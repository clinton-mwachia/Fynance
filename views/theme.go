package views

import (
	"image/color"

	"fyne.io/fyne/v2"
)

type themeVariant struct {
	fyne.Theme

	variant fyne.ThemeVariant
}

func (f *themeVariant) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.variant)
}
