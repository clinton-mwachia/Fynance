package utils

import (
	"errors"
	"fmt"
	"fynance/models"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

func AddNotification(notification models.Notification, window fyne.Window) error {
	db := GetDB()
	if db == nil {
		return errors.New("database connection not initialized")
	}

	err := db.Create(&notification).Error
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to save notification:\n%w", err), window)
	}
	return err
}

func ClearNotifications(userID uint, window fyne.Window) {
	db := GetDB()
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return
	}

	err := db.Where("user_id = ?", userID).Delete(&models.Notification{}).Error
	if err != nil {
		dialog.ShowError(err, window)
	}
}

func GetUnreadNotificationsCount(userID uint, window fyne.Window) int {
	db := GetDB()
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return 0
	}

	var count int64
	err := db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).
		Error

	if err != nil {
		dialog.ShowError(err, window)
		return 0
	}

	return int(count)
}

func FetchNotifications(userID uint, window fyne.Window) []models.Notification {
	db := GetDB()
	var notifications []models.Notification
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return nil
	}

	err := db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifications).
		Error

	if err != nil {
		dialog.ShowError(err, window)
		return nil
	}

	return notifications
}

func MarkNotificationsAsRead(userID uint, window fyne.Window) {
	db := GetDB()
	if db == nil {
		dialog.ShowError(errors.New("database connection not initialized"), window)
		return
	}

	err := db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).
		Error

	if err != nil {
		dialog.ShowError(err, window)
	}
}

func PlayNotificationSound(window fyne.Window) {
	filePath := "assets/confirmation-tone.wav"

	file, err := os.Open(filePath)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	defer file.Close()

	streamer, format, err := wav.Decode(file)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	defer streamer.Close()

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
