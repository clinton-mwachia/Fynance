package utils

import (
	"errors"
	"fynance/models"
	"math"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func GetMonthlyReport(window fyne.Window, months []string) ([]models.Report, error) {
	db := GetDB()
	if db == nil {
		return nil, errors.New("database connection not initialized")
	}

	currentYear := time.Now().Format("2006")

	getTotal := func(tableName string, month string) float64 {
		var total float64
		err := db.Table(tableName).
			Where("year = ? AND month = ?", currentYear, month).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&total).
			Error

		if err != nil {
			dialog.ShowInformation("Monthly Report", "Error fetching data for "+month, window)
			return 0
		}
		return math.Round(total*100) / 100
	}

	var reports []models.Report

	for _, month := range months {
		totalIncome := getTotal("incomes", month)
		totalExpense := getTotal("expenses", month)

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
	db := GetDB()
	if db == nil {
		return nil, errors.New("database connection not initialized")
	}

	type totalsMap map[string]float64

	getTotals := func(tableName string) (totalsMap, error) {
		var results []struct {
			Day   string  `gorm:"column:day"`
			Total float64 `gorm:"column:total"`
		}

		err := db.Table(tableName).
			Select("strftime('%Y-%m-%d', created_at) as day, SUM(amount) as total").
			Group("day").
			Scan(&results).
			Error

		if err != nil {
			return nil, err
		}

		totals := make(totalsMap)
		for _, result := range results {
			totals[result.Day] = math.Round(result.Total*100) / 100
		}
		return totals, nil
	}

	incomeTotals, err := getTotals("incomes")
	if err != nil {
		return nil, err
	}

	expenseTotals, err := getTotals("expenses")
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
			Balance:      math.Round((incomeTotals[day]-expenseTotals[day])*100) / 100,
		})
	}

	return reports, nil
}

func GetWeeklyReport(window fyne.Window) ([]models.Report, error) {
	db := GetDB()
	if db == nil {
		return nil, errors.New("database connection not initialized")
	}

	type totalsMap map[string]float64

	getTotals := func(tableName string) (totalsMap, error) {
		var results []struct {
			Week  string  `gorm:"column:week"`
			Total float64 `gorm:"column:total"`
		}

		err := db.Table(tableName).
			Select("strftime('%Y-W%W', created_at) as week, SUM(amount) as total").
			Group("week").
			Scan(&results).
			Error

		if err != nil {
			return nil, err
		}

		totals := make(totalsMap)
		for _, result := range results {
			totals[result.Week] = math.Round(result.Total*100) / 100
		}
		return totals, nil
	}

	incomeTotals, err := getTotals("incomes")
	if err != nil {
		return nil, err
	}

	expenseTotals, err := getTotals("expenses")
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
	sort.Strings(weeks)

	var reports []models.Report
	for _, week := range weeks {
		reports = append(reports, models.Report{
			Period:       week,
			TotalIncome:  incomeTotals[week],
			TotalExpense: expenseTotals[week],
			Balance:      math.Round((incomeTotals[week]-expenseTotals[week])*100) / 100,
		})
	}

	return reports, nil
}
