package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (DivisionExist) TableName() string {
	return "divisions"
}

type DivisionExist struct {
	Id        uuid.UUID  `json:"id" gorm:"column:id"`
	Name      string     `json:"name" gorm:"column:name"`
	CompanyId uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"column:deleted_at"`
}

func (DivisionReeng) TableName() string {
	return "divisions"
}

type DivisionReeng struct {
	Id        uuid.UUID  `json:"id" gorm:"column:id"`
	CompanyId uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	Name      string     `json:"name" gorm:"column:name"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
}
