package views

import (
	"fynance/helpers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func ParametersView(window fyne.Window) fyne.CanvasObject {
	userID := helpers.CurrentUserID
	header := Header(window)
	footer := Footer(window)
	content := container.NewAppTabs(
		container.NewTabItem("Income", IncomeDetailsView(window, userID)),
		container.NewTabItem("Expenses", ExpenseDetailsView(window, userID)),
	)
	return container.NewBorder(header, footer, nil, nil, content)
}
