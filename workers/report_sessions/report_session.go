package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/joho/godotenv/autoload"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"konnek-migration/models"
	"konnek-migration/utils"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	utils.Init()

	db := utils.GetDBNewConnection()
	defer func(db *gorm.DB) {
		err := db.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(db)

	dbReport := utils.GetDBReportConnection()
	defer func(dbReport *gorm.DB) {
		err := dbReport.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(dbReport)

	logID := uuid.NewV4()
	appName := "report_message"
	if os.Getenv("APP_NAME") != "" {
		appName = os.Getenv("APP_NAME")
	}

	logPrefix := fmt.Sprintf("[%v] [%s]", logID, appName)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	var lists []models.FetchReportSession

	// Get from file
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
		err = json.Unmarshal(fileContent, &lists)
		if err != nil {
			fmt.Printf("%s Error unmarshalling: %v\n", logPrefix, err)
			utils.WriteLog(fmt.Sprintf("%s Error unmarshalling JSON: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
			return
		}
		debug++
		utils.WriteLog(fmt.Sprintf("%s [GET_FROM_FILE] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, len(lists), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()

		err = os.Remove("data/" + os.Getenv("GET_FROM_FILE"))
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s Error Delete file: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
		}
	} else {
		//Fetch from database

		//Set the filters
		if os.Getenv("COMPANYID") != "" {
			db = db.Preload("Room", func(db *gorm.DB) *gorm.DB {
				return db.Where("company_id = ?", os.Getenv("COMPANYID"))
			})
		} else {
			db = db.Preload("Room")
		}

		if os.Getenv("START_DATE") != "" && os.Getenv("END_DATE") != "" {
			db = db.Where("created_at BETWEEN ? AND ?", os.Getenv("START_DATE"), os.Getenv("END_DATE"))
		} else if os.Getenv("START_DATE") != "" && os.Getenv("END_DATE") == "" {
			db = db.Where("created_at >=?", os.Getenv("START_DATE"))
		} else if os.Getenv("START_DATE") == "" && os.Getenv("END_DATE") != "" {
			db = db.Where("created_at <=?", os.Getenv("END_DATE"))
		}

		if os.Getenv("ORDER_BY") != "" {
			sortMap := map[string]string{
				"created_at": "created_at",
			}
			if strings.ToUpper(os.Getenv("ORDER_DIRECTION")) == "DESC" {
				db = db.Order(sortMap[os.Getenv("ORDER_BY")] + " DESC")
			} else {
				db = db.Order(sortMap[os.Getenv("ORDER_BY")])
			}
		}

		offset, _ := strconv.Atoi(os.Getenv("OFFSET"))
		limit, _ := strconv.Atoi(os.Getenv("LIMIT"))

		if offset > 0 {
			db = db.Offset(offset)
		}

		if limit > 0 {
			db = db.Limit(limit)
		}

		if err := db.Preload("Division").Preload("Agent").Find(&lists).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
			return
		}
		totalFetch := len(lists)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()
	}

	//Insert into database report
	var errorMessages []models.FetchReportSession
	var errorDuplicates []models.FetchReportSession
	totalInserted := 0

	for _, list := range lists {
		var m models.ReportSession
		m.TablePrefix = list.Room.CompanyId.String() + "_"
		//check table, create table if it doesn't exist
		//dbReport.AutoMigrate(&m)

		var waitingDuration int64
		var frDuration int64
		var resolveDuration int64
		var sessionDuration int64

		if list.QueTime != nil && list.AssignTime != nil {
			waitingDuration = int64(list.QueTime.Sub(*list.AssignTime).Seconds())
		}
		if list.AssignTime != nil && list.FirstResponseTime != nil {
			frDuration = int64(list.AssignTime.Sub(*list.FirstResponseTime).Seconds())
		}
		if list.AssignTime != nil && list.CloseTime != nil {
			resolveDuration = int64(list.AssignTime.Sub(*list.CloseTime).Seconds())
		}
		if list.OpenTime != nil && list.CloseTime != nil {
			sessionDuration = int64(list.OpenTime.Sub(*list.CloseTime).Seconds())
		}

		m.Id = list.Id
		m.CompanyId = list.Room.CompanyId
		m.CompanyName = list.Room.Company.Name
		m.CompanyCode = list.Room.Company.Code
		m.CustomerId = list.Room.Customer.Id
		m.CustomerUsername = list.Room.Customer.Username
		m.CustomerName = list.Room.Customer.Name
		m.CustomerTags = list.Room.Customer.Tags
		m.Channel = list.Room.ChannelCode
		m.RoomId = list.RoomId
		m.DivisionId = list.DivisionId
		m.DivisionName = list.Division.Name
		m.AgentUserId = list.AgentUserId
		m.AgentUsername = list.Agent.Username
		m.AgentName = list.Agent.Name
		m.Categories = list.Categories
		m.BotStatus = list.BotStatus
		m.Status = list.Status
		m.OpenTime = list.OpenTime
		m.QueTime = list.QueTime
		m.AssignTime = list.AssignTime
		m.FrTime = list.FirstResponseTime
		m.LrTime = list.LastAgentChatTime
		m.CloseTime = list.CloseTime
		m.WaitingDuration = waitingDuration
		m.FrDuration = frDuration
		m.ResolveDuration = resolveDuration
		m.SessionDuration = sessionDuration
		m.SlaFrom = list.SlaFrom
		m.SlaTo = list.SlaTo
		m.SlaThreshold = int64(list.SlaTreshold)
		m.SlaDuration = int64(list.SlaDurations)
		m.SlaStatus = list.SlaStatus
		m.OpenBy = list.OpenBy
		m.OpenUsername = list.UserOpenBy.Username
		m.OpenName = list.UserOpenBy.Name
		m.HandoverBy = list.HandoverBy
		m.HandoverUsername = list.UserHandoverBy.Username
		m.HandoverName = list.UserHandoverBy.Name
		m.CloseBy = list.CloseBy
		m.CloseUsername = list.UserCloseBy.Username
		m.CloseName = list.UserCloseBy.Name
		m.LastUpdate = time.Now()
		m.CreatedAt = time.Now()
		m.UpdatedAt = time.Now()
		m.CreatedBy = "migration-engine"
		m.UpdatedBy = "migration-engine"

		if err := dbReport.Create(&m).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; insert error: %v", logPrefix, err), utils.LogLevelError)
			list.Error = err.Error()
			if errCode, ok := err.(*pq.Error); ok {
				if errCode.Code == "23505" { //unique_violation
					errorDuplicates = append(errorDuplicates, list)
					continue
				}
			}
			errorMessages = append(errorMessages, list)
			continue
		}
		totalInserted++

		//summary report
		updateSummary := make(map[string]interface{})

		updateSummary["last_update"] = m.LastUpdate.Format(time.RFC3339Nano)
		updateSummary["updated_at"] = time.Now().Format(time.RFC3339Nano)
		updateSummary["updated_by"] = "migration-engine"
		switch m.Status {
		case models.SessionOpen:
			updateSummary["open"] = 1
		case models.SessionWaiting:
			updateSummary["waiting"] = 1
		case models.SessionAssigned:
			updateSummary["assigned"] = 1
		case models.SessionHandovered:
			updateSummary["handover"] = 1
		case models.SessionClosed:
			updateSummary["close"] = 1
		}
		updateSummary["total"] = 1

		switch m.SlaStatus {
		case models.SlaSuccess:
			updateSummary["sla_success"] = 1
		case models.SlaFailed:
			updateSummary["sla_fail"] = 1
		}

		updateSummary["waiting_duration"] = m.WaitingDuration
		updateSummary["fr_duration"] = m.FrDuration
		updateSummary["resolve_duration"] = m.ResolveDuration
		updateSummary["session_duration"] = m.SessionDuration

		summarySets := make([]string, 0)
		summaryFields := make([]string, 0)
		summaryVals := make([]string, 0)
		for k, v := range updateSummary {
			summaryFields = append(summaryFields, k)

			switch k {
			case "last_update", "updated_at", "updated_by":
				summarySets = append(summarySets, fmt.Sprintf("%s = '%s'", k, v))
				summaryVals = append(summaryVals, fmt.Sprintf("'%s'", v))
			default:
				summarySets = append(summarySets, fmt.Sprintf("%s = %s+%d", k, k, v))
				summaryVals = append(summaryVals, fmt.Sprintf("'%d'", v))
			}
		}

		//summary hourly per channel
		go func(logPrefix string, summarySets, summaryFields, summaryVals []string) {
			qry := fmt.Sprintf(`UPDATE "%s_summary_hourly_perchannel" SET %s WHERE datetime = '%s' AND channel = '%s';`, m.CompanyId, strings.Join(summarySets, ", "), m.OpenTime.Format(utils.LayoutDateTimeH+":00"), m.Channel)
			if qryRes := dbReport.Exec(qry); qryRes.Error != nil {
				utils.WriteLog(fmt.Sprintf("%s; %s, Error: %+v;", logPrefix, qry, qryRes.Error), utils.LogLevelError)
				utils.WriteToFile(fmt.Sprintf("summary_%s", time.Now().Format("2006_01_02")), qry)
			} else if qryRes.RowsAffected == 0 {
				fields := append(summaryFields, "datetime", "channel", "created_at", "created_by")
				vals := append(summaryVals,
					fmt.Sprintf("'%s'", m.OpenTime.Format(utils.LayoutDateTimeH+":00")),
					fmt.Sprintf("'%s'", m.Channel),
					fmt.Sprintf("'%s'", time.Now().Format(time.RFC3339Nano)),
					fmt.Sprintf("'%s'", "migration-engine"),
				)

				qry = fmt.Sprintf(`INSERT INTO "%s_summary_hourly_perchannel" (%s) VALUES (%s);`, m.CompanyId, strings.Join(fields, ", "), strings.Join(vals, ", "))
				if qryRes = dbReport.Exec(qry); qryRes.Error != nil {
					utils.WriteLog(fmt.Sprintf("%s; %s, Error: %+v;", logPrefix, qry, qryRes.Error), utils.LogLevelError)
					utils.WriteToFile(fmt.Sprintf("summary_%s", time.Now().Format("2006_01_02")), qry)
				}
			}
		}(logPrefix, summarySets, summaryFields, summaryVals)

		//summary daily per channel
		go func(logPrefix string, summarySets, summaryFields, summaryVals []string) {
			qry := fmt.Sprintf(`UPDATE "%s_summary_daily_perchannel" SET %s WHERE date = '%s' AND channel = '%s';`, m.CompanyId, strings.Join(summarySets, ", "), m.OpenTime.Format(utils.LayoutDate), m.Channel)
			if qryRes := dbReport.Exec(qry); qryRes.Error != nil {
				utils.WriteLog(fmt.Sprintf("%s; %s, Error: %+v;", logPrefix, qry, qryRes.Error), utils.LogLevelError)
				utils.WriteToFile(fmt.Sprintf("summary_%s", time.Now().Format("2006_01_02")), qry)
			} else if qryRes.RowsAffected == 0 {
				fields := append(summaryFields, "date", "channel", "created_at", "created_by")
				vals := append(summaryVals,
					fmt.Sprintf("'%s'", m.OpenTime.Format(utils.LayoutDate)),
					fmt.Sprintf("'%s'", m.Channel),
					fmt.Sprintf("'%s'", time.Now().Format(time.RFC3339Nano)),
					fmt.Sprintf("'%s'", "migration-engine"),
				)

				qry = fmt.Sprintf(`INSERT INTO "%s_summary_daily_perchannel" (%s) VALUES (%s);`, m.CompanyId, strings.Join(fields, ", "), strings.Join(vals, ", "))
				if qryRes = dbReport.Exec(qry); qryRes.Error != nil {
					utils.WriteLog(fmt.Sprintf("%s; %s, Error: %+v;", logPrefix, qry, qryRes.Error), utils.LogLevelError)
					utils.WriteToFile(fmt.Sprintf("summary_%s", time.Now().Format("2006_01_02")), qry)
				}
			}
		}(logPrefix, summarySets, summaryFields, summaryVals)

		//summary per channel
		go func(logPrefix string, summarySets, summaryFields, summaryVals []string) {
			qry := fmt.Sprintf(`UPDATE "%s_summary_perchannel" SET %s WHERE channel = '%s';`, m.CompanyId, strings.Join(summarySets, ", "), m.Channel)
			if qryRes := dbReport.Exec(qry); qryRes.Error != nil {
				utils.WriteLog(fmt.Sprintf("%s; %s, Error: %+v;", logPrefix, qry, qryRes.Error), utils.LogLevelError)
				utils.WriteToFile(fmt.Sprintf("summary_%s", time.Now().Format("2006_01_02")), qry)
			} else if qryRes.RowsAffected == 0 {
				fields := append(summaryFields, "channel", "created_at", "created_by")
				vals := append(summaryVals,
					fmt.Sprintf("'%s'", m.Channel),
					fmt.Sprintf("'%s'", time.Now().Format(time.RFC3339Nano)),
					fmt.Sprintf("'%s'", "migration-engine"),
				)

				qry = fmt.Sprintf(`INSERT INTO "%s_summary_perchannel" (%s) VALUES (%s);`, m.CompanyId, strings.Join(fields, ", "), strings.Join(vals, ", "))
				if qryRes = dbReport.Exec(qry); qryRes.Error != nil {
					utils.WriteLog(fmt.Sprintf("%s; %s, Error: %+v;", logPrefix, qry, qryRes.Error), utils.LogLevelError)
					utils.WriteToFile(fmt.Sprintf("summary_%s", time.Now().Format("2006_01_02")), qry)
				}
			}
		}(logPrefix, summarySets, summaryFields, summaryVals)

		//summary daily per customer
		go func(logPrefix string, summarySets, summaryFields, summaryVals []string) {
			qry := fmt.Sprintf(`UPDATE "%s_summary_daily_percustomer" SET %s WHERE date = '%s' AND customer_id = '%s';`, m.CompanyId, strings.Join(summarySets, ", "), m.OpenTime.Format(utils.LayoutDate), m.CustomerId)
			if qryRes := dbReport.Exec(qry); qryRes.Error != nil {
				utils.WriteLog(fmt.Sprintf("%s; %s, Error: %+v;", logPrefix, qry, qryRes.Error), utils.LogLevelError)
				utils.WriteToFile(fmt.Sprintf("summary_%s", time.Now().Format("2006_01_02")), qry)
			} else if qryRes.RowsAffected == 0 {
				fields := append(summaryFields, "date", "channel", "customer_id", "customer_username", "customer_name", "customer_tags", "created_at", "created_by")
				vals := append(summaryVals,
					fmt.Sprintf("'%s'", m.OpenTime.Format(utils.LayoutDate)),
					fmt.Sprintf("'%s'", m.Channel),
					fmt.Sprintf("'%s'", m.CustomerId),
					fmt.Sprintf("'%s'", m.CustomerUsername),
					fmt.Sprintf("'%s'", m.CustomerName),
					fmt.Sprintf("'%s'", m.CustomerTags),
					fmt.Sprintf("'%s'", time.Now().Format(time.RFC3339Nano)),
					fmt.Sprintf("'%s'", "migration-engine"),
				)

				qry = fmt.Sprintf(`INSERT INTO "%s_summary_daily_percustomer" (%s) VALUES (%s);`, m.CompanyId, strings.Join(fields, ", "), strings.Join(vals, ", "))
				if qryRes = dbReport.Exec(qry); qryRes.Error != nil {
					utils.WriteLog(fmt.Sprintf("%s; %s, Error: %+v;", logPrefix, qry, qryRes.Error), utils.LogLevelError)
					utils.WriteToFile(fmt.Sprintf("summary_%s", time.Now().Format("2006_01_02")), qry)
				}
			}
		}(logPrefix, summarySets, summaryFields, summaryVals)

		//summary per customer
		go func(logPrefix string, summarySets, summaryFields, summaryVals []string) {
			qry := fmt.Sprintf(`UPDATE "%s_summary_percustomer" SET %s WHERE customer_id = '%s';`, m.CompanyId, strings.Join(summarySets, ", "), m.CustomerId)
			if qryRes := dbReport.Exec(qry); qryRes.Error != nil {
				utils.WriteLog(fmt.Sprintf("%s; %s, Error: %+v;", logPrefix, qry, qryRes.Error), utils.LogLevelError)
				utils.WriteToFile(fmt.Sprintf("summary_%s", time.Now().Format("2006_01_02")), qry)
			} else if qryRes.RowsAffected == 0 {
				fields := append(summaryFields, "channel", "customer_id", "customer_username", "customer_name", "customer_tags", "created_at", "created_by")
				vals := append(summaryVals,
					fmt.Sprintf("'%s'", m.Channel),
					fmt.Sprintf("'%s'", m.CustomerId),
					fmt.Sprintf("'%s'", m.CustomerUsername),
					fmt.Sprintf("'%s'", m.CustomerName),
					fmt.Sprintf("'%s'", m.CustomerTags),
					fmt.Sprintf("'%s'", time.Now().Format(time.RFC3339Nano)),
					fmt.Sprintf("'%s'", "migration-engine"),
				)

				qry = fmt.Sprintf(`INSERT INTO "%s_summary_percustomer" (%s) VALUES (%s);`, m.CompanyId, strings.Join(fields, ", "), strings.Join(vals, ", "))
				if qryRes = dbReport.Exec(qry); qryRes.Error != nil {
					utils.WriteLog(fmt.Sprintf("%s; %s, Error: %+v;", logPrefix, qry, qryRes.Error), utils.LogLevelError)
					utils.WriteToFile(fmt.Sprintf("summary_%s", time.Now().Format("2006_01_02")), qry)
				}
			}
		}(logPrefix, summarySets, summaryFields, summaryVals)

		if m.AgentUserId != uuid.Nil {
			summaryAgentSets := make([]string, 0)
			summaryAgentFields := make([]string, 0)
			summaryAgentVals := make([]string, 0)
			_, totalIsSet := updateSummary["total"]

			//remove open/waiting (asumsi agent tidak akan memiliki room open/waiting)
			for j := 0; j < len(summaryFields); j++ {
				if summaryFields[j] == "open" || summaryFields[j] == "waiting" {
					continue
				}

				if !totalIsSet && summaryFields[j] == "assigned" && summaryVals[j] == "'1'" {
					summaryAgentSets = append(summaryAgentSets, "total=total+1")
					summaryAgentFields = append(summaryAgentFields, "total")
					summaryAgentVals = append(summaryAgentVals, "1")
				}

				summaryAgentSets = append(summaryAgentSets, summarySets[j])
				summaryAgentFields = append(summaryAgentFields, summaryFields[j])
				summaryAgentVals = append(summaryAgentVals, summaryVals[j])
			}

			//summary daily per agent
			go func(logPrefix string, summarySets, summaryFields, summaryVals []string) {
				qry := fmt.Sprintf(`UPDATE "%s_summary_daily_peragent" SET %s WHERE date = '%s' AND agent_id = '%s';`, m.CompanyId, strings.Join(summarySets, ", "), m.OpenTime.Format(utils.LayoutDate), m.AgentUserId)
				if qryRes := dbReport.Exec(qry); qryRes.Error != nil {
					utils.WriteLog(fmt.Sprintf("%s; %s, Error: %+v;", logPrefix, qry, qryRes.Error), utils.LogLevelError)
					utils.WriteToFile(fmt.Sprintf("summary_%s", time.Now().Format("2006_01_02")), qry)
				} else if qryRes.RowsAffected == 0 {
					fields := append(summaryFields, "date", "agent_id", "agent_username", "agent_name", "created_at", "created_by")
					vals := append(summaryVals,
						fmt.Sprintf("'%s'", m.OpenTime.Format(utils.LayoutDate)),
						fmt.Sprintf("'%s'", m.AgentUserId),
						fmt.Sprintf("'%s'", m.AgentUsername),
						fmt.Sprintf("'%s'", m.AgentName),
						fmt.Sprintf("'%s'", time.Now().Format(time.RFC3339Nano)),
						fmt.Sprintf("'%s'", "migration-engine"),
					)

					qry = fmt.Sprintf(`INSERT INTO "%s_summary_daily_peragent" (%s) VALUES (%s);`, m.CompanyId, strings.Join(fields, ", "), strings.Join(vals, ", "))
					if qryRes = dbReport.Exec(qry); qryRes.Error != nil {
						utils.WriteLog(fmt.Sprintf("%s; %s, Error: %+v;", logPrefix, qry, qryRes.Error), utils.LogLevelError)
						utils.WriteToFile(fmt.Sprintf("summary_%s", time.Now().Format("2006_01_02")), qry)
					}
				}
			}(logPrefix, summaryAgentSets, summaryAgentFields, summaryAgentVals)

			//summary per agent
			go func(logPrefix string, summarySets, summaryFields, summaryVals []string) {
				qry := fmt.Sprintf(`UPDATE "%s_summary_peragent" SET %s WHERE agent_id = '%s';`, m.CompanyId, strings.Join(summarySets, ", "), m.AgentUserId)
				if qryRes := dbReport.Exec(qry); qryRes.Error != nil {
					utils.WriteLog(fmt.Sprintf("%s; %s, Error: %+v;", logPrefix, qry, qryRes.Error), utils.LogLevelError)
					utils.WriteToFile(fmt.Sprintf("summary_%s", time.Now().Format("2006_01_02")), qry)
				} else if qryRes.RowsAffected == 0 {
					fields := append(summaryFields, "agent_id", "agent_username", "agent_name", "created_at", "created_by")
					vals := append(summaryVals,
						fmt.Sprintf("'%s'", m.AgentUserId),
						fmt.Sprintf("'%s'", m.AgentUsername),
						fmt.Sprintf("'%s'", m.AgentName),
						fmt.Sprintf("'%s'", time.Now().Format(time.RFC3339Nano)),
						fmt.Sprintf("'%s'", "migration-engine"),
					)

					qry = fmt.Sprintf(`INSERT INTO "%s_summary_peragent" (%s) VALUES (%s);`, m.CompanyId, strings.Join(fields, ", "), strings.Join(vals, ", "))
					if qryRes = dbReport.Exec(qry); qryRes.Error != nil {
						utils.WriteLog(fmt.Sprintf("%s; %s, Error: %+v;", logPrefix, qry, qryRes.Error), utils.LogLevelError)
						utils.WriteToFile(fmt.Sprintf("summary_%s", time.Now().Format("2006_01_02")), qry)
					}
				}
			}(logPrefix, summaryAgentSets, summaryAgentFields, summaryAgentVals)
		}
	}
	debug++
	utils.WriteLog(fmt.Sprintf("%s [INSERT] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, totalInserted, len(errorMessages), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

	//write error to file
	if len(errorMessages) > 0 {
		filename := fmt.Sprintf("%s_%s_%v", appName, time.Now().Format("2006_01_02"), time.Now().Unix())
		utils.WriteErrorMap(filename, errorMessages)
	}
	if len(errorDuplicates) > 0 {
		filename := fmt.Sprintf("%s_%s_%v_duplicate", appName, time.Now().Format("2006_01_02"), time.Now().Unix())
		utils.WriteErrorMap(filename, errorDuplicates)
	}

	utils.WriteLog(fmt.Sprintf("%s end; duration: %v", logPrefix, time.Now().Sub(tStart)), utils.LogLevelDebug)
}
