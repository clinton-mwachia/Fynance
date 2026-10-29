package models

import (
	"time"
)

type Income struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Category  string    `gorm:"type:varchar(255);not null" json:"category"`
	Month     string    `gorm:"type:varchar(255);not null" json:"month"`
	Year      string    `gorm:"type:varchar(255);not null" json:"year"`
	Amount    float64   `gorm:"type:bigint;not null" json:"amount"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
