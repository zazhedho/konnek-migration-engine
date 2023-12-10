package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

const (
	MessageList     = "list"
	MessageButton   = "button"
	MessageCarousel = "carousel"
	MessageLocation = "location"
	MessagePostback = "postback"
	MessageTemplate = "template"
)

const (
	TemplateList        = "list_reply"
	TemplateButton      = "button"
	TemplateButtonReply = "button_reply"
	TemplateCarousel    = "carousel"
	TemplateLocation    = "location"
)

var MessageTypeMap = map[string]string{
	TemplateList:        MessageList,
	TemplateButton:      MessageButton,
	TemplateButtonReply: MessageList,
	TemplateCarousel:    MessageCarousel,
	TemplateLocation:    MessageLocation,
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

// ChatMessage struct db existing
type ChatMessage struct {
	Id          uuid.UUID    `json:"id" gorm:"column:id"`
	Message     string       `json:"message" gorm:"column:message"`
	UserId      uuid.UUID    `json:"user_id" gorm:"column:user_id"`
	User        UserCompany  `json:"user" gorm:"foreignKey:UserId;AssociationForeignKey:Id;"`
	FromType    int          `json:"from_type" gorm:"column:from_type"`
	RoomId      uuid.UUID    `json:"room_id" gorm:"column:room_id"`
	ChatMedia   ChatMediaUrl `json:"chat_media" gorm:"foreignKey:ChatMessageId;AssociationForeignKey:Id;"`
	SessionId   uuid.UUID    `json:"session_id" gorm:"session_id"`
	TrxId       string       `json:"trx_id" gorm:"trx_id"`
	Status      int          `json:"status" gorm:"status"`
	CreatedAt   time.Time    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt   *time.Time   `json:"-" gorm:"column:deleted_at"`
	MsgId       string       `json:"msg_id" gorm:"column:msg_id"`
	MessageTime time.Time    `json:"message_time" gorm:"column:message_time"`
	Error       string       `json:"error" gorm:"-"`
}

func (ChatMediaUrl) TableName() string {
	return "chat_media"
}

type ChatMediaUrl struct {
	ChatMessageId uuid.UUID `json:"chat_message_id" gorm:"column:chat_message_id"`
	Media         string    `json:"media" gorm:"column:media"`
}

func (ChatMessages) TableName() string {
	return "chat_messages"
}

// ChatMessages struct db reengineering
type ChatMessages struct {
	Id                uuid.UUID  `json:"id" gorm:"column:id"`
	RoomId            uuid.UUID  `json:"room_id" gorm:"column:room_id"`
	SessionId         uuid.UUID  `json:"session_id" gorm:"column:session_id"`
	UserId            uuid.UUID  `json:"user_id" gorm:"column:user_id"`
	MessageId         string     `json:"message_id" gorm:"column:message_id"`
	ReplyId           string     `json:"reply_id" gorm:"column:reply_id"`
	ProviderMessageId string     `json:"provider_message_id" gorm:"column:provider_message_id"`
	FromType          string     `json:"from_type" gorm:"column:from_type"`
	Type              string     `json:"type" gorm:"column:type"`
	Text              string     `json:"text" gorm:"column:text"`
	Payload           string     `json:"payload" gorm:"column:payload"`
	Status            int        `json:"status" gorm:"column:status"`
	MessageTime       *time.Time `json:"message_time" gorm:"column:message_time"`
	ReceivedAt        *time.Time `json:"received_at" gorm:"column:received_at"`
	ProcessedAt       *time.Time `json:"processed_at" gorm:"column:processed_at"`
	CreatedAt         time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy         uuid.UUID  `json:"created_by" gorm:"column:created_by"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy         uuid.UUID  `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt         *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy         uuid.UUID  `json:"deleted_by" gorm:"column:deleted_by"`
	TextDeleted       string     `json:"-" gorm:"column:text_deleted"`
	DeleteTime        *time.Time `json:"-" gorm:"column:delete_time"`
	Retry             int        `json:"retry_sending" gorm:"column:retry"`
	RetryTime         *time.Time `json:"retry_time" gorm:"column:retry_time"`
}

type ButtonMessage struct {
	Body   []Action      `json:"body"`
	Header HeaderMessage `json:"header"`
	Footer FooterMessage `json:"footer"`
}

type ListMessage struct {
	Body struct {
		Sections []Section `json:"sections"`
		Title    string    `json:"title"`
	} `json:"body"`
	Header HeaderMessage `json:"header"`
	Footer FooterMessage `json:"footer"`
}

type TemplateMessage struct {
	Body struct {
		Text string `json:"text"`
	} `json:"body"`
	Header HeaderMessage `json:"header"`
	Footer FooterMessage `json:"footer"`
}

type MediaMessage struct {
	Id   string `json:"id"`
	Url  string `json:"url"`
	Name string `json:"name"`
	Size int    `json:"size"` //in kb
}

type PostbackMessage struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type LocationMessage struct {
	Lat        float64 `json:"latitude"`
	Lng        float64 `json:"longitude"`
	Address    string  `json:"address"`
	Name       string  `json:"name"`
	LivePeriod int     `json:"live_period"`
}

type CarouselMessage struct {
	Body   []BodyCarousel `json:"body"`
	Header HeaderMessage  `json:"header"`
	Footer FooterMessage  `json:"footer"`
}

type BodyCarousel struct {
	Actions     []Action `json:"actions"`
	Description string   `json:"description"`
	Title       string   `json:"title"`
	MediaUrl    string   `json:"media_url"`
}

type Action struct {
	Description string `json:"description"`
	Key         string `json:"key"`
	MediaURL    string `json:"media_url"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	Value       string `json:"value"`
}

type Section struct {
	Actions     []Action `json:"actions"`
	Description string   `json:"description"`
	MediaURL    string   `json:"media_url"`
	Title       string   `json:"title"`
}

type HeaderMessage struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	MediaUrl string `json:"media_url"`
}

type FooterMessage struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}
