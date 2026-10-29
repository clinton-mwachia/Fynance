package models

import (
	"time"
)

type Log struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Status    string    `gorm:"type:varchar(100);not null" json:"status"`
	Details   string    `gorm:"type:varchar(200);not null" json:"details"`
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
}
