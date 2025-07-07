package views

import (
	"fynance/helpers"
	"fynance/utils"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// notifications
func showNotifications(window fyne.Window) {
	// Fetch notifications from the database
	userID := helpers.CurrentUserID
	notifications = utils.FetchNotifications(userID, window)

	if len(notifications) == 0 {
		dialog.ShowInformation("Notifications", "No notifications found.", window)
		return
	}

	// Create a list to display notifications
	list := widget.NewList(
		func() int { return len(notifications) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(notifications[id].Message)
		},
	)

	// Create buttons for clearing and marking notifications
	markAsReadButton := widget.NewButton("Mark All as Read", func() {
		utils.MarkNotificationsAsRead(userID, window)
		updateNotificationCount(window)
		notificationIcon.Refresh()
		dialog.ShowInformation("Notifications", "All notifications marked as read.", window)
	})

	clearButton := widget.NewButton("Clear All", func() {
		utils.ClearNotifications(userID, window)
		updateNotificationCount(window)
		notificationIcon.Refresh()
		notifications = utils.FetchNotifications(userID, window)
		list.Refresh() // Refresh the list widget to update UI
		dialog.ShowInformation("Notifications", "All notifications cleared.", window)
	})

	// Create a scrollable container for the list
	scrollableList := container.NewVScroll(list)
	scrollableList.SetMinSize(fyne.NewSize(400, 250))

	// Create a container for the list and buttons
	content := container.NewVBox(
		scrollableList,
		container.NewHBox(markAsReadButton, clearButton),
	)

	// Show notifications in a new window or dialog
	dialog.ShowCustom("Notifications", "Close", content, window)
}

func updateNotificationCount(window fyne.Window) {
	userID := helpers.CurrentUserID
	unreadCount := utils.GetUnreadNotificationsCount(userID, window)
	notificationCountLabel.SetText(strconv.Itoa(unreadCount))
	notificationIcon.Refresh()
}
