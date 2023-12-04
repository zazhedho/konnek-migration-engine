package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/joho/godotenv/autoload"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"konnek-migration/models"
	"konnek-migration/utils"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Create source DB connection
	scDB := utils.GetDBConnection()
	defer func(scDB *gorm.DB) {
		err := scDB.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(scDB)

	// Create destination DB connection
	dstDB := utils.GetDBNewConnection()
	defer func(dstDB *gorm.DB) {
		err := dstDB.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(dstDB)

	logID := uuid.NewV4()
	logPrefix := fmt.Sprintf("[%v] [users]", logID)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	//Fetch the database

	//Set the filters
	if os.Getenv("COMPANYID") != "" {
		scDB = scDB.Where("company_id = ?", os.Getenv("COMPANYID"))
	}

	if os.Getenv("START_DATE") != "" && os.Getenv("END_DATE") != "" {
		scDB = scDB.Where("created_at BETWEEN ? AND ?", os.Getenv("START_DATE"), os.Getenv("END_DATE"))
	} else if os.Getenv("START_DATE") != "" && os.Getenv("END_DATE") == "" {
		scDB = scDB.Where("created_at >=?", os.Getenv("START_DATE"))
	} else if os.Getenv("START_DATE") == "" && os.Getenv("END_DATE") != "" {
		scDB = scDB.Where("created_at <=?", os.Getenv("END_DATE"))
	}

	if os.Getenv("ORDER_BY") != "" {
		sortMap := map[string]string{
			"created_at": "created_at",
		}
		if strings.ToUpper(os.Getenv("ORDER_DIRECTION")) == "DESC" {
			scDB = scDB.Order(sortMap[os.Getenv("ORDER_BY")] + " DESC")
		} else {
			scDB = scDB.Order(sortMap[os.Getenv("ORDER_BY")])
		}
	}

	offset, _ := strconv.Atoi(os.Getenv("OFFSET"))
	limit, _ := strconv.Atoi(os.Getenv("LIMIT"))

	if offset > 0 {
		scDB = scDB.Offset(offset)
	}

	if limit > 0 {
		scDB = scDB.Limit(limit)
	}

	var lists []models.User
	if err := scDB.Preload("Roles").Preload("Customer").Preload("Employee").Find(&lists).Error; err != nil {
		utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
		return
	}
	totalFetch := len(lists)

	debug++
	utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
	debugT = time.Now()

	//Insert into the new database
	var errorMessages []string
	totalInserted := 0
	var TempRole map[string]uuid.UUID

	for _, list := range lists {
		var rolesId uuid.UUID

		//get rolesId from temporary role list
		keyRole := fmt.Sprintf("%s%s", list.Roles.Name, list.CompanyId)
		if v, ok := TempRole[keyRole]; ok {
			rolesId = v
		} else {
			var getRoles models.Roles
			if err := dstDB.Where("name = ? AND company_id = ?", list.Roles.Name, list.CompanyId).Find(&getRoles).Error; err != nil {
				utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
				continue
			}
			TempRole[keyRole] = getRoles.Id
			rolesId = getRoles.Id
		}

		var customerChannel string
		var tags string
		var replyToken string
		var divisionId uuid.UUID
		name := list.Username // role admin_konnek tidak termasuk employees
		if list.Roles.Name == models.RoleCustomer {
			type RoomDetails struct {
				RoomId      uuid.UUID `json:"id" gorm:"column:id"`
				ChannelName string    `json:"channel_name" gorm:"column:channel_name"`
			}

			var roomDetails RoomDetails
			_ = scDB.Where("customer_user_id = ? ", list.Id).Find(&roomDetails).Error

			customerChannel = roomDetails.ChannelName
			name = list.Customer.Name
			tags = list.Customer.Tags
			replyToken = list.Customer.Reply_Token
		} else {
			name = list.Employee.Name
			divisionId = list.Employee.DivisionId
		}

		var m models.Users
		m.Id = list.Id
		m.CompanyId = list.CompanyId
		m.RolesId = rolesId
		m.Type = strconv.Itoa(models.UserTypeMap[list.Roles.Name])
		m.CustomerChannel = customerChannel
		m.Username = list.Username
		m.Password = list.Password
		m.Status = 1
		m.OnlineStatus = list.OnlineStatus
		m.LoginTime = &list.LoginTime
		m.LoginSession = uuid.Nil
		m.Email = list.Email
		m.Phone = list.PhoneNumber
		m.Name = name
		m.Avatar = list.AvatarUrl
		m.Description = list.Description
		m.Tags = tags
		m.CustomerReplyToken = replyToken
		m.DivisionId = divisionId
		m.SoundNotification = list.NotifSound
		m.CreatedAt = list.CreatedAt
		m.CreatedBy = uuid.Nil
		m.UpdatedAt = list.UpdatedAt
		m.UpdatedBy = uuid.Nil
		m.DeletedAt = list.DeletedAt
		m.DeletedBy = uuid.Nil

		//	reiInsertCount := 0
		//reInsert:
		if err := dstDB.Create(&m).Error; err != nil {
			if errCode, ok := err.(*pq.Error); ok {
				if errCode.Code == "23505" { //unique_violation
					//reiInsertCount++
					//m.Id = uuid.NewV4()
					//if reiInsertCount < 3 {
					//	goto reInsert
					//}
				}
			}
			//TODO: write query insert to file
			utils.WriteLog(fmt.Sprintf("%s; insert error: %v", logPrefix, err), utils.LogLevelError)
			errorMessages = append(errorMessages, fmt.Sprintf("id: %s; error: %s; ", m.Id, err.Error()))
			continue
		}
		totalInserted++
	}
	debug++
	utils.WriteLog(fmt.Sprintf("%s [INSERT] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalInserted, len(errorMessages), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

	utils.WriteLog(fmt.Sprintf("%s end; duration: %v", logPrefix, time.Now().Sub(tStart)), utils.LogLevelDebug)
}
