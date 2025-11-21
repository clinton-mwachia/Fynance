package utils

import (
	"context"
	"fynance/models"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AddNotification adds a new notification to the database
func AddNotification(notification models.Notification, window fyne.Window) {
	collection := GetCollection("notifications")

	notification.ID = primitive.NewObjectID() // Assign a new ObjectID
	notification.CreatedAt = primitive.NewDateTimeFromTime(time.Now())

	_, err := collection.InsertOne(context.TODO(), notification)
	if err != nil {
		dialog.ShowError(err, window)
	}
}

// ClearNotifications clears all notifications for a user
func ClearNotifications(userID primitive.ObjectID, window fyne.Window) {
	collection := GetCollection("notifications")
	filter := bson.M{"user_id": userID}

	_, err := collection.DeleteMany(context.TODO(), filter)
	if err != nil {
		dialog.ShowError(err, window)
	}
}

// GetUnreadNotificationsCount returns the count of unread notifications for a user
func GetUnreadNotificationsCount(userID primitive.ObjectID, window fyne.Window) int {
	collection := GetCollection("notifications")
	filter := bson.M{"user_id": userID, "is_read": false}

	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		dialog.ShowError(err, window)
		return 0
	}

	return int(count)
}

// FetchNotifications retrieves all notifications for a user
func FetchNotifications(userID primitive.ObjectID, window fyne.Window) []models.Notification {
	collection := GetCollection("notifications")
	filter := bson.M{"user_id": userID}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		dialog.ShowError(err, window)
		return nil
	}
	defer cur.Close(context.TODO())

	var notifications []models.Notification
	for cur.Next(context.TODO()) {
		var notif models.Notification
		err := cur.Decode(&notif)
		if err != nil {
			dialog.ShowError(err, window)
			continue
		}
		notifications = append(notifications, notif)
	}

	return notifications
}

// MarkNotificationsAsRead marks all notifications for a user as read
func MarkNotificationsAsRead(userID primitive.ObjectID, window fyne.Window) {
	collection := GetCollection("notifications")
	filter := bson.M{"user_id": userID, "is_read": false}
	update := bson.M{"$set": bson.M{"is_read": true}}

	_, err := collection.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		dialog.ShowError(err, window)
	}
}

// PlayNotificationSound plays a notification sound in WAV format
func PlayNotificationSound(window fyne.Window) {
	// Relative path for flexibility
	filePath := "assets/confirmation-tone.wav"

	// Open the sound file
	file, err := os.Open(filePath)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	defer file.Close()

	// Decode the WAV file
	streamer, format, err := wav.Decode(file)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	defer streamer.Close()

	// Initialize the speaker
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	// Play the sound and wait until it finishes
	done := make(chan bool) // Channel to wait for playback completion
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done // Block until the sound finishes playing
}
