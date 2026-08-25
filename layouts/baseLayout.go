package layouts

import (
	"fynance/pages"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func BaseLayout(window fyne.Window, title string, content fyne.CanvasObject) *fyne.Container {
	header := pages.Header(window, title)
	footer := pages.Footer(window)

	return container.NewBorder(header, footer, nil, nil, content)
}
