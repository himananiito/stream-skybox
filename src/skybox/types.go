package skybox

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	VrSetting2D180 = iota
	VrSetting2D360
	VrSetting2DScreen
	VrSettingLeftRight180
	VrSettingLeftRight360
	VrSettingLeftRightScreen
	VrSettingTopBottom180
	VrSettingTopBottom360
	VrSettingTopBottomScreen
)

// fixme if needed
func generateActionID() string {
	timestamp := time.Now().UnixNano()
	bs := sha256.Sum256([]byte(strconv.FormatInt(timestamp, 10)))
	return hex.EncodeToString(bs[:32])
}

// interface to string
func i2s(data interface{}) string {
	switch d := data.(type) {
	case string:
		return d
	default:
		return fmt.Sprintf("%v", d)
	}
}

// interface to float64
func i2f(data interface{}) float64 {
	switch d := data.(type) {
	case float64:
		return d
	case string:
		f, _ := strconv.ParseFloat(d, 64)
		return f
	default:
		f, _ := strconv.ParseFloat(fmt.Sprintf("%v", d), 64)
		return f
	}
}

type AddDeviceResult struct {
	Command    string `json:"command"`    // addDeviceResult
	Success    bool   `json:"success"`    // true
	Version    string `json:"version"`    // 10
	Os         string `json:"os"`         // win
	IsLoggedIn bool   `json:"isLoggedIn"` // true
}

func NewAddDeviceResult() AddDeviceResult {
	return AddDeviceResult{
		Command:    "addDeviceResult",
		Success:    true,
		Version:    "10",
		Os:         "win",
		IsLoggedIn: true,
	}
}

type Media struct {
	ID                   string        `json:"id"`                   // "\p{Hex}{32}"
	Name                 string        `json:"name"`                 // "videoname"
	Duration             int64         `json:"duration"`             // seconds * 1000
	Size                 int64         `json:"size"`                 // 938092625
	URL                  string        `json:"url"`                  // "http://192.168.100.12:6890/stream/\p{Hex}{32}"
	Thumbnail            string        `json:"thumbnail"`            // "http://192.168.100.12:x/thumbnail/\p{Hex}{32}"
	ThumbnailWidth       int           `json:"thumbnailWidth"`       // 186
	ThumbnailHeight      int           `json:"thumbnailHeight"`      // 120
	LastModified         int64         `json:"lastModified"`         // 1579442054395
	DefaultVRSetting     int           `json:"defaultVRSetting"`     // 2
	UserVRSetting        int           `json:"userVRSetting"`        // 3
	Width                int           `json:"width"`                // 3840
	Height               int           `json:"height"`               // 1920
	OrientDegree         string        `json:"orientDegree"`         // "0"
	Subtitles            []interface{} `json:"subtitles" gorm:"-"`   // []
	RatioTypeFor2DScreen string        `json:"ratioTypeFor2DScreen"` // default
	RotationFor2DScreen  int           `json:"rotationFor2DScreen"`  // 0
	Exists               bool          `json:"exists"`               // true
	IsBadMedia           bool          `json:"isBadMedia" gorm:"-"`  // false
	AddedTime            int64         `json:"addedTime"`            // 1582462306906
	IsSample             bool          `json:"-"`                    // permanently sample video
}

type ActiveSetTime struct {
	Command  string  `json:"command"`
	ActionID string  `json:"actionId"`
	Time     float64 `json:"time"`
	DeviceID string  `json:"deviceId"`
}

func newActiveSetTime(time float64, actionID, deviceID string) ActiveSetTime {
	return ActiveSetTime{
		Command:  "activeSetTime",
		Time:     time,
		ActionID: actionID,
		DeviceID: deviceID,
	}
}

type UpdatePlayerState struct {
	Command        string `json:"command"`
	LatestActionID string `json:"latestActionId"`
	State          string `json:"state"` // "stopped"
	// if not stopped { // _playlist.currentMedia
	MediaID         interface{} `json:"mediaId,omitempty"`
	Time            interface{} `json:"time,omitempty"`
	Name            interface{} `json:"name,omitempty"`
	Duration        interface{} `json:"duration,omitempty"`
	Size            interface{} `json:"size,omitempty"`
	IsBadMedia      interface{} `json:"isBadMedia,omitempty"`
	ThumbnailWidth  interface{} `json:"thumbnailWidth,omitempty"`
	ThumbnailHeight interface{} `json:"thumbnailHeight,omitempty"`
	LastModified    interface{} `json:"lastModified,omitempty"`
	VRSetting       interface{} `json:"VRSetting,omitempty"`
	Width           interface{} `json:"width,omitempty"`
	Height          interface{} `json:"height,omitempty"`
	// }
	ShowMirrorScreen bool `json:"showMirrorScreen"`
	// settings
	Speed          string        `json:"speed"`
	AbLoopPointA   float64       `json:"abLoopPointA"`
	AbLoopPointB   float64       `json:"abLoopPointB"`
	RandomMode     string        `json:"randomMode"`
	LoopMode       string        `json:"loopMode"`
	RandomPlaylist []interface{} `json:"randomPlaylist"`
}

func newStoppedUpdatePlayerState() UpdatePlayerState {
	return UpdatePlayerState{
		Command:        "updatePlayerState",
		LatestActionID: generateActionID(), // [FIXME]
		State:          "stopped",
		// [FIXME] setting
		Speed:          "1",
		AbLoopPointA:   -1,
		AbLoopPointB:   -1,
		RandomMode:     "order",
		LoopMode:       "playlist",
		RandomPlaylist: []interface{}{},
	}
}

type ActiveSetVRSetting struct {
	Command     string  `json:"command"`
	SettingCode float64 `json:"settingCode"`
	DeviceID    string  `json:"deviceId"`
}

func newActiveSetVRSetting(settingCode float64, deviceID string) ActiveSetVRSetting {
	return ActiveSetVRSetting{
		Command:     "activeSetVRSetting",
		SettingCode: settingCode,
		DeviceID:    deviceID,
	}
}

type ActiveDisconnect struct {
	Command      string `json:"command"`
	WindowClosed bool   `json:"windowClosed"`
}

func newActiveDisconnect(windowClosed bool) ActiveDisconnect {
	return ActiveDisconnect{
		Command:      "activeDisconnect",
		WindowClosed: windowClosed,
	}
}

type UpdatePlayerAbLoop struct {
	Command  string  `json:"command"`
	PointA   float64 `json:"pointA"`
	PointB   float64 `json:"pointB"`
	DeviceID string  `json:"deviceId"`
}

func newUpdatePlayerAbLoop(pointA, pointB float64, deviceID string) UpdatePlayerAbLoop {
	return UpdatePlayerAbLoop{
		Command:  "updatePlayerAbLoop",
		PointA:   pointA,
		PointB:   pointB,
		DeviceID: deviceID,
	}
}

type UpdatePlayerSpeed struct {
	Command  string `json:"command"`
	Speed    string `json:"speed"`
	DeviceID string `json:"deviceId"`
}

func newUpdatePlayerSpeed(speed string, deviceID string) UpdatePlayerSpeed {
	return UpdatePlayerSpeed{
		Command:  "updatePlayerSpeed",
		Speed:    speed,
		DeviceID: deviceID,
	}
}

type activeCommon struct {
	Command  string `json:"command"`
	DeviceID string `json:"deviceId"`
	ActionID string `json:"actionId"`
}

func NewActiveCommon(command, deviceID string) activeCommon {
	return activeCommon{
		Command:  fmt.Sprintf("active%s", strings.Title(strings.ToLower(command))),
		DeviceID: deviceID,
		ActionID: generateActionID(),
	}
}

type updatePlayerRandomAndLoopMode struct {
	Command    string `json:"command"`
	RandomMode string `json:"randomMode"`
	LoopMode   string `json:"loopMode"`
	DeviceID   string `json:"deviceId"`
}

func newUpdatePlayerRandomAndLoopMode(randomMode, loopMode, deviceID string) updatePlayerRandomAndLoopMode {
	return updatePlayerRandomAndLoopMode{
		Command:    "updatePlayerRandomAndLoopMode",
		RandomMode: randomMode,
		LoopMode:   loopMode,
		DeviceID:   deviceID,
	}
}

type deleteAllMedias struct {
	Command string `json:"command"` // deleteAllMedias
}

func newDeleteAllMedias() deleteAllMedias {
	return deleteAllMedias{
		Command: "deleteAllMedias",
	}
}

type updatePlaylist struct {
	Command string   `json:"command"` // updatePlaylist
	List    []string `json:"list"`    // [FIXME] []
}

func newUpdatePlaylist(list []string) updatePlaylist {
	return updatePlaylist{
		Command: "updatePlaylist",
		List:    list,
	}
}

type updateReadyMediaToClients struct {
	Command string `json:"command"` // "updateReadyMediaToClients"
	Media   Media  `json:"media"`   //
}

type getMediaListResult struct {
	Command string  `json:"command"` // getMediaListResult
	List    []Media `json:"list"`
}

func newGetMediaListResult(mediaList []Media) getMediaListResult {
	return getMediaListResult{
		Command: "getMediaListResult",
		List:    mediaList,
	}
}

// [FIXME]
type activePlay struct {
	Command              string        `json:"command"` // "activePlay"
	ActionID             string        `json:"actionId"`
	DeviceID             string        `json:"deviceId"`
	ID                   string        `json:"id"`                   //
	Exists               bool          `json:"exists"`               // true
	Name                 string        `json:"name"`                 // ""
	Size                 int64         `json:"size"`                 //
	Duration             int64         `json:"duration"`             //
	StreamType           string        `json:"streamType"`           //"RTSP"
	DefaultVRSetting     int           `json:"defaultVRSetting"`     // 2
	UserVRSetting        int           `json:"userVRSetting"`        // 3
	StreamURL            string        `json:"streamUrl"`            // "rtsp://192.168.100.12:8554/test"
	PlayTime             int           `json:"playTime"`             // 0
	Width                int           `json:"width"`                // 3840
	Height               int           `json:"height"`               // 1920
	OrientDegree         string        `json:"orientDegree"`         // "0"
	RatioTypeFor2DScreen string        `json:"ratioTypeFor2DScreen"` // "default"
	RotationFor2DScreen  int           `json:"rotationFor2DScreen"`  // 0
	Speed                string        `json:"speed"`                // "1"
	AbLoopPointA         int           `json:"abLoopPointA"`         // -1
	AbLoopPointB         int           `json:"abLoopPointB"`         // -1
	RandomMode           string        `json:"randomMode"`           // "order"
	LoopMode             string        `json:"loopMode"`             // "playlist"
	RandomPlaylist       []interface{} `json:"randomPlaylist"`       // [FIXME] []
	URL                  string        `json:"url"`                  //
}

func newActivePlay(id, deviceID string) activePlay {
	return activePlay{
		Command:              "activePlay",
		DeviceID:             deviceID,
		ID:                   id,
		Exists:               true,
		Name:                 "test",
		Size:                 938092625,
		Duration:             1909000,
		StreamType:           "RTSP",
		DefaultVRSetting:     3,
		UserVRSetting:        3,
		PlayTime:             0,
		Width:                3840,
		Height:               1920,
		OrientDegree:         "180",
		RatioTypeFor2DScreen: "default",
		RotationFor2DScreen:  0,
		Speed:                "1",
		AbLoopPointA:         -1,
		AbLoopPointB:         -1,
		RandomMode:           "order",
		LoopMode:             "playlist",
		RandomPlaylist:       []interface{}{},
		URL:                  "http://192.168.100.9:6888/stream/playlist.m3u8",
	}
}
