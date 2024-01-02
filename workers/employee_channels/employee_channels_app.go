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
	appName := "employee_channels"
	if os.Getenv("APP_NAME") != "" {
		appName = os.Getenv("APP_NAME")
	}
	logPrefix := fmt.Sprintf("[%v] [%s]", logID, appName)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	var employeeChannelSc []models.EmployeeChannelExist

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
		err = json.Unmarshal(fileContent, &employeeChannelSc)
		if err != nil {
			fmt.Printf("%s Error unmarshalling: %v\n", logPrefix, err)
			utils.WriteLog(fmt.Sprintf("%s Error unmarshalling JSON: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
			return
		}
		debug++
		utils.WriteLog(fmt.Sprintf("%s [GET_FROM_FILE] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, len(employeeChannelSc), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()

		err = os.Remove("../../data/" + os.Getenv("GET_FROM_FILE"))
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s Error Delete file: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
		}
	} else {
		//Fetch from database

		qry := `select ec.id, u.id as user_id, 
			case when c.name = 'widget' then 'web' else c.name end as name,
			u.company_id, ec.created_at, ec.updated_at, ec.deleted_at 
			from employee_channels ec 
			join employees e on ec.employee_id = e.id 
			join users u on e.user_id = u.id 
			join channels c on ec.channel_id = c.id
			WHERE 1=1 AND u.company_id IN (?)`

		//if os.Getenv("COMPANYID") != "" {
		//	companyId := strings.Split(os.Getenv("COMPANYID"), ",")
		//	//scDB = scDB.Where("company_id IN (?)", companyId)
		//	qry += fmt.Sprintf(" AND u.company_id IN (%v)", companyId)
		//}

		//Fetch companies existing
		companyId := strings.Split(os.Getenv("COMPANYID"), ",")
		if err := scDB.Raw(qry, companyId).Scan(&employeeChannelSc).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
			return
		}

		totalEmpChannel := len(employeeChannelSc)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalEmpChannel, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()
	}

	insertedCount := 0
	successCount := 0
	errorCount := 0
	var errorMessages []models.EmployeeChannelExist
	var errorDuplicates []models.EmployeeChannelExist

	for _, employeeChannel := range employeeChannelSc {
		var employeeChannelDst models.EmployeeChannelReeng
		employeeChannelDst.Id = employeeChannel.Id
		employeeChannelDst.UserId = employeeChannel.UserID
		employeeChannelDst.ChannelCode = employeeChannel.ChannelName
		employeeChannelDst.CompanyId = employeeChannel.CompanyID
		employeeChannelDst.CreatedAt = employeeChannel.CreatedAt
		employeeChannelDst.CreatedBy = uuid.Nil
		employeeChannelDst.UpdatedAt = employeeChannel.UpdatedAt
		employeeChannelDst.UpdatedBy = uuid.Nil
		employeeChannelDst.DeletedAt = employeeChannel.DeletedAt

		insertedCount++
		if err := dstDB.Create(&employeeChannelDst).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; [FAILED] [INSERT] Error: %v", logPrefix, err), utils.LogLevelError)

			if errCode, ok := err.(*pq.Error); ok {
				if errCode.Code == "23505" { //unique_violation
					errorDuplicates = append(errorDuplicates, employeeChannel)
					continue
				}
			}
			errorCount++
			errorMessages = append(errorMessages, employeeChannel)
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
