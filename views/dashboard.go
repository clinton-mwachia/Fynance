package views

import (
	"fynance/utils"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func Dashboard(window fyne.Window) *fyne.Container {
	header := Header(window)
	footer := Footer(window)

	// fetch to totals
	totalIncome := utils.TotalIncome(window)
	totalExpenses := utils.TotalExpenses(window)
	balance := totalIncome - totalExpenses

	// Creat statistics boxes
	totalIncomeBox := createStatisticsBox("Total Income", strconv.FormatFloat(totalIncome, 'f', 2, 64))
	totalExpenseBox := createStatisticsBox("Total Expenses", strconv.FormatFloat(totalExpenses, 'f', 2, 64))
	balanceBox := createStatisticsBox("Balance", strconv.FormatFloat(balance, 'f', 2, 64))

	// Layout for the statistics boxes
	statsContainer := container.New(layout.NewGridLayout(3),
		totalIncomeBox,
		totalExpenseBox,
		balanceBox,
	)

	return container.NewBorder(header, footer, nil, nil, statsContainer)
}

// createStatisticsBox creates a statistics display box
func createStatisticsBox(title, value string) fyne.CanvasObject {
	// Create a border rectangle
	border := canvas.NewRectangle(color.Gray{0x99})
	border.StrokeWidth = 2
	border.StrokeColor = color.Gray{0x99}
	border.FillColor = color.Transparent

	// Create the content
	content := container.NewVBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(value, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	// Add padding around the content
	paddedContent := container.NewPadded(content)

	// Create a container that will show both the border and content
	return container.NewStack(border, paddedContent)
}
