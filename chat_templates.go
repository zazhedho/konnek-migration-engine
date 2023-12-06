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
	appName := "chat_templates"
	if os.Getenv("APP_NAME") != "" {
		appName = os.Getenv("APP_NAME")
	}
	logPrefix := fmt.Sprintf("[%v] [%s]", logID, appName)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	var chatTemplSc []models.ChatTemplateExist
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
		err = json.Unmarshal(fileContent, &chatTemplSc)
		if err != nil {
			fmt.Printf("%s Error unmarshalling: %v\n", logPrefix, err)
			utils.WriteLog(fmt.Sprintf("%s Error unmarshalling JSON: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
			return
		}
		debug++
		utils.WriteLog(fmt.Sprintf("%s [GET_FROM_FILE] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, len(chatTemplSc), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()

		err = os.Remove("data/" + os.Getenv("GET_FROM_FILE"))
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s Error Delete file: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
		}
	} else {
		//Fetch from database

		if os.Getenv("COMPANYID") != "" {
			scDB = scDB.Where("company_id = ?", os.Getenv("COMPANYID"))
		}

		//Fetch companies existing
		if err := scDB.Find(&chatTemplSc).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
			return
		}

		totalChatTempl := len(chatTemplSc)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalChatTempl, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()
	}

	insertedCount := 0
	successCount := 0
	errorCount := 0
	var errorMessages []models.ChatTemplateExist
	var errorDuplicates []models.ChatTemplateExist

	for _, chatTemplate := range chatTemplSc {
		var chatTemplDst models.ChatTemplateReeng

		chatTemplDst.Id = chatTemplate.Id
		chatTemplDst.Keyword = chatTemplate.Keyword
		chatTemplDst.Text = chatTemplate.Text
		chatTemplDst.CompanyId = chatTemplate.CompanyId
		chatTemplDst.CreatedAt = chatTemplate.CreatedAt
		chatTemplDst.CreatedBy = chatTemplate.CreatedBy
		chatTemplDst.UpdatedAt = chatTemplate.UpdatedAt
		chatTemplDst.UpdatedBy = uuid.Nil
		chatTemplDst.DeletedAt = chatTemplate.DeletedAt

		insertedCount++
		//reiInsertCount := 0
		//reInsert:
		if err := dstDB.Create(&chatTemplDst).Error; err != nil {
			chatTemplate.Error = err.Error()
			if errCode, ok := err.(*pq.Error); ok {
				if errCode.Code == "23505" { //unique_violation
					errorDuplicates = append(errorDuplicates, chatTemplate)
					continue
				}
			}
			utils.WriteLog(fmt.Sprintf("%s; [FAILED] [INSERT] Error: %v", logPrefix, err), utils.LogLevelError)
			errorCount++
			errorMessages = append(errorMessages, chatTemplate)
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
