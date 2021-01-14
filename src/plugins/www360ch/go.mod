module stream-skybox.local/plugins/www360ch

go 1.14

require (
	fyne.io/fyne v1.4.2
	github.com/gofiber/fiber v1.14.6
	github.com/gofiber/fiber/v2 v2.1.0
	golang.org/x/text v0.3.4
	stream-skybox.local/common v0.0.0-00010101000000-000000000000
	stream-skybox.local/plugins v0.0.0-00010101000000-000000000000
	stream-skybox.local/skybox v0.0.0-00010101000000-000000000000
)

replace (
	stream-skybox.local/common => ../../common
	stream-skybox.local/plugins => ../
	stream-skybox.local/skybox => ../../skybox
)
