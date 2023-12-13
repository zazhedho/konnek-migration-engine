package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

func (Session) TableName() string {
	return "sessions"
}

type Session struct {
	Id     uuid.UUID `json:"id" gorm:"column:id"`
	RoomId uuid.UUID `json:"room_id" gorm:"column:room_id"`
	//RoomDetail        RoomDetails   `json:"room_detail" gorm:"foreignKey:RoomId;AssociationForeignKey:Id;"`
	ChatMessage       ChatMessageId `json:"chat_message" gorm:"foreignKey:SessionId;AssociationForeignKey:Id;"`
	OpenBy            uuid.UUID     `json:"open_by" gorm:"column:open_by"`
	CloseBy           uuid.UUID     `json:"close_by" gorm:"column:close_by"`
	HandOverBy        uuid.UUID     `json:"hand_over_by" gorm:"column:hand_over_by"`
	CreatedAt         time.Time     `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time     `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt         *time.Time    `json:"deleted_at" gorm:"column:deleted_at"`
	Categories        string        `json:"categories" gorm:"column:categories"`
	QueTime           time.Time     `json:"queue_time" gorm:"column:queue_time"`
	AssignTime        time.Time     `json:"assign_time" gorm:"column:assign_time"`
	FirstResponseTime time.Time     `json:"first_response_time" gorm:"column:first_response_time"`
	LastAgentChatTime time.Time     `json:"last_agent_chat_time" gorm:"column:last_agent_chat_time"`
	CloseTime         time.Time     `json:"close_time" gorm:"column:close_time"`
	SlaFrom           string        `json:"sla_from" gorm:"columns:sla_from"`
	SlaTo             string        `json:"sla_to" gorm:"column:sla_to"`
	SlaTreshold       int           `json:"sla_treshold" gorm:"column:sla_threshold"`
	SlaDurations      int           `json:"sla_durations" gorm:"column:sla_durations"`
	SlaStatus         string        `json:"sla_status" gorm:"column:sla_status"`
	AgentUserId       uuid.UUID     `json:"agent_user_id" gorm:"column:agent_user_id"`
	DivisionId        uuid.UUID     `json:"division_id" gorm:"column:division_id"`
	BotStatus         bool          `json:"bot_status" gorm:"bot_status"`
	Error             string        `json:"error" gorm:"-"`
}

func (Sessions) TableName() string {
	return "sessions"
}

type Sessions struct {
	Id          uuid.UUID `json:"id" gorm:"column:id"`
	SeqId       int64     `json:"seq_id" gorm:"AUTO_INCREMENT"`
	RoomId      uuid.UUID `json:"room_id" gorm:"column:room_id"`
	DivisionId  uuid.UUID `json:"division_id" gorm:"column:division_id"`
	AgentUserId uuid.UUID `json:"agent_user_id" gorm:"column:agent_user_id"`

	LastMessageId uuid.UUID `json:"last_chat_message_id" gorm:"column:last_chat_message_id"`
	Categories    string    `json:"categories" gorm:"column:categories"`

	BotStatus bool `json:"bot_status" gorm:"bot_status"`
	Status    int  `json:"status" gorm:"column:status"`

	OpenBy     uuid.UUID `json:"open_by" gorm:"column:open_by"`
	HandoverBy uuid.UUID `json:"handover_by" gorm:"column:handover_by"`
	CloseBy    uuid.UUID `json:"close_by" gorm:"column:close_by"`

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
}

func (ChatMessageId) TableName() string {
	return "chat_messages"
}

type ChatMessageId struct {
	Id        uuid.UUID `json:"id" gorm:"column:id"`
	SessionId uuid.UUID `json:"session_id" gorm:"column:session_id"`
}
