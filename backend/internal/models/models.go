package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Role is the user's access level.
type Role string

const (
	RoleMember Role = "member"
	RoleAdmin  Role = "admin"
)

// CategoryType groups categories into the four budgeting buckets.
type CategoryType string

const (
	CategoryFixed    CategoryType = "fixed"
	CategoryVariable CategoryType = "variable"
	CategoryWants    CategoryType = "wants"
	CategoryDebts    CategoryType = "debts"
)

// ValidCategoryType reports whether t is one of the four allowed buckets.
func ValidCategoryType(t CategoryType) bool {
	switch t {
	case CategoryFixed, CategoryVariable, CategoryWants, CategoryDebts:
		return true
	}
	return false
}

// SavingsTargetType selects how the global savings target is interpreted.
type SavingsTargetType string

const (
	SavingsPercent SavingsTargetType = "percent"
	SavingsFixed   SavingsTargetType = "fixed"
)

// User is an application account.
type User struct {
	ID            string          `gorm:"type:char(36);primaryKey" json:"id"`
	Email         string          `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash  string          `gorm:"type:varchar(255);not null" json:"-"`
	Role          Role            `gorm:"type:varchar(16);not null;default:member" json:"role"`
	MonthlyIncome decimal.Decimal `gorm:"type:decimal(12,2);not null;default:0" json:"monthlyIncome"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`

	Categories []Category `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	Expenses   []Expense  `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

// Category is a user-owned spending category tied to one of the four buckets.
type Category struct {
	ID        string       `gorm:"type:char(36);primaryKey" json:"id"`
	Name      string       `gorm:"type:varchar(100);not null;uniqueIndex:idx_user_name,priority:2" json:"name"`
	Type      CategoryType `gorm:"type:varchar(16);not null" json:"type"`
	UserID    string       `gorm:"type:char(36);not null;uniqueIndex:idx_user_name,priority:1;index" json:"userId"`
	CreatedAt time.Time    `json:"createdAt"`
}

// Expense is a single logged spend, owned by a user and tied to a category.
type Expense struct {
	ID         string          `gorm:"type:char(36);primaryKey" json:"id"`
	Amount     decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"amount"`
	Note       *string         `gorm:"type:text" json:"note"`
	SpentOn    time.Time       `gorm:"type:date;not null;index:idx_user_spent,priority:2" json:"spentOn"`
	UserID     string          `gorm:"type:char(36);not null;index:idx_user_spent,priority:1" json:"userId"`
	CategoryID string          `gorm:"type:char(36);not null;index" json:"categoryId"`
	Category   *Category       `gorm:"constraint:OnDelete:RESTRICT" json:"category,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

// AppSetting is a global singleton (id = "global") managed by admins.
type AppSetting struct {
	ID                 string            `gorm:"type:varchar(16);primaryKey" json:"id"`
	SavingsTargetType  SavingsTargetType `gorm:"type:varchar(16);not null;default:percent" json:"savingsTargetType"`
	SavingsTargetValue decimal.Decimal   `gorm:"type:decimal(12,2);not null;default:20" json:"savingsTargetValue"`
	UpdatedAt          time.Time         `json:"updatedAt"`
	UpdatedBy          *string           `gorm:"type:char(36)" json:"updatedBy"`
}

// GlobalSettingID is the fixed primary key of the singleton AppSetting row.
const GlobalSettingID = "global"

// RefreshToken stores a hashed refresh token so sessions can be revoked.
type RefreshToken struct {
	ID        string    `gorm:"type:char(36);primaryKey" json:"-"`
	UserID    string    `gorm:"type:char(36);not null;index" json:"-"`
	TokenHash string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"-"`
	Revoked   bool      `gorm:"not null;default:false" json:"-"`
	CreatedAt time.Time `json:"-"`
}

// BeforeCreate assigns a UUID primary key when one is not already set.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	return nil
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return nil
}

func (e *Expense) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	return nil
}

func (r *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
	return nil
}
