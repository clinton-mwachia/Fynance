package utils

import (
	"context"
	"fynance/models"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MonthlyExpense represents the aggregated result
type MonthlyExpense struct {
	Month string  `bson:"_id"`
	Total float64 `bson:"total"`
}

// AddExpense adds a new Expense to the database.
func AddExpense(Expense models.Expense, window fyne.Window) error {
	collection := GetCollection("expenses")
	_, err := collection.InsertOne(context.TODO(), Expense)
	return err
}

// GetAllExpenses retrieves all Expenses from the database.
func GetAllExpenses(window fyne.Window) []models.Expense {
	collection := GetCollection("expenses")
	var Expenses []models.Expense

	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		dialog.ShowError(err, window)
		return Expenses
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &Expenses); err != nil {
		dialog.ShowError(err, window)
	}

	return Expenses
}

// GetExpenseByID retrieves a single Expense by its ID from the database.
func GetExpenseByID(id primitive.ObjectID, window fyne.Window) models.Expense {
	collection := GetCollection("expenses")
	var Expense models.Expense

	err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&Expense)
	if err != nil {
		dialog.ShowError(err, window)
	}

	return Expense
}

// UpdateExpense updates an existing Expense in the database.
func UpdateExpense(Expense models.Expense, window fyne.Window) error {
	collection := GetCollection("expenses")
	_, err := collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": Expense.ID},
		bson.M{"$set": Expense},
	)
	return err
}

// DeleteExpense deletes a Expense from the database.
func DeleteExpense(id primitive.ObjectID, window fyne.Window) error {
	collection := GetCollection("expenses")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": id})
	return err
}

// GetExpensesPaginated fetches Expenses with pagination from the database
func GetExpensesPaginated(page, limit int, w fyne.Window, updateProgress func(float64)) []models.Expense {
	collection := GetCollection("expenses")

	skip := (page - 1) * limit
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	var expenses []models.Expense

	cursor, err := collection.Find(context.TODO(), bson.M{}, findOptions)
	if err != nil {
		dialog.ShowError(err, w)
		return expenses
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &expenses); err != nil {
		dialog.ShowError(err, w)
	}
	// Process todos and update progress
	count := 0
	for cursor.Next(context.TODO()) {
		var expense models.Expense
		if err := cursor.Decode(&expense); err != nil {
			dialog.ShowError(err, w)
			continue
		}
		expenses = append(expenses, expense)
		count++

		// Update progress dynamically
		if updateProgress != nil {
			updateProgress(float64(count) / float64(count))
		}
	}

	if err := cursor.Err(); err != nil {
		dialog.ShowError(err, w)
	}

	return expenses
}

// CountExpenses returns the total count of Expenses for a user
func CountExpenses(w fyne.Window) int64 {
	collection := GetCollection("expenses")
	count, err := collection.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		dialog.ShowError(err, w)
	}
	return count
}

// search Expenses by quering the db
func SearchExpenses(searchText string, window fyne.Window) []models.Expense {
	collection := GetCollection("expenses")

	// Create a case-insensitive regex pattern for the search
	searchPattern := bson.M{
		"$regex":   searchText,
		"$options": "i", // Case-insensitive
	}

	filter := bson.M{
		"$or": []bson.M{
			{"category": searchPattern},
			{"month": searchPattern},
		},
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		dialog.ShowError(err, window)
		return nil
	}
	defer cursor.Close(context.TODO())

	var results []models.Expense
	if err = cursor.All(context.TODO(), &results); err != nil {
		dialog.ShowError(err, window)
		return nil
	}

	return results

}

// total income by month in current year
func SumExpenseByMonth(month string) (MonthlyExpense, error) {
	collection := GetCollection("expenses")

	// Get current year
	currentYear := time.Now().Format("2006")

	// MongoDB aggregation pipeline
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "year", Value: currentYear}, {Key: "month", Value: month}}}}, // Filter by year and month
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$month"},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
		}}},
	}

	ctx := context.Background()
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return MonthlyExpense{}, err
	}
	defer cursor.Close(ctx)

	// Default result (if no income is found)
	result := MonthlyExpense{
		Month: month,
		Total: 0,
	}

	// Parse result if data exists
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return MonthlyExpense{}, err
		}
	}

	return result, nil
}

// Returns the total expenses amount for that year
func TotalExpenses(w fyne.Window) float64 {
	collection := GetCollection("expenses")

	// get current year
	currentYear := time.Now().Format("2006")

	// Filter by the "year" field
	filter := bson.M{
		"year": currentYear,
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		dialog.ShowError(err, w)
		return 0
	}
	defer cursor.Close(context.TODO())

	var total float64
	for cursor.Next(context.TODO()) {
		var expense models.Expense
		if err := cursor.Decode(&expense); err != nil {
			dialog.ShowError(err, w)
			continue
		}
		total += expense.Amount
	}

	if err := cursor.Err(); err != nil {
		dialog.ShowError(err, w)
	}

	return total
}
