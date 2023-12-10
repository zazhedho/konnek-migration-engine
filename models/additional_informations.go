package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (AdditionalInformation) TableName() string {
	return "additional_informations"
}

type AdditionalInformation struct {
	Id          uuid.UUID      `json:"id" gorm:"column:id"`
	CustomerId  uuid.UUID      `json:"customer_id" gorm:"column:customer_id"`
	Customer    CustomerDetail `json:"customer" gorm:"Foreignkey:CustomerId;association_foreignkey:Id"`
	Title       string         `json:"title" gorm:"column:title"`
	Description string         `json:"description" gorm:"column:description"`
	Category    bool           `json:"category" gorm:"column:category"`
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt   *time.Time     `json:"-" gorm:"column:deleted_at"`
	Error       string         `json:"error" gorm:"-"`
}

func (CustomerDetail) TableName() string {
	return "customers"
}

type CustomerDetail struct {
	Id          uuid.UUID   `json:"id" gorm:"column:id"`
	UserId      uuid.UUID   `json:"user_id" gorm:"column:user_id"`
	User        UserCompany `json:"user" gorm:"foreignKey:UserId;AssociationForeignKey:Id;"`
	Name        string      `json:"name" gorm:"column:name"`
	Tags        string      `json:"tags" gorm:"column:tags"`
	Reply_Token string      `json:"reply_token" gorm:"column:reply_token"`
}

func (UserCompany) TableName() string {
	return "users"
}

type UserCompany struct {
	Id        uuid.UUID `json:"id" gorm:"column:id"`
	CompanyId uuid.UUID `json:"company_id" gorm:"column:company_id"`
	Username  string    `json:"username" gorm:"column:username"`
}
