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

	// Create destination DB connection
	dstDB := utils.GetDBNewConnection()
	defer func(dstDB *gorm.DB) {
		err := dstDB.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(dstDB)

	logID := uuid.NewV4()
	appName := "history_change_unavailable_reason"
	if os.Getenv("APP_NAME") != "" {
		appName = os.Getenv("APP_NAME")
	}

	logPrefix := fmt.Sprintf("[%v] [%s]", logID, appName)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	var lists []models.HistoryChangeReasonAvailability

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
		debug++
		utils.WriteLog(fmt.Sprintf("%s [GET_FROM_FILE] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, len(lists), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()

		err = os.Remove("../../data/" + os.Getenv("GET_FROM_FILE"))
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s Error Delete file: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
		}
	} else {
		//Fetch the database
		scDB = scDB.Unscoped()
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
				os.Getenv("ORDER_BY"): os.Getenv("ORDER_BY"),
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

		if err := scDB.Find(&lists).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
			return
		}
		totalFetch := len(lists)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()
	}

	//Insert into the new database
	var errorMessages []models.HistoryChangeReasonAvailability
	var errorDuplicates []models.HistoryChangeReasonAvailability
	totalInserted := 0
	var m models.HistoryChangeUnavailableReason
	for _, list := range lists {
		m.Id = list.Id
		m.CompanyId = list.CompanyId
		m.UserId = list.UserId
		m.OldType = list.OldType
		m.NewType = list.NewType
		m.OldReason = list.OldReason
		m.NewReason = list.NewReason
		m.CreatedAt = list.CreatedAt
		m.CreatedBy = uuid.Nil
		m.UpdatedAt = list.CreatedAt
		m.UpdatedBy = uuid.Nil
		m.DeletedBy = uuid.Nil

		//	reiInsertCount := 0
		//reInsert:
		if err := dstDB.Create(&m).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; insert error: %v", logPrefix, err), utils.LogLevelError)
			list.Error = err.Error()
			if errCode, ok := err.(*pq.Error); ok {
				if errCode.Code == "23505" { //unique_violation
					errorDuplicates = append(errorDuplicates, list)
					continue
				}
			}
			errorMessages = append(errorMessages, list)
			continue
		}
		totalInserted++
	}
	debug++
	utils.WriteLog(fmt.Sprintf("%s [INSERT] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalInserted, len(errorMessages), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

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
