package skybox

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type Server struct {
	Library *Library
	guard   bool
	mtx     sync.RWMutex
	localIP string
	wsPort  int
}

func NewServer(wsPort int) *Server {
	lib := NewLibrary()

	srv := &Server{
		Library: lib,
		wsPort:  wsPort,
	}

	go srv.serveUDP()

	return srv
}

func (srv *Server) setLocalIP(ip string) {
	srv.localIP = ip
}
func (srv *Server) getLocalIP() string {
	return srv.localIP
}
func (srv *Server) GetHost() string {
	if srv.wsPort == 0 {
		return ""
	}
	ip := srv.getLocalIP()
	if ip == "" {
		return ""
	}
	return fmt.Sprintf("%v:%v", ip, srv.wsPort)
}

func (srv *Server) visible() bool {
	return !srv.GetGuard() && srv.Library.hasMedia()
}

func (srv *Server) SetGuard(b bool) {
	srv.mtx.Lock()
	defer srv.mtx.Unlock()
	srv.guard = b
}
func (srv *Server) GetGuard() bool {
	srv.mtx.RLock()
	defer srv.mtx.RUnlock()
	return srv.guard
}

func (srv *Server) GetMedias() []Media {
	medias := srv.Library.GetMedias()
	localIP := srv.getLocalIP()
	for i := 0; i < len(medias); i++ {
		medias[i].Subtitles = []interface{}{}
		if strings.HasPrefix(medias[i].URL, "/") {
			medias[i].URL = fmt.Sprintf("http://%v:%v", localIP, srv.wsPort) + medias[i].URL
		}
	}
	return medias
}

func (srv *Server) Callback(c *websocket.Conn) {

	log.Println("enter ws callback")
	defer log.Println("exit ws callback")

	if err := EioOpen(c); err != nil {
		log.Println(err)
		return
	}
	if err := SioConnect(c); err != nil {
		log.Println(err)
		return
	}

	for {
		if mt, msg, err := c.ReadMessage(); err != nil {
			log.Println("read:", err)
			break
		} else {
			log.Println("read:", mt, string(msg))

			switch mt {
			case 1:
				eioType := msg[0]
				msg = msg[1:]
				switch eioType {
				case EioTypeMessage:
					sioType := msg[0]
					msg := msg[1:]
					switch sioType {
					case SioTypeEvent:
						//log.Println("got sio event", string(msg))

						var sioData []string
						err = json.Unmarshal(msg, &sioData)
						if err != nil {
							panic(err)
						}
						var msg map[string]interface{}
						err = json.Unmarshal([]byte(sioData[1]), &msg)
						if err != nil {
							panic(err)
						}

						// common
						cmd, _ := msg["command"].(string)
						deviceID, _ := msg["deviceId"].(string)
						id, _ := msg["id"].(string)

						/*
							if cmd == "upgradePriority" {
								break
							}
						*/

						switch cmd {
						case "addDevice":
							SioServerMessage(c, NewAddDeviceResult())

						case "getMediaList", "refreshMediaList":
							// TODO
							mediaList := srv.GetMedias()
							// sendDeleteAllMedias
							SioServerMessage(c, newGetMediaListResult(mediaList))

						case "play":
							// FIXME
							SioServerMessage(c, newActivePlay(id, deviceID))

						case "pause", "resume", "stop":
							SioServerMessage(c, NewActiveCommon(cmd, deviceID))

						case "getPlaylist":

							// FIXME
							list := []string{}
							SioServerMessage(c, newUpdatePlaylist(list))

						case "setPlayerSpeed":
							SioServerMessage(c, newUpdatePlayerSpeed(
								i2s(msg["speed"]),
								i2s(msg["deviceId"]),
							))

						case "setPlayerAbLoop":
							SioServerMessage(c, newUpdatePlayerAbLoop(
								i2f(msg["pointA"]),
								i2f(msg["pointB"]),
								i2s(msg["deviceId"]),
							))

						case "setPlayerRandomAndLoopMode":
							SioServerMessage(c, newUpdatePlayerRandomAndLoopMode(
								i2s(msg["randomMode"]),
								i2s(msg["loopMode"]),
								i2s(msg["deviceId"]),
							))

						case "disconnect":
							SioServerMessage(c, newActiveDisconnect(true))

						case "setVRSetting":
							SioServerMessage(c, newActiveSetVRSetting(
								i2f(msg["settingCode"]),
								i2s(msg["deviceId"]),
							))

						case "setTime":
							SioServerMessage(c, newActiveSetTime(
								i2f(msg["time"]),
								generateActionID(),
								i2s(msg["deviceId"]),
							))

						case "getPlayerState":
							// [FIXME]
							SioServerMessage(c, newStoppedUpdatePlayerState())

						default:
							log.Printf("\n\n[FIXME] %q\n\n\n", msg)

						}

					default:
						log.Fatalln("got", string(msg))
					}
				case EioTypePing:
					EioPong(c)
				case EioTypeClose:
					c.Close()
					return
				default:
					log.Fatalln("[FIXME] eio type", eioType)
				}
			default:
				log.Fatalln("[FIXME] websocket message type", mt)
			}
		}
	}
	return
}
