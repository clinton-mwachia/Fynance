package views

import (
	"fynance/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Sidebar(window fyne.Window, showParameters, showIncome,
	showExpenses, showReport, showContact, showDashboard,
	showLogin func(), userID primitive.ObjectID) *fyne.Container {

	// Define buttons with their labels and actions
	buttonConfigs := []struct {
		label    string
		callback func()
	}{
		{"Dashboard", showDashboard},
		{"Parameters", showParameters},
		{"Income", showIncome},
		{"Expenses", showExpenses},
		{"Report", showReport},
		{"Contact", showContact},
	}

	// Create buttons from configurations
	var buttons []fyne.CanvasObject

	// Add other buttons below
	for _, config := range buttonConfigs {
		buttons = append(buttons, widget.NewButton(config.label, config.callback))
	}

	// Add spacer and logout button
	buttons = append(buttons, layout.NewSpacer())
	buttons = append(buttons, widget.NewButton("Logout", func() {
		utils.Logger("User Logged out", "SUCCESS", window)

		showLogin()
	}))

	// Return the sidebar container
	return container.NewVBox(buttons...)
}
