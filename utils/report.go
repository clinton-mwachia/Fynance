package utils

import (
	"context"
	"fynance/models"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// getMonthlyFinance calculates total income, expenses, and balance for multiple months
func GetMonthlyReport(window fyne.Window, months []string) ([]models.Report, error) {
	incomeCollection := GetCollection("income")
	expenseCollection := GetCollection("expenses")
	// Get current year
	currentYear := time.Now().Format("2006")

	// Function to get total amount from aggregation
	getTotal := func(collection *mongo.Collection, month string) float64 {
		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.D{{Key: "year", Value: currentYear}, {Key: "month", Value: month}}}},
			{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "total", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
			}}},
		}

		ctx := context.Background()
		cursor, err := collection.Aggregate(ctx, pipeline)
		if err != nil {
			dialog.ShowInformation("Aggregating", "Error fetching data for "+month, window)
			return 0 // Default to 0 in case of an error
		}
		defer cursor.Close(ctx)

		var result struct {
			Total float64 `bson:"total"`
		}
		if cursor.Next(ctx) {
			if err := cursor.Decode(&result); err != nil {
				dialog.ShowInformation("Aggregating", "Error decoding result for "+month, window)
				return 0 // Default to 0 if decoding fails
			}
			return math.Round(result.Total*100) / 100
		}
		return 0 // Default to 0 if no data is found
	}

	var results []models.Report

	for _, month := range months {
		// Fetch totals for the month
		totalIncome := getTotal(incomeCollection, month)
		totalExpense := getTotal(expenseCollection, month)

		// Calculate balance
		balance := totalIncome - totalExpense

		// Append result
		results = append(results, models.Report{
			Month:        month,
			TotalIncome:  totalIncome,
			TotalExpense: totalExpense,
			Balance:      balance,
		})
	}

	return results, nil
}
