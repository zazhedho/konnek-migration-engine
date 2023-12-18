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

	dbReport := utils.GetDBReportConnection()
	defer func(dbReport *gorm.DB) {
		err := dbReport.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(dbReport)

	logID := uuid.NewV4()
	appName := "companies"
	if os.Getenv("APP_NAME") != "" {
		appName = os.Getenv("APP_NAME")
	}
	logPrefix := fmt.Sprintf("[%v] [%s]", logID, appName)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	var companiesSc []models.CompanyExist

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
		err = json.Unmarshal(fileContent, &companiesSc)
		if err != nil {
			fmt.Printf("%s Error unmarshalling: %v\n", logPrefix, err)
			utils.WriteLog(fmt.Sprintf("%s Error unmarshalling JSON: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
			return
		}
		debug++
		utils.WriteLog(fmt.Sprintf("%s [GET_FROM_FILE] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, len(companiesSc), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
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
			scDB = scDB.Where("id IN (?)", companyId)
		}

		//Fetch companies existing
		if err := scDB.Find(&companiesSc).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
			return
		}

		totalCompany := len(companiesSc)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalCompany, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()
	}

	insertedCount := 0
	successCount := 0
	errorCount := 0
	var errorMessages []models.CompanyExist
	var errorDuplicates []models.CompanyExist

	for _, company := range companiesSc {
		var companyDst models.CompanyReeng

		companyDst.Id = company.Id
		companyDst.Code = company.CompanyCode
		companyDst.Name = company.Name
		companyDst.Email = company.Email
		companyDst.LimitUser = company.LimitUser
		companyDst.StartPeriod = company.StartPeriod
		companyDst.EndPeriod = company.EndPeriod
		companyDst.Status = company.Status
		companyDst.CreatedAt = company.CreatedAt
		companyDst.CreatedBy = uuid.Nil
		companyDst.UpdatedAt = company.UpdatedAt
		companyDst.UpdatedBy = uuid.Nil
		companyDst.DeletedAt = company.DeletedAt
		companyDst.DeletedBy = uuid.Nil

		insertedCount++
		if err := dstDB.Create(&companyDst).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; [FAILED] [INSERT] Error: %v", logPrefix, err), utils.LogLevelError)

			if errCode, ok := err.(*pq.Error); ok {
				if errCode.Code == "23505" { //unique_violation
					errorDuplicates = append(errorDuplicates, company)
					continue
				}
			}
			errorCount++
			errorMessages = append(errorMessages, company)
			continue
		}

		successCount++

		if err := dbReport.Exec(getTableStructureChatReport(company.Id.String())).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; create table report Error: %+v;", logPrefix, err), utils.LogLevelError)
		}
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

func getTableStructureChatReport(companyId string) string {
	return `
CREATE TABLE "` + companyId + `_session" (
  "id" uuid NOT NULL,
  "company_id" uuid,
  "company_code" varchar(25),
  "company_name" varchar(50),
  "customer_id" uuid,
  "customer_username" varchar(100),
  "customer_name" varchar(100),
  "customer_tags" text,
  "channel" varchar(10),
  "room_id" uuid,
  "division_id" uuid,
  "division_name" varchar(100),
  "agent_id" uuid,
  "agent_username" varchar(100),
  "agent_name" varchar(100),
  "categories" text,
  "bot_status" bool,
  "status" int2,
  "open_time" timestamptz,
  "queue_time" timestamptz,
  "assign_time" timestamptz,
  "fr_time" timestamptz,
  "lr_time" timestamptz,
  "close_time" timestamptz,
  "waiting_duration" int8,
  "fr_duration" int8,
  "resolve_duration" int8,
  "session_duration" int8,
  "sla_from" varchar(25),
  "sla_to" varchar(25),
  "sla_threshold" int2,
  "sla_duration" int8,
  "sla_status" varchar(10),
  "open_by" uuid,
  "open_username" varchar(100),
  "open_name" varchar(100),
  "handover_by" uuid,
  "handover_username" varchar(100),
  "handover_name" varchar(100),
  "close_by" uuid,
  "close_username" varchar(100),
  "close_name" varchar(100),
  "last_update" timestamptz NOT NULL DEFAULT now(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "created_by" varchar(100),
  "updated_at" timestamptz,
  "updated_by" varchar(100),
  PRIMARY KEY ("id")
);

CREATE INDEX "idx_` + companyId + `_session_company_id" ON "` + companyId + `_session" (
  "company_id"
);

CREATE INDEX "idx_` + companyId + `_session_customer_id" ON "` + companyId + `_session" (
  "customer_id"
);

CREATE INDEX "idx_` + companyId + `_session_customer_name" ON "` + companyId + `_session" (
  "company_name"
);

CREATE INDEX "idx_` + companyId + `_session_channel" ON "` + companyId + `_session" (
  "channel"
);

CREATE INDEX "idx_` + companyId + `_session_agent_name" ON "` + companyId + `_session" (
  "agent_name"
);

CREATE INDEX "idx_` + companyId + `_session_status" ON "` + companyId + `_session" (
  "status"
);

CREATE INDEX "idx_` + companyId + `_session_categoories" ON "` + companyId + `_session" (
  "categories"
);

CREATE INDEX "idx_` + companyId + `_session_sla_status" ON "` + companyId + `_session" (
  "sla_status"
);

CREATE INDEX "idx_` + companyId + `_session_open_time" ON "` + companyId + `_session" (
  "open_time"
);

CREATE INDEX "idx_` + companyId + `_session_last_update" ON "` + companyId + `_session" (
  "last_update"
);

COMMENT ON COLUMN "` + companyId + `_session"."id" IS 'sessionId';
COMMENT ON COLUMN "` + companyId + `_session"."waiting_duration" IS 'assign - queue';
COMMENT ON COLUMN "` + companyId + `_session"."fr_duration" IS 'fr - assign';
COMMENT ON COLUMN "` + companyId + `_session"."resolve_duration" IS 'close - assign';
COMMENT ON COLUMN "` + companyId + `_session"."session_duration" IS 'close - open';

CREATE TABLE "` + companyId + `_message" (
  "id" uuid NOT NULL,
  "message_id" varchar(36),
  "reply_id" varchar(36),
  "room_id" uuid,
  "session_id" uuid,
  "from_type" char(1),
  "user_id" uuid,
  "username" varchar(100),
  "user_fullname" varchar(100),
  "type" varchar(20),
  "text" text,
  "payload" text,
  "status" int2,
  "message_time" timestamptz,
  "last_update" timestamptz NOT NULL DEFAULT now(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "created_by" varchar(100),
  "updated_at" timestamptz,
  "updated_by" varchar(100),
  PRIMARY KEY ("id")
);

CREATE INDEX "idx_` + companyId + `_message_message_id" ON "` + companyId + `_message" (
  "message_id"
);

CREATE INDEX "idx_` + companyId + `_message_room_id" ON "` + companyId + `_message" (
  "room_id"
);

CREATE INDEX "idx_` + companyId + `_message_session_id" ON "` + companyId + `_message" (
  "session_id"
);

CREATE INDEX "idx_` + companyId + `_message_text" ON "` + companyId + `_message" (
  "text"
);

CREATE INDEX "idx_` + companyId + `_message_message_time" ON "` + companyId + `_message" (
  "message_time"
);

CREATE INDEX "idx_` + companyId + `_message_last_update" ON "` + companyId + `_message" (
  "last_update"
);


CREATE TABLE "` + companyId + `_summary_hourly_perchannel" (
  "datetime" varchar(16) NOT NULL,
  "channel" varchar(20),
  "open" int8 NOT NULL DEFAULT 0,
  "waiting" int8 NOT NULL DEFAULT 0,
  "assigned" int8 NOT NULL DEFAULT 0,
  "handover" int8 NOT NULL DEFAULT 0,
  "close" int8 NOT NULL DEFAULT 0,
  "total" int8 NOT NULL DEFAULT 0,
  "sla_success" int8 NOT NULL DEFAULT 0,
  "sla_fail" int8 NOT NULL DEFAULT 0,
  "waiting_duration" int8 NOT NULL DEFAULT 0,
  "fr_duration" int8 NOT NULL DEFAULT 0,
  "resolve_duration" int8 NOT NULL DEFAULT 0,
  "session_duration" int8 NOT NULL DEFAULT 0,
  "last_update" timestamptz NOT NULL DEFAULT now(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "created_by" varchar(100),
  "updated_at" timestamptz,
  "updated_by" varchar(100),
  PRIMARY KEY ("datetime", "channel")
);


CREATE TABLE "` + companyId + `_summary_daily_perchannel" (
  "date" date NOT NULL,
  "channel" varchar(20),
  "open" int8 NOT NULL DEFAULT 0,
  "waiting" int8 NOT NULL DEFAULT 0,
  "assigned" int8 NOT NULL DEFAULT 0,
  "handover" int8 NOT NULL DEFAULT 0,
  "close" int8 NOT NULL DEFAULT 0,
  "total" int8 NOT NULL DEFAULT 0,
  "sla_success" int8 NOT NULL DEFAULT 0,
  "sla_fail" int8 NOT NULL DEFAULT 0,
  "waiting_duration" int8 NOT NULL DEFAULT 0,
  "fr_duration" int8 NOT NULL DEFAULT 0,
  "resolve_duration" int8 NOT NULL DEFAULT 0,
  "session_duration" int8 NOT NULL DEFAULT 0,
  "last_update" timestamptz NOT NULL DEFAULT now(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "created_by" varchar(100),
  "updated_at" timestamptz,
  "updated_by" varchar(100),
  PRIMARY KEY ("date", "channel")
);


CREATE TABLE "` + companyId + `_summary_perchannel" (
  "channel" varchar(20),
  "open" int8 NOT NULL DEFAULT 0,
  "waiting" int8 NOT NULL DEFAULT 0,
  "assigned" int8 NOT NULL DEFAULT 0,
  "handover" int8 NOT NULL DEFAULT 0,
  "close" int8 NOT NULL DEFAULT 0,
  "total" int8 NOT NULL DEFAULT 0,
  "sla_success" int8 NOT NULL DEFAULT 0,
  "sla_fail" int8 NOT NULL DEFAULT 0,
  "waiting_duration" int8 NOT NULL DEFAULT 0,
  "fr_duration" int8 NOT NULL DEFAULT 0,
  "resolve_duration" int8 NOT NULL DEFAULT 0,
  "session_duration" int8 NOT NULL DEFAULT 0,
  "last_update" timestamptz NOT NULL DEFAULT now(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "created_by" varchar(100),
  "updated_at" timestamptz,
  "updated_by" varchar(100),
  PRIMARY KEY ("channel")
);


CREATE TABLE "` + companyId + `_summary_daily_percustomer" (
  "date" date NOT NULL,
  "customer_id" uuid NOT NULL,
  "customer_username" varchar(100),
  "customer_name" varchar(100),
  "customer_tags" text,
  "channel" varchar(20),
  "open" int8 NOT NULL DEFAULT 0,
  "waiting" int8 NOT NULL DEFAULT 0,
  "assigned" int8 NOT NULL DEFAULT 0,
  "handover" int8 NOT NULL DEFAULT 0,
  "close" int8 NOT NULL DEFAULT 0,
  "total" int8 NOT NULL DEFAULT 0,
  "sla_success" int8 NOT NULL DEFAULT 0,
  "sla_fail" int8 NOT NULL DEFAULT 0,
  "waiting_duration" int8 NOT NULL DEFAULT 0,
  "fr_duration" int8 NOT NULL DEFAULT 0,
  "resolve_duration" int8 NOT NULL DEFAULT 0,
  "session_duration" int8 NOT NULL DEFAULT 0,
  "last_update" timestamptz NOT NULL DEFAULT now(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "created_by" varchar(100),
  "updated_at" timestamptz,
  "updated_by" varchar(100),
  PRIMARY KEY ("date", "customer_id")
);


CREATE TABLE "` + companyId + `_summary_percustomer" (
  "customer_id" uuid NOT NULL,
  "customer_username" varchar(100),
  "customer_name" varchar(100),
  "customer_tags" text,
  "channel" varchar(20),
  "open" int8 NOT NULL DEFAULT 0,
  "waiting" int8 NOT NULL DEFAULT 0,
  "assigned" int8 NOT NULL DEFAULT 0,
  "handover" int8 NOT NULL DEFAULT 0,
  "close" int8 NOT NULL DEFAULT 0,
  "total" int8 NOT NULL DEFAULT 0,
  "sla_success" int8 NOT NULL DEFAULT 0,
  "sla_fail" int8 NOT NULL DEFAULT 0,
  "waiting_duration" int8 NOT NULL DEFAULT 0,
  "fr_duration" int8 NOT NULL DEFAULT 0,
  "resolve_duration" int8 NOT NULL DEFAULT 0,
  "session_duration" int8 NOT NULL DEFAULT 0,
  "last_update" timestamptz NOT NULL DEFAULT now(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "created_by" varchar(100),
  "updated_at" timestamptz,
  "updated_by" varchar(100),
  PRIMARY KEY ("customer_id")
);


CREATE TABLE "` + companyId + `_summary_daily_peragent" (
  "date" date NOT NULL,
  "agent_id" uuid NOT NULL,
  "agent_username" varchar(100),
  "agent_name" varchar(100),
  "open" int8 NOT NULL DEFAULT 0,
  "waiting" int8 NOT NULL DEFAULT 0,
  "assigned" int8 NOT NULL DEFAULT 0,
  "handover" int8 NOT NULL DEFAULT 0,
  "close" int8 NOT NULL DEFAULT 0,
  "total" int8 NOT NULL DEFAULT 0,
  "sla_success" int8 NOT NULL DEFAULT 0,
  "sla_fail" int8 NOT NULL DEFAULT 0,
  "waiting_duration" int8 NOT NULL DEFAULT 0,
  "fr_duration" int8 NOT NULL DEFAULT 0,
  "resolve_duration" int8 NOT NULL DEFAULT 0,
  "session_duration" int8 NOT NULL DEFAULT 0,
  "last_update" timestamptz NOT NULL DEFAULT now(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "created_by" varchar(100),
  "updated_at" timestamptz,
  "updated_by" varchar(100),
  PRIMARY KEY ("date", "agent_id")
);


CREATE TABLE "` + companyId + `_summary_peragent" (
  "agent_id" uuid NOT NULL,
  "agent_username" varchar(100),
  "agent_name" varchar(100),
  "open" int8 NOT NULL DEFAULT 0,
  "waiting" int8 NOT NULL DEFAULT 0,
  "assigned" int8 NOT NULL DEFAULT 0,
  "handover" int8 NOT NULL DEFAULT 0,
  "close" int8 NOT NULL DEFAULT 0,
  "total" int8 NOT NULL DEFAULT 0,
  "sla_success" int8 NOT NULL DEFAULT 0,
  "sla_fail" int8 NOT NULL DEFAULT 0,
  "waiting_duration" int8 NOT NULL DEFAULT 0,
  "fr_duration" int8 NOT NULL DEFAULT 0,
  "resolve_duration" int8 NOT NULL DEFAULT 0,
  "session_duration" int8 NOT NULL DEFAULT 0,
  "last_update" timestamptz NOT NULL DEFAULT now(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "created_by" varchar(100),
  "updated_at" timestamptz,
  "updated_by" varchar(100),
  PRIMARY KEY ("agent_id")
);
`
}
