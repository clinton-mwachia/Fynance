package main

import (
	"fynance/appTheme"
	router "fynance/routes"
	"fynance/utils"
	"fynance/views"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
)

func main() {
	application := app.NewWithID("fynance.com")
	window := application.NewWindow("Fynance")
	// connect to DB
	utils.ConnectDB("mongodb://localhost:27017", window)

	// Load settings
	loadTheme(window)

	// Create router
	r := router.NewRouter(window)

	// Start with login
	r.ShowLogin()

	window.Resize(fyne.NewSize(400, 300))
	window.CenterOnScreen()
	window.ShowAndRun()
}

func loadTheme(window fyne.Window) {

	settings, err := views.LoadSettings()

	if err != nil {
		dialog.ShowInformation("Loading settings", "Error loading settings: "+err.Error(), window)
		return
	}

	variant := theme.VariantLight

	if settings.IsDarkMode {
		variant = theme.VariantDark
	}

	fyne.CurrentApp().Settings().SetTheme(
		&appTheme.ThemeVariant{
			Theme:   theme.DefaultTheme(),
			Variant: variant,
		},
	)
}
