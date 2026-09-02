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

func Sidebar(window fyne.Window, showParameters, showIncome,
	showExpenses, showDailyReport, showWeeklyReport, showMonthlyReport, showContact, showDashboard,
	showLogin func()) *fyne.Container {

	sidebarItems := []fyne.CanvasObject{}

	sidebarItems = append(sidebarItems, widget.NewButton("Dashboard", showDashboard))
	sidebarItems = append(sidebarItems, widget.NewButton("Parameters", showParameters))
	sidebarItems = append(sidebarItems, widget.NewButton("Income", showIncome))
	sidebarItems = append(sidebarItems, widget.NewButton("Expenses", showExpenses))

	reportButtons := container.NewVBox(
		widget.NewButton("Daily", showDailyReport),
		widget.NewButton("Weekly", showWeeklyReport),
		widget.NewButton("Monthly", showMonthlyReport),
	)

	accordionItem := widget.NewAccordionItem("Reports", reportButtons)

	reportsAccordion := widget.NewAccordion(accordionItem)

	sidebarItems = append(sidebarItems, reportsAccordion)

	sidebarItems = append(sidebarItems, widget.NewButton("Contact", showContact))

	sidebarItems = append(sidebarItems, layout.NewSpacer())

	sidebarItems = append(sidebarItems, widget.NewButton("Logout", func() {
		utils.Logger("User Logged out", "SUCCESS", window)
		helpers.CurrentUserID = primitive.NilObjectID
		showLogin()
	}))

	return container.NewVBox(sidebarItems...)
}
