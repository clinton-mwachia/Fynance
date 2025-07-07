package views

import (
	"fynance/auth"
	"fynance/helpers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func RegisterView(window fyne.Window, showDashboard func()) *fyne.Container {
	// Load background image
	bgImage := canvas.NewImageFromFile("assets/background.png")
	bgImage.FillMode = canvas.ImageFillStretch

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")

	phoneEntry := widget.NewEntry()
	phoneEntry.SetPlaceHolder("+254700000000")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	registerButton := widget.NewButton("Register", func() {

		username := usernameEntry.Text
		phone := phoneEntry.Text
		password := passwordEntry.Text

		if username == "" || phone == "" || password == "" {
			dialog.ShowInformation("User Register", "All fields are required", window)
		} else {

			err := auth.Register(username, phone, password)
			if err != nil {
				dialog.ShowInformation("User Register", "Cannot Register account: "+err.Error(), window)
			} else {
				dialog.ShowInformation("Registration Successful", "Please login, "+username, window)
				window.SetContent(LoginView(window, showDashboard))
			}
		}

	})

	loginButton := widget.NewButton("Login", func() {
		window.SetContent(LoginView(window, showDashboard))
	})

	form := container.NewVBox(
		usernameEntry,
		phoneEntry,
		passwordEntry,
		registerButton,
		loginButton,
	)

	centeredForm := helpers.NewFixedWidthCenter(form, 300)

	return container.NewStack(bgImage, container.NewCenter(centeredForm))
}
