package utils

import (
	"errors"
	"fmt"
	"fynance/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"gorm.io/gorm"
)

func AddLog(log models.Log, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	err := db.Create(&log).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to save Log:\n%w", err), window)
	}
	return err
}

func GetAllLogs(window fyne.Window) []models.Log {
	db := GetDB()
	var logs []models.Log

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return logs
	}

	err := db.Order("created_at ASC").Find(&logs).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load logs:\n%w", err), window)
	}

	return logs
}

func GetLogByID(id uint, window fyne.Window) models.Log {
	db := GetDB()
	var log models.Log

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return log
	}

	err := db.First(&log, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dialog.ShowError(fmt.Errorf("log record with ID %d could not be found", id), window)
		} else {
			dialog.ShowError(fmt.Errorf("failed to fetch log payload:\n%w", err), window)
		}
	}

	return log
}

func DeleteLog(id uint, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	var log models.Log
	err := db.First(&log, id).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("cannot delete log: record not found\n%w", err), window)
		return err
	}

	return db.Delete(&log).Error
}

func DeleteAllLogs(window fyne.Window) {
	db := GetDB()
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return
	}

	err := db.Table("logs").Where("1 = 1").Delete(struct{}{}).Error
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	dialog.ShowInformation("Success", "Logs deleted successfully!", window)
}

func GetLogsPaginated(page, limit int, w fyne.Window, updateProgress func(float64)) []models.Log {
	db := GetDB()
	var logs []models.Log

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), w)
		return logs
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	if updateProgress != nil {
		updateProgress(0.2)
	}

	err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("logs pagination query failure:\n%w", err), w)
		if updateProgress != nil {
			updateProgress(0.0)
		}
		return logs
	}

	if updateProgress != nil {
		updateProgress(1.0)
	}

	return logs
}

func CountLogs(w fyne.Window) int64 {
	db := GetDB()
	var count int64

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), w)
		return 0
	}

	err := db.Model(&models.Log{}).Count(&count).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to count logs:\n%w", err), w)
	}
	return count
}

func SearchLogs(searchText string, window fyne.Window) []models.Log {
	db := GetDB()
	var results []models.Log

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return nil
	}

	queryPattern := "%" + searchText + "%"

	err := db.Where("status LIKE ? OR details LIKE ?", queryPattern, queryPattern).Find(&results).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("logs search execution failure:\n%w", err), window)
		return nil
	}

	return results
}
