package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (CompanyExist) TableName() string {
	return "companies"
}

type CompanyExist struct {
	Id                   uuid.UUID  `json:"id" gorm:"column:id"`
	Name                 string     `json:"name" gorm:"column:name"`
	Status               bool       `json:"status" gorm:"column:status"`
	CompanyCode          string     `json:"company_code" gorm:"column:company_code"`
	LimitUser            int        `json:"limit_user" gorm:"column:limit_user"`
	Email                string     `json:"email" gorm:"column:email"`
	StartPeriod          time.Time  `json:"start_period" gorm:"column:start_period"`
	EndPeriod            time.Time  `json:"end_period" gorm:"column:end_period"`
	CreatedAt            time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt            time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt            *time.Time `json:"-" gorm:"column:deleted_at"`
	CorporateIdSandeza   string     `json:"corporate_id_sandeza" gorm:"column:corporate_id_sandeza"`
	CorporateNameSandeza string     `json:"corporate_name_sandeza" gorm:"column:corporate_name_sandeza"`
	DivisionIdSandeza    string     `json:"division_id_sandeza" gorm:"column:division_id_sandeza"`
	DivisionNameSandeza  string     `json:"division_name_sandeza" gorm:"column:division_name_sandeza"`
	UsernameCGW          string     `json:"username_cgw" gorm:"column:username_cgw"`
	PasswordCGW          string     `json:"password_cgw" gorm:"column:password_cgw"`
	WabaIdSandeza        string     `json:"waba_id_sandeza" gorm:"column:waba_id_sandeza"`
	SenderIdSandeza      string     `json:"sender_id_sandeza" gorm:"column:sender_id_sandeza"`
	Error                string     `json:"error" gorm:"-"`
}

func (CompanyReeng) TableName() string {
	return "companies"
}

type CompanyReeng struct {
	Id           uuid.UUID  `json:"id" gorm:"column:id"`
	Name         string     `json:"name" gorm:"column:name"`
	Status       bool       `json:"status" gorm:"column:status"`
	Code         string     `json:"code" gorm:"column:code"`
	LimitUser    int        `json:"user_limit" gorm:"column:user_limit"`
	Email        string     `json:"email" gorm:"column:email"`
	StartPeriod  time.Time  `json:"start_period" gorm:"column:start_period"`
	EndPeriod    time.Time  `json:"end_period" gorm:"column:end_period"`
	Hosts        string     `json:"hosts" gorm:"column:hosts"`
	CreatedAt    time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy    uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy    uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt    *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy    uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
	BoClientCode string     `json:"bo_client_code" gorm:"column:bo_client_code"`
	BoDivCode    string     `json:"bo_div_code" gorm:"column:bo_div_code"`
}
