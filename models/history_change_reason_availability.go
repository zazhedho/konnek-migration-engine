package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (HistoryChangeReasonAvailability) TableName() string {
	return "history_change_reason_availability"
}

type HistoryChangeReasonAvailability struct {
	Id        uuid.UUID `json:"availability_id" gorm:"column:availability_id"`
	CompanyId uuid.UUID `json:"company_id" gorm:"column:company_id"`
	UserId    uuid.UUID `json:"user_id" gorm:"column:user_id"`
	OldType   string    `json:"old_type" gorm:"column:old_type"`
	NewType   string    `json:"new_type" gorm:"column:new_type"`
	OldReason string    `json:"old_reason" gorm:"column:old_reason"`
	NewReason string    `json:"new_reason" gorm:"column:new_reason"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}
