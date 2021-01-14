package www360ch

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"stream-skybox.local/skybox"
)

// r.PLATFORM_RIFT = 101,
// r.PLATFORM_GEARVR = 102,
// r.PLATFORM_PC_WEB = 201,
// r.PLATFORM_PC_WEB_IE = 202,
// r.PLATFORM_MOBILE_WEB = 301,
// r.PLATFORM_MOBILE_WEB_IOS = 302,
// r.PLATFORM_MOBILE_WEB_IOS_HLS = 301,
// r.PLATFORM_MOBILE_WEB_IOS_LOW_MP4 = 303,
// r.PLATFORM_MOBILE_WEB_IOS_LOW_HLS = 304,
// r.PLATFORM_PLANE = 401,
// r.PLATFORM_PLANE_MOBILE_WEB = 501,
// r.PLATFORM_PLANE_MOBILE_WEB_IOS = 502,
// r.PLATFORM_PLANE_MOBILE_WEB_IOS_HLS = 501,
// r.PLATFORM_PLANE_MOBILE_WEB_IOS_LOW_MP4 = 503,
// r.PLATFORM_PLANE_MOBILE_WEB_IOS_LOW_HLS = 504,
// r.PLATFORM_STEREO_VR_4K = 601,
// r.PLATFORM_STEREO_VR_2K = 602,
// r.PLATFORM_VR_4K_60FPS_15M = 800,
// r.PLATFORM_VR_4K_60FPS_12M = 801,
// r.PLATFORM_VR_2P5K_60FPS = 802,
// r.PLATFORM_STEREO_VR_4K_60FPS_15M = 810,
// r.PLATFORM_STEREO_VR_4K_60FPS_12M = 811,
// r.PLATFORM_STEREO_VR_2P5K_60FPS = 812,

type responseSource struct {
	URL string `json:"u"`
	P   int    `json:"p"`
}

type responseOs struct {
	I     int         `json:"i"`
	L     string      `json:"l"`
	URI   string      `json:"src"`
	SType interface{} `json:"stype"`
	VType interface{} `json:"vtype"`
}
type vp struct {
	Yit int `json:"yit"`
}
type responseVideo struct {
	ID           int              `json:"i"`
	Version      int              `json:"version"`
	Name         string           `json:"n"`
	Params       string           `json:"p"`
	QualityType  int              `json:"qt"`
	Time         int              `json:"time"`
	Sources      []responseSource `json:"src"`
	ThumbnailURL string           `json:"turl"`
	UV           *interface{}     `json:"uv"`
	VP           []vp             `json:"vp"`
	SA           int64            `json:"sa"`
	OS           []responseOs     `json:"os"`
	_params      *params
}

func (v *responseVideo) getSource(p int) string {
	for _, s := range v.Sources {
		if s.P == p {
			return s.URL
		}
	}
	return ""
}

func (o responseOs) getSType() interface{} {
	return o.SType
}
func (o responseOs) getVType() interface{} {
	return o.VType
}
func (o responseOs) getPlaylist() string {
	uri := o.URI
	switch getVRSetting(o) {
	case skybox.VrSettingTopBottom360, skybox.VrSettingLeftRight180:
		uri = regexp.MustCompile(`/pid/7[0-3]`).ReplaceAllString(uri, "/pid/78")
		uri = regexp.MustCompile(`_7[0-3](\d+)\.m3u8`).ReplaceAllString(uri, "_78$1.m3u8")
	case skybox.VrSetting2D360:
		uri = regexp.MustCompile(`/pid/7[0-3]`).ReplaceAllString(uri, "/pid/70")
		uri = regexp.MustCompile(`_7[0-3](\d+)\.m3u8`).ReplaceAllString(uri, "_70$1.m3u8")
	}

	return uri
}

type params struct {
	SType interface{} `json:"stype"`
	VType interface{} `json:"vtype"`
}

func (v responseVideo) getSType() interface{} {
	return v.getParams().SType
}

func (v responseVideo) getVType() interface{} {
	return v.getParams().VType
}

func (v *responseVideo) getParams() params {
	if v._params != nil {
		return *v._params
	}
	var data params
	if v.Params != "" {
		if err := json.Unmarshal([]byte(v.Params), &data); err != nil {
			log.Println(err, v.Params)
		}
	}
	v._params = &data
	return data
}

type getVRSettingOption interface {
	getSType() interface{}
	getVType() interface{}
}

func getVRSetting(o getVRSettingOption) int {
	switch o.getSType() {
	case 1, "1":
		return skybox.VrSettingTopBottom360
	case 2, "2":
		return skybox.VrSettingLeftRight180
	}

	switch o.getVType() {
	case 1, "1":
		return skybox.VrSetting2DScreen
	}

	return skybox.VrSetting2D360
}

func (v *responseVideo) enabled() bool {
	if 0 < len(v.Sources) {
		if strings.Contains(v.Sources[0].URL, "/res/p/") {
			if v.UV == nil {
				if 0 < len(v.VP) && v.VP[0].Yit == 0 {
					return true
				}
				return false
			}
			return true
		}
		return true

	}
	return false
}

func (v *responseVideo) getPlaylist() string {
	var uri string
	switch getVRSetting(v) {
	case skybox.VrSettingTopBottom360, skybox.VrSettingLeftRight180:
		if 0 < v.QualityType {
			if uri = v.getSource(810); uri != "" {
				return uri
			}
			if uri = v.getSource(812); uri != "" {
				uri = strings.ReplaceAll(uri, "/pid/812/", "/pid/810/")
				uri = strings.ReplaceAll(uri, "_812.m3u8", "_810.m3u8")
				return uri
			}
		}
		return v.getSource(601)
	case skybox.VrSetting2DScreen:
		return v.getSource(401)
	default:
		if 0 < v.QualityType {
			if uri = v.getSource(800); uri != "" {
				return uri
			}
			if uri = v.getSource(802); uri != "" {
				uri = strings.ReplaceAll(uri, "/pid/802/", "/pid/800/")
				uri = strings.ReplaceAll(uri, "_802.m3u8", "_800.m3u8")
				return uri
			}
		}
		return v.getSource(102)
	}
}

type mediaInfo struct {
	name      string
	uri       string
	vrSetting int
}

func (v *responseVideo) getInfo() []mediaInfo {
	var info []mediaInfo
	if 0 < len(v.OS) {
		for _, o := range v.OS {
			info = append(info, mediaInfo{
				name:      fmt.Sprintf("%s - %s", v.Name, o.L),
				uri:       o.getPlaylist(),
				vrSetting: getVRSetting(o),
			})
		}
	} else {
		info = append(info, mediaInfo{
			name:      v.Name,
			uri:       v.getPlaylist(),
			vrSetting: getVRSetting(v),
		})
	}

	return info
}

type responseResult struct {
	Videos []responseVideo `json:"videos"`
	Paging paging          `json:"paging"`
}
type paging struct {
	P  int `json:"p"`
	MP int `json:"mp"`
	C  int `json:"c"`
}
type response struct {
	Result responseResult `json:"result"`
}
