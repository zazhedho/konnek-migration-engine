package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (Channel) TableName() string {
	return "channels"
}

type Channel struct {
	Id        uuid.UUID  `json:"id" gorm:"column:id"`
	Name      string     `json:"name" gorm:"column:name"`
	Icon      string     `json:"icon" gorm:"column:icon"`
	status    bool       `json:"status" gorm:"column:status"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"column:deleted_at"`
}
