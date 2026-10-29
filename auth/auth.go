package auth

import (
	"errors"
	"fmt"
	"fynance/helpers"
	"fynance/models"
	"fynance/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"gorm.io/gorm"
)

func Register(user *models.User) error {
	db := utils.GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	var existingUser models.User
	err := db.Where("username = ?", user.Username).First(&existingUser).Error
	if err == nil {
		return errors.New("the username is already taken")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("database verification error: %w", err)
	}

	hashedPassword, err := helpers.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("failed to secure password: %w", err)
	}

	user.Password = hashedPassword

	err = db.Create(&user).Error
	if err != nil {
		return fmt.Errorf("database insertion error: %w", err)
	}

	return nil
}

func Login(username, password string, updateProgress func(progress float64)) (*models.User, error) {
	db := utils.GetDB()
	if db == nil {
		return nil, errors.New("database connection not initialized")
	}

	var user models.User

	updateProgress(0.1)

	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		updateProgress(0.0)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("User not found")
		}
		return nil, fmt.Errorf("database lookup error: %w", err)
	}

	updateProgress(0.5)

	if !helpers.CheckPasswordHash(password, user.Password) {
		updateProgress(0.0)
		return nil, errors.New("Invalid username or password")
	}

	updateProgress(1.0)
	return &user, nil
}

func ResetPassword(userID uint, newPassword string, window fyne.Window) error {
	db := utils.GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	newHashedPassword, err := helpers.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	err = db.Model(&models.User{}).Where("id = ?", userID).Update("password", newHashedPassword).Error
	if err != nil {
		dialog.ShowError(err, window)
		return err
	}

	return nil
}
