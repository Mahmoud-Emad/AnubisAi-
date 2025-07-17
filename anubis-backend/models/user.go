package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null" validate:"required,min=3,max=50"`
	Password  string    `json:"-" gorm:"not null" validate:"required,min=8"`
	FirstName string    `json:"first_name" gorm:"not null" validate:"required,min=1,max=100"`
	LastName  string    `json:"last_name" gorm:"not null" validate:"required,min=1,max=100"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	IsAdmin   bool      `json:"is_admin" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Settings []UserSetting `json:"settings,omitempty" gorm:"foreignKey:UserID"`
	Memories []UserMemory  `json:"memories,omitempty" gorm:"foreignKey:UserID"`
	Tasks    []TaskExecution `json:"tasks,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate hook to set UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserSetting represents user-specific settings
type UserSetting struct {
	ID        uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:char(36);not null;index"`
	Key       string    `json:"key" gorm:"not null" validate:"required"`
	Value     string    `json:"value" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User User `json:"-" gorm:"foreignKey:UserID"`
}

// BeforeCreate hook to set UUID
func (us *UserSetting) BeforeCreate(tx *gorm.DB) error {
	if us.ID == uuid.Nil {
		us.ID = uuid.New()
	}
	return nil
}

// UserMemory represents AI memories for a user
type UserMemory struct {
	ID        uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:char(36);not null;index"`
	Title     string    `json:"title" gorm:"not null" validate:"required,max=200"`
	Content   string    `json:"content" gorm:"type:text;not null" validate:"required"`
	Tags      string    `json:"tags" gorm:"type:text"` // JSON array as string
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User User `json:"-" gorm:"foreignKey:UserID"`
}

// BeforeCreate hook to set UUID
func (um *UserMemory) BeforeCreate(tx *gorm.DB) error {
	if um.ID == uuid.Nil {
		um.ID = uuid.New()
	}
	return nil
}

// TaskExecution represents a task execution record
type TaskExecution struct {
	ID          uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:char(36);index"`
	TaskName    string    `json:"task_name" gorm:"not null" validate:"required"`
	Parameters  string    `json:"parameters" gorm:"type:text"` // JSON as string
	Response    string    `json:"response" gorm:"type:text"`   // JSON as string
	Status      string    `json:"status" gorm:"not null;default:'pending'" validate:"required,oneof=pending running success failed"`
	ErrorMsg    string    `json:"error_message,omitempty" gorm:"type:text"`
	Duration    int64     `json:"duration_ms,omitempty"` // Duration in milliseconds
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate hook to set UUID
func (te *TaskExecution) BeforeCreate(tx *gorm.DB) error {
	if te.ID == uuid.Nil {
		te.ID = uuid.New()
	}
	return nil
}

// PasswordReset represents password reset tokens
type PasswordReset struct {
	ID        uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:char(36);not null;index"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User User `json:"-" gorm:"foreignKey:UserID"`
}

// BeforeCreate hook to set UUID
func (pr *PasswordReset) BeforeCreate(tx *gorm.DB) error {
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	return nil
}

// IsExpired checks if the password reset token is expired
func (pr *PasswordReset) IsExpired() bool {
	return time.Now().After(pr.ExpiresAt)
}

// IsUsed checks if the password reset token has been used
func (pr *PasswordReset) IsUsed() bool {
	return pr.UsedAt != nil
}
