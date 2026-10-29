package pages

import (
	"fynance/helpers"
	"fynance/utils"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func ShowNotifications(window fyne.Window) {

	userID := helpers.CurrentUserID
	notifications = utils.FetchNotifications(userID, window)

	if len(notifications) == 0 {
		dialog.ShowInformation("Notifications", "No notifications found.", window)
		return
	}

	list := widget.NewList(
		func() int { return len(notifications) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(notifications[id].Message)
		},
	)

	markAsReadButton := widget.NewButton("Mark All as Read", func() {
		utils.MarkNotificationsAsRead(userID, window)
		UpdateNotificationCount(window)
		notificationIcon.Refresh()
		dialog.ShowInformation("Notifications", "All notifications marked as read.", window)
	})

	clearButton := widget.NewButton("Clear All", func() {
		utils.ClearNotifications(userID, window)
		UpdateNotificationCount(window)
		notificationIcon.Refresh()
		notifications = utils.FetchNotifications(userID, window)
		list.Refresh()
		dialog.ShowInformation("Notifications", "All notifications cleared.", window)
	})

	scrollableList := container.NewVScroll(list)
	scrollableList.SetMinSize(fyne.NewSize(400, 250))

	content := container.NewVBox(
		scrollableList,
		container.NewHBox(markAsReadButton, clearButton),
	)

	dialog.ShowCustom("Notifications", "Close", content, window)
}

func UpdateNotificationCount(window fyne.Window) {
	userID := helpers.CurrentUserID
	unreadCount := utils.GetUnreadNotificationsCount(userID, window)
	fyne.Do(func() {
		notificationCountLabel.SetText(strconv.Itoa(unreadCount))
		notificationIcon.Refresh()
	})
}
