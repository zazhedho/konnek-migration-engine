package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (m ReportMessage) TableName() string {
	return m.TablePrefix + "message"
}

type ReportMessage struct {
	TablePrefix string `gorm:"-"`

	Id        uuid.UUID `json:"id" gorm:"column:id"`
	MessageId string    `json:"message_id" gorm:"column:message_id"`
	ReplyId   string    `json:"reply_id" gorm:"column:reply_id"`

	RoomId    uuid.UUID `json:"room_id" gorm:"column:room_id"`
	SessionId uuid.UUID `json:"session_id" gorm:"column:session_id"`
	FromType  string    `json:"from_type" gorm:"column:from_type"`

	UserId       uuid.UUID `json:"user_id" gorm:"column:user_id"`
	Username     string    `json:"username" gorm:"column:username"`
	UserFullname string    `json:"user_fullname" gorm:"column:user_fullname"`

	Type        string     `json:"type" gorm:"column:type"`
	Text        string     `json:"text" gorm:"column:text"`
	Payload     string     `json:"payload" gorm:"column:payload"`
	Status      int        `json:"status" gorm:"column:status"`
	MessageTime *time.Time `json:"message_time" gorm:"column:message_time"`

	LastUpdate time.Time `json:"last_update" gorm:"column:last_update"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
	CreatedBy  string    `json:"created_by" gorm:"column:created_by"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy  string    `json:"updated_by" gorm:"column:updated_by"`
}

func (ChatMessageRpt) TableName() string {
	return "chat_messages"
}

// ChatMessages struct db reengineering
type ChatMessageRpt struct {
	Id                uuid.UUID    `json:"id" gorm:"column:id"`
	RoomId            uuid.UUID    `json:"room_id" gorm:"column:room_id"`
	SessionId         uuid.UUID    `json:"session_id" gorm:"column:session_id"`
	UserId            uuid.UUID    `json:"user_id" gorm:"column:user_id"`
	Users             UserMessages `json:"user" gorm:"Foreignkey:UserId;association_foreignkey:Id;"`
	MessageId         string       `json:"message_id" gorm:"column:message_id"`
	ReplyId           string       `json:"reply_id" gorm:"column:reply_id"`
	ProviderMessageId string       `json:"provider_message_id" gorm:"column:provider_message_id"`
	FromType          string       `json:"from_type" gorm:"column:from_type"`
	Type              string       `json:"type" gorm:"column:type"`
	Text              string       `json:"text" gorm:"column:text"`
	Payload           string       `json:"payload" gorm:"column:payload"`
	Status            int          `json:"status" gorm:"column:status"`
	MessageTime       *time.Time   `json:"message_time" gorm:"column:message_time"`
	ReceivedAt        *time.Time   `json:"received_at" gorm:"column:received_at"`
	ProcessedAt       *time.Time   `json:"processed_at" gorm:"column:processed_at"`
	CreatedAt         time.Time    `json:"created_at" gorm:"column:created_at"`
	CreatedBy         uuid.UUID    `json:"created_by" gorm:"column:created_by"`
	UpdatedAt         time.Time    `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy         uuid.UUID    `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt         *time.Time   `json:"-" gorm:"column:deleted_at"`
	DeletedBy         uuid.UUID    `json:"deleted_by" gorm:"column:deleted_by"`
	TextDeleted       string       `json:"-" gorm:"column:text_deleted"`
	DeleteTime        *time.Time   `json:"-" gorm:"column:delete_time"`
	Retry             int          `json:"retry_sending" gorm:"column:retry"`
	RetryTime         *time.Time   `json:"retry_time" gorm:"column:retry_time"`
	Error             string       `json:"error" gorm:"-"`
}

func (UserMessages) TableName() string {
	return "users"
}

type UserMessages struct {
	Id        uuid.UUID `json:"id" gorm:"column:id"`
	Username  string    `json:"username" gorm:"column:username"`
	Name      string    `json:"name" gorm:"column:name"`
	CompanyId uuid.UUID `json:"company_id" gorm:"column:company_id"`
}
