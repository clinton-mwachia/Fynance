package appTheme

import (
	"image/color"

	"fyne.io/fyne/v2"
)

type ThemeVariant struct {
	fyne.Theme

	Variant fyne.ThemeVariant
}

func (f *ThemeVariant) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.Variant)
}
