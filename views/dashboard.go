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

	// fetch to tal incomes
	totalIncome := utils.TotalIncome(window)

	// Creat statistics boxes
	incomeBox := createStatisticsBox("Total Income", strconv.FormatFloat(totalIncome, 'f', 2, 64))

	// Layout for the statistics boxes
	statsContainer := container.New(layout.NewGridLayout(1),
		incomeBox,
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
