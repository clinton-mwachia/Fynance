package utils

import (
	"fynance/models"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Logger(details string, status string, window fyne.Window) {
	parsedTime, err := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))

	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	myLog := models.Log{
		ID:        primitive.NewObjectID(),
		Timestamp: parsedTime,
		Details:   details,
		Status:    status,
	}
	AddLog(myLog, window)
}
