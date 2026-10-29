package models

import (
	"time"
)

type IncomeDetail struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	IncomeCategory string    `gorm:"type:varchar(255);not null" json:"income_category"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
