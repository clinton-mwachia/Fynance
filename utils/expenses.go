package utils

import (
	"context"
	"errors"
	"fmt"
	"fynance/models"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gorm.io/gorm"
)

type MonthlyExpense struct {
	Month string  `bson:"_id"`
	Total float64 `bson:"count"`
}

func AddExpense(expense models.Expense, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	err := db.Create(&expense).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to save Expense:\n%w", err), window)
	}
	return err
}

func GetAllExpenses(window fyne.Window) []models.Expense {
	db := GetDB()
	var expenses []models.Expense

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return expenses
	}

	err := db.Order("created_at ASC").Find(&expenses).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load expenses:\n%w", err), window)
	}

	return expenses
}

func GetExpenseByID(id uint, window fyne.Window) models.Expense {
	db := GetDB()
	var expense models.Expense

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return expense
	}

	err := db.First(&expense, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dialog.ShowError(fmt.Errorf("expense record with ID %d could not be found", id), window)
		} else {
			dialog.ShowError(fmt.Errorf("failed to fetch expense payload:\n%w", err), window)
		}
	}

	return expense
}

func UpdateExpense(expense models.Expense, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	err := db.Save(&expense).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to update expense record:\n%w", err), window)
	}
	return err
}

func DeleteExpense(id uint, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	var expense models.Expense
	err := db.First(&expense, id).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("cannot delete: record not found\n%w", err), window)
		return err
	}

	return db.Delete(&expense).Error
}

func GetExpensesPaginated(page, limit int, w fyne.Window, updateProgress func(float64)) []models.Expense {
	db := GetDB()
	var expenses []models.Expense

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), w)
		return expenses
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	if updateProgress != nil {
		updateProgress(0.2)
	}

	err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&expenses).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("expenses pagination query failure:\n%w", err), w)
		if updateProgress != nil {
			updateProgress(0.0)
		}
		return expenses
	}

	if updateProgress != nil {
		updateProgress(1.0)
	}

	return expenses
}

func CountExpenses(w fyne.Window) int64 {
	db := GetDB()
	var count int64

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), w)
		return 0
	}

	err := db.Model(&models.Expense{}).Count(&count).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to count expenses:\n%w", err), w)
	}
	return count
}

func SearchExpenses(searchText string, window fyne.Window) []models.Expense {
	db := GetDB()
	var results []models.Expense

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return nil
	}

	queryPattern := "%" + searchText + "%"
	err := db.Where("category LIKE ?", queryPattern).Find(&results).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("expenses search execution failure:\n%w", err), window)
		return nil
	}

	return results
}

func SumExpenseByMonth(month string, window fyne.Window) (MonthlyExpense, error) {
	db := GetDB()
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return MonthlyExpense{}, errors.New("database connection not initialized")
	}

	currentYear := time.Now().Format("2006")

	result := MonthlyExpense{
		Month: month,
		Total: 0,
	}

	err := db.Table("expenses").
		Where("year = ? AND month = ?", currentYear, month).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&result.Total).
		Error

	if err != nil {
		return MonthlyExpense{}, err
	}

	return result, nil
}

func TotalExpenses(w fyne.Window) float64 {
	db := GetDB()
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), w)
		return 0
	}

	currentYear := time.Now().Format("2006")

	var total float64

	err := db.Table("expenses").
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

func GetExpenseStats(ctx context.Context) (map[string]float64, error) {
	db := GetDB()
	if db == nil {
		return nil, errors.New("database connection not initialized")
	}

	var results []struct {
		Category string  `gorm:"column:category"`
		Count    float64 `gorm:"column:count"`
	}

	err := db.WithContext(ctx).
		Table("expenses").
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

func BulkInsertExpenses(expenses []models.Expense, window fyne.Window, progressBar *widget.ProgressBar) {
	db := GetDB()
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return
	}

	totalExpenses := len(expenses)
	if totalExpenses == 0 {
		progressBar.SetValue(0)
		dialog.ShowInformation("Success", "No expenses to add.", window)
		return
	}

	go func() {
		now := time.Now()
		batchSize := 100

		err := db.Transaction(func(tx *gorm.DB) error {
			for i := 0; i < totalExpenses; i += batchSize {
				end := i + batchSize
				if end > totalExpenses {
					end = totalExpenses
				}

				batch := expenses[i:end]

				for j := range batch {
					batch[j].CreatedAt = now
					batch[j].UpdatedAt = now
				}

				if err := tx.Create(&batch).Error; err != nil {
					return err
				}

				progress := end
				progressBar.SetValue(float64(progress) / float64(totalExpenses))
			}
			return nil
		})

		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		dialog.ShowInformation("Success", "Expenses added successfully!", window)
	}()
}
