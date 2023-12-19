package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"konnek-migration/models"
	"konnek-migration/utils"
	"os"
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
	appName := "channel_configs"
	if os.Getenv("APP_NAME") != "" {
		appName = os.Getenv("APP_NAME")
	}
	logPrefix := fmt.Sprintf("[%v] [%s]", logID, appName)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	var channelConfigSc []models.ChannelConfigExist

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
		err = json.Unmarshal(fileContent, &channelConfigSc)
		if err != nil {
			fmt.Printf("%s Error unmarshalling: %v\n", logPrefix, err)
			utils.WriteLog(fmt.Sprintf("%s Error unmarshalling JSON: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
			return
		}
		debug++
		utils.WriteLog(fmt.Sprintf("%s [GET_FROM_FILE] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, len(channelConfigSc), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()

		err = os.Remove("../../data/" + os.Getenv("GET_FROM_FILE"))
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s Error Delete file: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
		}
	} else {
		//Fetch from database
		scDB = scDB.Preload("Channel")

		if os.Getenv("COMPANYID") != "" {
			companyId := strings.Split(os.Getenv("COMPANYID"), ",")
			scDB = scDB.Where("company_id IN (?)", companyId)
		}

		if os.Getenv("CHANNELID") != "" {
			scDB = scDB.Where("channel_id = ?", os.Getenv("CHANNELID"))
		}

		//Fetch companies existing
		if err := scDB.Find(&channelConfigSc).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
			return
		}

		totalConfig := len(channelConfigSc)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalConfig, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()
	}

	insertedCount := 0
	successCount := 0
	errorCount := 0

	var errorMessages []models.ChannelConfigExist
	var errorDuplicates []models.ChannelConfigExist

	for _, channelConfig := range channelConfigSc {
		var channelConfigDst models.ChannelConfigReeng

		channelConfigDst.Id = channelConfig.Id
		channelConfigDst.CompanyId = channelConfig.CompanyId

		channelName := channelConfig.Channel.Name
		if channelName == "widget" {
			channelName = "web"
		}
		channelConfigDst.ChannelCode = channelName
		channelConfigDst.Key = channelConfig.Key
		channelConfigDst.Content = channelConfig.Content
		channelConfigDst.CreatedAt = channelConfig.CreatedAt
		channelConfigDst.CreatedBy = uuid.Nil
		channelConfigDst.UpdatedAt = channelConfig.UpdatedAt
		channelConfigDst.UpdatedBy = uuid.Nil
		channelConfigDst.DeletedAt = channelConfig.DeletedAt
		channelConfigDst.DeletedBy = uuid.Nil
		channelConfigDst.ErrMessage = channelConfig.ErrMessage

		insertedCount++
		//	reiInsertCount := 0
		//reInsert:
		if err := dstDB.Create(&channelConfigDst).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; [INSERT] Error: %v", logPrefix, err), utils.LogLevelError)
			channelConfig.Error = err.Error()
			if errCode, ok := err.(*pq.Error); ok {
				if errCode.Code == "23505" { //unique_violation
					errorDuplicates = append(errorDuplicates, channelConfig)
					continue
				}
			}

			errorCount++
			errorMessages = append(errorMessages, channelConfig)
			continue
		}

		successCount++
	}

	debug++
	utils.WriteLog(fmt.Sprintf("%s [INSERT] TOTAL_INSERTED: %d; TOTAL_SUCCESS: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, insertedCount, successCount, errorCount, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

	//write error to file
	if len(errorMessages) > 0 {
		filename := fmt.Sprintf("%s_%s", appName, time.Now().Format("2006_01_02"))
		utils.WriteErrorMap(filename, errorMessages)
	}
	if len(errorDuplicates) > 0 {
		filename := fmt.Sprintf("%s_%s_duplicate", appName, time.Now().Format("2006_01_02"))
		utils.WriteErrorMap(filename, errorDuplicates)
	}

	utils.WriteLog(fmt.Sprintf("%s end; duration: %v", logPrefix, time.Now().Sub(tStart)), utils.LogLevelDebug)

}
