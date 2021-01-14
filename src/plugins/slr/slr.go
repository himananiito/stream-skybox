package slr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"stream-skybox.local/common"
	"stream-skybox.local/skybox"
)

type slr struct {
	pageConfs []pageConf
	mtx       sync.RWMutex
}

type pageConf struct {
	uri string
	num int
}

func NewSLR() *slr {
	return &slr{}
}

func (w *slr) setPageConfURI(uri string, index int) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	if index < len(w.pageConfs) {
		w.pageConfs[index].uri = uri
	}
}

func (w *slr) setPageConfNum(num, index int) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	if index < len(w.pageConfs) {
		w.pageConfs[index].num = num
	}
}
func (w *slr) getPageConfLength() int {
	w.mtx.RLock()
	defer w.mtx.RUnlock()
	return len(w.pageConfs)
}
func (w *slr) getPageConf(i int) pageConf {
	w.mtx.RLock()
	defer w.mtx.RUnlock()
	if i < len(w.pageConfs) {
		return w.pageConfs[i]
	}
	return pageConf{}
}

func (w *slr) GetSettingCanvas(main fyne.Window) *widget.TabItem {

	w.pageConfs = []pageConf{}

	vbox := widget.NewVBox()
	for i := 0; i < 4; i++ {
		func(index int) {

			w.pageConfs = append(w.pageConfs, pageConf{})

			entry1 := widget.NewEntry()
			entry1.SetPlaceHolder("リスト取得起点URL")
			entry1.OnChanged = func(s string) {
				w.setPageConfURI(s, index)
			}

			entry2 := widget.NewEntry()
			entry2.SetPlaceHolder("ページ数")
			entry2.OnChanged = func(s string) {
				n, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					log.Println(n, err)
					entry2.Text = "0"
					w.setPageConfNum(0, index)
				} else {
					w.setPageConfNum(int(n), index)
				}
			}

			hbox := widget.NewHBox(entry2, widget.NewLabel("ページ取得（最大）"))

			card1 := widget.NewCard("", "", widget.NewVBox(entry1, hbox))

			vbox.Append(card1)
		}(i)
	}

	return widget.NewTabItem("SLR.com", vbox)
}

var pathRouterInt = "/stream/slr/:id/:url.int"

func (w *slr) SetRoute(app *fiber.App) {
	app.Get(pathRouterInt, func(c *fiber.Ctx) error {
		uri := common.B64Dec(c.Params("url"))

		log.Println(uri)
		log.Println(c.Request().Header.String())

		uri = getSourceSLR(uri)
		if uri == "" {
			return errors.New("no source found")
		}
		c.Location(uri)
		c.SendStatus(fiber.StatusFound)
		return nil
	})
}

func getSourceSLR(uri string) string {

	// 中間ページをダウンロード

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		panic(err)
	}

	cookie := common.GetBrowserCookie("sexlikereal.com", "sess")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	client := common.GetClient()

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Println(err)
		return ""
	}
	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return ""
	}
	mParent := regexp.MustCompile(`(?s)window\.vrPlayerSettings\s*=\s*({.*?});`).FindSubmatch(bs)
	if len(mParent) < 2 {
		return "vrPlayerSettings not found"
	}

	var angle int
	if m := regexp.MustCompile(`(?s)"angle"\s*:\s*(\d+)`).FindSubmatch(mParent[1]); 2 <= len(m) {
		n, err := strconv.ParseInt(string(m[1]), 10, 64)
		if err != nil {
			log.Fatalln(err)
		}
		angle = int(n)
	}

	var format string
	if m := regexp.MustCompile(`(?s)"format"\s*:\s*"(.*?)"`).FindSubmatch(mParent[1]); 2 <= len(m) {
		format = string(m[1])
	}

	var fullVideo int64
	if m := regexp.MustCompile(`(?s)"fullVideo"\s*:\s*(\d+)`).FindSubmatch(mParent[1]); 2 <= len(m) {
		fullVideo, _ = strconv.ParseInt(string(m[1]), 10, 64)
	}
	log.Println("fullVideo", fullVideo)

	m := regexp.MustCompile(`(?s)"src"\s*:\s*(\[.*?\])`).FindSubmatch(mParent[1])
	if len(m) < 2 {
		log.Println("src not found")
		return ""
	}
	//log.Println(string(m[1]))

	type source struct {
		URL      string `json:"url"`
		MimeType string `json:"mimeType"`
		Quality  string `json:"quality"`
		Encoding string `json:"encoding"`
	}

	sources := []source{}
	if err := json.Unmarshal(m[1], &sources); err != nil {
		log.Println(err)
		return ""
	}

	var movieSrc string
	var quality int64
	for _, src := range sources {
		q, err := strconv.ParseInt(strings.TrimSuffix(src.Quality, "p"), 10, 64)
		if err != nil {
			log.Println(err)
			continue
		}

		var update bool
		switch format {
		case "sbs2l", "sbs", "LR", "mono":
			if q <= 2160 && quality < q {
				update = true
			} else if quality == q && strings.ToLower(src.Encoding) == "h265" {
				update = true
			}

		case "ab2l", "ab", "TB":
			if q < 2*2160 && quality < q {
				update = true
			} else if quality == q && strings.ToLower(src.Encoding) == "h265" {
				update = true
			}
		}

		if update {
			quality = q
			movieSrc = src.URL
		}
	}

	log.Println(movieSrc)

	log.Println(angle, format, movieSrc, quality)
	return movieSrc
}

func (w *slr) Crawl(chMedia chan<- skybox.Media) {
	//host := "www.sexlikereal.com"
	client := common.GetClient()

	// FIXME
	//gen := newLocalURLGenerator()
	for confIndex := 0; confIndex < w.getPageConfLength(); confIndex++ {
		pageConf := w.getPageConf(confIndex)
		uri := pageConf.uri
		num := pageConf.num

		var prevPage int
		for i := 0; i < num; i++ {

			u, err := url.Parse(uri)
			if err != nil {
				log.Println(err)
				continue
			}
			switch u.Scheme {
			case "http", "https":
			default:
				continue
			}
			uri = u.String()

			log.Println("start", uri)

			urlMain, err := url.Parse(uri)
			if err != nil {
				log.Println(err)
				//return list
			}

			req, err := http.NewRequest("GET", uri, nil)
			if err != nil {
				panic(err)
			}

			//if false {
			cookie := common.GetBrowserCookie("sexlikereal.com", "sess")
			if cookie != "" {
				req.Header.Set("Cookie", cookie)
			}
			//	}

			req.Header.Set("Accept", "*/*")

			res, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer res.Body.Close()

			log.Println(res)

			doc, err := goquery.NewDocumentFromReader(res.Body)
			if err != nil {
				panic(err)
			}

			var currentPage int
			doc.Find(`a.o-btn--base.u-disabled`).Each(func(_ int, s *goquery.Selection) {
				if m := regexp.MustCompile(`\A\D*(\d+)\D*\z`).FindStringSubmatch(s.Text()); 1 < len(m) {
					if n, err := strconv.ParseInt(m[1], 10, 64); err != nil {
						log.Println(err)
					} else {
						currentPage = int(n)
					}
				}
			})

			if 0 < prevPage {
				if prevPage+1 != currentPage {
					break
				}
			}

			doc.Find(`article.c-grid-item--scene a[href][data-like-id]`).Each(func(_ int, s *goquery.Selection) {
				page, _ := s.Attr("href")
				urlSub, err := url.Parse(page)
				if err != nil {
					log.Println(err)
					return
				}
				pageURL := urlMain.ResolveReference(urlSub).String()

				log.Println(pageURL)

				//getSourceSLR(pageURL)

				videoID, _ := s.Attr("data-like-id")

				s = s.Find("img[data-srcset][alt]").First()
				imageSrc, _ := s.Attr("data-srcset")
				title, _ := s.Attr("alt")
				if imageSrc == "" {
					return
				}

				//id := generateId(fmt.Sprintf("%s:%d", host, videoID))
				id := fmt.Sprintf("slr:%v", videoID)

				// TODO 情報があるならURLとVRSettingを解決する
				//uri := gen.generateLocalURI(host, id, pageURL, "int")
				log.Println(videoID, title, id, pageURL)

				mediaURI := pathRouterInt
				mediaURI = strings.ReplaceAll(mediaURI, ":id", id)
				mediaURI = strings.ReplaceAll(pathRouterInt, ":url", common.B64Enc(pageURL))

				chMedia <- skybox.Media{
					ID:               id,
					Name:             title,
					URL:              mediaURI,
					Thumbnail:        imageSrc,
					DefaultVRSetting: skybox.VrSettingLeftRight180,
					UserVRSetting:    skybox.VrSettingLeftRight180,
					Width:            3840,
					Height:           2160,
				}
			})

			var nextUri string
			if 0 < currentPage {
				doc.Find(`a[href].o-btn--outlined.u-transition--base:not(.u-relative)`).Each(func(_ int, s *goquery.Selection) {

					if !strings.Contains(strings.ToLower(s.Text()), "next") {
						return
					}

					href, _ := s.Attr("href")
					urlSub, err := url.Parse(href)
					if err != nil {
						log.Println(err)
						return
					}
					nextUri = urlMain.ResolveReference(urlSub).String()
				})
			}
			if nextUri == "" || nextUri == uri {
				break
			}

			prevPage = currentPage

		}
	}

	//return list
}
