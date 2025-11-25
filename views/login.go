package views

import (
	"fynance/auth"
	"fynance/helpers"
	"fynance/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"go.mongodb.org/mongo-driver/mongo"
)

func LoginView(window fyne.Window, showDashboard func()) *fyne.Container {
	// Load background image
	bgImage := canvas.NewImageFromFile("assets/background.png")
	bgImage.FillMode = canvas.ImageFillStretch

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	loginButton := widget.NewButton("Login", func() {
		username := usernameEntry.Text
		password := passwordEntry.Text

		if username == "" || password == "" {
			dialog.ShowInformation("User Login", "All fields are required", window)
		}

		// Show progress bar dialog
		progressBar := widget.NewProgressBar()
		progressBar.SetValue(0)
		progressDialog := dialog.NewCustom("Logging In User", "Cancel", progressBar, window)
		progressDialog.Show()

		user, err := auth.Login(username, password, func(progress float64) {
			progressBar.SetValue(progress)
		})
		if err != nil {
			progressDialog.Hide()
			if err == mongo.ErrNoDocuments {
				utils.Logger("User not found", "ERROR", window)
				dialog.ShowInformation("User Login", "User not found", window)
			} else {
				utils.Logger(username+" wrong password/username", "ERROR", window)
				dialog.ShowInformation("User Login", "Wrong password/username ", window)
			}
		} else {
			progressDialog.Hide()
			detail := user.Username + " Logged in"
			utils.Logger(detail, "SUCCESS", window)
			helpers.CurrentUserID = user.ID
			showDashboard()
			dialog.ShowInformation("Login Successful", "Welcome, "+user.Username, window)
		}
	})

	registerButton := widget.NewButton("Register", func() {
		window.SetContent(RegisterView(window, showDashboard))
	})

	// enter key to login
	passwordEntry.OnSubmitted = func(s string) {
		loginButton.OnTapped()
	}

	form := container.NewVBox(
		usernameEntry,
		passwordEntry,
		loginButton,
		registerButton,
	)

	centeredForm := helpers.NewFixedWidthCenter(form, 300) // Set width to 300

	return container.NewStack(bgImage, container.NewCenter(centeredForm))

}
