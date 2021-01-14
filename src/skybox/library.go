package skybox

import (
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Library struct {
	ignoreSample  bool
	propMtx       sync.RWMutex
	chInsertMedia chan Media
	db            *gorm.DB
}

func NewLibrary() *Library {

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Media{})

	chInsertMedia := make(chan Media, 10)
	lib := &Library{
		chInsertMedia: chInsertMedia,
		db:            db,
	}

	go func() {
		for m := range chInsertMedia {
			lib.InsertOrUpdate(&m)
		}
	}()
	return lib
}

func (lib *Library) InsertOrUpdate(m *Media) {
	var test Media
	lib.db.Find(&test, "id = ?", m.ID)
	if test.ID == m.ID {
		lib.db.Updates(m)
	} else {
		// set initial value
		if m.Size == 0 {
			m.Size = 1
		}
		if m.Duration == 0 {
			m.Duration = 1
		}
		if m.ThumbnailWidth == 0 {
			m.ThumbnailWidth = 320
		}
		if m.ThumbnailHeight == 0 {
			m.ThumbnailHeight = 180
		}
		m.OrientDegree = "0"
		m.RatioTypeFor2DScreen = "default"
		m.Exists = true
		if m.Width == 0 {
			m.Width = 3840
		}
		if m.Height == 0 {
			m.Height = 2160
		}

		lib.db.Create(m)
	}
}

func (lib *Library) GetInsert() chan Media {
	return lib.chInsertMedia
}

func (lib *Library) SetIgnoreSample(b bool) {
	lib.propMtx.Lock()
	defer lib.propMtx.Unlock()
	lib.ignoreSample = b
}
func (lib *Library) GetIgnoreSample() bool {
	lib.propMtx.RLock()
	defer lib.propMtx.RUnlock()
	return lib.ignoreSample
}

func (lib *Library) GetMedias() []Media {
	// TODO filtering
	var medias []Media
	lib.db.Find(&medias)
	return medias
}

func (lib *Library) GetPlaylist() []string {
	// TODO FIXME
	medias := lib.GetMedias()
	lst := make([]string, 0, len(medias))
	for _, m := range medias {
		lst = append(lst, m.ID)
	}
	return lst
}

func (lib *Library) hasMedia() bool {
	var m Media
	lib.db.Take(&m)
	return m.ID != ""
}
