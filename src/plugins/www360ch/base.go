package www360ch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/unicode/norm"
	"stream-skybox.local/common"
	"stream-skybox.local/plugins"
	"stream-skybox.local/skybox"
)

type www360ch struct {
	host          string
	cookieHost    string
	cookieNames   []string
	optionCookies func() []string
	getCanvas     func(*www360ch) *widget.TabItem
	mtx           sync.RWMutex
	isPurchased   bool
	isWatchList   bool
	isNewVideo    bool
	segmentID     int
}

func New360ch() *www360ch {
	return &www360ch{
		host:        "www.360ch.tv",
		cookieHost:  "360ch.tv",
		cookieNames: []string{"ch360pt", "ch360t"},
		getCanvas:   getWww360chCanvas,
		segmentID:   1,
	}
}

func NewPicmo1() *www360ch {
	return &www360ch{
		host:        "www.picmo.jp",
		cookieHost:  "picmo.jp",
		cookieNames: []string{"picmopt", "picmot"},
		getCanvas:   getPicmoCanvas,
		segmentID:   1,
	}
}

func NewPicmo2() *www360ch {
	return &www360ch{
		host:          "r.picmo.jp",
		cookieHost:    "picmo.jp",
		cookieNames:   []string{"picmopt", "picmot"},
		optionCookies: picmo2OptionCookies,
		getCanvas:     getPicmo2Canvas,
		segmentID:     2,
	}
}

const (
	propPurchased = iota
	propWatchList
	propNewVideo
)

func (w *www360ch) setProperty(prop int, b bool) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	switch prop {
	case propPurchased:
		w.isPurchased = b
	case propWatchList:
		w.isWatchList = b
	case propNewVideo:
		w.isNewVideo = b
	}
}
func (w *www360ch) getProperty(prop int) bool {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	var b bool
	switch prop {
	case propPurchased:
		b = w.isPurchased
	case propWatchList:
		b = w.isWatchList
	case propNewVideo:
		b = w.isNewVideo
	}
	return b
}

func picmo2OptionCookies() []string {
	return []string{fmt.Sprintf("vrfy-2=%d", time.Now().Unix())}
}

func (w *www360ch) UpdateMediaList(opt plugins.UpdateMediaListOption) {

}

var tsRouteFormat = "/stream/%s/:id/:url.ts"
var keyRouteFormat = "/stream/%s/:id/:url.key"
var m3u8RouteFormat = "/stream/%s/:id/:url.m3u8"

func (w *www360ch) SetRoute(app *fiber.App) {

	app.Get(fmt.Sprintf(tsRouteFormat, w.host), func(c *fiber.Ctx) error {
		uri := common.B64Dec(c.Params("url"))
		// redirect not work
		// c.Location(uri)
		// c.SendStatus(fiber.StatusFound)
		// return nil
		res, err := common.GetClient().Get(uri)
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()

		_, err = io.Copy(c.Type("video/mp2t"), res.Body)
		return err
	})

	app.Get(fmt.Sprintf(keyRouteFormat, w.host), func(c *fiber.Ctx) error {
		uri := common.B64Dec(c.Params("url"))
		// redirect not work
		// c.Location(uri)
		// c.SendStatus(fiber.StatusFound)
		// return nil
		res, err := common.GetClient().Get(uri)
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()

		_, err = io.Copy(c.Type("application/octet-stream"), res.Body)
		return err
	})

	app.Get(fmt.Sprintf(m3u8RouteFormat, w.host), func(c *fiber.Ctx) error {

		host := string(c.Request().Host())
		if host == "" {
			log.Println("host empty")
			return errors.New("host not set")
		}

		uri := common.B64Dec(c.Params("url"))

		//log.Println(uri)
		//log.Println(c.Request().Header.String())

		m3u8URL, err := url.Parse(uri)
		if err != nil {
			log.Println(err)
			return err
		}

		s, err := w.getM3U8(m3u8URL, c.Params("id"), host)
		if err != nil {
			log.Println(err)
			return err
		}

		return c.Type("application/x-mpegURL").SendString(s)
	})
}

func (w *www360ch) setCookie(req *http.Request) {
	var cookie string
	cookie = common.GetBrowserCookie(w.cookieHost, w.cookieNames...)
	if w.optionCookies != nil {
		cookies := w.optionCookies()
		if cookie != "" {
			cookie += "; " + strings.Join(cookies, "; ")
		} else {
			cookie = strings.Join(cookies, "; ")
		}
	}

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
}

func (w *www360ch) crawlSingle(endpoint string, chMedia chan<- skybox.Media) (enableVideos int, allVideos int, morePages bool) {
	uri := fmt.Sprintf("https://%s/%s", w.host, endpoint)
	urlMain, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		panic(err)
	}
	w.setCookie(req)

	client := common.GetClient()
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return 0, 0, false
	}
	defer res.Body.Close()
	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var data response
	err = json.Unmarshal(bs, &data)
	if err != nil {
		log.Println(err)
		return
	}

	allVideos = len(data.Result.Videos)

	for _, v := range data.Result.Videos {
		if !v.enabled() {
			continue
		}

		for vIndex, info := range v.getInfo() {
			if info.uri == "" {
				continue
			}

			//id := generateId(fmt.Sprintf("%s:%d:%d:%d", d.getHost(), v.ID, v.Version, i))
			name := info.name
			name = regexp.MustCompile(`^《.{1,6}》`).ReplaceAllString(name, "")
			switch info.vrSetting {
			case skybox.VrSetting2DScreen:
				name = "(2D)" + name
			default:
				if 0 < v.QualityType {
					name = "(HQ)" + name
				} else {
					switch info.vrSetting {
					case skybox.VrSettingTopBottom360, skybox.VrSettingLeftRight180:
						name = "(3D)" + name
					}
				}
			}

			name = norm.NFKC.String(name)
			name = regexp.MustCompile(`\p{Cc}+`).ReplaceAllString(name, " ")
			name = regexp.MustCompile(`\p{Zs}+`).ReplaceAllString(name, " ")
			name = regexp.MustCompile(`\p{So}+`).ReplaceAllString(name, "")
			name = strings.ReplaceAll(name, ".", "")

			// [FIXME]
			id := fmt.Sprintf("%v:%v:%v", w.host, v.ID, vIndex)

			urlSub, err := url.Parse(v.ThumbnailURL)
			if err != nil {
				log.Println(err)
			}
			thumbnail := urlMain.ResolveReference(urlSub).String()

			uri := fmt.Sprintf(m3u8RouteFormat, w.host)
			uri = strings.ReplaceAll(uri, ":id", id)
			uri = strings.ReplaceAll(uri, ":url", common.B64Enc(info.uri))

			chMedia <- skybox.Media{
				ID:                   id,
				Name:                 name,
				Duration:             int64(v.Time) * 1000,
				Size:                 int64(v.ID * 1024 / 1000),
				URL:                  uri,
				Thumbnail:            thumbnail,
				ThumbnailWidth:       186,
				ThumbnailHeight:      120,
				LastModified:         v.SA,
				DefaultVRSetting:     info.vrSetting,
				UserVRSetting:        info.vrSetting,
				Subtitles:            []interface{}{},
				Width:                3840,
				Height:               2160,
				OrientDegree:         "0",
				RatioTypeFor2DScreen: "default",
				RotationFor2DScreen:  0,
				Exists:               true,
				IsBadMedia:           false,
				AddedTime:            v.SA,
			}
			enableVideos++
		}
	}

	morePages = data.Result.Paging.P <= data.Result.Paging.MP
	return
}

func (w *www360ch) crawlPurchased(chMedia chan<- skybox.Media) {
	itemNum := 30
	for i := 1; i < 5; i++ {
		endpoint := fmt.Sprintf("ajax/user-video/index/page/%d/itemNum/%d/offset/0/orderType/0/order/0/segmentId/%d", i, itemNum, w.segmentID)
		_, _, morePages := w.crawlSingle(endpoint, chMedia)
		if !morePages {
			break
		}
	}
}

func (w *www360ch) crawlWatchList(chMedia chan<- skybox.Media) {
	itemNum := 30
	for i := 1; i < 5; i++ {
		endpoint := fmt.Sprintf("ajax/video/my-favorite-video/page/%d/itemNum/%d/offset/0/segmentId/%d", i, itemNum, w.segmentID)
		_, _, morePages := w.crawlSingle(endpoint, chMedia)
		if !morePages {
			break
		}
	}
}

func (w *www360ch) crawlNewVideo(chMedia chan<- skybox.Media) {
	itemNum := 30
	for i := 1; i < 5; i++ {
		endpoint := fmt.Sprintf("ajax/video/get-new-videos/page/%d/itemNum/%d/targetType/all", i, itemNum)
		_, _, morePages := w.crawlSingle(endpoint, chMedia)
		if !morePages {
			break
		}
	}
}

func (w *www360ch) Crawl(chMedia chan<- skybox.Media) {
	if w.getProperty(propPurchased) {
		w.crawlPurchased(chMedia)
	}
	if w.getProperty(propWatchList) {
		w.crawlWatchList(chMedia)
	}
	if w.getProperty(propNewVideo) {
		w.crawlNewVideo(chMedia)
	}
}

func (w *www360ch) GetSettingCanvas(main fyne.Window) *widget.TabItem {
	return w.getCanvas(w)
}

func getWww360chCanvas(w *www360ch) *widget.TabItem {
	return w.getCanvasCommon("360Channel", fmt.Sprintf("https://%s/", w.host))
}

func getPicmoCanvas(w *www360ch) *widget.TabItem {
	return w.getCanvasCommon("PICMO/一般", fmt.Sprintf("https://%s/", w.host))
}

func getPicmo2Canvas(w *www360ch) *widget.TabItem {
	return w.getCanvasCommon("PICMO/その他", fmt.Sprintf("https://%s/", w.host))
}

func (w *www360ch) getCanvasCommon(tabTitle, uriString string) *widget.TabItem {
	uri, err := url.Parse(uriString)
	if err != nil {
		panic(err)
	}
	link := widget.NewHyperlink("PCのブラウザでログインして下さい", uri)

	check1 := widget.NewCheck("購入済を取得", func(value bool) {
		w.setProperty(propPurchased, value)
	})
	check1.Checked = w.getProperty(propPurchased)

	check2 := widget.NewCheck("ウォッチリストを取得", func(value bool) {
		w.setProperty(propWatchList, value)
	})
	check2.Checked = w.getProperty(propWatchList)

	check3 := widget.NewCheck("新着動画を取得", func(value bool) {
		w.setProperty(propNewVideo, value)
	})
	check3.Checked = w.getProperty(propNewVideo)

	return widget.NewTabItem(tabTitle, widget.NewVBox(link, check1, check2, check3))
}
