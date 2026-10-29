package pages

import (
	"fynance/models"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	notificationCountLabel *widget.Label
	notificationIcon       *widget.Button
	darkModeIcon           *widget.Button
	settingsIcon           *widget.Button
	notifications          []models.Notification
	timeLabel              *widget.Label
)

var statusLabel *widget.Label

func isOnline() bool {
	_, err := http.Get("https//google.com")
	return err == nil
}

func showStatus(status string, window fyne.Window) {
	fyne.Do(func() {
		statusLabel.SetText(status)
		statusLabel.Show()
	})
	go func() {
		time.Sleep(5 * time.Second)
		fyne.Do(func() {
			window.Canvas().Refresh(statusLabel)
			statusLabel.Hide()
		})
	}()
}

func monitorNetworkStatus(window fyne.Window) {
	for {
		if isOnline() {
			showStatus("Online", window)
		} else {
			showStatus("Offline", window)
		}
		time.Sleep(5 * time.Second)
	}
}

func Header(window fyne.Window, title string) *fyne.Container {
	statusLabel = widget.NewLabel("")
	statusLabel.Hide()

	go monitorNetworkStatus(window)

	notificationCountLabel = widget.NewLabel("0")
	notificationIcon = widget.NewButtonWithIcon("", theme.MailComposeIcon(), func() {
		ShowNotifications(window)
	})

	settingsIcon = widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		ShowSettings(window)
	})
	var themeIcon = theme.VisibilityIcon()
	if isDarkMode {
		themeIcon = theme.VisibilityOffIcon()
	} else {
		themeIcon = theme.VisibilityIcon()
	}

	darkModeIcon = widget.NewButtonWithIcon("", themeIcon, func() {
		toggleTheme(window)
	})

	timeLabel = widget.NewLabel(time.Now().Format("15:04:05 PM"))
	go func() {
		for {
			time.Sleep(time.Second)
			current := time.Now().Format("15:04:05 PM")

			if timeLabel.Text != current {
				fyne.Do(func() {
					timeLabel.SetText(current)
					window.Canvas().Refresh(timeLabel)
				})
			}
		}
	}()

	UpdateNotificationCount(window)

	header := container.NewHBox(
		widget.NewLabel(title),
		statusLabel,
		layout.NewSpacer(),
		timeLabel,
		darkModeIcon,
		settingsIcon,
		notificationIcon,
		notificationCountLabel,
	)

	return header
}
