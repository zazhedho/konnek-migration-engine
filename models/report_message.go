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
