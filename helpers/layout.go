package helpers

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type fixedWidthLayout struct {
	width float32
}

func (f *fixedWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, obj := range objects {
		obj.Resize(fyne.NewSize(f.width, obj.MinSize().Height))
		obj.Move(fyne.NewPos((size.Width-f.width)/2, (size.Height-obj.MinSize().Height)/2))
	}
}

func (f *fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minHeight := float32(0)
	for _, obj := range objects {
		minHeight += obj.MinSize().Height
	}
	return fyne.NewSize(f.width, minHeight)
}

// NewFixedWidthCenter creates a container with a fixed width that is centered.
func NewFixedWidthCenter(content fyne.CanvasObject, width float32) *fyne.Container {
	layout := &fixedWidthLayout{width: width}
	return container.New(layout, content)
}
