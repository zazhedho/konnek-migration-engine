package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"konnek-migration/models"
	"konnek-migration/sdk"
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
	appName := "company_configs"
	if os.Getenv("APP_NAME") != "" {
		appName = os.Getenv("APP_NAME")
	}
	logPrefix := fmt.Sprintf("[%v] [%s]", logID, appName)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	var comConfSc []models.Configuration

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
		err = json.Unmarshal(fileContent, &comConfSc)
		if err != nil {
			fmt.Printf("%s Error unmarshalling: %v\n", logPrefix, err)
			utils.WriteLog(fmt.Sprintf("%s Error unmarshalling JSON: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
			return
		}
		debug++
		utils.WriteLog(fmt.Sprintf("%s [GET_FROM_FILE] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, len(comConfSc), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()

		err = os.Remove("../../data/" + os.Getenv("GET_FROM_FILE"))
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s Error Delete file: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
		}
	} else {
		//Fetch from database
		scDB = scDB.Unscoped()
		if os.Getenv("COMPANYID") != "" {
			companyId := strings.Split(os.Getenv("COMPANYID"), ",")
			scDB = scDB.Where("company_id IN (?)", companyId)
		}

		//Fetch companies existing
		if err := scDB.Find(&comConfSc).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
			return
		}

		totalConfigs := len(comConfSc)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalConfigs, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()
	}

	insertedCount := 0
	successCount := 0
	errorCount := 0

	var errorMessages []models.Configuration
	var errorDuplicates []models.Configuration

	for _, comConf := range comConfSc {
		var comConfDst models.CompanyConfig

		comConfDst.Id = comConf.Id
		comConfDst.CompanyId = comConf.CompanyId
		comConfDst.Bot = comConf.Bot
		comConfDst.Whatsapp = comConf.Whatsapp
		comConfDst.Line = comConf.Line
		comConfDst.Telegram = comConf.Telegram
		comConfDst.Facebook = comConf.FacebookMessenger
		comConfDst.WebWidget = comConf.Widget
		comConfDst.AutoAssign = comConf.AutoAssign
		comConfDst.ChatLimit = comConf.ChatLimit
		comConfDst.SlaFrom = comConf.SlaFrom
		comConfDst.SlaTo = comConf.SlaTo
		comConfDst.SlaThreshold = comConf.SlaThreshold
		comConfDst.WelcomeGreeting = comConf.Greeting
		comConfDst.WelcomeGreetingFlag = comConf.FlagGreeting
		comConfDst.WelcomeGreetingOption = comConf.GreetingOptions
		comConfDst.WelcomeGreetingOptionFlag = comConf.GreetingOptionsFlag
		comConfDst.WaitingGreeting = comConf.WaitingGreeting
		comConfDst.WaitingGreetingFlag = comConf.WaitingGreetingFlag
		comConfDst.AssignedGreeting = comConf.AssignedGreeting
		comConfDst.AssignedGreetingFlag = comConf.AssignedGreetingFlag
		comConfDst.ClosingGreeting = comConf.ClosingGreeting
		comConfDst.ClosingGreetingFlag = comConf.ClosingGreetingFlag
		comConfDst.CsatFlag = comConf.CsatFlag
		comConfDst.UnavailableReasonFlag = comConf.ReasonFlag
		comConfDst.CreatedAt = comConf.CreatedAt
		comConfDst.CreatedBy = uuid.Nil
		comConfDst.UpdatedAt = comConf.UpdatedAt
		comConfDst.UpdatedBy = uuid.Nil
		comConfDst.DeletedAt = comConf.DeletedAt
		comConfDst.DeletedBy = uuid.Nil

		switch comConf.SdkWhatsapp {
		case sdk.SdkWAKataAiEx:
			comConfDst.SdkWhatsapp = sdk.SdkWhatsappKataAi
		case sdk.SdkWABotikaEx:
			comConfDst.SdkWhatsapp = sdk.SdkWhatsappBotika
		case sdk.SdkWAMetaEx:
			comConfDst.SdkWhatsapp = sdk.SdkWhatsappMeta
		default:
			comConfDst.SdkWhatsapp = sdk.SdkWhatsappMeta
		}

		comConfDst.SdkTelegram = sdk.SdkTelegram
		comConfDst.SdkLine = sdk.SdkLine
		comConfDst.SdkFacebook = sdk.SdkFacebook
		comConfDst.SdkBot = sdk.SdkBotKataAi
		comConfDst.Blacklist = comConf.BlackList
		comConfDst.InquirySandeza = comConf.InquirySandeza
		comConfDst.KeywordFilterStatus = comConf.KeywordFilterStatus
		comConfDst.KeywordFilter = comConf.KeywordFilter
		comConfDst.KeywordGreetings = comConf.KeywordGreetings
		comConfDst.MaintenanceStatus = comConf.MaintenanceStatus
		comConfDst.MaintenanceMessage = comConf.MaintenanceMessage
		comConfDst.KeywordMaxInvalid = comConf.KeywordGreetingsLimit

		KeyInterval := comConf.KeywordGreetingsLimitDuration / 60

		comConfDst.KeywordInterval = KeyInterval
		comConfDst.KeywordBlockDuration = comConf.KeywordGreetingsLimitDuration

		insertedCount++
		if err := dstDB.Create(&comConfDst).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; [INSERT] Error: %v", logPrefix, err), utils.LogLevelError)
			if errCode, ok := err.(*pq.Error); ok {
				if errCode.Code == "23505" { //unique_violation
					errorDuplicates = append(errorDuplicates, comConf)
					continue
				}
			}

			errorCount++
			errorMessages = append(errorMessages, comConf)
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
