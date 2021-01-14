package plugins

import (
	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"github.com/gofiber/fiber/v2"
	"stream-skybox.local/skybox"
)

type Plugin interface {
	GetSettingCanvas(fyne.Window) *widget.TabItem
	Crawl(chan<- skybox.Media)
	SetRoute(app *fiber.App)
}

type UpdateMediaListOption map[string]interface{}
