package main

import (
	"embed"
	"fmt"
	"fynance/appTheme"
	router "fynance/routes"
	"fynance/utils"
	"fynance/views"
	"io/fs"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
)

//go:embed templates/*
var embeddedTemplates embed.FS

const targetFolder = "templates"

func initTemplates(window fyne.Window) {
	// Create the "templates" folder on the user's machine if missing
	if err := os.MkdirAll(targetFolder, 0755); err != nil {
		dialog.ShowInformation("Creating Templates Folder",
			fmt.Sprintf("Failed to create directory: %s"+err.Error()), window)
		return
	}

	// Read the files from the embedded binary system
	err := fs.WalkDir(embeddedTemplates, targetFolder, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		// Read file bytes from binary memory
		data, err := embeddedTemplates.ReadFile(path)
		if err != nil {
			return err
		}

		// Determine destination path (e.g., templates/template1.xlsx)
		destPath := filepath.Join(targetFolder, d.Name())

		// Only write the file if it does not already exist
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			err = os.WriteFile(destPath, data, 0644)
			if err != nil {
				dialog.ShowInformation("Creating Templates Folder",
					fmt.Sprintf("Failed to write %s: %v"+err.Error()+destPath), window)
			}
		}
		return nil
	})

	if err != nil {
		dialog.ShowInformation("Extracting Templates",
			fmt.Sprintf("Error extracting templates: %v", err), window)
	}
}

func main() {
	application := app.NewWithID("fynance.com")
	window := application.NewWindow("Fynance")

	initTemplates(window)
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
