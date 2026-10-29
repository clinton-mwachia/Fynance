package utils

import (
	"errors"
	"fmt"
	"fynance/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"gorm.io/gorm"
)

func AddExpenseDetail(ExpenseDetail models.ExpenseDetail, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	err := db.Create(&ExpenseDetail).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to save Expense Detail:\n%w", err), window)
	}
	return err
}

func GetAllExpenseDetails(window fyne.Window) []models.ExpenseDetail {
	db := GetDB()
	var expenseDetails []models.ExpenseDetail

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return expenseDetails
	}

	err := db.Order("expense_category ASC").Find(&expenseDetails).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load expense details:\n%w", err), window)
	}

	return expenseDetails
}

func GetExpenseDetailByID(id uint, window fyne.Window) models.ExpenseDetail {
	db := GetDB()
	var expenseDetail models.ExpenseDetail

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return expenseDetail
	}

	err := db.First(&expenseDetail, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dialog.ShowError(fmt.Errorf("expense Detail record with ID %d could not be found", id), window)
		} else {
			dialog.ShowError(fmt.Errorf("failed to fetch expense Detail payload:\n%w", err), window)
		}
	}

	return expenseDetail
}

func UpdateExpenseDetail(expenseDetail models.ExpenseDetail, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	err := db.Save(&expenseDetail).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to update expense detail record:\n%w", err), window)
	}
	return err
}

func DeleteExpenseDetail(id uint, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	var expenseDetail models.ExpenseDetail
	err := db.First(&expenseDetail, id).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("cannot delete: record not found\n%w", err), window)
		return err
	}

	return db.Delete(&expenseDetail).Error
}

func GetExpenseDetailsPaginated(page, limit int, w fyne.Window, updateProgress func(float64)) []models.ExpenseDetail {
	db := GetDB()
	var expensesDetails []models.ExpenseDetail

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), w)
		return expensesDetails
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	if updateProgress != nil {
		updateProgress(0.2)
	}

	err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&expensesDetails).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("expense details pagination query failure:\n%w", err), w)
		if updateProgress != nil {
			updateProgress(0.0)
		}
		return expensesDetails
	}

	if updateProgress != nil {
		updateProgress(1.0)
	}

	return expensesDetails
}

func CountExpenseDetails(w fyne.Window) int64 {
	db := GetDB()
	var count int64

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), w)
		return 0
	}

	err := db.Model(&models.ExpenseDetail{}).Count(&count).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to count expense details:\n%w", err), w)
	}
	return count
}

func SearchExpenseDetails(searchText string, window fyne.Window) []models.ExpenseDetail {
	db := GetDB()
	var results []models.ExpenseDetail

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return nil
	}

	queryPattern := "%" + searchText + "%"
	err := db.Where("expense_category LIKE ?", queryPattern).Find(&results).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("expense details search execution failure:\n%w", err), window)
		return nil
	}

	return results
}
