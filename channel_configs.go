package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"konnek-migration/models"
	"konnek-migration/utils"
	"os"
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
	logPrefix := fmt.Sprintf("[%v] [Channel Configs]", logID)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	var channelConfigSc []models.ChannelConfigExist

	scDB = scDB.Preload("Channel")

	if os.Getenv("COMPANYID") != "" {
		scDB = scDB.Where("company_id = ?", os.Getenv("COMPANYID"))
	}

	//Fetch companies existing
	if err := scDB.Find(&channelConfigSc).Error; err != nil {
		utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
		return
	}

	totalCompany := len(channelConfigSc)

	debug++
	utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalCompany, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
	debugT = time.Now()

	insertedCount := 0
	successCount := 0
	errorCount := 0

	//var errorMessages []string

	//for _, channelConfig := range channelConfigSc {
	//var channelConfigDst models.ChannelConfigReeng

	//insertedCount++
	//	reiInsertCount := 0
	//reInsert:
	//if err := dstDB.Create(&comConfDst).Error; err != nil {
	//	if errCode, ok := err.(*pq.Error); ok {
	//		if errCode.Code == "23505" { //unique_violation
	//			reiInsertCount++
	//			comConfDst.Id = uuid.NewV4()
	//			if reiInsertCount < 3 {
	//				goto reInsert
	//			}
	//		}
	//	}
	//	utils.WriteLog(fmt.Sprintf("%s; [INSERT] Error: %v", logPrefix, err), utils.LogLevelError)
	//	errorCount++
	//	errorMessages = append(errorMessages, fmt.Sprintf("%s [INSERT] Error: %v ; DATA: %v", time.Now(), err, comConfDst))
	//	continue
	//}

	//successCount++
	//}
	//
	//// Write error messages to a text file
	//formattedTime := time.Now().Format("2006-01-02_150405")
	//errorFileLog := fmt.Sprintf("error_messages_companies_%s.log", formattedTime)
	//if len(errorMessages) > 0 {
	//	createFile, errCreate := os.Create(errorFileLog)
	//	if errCreate != nil {
	//		utils.WriteLog(fmt.Sprintf("%s [ERROR] Create File Error Log; Error: %v;", logPrefix, errCreate), utils.LogLevelError)
	//		return
	//	}
	//	defer createFile.Close()
	//
	//	for _, errMsg := range errorMessages {
	//		_, errWrite := createFile.WriteString(errMsg + "\n")
	//		if errWrite != nil {
	//			utils.WriteLog(fmt.Sprintf("%s [ERROR] Write File Error Log; Filename: %s; Error: %v;", logPrefix, errorFileLog, errCreate), utils.LogLevelError)
	//		}
	//	}
	//
	//	utils.WriteLog(fmt.Sprintf("%s [ERROR] Error messages written to %s", logPrefix, errorFileLog), utils.LogLevelError)
	//}

	debug++
	utils.WriteLog(fmt.Sprintf("%s [INSERT] TOTAL_INSERTED: %d; TOTAL_SUCCESS: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, insertedCount, successCount, errorCount, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

	utils.WriteLog(fmt.Sprintf("%s end; duration: %v", logPrefix, time.Now().Sub(tStart)), utils.LogLevelDebug)
}
