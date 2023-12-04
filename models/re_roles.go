package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (Roles) TableName() string {
	return "roles"
}

type Roles struct {
	Id            uuid.UUID   `json:"id" gorm:"column:id"`
	CompanyId     uuid.UUID   `json:"company_id" gorm:"column:company_id"`
	Name          string      `json:"name" gorm:"column:name"`
	IsAgent       bool        `json:"is_agent" gorm:"column:is_agent"`
	IsAdmin       bool        `json:"is_admin" gorm:"column:is_admin"`
	UrlAfterLogin string      `json:"url_after_login" gorm:"column:url_after_login"`
	Status        bool        `json:"status" gorm:"column:status"`
	MenuAccess    interface{} `json:"access" gorm:"menu_access"`
	CreatedAt     time.Time   `json:"created_at" gorm:"column:created_at"`
	CreatedBy     uuid.UUID   `json:"created_by" gorm:"column:created_by"`
	UpdatedAt     time.Time   `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy     uuid.UUID   `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt     *time.Time  `json:"-" gorm:"column:deleted_at"`
	DeletedBy     uuid.UUID   `json:"deleted_by" gorm:"column:deleted_by"`
}
