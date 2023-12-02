package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (UnavailableReason) TableName() string {
	return "unavailable_reasons"
}

type UnavailableReason struct {
	Id        uuid.UUID  `json:"id" gorm:"column:id"`
	CompanyId uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	Type      string     `json:"type" gorm:"column:type"`
	Reason    string     `json:"reason" gorm:"column:reason"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
}
