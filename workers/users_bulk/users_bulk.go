package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/joho/godotenv/autoload"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"konnek-migration/models"
	"konnek-migration/utils"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	utils.Init()

	// Create source DB connection
	scDB := utils.GetDBConnection()
	defer func(scDB *gorm.DB) {
		err := scDB.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(scDB)

	scDb2 := scDB

	// Create destination DB connection
	dstDB := utils.GetDBNewConnection()
	defer func(dstDB *gorm.DB) {
		err := dstDB.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(dstDB)

	logID := uuid.NewV4()
	appName := "users"
	if os.Getenv("APP_NAME") != "" {
		appName = os.Getenv("APP_NAME")
	}

	logPrefix := fmt.Sprintf("[%v] [%s]", logID, appName)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()

	var lists []models.User

	startDate := os.Getenv("START_DATE")
	endDate := os.Getenv("END_DATE")
	limit, _ := strconv.Atoi(os.Getenv("LIMIT"))

	loopCount := 0
reFetch:
	scDB = utils.GetDBConnection()

	debug := 0
	debugT := time.Now()

	scDB = scDB.Unscoped()
	//Set the filters
	if os.Getenv("COMPANYID") != "" {
		scDB = scDB.Where("company_id = ?", os.Getenv("COMPANYID"))
	}

	if os.Getenv("ROLES_ID") != "" {
		idSlice := strings.Split(os.Getenv("ROLES_ID"), ",")
		scDB = scDB.Where("roles_id IN (?)", idSlice)
	}

	if startDate != "" && endDate != "" {
		scDB = scDB.Where("created_at BETWEEN ? AND ?", startDate, endDate)
	} else if startDate != "" && endDate == "" {
		scDB = scDB.Where("created_at >=?", startDate)
	} else if startDate == "" && endDate != "" {
		scDB = scDB.Where("created_at <=?", endDate)
	}

	if os.Getenv("ORDER_BY") != "" {
		sortMap := map[string]string{
			os.Getenv("ORDER_BY"): os.Getenv("ORDER_BY"),
		}
		if strings.ToUpper(os.Getenv("ORDER_DIRECTION")) == "DESC" {
			scDB = scDB.Order(sortMap[os.Getenv("ORDER_BY")] + " DESC")
		} else {
			scDB = scDB.Order(sortMap[os.Getenv("ORDER_BY")])
		}
	}

	offset, _ := strconv.Atoi(os.Getenv("OFFSET"))
	if offset > 0 {
		scDB = scDB.Offset(offset)
	}
	if limit > 0 {
		scDB = scDB.Limit(limit)
	}

	totalFetch := 0

	// Get from file
	if os.Getenv("GET_FROM_FILE") != "" {
		utils.WriteLog(fmt.Sprintf("%s get from file %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelDebug)
		// Read the JSON file
		fileContent, err := ioutil.ReadFile("../../data/" + os.Getenv("GET_FROM_FILE"))
		if err != nil {
			fmt.Printf("%s Error reading file: %v\n", logPrefix, err)
			utils.WriteLog(fmt.Sprintf("%s Error reading file: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
			return
		}

		// Unmarshal the JSON data into the struct
		err = json.Unmarshal(fileContent, &lists)
		if err != nil {
			fmt.Printf("%s Error unmarshalling: %v\n", logPrefix, err)
			utils.WriteLog(fmt.Sprintf("%s Error unmarshalling JSON: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
			return
		}
		totalFetch = len(lists)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [GET_FROM_FILE] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, len(lists), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()

		err = os.Remove("../../data/" + os.Getenv("GET_FROM_FILE"))
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s Error Delete file: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
		}
	} else {
		//Fetch from database
		if err := scDB.Preload("Roles").Preload("Customer").Preload("Employee").Find(&lists).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
			return
		}
		totalFetch = len(lists)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [FETCH] [>= '%v' <= '%v' LIMIT: %v] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, startDate, endDate, limit, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()
	}

	//Insert into the new database
	var errorMessages []models.User
	var errorDuplicates []models.User
	totalInserted := 0
	TempRole := make(map[string]uuid.UUID)

	for i, list := range lists {
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
			_ = scDb2.Raw("SELECT id, channel_name FROM room_details WHERE customer_user_id = ? ", list.Id).Scan(&roomDetails).Error

			customerChannel = roomDetails.ChannelName
			name = list.Customer.Name
			tags = list.Customer.Tags
			replyToken = list.Customer.Reply_Token
		} else {
			name = list.Employee.Name
			divisionId = list.Employee.DivisionId
		}

		userType := ""
		if _, ok := models.UserTypeMap[list.Roles.Name]; ok {
			userType = strconv.Itoa(models.UserTypeMap[list.Roles.Name])
		}

		var m models.Users
		m.Id = list.Id
		m.CompanyId = list.CompanyId
		m.RolesId = rolesId
		m.Type = userType
		m.CustomerChannel = customerChannel
		m.Username = list.Username
		m.Password = list.Password
		m.LastChangePwd = &list.CreatedAt
		m.Status = 1
		m.OnlineStatus = list.OnlineStatus
		m.LoginTime = list.LoginTime
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

		err := dstDB.Create(&m).Error
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s; [%v][>= '%v' <= '%v' LIMIT: %v] TOTAL_FETCH: %d; insert error: %v; id: %v", logPrefix, i, startDate, endDate, limit, totalFetch, err, list.Id), utils.LogLevelError)
			if errCode, ok := err.(*pq.Error); ok {
				list.Error = err.Error()
				if errCode.Code == "23505" { //unique_violation
					errorDuplicates = append(errorDuplicates, list)
				} else {
					errorMessages = append(errorMessages, list)
				}
			}
		}
		totalInserted++

		if i >= limit-1 {
			debug++
			utils.WriteLog(fmt.Sprintf("%s [PSQL] [>= '%v' <= '%v' LIMIT: %v] TOTAL_FETCH: %d; TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, startDate, endDate, limit, totalFetch, totalInserted, len(errorMessages), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

			startDate = list.CreatedAt.Format("2006-01-02 15:04:05.999999999+07")
			utils.WriteLog(fmt.Sprintf("%s [%v] last created_at:%v; set startDate:%v; endDate:%v; TOTAL_INSERTED: %d; DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;\n", logPrefix, loopCount, list.CreatedAt, startDate, endDate, totalInserted, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

			loopCount++
			goto reFetch
		}
	}
	//debug++
	//utils.WriteLog(fmt.Sprintf("%s [INSERT] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalInserted, len(errorMessages), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

	//write error to file
	if len(errorMessages) > 0 {
		filename := fmt.Sprintf("%s_%s_%v", appName, time.Now().Format("2006_01_02"), time.Now().Unix())
		utils.WriteErrorMap(filename, errorMessages)
	}
	if len(errorDuplicates) > 0 {
		filename := fmt.Sprintf("%s_%s_%v_duplicate", appName, time.Now().Format("2006_01_02"), time.Now().Unix())
		utils.WriteErrorMap(filename, errorDuplicates)
	}

	utils.WriteLog(fmt.Sprintf("%s end; duration: %v", logPrefix, time.Now().Sub(tStart)), utils.LogLevelDebug)
}
