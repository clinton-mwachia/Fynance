package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ContactView(window fyne.Window) fyne.CanvasObject {
	header := Header(window)
	footer := Footer(window)

	title := widget.NewLabelWithStyle("Need an upgrade or an app made just for you?",
		fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabelWithStyle("Contact Us Through:", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	email := widget.NewLabel("Email: clintonmwachia9@gmail.com")
	phone1 := widget.NewLabel("Phone1: +254746646331")
	phone2 := widget.NewLabel("Phone2: +254738816913")
	whatsapp := widget.NewLabel("Whatsapp: +254746646331")

	content := container.NewVBox(title, subtitle, whatsapp, phone1, phone2, email)

	contentCard := widget.NewCard("", "", content)

	return container.NewBorder(header, footer, nil, nil, container.NewCenter(contentCard))
}
