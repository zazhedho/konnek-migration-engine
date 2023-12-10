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
	debug := 0
	debugT := time.Now()

	var dataSessions []models.Session
	// Get from file
	if os.Getenv("GET_FROM_FILE") != "" {
		utils.WriteLog(fmt.Sprintf("%s get from file %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelDebug)
		// Read the JSON file
		fileContent, err := ioutil.ReadFile("data/" + os.Getenv("GET_FROM_FILE"))
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
		debug++
		utils.WriteLog(fmt.Sprintf("%s [GET_FROM_FILE] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, len(dataSessions), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()

		err = os.Remove("data/" + os.Getenv("GET_FROM_FILE"))
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s Error Delete file: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
		}
	} else {
		//Fetch the data from existing PSQL database
		//Set the filters
		if os.Getenv("COMPANYID") != "" {
			scDB = scDB.Preload("RoomDetail", func(db *gorm.DB) *gorm.DB {
				return db.Where("company_id = ?", os.Getenv("COMPANYID"))
			})
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

		// query data dari source PSQL DB
		if err := scDB.Preload("ChatMessage", func(db *gorm.DB) *gorm.DB {
			return db.Order("chat_messages.created_at DESC")
		}).Find(&dataSessions).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
		}
		totalFetch := len(dataSessions)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()
	}

	//Insert into the new PSQL database
	var errorMessages []models.Session
	var errorDuplicates []models.Session
	totalInserted := 0 //success insert
	for _, dataSession := range dataSessions {
		status := 0
		if dataSession.QueTime.IsZero() {
			status = 0 //Open
		} else if !dataSession.QueTime.IsZero() && dataSession.AgentUserId == uuid.Nil {
			status = 1 //Waiting
		} else if dataSession.AgentUserId != uuid.Nil && !dataSession.AssignTime.IsZero() {
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

		if err := dstDB.Create(&mSessions).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; insert error: %v", logPrefix, err), utils.LogLevelError)
			dataSession.Error = err.Error()
			if errCode, ok := err.(*pq.Error); ok {
				if errCode.Code == "23505" { //unique_violation
					errorDuplicates = append(errorDuplicates, dataSession)
					continue
				}
			}
			errorMessages = append(errorMessages, dataSession)
			continue
		}
		totalInserted++
	}
	debug++
	utils.WriteLog(fmt.Sprintf("%s [PSQL] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalInserted, len(errorMessages), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

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
