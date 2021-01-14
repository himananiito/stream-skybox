module stream-skybox.local/plugins/slr

go 1.14

require (
	fyne.io/fyne v1.4.2
	github.com/PuerkitoBio/goquery v1.6.0
	github.com/gofiber/fiber v1.14.6
	github.com/gofiber/fiber/v2 v2.2.5
	stream-skybox.local/common v0.0.0-00010101000000-000000000000
	stream-skybox.local/skybox v0.0.0-00010101000000-000000000000
)

replace (
	stream-skybox.local/common => ../../common
	stream-skybox.local/skybox => ../../skybox
)
