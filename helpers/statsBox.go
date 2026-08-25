package helpers

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func StatBox(title, value string) fyne.CanvasObject {
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
