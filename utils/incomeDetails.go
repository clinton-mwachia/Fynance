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

// AddDetail adds a new Detail to the database.
func AddDetail(Detail models.IncomeDetail, window fyne.Window) error {
	collection := GetCollection("income_details")
	_, err := collection.InsertOne(context.TODO(), Detail)
	return err
}

// GetAllDetails retrieves all Details from the database.
func GetAllDetails(window fyne.Window) []models.IncomeDetail {
	collection := GetCollection("income_details")
	var Details []models.IncomeDetail

	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		dialog.ShowError(err, window)
		return Details
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &Details); err != nil {
		dialog.ShowError(err, window)
	}

	return Details
}

// GetDetailByID retrieves a single Detail by its ID from the database.
func GetDetailByID(id primitive.ObjectID, window fyne.Window) models.IncomeDetail {
	collection := GetCollection("income_details")
	var Detail models.IncomeDetail

	err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&Detail)
	if err != nil {
		dialog.ShowError(err, window)
	}

	return Detail
}

// UpdateDetail updates an existing Detail in the database.
func UpdateDetail(Detail models.IncomeDetail, window fyne.Window) error {
	collection := GetCollection("income_details")
	_, err := collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": Detail.ID},
		bson.M{"$set": Detail},
	)
	return err
}

// DeleteDetail deletes a Detail from the database.
func DeleteDetail(id primitive.ObjectID, window fyne.Window) error {
	collection := GetCollection("income_details")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": id})
	return err
}

// GetDetailsPaginated fetches Details with pagination from the database
func GetDetailsPaginated(page, limit int, w fyne.Window) []models.IncomeDetail {
	collection := GetCollection("income_details")

	skip := (page - 1) * limit
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	var income_details []models.IncomeDetail

	cursor, err := collection.Find(context.TODO(), bson.M{}, findOptions)
	if err != nil {
		dialog.ShowError(err, w)
		return income_details
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &income_details); err != nil {
		dialog.ShowError(err, w)
	}
	// Process Details and update progress
	count := 0
	for cursor.Next(context.TODO()) {
		var income_detail models.IncomeDetail
		if err := cursor.Decode(&income_detail); err != nil {
			dialog.ShowError(err, w)
			continue
		}
		income_details = append(income_details, income_detail)
		count++

	}

	if err := cursor.Err(); err != nil {
		dialog.ShowError(err, w)
	}

	return income_details
}

// CountDetails returns the total count of Details for a user
func CountDetails(w fyne.Window) int64 {
	collection := GetCollection("income_details")
	count, err := collection.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		dialog.ShowError(err, w)
	}
	return count
}

// search Details by quering the db
func SearchDetails(searchText string, window fyne.Window) []models.IncomeDetail {
	collection := GetCollection("income_details")

	// Create a case-insensitive regex pattern for the search
	searchPattern := bson.M{
		"$regex":   searchText,
		"$options": "i", // Case-insensitive
	}

	filter := bson.M{
		"$or": []bson.M{
			{"income_category": searchPattern},
		},
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		dialog.ShowError(err, window)
		return nil
	}
	defer cursor.Close(context.TODO())

	var results []models.IncomeDetail
	if err = cursor.All(context.TODO(), &results); err != nil {
		dialog.ShowError(err, window)
		return nil
	}

	return results

}
