package main

import (
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"stream-skybox.local/plugins"
	"stream-skybox.local/plugins/conf"
	"stream-skybox.local/plugins/dmm"
	"stream-skybox.local/plugins/slr"
	"stream-skybox.local/plugins/www360ch"
	"stream-skybox.local/skybox"
)

func checkEnv() {
	font := os.Getenv("FYNE_FONT")
	if font == "" {
		os.Setenv("FYNE_FONT", `C:\Windows\Fonts\YuGothM.ttc`)
	}
}

func main() {
	checkEnv()
	wsPort := 6888
	sbServer := skybox.NewServer(wsPort)

	gui := app.New()
	wnd := gui.NewWindow("StreamSkybox")

	mainConf := conf.NewMainConf()
	pgins := []plugins.Plugin{
		mainConf,
		www360ch.New360ch(),
		www360ch.NewPicmo1(),
		www360ch.NewPicmo2(),
		slr.NewSLR(),
		dmm.NewDMM(),
	}

	tabs := widget.NewTabContainer()
	for _, p := range pgins {
		tabs.Append(p.GetSettingCanvas(wnd))
	}
	tabs.SelectTabIndex(0)

	mainConf.SetRefreshCallback(func() {
		sbServer.SetGuard(true)
		defer sbServer.SetGuard(false)

		for _, p := range pgins {
			p.Crawl(sbServer.Library.GetInsert())
		}
	})
	tabs.SetTabLocation(widget.TabLocationLeading)
	wnd.SetContent(tabs)

	/// server
	fib := fiber.New()

	fib.Get("/*", func(c *fiber.Ctx) error {
		log.Println(c.Request())
		//log.Println(c.Request().Header.String())
		return c.Next()
	})

	// register fiber routing
	for _, p := range pgins {
		p.SetRoute(fib)
	}

	fib.Use("/", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})
	fib.Get("/socket.io/", websocket.New(sbServer.Callback))

	go func() {
		log.Fatal(fib.Listen(fmt.Sprintf(":%d", wsPort)))
	}()

	// gui
	wnd.ShowAndRun()
}
