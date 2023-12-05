package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (CustomerInformation) TableName() string {
	return "customer_informations"
}

type CustomerInformation struct {
	Id          uuid.UUID  `json:"id" gorm:"column:id"`
	UserId      uuid.UUID  `json:"user_id" gorm:"column:user_id"`
	Category    int        `json:"category" gorm:"column:category"`
	Title       string     `json:"title" gorm:"column:title"`
	Description string     `json:"description" gorm:"column:description"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy   uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy   uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt   *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy   uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
}
