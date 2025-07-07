package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IncomeDetail struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	IncomeCategory string             `bson:"income_category"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
}

// time.Now().Format("2006-01-02 15:04:05")
