package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (PinRoom) TableName() string {
	return "pin_rooms"
}

type PinRoom struct {
	Id         uuid.UUID   `json:"id" gorm:"column:id"`
	UserId     uuid.UUID   `json:"user_id" gorm:"column:user_id"`
	RoomId     uuid.UUID   `json:"room_id" gorm:"column:room_id"`
	RoomDetail RoomDetails `json:"room_detail" gorm:"foreignKey:RoomId;AssociationForeignKey:Id;"`
	CreatedAt  time.Time   `json:"created_at" gorm:"column:created_at"`
	UpdatedAt  time.Time   `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt  *time.Time  `json:"-" gorm:"column:deleted_at"`
	Error      string      `json:"error" gorm:"-"`
}

func (PinRooms) TableName() string {
	return "pin_rooms"
}

type PinRooms struct {
	Id        uuid.UUID  `json:"id" gorm:"column:id"`
	UserId    uuid.UUID  `json:"user_id" gorm:"column:user_id"`
	RoomId    uuid.UUID  `json:"room_id" gorm:"column:room_id"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
}
