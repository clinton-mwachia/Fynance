package utils

import (
	"errors"
	"fmt"
	"fynance/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"gorm.io/gorm"
)

func GetAllUsers(window fyne.Window) []models.User {
	db := GetDB()
	var users []models.User

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return users
	}

	err := db.Order("username ASC").Find(&users).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load records:\n%w", err), window)
	}

	return users
}

func GetUserByID(id uint, window fyne.Window) models.User {
	db := GetDB()
	var user models.User

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return user
	}

	err := db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dialog.ShowError(fmt.Errorf("user record with ID %d could not be found", id), window)
		} else {
			dialog.ShowError(fmt.Errorf("failed to fetch user payload:\n%w", err), window)
		}
	}

	return user
}

func UpdateUser(user models.User, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	return db.Save(&user).Error
}

func DeleteUser(id uint, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	var user models.User
	err := db.First(&user, id).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("cannot delete: record not found\n%w", err), window)
		return err
	}

	return db.Delete(&user).Error
}

func GetUsersPaginated(page, limit int, w fyne.Window, updateProgress func(float64)) []models.User {
	db := GetDB()
	var users []models.User

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), w)
		return users
	}

	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	if updateProgress != nil {
		updateProgress(0.2)
	}

	err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("pagination query failure:\n%w", err), w)
		if updateProgress != nil {
			updateProgress(0.0)
		}
		return users
	}

	if updateProgress != nil {
		updateProgress(1.0)
	}

	return users
}

func CountUsers(w fyne.Window) int64 {
	db := GetDB()
	var count int64

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), w)
		return 0
	}

	err := db.Model(&models.User{}).Count(&count).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to count records:\n%w", err), w)
	}

	return count
}

func SearchUsers(searchText string, window fyne.Window) []models.User {
	db := GetDB()
	var results []models.User

	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return nil
	}

	queryPattern := "%" + searchText + "%"

	err := db.Where("fullname LIKE ? OR username LIKE ? OR id_number LIKE ? OR phone LIKE ?",
		queryPattern, queryPattern, queryPattern, queryPattern).Find(&results).Error

	if err != nil {
		dialog.ShowError(fmt.Errorf("search tracking processing failure:\n%w", err), window)
		return nil
	}

	return results
}
