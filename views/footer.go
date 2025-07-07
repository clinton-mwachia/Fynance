package views

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func Footer(window fyne.Window) *fyne.Container {

	// get current year
	currentTime := time.Now()
	currentYear := currentTime.Year()

	// Footer container
	footer := container.NewCenter(
		widget.NewLabel(fmt.Sprintf("© %d Moshe Crafts. All Rights Reserved.", currentYear)),
	)
	return footer
}
