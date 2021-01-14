package conf

import (
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/widget"
	"github.com/gofiber/fiber/v2"
	"stream-skybox.local/skybox"
)

type mainConf struct {
	refreshCallback func()
}

func NewMainConf() *mainConf {
	return &mainConf{}
}

func (m *mainConf) SetRefreshCallback(f func()) {
	m.refreshCallback = f
}

func (m *mainConf) GetSettingCanvas(main fyne.Window) *widget.TabItem {
	box := widget.NewVBox()

	var button *widget.Button
	button = widget.NewButton("データ更新", func() {
		dialog.NewCustomConfirm("データを取得しますか？（しばらくかかります）", "実行", "中止", widget.NewVBox(), func(b bool) {
			if b {
				orig := button.Text
				button.Text = "取得中"
				button.Disable()
				go func() {
					defer func() {
						button.Text = orig
						button.Enable()
					}()

					if m.refreshCallback != nil {
						m.refreshCallback()
					}

				}()
			}
		}, main)
	})
	box.Append(button)

	check1 := widget.NewCheck("サンプル動画を非表示", func(value bool) {
		log.Println("Check set to", value)
	})
	box.Append(check1)

	check2 := widget.NewCheck("2D動画を非表示", func(value bool) {
		log.Println("Check set to", value)
	})
	box.Append(check2)

	return widget.NewTabItem("メイン設定", box)
}

func (m *mainConf) Crawl(chan<- skybox.Media) {
	// nop
}
func (m *mainConf) SetRoute(app *fiber.App) {

}
