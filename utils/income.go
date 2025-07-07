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

// MonthlyIncome represents the aggregated result
type MonthlyIncome struct {
	Month string  `bson:"_id"`
	Total float64 `bson:"total"`
}

// AddIncome adds a new Income to the database.
func AddIncome(Income models.Income, window fyne.Window) error {
	collection := GetCollection("income")
	_, err := collection.InsertOne(context.TODO(), Income)
	return err
}

// GetAllIncomes retrieves all Incomes from the database.
func GetAllIncomes(window fyne.Window) []models.Income {
	collection := GetCollection("income")
	var Incomes []models.Income

	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		dialog.ShowError(err, window)
		return Incomes
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &Incomes); err != nil {
		dialog.ShowError(err, window)
	}

	return Incomes
}

// GetIncomeByID retrieves a single Income by its ID from the database.
func GetIncomeByID(id primitive.ObjectID, window fyne.Window) models.Income {
	collection := GetCollection("income")
	var Income models.Income

	err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&Income)
	if err != nil {
		dialog.ShowError(err, window)
	}

	return Income
}

// UpdateIncome updates an existing Income in the database.
func UpdateIncome(Income models.Income, window fyne.Window) error {
	collection := GetCollection("income")
	_, err := collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": Income.ID},
		bson.M{"$set": Income},
	)
	return err
}

// DeleteIncome deletes a Income from the database.
func DeleteIncome(id primitive.ObjectID, window fyne.Window) error {
	collection := GetCollection("income")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": id})
	return err
}

// GetIncomesPaginated fetches Incomes with pagination from the database
func GetIncomesPaginated(page, limit int, w fyne.Window, updateProgress func(float64)) []models.Income {
	collection := GetCollection("income")

	skip := (page - 1) * limit
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	var incomes []models.Income

	cursor, err := collection.Find(context.TODO(), bson.M{}, findOptions)
	if err != nil {
		dialog.ShowError(err, w)
		return incomes
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &incomes); err != nil {
		dialog.ShowError(err, w)
	}
	// Process todos and update progress
	count := 0
	for cursor.Next(context.TODO()) {
		var income models.Income
		if err := cursor.Decode(&income); err != nil {
			dialog.ShowError(err, w)
			continue
		}
		incomes = append(incomes, income)
		count++

		// Update progress dynamically
		if updateProgress != nil {
			updateProgress(float64(count) / float64(count))
		}
	}

	if err := cursor.Err(); err != nil {
		dialog.ShowError(err, w)
	}

	return incomes
}

// CountIncomes returns the total count of Incomes for a user
func CountIncomes(w fyne.Window) int64 {
	collection := GetCollection("income")
	count, err := collection.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		dialog.ShowError(err, w)
	}
	return count
}

// search Incomes by quering the db
func SearchIncomes(searchText string, window fyne.Window) []models.Income {
	collection := GetCollection("income")

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

	var results []models.Income
	if err = cursor.All(context.TODO(), &results); err != nil {
		dialog.ShowError(err, window)
		return nil
	}

	return results

}

// total income by month in current year
func SumIncomeByMonth(month string) (MonthlyIncome, error) {
	collection := GetCollection("income")

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
		return MonthlyIncome{}, err
	}
	defer cursor.Close(ctx)

	// Default result (if no income is found)
	result := MonthlyIncome{
		Month: month,
		Total: 0,
	}

	// Parse result if data exists
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return MonthlyIncome{}, err
		}
	}

	return result, nil
}
