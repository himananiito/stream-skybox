package skybox

import (
	"encoding/json"

	"github.com/gofiber/websocket/v2"
)

const (
	// https://github.com/socketio/socket.io-protocol
	SioTypeConnect      = '0'
	SioTypeDisconnect   = '1'
	SioTypeEvent        = '2'
	SioTypeAck          = '3'
	SioTypeConnectError = '4'
	SioTypeBinaryEvent  = '5'
	SioTypeBinaryAck    = '6'
)

func SioConnect(c *websocket.Conn) error {
	if err := c.WriteMessage(websocket.TextMessage, []byte{EioTypeMessage, SioTypeConnect}); err != nil {
		return err
	}
	return nil
}

func SioServerMessage(c *websocket.Conn, data interface{}) error {
	return SioEvent(c, []interface{}{"serverMessage", data})
}

func SioEvent(c *websocket.Conn, data interface{}) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}

	//log.Println(string(bs))

	msg := append([]byte{EioTypeMessage, SioTypeEvent}, bs...)

	if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
		return err
	}
	return nil
}
