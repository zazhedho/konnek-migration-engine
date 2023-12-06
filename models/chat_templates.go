package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (ChatTemplateExist) TableName() string {
	return "chat_templates"
}

type ChatTemplateExist struct {
	Id        uuid.UUID  `json:"id" gorm:"column:id"`
	Keyword   string     `json:"keyword" gorm:"column:keyword"`
	Text      string     `json:"text" gorm:"column:text"`
	CreatedBy uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	CompanyId uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"column:deleted_at"`
	Error     string     `json:"error" gorm:"-"`
}

func (ChatTemplateReeng) TableName() string {
	return "chat_templates"
}

type ChatTemplateReeng struct {
	Id        uuid.UUID  `json:"id" gorm:"column:id"`
	CompanyId uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	Keyword   string     `json:"keyword" gorm:"column:keyword"`
	Text      string     `json:"text" gorm:"column:text"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
}
