package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ResearchSession 研究会话模型
type ResearchSession struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      string    `gorm:"not null;index" json:"user_id"`
	Title       string    `gorm:"not null;size:200" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Status      string    `gorm:"not null;default:'active';size:20" json:"status"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// 关联关系
	Tasks []ResearchTask `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"tasks,omitempty"`
}

// TableName 指定表名
func (ResearchSession) TableName() string {
	return "research_sessions"
}

// BeforeCreate 创建前的钩子
func (rs *ResearchSession) BeforeCreate(tx *gorm.DB) error {
	if rs.ID == uuid.Nil {
		rs.ID = uuid.New()
	}
	return nil
}

// ResearchTask 研究任务模型
type ResearchTask struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionID   uuid.UUID `gorm:"type:uuid;not null;index" json:"session_id"`
	TaskType    string    `gorm:"not null;size:50" json:"task_type"`
	Title       string    `gorm:"not null;size:200" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Status      string    `gorm:"not null;default:'pending';size:20" json:"status"`
	Result      string    `gorm:"type:text" json:"result"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`

	// 关联关系
	Session ResearchSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	ToolCalls []ToolCall `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"tool_calls,omitempty"`
}

// TableName 指定表名
func (ResearchTask) TableName() string {
	return "research_tasks"
}

// BeforeCreate 创建前的钩子
func (rt *ResearchTask) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

// ToolCall 工具调用记录模型
type ToolCall struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID       uuid.UUID `gorm:"type:uuid;not null;index" json:"task_id"`
	ToolName     string    `gorm:"not null;size:100" json:"tool_name"`
	InputData    string    `gorm:"type:jsonb" json:"input_data"`
	OutputData   string    `gorm:"type:jsonb" json:"output_data"`
	ExecutionTime int       `gorm:"default:0" json:"execution_time"`
	Status       string    `gorm:"not null;default:'success';size:20" json:"status"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	// 关联关系
	Task ResearchTask `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}

// TableName 指定表名
func (ToolCall) TableName() string {
	return "tool_calls"
}

// BeforeCreate 创建前的钩子
func (tc *ToolCall) BeforeCreate(tx *gorm.DB) error {
	if tc.ID == uuid.Nil {
		tc.ID = uuid.New()
	}
	return nil
}

// MemoryItem 记忆项模型
type MemoryItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionID uuid.UUID `gorm:"type:uuid;not null;index" json:"session_id"`
	Role      string    `gorm:"not null;size:20" json:"role"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Embedding []float32 `gorm:"type:vector(1536)" json:"embedding"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	// 关联关系
	Session ResearchSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
}

// TableName 指定表名
func (MemoryItem) TableName() string {
	return "memory_items"
}

// BeforeCreate 创建前的钩子
func (mi *MemoryItem) BeforeCreate(tx *gorm.DB) error {
	if mi.ID == uuid.Nil {
		mi.ID = uuid.New()
	}
	return nil
} 