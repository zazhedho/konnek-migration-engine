package main

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/joho/godotenv/autoload"
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

	//err = godotenv.Load("workers/conversation_mongo/.env")
	//if err != nil {
	//	log.Fatalf("Error loading service .env")
	//}

	logID := uuid.NewV4()
	appName := "conversation_mongo"
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
		dstDB = dstDB.Where("company_id = ?", os.Getenv("COMPANYID"))
	}
	if os.Getenv("ROOM_ID") != "" {
		dstDB = dstDB.Where("cm.room_id = ?", os.Getenv("ROOM_ID"))
	}
	if os.Getenv("FROM_TYPE") != "" {
		dstDB = dstDB.Where("cm.from_type = ?", os.Getenv("FROM_TYPE"))
	}

	if os.Getenv("START_DATE") != "" && os.Getenv("END_DATE") != "" {
		dstDB = dstDB.Where("cm.created_at BETWEEN ? AND ?", os.Getenv("START_DATE"), os.Getenv("END_DATE"))
	} else if os.Getenv("START_DATE") != "" && os.Getenv("END_DATE") == "" {
		dstDB = dstDB.Where("cm.created_at >=?", os.Getenv("START_DATE"))
	} else if os.Getenv("START_DATE") == "" && os.Getenv("END_DATE") != "" {
		dstDB = dstDB.Where("cm.created_at <=?", os.Getenv("END_DATE"))
	}

	if os.Getenv("ORDER_BY") != "" {
		sortMap := map[string]string{
			"created_at": "created_at",
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
	var dataConversations []models.FetchConversation
	dstDB.Select("cm.id, cm.seq_id, cm.room_id, cm.session_id, cm.user_id, cm.message_id, cm.reply_id, cm.from_type, cm.type, cm.text, cm.payload, cm.status, cm.message_time, cm.created_at, u.name AS created_by").
		Table("chat_messages cm").
		Joins("JOIN users u ON u.id = cm.user_id").
		Find(&dataConversations)

	debug++
	logPrefix += "[dstDB]"
	utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, len(dataConversations), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

	var errorMessages []string
	totalInsertedMdb := 0
	errCountMdb := 0
	mongDb := mongos.NewConversation(mongoDB)
	for _, fetchConversation := range dataConversations {
		mongoData := mongos.Conversation{
			MongoId:   primitive.NewObjectID(),
			SeqId:     fetchConversation.SeqId,
			RoomId:    fetchConversation.RoomId.String(),
			SessionId: fetchConversation.SessionId.String(),
			Id:        fetchConversation.Id.String(),
			UserId:    fetchConversation.UserId.String(),
			MessageId: fetchConversation.MessageId,
			ReplyId:   fetchConversation.ReplyId,
			FromType:  fetchConversation.FromType,
			Type:      fetchConversation.Type,
			Text:      fetchConversation.Text,
			Payload:   fetchConversation.Payload,
			Status:    fetchConversation.Status,
			Time:      fetchConversation.MessageTime,
			CreatedAt: &fetchConversation.CreatedAt,
			CreatedBy: fetchConversation.CreatedBy,
		}

		filter := bson.M{
			"id": fetchConversation.Id.String(),
		}
		//first, delete conversations to avoid duplicates
		_, err = mongDb.DeleteMany(os.Getenv("COMPANYID")+"_conversation", filter)
		//insert into mongodb _conversation
		if _, err = mongDb.Store(os.Getenv("COMPANYID")+"_conversation", mongoData); err != nil {
			utils.WriteLog(fmt.Sprintf("%s; failed insert conversation: %s; error: %v", logPrefix, fetchConversation.Id, err), utils.LogLevelError)
			errCountMdb++
			errorMessages = append(errorMessages, fmt.Sprintf("[%v] insert error: %v | DATA mongoDB: %v", time.Now(), err.Error(), mongoData))
			continue
		}
		totalInsertedMdb++
	}
	debug++
	utils.WriteLog(fmt.Sprintf("%s [MongoDB] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalInsertedMdb, errCountMdb, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
}
