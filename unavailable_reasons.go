package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/joho/godotenv/autoload"
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
	scDB := utils.GetDBConnection(os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_PORT"), os.Getenv("USERNAME_DB"), os.Getenv("DATABASE_NAME"), os.Getenv("PASSWORD_DB"))
	defer func(scDB *gorm.DB) {
		err := scDB.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(scDB)

	// Create destination DB connection
	dstDB := utils.GetDBConnection(os.Getenv("RE_DATABASE_HOST"), os.Getenv("RE_DATABASE_PORT"), os.Getenv("RE_USERNAME_DB"), os.Getenv("RE_DATABASE_NAME"), os.Getenv("RE_PASSWORD_DB"))
	defer func(dstDB *gorm.DB) {
		err := dstDB.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(dstDB)

	logID := uuid.NewV4()
	logPrefix := fmt.Sprintf("[%v] [unavailable_reason]", logID)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	//Fetch the database

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

	var list []models.MtrReasonAvailability
	if err := scDB.Find(&list).Error; err != nil {
		utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
		return
	}
	debug++
	utils.WriteLog(fmt.Sprintf("%s [FETCH] DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
	debugT = time.Now()

	//Insert into the new database

	utils.WriteLog(fmt.Sprintf("%s end; duration: %v", logPrefix, time.Now().Sub(tStart)), utils.LogLevelDebug)
}
