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
	logPrefix := fmt.Sprintf("[%v][pin_rooms]", logID)
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
	var dataPinRooms []models.PinRoom
	if err := scDB.Find(&dataPinRooms).Error; err != nil {
		utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
		return
	}
	totalFetch := len(dataPinRooms)

	debug++
	utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
	debugT = time.Now()

	//Insert into the new PSQL database
	var errorMessages []string
	totalInserted := 0 //success insert
	errCount := 0
	for _, dataPinRoom := range dataPinRooms {
		mPinRooms := models.PinRooms{
			Id:        dataPinRoom.Id,
			UserId:    dataPinRoom.UserId,
			RoomId:    dataPinRoom.RoomId,
			CreatedAt: dataPinRoom.CreatedAt,
			CreatedBy: uuid.Nil,
			UpdatedAt: dataPinRoom.CreatedAt,
			UpdatedBy: uuid.Nil,
			DeletedAt: dataPinRoom.DeletedAt,
			DeletedBy: uuid.Nil,
		}

		if err := dstDB.Create(&mPinRooms).Error; err != nil {
			var errCode *pq.Error
			if errors.As(err, &errCode) {
				if errCode.Code == "23505" { //unique_violation
					mPinRooms.Id = uuid.NewV4()
					if err = dstDB.Create(&mPinRooms).Error; err != nil {
						utils.WriteLog(fmt.Sprintf("%s; failed insert to SQL: %s; error: %v", logPrefix, mPinRooms.Id, err), utils.LogLevelError)
					}
				}
			}

			utils.WriteLog(fmt.Sprintf("%s; failed insert to SQL: %s; error: %v", logPrefix, mPinRooms.Id, err), utils.LogLevelError)
			errCount++
			errorMessages = append(errorMessages, fmt.Sprintf("[%v] insert error: %v | DATA SQL: %v", time.Now(), err.Error(), mPinRooms))
			continue
		}
		totalInserted++
	}
	debug++
	utils.WriteLog(fmt.Sprintf("%s [PSQL] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalInserted, errCount, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
}
