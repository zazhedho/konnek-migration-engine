package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (AdditionalInformation) TableName() string {
	return "additional_informations"
}

type AdditionalInformation struct {
	Id          uuid.UUID  `json:"id" gorm:"column:id"`
	CustomerId  uuid.UUID  `json:"customer_id" gorm:"column:customer_id"`
	Customer    Customer   `json:"customer" gorm:"Foreignkey:CustomerId;association_foreignkey:Id"`
	Title       string     `json:"title" gorm:"column:title"`
	Description string     `json:"description" gorm:"column:description"`
	Category    bool       `json:"category" gorm:"column:category"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt   *time.Time `json:"-" gorm:"column:deleted_at"`
	Error       string     `json:"error" gorm:"-"`
}
