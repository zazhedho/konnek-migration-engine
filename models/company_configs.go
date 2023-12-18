package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (Configuration) TableName() string {
	return "configurations"
}

type Configuration struct {
	Id                            uuid.UUID  `json:"id" gorm:"column:id"`
	ChatLimit                     int        `json:"limit_chat" gorm:"column:limit_chat"`
	SetLimitChat                  bool       `json:"set_limit_chat" gorm:"column:set_limit_chat"`
	CompanyId                     uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	Bot                           bool       `json:"bot" gorm:"column:bot"`
	Whatsapp                      bool       `json:"whatsapp" gorm:"column:whatsapp"`
	Line                          bool       `json:"line" gorm:"column:line"`
	AutoAssign                    bool       `json:"auto_assign" gorm:"column:auto_assign"`
	Telegram                      bool       `json:"telegram" gorm:"column:telegram"`
	Widget                        bool       `json:"widget" gorm:"column:widget"`
	FacebookMessenger             bool       `json:"facebook_messenger" gorm:"column:facebook_messenger"`
	FlagGreeting                  bool       `json:"flag_greeting" gorm:"column:flag_greeting"`
	Greeting                      string     `json:"greeting" gorm:"column:greeting"`
	GreetingOptionsFlag           bool       `json:"greeting_options_flag" gorm:"column:greeting_options_flag"`
	GreetingOptions               string     `json:"greeting_options" gorm:"column:greeting_options"`
	WaitingGreetingFlag           bool       `json:"waiting_greeting_flag" gorm:"column:waiting_greeting_flag"`
	WaitingGreeting               string     `json:"waiting_greeting" gorm:"column:waiting_greeting"`
	AssignedGreetingFlag          bool       `json:"assigned_greeting_flag" gorm:"column:assigned_greeting_flag"`
	AssignedGreeting              string     `json:"assigned_greeting" gorm:"column:assigned_greeting"`
	ClosingGreetingFlag           bool       `json:"closing_greeting_flag" gorm:"column:closing_greeting_flag"`
	ClosingGreeting               string     `json:"closing_greeting" gorm:"column:closing_greeting"`
	SlaFrom                       string     `json:"sla_from" gorm:"column:sla_from"`
	SlaTo                         string     `json:"sla_to" gorm:"column:sla_to"`
	SlaThreshold                  int        `json:"sla_threshold" gorm:"column:sla_threshold"`
	CreatedAt                     time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt                     time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt                     *time.Time `json:"-" gorm:"column:deleted_at"`
	CsatFlag                      bool       `json:"csat_flag" gorm:"column:csat_flag"`
	ReasonFlag                    bool       `json:"reason_flag" gorm:"column:reason_flag"`
	BlackList                     string     `json:"blacklist" gorm:"column:blacklist"`
	KeywordFilterStatus           bool       `json:"keyword_filter_status" gorm:"column:keyword_filter_status"`
	KeywordFilter                 string     `json:"keyword_filter" gorm:"column:keyword_filter"`
	KeywordGreetings              string     `json:"keyword_greetings" gorm:"column:keyword_greetings"`
	UrlWebhook                    string     `json:"url_webhook" gorm:"column:url_webhook"`
	SdkWhatsapp                   string     `json:"sdk_whatsapp" gorm:"column:sdk_whatsapp"`
	InquirySandeza                bool       `json:"inquiry_sandeza" gorm:"column:inquiry_sandeza"`
	MaintenanceStatus             bool       `json:"maintenance_status" gorm:"column:maintenance_status"`
	MaintenanceMessage            string     `json:"maintenance_message" gorm:"column:maintenance_message"`
	KeywordGreetingsLimit         int        `json:"keyword_greetings_limit" gorm:"column:keyword_greetings_limit"`
	KeywordGreetingsLimitDuration int        `json:"keyword_greetings_limit_duration" gorm:"column:keyword_greetings_limit_duration"`
	Error                         string     `json:"error" gorm:"-"`
}

func (ConfigurationApp) TableName() string {
	return "configurations"
}

type ConfigurationApp struct {
	Id                   uuid.UUID  `json:"id" gorm:"column:id"`
	ChatLimit            int        `json:"limit_chat" gorm:"column:limit_chat"`
	SetLimitChat         bool       `json:"set_limit_chat" gorm:"column:set_limit_chat"`
	CompanyId            uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	Bot                  bool       `json:"bot" gorm:"column:bot"`
	Whatsapp             bool       `json:"whatsapp" gorm:"column:whatsapp"`
	Line                 bool       `json:"line" gorm:"column:line"`
	AutoAssign           bool       `json:"auto_assign" gorm:"column:auto_assign"`
	Telegram             bool       `json:"telegram" gorm:"column:telegram"`
	Widget               bool       `json:"widget" gorm:"column:widget"`
	FacebookMessenger    bool       `json:"facebook_messenger" gorm:"column:facebook_messenger"`
	FlagGreeting         bool       `json:"flag_greeting" gorm:"column:flag_greeting"`
	Greeting             string     `json:"greeting" gorm:"column:greeting"`
	GreetingOptionsFlag  bool       `json:"greeting_options_flag" gorm:"column:greeting_options_flag"`
	GreetingOptions      string     `json:"greeting_options" gorm:"column:greeting_options"`
	WaitingGreetingFlag  bool       `json:"waiting_greeting_flag" gorm:"column:waiting_greeting_flag"`
	WaitingGreeting      string     `json:"waiting_greeting" gorm:"column:waiting_greeting"`
	AssignedGreetingFlag bool       `json:"assigned_greeting_flag" gorm:"column:assigned_greeting_flag"`
	AssignedGreeting     string     `json:"assigned_greeting" gorm:"column:assigned_greeting"`
	ClosingGreetingFlag  bool       `json:"closing_greeting_flag" gorm:"column:closing_greeting_flag"`
	ClosingGreeting      string     `json:"closing_greeting" gorm:"column:closing_greeting"`
	SlaFrom              string     `json:"sla_from" gorm:"column:sla_from"`
	SlaTo                string     `json:"sla_to" gorm:"column:sla_to"`
	SlaThreshold         int        `json:"sla_threshold" gorm:"column:sla_threshold"`
	CreatedAt            time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt            time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt            *time.Time `json:"-" gorm:"column:deleted_at"`
	CsatFlag             bool       `json:"csat_flag" gorm:"column:csat_flag"`
	ReasonFlag           bool       `json:"reason_flag" gorm:"column:reason_flag"`
	BlackList            string     `json:"blacklist" gorm:"column:blacklist"`
	KeywordFilterStatus  bool       `json:"keyword_filter_status" gorm:"column:keyword_filter_status"`
	KeywordFilter        string     `json:"keyword_filter" gorm:"column:keyword_filter"`
	KeywordGreetings     string     `json:"keyword_greetings" gorm:"column:keyword_greetings"`
	UrlWebhook           string     `json:"url_webhook" gorm:"column:url_webhook"`
	SdkWhatsapp          string     `json:"sdk_whatsapp" gorm:"column:sdk_whatsapp"`
	InquirySandeza       bool       `json:"inquiry_sandeza" gorm:"column:inquiry_sandeza"`
	MaxChat              int        `json:"max_chat" gorm:"column:max_chat"`
	IntervalChat         int        `json:"interval_chat" gorm:"column:interval_chat"`
	BlockDuration        int        `json:"block_duration" gorm:"column:block_duration"`
	Error                string     `json:"error" gorm:"-"`
}

func (CompanyConfig) TableName() string {
	return "company_configs"
}

type CompanyConfig struct {
	Id                        uuid.UUID  `json:"id" gorm:"column:id"`
	CompanyId                 uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	Bot                       bool       `json:"bot" gorm:"column:bot"`
	Whatsapp                  bool       `json:"whatsapp" gorm:"column:whatsapp"`
	Line                      bool       `json:"line" gorm:"column:line"`
	Telegram                  bool       `json:"telegram" gorm:"column:telegram"`
	Facebook                  bool       `json:"facebook" gorm:"column:facebook"`
	WebWidget                 bool       `json:"web_widget" gorm:"column:web_widget"`
	AppWidget                 bool       `json:"app_widget" gorm:"column:app_widget"`
	SdkBot                    string     `json:"sdk_bot" gorm:"column:sdk_bot"`
	SdkWhatsapp               string     `json:"sdk_whatsapp" gorm:"column:sdk_whatsapp"`
	SdkLine                   string     `json:"sdk_line" gorm:"column:sdk_line"`
	SdkTelegram               string     `json:"sdk_telegram" gorm:"column:sdk_telegram"`
	SdkFacebook               string     `json:"sdk_facebook" gorm:"column:sdk_facebook"`
	AutoAssign                bool       `json:"auto_assign" gorm:"column:auto_assign"`
	ChatLimit                 int        `json:"chat_limit" gorm:"column:chat_limit"`
	SlaFrom                   string     `json:"sla_from" gorm:"column:sla_from"`
	SlaTo                     string     `json:"sla_to" gorm:"column:sla_to"`
	SlaThreshold              int        `json:"sla_threshold" gorm:"column:sla_threshold"`
	WelcomeGreeting           string     `json:"welcome_greeting" gorm:"column:welcome_greeting"`
	WelcomeGreetingFlag       bool       `json:"welcome_greeting_flag" gorm:"column:welcome_greeting_flag"`
	WelcomeGreetingOption     string     `json:"welcome_greeting_option" gorm:"column:welcome_greeting_options"`
	WelcomeGreetingOptionFlag bool       `json:"welcome_greeting_option_flag" gorm:"column:welcome_greeting_options_flag"`
	WaitingGreeting           string     `json:"waiting_greeting" gorm:"column:waiting_greeting"`
	WaitingGreetingFlag       bool       `json:"waiting_greeting_flag" gorm:"column:waiting_greeting_flag"`
	AssignedGreeting          string     `json:"assigned_greeting" gorm:"column:assigned_greeting"`
	AssignedGreetingFlag      bool       `json:"assigned_greeting_flag" gorm:"column:assigned_greeting_flag"`
	ClosingGreeting           string     `json:"closing_greeting" gorm:"column:closing_greeting"`
	ClosingGreetingFlag       bool       `json:"closing_greeting_flag" gorm:"column:closing_greeting_flag"`
	CsatFlag                  bool       `json:"csat_flag" gorm:"column:csat_flag"`
	UnavailableReasonFlag     bool       `json:"unavailable_reason_flag" gorm:"column:unavailable_reason_flag"`
	InquirySandeza            bool       `json:"inquiry_sandeza" gorm:"column:inquiry_sandeza"`
	Blacklist                 string     `json:"blacklist" gorm:"column:blacklist"`
	KeywordFilterStatus       bool       `json:"keyword_filter_status" gorm:"column:keyword_filter_status"`
	KeywordFilter             string     `json:"keyword_filter" gorm:"column:keyword_filter"`
	KeywordGreetings          string     `json:"keyword_greetings" gorm:"column:keyword_greetings"`
	MaintenanceStatus         bool       `json:"maintenance_status" gorm:"column:maintenance_status"`
	MaintenanceMessage        string     `json:"maintenance_message" gorm:"column:maintenance_message"`
	AutoAssignPersession      bool       `json:"auto_assign_persession" gorm:"column:auto_assign_persession"`
	SpamMaxChat               int        `json:"spam_max_chat" gorm:"column:spam_max_chat"`
	SpamIntervalChat          int        `json:"spam_interval_chat" gorm:"column:spam_interval_chat"`   //in seconds
	SpamBlockDuration         int        `json:"spam_block_duration" gorm:"column:spam_block_duration"` //in minutes
	SpamPerContentType        bool       `json:"spam_per_content_type" gorm:"column:spam_per_content_type"`
	SpamMessage               string     `json:"spam_message" gorm:"column:spam_message"`
	KeywordMaxInvalid         int        `json:"keyword_max_invalid" gorm:"column:keyword_max_invalid"`
	KeywordInterval           int        `json:"keyword_interval" gorm:"column:keyword_interval"`             //in seconds
	KeywordBlockDuration      int        `json:"keyword_block_duration" gorm:"column:keyword_block_duration"` //in minutes
	CreatedAt                 time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy                 uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt                 time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy                 uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt                 *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy                 uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
}
