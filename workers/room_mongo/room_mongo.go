package main

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"konnek-migration/models"
	"konnek-migration/mongos"
	"konnek-migration/utils"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	utils.Init()

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
	appName := "room_mongo"
	if os.Getenv("APP_NAME") != "" {
		appName = os.Getenv("APP_NAME")
	}
	logPrefix := fmt.Sprintf("[%v][%v]", logID, appName)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	//Fetch the data from existing PSQL database

	//Set the filters
	if os.Getenv("COMPANYID") != "" {
		dstDB = dstDB.Where("r.company_id = ?", os.Getenv("COMPANYID"))
	}
	if os.Getenv("ROOM_ID") != "" {
		dstDB = dstDB.Where("r.id = ?", os.Getenv("ROOM_ID"))
	}

	if os.Getenv("START_DATE") != "" && os.Getenv("END_DATE") != "" {
		dstDB = dstDB.Where("r.created_at BETWEEN ? AND ?", os.Getenv("START_DATE"), os.Getenv("END_DATE"))
	} else if os.Getenv("START_DATE") != "" && os.Getenv("END_DATE") == "" {
		dstDB = dstDB.Where("r.created_at >=?", os.Getenv("START_DATE"))
	} else if os.Getenv("START_DATE") == "" && os.Getenv("END_DATE") != "" {
		dstDB = dstDB.Where("r.created_at <=?", os.Getenv("END_DATE"))
	}

	if os.Getenv("ORDER_BY") != "" {
		sortMap := map[string]string{
			os.Getenv("ORDER_BY"): "r." + os.Getenv("ORDER_BY"),
		}
		if strings.ToUpper(os.Getenv("ORDER_DIRECTION")) == "DESC" {
			dstDB = dstDB.Order(sortMap[os.Getenv("ORDER_BY")] + " DESC")
		} else {
			dstDB = dstDB.Order(sortMap[os.Getenv("ORDER_BY")])
		}
	}

	offset, _ := strconv.Atoi(os.Getenv("OFFSET"))
	limit, _ := strconv.Atoi(os.Getenv("LIMIT"))
	if offset > 0 {
		dstDB = dstDB.Offset(offset)
	}
	if limit > 0 {
		dstDB = dstDB.Limit(limit)
	}

	// Query data dari new PSQL database untuk di store ke mongoDB
	var dataRooms []models.FetchRoom
	err = dstDB.Select("r.company_id, r.channel_code, r.customer_user_id, r.id, s.id AS session_id, s.seq_id, s.division_id, s.agent_user_id, s.last_chat_message_id, s.categories, s.bot_status, s.status, s.open_time, s.queue_time, s.assign_time, s.first_response_time, s.last_agent_chat_time, s.close_time, cm.message_id, cm.reply_id AS message_reply_id, cm.from_type AS message_from_type, cm.type AS message_type, cm.text AS message_text, cm.payload AS message_payload, cm.status AS message_status, cm.message_time, cm.created_at AS message_created_at, cm.id AS chat_message_id, cm.seq_id AS conversation_seq_id, cust.name AS customer_name, cust.username AS customer_username, agent.name AS agent_name, agent.username AS agent_username").
		Table("rooms r").
		Joins("JOIN sessions s ON s.id = r.last_session_id").
		Joins("JOIN users cust ON cust.id = r.customer_user_id").
		Joins("LEFT JOIN chat_messages cm ON cm.room_id = r.id").
		Joins("LEFT JOIN users agent ON agent.id = s.agent_user_id").
		Find(&dataRooms).Error
	if err != nil {
		utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
		return
	}

	debug++
	logPrefix += "[dstDB]"
	utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, len(dataRooms), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

	var errorMessages []string
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
			//first, delete room_open to avoid duplicates
			_, err = mongDb.DeleteMany(os.Getenv("COMPANYID")+"_room_open", filter)

			if _, err = mongDb.Store(fetchRoom.CompanyId.String()+"_room_open", mongoData); err != nil {
				utils.WriteLog(fmt.Sprintf("%s; failed insert room_open: %s; error: %v", logPrefix, fetchRoom.Id.String(), err), utils.LogLevelError)
				errCountMdb++
				errorMessages = append(errorMessages, fmt.Sprintf("[%v] insert error: %v | DATA mongoDB: %v", time.Now(), err.Error(), mongoData))
				continue
			}

			//delete from mongoDB room_close after insert into mongoDB room_open
			if _, err = mongDb.Delete(fetchRoom.CompanyId.String()+"_room_closed", filter); err != nil {
				utils.WriteLog(fmt.Sprintf("%s; failed delete from mongoDB: %s; error: %v", logPrefix, fetchRoom.Id.String(), err), utils.LogLevelError)
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
	debug++
	utils.WriteLog(fmt.Sprintf("%s [MongoDB] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalInsertedMdb, errCountMdb, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
}
