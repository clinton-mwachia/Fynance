package views

import (
	"fmt"
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

// Check if app is online
func isOnline() bool {
	_, err := http.Get("https://www.google.com")
	return err == nil
}

// Function to display the status label and hide it after 5 seconds
func showStatus(status string, window fyne.Window) {
	statusLabel.SetText(status)
	statusLabel.Show()
	go func() {
		time.Sleep(5 * time.Second)
		window.Canvas().Refresh(statusLabel)
		statusLabel.Hide()
	}()
}

// Function to monitor network status
func monitorNetworkStatus(window fyne.Window) {
	for {
		if isOnline() {
			showStatus("Online", window)
		} else {
			showStatus("Offline", window)
		}
		time.Sleep(5 * time.Second) // Check every 5 seconds
	}
}

func Header(window fyne.Window) *fyne.Container {
	statusLabel = widget.NewLabel("")
	statusLabel.Hide() // Initially hidden

	go monitorNetworkStatus(window) // Start monitoring network status

	// Notification icon button with initial count
	notificationCountLabel = widget.NewLabel("0")
	notificationIcon = widget.NewButtonWithIcon("", theme.MailComposeIcon(), func() {
		showNotifications(window)
	})

	// settings icon
	settingsIcon = widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		showSettings(window)
	})
	var themeIcon = theme.VisibilityIcon()
	if isDarkMode {
		themeIcon = theme.VisibilityOffIcon()
	} else {
		themeIcon = theme.VisibilityIcon()
	}

	// dark mode icon
	darkModeIcon = widget.NewButtonWithIcon("", themeIcon, func() {
		toggleTheme(window)
	})

	// Real-time local time label
	timeLabel = widget.NewLabel(time.Now().Format("15:04:05 PM"))
	go func() {
		for {
			time.Sleep(time.Second)
			current := time.Now().Format("15:04:05 PM")
			// Only update if value changes
			if timeLabel.Text != current {
				timeLabel.SetText(current)
				window.Canvas().Refresh(timeLabel)
			}
		}
	}()

	// Set initial count
	updateNotificationCount(window)

	currentTime := time.Now()
	currentYear := currentTime.Year()

	// Header container
	header := container.NewHBox(
		widget.NewLabel(fmt.Sprintf("Fynance - %d", currentYear)),
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
