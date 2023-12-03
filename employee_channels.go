package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
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
	logPrefix := fmt.Sprintf("[%v] [Employee Channel]", logID)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	var employeeChannelSc []models.EmployeeChannelExist

	qry := `select ec.id, u.id as user_id, 
			case when c.name = 'widget' then 'web' else c.name end as name,
			u.company_id, ec.created_at, ec.updated_at 
			from employee_channels ec 
			join employees e on ec.employee_id = e.id 
			join users u on e.user_id = u.id 
			join channels c on ec.channel_id = c.id 
			where 1=1 and ec.deleted_at is null`

	if os.Getenv("COMPANYID") != "" {
		qry += fmt.Sprintf(" AND u.company_id = '%v'", os.Getenv("COMPANYID"))
	}

	//Fetch companies existing
	if err := scDB.Raw(qry).Scan(&employeeChannelSc).Error; err != nil {
		utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
		return
	}

	totalCompany := len(employeeChannelSc)

	debug++
	utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalCompany, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
	debugT = time.Now()

	insertedCount := 0
	successCount := 0
	errorCount := 0
	var errorMessages []string

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

		insertedCount++
		reiInsertCount := 0
	reInsert:
		if err := dstDB.Create(&employeeChannelDst).Error; err != nil {
			if errCode, ok := err.(*pq.Error); ok {
				if errCode.Code == "23505" { //unique_violation
					reiInsertCount++
					employeeChannelDst.Id = uuid.NewV4()
					if reiInsertCount < 3 {
						goto reInsert
					}
				}
			}
			utils.WriteLog(fmt.Sprintf("%s; [FAILED] [INSERT] Error: %v", logPrefix, err), utils.LogLevelError)
			errorCount++
			errorMessages = append(errorMessages, fmt.Sprintf("%s [FAILED ][INSERT] Error: %v ; DATA: %v", time.Now(), err, employeeChannelDst))
			continue
		}

		successCount++
	}

	// Write error messages to a text file
	formattedTime := time.Now().Format("2006-01-02_150405")
	errorFileLog := fmt.Sprintf("error_messages_employee_channels_%s.log", formattedTime)
	if len(errorMessages) > 0 {
		createFile, errCreate := os.Create(errorFileLog)
		if errCreate != nil {
			utils.WriteLog(fmt.Sprintf("%s [ERROR] Create File Error Log; Error: %v;", logPrefix, errCreate), utils.LogLevelError)
			return
		}
		defer createFile.Close()

		for _, errMsg := range errorMessages {
			_, errWrite := createFile.WriteString(errMsg + "\n")
			if errWrite != nil {
				utils.WriteLog(fmt.Sprintf("%s [ERROR] Write File Error Log; Filename: %s; Error: %v;", logPrefix, errorFileLog, errCreate), utils.LogLevelError)
			}
		}

		utils.WriteLog(fmt.Sprintf("%s [ERROR] Error messages written to %s", logPrefix, errorFileLog), utils.LogLevelError)
	}

	debug++
	utils.WriteLog(fmt.Sprintf("%s [INSERT] TOTAL_INSERTED: %d; TOTAL_SUCCESS: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, insertedCount, successCount, errorCount, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

	utils.WriteLog(fmt.Sprintf("%s end; duration: %v", logPrefix, time.Now().Sub(tStart)), utils.LogLevelDebug)
}
