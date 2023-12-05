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
	"log"
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
	logPrefix := fmt.Sprintf("[%v][rooms]", logID)
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
	var dataRoomDetails []models.RoomDetails
	if err := scDB.Find(&dataRoomDetails).Error; err != nil {
		utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
		return
	}
	totalFetch := len(dataRoomDetails)

	debug++
	utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
	debugT = time.Now()

	//Insert into the new PSQL database
	var errorMessages []string
	totalInsertedSQL := 0 //success insert
	errCountSQL := 0
	for _, dataRoomDetail := range dataRoomDetails {
		mRooms := models.Rooms{
			ID:             dataRoomDetail.ID,
			CompanyID:      dataRoomDetail.CompanyID,
			ChannelCode:    dataRoomDetail.ChannelName,
			CustomerUserID: dataRoomDetail.CustomerUserID,
			LastSessionID:  dataRoomDetail.SessionID,
			CreatedAt:      dataRoomDetail.CreatedAt,
			CreatedBy:      uuid.Nil,
			UpdatedAt:      dataRoomDetail.CreatedAt,
			UpdatedBy:      uuid.Nil,
			DeletedAt:      dataRoomDetail.DeletedAt,
			DeletedBy:      uuid.Nil,
		}

		if err := dstDB.Create(&mRooms).Error; err != nil {
			var errCode *pq.Error
			if errors.As(err, &errCode) {
				if errCode.Code == "23505" { //unique_violation
					mRooms.ID = uuid.NewV4()
					if err = dstDB.Create(&mRooms).Error; err != nil {
						utils.WriteLog(fmt.Sprintf("%s; failed insert to SQL: %s; error: %v", logPrefix, dataRoomDetail.ID, err), utils.LogLevelError)
					}
				}
			}

			utils.WriteLog(fmt.Sprintf("%s; failed insert to SQL: %s; error: %v", logPrefix, dataRoomDetail.ID, err), utils.LogLevelError)
			errCountSQL++
			errorMessages = append(errorMessages, fmt.Sprintf("[%v] insert error: %v | DATA SQL: %v", time.Now(), err.Error(), mRooms))
			continue
		}
		totalInsertedSQL++
	}

	// Write error messages to a text file
	formattedTime := time.Now().Format("2006-01-02_150405")
	logErrorFilename := fmt.Sprintf("error_messages_%s.log", formattedTime)
	if len(errorMessages) > 0 {
		createFIle, errCreate := os.Create(logErrorFilename)
		if errCreate != nil {
			utils.WriteLog(fmt.Sprintf("%s [ERROR] Failed write to %s | Error: %v", logPrefix, logErrorFilename, errCreate), utils.LogLevelError)
			return
		}
		defer createFIle.Close()

		for _, errMsg := range errorMessages {
			_, errWrite := createFIle.WriteString(errMsg + "\n")
			if errWrite != nil {
				log.Printf("Failed to write to error_messages_%s.log : %v", formattedTime, errWrite)
				utils.WriteLog(fmt.Sprintf("%s [ERROR] Failed write to %s | Error: %v", logPrefix, logErrorFilename, errWrite), utils.LogLevelError)
			}
		}
		utils.WriteLog(fmt.Sprintf("%s [SUCCESS] Messages error written to %s", logPrefix, logErrorFilename), utils.LogLevelInfo)
	}
	debug++
	utils.WriteLog(fmt.Sprintf("%s [PSQL] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalInsertedSQL, errCountSQL, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
}
