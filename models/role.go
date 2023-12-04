package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (Role) TableName() string {
	return "roles"
}

type Role struct {
	Id        uuid.UUID  `json:"id" gorm:"column:id"`
	Name      string     `json:"name" gorm:"column:name"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"column:deleted_at"`
}
