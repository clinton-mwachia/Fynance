package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ParametersView(window fyne.Window, userID primitive.ObjectID) fyne.CanvasObject {
	header := Header(window)
	footer := Footer(window)
	content := container.NewAppTabs(
		container.NewTabItem("Income", IncomeDetailsView(window, userID)),
		container.NewTabItem("Expenses", ExpenseDetailsView(window, userID)),
	)
	return container.NewBorder(header, footer, nil, nil, content)
}
