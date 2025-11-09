package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes" // ⭐ 1. 导入 gorm/datatypes 包
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username  string         `gorm:"uniqueIndex;not null" json:"username"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	Nickname  string         `gorm:"size:100" json:"nickname"`
	Avatar    string         `gorm:"size:255" json:"avatar"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ChatMessage 聊天消息模型
type ChatMessage struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	ConversationID uuid.UUID `gorm:"type:uuid;index;not null" json:"conversation_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Role      string         `gorm:"size:20;not null" json:"role"`     // "user" or "assistant"
	MessageID uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"message_id"`
	Metadata  datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ChatSession 聊天会话模型
type ChatSession struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	User        User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Title       string         `gorm:"size:200" json:"title"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	Messages    []ChatMessage  `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserPreference 用户偏好设置模型
type UserPreference struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	User          User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	LLMModel      string    `gorm:"size:100;default:'gpt-3.5-turbo'" json:"llm_model"`
	Temperature   float32   `gorm:"default:0.7" json:"temperature"`
	MaxTokens     int       `gorm:"default:2000" json:"max_tokens"`
	SystemPrompt  string    `gorm:"type:text" json:"system_prompt,omitempty"`
	ContextWindow int       `gorm:"default:10" json:"context_window"` // 上下文窗口大小
	MemoryEnabled bool      `gorm:"default:true" json:"memory_enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// RefreshToken 刷新令牌模型
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Token     string    `gorm:"size:500;not null;uniqueIndex" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsRevoked bool      `gorm:"default:false" json:"is_revoked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
