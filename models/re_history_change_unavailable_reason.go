package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (HistoryChangeUnavailableReason) TableName() string {
	return "history_change_unavailable_reason"
}

type HistoryChangeUnavailableReason struct {
	Id        uuid.UUID  `json:"availability_id" gorm:"column:availability_id"`
	CompanyId uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	UserId    uuid.UUID  `json:"user_id" gorm:"column:user_id"`
	OldType   string     `json:"old_type" gorm:"column:old_type"`
	NewType   string     `json:"new_type" gorm:"column:new_type"`
	OldReason string     `json:"old_reason" gorm:"column:old_reason"`
	NewReason string     `json:"new_reason" gorm:"column:new_reason"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
}
