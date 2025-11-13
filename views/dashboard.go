package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func Dashboard(window fyne.Window) *fyne.Container {
	header := Header(window)
	footer := Footer(window)

	label := widget.NewLabel("Dashboard")
	return container.NewBorder(header, footer, nil, nil, label)
}
