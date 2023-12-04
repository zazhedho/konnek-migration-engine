package main

import (
	"errors"
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
	logPrefix := fmt.Sprintf("[%v][sessions]", logID)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	//Fetch the data from existing PSQL database
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

	// query data dari source PSQL DB
	var dataSessions []models.Session
	if err := scDB.Preload("ChatMessage", func(db *gorm.DB) *gorm.DB {
		return db.Order("chat_messages.created_at DESC")
	}).Find(&dataSessions).Error; err != nil {
		utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
	}
	totalFetch := len(dataSessions)

	debug++
	utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
	debugT = time.Now()

	//Insert into the new PSQL database
	var errorMessages []string
	totalInserted := 0 //success insert
	errCount := 0
	for _, dataSession := range dataSessions {
		status := 0
		if dataSession.QueTime.IsZero() {
			status = 0
		} else if !dataSession.QueTime.IsZero() && dataSession.AgentUserId == uuid.Nil {
			status = 1
		} else if dataSession.AgentUserId != uuid.Nil && !dataSession.AssignTime.IsZero() {
			status = 2
		} else if !dataSession.CloseTime.IsZero() {
			status = -1
		}

		mSessions := models.Sessions{
			Id:                dataSession.Id,
			RoomId:            dataSession.RoomId,
			DivisionId:        dataSession.DivisionId,
			AgentUserId:       dataSession.AgentUserId,
			LastMessageId:     dataSession.ChatMessage.Id, //query ke table chat_messages
			Categories:        dataSession.Categories,
			BotStatus:         dataSession.BotStatus,
			Status:            status, //query status dari room_details existing
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
			var errCode *pq.Error
			if errors.As(err, &errCode) {
				if errCode.Code == "23505" { //unique_violation
					mSessions.Id = uuid.NewV4()
					if err = dstDB.Create(&mSessions).Error; err != nil {
						utils.WriteLog(fmt.Sprintf("%s; failed insert to SQL: %s; error: %v", logPrefix, mSessions.Id, err), utils.LogLevelError)
					}
				}
			}

			utils.WriteLog(fmt.Sprintf("%s; failed insert to SQL: %s; error: %v", logPrefix, mSessions.Id, err), utils.LogLevelError)
			errCount++
			errorMessages = append(errorMessages, fmt.Sprintf("[%v] insert error: %v | DATA SQL: %v", time.Now(), err.Error(), mSessions))
			continue
		}
		totalInserted++
	}
	debug++
	utils.WriteLog(fmt.Sprintf("%s [PSQL] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalInserted, errCount, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
}
