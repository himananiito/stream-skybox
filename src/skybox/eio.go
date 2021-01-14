package skybox

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/websocket/v2"
)

const (
	// https://github.com/socketio/engine.io-protocol
	EioTypeOpen    = '0'
	EioTypeClose   = '1'
	EioTypePing    = '2'
	EioTypePong    = '3'
	EioTypeMessage = '4'
	EioTypeUpgrade = '5'
	EioTypeNoop    = '6'
)

type openMessage struct {
	SID          string   `json:"sid"`
	Upgrades     []string `json:"upgrades"`
	PingInterval int      `json:"pingInterval"`
	PingTimeout  int      `json:"pingTimeout"`
}

// fixme if needed
func getSID() string {
	bs := sha512.Sum512([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	msg := base64.StdEncoding.EncodeToString(bs[:])
	return msg[:20]
}

func EioOpen(c *websocket.Conn) error {
	bs, err := json.Marshal(openMessage{
		SID:          getSID(),
		PingInterval: 25000,
		PingTimeout:  5000,
	})
	if err != nil {
		panic(err)
	}
	msg := append([]byte{EioTypeOpen}, bs...)
	if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
		return err
	}
	return nil
}

func EioPong(c *websocket.Conn) error {
	if err := c.WriteMessage(websocket.TextMessage, []byte{EioTypePong}); err != nil {
		return err
	}
	return nil
}
