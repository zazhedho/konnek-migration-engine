package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

// Status
const (
	SessionHandovered = -2
	SessionClosed     = -1
	SessionOpen       = 0
	SessionWaiting    = 1
	SessionAssigned   = 2
)

const (
	SlaSuccess = "Success"
	SlaFailed  = "Failed"
)

func (m ReportSession) TableName() string {
	return m.TablePrefix + "session"
}

type ReportSession struct {
	TablePrefix string `gorm:"-"`

	Id uuid.UUID `json:"id" gorm:"column:id"`

	CompanyId   uuid.UUID `json:"company_id" gorm:"column:company_id"`
	CompanyCode string    `json:"company_code" gorm:"column:company_code"`
	CompanyName string    `json:"company_name" gorm:"column:company_name"`

	CustomerId       uuid.UUID `json:"customer_id" gorm:"column:customer_id"`
	CustomerUsername string    `json:"customer_username" gorm:"column:customer_username"`
	CustomerName     string    `json:"customer_name" gorm:"column:customer_name"`
	CustomerTags     string    `json:"customer_tags" gorm:"column:customer_tags"`
	Channel          string    `json:"channel" gorm:"column:channel"`
	RoomId           uuid.UUID `json:"room_id" gorm:"column:room_id"`

	DivisionId   uuid.UUID `json:"division_id" gorm:"column:division_id"`
	DivisionName string    `json:"division_name" gorm:"column:division_name"`

	AgentUserId   uuid.UUID `json:"agent_id" gorm:"column:agent_id"`
	AgentUsername string    `json:"agent_username" gorm:"column:agent_username"`
	AgentName     string    `json:"agent_name" gorm:"column:agent_name"`

	Categories string `json:"categories" gorm:"column:categories"`
	BotStatus  bool   `json:"bot_status" gorm:"bot_status"`
	Status     int    `json:"status" gorm:"column:status"`

	OpenTime   *time.Time `json:"open_time" gorm:"column:open_time"`
	QueTime    *time.Time `json:"queue_time" gorm:"column:queue_time"`
	AssignTime *time.Time `json:"assign_time" gorm:"column:assign_time"`
	FrTime     *time.Time `json:"fr_time" gorm:"column:fr_time"`
	LrTime     *time.Time `json:"lr_time" gorm:"column:lr_time"`
	CloseTime  *time.Time `json:"close_time" gorm:"column:close_time"`

	WaitingDuration int64 `json:"waiting_duration" gorm:"column:waiting_duration"` //assign - queue
	FrDuration      int64 `json:"fr_duration" gorm:"column:fr_duration"`           //fr - assign
	ResolveDuration int64 `json:"resolve_duration" gorm:"column:resolve_duration"` // close - assign
	SessionDuration int64 `json:"session_duration" gorm:"column:session_duration"` // close - open

	SlaFrom      string `json:"sla_from" gorm:"columns:sla_from"`
	SlaTo        string `json:"sla_to" gorm:"column:sla_to"`
	SlaThreshold int64  `json:"sla_threshold" gorm:"column:sla_threshold"`
	SlaDuration  int64  `json:"sla_duration" gorm:"column:sla_duration"`
	SlaStatus    string `json:"sla_status" gorm:"column:sla_status"`

	OpenBy       uuid.UUID `json:"open_by" gorm:"column:open_by"`
	OpenUsername string    `json:"open_username" gorm:"column:open_username"`
	OpenName     string    `json:"open_name" gorm:"column:open_name"`

	HandoverBy       uuid.UUID `json:"handover_by" gorm:"column:handover_by"`
	HandoverUsername string    `json:"handover_username" gorm:"column:handover_username"`
	HandoverName     string    `json:"handover_name" gorm:"column:handover_name"`

	CloseBy       uuid.UUID `json:"close_by" gorm:"column:close_by"`
	CloseUsername string    `json:"close_username" gorm:"column:close_username"`
	CloseName     string    `json:"close_name" gorm:"column:close_name"`

	LastUpdate time.Time `json:"last_update" gorm:"column:last_update"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
	CreatedBy  string    `json:"created_by" gorm:"column:created_by"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy  string    `json:"updated_by" gorm:"column:updated_by"`
}

func (FetchReportSession) TableName() string {
	return "sessions"
}

type FetchReportSession struct {
	Id uuid.UUID `json:"id" gorm:"column:id"`
	//SeqId       int64     `json:"seq_id" gorm:"AUTO_INCREMENT"`

	RoomId uuid.UUID       `json:"room_id" gorm:"column:room_id"`
	Room   FetchReportRoom `json:"room" gorm:"Foreignkey:RoomId;association_foreignkey:Id;"`

	DivisionId uuid.UUID     `json:"division_id" gorm:"column:division_id"`
	Division   DivisionReeng `json:"division" gorm:"Foreignkey:DivisionId;association_foreignkey:Id;"`

	AgentUserId uuid.UUID `json:"agent_user_id" gorm:"column:agent_user_id"`
	Agent       Users     `json:"agent" gorm:"Foreignkey:AgentUserId;association_foreignkey:Id;"`

	//LastMessageId uuid.UUID `json:"last_chat_message_id" gorm:"column:last_chat_message_id"`
	Categories string `json:"categories" gorm:"column:categories"`

	BotStatus bool `json:"bot_status" gorm:"bot_status"`
	Status    int  `json:"status" gorm:"column:status"`

	OpenBy     uuid.UUID `json:"open_by" gorm:"column:open_by"`
	UserOpenBy Users     `json:"user_open_by" gorm:"Foreignkey:OpenBy;association_foreignkey:Id;"`

	HandoverBy     uuid.UUID `json:"handover_by" gorm:"column:handover_by"`
	UserHandoverBy Users     `json:"user_handover_by" gorm:"Foreignkey:HandoverBy;association_foreignkey:Id;"`

	CloseBy     uuid.UUID `json:"close_by" gorm:"column:close_by"`
	UserCloseBy Users     `json:"user_close_by" gorm:"Foreignkey:CloseBy;association_foreignkey:Id;"`

	SlaFrom      string `json:"sla_from" gorm:"columns:sla_from"`
	SlaTo        string `json:"sla_to" gorm:"column:sla_to"`
	SlaTreshold  int    `json:"sla_treshold" gorm:"column:sla_threshold"`
	SlaDurations int    `json:"sla_durations" gorm:"column:sla_durations"`
	SlaStatus    string `json:"sla_status" gorm:"column:sla_status"`

	OpenTime          *time.Time `json:"open_time" gorm:"column:open_time"`
	QueTime           *time.Time `json:"queue_time" gorm:"column:queue_time"`
	AssignTime        *time.Time `json:"assign_time" gorm:"column:assign_time"`
	FirstResponseTime *time.Time `json:"first_response_time" gorm:"column:first_response_time"`
	LastAgentChatTime *time.Time `json:"last_agent_chat_time" gorm:"column:last_agent_chat_time"`
	CloseTime         *time.Time `json:"close_time" gorm:"column:close_time"`

	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`

	Error string `json:"error" gorm:"-"`
}

func (FetchReportRoom) TableName() string {
	return "rooms"
}

type FetchReportRoom struct {
	Id uuid.UUID `json:"id" gorm:"column:id"`

	CompanyId uuid.UUID    `json:"company_id" gorm:"column:company_id"`
	Company   CompanyReeng `json:"company" gorm:"Foreignkey:CompanyId;association_foreignkey:Id;"`

	CustomerUserId uuid.UUID `json:"customer_user_id" gorm:"column:customer_user_id"`
	Customer       Users     `json:"customer" gorm:"Foreignkey:CustomerUserId;association_foreignkey:Id;"`

	ChannelCode string `json:"channel_code" gorm:"column:channel_code"`
}
