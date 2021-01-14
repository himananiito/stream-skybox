module stream-skybox

go 1.14

require (
	fyne.io/fyne v1.4.2
	github.com/PuerkitoBio/goquery v1.6.0
	github.com/andybalholm/brotli v1.0.1 // indirect
	github.com/gofiber/fiber/v2 v2.2.5
	github.com/gofiber/websocket/v2 v2.0.2
	github.com/googollee/go-engine.io v1.4.2 // indirect
	github.com/googollee/go-socket.io v1.4.4 // indirect
	github.com/klauspost/compress v1.11.3 // indirect
	github.com/mattn/go-sqlite3 v1.14.5
	github.com/savsgio/gotils v0.0.0-20200909101946-939aa3fc74fb // indirect
	github.com/zellyn/kooky v0.0.0-20201108220156-bec09c12c339
	golang.org/x/text v0.3.4
	gorm.io/gorm v1.20.9 // indirect
	stream-skybox.local/common v0.0.0-00010101000000-000000000000
	stream-skybox.local/plugins v0.0.0-00010101000000-000000000000
	stream-skybox.local/plugins/conf v0.0.0-00010101000000-000000000000
	stream-skybox.local/plugins/dmm v0.0.0-00010101000000-000000000000
	stream-skybox.local/plugins/slr v0.0.0-00010101000000-000000000000
	stream-skybox.local/plugins/www360ch v0.0.0-00010101000000-000000000000
	stream-skybox.local/skybox v0.0.0-00010101000000-000000000000
)

replace (
	stream-skybox.local/common => ./common
	stream-skybox.local/plugins => ./plugins
	stream-skybox.local/plugins/conf => ./plugins/conf
	stream-skybox.local/plugins/dmm => ./plugins/dmm
	stream-skybox.local/plugins/slr => ./plugins/slr
	stream-skybox.local/plugins/www360ch => ./plugins/www360ch
	stream-skybox.local/skybox => ./skybox
)
