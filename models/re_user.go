package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

const (
	RoleTypeInternal = 1
	RoleTypeClient   = 2
	RoleTypeCustomer = 3
	RoleTypeBot      = 4
)

var UserTypeMap = map[string]int{
	RoleAdminKonnek:      RoleTypeInternal,
	RoleAdmin:            RoleTypeClient,
	RoleSupervisor:       RoleTypeClient,
	RoleAgent:            RoleTypeClient,
	RoleRtfm:             RoleTypeClient,
	RoleQualityAssurance: RoleTypeClient,
	RoleCustomer:         RoleTypeCustomer,
	RoleBot:              RoleTypeBot,
}

func (Users) TableName() string {
	return "users"
}

type Users struct {
	Id                 uuid.UUID  `json:"id" gorm:"column:id"`
	CompanyId          uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	RolesId            uuid.UUID  `json:"roles_id" gorm:"column:roles_id"`
	Type               string     `json:"type" gorm:"column:type"`
	CustomerChannel    string     `json:"customer_channel" gorm:"column:customer_channel"`
	Username           string     `json:"username" gorm:"column:username"`
	Password           string     `json:"password" gorm:"column:password"`
	LastChangePwd      *time.Time `json:"last_change_password" gorm:"column:last_change_password"`
	Status             int        `json:"status" gorm:"column:status"`
	OnlineStatus       int        `json:"online_status" gorm:"column:online_status"`
	LoginTime          *time.Time `json:"login_time" gorm:"column:login_time"`
	LoginSession       uuid.UUID  `json:"login_session" gorm:"column:login_session"`
	Email              string     `json:"email" gorm:"column:email"`
	Phone              string     `json:"phone" gorm:"column:phone"`
	Name               string     `json:"name" gorm:"column:name"`
	Avatar             string     `json:"avatar" gorm:"column:avatar"`
	Description        string     `json:"description" gorm:"column:description"`
	Tags               string     `json:"tags" gorm:"column:tags"`
	CustomerReplyToken string     `json:"customer_reply_token" gorm:"column:customer_reply_token"`
	DivisionId         uuid.UUID  `json:"division_id" gorm:"column:division_id"`
	RoomOpenAgent      int        `json:"room_open_agent" gorm:"column:room_open_agent"`
	RoomCloseAgent     int        `json:"room_close_agent" gorm:"column:room_close_agent"`
	SoundNotification  bool       `json:"sound_notification" gorm:"column:sound_notification"`
	CreatedAt          time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy          uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy          uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt          *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy          uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
}
