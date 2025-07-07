package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Log struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Status    string             `bson:"status"`
	Details   string             `bson:"details,omitempty"`
	Timestamp time.Time          `bson:"timestamp"`
}
