package helpers

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func StatBox(title, value string) fyne.CanvasObject {

	border := canvas.NewRectangle(color.Gray{0x99})
	border.StrokeWidth = 2
	border.StrokeColor = color.Gray{0x99}
	border.FillColor = color.Transparent

	content := container.NewVBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(value, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	paddedContent := container.NewPadded(content)

	return container.NewStack(border, paddedContent)
}
