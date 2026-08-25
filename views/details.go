package views

import (
	"fynance/helpers"
	"fynance/layouts"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func ParametersView(window fyne.Window) *fyne.Container {
	userID := helpers.CurrentUserID

	content := container.NewAppTabs(
		container.NewTabItem("Income", IncomeDetailsView(window, userID)),
		container.NewTabItem("Expenses", ExpenseDetailsView(window, userID)),
	)
	return layouts.BaseLayout(window, "Parameters", content)
}
