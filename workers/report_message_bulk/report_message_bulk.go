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

	var lists []models.FetchReportReportMessage

	startDate := os.Getenv("START_DATE")
	endDate := os.Getenv("END_DATE")
	limit, _ := strconv.Atoi(os.Getenv("LIMIT"))

	loopCount := 0
reFetch:
	db = utils.GetDBNewConnection()

	debug := 0
	debugT := time.Now()

	db = db.Unscoped()
	//Set the filters
	if os.Getenv("COMPANYID") != "" {
		db = db.Joins("JOIN rooms ON chat_messages.room_id = rooms.id").Where("rooms.company_id = ?", os.Getenv("COMPANYID"))
	}

	if startDate != "" && endDate != "" {
		db = db.Where("chat_messages.created_at BETWEEN ? AND ?", startDate, endDate)
	} else if startDate != "" && endDate == "" {
		db = db.Where("chat_messages.created_at >=?", startDate)
	} else if startDate == "" && endDate != "" {
		db = db.Where("chat_messages.created_at <=?", endDate)
	}

	if os.Getenv("ORDER_BY") != "" {
		sortMap := map[string]string{
			os.Getenv("ORDER_BY"): "chat_messages." + os.Getenv("ORDER_BY"),
		}
		if strings.ToUpper(os.Getenv("ORDER_DIRECTION")) == "DESC" {
			db = db.Order(sortMap[os.Getenv("ORDER_BY")] + " DESC")
		} else {
			db = db.Order(sortMap[os.Getenv("ORDER_BY")])
		}
	}

	offset, _ := strconv.Atoi(os.Getenv("OFFSET"))
	if offset > 0 {
		db = db.Offset(offset)
	}
	if limit > 0 {
		db = db.Limit(limit)
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
		if err := db.Preload("Users").Find(&lists).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
			return
		}
		totalFetch = len(lists)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()
	}

	//Insert into database report
	var errorMessages []models.FetchReportReportMessage
	var errorDuplicates []models.FetchReportReportMessage
	totalInserted := 0

	for i, list := range lists {
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

		err := dbReport.Create(&m).Error
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s; [%v][>= '%v' <= '%v' LIMIT: %v] TOTAL_FETCH: %d; insert error: %v; id: %v", logPrefix, i, startDate, endDate, limit, totalFetch, err, list.Id), utils.LogLevelError)
			list.Error = err.Error()
			if errCode, ok := err.(*pq.Error); ok {
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
