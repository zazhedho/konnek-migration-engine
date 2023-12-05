package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (ChannelConfigExist) TableName() string {
	return "channel_config"
}

type ChannelConfigExist struct {
	Id         uuid.UUID  `json:"id" gorm:"column:id"`
	CompanyId  uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	ChannelId  uuid.UUID  `json:"channel_id" gorm:"column:channel_id"`
	Channel    Channel    `json:"channel" gorm:"Foreignkey:ChannelId;association_foreignkey:Id;"`
	Key        string     `json:"key" gorm:"column:key"`
	Content    string     `json:"content" gorm:"column:content"`
	ErrMessage string     `json:"err_message" gorm:"column:err_message"`
	CreatedAt  time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt  *time.Time `json:"-" gorm:"column:deleted_at"`
	Error      string     `json:"error" gorm:"-"`
}

func (ChannelConfigReeng) TableName() string {
	return "channel_configs"
}

type ChannelConfigReeng struct {
	Id          uuid.UUID  `json:"id" gorm:"column:id"`
	CompanyId   uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	ChannelCode string     `json:"channel_code" gorm:"column:channel_code"`
	Key         string     `json:"key" gorm:"column:key"`
	Content     string     `json:"content" gorm:"column:content"`
	ErrMessage  string     `json:"err_message" gorm:"column:err_message"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy   uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy   uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt   *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy   uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
}
