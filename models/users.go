package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

const (
	RoleAdminKonnek = "admin_konnek"

	RoleAdmin            = "admin"
	RoleSupervisor       = "supervisor"
	RoleAgent            = "agent"
	RoleRtfm             = "RTFM"
	RoleQualityAssurance = "quality assurance"

	RoleCustomer = "customer"
	RoleBot      = "bot"
)

func (User) TableName() string {
	return "users"
}

type User struct {
	Id           uuid.UUID  `json:"id" gorm:"column:id"`
	PhoneNumber  string     `json:"phone_number" gorm:"column:phone_number"`
	CompanyId    uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	Password     string     `json:"-" gorm:"password"`
	Email        string     `json:"email" gorm:"column:email"`
	AvatarUrl    string     `json:"avatar_url" gorm:"column:avatar_url"`
	Username     string     `json:"username" gorm:"column:username"`
	Status       bool       `json:"status" gorm:"column:status"`
	RolesId      uuid.UUID  `json:"roles_id" gorm:"column:roles_id"`
	Roles        Role       `json:"role" gorm:"Foreignkey:RolesId;association_foreignkey:Id;"`
	Customer     Customer   `json:"customer" gorm:"Foreignkey:Id;association_foreignkey:UserId"`
	Employee     Employee   `json:"employee" gorm:"Foreignkey:Id;association_foreignkey:UserId"`
	Description  string     `json:"description" gorm:"column:description"`
	CreatedAt    time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt    *time.Time `json:"-" gorm:"column:deleted_at"`
	OnlineStatus int        `json:"-" gorm:"column:online_status"`
	LoginTime    *time.Time `json:"login_time" gorm:"login_time"`
	NotifSound   bool       `json:"notification_sound" gorm:"column:notification_sound"`
	Error        string     `json:"error" gorm:"-"`
}

func (Customer) TableName() string {
	return "customers"
}

type Customer struct {
	Id          uuid.UUID `json:"id" gorm:"column:id"`
	UserId      uuid.UUID `json:"user_id" gorm:"column:user_id"`
	Name        string    `json:"name" gorm:"column:name"`
	Tags        string    `json:"tags" gorm:"column:tags"`
	Reply_Token string    `json:"reply_token" gorm:"column:reply_token"`
}

func (Employee) TableName() string {
	return "employees"
}

type Employee struct {
	Id         uuid.UUID `json:"id" gorm:"column:id"`
	UserId     uuid.UUID `json:"user_id" gorm:"column:user_id"`
	Name       string    `json:"name" gorm:"column:name"`
	DivisionId uuid.UUID `json:"division_id" gorm:"column:division_id"`
}
