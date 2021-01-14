package skybox

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
)

type udpClient struct {
	Command    string `json:"command"`    // "search"
	Project    string `json:"project"`    // "direwolf"
	DeviceID   string `json:"deviceId"`   //
	DeviceType string `json:"deviceType"` // "vr"
	UDPPort    string `json:"udpPort"`    // "6881"
}

type udpServer struct {
	UDP          bool     `json:"udp"`          // true
	Project      string   `json:"project"`      // "direwolf server"
	Command      string   `json:"command"`      // "searchResult"
	DeviceID     string   `json:"deviceId"`     //
	ComputerID   string   `json:"computerId"`   //
	ComputerName string   `json:"computerName"` //
	IP           string   `json:"ip"`
	IPs          []string `json:"ips"`
	Port         int      `json:"port"`
}

var errLocalIPNotDetected = errors.New("local IP not detected")

func getLocalIP(remote *net.UDPAddr) (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
			continue
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				a := v.IP.Mask(v.Mask).String()
				b := remote.IP.Mask(v.Mask).String()
				if a == b {
					return v.IP.String(), nil
				}
			}
		}
	}

	return "", errLocalIPNotDetected
}

func calcSha256(s string) string {
	bs := sha256.Sum256([]byte(s))
	return hex.EncodeToString(bs[:32])
}

func (srv *Server) serveUDP() {
	// computer name
	hostName, err := os.Hostname()
	if err != nil {
		log.Println(err)
	}
	computerID := calcSha256(hostName)

	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: 6879,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		log.Fatalln(err)
		return
	}

	defer conn.Close()

	message := make([]byte, 1*1000*1000)
	for {
		rlen, remote, err := conn.ReadFromUDP(message[:])
		if err != nil {
			log.Println(err)
			continue
		}
		if !srv.visible() {
			continue
		}

		localIP, err := getLocalIP(remote)
		if err != nil {
			// FIXME
			log.Println(err)
			return
		}
		srv.setLocalIP(localIP)

		var req udpClient
		err = json.Unmarshal(message[:rlen], &req)
		if err != nil {
			log.Println(err)
			continue
		}

		res, err := json.Marshal(udpServer{
			UDP:          true,
			Project:      "direwolf server",
			Command:      "searchResult",
			DeviceID:     req.DeviceID,
			ComputerID:   computerID,
			ComputerName: hostName,
			IP:           localIP,
			IPs:          []string{localIP},
			Port:         srv.wsPort,
		})
		if err != nil {
			log.Println(err)
			continue
		}
		_, err = conn.WriteTo(res, remote)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
