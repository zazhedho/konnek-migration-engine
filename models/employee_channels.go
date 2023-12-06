package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (EmployeeChannelExist) TableName() string {
	return "employee_channels"
}

type EmployeeChannelExist struct {
	Id          uuid.UUID `json:"id" gorm:"column:id"`
	UserID      uuid.UUID `json:"user_id" gorm:"column:user_id"`
	CompanyID   uuid.UUID `json:"company_id" gorm:"column:company_id"`
	ChannelName string    `json:"name" gorm:"column:name"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
	Error       string    `json:"error" gorm:"-"`
}

func (EmployeeChannelReeng) TableName() string {
	return "employee_channels"
}

type EmployeeChannelReeng struct {
	Id          uuid.UUID  `json:"-" gorm:"column:id"`
	CompanyId   uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	UserId      uuid.UUID  `json:"user_id" gorm:"column:user_id"`
	ChannelCode string     `json:"channel_code" gorm:"column:channel_code"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy   uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy   uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt   *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy   uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
}
