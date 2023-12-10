package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (Participants) TableName() string {
	return "participants"
}

type Participants struct {
	Id        uuid.UUID  `gorm:"column:id" json:"id"`
	SessionID uuid.UUID  `gorm:"column:session_id" json:"session_id"`
	UserID    uuid.UUID  `gorm:"column:user_id" json:"user_id"`
	User      User       `json:"user" gorm:"foreignKey:UserID;AssociationForeignKey:Id;"`
	Status    bool       `gorm:"column:status" json:"status"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	CreatedBy uuid.UUID  `gorm:"column:created_by" json:"created_by"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	UpdatedBy uuid.UUID  `gorm:"column:updated_by" json:"updated_by"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"-"`
	DeletedBy uuid.UUID  `gorm:"column:deleted_by" json:"deleted_by"`
	Error     string     `json:"error" gorm:"-"`
}
