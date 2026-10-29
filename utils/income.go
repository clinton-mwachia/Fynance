package utils

import (
	"context"
	"errors"
	"fynance/models"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gorm.io/gorm"
)

type MonthlyIncome struct {
	Month string  `bson:"_id"`
	Total float64 `bson:"count"`
}

func AddIncome(income models.Income, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}
	return db.Create(&income).Error
}

func GetAllIncomes(window fyne.Window) []models.Income {
	db := GetDB()
	var incomes []models.Income
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return incomes
	}

	if err := db.Find(&incomes).Error; err != nil {
		dialog.ShowError(err, window)
	}
	return incomes
}

func GetIncomeByID(id uint, window fyne.Window) models.Income {
	db := GetDB()
	var income models.Income
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return income
	}

	if err := db.First(&income, id).Error; err != nil {
		dialog.ShowError(err, window)
	}
	return income
}

func UpdateIncome(income models.Income, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	return db.Save(&income).Error
}

func DeleteIncome(id uint, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}
	return db.Delete(&models.Income{}, id).Error
}

func GetIncomesPaginated(page, limit int, w fyne.Window, updateProgress func(float64)) []models.Income {
	db := GetDB()
	var incomes []models.Income
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), w)
		return incomes
	}

	offset := (page - 1) * limit

	err := db.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&incomes).
		Error

	if err != nil {
		dialog.ShowError(err, w)
		return incomes
	}

	if updateProgress != nil && len(incomes) > 0 {
		updateProgress(1.0)
	}

	return incomes
}

func CountIncomes(w fyne.Window) int64 {
	db := GetDB()
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), w)
		return 0
	}

	var count int64
	if err := db.Model(&models.Income{}).Count(&count).Error; err != nil {
		dialog.ShowError(err, w)
	}
	return count
}

func SearchIncomes(searchText string, window fyne.Window) []models.Income {
	db := GetDB()
	var results []models.Income
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return nil
	}

	queryParam := "%" + searchText + "%"

	err := db.Where("category LIKE ? OR month LIKE ?", queryParam, queryParam).
		Find(&results).
		Error

	if err != nil {
		dialog.ShowError(err, window)
		return nil
	}
	return results
}

func SumIncomeByMonth(month string) (MonthlyIncome, error) {
	db := GetDB()
	if db == nil {
		return MonthlyIncome{}, errors.New("database connection not initialized")
	}

	currentYear := time.Now().Format("2006")
	result := MonthlyIncome{
		Month: month,
		Total: 0,
	}

	err := db.Table("incomes").
		Where("year = ? AND month = ?", currentYear, month).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&result.Total).
		Error

	if err != nil {
		return MonthlyIncome{}, err
	}
	return result, nil
}

func TotalIncome(w fyne.Window) float64 {
	db := GetDB()
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), w)
		return 0
	}

	currentYear := time.Now().Format("2006")
	var total float64

	err := db.Table("incomes").
		Where("year = ?", currentYear).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).
		Error

	if err != nil {
		dialog.ShowError(err, w)
		return 0
	}
	return total
}

func GetIncomeStats(ctx context.Context) (map[string]float64, error) {
	db := GetDB()
	if db == nil {
		return nil, errors.New("database connection not initialized")
	}

	var results []struct {
		Category string  `gorm:"column:category"`
		Count    float64 `gorm:"column:count"`
	}

	err := db.WithContext(ctx).
		Table("incomes").
		Select("category, SUM(amount) as count").
		Group("category").
		Order("count DESC").
		Limit(5).
		Scan(&results).
		Error

	if err != nil {
		return nil, err
	}

	stats := make(map[string]float64)
	for _, result := range results {
		stats[result.Category] = result.Count
	}
	return stats, nil
}

func BulkInsertIncome(incomes []models.Income, window fyne.Window, progressBar *widget.ProgressBar) {
	db := GetDB()
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return
	}

	totalIncomes := len(incomes)
	if totalIncomes == 0 {
		progressBar.SetValue(0)
		dialog.ShowInformation("Success", "No incomes to add.", window)
		return
	}

	go func() {
		now := time.Now()
		batchSize := 100

		err := db.Transaction(func(tx *gorm.DB) error {
			for i := 0; i < totalIncomes; i += batchSize {
				end := i + batchSize
				if end > totalIncomes {
					end = totalIncomes
				}

				batch := incomes[i:end]
				for j := range batch {
					batch[j].CreatedAt = now
					batch[j].UpdatedAt = now
				}

				if err := tx.Table("incomes").Create(&batch).Error; err != nil {
					return err
				}

				progressBar.SetValue(float64(end) / float64(totalIncomes))
			}
			return nil
		})

		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		dialog.ShowInformation("Success", "Incomes added successfully!", window)
	}()
}
