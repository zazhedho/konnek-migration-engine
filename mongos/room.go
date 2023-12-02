package mongos

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type mongoRoom struct {
	MongoDb *mongo.Database
}

func NewRoom(db *mongo.Database) *mongoRoom {
	return &mongoRoom{
		MongoDb: db,
	}
}

type User struct {
	Id       string `json:"id" bson:"id"`
	Username string `json:"username" bson:"username"`
	Name     string `json:"name" bson:"name"`
}

type Room struct {
	MongoId     interface{} `json:"-" bson:"-"`
	CompanyId   string      `json:"company_id" bson:"company_id"`
	DivisionId  string      `json:"division_id" bson:"division_id"`
	RoomId      string      `json:"room_id" bson:"room_id"`
	SessionId   string      `json:"session_id" bson:"session_id"`
	ChannelCode string      `json:"channel_code" bson:"channel"`
	SeqId       int64       `json:"seq_id" bson:"seq_id"`

	Customer User `json:"customer" bson:"customer"`
	Agent    User `json:"agent" bson:"agent"`

	Categories interface{} `json:"categories" bson:"categories"`
	BotStatus  bool        `json:"bot_status" bson:"bot_status"`
	Status     int         `json:"status" bson:"status"`

	OpenTime          *time.Time `json:"open_time" bson:"open_time"`
	QueTime           *time.Time `json:"queue_time" bson:"queue_time"`
	AssignTime        *time.Time `json:"assign_time" bson:"assign_time"`
	FirstResponseTime *time.Time `json:"first_response_time" bson:"fr_time"`
	LastAgentChatTime *time.Time `json:"last_agent_chat_time" bson:"lr_time"`
	CloseTime         *time.Time `json:"close_time" bson:"close_time"`

	OpenBy     User `json:"open_by" bson:"open_by"`
	HandoverBy User `json:"handover_by" bson:"handover_by"`
	CloseBy    User `json:"close_by" bson:"close_by"`

	Conversations []Conversation `json:"conversations" bson:"conversations"`
}

func (mg *mongoRoom) Store(coll string, doc Room) (Room, error) {
	res, err := mg.MongoDb.Collection(coll).InsertOne(context.TODO(), doc)
	if err != nil {
		return doc, err
	}
	doc.MongoId = res.InsertedID

	_, _ = mg.MongoDb.Collection(coll).Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{{"seq_id", -1}}},
		{Keys: bson.D{{"lr_time", -1}}},
		{Keys: bson.D{{"open_time", 1}}},
		{Keys: bson.D{{"room_id", 1}}},
		{Keys: bson.D{{"agent.id", 1}}},
		{Keys: bson.D{{"conversations.time", -1}}},
		{Keys: bson.D{
			{"customer.name", "text"},
			{"customer.username", "text"},
			{"agent.name", "text"},
			{"agent.username", "text"},
		}},
	})

	return doc, nil
}

func (mg *mongoRoom) Delete(coll string, filter bson.M) (*mongo.DeleteResult, error) {
	return mg.MongoDb.Collection(coll).DeleteOne(context.TODO(), filter)
}

func (mg *mongoRoom) DeleteMany(coll string, filter bson.M) (*mongo.DeleteResult, error) {
	return mg.MongoDb.Collection(coll).DeleteMany(context.TODO(), filter)
}
