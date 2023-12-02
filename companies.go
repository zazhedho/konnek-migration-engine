package main

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"konnek-migration/utils"
	"net/http"
	"os"
	"time"
)

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

func main() {
	// Create source DB connection
	scDB := utils.GetDBConnection(os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_PORT"), os.Getenv("USERNAME_DB"), os.Getenv("DATABASE_NAME"), os.Getenv("PASSWORD_DB"))
	defer func(scDB *gorm.DB) {
		err := scDB.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(scDB)

	// Create destination DB connection
	dstDB := utils.GetDBConnection(os.Getenv("RE_DATABASE_HOST"), os.Getenv("RE_DATABASE_PORT"), os.Getenv("RE_USERNAME_DB"), os.Getenv("RE_DATABASE_NAME"), os.Getenv("RE_PASSWORD_DB"))
	defer func(dstDB *gorm.DB) {
		err := dstDB.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(dstDB)

	if err := scDB.Order("created_at", order).Limit(perpage).Offset(offset).Find(&company).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			sentry.CaptureException(err)
		}
		return []models.Company{}, http.StatusBadRequest, errors.New("failed to fetch company list" + err.Error())
	}
}
