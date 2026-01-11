package utils

import (
	"fynance/models"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Logger(details string, status string, window fyne.Window) {
	parsedTime, err := time.Parse("02-01-2006 15:04:05", time.Now().Format("02-01-2006 15:04:05"))

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
