package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Income struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Category  string             `bson:"category"`
	Month     string             `bson:"month"`
	Year      string             `bson:"year"`
	Amount    float64            `bson:"amount"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

// time.Now().Format("2006-01-02 15:04:05")
