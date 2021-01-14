package www360ch

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unsafe"

	"stream-skybox.local/common"
)

func (w *www360ch) getM3U8(m3u8URL *url.URL, id, host string) (string, error) {
	req, err := http.NewRequest("GET", m3u8URL.String(), nil)
	if err != nil {
		return "", err
	}
	w.setCookie(req)

	client := common.GetClient()

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	s := *(*string)(unsafe.Pointer(&bs))

	s = regexp.MustCompile(`(?m)^[^#].*\.ts.*$`).ReplaceAllStringFunc(s, func(s string) string {
		urlSub, err := url.Parse(s)
		if err != nil {
			panic(err)
		}

		orig := m3u8URL.ResolveReference(urlSub).String()
		uri := "http://" + host + fmt.Sprintf(tsRouteFormat, w.host)
		uri = strings.ReplaceAll(uri, ":id", "id")
		uri = strings.ReplaceAll(uri, ":url", common.B64Enc(orig))
		return uri
	})

	s = regexp.MustCompile(`(?m)^[^#].*\.m3u8.*$`).ReplaceAllStringFunc(s, func(s string) string {
		urlSub, err := url.Parse(s)
		if err != nil {
			panic(err)
		}
		s = m3u8URL.ResolveReference(urlSub).String()
		return common.B64Enc(s) + ".m3u8"
	})
	s = regexp.MustCompile(`(?s)\.m3u8\r?\n.*\z`).ReplaceAllString(s, ".m3u8\n")

	mIdx := regexp.MustCompile(`#EXT-X-KEY:METHOD=AES-128,URI="(.*?)"`).FindStringSubmatchIndex(s)
	if 0 < len(mIdx) {
		urlSub, err := url.Parse(s[mIdx[2]:mIdx[3]])
		if err != nil {
			panic(err)
		}
		orig := m3u8URL.ResolveReference(urlSub).String()
		uri := "http://" + host + fmt.Sprintf(keyRouteFormat, w.host)
		uri = strings.ReplaceAll(uri, ":id", id)
		uri = strings.ReplaceAll(uri, ":url", common.B64Enc(orig))

		s = s[:mIdx[2]] + uri + s[mIdx[3]:]
	}

	return s, nil
}
