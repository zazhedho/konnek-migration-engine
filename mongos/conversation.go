package mongos

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type mongoConversation struct {
	MongoDb *mongo.Database
}

func NewConversation(db *mongo.Database) *mongoConversation {
	return &mongoConversation{
		MongoDb: db,
	}
}

type Conversation struct {
	MongoId   interface{} `json:"-" bson:"-"`
	SeqId     int64       `json:"seq_id" bson:"seq_id"`
	RoomId    string      `json:"room_id" bson:"room_id"`
	SessionId string      `json:"session_id" bson:"session_id"`
	Id        string      `json:"id" bson:"id"`
	UserId    string      `json:"user_id" bson:"user_id"`
	MessageId string      `json:"message_id" bson:"message_id"`
	ReplyId   string      `json:"reply_id" bson:"reply_id"`
	FromType  string      `json:"from_type" bson:"from_type"`
	Type      string      `json:"type" bson:"type"`
	Text      string      `json:"text" bson:"text"`
	Payload   string      `json:"payload" bson:"payload"`
	Status    int         `json:"status" bson:"status"`
	Time      *time.Time  `json:"message_time" bson:"time"`
	CreatedAt *time.Time  `json:"created_at" bson:"created_at"`
	CreatedBy string      `json:"created_by" bson:"created_by"`
}

func (mg *mongoConversation) Store(coll string, doc Conversation) (Conversation, error) {
	res, err := mg.MongoDb.Collection(coll).InsertOne(context.TODO(), doc)
	if err != nil {
		return doc, err
	}
	doc.MongoId = res.InsertedID

	_, _ = mg.MongoDb.Collection(coll).Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{
			{"message_id", 1},
		}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"seq_id", -1}}},
		{Keys: bson.D{{"time", -1}}},
		{Keys: bson.D{{"room_id", 1}}},
		{Keys: bson.D{{"id", 1}}},
		{Keys: bson.D{
			{"session_id", "text"},
			{"text", "text"},
		}},
	})

	return doc, nil
}

func (mg *mongoConversation) DeleteMany(coll string, filter bson.M) (*mongo.DeleteResult, error) {
	return mg.MongoDb.Collection(coll).DeleteMany(context.TODO(), filter)
}
