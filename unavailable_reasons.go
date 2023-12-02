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
	logPrefix := fmt.Sprintf("[%v] [unavailable_reason]", logID)
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

	var lists []models.MtrReasonAvailability
	if err := scDB.Find(&lists).Error; err != nil {
		utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
		return
	}
	totalFetch := len(lists)

	debug++
	utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
	debugT = time.Now()

	//Insert into the new database
	//var errorMessages []string
	var m models.UnavailableReason
	for _, list := range lists {
		m.Id = list.Id
		m.CompanyId = list.CompanyId
		m.Type = list.Type
		m.Reason = list.Reason
		m.CreatedAt = list.CreatedAt
		m.CreatedBy = list.CreatedBy
		m.UpdatedAt = list.UpdatedAt
		m.UpdatedBy = list.UpdatedBy
		m.DeletedAt = list.DeletedAt
		m.DeletedBy = list.DeletedBy

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
			//errorMessages = append(errorMessages, fmt.Sprintf("error: %s; query: %s", err.Error()))
			continue
		}
	}

	utils.WriteLog(fmt.Sprintf("%s end; duration: %v", logPrefix, time.Now().Sub(tStart)), utils.LogLevelDebug)
}
