package utils

import (
	"errors"
	"fmt"
	"fynance/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"gorm.io/gorm"
)

func AddIncomeDetail(incomeDetail models.IncomeDetail, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	err := db.Create(&incomeDetail).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to save Income Detail:\n%w", err), window)
	}
	return err
}

func GetAllIncomeDetails(window fyne.Window) []models.IncomeDetail {
	db := GetDB()
	var incomeDetails []models.IncomeDetail

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return incomeDetails
	}

	err := db.Order("income_category ASC").Find(&incomeDetails).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load income details:\n%w", err), window)
	}

	return incomeDetails
}

func GetIncomeDetailByID(id uint, window fyne.Window) models.IncomeDetail {
	db := GetDB()
	var incomeDetail models.IncomeDetail

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return incomeDetail
	}

	err := db.First(&incomeDetail, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dialog.ShowError(fmt.Errorf("income Detail record with ID %d could not be found", id), window)
		} else {
			dialog.ShowError(fmt.Errorf("failed to fetch income Detail payload:\n%w", err), window)
		}
	}

	return incomeDetail
}

func UpdateIncomeDetail(incomeDetail models.IncomeDetail, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	err := db.Save(&incomeDetail).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to update income detail record:\n%w", err), window)
	}
	return err
}

func DeleteIncomeDetail(id uint, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	var incomeDetail models.IncomeDetail
	err := db.First(&incomeDetail, id).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("cannot delete: record not found\n%w", err), window)
		return err
	}

	return db.Delete(&incomeDetail).Error
}

func GetIncomeDetailsPaginated(page, limit int, w fyne.Window, updateProgress func(float64)) []models.IncomeDetail {
	db := GetDB()
	var incomesDetails []models.IncomeDetail

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), w)
		return incomesDetails
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	if updateProgress != nil {
		updateProgress(0.2)
	}

	err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&incomesDetails).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("income details pagination query failure:\n%w", err), w)
		if updateProgress != nil {
			updateProgress(0.0)
		}
		return incomesDetails
	}

	if updateProgress != nil {
		updateProgress(1.0)
	}

	return incomesDetails
}

func CountIncomeDetails(w fyne.Window) int64 {
	db := GetDB()
	var count int64

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), w)
		return 0
	}

	err := db.Model(&models.IncomeDetail{}).Count(&count).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to count income details:\n%w", err), w)
	}
	return count
}

func SearchIncomeDetails(searchText string, window fyne.Window) []models.IncomeDetail {
	db := GetDB()
	var results []models.IncomeDetail

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return nil
	}

	queryPattern := "%" + searchText + "%"
	err := db.Where("income_category LIKE ?", queryPattern).Find(&results).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("income details search execution failure:\n%w", err), window)
		return nil
	}

	return results
}
