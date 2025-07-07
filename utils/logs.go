package utils

import (
	"context"
	"fynance/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AddLog adds a new todo to the database.
func AddLog(log models.Log, window fyne.Window) {
	collection := GetCollection("logs")
	_, err := collection.InsertOne(context.TODO(), log)
	if err != nil {
		dialog.ShowError(err, window)
	}
}

// GetAllLogs retrieves all logs from the database.
func GetAllLogs(window fyne.Window) []models.Log {
	collection := GetCollection("logs")
	var logs []models.Log

	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		dialog.ShowError(err, window)
		return logs
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &logs); err != nil {
		dialog.ShowError(err, window)
	}

	return logs
}

// GetLogByID retrieves a single log by its ID from the database.
func GetLogByID(id primitive.ObjectID, window fyne.Window) models.Log {
	collection := GetCollection("logs")
	var log models.Log

	err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&log)
	if err != nil {
		dialog.ShowError(err, window)
	}

	return log
}

// DeleteLog deletes a log from the database.
func DeleteLog(id primitive.ObjectID, window fyne.Window) {
	collection := GetCollection("logs")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		dialog.ShowError(err, window)
	} else {
		dialog.ShowInformation("Success", "Log deleted successfully!", window)
	}
}

// DeleteLog deletes a log from the database.
func DeleteAllLogs(window fyne.Window) {
	collection := GetCollection("logs")
	_, err := collection.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		dialog.ShowError(err, window)
	} else {
		dialog.ShowInformation("Success", "Log deleted successfully!", window)
	}
}

// GetLogsPaginated fetches todos with pagination from the database
func GetLogsPaginated(page, limit int, w fyne.Window) []models.Log {
	collection := GetCollection("logs")

	skip := (page - 1) * limit
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	var logs []models.Log

	cursor, err := collection.Find(context.TODO(), bson.M{}, findOptions)
	if err != nil {
		dialog.ShowError(err, w)
		return logs
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &logs); err != nil {
		dialog.ShowError(err, w)
	}

	return logs
}

// search logs by quering the db
func SearchLogs(searchText string, window fyne.Window) []models.Log {
	collection := GetCollection("logs")

	// Create a case-insensitive regex pattern for the search
	searchPattern := bson.M{
		"$regex":   searchText,
		"$options": "i", // Case-insensitive
	}

	filter := bson.M{
		"$or": []bson.M{
			{"timestamp": searchPattern},
			{"details": searchPattern},
		},
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		dialog.ShowError(err, window)
		return nil
	}
	defer cursor.Close(context.TODO())

	var results []models.Log
	if err = cursor.All(context.TODO(), &results); err != nil {
		dialog.ShowError(err, window)
		return nil
	}

	return results

}

// CountLogs returns the total count of logs
func CountLogs(w fyne.Window) int64 {
	collection := GetCollection("logs")
	count, err := collection.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		dialog.ShowError(err, w)
	}
	return count
}
