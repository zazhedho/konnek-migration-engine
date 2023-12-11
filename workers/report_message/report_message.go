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
	db := utils.GetDBNewConnection()
	defer func(db *gorm.DB) {
		err := db.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(db)

	dbReport := utils.GetDBReportConnection()
	defer func(dbReport *gorm.DB) {
		err := dbReport.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(dbReport)

	logID := uuid.NewV4()
	appName := "report_message"
	if os.Getenv("APP_NAME") != "" {
		appName = os.Getenv("APP_NAME")
	}

	logPrefix := fmt.Sprintf("[%v] [%s]", logID, appName)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	var lists []models.FetchReportReportMessage

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
		//Fetch from database

		//Set the filters
		if os.Getenv("COMPANYID") != "" {
			db = db.Preload("Users", func(db *gorm.DB) *gorm.DB {
				return db.Where("company_id = ?", os.Getenv("COMPANYID"))
			})
		} else {
			db = db.Preload("Users")
		}

		if os.Getenv("START_DATE") != "" && os.Getenv("END_DATE") != "" {
			db = db.Where("created_at BETWEEN ? AND ?", os.Getenv("START_DATE"), os.Getenv("END_DATE"))
		} else if os.Getenv("START_DATE") != "" && os.Getenv("END_DATE") == "" {
			db = db.Where("created_at >=?", os.Getenv("START_DATE"))
		} else if os.Getenv("START_DATE") == "" && os.Getenv("END_DATE") != "" {
			db = db.Where("created_at <=?", os.Getenv("END_DATE"))
		}

		if os.Getenv("ORDER_BY") != "" {
			sortMap := map[string]string{
				"created_at": "created_at",
			}
			if strings.ToUpper(os.Getenv("ORDER_DIRECTION")) == "DESC" {
				db = db.Order(sortMap[os.Getenv("ORDER_BY")] + " DESC")
			} else {
				db = db.Order(sortMap[os.Getenv("ORDER_BY")])
			}
		}

		offset, _ := strconv.Atoi(os.Getenv("OFFSET"))
		limit, _ := strconv.Atoi(os.Getenv("LIMIT"))

		if offset > 0 {
			db = db.Offset(offset)
		}

		if limit > 0 {
			db = db.Limit(limit)
		}

		if err := db.Find(&lists).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
			return
		}
		totalFetch := len(lists)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()
	}

	//Insert into database report
	var errorMessages []models.FetchReportReportMessage
	var errorDuplicates []models.FetchReportReportMessage
	totalInserted := 0

	for _, list := range lists {
		var m models.ReportMessage
		m.TablePrefix = list.Users.CompanyId.String() + "_"
		//check table, create table if it doesn't exist
		//dbReport.AutoMigrate(&m)

		m.Id = list.Id
		m.MessageId = list.MessageId
		m.ReplyId = list.ReplyId
		m.RoomId = list.RoomId
		m.SessionId = list.SessionId
		m.FromType = list.FromType
		m.UserId = list.UserId
		m.Username = list.Users.Username
		m.UserFullname = list.Users.Name
		m.Type = list.Type
		m.Text = list.Text
		m.Payload = list.Payload
		m.Status = list.Status
		m.MessageTime = list.MessageTime
		m.LastUpdate = time.Now()
		m.CreatedAt = time.Now()
		m.CreatedBy = "migration-engine"
		m.UpdatedAt = time.Now()
		m.UpdatedBy = "migration-engine"

		if err := dbReport.Create(&m).Error; err != nil {
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
