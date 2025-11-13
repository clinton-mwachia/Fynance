package main

import (
	"fynance/appTheme"
	"fynance/helpers"
	"fynance/utils"
	"fynance/views"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
)

func main() {
	application := app.New()
	window := application.NewWindow("Fynance")
	// connect to DB
	utils.ConnectDB("mongodb://localhost:27017", window)

	// Placeholder for functions that need to reference each other
	var showParameters, showIncome, showExpenses, showReport, showContact, showDashboard, showLogin func()

	// Load the settings on app startup
	settings, err := views.LoadSettings()
	if err != nil {
		dialog.ShowInformation("Loading settings", "Error loading settings: "+err.Error(), window)
	}

	if settings.IsDarkMode {
		fyne.CurrentApp().Settings().SetTheme(&appTheme.ThemeVariant{Theme: theme.DefaultTheme(), Variant: theme.VariantDark})
	} else {
		fyne.CurrentApp().Settings().SetTheme(&appTheme.ThemeVariant{Theme: theme.DefaultTheme(), Variant: theme.VariantLight})
	}

	// Function to show the details view
	showParameters = func() {
		sidebar := views.Sidebar(window, showParameters, showIncome,
			showExpenses, showReport, showContact, showDashboard, showLogin, helpers.CurrentUserID)
		parameters := views.ParametersView(window, helpers.CurrentUserID)
		window.SetContent(container.NewBorder(nil, nil, sidebar, nil, parameters))
	}

	// Function to show the income view
	showIncome = func() {
		sidebar := views.Sidebar(window, showParameters, showIncome,
			showExpenses, showReport, showContact, showDashboard, showLogin, helpers.CurrentUserID)
		income := views.IncomeView(window, helpers.CurrentUserID)
		window.SetContent(container.NewBorder(nil, nil, sidebar, nil, income))
	}

	// Function to show the expenses view
	showExpenses = func() {
		sidebar := views.Sidebar(window, showParameters, showIncome,
			showExpenses, showReport, showContact, showDashboard, showLogin, helpers.CurrentUserID)
		expenses := views.ExpenseView(window, helpers.CurrentUserID)
		window.SetContent(container.NewBorder(nil, nil, sidebar, nil, expenses))
	}

	// Function to show the report view
	showReport = func() {
		sidebar := views.Sidebar(window, showParameters, showIncome,
			showExpenses, showReport, showContact, showDashboard, showLogin, helpers.CurrentUserID)
		report := views.Report(window)
		window.SetContent(container.NewBorder(nil, nil, sidebar, nil, report))
	}

	// Function to show the contact view
	showContact = func() {
		sidebar := views.Sidebar(window, showParameters, showIncome,
			showExpenses, showReport, showContact, showDashboard, showLogin, helpers.CurrentUserID)
		contact := views.ContactView(window)
		window.SetContent(container.NewBorder(nil, nil, sidebar, nil, contact))
	}

	// Function to show the dashboard view
	showDashboard = func() {
		sidebar := views.Sidebar(window, showParameters, showIncome,
			showExpenses, showReport, showContact, showDashboard, showLogin, helpers.CurrentUserID)
		dashboard := views.Dashboard(window)
		window.SetContent(container.NewBorder(nil, nil, sidebar, nil, dashboard))
	}

	// Function to show the login view
	showLogin = func() {
		window.SetContent(views.LoginView(window, showDashboard))
	}

	// Initial view when the application starts
	showLogin()
	window.Resize(fyne.NewSize(600, 500))
	window.CenterOnScreen()
	window.ShowAndRun()
}
