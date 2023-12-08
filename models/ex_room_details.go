package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (RoomDetails) TableName() string {
	return "room_details"
}

type RoomDetails struct {
	ID             uuid.UUID  `json:"id" gorm:"column:id"`
	CompanyID      uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	ChannelName    string     `json:"channel_name" gorm:"column:channel_name"`
	CustomerUserID uuid.UUID  `json:"customer_user_id" gorm:"column:customer_user_id"`
	SessionID      uuid.UUID  `json:"session_id" gorm:"column:session_id"`
	CreatedAt      time.Time  `json:"created_at" gorm:"column:created_at"`
	DeletedAt      *time.Time `json:"-" gorm:"column:deleted_at"`
	Error          string     `json:"error" gorm:"-"`
}
