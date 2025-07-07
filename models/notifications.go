package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Notification struct for storing notification data
type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Message   string             `bson:"message"`
	IsRead    bool               `bson:"is_read"`
	CreatedAt primitive.DateTime `bson:"created_at"`
}
