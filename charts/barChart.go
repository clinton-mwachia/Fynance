package charts

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type DataPoint struct {
	Count float64
	Color color.Color
}

type BarChart struct {
	container *fyne.Container
	maxHeight float32
	barWidth  float32
	spacing   float32
}

func NewBarChart(maxHeight, barWidth, spacing float32) *BarChart {
	// Create main container with border
	mainBorder := canvas.NewRectangle(color.Gray{0x99})
	mainBorder.StrokeWidth = 2
	mainBorder.StrokeColor = color.Gray{0x99}
	mainBorder.FillColor = color.Transparent

	innerContainer := container.NewHBox()
	paddedContainer := container.NewPadded(innerContainer)

	mainContainer := container.NewStack(mainBorder, paddedContainer)

	return &BarChart{
		container: mainContainer,
		maxHeight: maxHeight,
		barWidth:  barWidth,
		spacing:   spacing,
	}
}

func (b *BarChart) UpdateData(data map[string]DataPoint) {
	// Get the inner HBox container
	innerContainer := b.container.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container)
	innerContainer.Objects = nil

	var totalCount float64
	for _, v := range data {
		totalCount += v.Count
	}
	if totalCount == 0 {
		totalCount = 1
	}

	for label, value := range data {
		height := b.maxHeight * (float32(value.Count) / float32(totalCount))

		bar := canvas.NewRectangle(value.Color)
		bar.SetMinSize(fyne.NewSize(b.barWidth, height))

		spacer := canvas.NewRectangle(color.Transparent)
		spacer.SetMinSize(fyne.NewSize(b.barWidth, b.maxHeight-height))

		barContainer := container.NewVBox(spacer, bar)

		labelWidget := widget.NewLabel(label)

		barSection := container.NewVBox(
			barContainer,
			container.NewHBox(layout.NewSpacer(), labelWidget, layout.NewSpacer()),
		)

		paddedContainer := container.NewPadded(barSection)

		innerContainer.Add(paddedContainer)
	}
	b.container.Refresh()
}

func (b *BarChart) Container() *fyne.Container {
	return b.container
}
