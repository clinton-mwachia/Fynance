package views

import (
	"encoding/json"
	"errors"
	"fynance/auth"
	"fynance/helpers"
	"fynance/models"
	"fynance/utils"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/crypto/bcrypt"
)

// Struct to hold app settings
type AppSettings struct {
	IsDarkMode bool   `json:"is_dark_mode"`
	PageSize   string `json:"page_size"`
}

const settingsFilePath = "settings.json"

// LoadSettings loads the app settings from a JSON file
func LoadSettings() (*AppSettings, error) {
	// Check if the settings file exists
	if _, err := os.Stat(settingsFilePath); os.IsNotExist(err) {
		// If it doesn't exist, return default settings
		return &AppSettings{IsDarkMode: false, PageSize: "10"}, nil
	}

	// Read the settings file
	fileBytes, err := os.ReadFile(settingsFilePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into the AppSettings struct
	var settings AppSettings
	err = json.Unmarshal(fileBytes, &settings)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// SaveSettings saves the app settings to a JSON file
func SaveSettings(settings *AppSettings) error {
	// Marshal the settings struct into JSON format
	fileBytes, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	// Write the JSON data to the settings file
	err = os.WriteFile(settingsFilePath, fileBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Variable to track current theme mode
var isDarkMode bool = false

// Function to apply the theme based on the current mode
func applyTheme() {
	if isDarkMode {
		fyne.CurrentApp().Settings().SetTheme(&themeVariant{Theme: theme.DefaultTheme(), variant: theme.VariantDark})
		darkModeIcon.SetIcon(theme.VisibilityIcon())
	} else {
		fyne.CurrentApp().Settings().SetTheme(&themeVariant{Theme: theme.DefaultTheme(), variant: theme.VariantLight})
		darkModeIcon.SetIcon(theme.VisibilityOffIcon())
	}
}

// Function to toggle between light and dark mode and save the settings
func toggleTheme(window fyne.Window) {
	// load settings
	saved_settings, err := LoadSettings()
	if err != nil {
		dialog.ShowInformation("Loading settings", "Error loading settings: "+err.Error(), window)
	}
	isDarkMode = !isDarkMode
	applyTheme()

	// Save the current theme setting
	settings := &AppSettings{PageSize: saved_settings.PageSize, IsDarkMode: isDarkMode}
	err = SaveSettings(settings)
	if err != nil {
		dialog.ShowInformation("User Settings", "Error saving settings", window)
	}
}

// FUNCTION TO TOGGLE THE PAGE SIZE
func updatePageSize(pageSize string, window fyne.Window) {
	// load settings
	saved_settings, err := LoadSettings()
	if err != nil {
		dialog.ShowInformation("Loading settings", "Error loading settings: "+err.Error(), window)
	}
	// Save the current theme setting
	settings := &AppSettings{IsDarkMode: saved_settings.IsDarkMode, PageSize: pageSize}

	err = SaveSettings(settings)
	if err != nil {
		dialog.ShowInformation("User Settings:Page size", "Error updating page size: "+err.Error(), window)
	}

}

// showSettings displays the settings view with user details and update options
func showSettings(window fyne.Window) {
	var user models.User
	var ImageFile *canvas.Image

	loadUser := func() {
		userID := helpers.CurrentUserID
		user = utils.GetUserByID(userID, window)
	}
	loadUser()

	ImageFile = canvas.NewImageFromFile("assets/profile.jpg")

	ImageFile.SetMinSize(fyne.NewSize(200, 200))
	ImageFile.FillMode = canvas.ImageFillContain // Ensure proper scaling
	// User details section (Right Side)
	userDetailsContainer := container.NewVBox()
	userDetailsContainer.Resize(fyne.NewSize(300, 200))

	// refresh user details in the UI
	refreshUserDetails := func() {
		loadUser()

		userDetailsContainer.Objects = []fyne.CanvasObject{
			container.NewGridWithColumns(2,
				widget.NewLabelWithStyle("Username: "+user.Username, fyne.TextAlignLeading, fyne.TextStyle{}),

				widget.NewLabelWithStyle("Phone: "+user.Phone, fyne.TextAlignLeading, fyne.TextStyle{})),
		}
		userDetailsContainer.Refresh()
	}

	refreshUserDetails()

	// Variable to track current theme mode

	content := container.NewHBox(
		ImageFile,
		container.NewVBox(
			userDetailsContainer,
			container.NewGridWithColumns(2,
				widget.NewButton("Update Details", func() {
					showUpdateUserDetailsDialog(window, user, refreshUserDetails)
				}),
				widget.NewButton("Change Password", func() {
					showChangePasswordDialog(window, user)
				}),
			),
			container.NewGridWithColumns(1,
				container.NewVBox(
					widget.NewLabel("Items Per Page"),
					widget.NewSelect([]string{"5", "10", "20", "30"}, func(value string) {
						updatePageSize(value, window)
					})),
			),
		),
	)

	// Center the content
	centeredContent := container.NewCenter(content)
	containerWithWidth := container.New(layout.NewStackLayout(), centeredContent)
	containerWithWidth.Resize(fyne.NewSize(600, 500))

	dialog.ShowCustom("Settings", "Close", content, window)
}

// showUpdateUserDetailsDialog displays a dialog for updating user details
func showUpdateUserDetailsDialog(window fyne.Window, user models.User, updateUserDetailsInView func()) {
	phoneEntry := widget.NewEntry()
	phoneEntry.SetPlaceHolder("Enter Phone")
	phoneEntry.SetText(user.Phone)

	formItems := []*widget.FormItem{
		{Text: "Phone", Widget: phoneEntry},
	}

	form := helpers.NewFixedWidthCenter(container.NewVBox(widget.NewForm(formItems...)), 300)

	dialog.ShowCustomConfirm("Update User Details", "Save", "Cancel", container.NewCenter(form), func(ok bool) {
		if !ok {
			return
		}

		user.Phone = phoneEntry.Text

		// Update user details in the database
		utils.UpdateUser(user, window)

		// Update user details in the view
		updateUserDetailsInView()

		dialog.ShowInformation("Success", "User details updated successfully.", window)

	}, window)
}

// showChangePasswordDialog displays a dialog for changing the user's password
func showChangePasswordDialog(window fyne.Window, user models.User) {
	currentPasswordEntry := widget.NewPasswordEntry()
	currentPasswordEntry.SetPlaceHolder("Enter Current Password")

	newPasswordEntry := widget.NewPasswordEntry()
	newPasswordEntry.SetPlaceHolder("Enter New Password")

	confirmPasswordEntry := widget.NewPasswordEntry()
	confirmPasswordEntry.SetPlaceHolder("Confirm New Password")

	formItems := []*widget.FormItem{
		{Text: "Current Password", Widget: currentPasswordEntry},
		{Text: "New Password", Widget: newPasswordEntry},
		{Text: "Confirm Password", Widget: confirmPasswordEntry},
	}

	form := helpers.NewFixedWidthCenter(container.NewVBox(widget.NewForm(formItems...)), 400)

	dialog.ShowCustomConfirm("Change Password", "Save", "Cancel", container.NewCenter(form), func(ok bool) {
		if !ok {
			return
		}

		oldPassword := currentPasswordEntry.Text
		newPassword := newPasswordEntry.Text
		confirmPassword := confirmPasswordEntry.Text

		// Verify that old password matches the stored password hash.
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
			dialog.ShowError(errors.New("old password is incorrect"), window)
			return
		}

		// Check if the new password is the same as the old password.
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(newPassword)); err == nil {
			dialog.ShowError(errors.New("new password cannot be the same as the old password"), window)
			return
		}

		if newPassword != confirmPassword {
			dialog.ShowError(errors.New("new password and confirm password do not match"), window)
			return
		}

		// Update password in the database
		err := auth.UpdateUserPassword(user.ID, newPassword, window)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		dialog.ShowInformation("Success", "Password changed successfully. Please log in again.", window)

	}, window)
}
