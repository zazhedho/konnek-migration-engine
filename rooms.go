package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/joho/godotenv/autoload"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"konnek-migration/models"
	"konnek-migration/mongos"
	"konnek-migration/utils"
	"log"
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

	// mongoDB connection
	conn, mongoDB, err := utils.MongoDBConnection()
	if err != nil {
		utils.WriteLog(fmt.Sprintf("Connection MongoDB; ERROR: %+v", err), utils.LogLevelError)
	}
	defer func(conn *mongo.Client) {
		if err = conn.Disconnect(context.Background()); err != nil {
			utils.WriteLog(fmt.Sprintf("Disconnecting Mongo ERROR: %v", err), utils.LogLevelError)
		}
	}(conn)

	logID := uuid.NewV4()
	logPrefix := fmt.Sprintf("[%v][rooms]", logID)
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
	var dataRoomDetails []models.RoomDetails
	if err = scDB.Find(&dataRoomDetails).Error; err != nil {
		utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
		return
	}
	totalFetch := len(dataRoomDetails)

	debug++
	utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
	debugT = time.Now()

	//Insert into the new PSQL database
	var errorMessages []string
	totalInsertedSQL := 0 //success insert
	errCountSQL := 0
	for _, dataRoomDetail := range dataRoomDetails {
		mRooms := models.Rooms{
			ID:             dataRoomDetail.ID,
			CompanyID:      dataRoomDetail.CompanyID,
			ChannelCode:    dataRoomDetail.ChannelName,
			CustomerUserID: dataRoomDetail.CustomerUserID,
			LastSessionID:  dataRoomDetail.SessionID,
			CreatedAt:      dataRoomDetail.CreatedAt,
			CreatedBy:      uuid.Nil,
			UpdatedAt:      dataRoomDetail.CreatedAt,
			UpdatedBy:      uuid.Nil,
			DeletedAt:      dataRoomDetail.DeletedAt,
			DeletedBy:      uuid.Nil,
		}

		if err = dstDB.Create(&mRooms).Error; err != nil {
			var errCode *pq.Error
			if errors.As(err, &errCode) {
				if errCode.Code == "23505" { //unique_violation
					mRooms.ID = uuid.NewV4()
					if err = dstDB.Create(&mRooms).Error; err != nil {
						utils.WriteLog(fmt.Sprintf("%s; failed insert to SQL: %s; error: %v", logPrefix, dataRoomDetail.ID, err), utils.LogLevelError)
					}
				}
			}

			utils.WriteLog(fmt.Sprintf("%s; failed insert to SQL: %s; error: %v", logPrefix, dataRoomDetail.ID, err), utils.LogLevelError)
			errCountSQL++
			errorMessages = append(errorMessages, fmt.Sprintf("[%v] insert error: %v | DATA SQL: %v", time.Now(), err.Error(), mRooms))
			continue
		}
		totalInsertedSQL++
	}
	debug++
	utils.WriteLog(fmt.Sprintf("%s [PSQL] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalInsertedSQL, errCountSQL, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

	// Query data dari new PSQL database untuk di store ke mongoDB
	var dataRooms []models.FetchRoom
	sqlQuery := `
		SELECT 
		    r.company_id, r.channel_code, r.customer_user_id, r.id, s.id AS session_id, s.seq_id, s.division_id, s.agent_user_id, s.last_chat_message_id, s.categories, s.bot_status, s.status, s.open_time, s.queue_time, s.assign_time, s.first_response_time, s.last_agent_chat_time, s.close_time, cm.message_id, cm.reply_id AS message_reply_id, cm.from_type AS message_from_type, cm.type AS message_type, cm.text AS message_text, cm.payload AS message_payload, cm.status AS message_status, cm.message_time, cm.created_at AS message_created_at, cm.id AS chat_message_id, cm.seq_id AS conversation_seq_id, cust.name AS customer_name, cust.username AS customer_username, agent.name AS agent_name, agent.username AS agent_username 
		FROM rooms r 
		JOIN sessions s ON s.id = r.last_session_id 
		JOIN users cust ON cust.id = r.customer_user_id 
		LEFT JOIN chat_messages cm ON cm.room_id = r.id
		LEFT JOIN users agent ON agent.id = s.agent_user_id
		WHERE 1=1
	`
	if os.Getenv("COMPANYID") != "" {
		sqlQuery += fmt.Sprintf(` AND r.company_id = '%s'`, os.Getenv("COMPANYID"))
	}
	if os.Getenv("START_DATE") != "" && os.Getenv("END_DATE") != "" {
		sqlQuery += fmt.Sprintf(` AND r.created_at BETWEEN '%s' AND '%s'`, os.Getenv("START_DATE"), os.Getenv("END_DATE"))
	} else if os.Getenv("START_DATE") != "" && os.Getenv("END_DATE") == "" {
		sqlQuery += fmt.Sprintf(` AND r.created_at >= '%s'`, os.Getenv("START_DATE"))
	} else if os.Getenv("START_DATE") == "" && os.Getenv("END_DATE") != "" {
		sqlQuery += fmt.Sprintf(` AND r.created_at <= '%s'`, os.Getenv("END_DATE"))
	}
	if os.Getenv("ROOM_ID") != "" {
		sqlQuery += fmt.Sprintf(` AND r.id = '%s'`, os.Getenv("ROOM_ID"))
	}
	dstDB.Raw(sqlQuery).Scan(&dataRooms)

	debug++
	logPrefix += "[dstDB]"
	utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, len(dataRooms), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

	totalInsertedMdb := 0
	errCountMdb := 0
	mongDb := mongos.NewRoom(mongoDB)
	for _, fetchRoom := range dataRooms {
		mongoData := mongos.Room{
			MongoId:     primitive.NewObjectID(),
			CompanyId:   fetchRoom.CompanyId.String(),
			DivisionId:  fetchRoom.DivisionId.String(),
			RoomId:      fetchRoom.Id.String(),
			SessionId:   fetchRoom.SessionId.String(),
			ChannelCode: fetchRoom.ChannelCode,
			SeqId:       fetchRoom.SeqId,
			Customer: mongos.User{
				Id:       fetchRoom.CustomerUserId.String(),
				Username: fetchRoom.CustomerUsername,
				Name:     fetchRoom.CustomerName,
			},
			Agent: mongos.User{
				Id:       fetchRoom.AgentUserId.String(),
				Username: fetchRoom.AgentUsername,
				Name:     fetchRoom.AgentName,
			},
			Categories:        fetchRoom.Categories,
			BotStatus:         fetchRoom.BotStatus,
			Status:            fetchRoom.Status,
			OpenTime:          fetchRoom.OpenTime,
			QueTime:           fetchRoom.QueTime,
			AssignTime:        fetchRoom.AssignTime,
			FirstResponseTime: fetchRoom.FirstResponseTime,
			LastAgentChatTime: fetchRoom.LastAgentChatTime,
			CloseTime:         fetchRoom.CloseTime,
			OpenBy: mongos.User{
				Id:       fetchRoom.CustomerUserId.String(),
				Username: fetchRoom.CustomerUsername,
				Name:     fetchRoom.CustomerName,
			},
			HandoverBy: mongos.User{
				Id:       fetchRoom.AgentUserId.String(),
				Username: fetchRoom.AgentUsername,
				Name:     fetchRoom.AgentName,
			},
			CloseBy: mongos.User{
				Id:       fmt.Sprintf("%s", uuid.Nil),
				Username: "",
				Name:     "",
			},
			Conversations: []mongos.Conversation{
				{
					SeqId:     fetchRoom.ConversationSeqId,
					RoomId:    fetchRoom.Id.String(),
					SessionId: fetchRoom.SessionId.String(),
					Id:        fetchRoom.ChatMessageId,
					UserId:    fetchRoom.CustomerUserId.String(),
					MessageId: fetchRoom.MessageId,
					ReplyId:   fetchRoom.ReplyId,
					FromType:  fetchRoom.FromType,
					Type:      fetchRoom.Type,
					Text:      fetchRoom.Text,
					Payload:   fetchRoom.Payload,
					Status:    fetchRoom.MessageStatus,
					Time:      fetchRoom.MessageTime,
					CreatedAt: &fetchRoom.MessageCreatedAt,
					CreatedBy: fetchRoom.CustomerName,
				},
			},
		}

		filter := bson.M{
			"room_id": fetchRoom.Id.String(),
		}
		//insert into mongodb room_open
		if fetchRoom.Status != -1 {
			if _, err = mongDb.Store(fetchRoom.CompanyId.String()+"_room_open", mongoData); err != nil {
				utils.WriteLog(fmt.Sprintf("%s; failed insert room_open: %s; error: %v", logPrefix, fetchRoom.Id, err), utils.LogLevelError)
				errCountMdb++
				errorMessages = append(errorMessages, fmt.Sprintf("[%v] insert error: %v | DATA mongoDB: %v", time.Now(), err.Error(), mongoData))
				continue
			}

			//delete from mongoDB room_close
			if _, err = mongDb.Delete(fetchRoom.CompanyId.String()+"_room_closed", filter); err != nil {
				utils.WriteLog(fmt.Sprintf("%s; failed delete from mongoDB: %s; error: %v", logPrefix, fetchRoom.Id, err), utils.LogLevelError)
				if fetchRoom.AgentUserId != uuid.Nil { //delete from mongoDB room_closed_agent if agent exists
					_, err = mongDb.Delete(fetchRoom.AgentUserId.String()+"_room_closed_agent", filter)
				}
			}
		} else {
			docRoomClosed := mongoData
			docRoomClosed.CloseBy = mongos.User{
				Id:       fetchRoom.AgentUserId.String(),
				Username: fetchRoom.AgentUsername,
				Name:     fetchRoom.AgentName,
			}

			//insert into mongoDB room_closed
			if _, err = mongDb.Store(fetchRoom.CompanyId.String()+"_room_closed", docRoomClosed); err != nil {
				utils.WriteLog(fmt.Sprintf("%s; failed insert room_closed: %s; error: %v", logPrefix, fetchRoom.Id, err), utils.LogLevelError)
				errCountMdb++
				errorMessages = append(errorMessages, fmt.Sprintf("[%v] insert error: %v | DATA mongoDB: %v", time.Now(), err.Error(), docRoomClosed))
				continue
			}
			if fetchRoom.AgentUserId != uuid.Nil { //insert into mongoDB room_closed_agent
				if _, err = mongDb.Store(fetchRoom.AgentUserId.String()+"_room_closed_agent", docRoomClosed); err != nil {
					utils.WriteLog(fmt.Sprintf("%s; failed insert room_closed_agent: %s; error: %v", logPrefix, fetchRoom.Id, err), utils.LogLevelError)
					errCountMdb++
					errorMessages = append(errorMessages, fmt.Sprintf("[%v] insert error: %v | DATA mongoDB: %v", time.Now(), err.Error(), docRoomClosed))
					continue
				}
			}
			//delete from mongoDB room_open
			if _, err = mongDb.Delete(fetchRoom.CompanyId.String()+"_room_open", filter); err != nil {
				utils.WriteLog(fmt.Sprintf("%s; failed delete from room_open: %s; error: %v", logPrefix, fetchRoom.Id, err), utils.LogLevelError)
			}
		}
		totalInsertedMdb++
	}

	// Write error messages to a text file
	formattedTime := time.Now().Format("2006-01-02_150405")
	logErrorFilename := fmt.Sprintf("error_messages_%s.log", formattedTime)
	if len(errorMessages) > 0 {
		createFIle, errCreate := os.Create(logErrorFilename)
		if errCreate != nil {
			utils.WriteLog(fmt.Sprintf("%s [ERROR] Failed write to %s | Error: %v", logPrefix, logErrorFilename, errCreate), utils.LogLevelError)
			return
		}
		defer createFIle.Close()

		for _, errMsg := range errorMessages {
			_, errWrite := createFIle.WriteString(errMsg + "\n")
			if errWrite != nil {
				log.Printf("Failed to write to error_messages_%s.log : %v", formattedTime, errWrite)
				utils.WriteLog(fmt.Sprintf("%s [ERROR] Failed write to %s | Error: %v", logPrefix, logErrorFilename, errWrite), utils.LogLevelError)
			}
		}
		utils.WriteLog(fmt.Sprintf("%s [SUCCESS] Messages error written to %s", logPrefix, logErrorFilename), utils.LogLevelInfo)
	}
	debug++
	utils.WriteLog(fmt.Sprintf("%s [MongoDB] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalInsertedMdb, errCountMdb, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
}
