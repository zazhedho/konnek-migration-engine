package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (HistoryAvailabilityUser) TableName() string {
	return "history_availability_users"
}

type HistoryAvailabilityUser struct {
	Id              uuid.UUID  `json:"id" gorm:"column:id"`
	CompanyId       uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	UserId          uuid.UUID  `json:"user_id" gorm:"column:user_id"`
	Activity        string     `json:"activity" gorm:"column:activity"`
	Reason          string     `json:"reason" gorm:"column:reason"`
	AvailableTime   time.Time  `json:"available_time" gorm:"column:available_time"`
	RequestTime     time.Time  `json:"request_time" gorm:"column:request_time"`
	UnAvailableTime time.Time  `json:"unavailable_time" gorm:"column:unavailable_time"`
	ReavailableTime time.Time  `json:"reavailable_time" gorm:"column:reavailable_time"`
	Type            string     `json:"type" gorm:"column:type"`
	LoginId         uuid.UUID  `json:"login_id" gorm:"column:login_id"`
	CreatedAt       time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy       uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy       uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt       *time.Time `json:"deleted_at" gorm:"column:deleted_at"`
	DeletedBy       uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
}
