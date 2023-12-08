package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (SessionCategory) TableName() string {
	return "session_categories"
}

type SessionCategory struct {
	Id          uuid.UUID  `json:"id" gorm:"column:id"`
	CompanyId   uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	Name        string     `json:"name" gorm:"column:name"`
	Description string     `json:"description" gorm:"column:description"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy   uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy   uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt   *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy   uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
	Error       string     `json:"error" gorm:"-"`
}

func (SessionCategories) TableName() string {
	return "session_categories"
}

type SessionCategories struct {
	Id          uuid.UUID  `json:"id" gorm:"column:id"`
	CompanyId   uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	Name        string     `json:"name" gorm:"column:name"`
	Description string     `json:"description" gorm:"column:description"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy   uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy   uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt   *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy   uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
}
