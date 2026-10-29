package models

import (
	"time"
)

type ExpenseDetail struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ExpenseCategory string    `gorm:"type:varchar(255);not null" json:"expense_category"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
