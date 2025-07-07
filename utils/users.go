package utils

import (
	"context"
	"fynance/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetAllUsers retrieves all users from the database.
func GetAllUsers(window fyne.Window) []models.User {
	collection := GetCollection("users")
	var users []models.User

	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		dialog.ShowError(err, window)
		return users
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &users); err != nil {
		dialog.ShowError(err, window)
	}

	return users
}

// GetUserByID retrieves a single user by its ID from the database.
func GetUserByID(id primitive.ObjectID, window fyne.Window) models.User {
	collection := GetCollection("users")
	var user models.User

	err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		dialog.ShowError(err, window)
	}

	return user
}

// UpdateUser updates an existing user in the database.
func UpdateUser(user models.User, window fyne.Window) error {
	collection := GetCollection("users")
	_, err := collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": user.ID},
		bson.M{"$set": user},
	)
	return err
}

// DeleteUser deletes a user from the database.
func DeleteUser(id primitive.ObjectID, window fyne.Window) error {
	collection := GetCollection("users")
	// Retrieve the user before deleting
	var user models.User
	err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		dialog.ShowError(err, window)
	}

	// Delete the user from the database
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": id})

	return err

}
