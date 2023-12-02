package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func (Rooms) TableName() string {
	return "rooms"
}

type Rooms struct {
	ID             uuid.UUID  `json:"id" gorm:"column:id"`
	CompanyID      uuid.UUID  `json:"company_id" gorm:"column:company_id"`
	ChannelCode    string     `json:"channel_code" gorm:"column:channel_code"`
	CustomerUserID uuid.UUID  `json:"customer_user_id" gorm:"column:customer_user_id"`
	LastSessionID  uuid.UUID  `json:"last_session_id" gorm:"column:last_session_id"`
	CreatedAt      time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy      uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy      uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt      *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy      uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
}

type FetchRoom struct {
	CompanyId      uuid.UUID `json:"company_id" gorm:"column:company_id"`
	DivisionId     uuid.UUID `json:"division_id" gorm:"column:division_id"`
	Id             uuid.UUID `json:"room_id" gorm:"column:id"`
	SessionId      uuid.UUID `json:"session_id" gorm:"column:session_id"`
	ChannelCode    string    `json:"channel_code" gorm:"column:channel_code"`
	CustomerUserId uuid.UUID `json:"customer_user_id" gorm:"column:customer_user_id"`
	AgentUserId    uuid.UUID `json:"agent_user_id" gorm:"column:agent_user_id"`
	SeqId          int64     `json:"seq_id" gorm:"column:seq_id"`
	LastMessageId  uuid.UUID `json:"last_chat_message_id" gorm:"column:last_chat_message_id"`
	Categories     string    `json:"categories" gorm:"column:categories"`
	BotStatus      bool      `json:"bot_status" gorm:"column:bot_status"`
	Status         int       `json:"status" gorm:"column:status"`

	OpenTime          *time.Time `json:"open_time" gorm:"column:open_time"`
	QueTime           *time.Time `json:"queue_time" gorm:"column:queue_time"`
	AssignTime        *time.Time `json:"assign_time" gorm:"column:assign_time"`
	FirstResponseTime *time.Time `json:"first_response_time" gorm:"column:first_response_time"`
	LastAgentChatTime *time.Time `json:"last_agent_chat_time" gorm:"column:last_agent_chat_time"`
	CloseTime         *time.Time `json:"close_time" gorm:"column:close_time"`

	CustomerName     string `json:"customer_name" gorm:"customer_name"`
	CustomerUsername string `json:"customer_username" gorm:"customer_username"`
	AgentName        string `json:"agent_name" gorm:"agent_name"`
	AgentUsername    string `json:"agent_username" gorm:"agent_username"`

	ChatMessageId    string     `json:"chat_message_id" gorm:"chat_message_id"`
	MessageId        string     `json:"message_id" gorm:"column:message_id"`
	ReplyId          string     `json:"reply_id" gorm:"column:message_reply_id"`
	FromType         string     `json:"from_type" gorm:"column:message_from_type"`
	Type             string     `json:"type" gorm:"column:message_type"`
	Text             string     `json:"text" gorm:"column:message_text"`
	Payload          string     `json:"payload" gorm:"column:message_payload"`
	MessageStatus    int        `json:"message_status" gorm:"column:message_status"`
	MessageTime      *time.Time `json:"message_time" gorm:"column:message_time"`
	MessageCreatedAt time.Time  `json:"message_created_at" gorm:"column:message_created_at"`

	ConversationSeqId int64 `json:"conversation_seq_id" gorm:"conversation_seq_id"`
}
