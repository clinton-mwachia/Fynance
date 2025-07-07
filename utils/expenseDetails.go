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
func AddExpenseDetail(ExpenseDetail models.ExpenseDetail, window fyne.Window) error {
	collection := GetCollection("expense_details")
	_, err := collection.InsertOne(context.TODO(), ExpenseDetail)
	return err
}

// GetAllExpenseDetails retrieves all ExpenseDetails from the database.
func GetAllExpenseDetails(window fyne.Window) []models.ExpenseDetail {
	collection := GetCollection("expense_details")
	var ExpenseDetails []models.ExpenseDetail

	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		dialog.ShowError(err, window)
		return ExpenseDetails
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &ExpenseDetails); err != nil {
		dialog.ShowError(err, window)
	}

	return ExpenseDetails
}

// GetExpenseDetailByID retrieves a single ExpenseDetail by its ID from the database.
func GetExpenseDetailByID(id primitive.ObjectID, window fyne.Window) models.ExpenseDetail {
	collection := GetCollection("expense_details")
	var ExpenseDetail models.ExpenseDetail

	err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&ExpenseDetail)
	if err != nil {
		dialog.ShowError(err, window)
	}

	return ExpenseDetail
}

// UpdateExpenseDetail updates an existing ExpenseDetail in the database.
func UpdateExpenseDetail(ExpenseDetail models.ExpenseDetail, window fyne.Window) error {
	collection := GetCollection("expense_details")
	_, err := collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": ExpenseDetail.ID},
		bson.M{"$set": ExpenseDetail},
	)
	return err
}

// DeleteExpenseDetail deletes a ExpenseDetail from the database.
func DeleteExpenseDetail(id primitive.ObjectID, window fyne.Window) error {
	collection := GetCollection("expense_details")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": id})
	return err
}

// GetExpenseDetailsPaginated fetches ExpenseDetails with pagination from the database
func GetExpenseDetailsPaginated(page, limit int, w fyne.Window) []models.ExpenseDetail {
	collection := GetCollection("expense_details")

	skip := (page - 1) * limit
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	var expense_details []models.ExpenseDetail

	cursor, err := collection.Find(context.TODO(), bson.M{}, findOptions)
	if err != nil {
		dialog.ShowError(err, w)
		return expense_details
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &expense_details); err != nil {
		dialog.ShowError(err, w)
	}
	// Process Details and update progress
	count := 0
	for cursor.Next(context.TODO()) {
		var expense_detail models.ExpenseDetail
		if err := cursor.Decode(&expense_detail); err != nil {
			dialog.ShowError(err, w)
			continue
		}
		expense_details = append(expense_details, expense_detail)
		count++

	}

	if err := cursor.Err(); err != nil {
		dialog.ShowError(err, w)
	}

	return expense_details
}

// CountExpenseDetails returns the total count of ExpenseDetails for a user
func CountExpenseDetails(w fyne.Window) int64 {
	collection := GetCollection("expense_details")
	count, err := collection.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		dialog.ShowError(err, w)
	}
	return count
}

// search ExpenseDetails by quering the db
func SearchExpenseDetails(searchText string, window fyne.Window) []models.ExpenseDetail {
	collection := GetCollection("expense_details")

	// Create a case-insensitive regex pattern for the search
	searchPattern := bson.M{
		"$regex":   searchText,
		"$options": "i", // Case-insensitive
	}

	filter := bson.M{
		"$or": []bson.M{
			{"expense_category": searchPattern},
		},
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		dialog.ShowError(err, window)
		return nil
	}
	defer cursor.Close(context.TODO())

	var results []models.ExpenseDetail
	if err = cursor.All(context.TODO(), &results); err != nil {
		dialog.ShowError(err, window)
		return nil
	}

	return results

}
