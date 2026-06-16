package utils

import (
	"context"
	"fynance/models"
	"math"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetMonthlyReport(window fyne.Window, months []string) ([]models.Report, error) {
	incomeCollection := GetCollection("income")
	expenseCollection := GetCollection("expenses")

	currentYear := time.Now().Format("2006")

	getTotal := func(collection *mongo.Collection, month string) float64 {
		pipeline := mongo.Pipeline{
			{{
				Key: "$match",
				Value: bson.D{
					{Key: "year", Value: currentYear},
					{Key: "month", Value: month},
				},
			}},
			{{
				Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: nil},
					{Key: "total", Value: bson.D{
						{Key: "$sum", Value: "$amount"},
					}},
				},
			}},
		}

		ctx := context.Background()

		cursor, err := collection.Aggregate(ctx, pipeline)
		if err != nil {
			dialog.ShowInformation(
				"Monthly Report",
				"Error fetching data for "+month,
				window,
			)
			return 0
		}
		defer cursor.Close(ctx)

		var result struct {
			Total float64 `bson:"total"`
		}

		if cursor.Next(ctx) {
			if err := cursor.Decode(&result); err != nil {
				dialog.ShowInformation(
					"Monthly Report",
					"Error decoding result for "+month,
					window,
				)
				return 0
			}

			return math.Round(result.Total*100) / 100
		}

		return 0
	}

	var reports []models.Report

	for _, month := range months {
		totalIncome := getTotal(incomeCollection, month)
		totalExpense := getTotal(expenseCollection, month)

		reports = append(reports, models.Report{
			Period:       month,
			TotalIncome:  totalIncome,
			TotalExpense: totalExpense,
			Balance:      math.Round((totalIncome-totalExpense)*100) / 100,
		})
	}

	return reports, nil
}

func GetDailyReport(window fyne.Window) ([]models.Report, error) {
	incomeCollection := GetCollection("income")
	expenseCollection := GetCollection("expenses")

	//currentYear := time.Now().Format("2006")
	//currentMonth := time.Now().Format("January")

	type totalsMap map[string]float64

	getTotals := func(collection *mongo.Collection) (totalsMap, error) {
		pipeline := mongo.Pipeline{
			{{
				Key: "$group",
				Value: bson.D{
					{
						Key: "_id",
						Value: bson.D{
							{
								Key: "$dateToString",
								Value: bson.D{
									{Key: "format", Value: "%Y-%m-%d"},
									{Key: "date", Value: "$created_at"},
								},
							},
						},
					},
					{
						Key: "total",
						Value: bson.D{
							{Key: "$sum", Value: "$amount"},
						},
					},
				},
			}},
		}

		ctx := context.Background()

		cursor, err := collection.Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		totals := make(totalsMap)

		for cursor.Next(ctx) {
			var result struct {
				Day   string  `bson:"_id"`
				Total float64 `bson:"total"`
			}

			if err := cursor.Decode(&result); err != nil {
				return nil, err
			}

			totals[result.Day] = math.Round(result.Total*100) / 100
		}

		return totals, nil
	}

	incomeTotals, err := getTotals(incomeCollection)
	if err != nil {
		return nil, err
	}

	expenseTotals, err := getTotals(expenseCollection)
	if err != nil {
		return nil, err
	}

	periods := make(map[string]bool)

	for day := range incomeTotals {
		periods[day] = true
	}

	for day := range expenseTotals {
		periods[day] = true
	}

	var days []string
	for day := range periods {
		days = append(days, day)
	}

	sort.Strings(days)

	var reports []models.Report

	for _, day := range days {
		reports = append(reports, models.Report{
			Period:       day,
			TotalIncome:  incomeTotals[day],
			TotalExpense: expenseTotals[day],
			Balance:      incomeTotals[day] - expenseTotals[day],
		})
	}

	return reports, nil
}

func GetWeeklyReport(window fyne.Window) ([]models.Report, error) {
	incomeCollection := GetCollection("income")
	expenseCollection := GetCollection("expenses")

	type totalsMap map[string]float64

	getTotals := func(collection *mongo.Collection) (totalsMap, error) {
		pipeline := mongo.Pipeline{
			// Stage 1: Group by year and ISO week format (e.g., "2026-W25")
			{{
				Key: "$group",
				Value: bson.D{
					{
						Key: "_id",
						Value: bson.D{
							{
								Key: "$dateToString",
								Value: bson.D{
									{Key: "format", Value: "%G-W%V"},
									{Key: "date", Value: "$created_at"},
								},
							},
						},
					},
					{
						Key: "total",
						Value: bson.D{
							{Key: "$sum", Value: "$amount"},
						},
					},
				},
			}},
		}

		ctx := context.Background()

		cursor, err := collection.Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		totals := make(totalsMap)

		for cursor.Next(ctx) {
			// Corrected struct mapping to match the dynamic string ID generated above
			var result struct {
				Week  string  `bson:"_id"`
				Total float64 `bson:"total"`
			}

			if err := cursor.Decode(&result); err != nil {
				return nil, err
			}

			totals[result.Week] = math.Round(result.Total*100) / 100
		}

		return totals, nil
	}

	incomeTotals, err := getTotals(incomeCollection)
	if err != nil {
		return nil, err
	}

	expenseTotals, err := getTotals(expenseCollection)
	if err != nil {
		return nil, err
	}

	periods := make(map[string]bool)
	for week := range incomeTotals {
		periods[week] = true
	}
	for week := range expenseTotals {
		periods[week] = true
	}

	var weeks []string
	for week := range periods {
		weeks = append(weeks, week)
	}

	// This now correctly sorts chronologically (e.g., "2025-W52" comes before "2026-W01")
	sort.Strings(weeks)

	var reports []models.Report

	for _, week := range weeks {
		reports = append(reports, models.Report{
			Period:       week, // Displays clean format like "2026-W25"
			TotalIncome:  incomeTotals[week],
			TotalExpense: expenseTotals[week],
			Balance:      math.Round((incomeTotals[week]-expenseTotals[week])*100) / 100, // Safe rounding for float math
		})
	}

	return reports, nil
}
