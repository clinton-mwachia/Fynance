package views

import (
	"fynance/helpers"
	"fynance/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Add your specific view callbacks for daily, weekly, and monthly reports to the parameters
func Sidebar(window fyne.Window, showParameters, showIncome,
	showExpenses, showDailyReport, showWeeklyReport, showMonthlyReport, showContact, showDashboard,
	showLogin func(), userID primitive.ObjectID) *fyne.Container {

	// 1. Core items container
	sidebarItems := []fyne.CanvasObject{}

	// 2. Top navigation buttons
	sidebarItems = append(sidebarItems, widget.NewButton("Dashboard", showDashboard))
	sidebarItems = append(sidebarItems, widget.NewButton("Parameters", showParameters))
	sidebarItems = append(sidebarItems, widget.NewButton("Income", showIncome))
	sidebarItems = append(sidebarItems, widget.NewButton("Expenses", showExpenses))

	// 3. Create the inner Report options layout
	reportButtons := container.NewVBox(
		widget.NewButton("Daily Report", showDailyReport),
		widget.NewButton("Weekly Report", showWeeklyReport),
		widget.NewButton("Monthly Report", showMonthlyReport),
	)

	// 4. Wrap inner buttons inside an Accordion Item
	accordionItem := widget.NewAccordionItem("Reports", reportButtons)

	// Create the Accordion layout container
	reportsAccordion := widget.NewAccordion(accordionItem)

	// Append accordion right after Expenses
	sidebarItems = append(sidebarItems, reportsAccordion)

	// 5. Remaining bottom navigation links
	sidebarItems = append(sidebarItems, widget.NewButton("Contact", showContact))

	// 6. Flexible spatial filler pushing logout down
	sidebarItems = append(sidebarItems, layout.NewSpacer())

	// 7. System action item
	sidebarItems = append(sidebarItems, widget.NewButton("Logout", func() {
		utils.Logger("User Logged out", "SUCCESS", window)
		helpers.CurrentUserID = primitive.NilObjectID
		showLogin()
	}))

	// Return everything arranged vertically
	return container.NewVBox(sidebarItems...)
}
