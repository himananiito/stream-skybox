package dmm

import (
	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"github.com/gofiber/fiber/v2"
	"stream-skybox.local/skybox"
)

type dmm struct {
}

func NewDMM() *dmm {
	return &dmm{}
}

func (w *dmm) GetSettingCanvas(main fyne.Window) *widget.TabItem {

	box := widget.NewVBox()
	box.Append(widget.NewLabel("実装中"))

	return widget.NewTabItem("DMM", box)
}

func (w *dmm) Crawl(chan<- skybox.Media) {
	// TODO
}

func (w *dmm) SetRoute(app *fiber.App) {
	// TODO
}
