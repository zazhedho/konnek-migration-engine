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
			utils.WriteLog(fmt.Sprintf("Close Connection sourceDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(scDB)

	// Create destination DB connection
	dstDB := utils.GetDBNewConnection()
	defer func(dstDB *gorm.DB) {
		err := dstDB.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection destinationsDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(dstDB)

	logID := uuid.NewV4()
	appName := "sessions"
	if os.Getenv("APP_NAME") != "" {
		appName = os.Getenv("APP_NAME")
	}
	logPrefix := fmt.Sprintf("[%v][%v]", logID, appName)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()

	var dataSessions []models.Session

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
	scDB = scDB.Joins("JOIN room_details ON sessions.room_id = room_details.id")
	if os.Getenv("COMPANYID") != "" {
		scDB = scDB.Where("room_details.company_id = ?", os.Getenv("COMPANYID"))
	}

	if startDate != "" && endDate != "" {
		scDB = scDB.Where("sessions.created_at BETWEEN ? AND ?", startDate, endDate)
	} else if startDate != "" && endDate == "" {
		scDB = scDB.Where("sessions.created_at >=?", startDate)
	} else if startDate == "" && endDate != "" {
		scDB = scDB.Where("sessions.created_at <=?", endDate)
	}

	if os.Getenv("ORDER_BY") != "" {
		sortMap := map[string]string{
			os.Getenv("ORDER_BY"): "sessions." + os.Getenv("ORDER_BY"),
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
		err = json.Unmarshal(fileContent, &dataSessions)
		if err != nil {
			fmt.Printf("%s Error unmarshalling: %v\n", logPrefix, err)
			utils.WriteLog(fmt.Sprintf("%s Error unmarshalling JSON: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
			return
		}
		totalFetch = len(dataSessions)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [GET_FROM_FILE] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, len(dataSessions), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()

		err = os.Remove("../../data/" + os.Getenv("GET_FROM_FILE"))
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s Error Delete file: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
		}
	} else {
		//Fetch the data from existing PSQL database
		// query data dari source PSQL DB
		if err := scDB.Preload("ChatMessage", func(db *gorm.DB) *gorm.DB {
			return db.Order("chat_messages.created_at DESC")
		}).Find(&dataSessions).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
			return
		}
		totalFetch = len(dataSessions)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [FETCH] [>= '%v' <= '%v' LIMIT: %v] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, startDate, endDate, limit, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()
	}

	//Insert into the new PSQL database
	var errorMessages []models.Session
	var errorDuplicates []models.Session
	totalInserted := 0 //success insert
	debugT = time.Now()

	for i, dataSession := range dataSessions {
		status := 0
		if dataSession.QueTime.IsZero() && dataSession.CloseTime.IsZero() {
			status = 0 //Open
		} else if !dataSession.QueTime.IsZero() && dataSession.AgentUserId == uuid.Nil && dataSession.CloseTime.IsZero() {
			status = 1 //Waiting
		} else if dataSession.AgentUserId != uuid.Nil && !dataSession.AssignTime.IsZero() && dataSession.CloseTime.IsZero() {
			status = 2 //Assigned
		} else if !dataSession.CloseTime.IsZero() {
			status = -1 //Closed
		}

		mSessions := models.Sessions{
			Id:                dataSession.Id,
			RoomId:            dataSession.RoomId,
			DivisionId:        dataSession.DivisionId,
			AgentUserId:       dataSession.AgentUserId,
			LastMessageId:     dataSession.ChatMessage.Id, //query ke table chat_messages
			Categories:        dataSession.Categories,
			BotStatus:         dataSession.BotStatus,
			Status:            status,
			OpenBy:            dataSession.OpenBy,
			HandoverBy:        dataSession.HandOverBy,
			CloseBy:           dataSession.CloseBy,
			SlaFrom:           dataSession.SlaFrom,
			SlaTo:             dataSession.SlaTo,
			SlaTreshold:       dataSession.SlaTreshold,
			SlaDurations:      dataSession.SlaDurations,
			SlaStatus:         dataSession.SlaStatus,
			OpenTime:          &dataSession.CreatedAt,
			QueTime:           &dataSession.QueTime,
			AssignTime:        &dataSession.AssignTime,
			FirstResponseTime: &dataSession.FirstResponseTime,
			LastAgentChatTime: &dataSession.LastAgentChatTime,
			CloseTime:         &dataSession.CloseTime,
			CreatedAt:         dataSession.CreatedAt,
			CreatedBy:         uuid.Nil,
			UpdatedAt:         dataSession.CreatedAt,
			UpdatedBy:         uuid.Nil,
			DeletedAt:         dataSession.DeletedAt,
			DeletedBy:         uuid.Nil,
		}

		err := dstDB.Create(&mSessions).Error
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s; [%v][>= '%v' <= '%v' LIMIT: %v] TOTAL_FETCH: %d; insert error: %v; id: %v", logPrefix, i, startDate, endDate, limit, totalFetch, err, dataSession.Id), utils.LogLevelError)
			dataSession.Error = err.Error()
			if errCode, ok := err.(*pq.Error); ok {
				if errCode.Code == "23505" { //unique_violation
					errorDuplicates = append(errorDuplicates, dataSession)
				} else {
					errorMessages = append(errorMessages, dataSession)
				}
			}
		}
		totalInserted++

		if i >= limit-1 {
			debug++
			utils.WriteLog(fmt.Sprintf("%s [PSQL] [>= '%v' <= '%v' LIMIT: %v] TOTAL_FETCH: %d; TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, startDate, endDate, limit, totalFetch, totalInserted, len(errorMessages), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

			startDate = dataSession.CreatedAt.Format("2006-01-02 15:04:05.999999999+07")
			utils.WriteLog(fmt.Sprintf("%s [%v] last created_at:%v; set startDate:%v; endDate:%v; TOTAL_INSERTED: %d; DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;\n", logPrefix, loopCount, dataSession.CreatedAt, startDate, endDate, totalInserted, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

			loopCount++
			goto reFetch
		}
	}
	//debug++
	//utils.WriteLog(fmt.Sprintf("%s [PSQL] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalInserted, len(errorMessages), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

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
