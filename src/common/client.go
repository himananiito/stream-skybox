package common

import (
	"net"
	"net/http"
	"net/url"
	"os"
)

func GetClient() *http.Client {
	client := &http.Client{}

	tr := &http.Transport{}
	// debug purpose
	proxy := os.Getenv("STREAM_SKYBOX_PROXY")
	if proxy != "" {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(proxy)
		}
	}
	if os.Getenv("STREAM_SKYBOX_IPV4") == "1" {
		tr.Dial = func(network, addr string) (net.Conn, error) {
			dstAddr, err := net.ResolveTCPAddr("tcp4", addr)
			if err != nil {
				return nil, err
			}
			return net.DialTCP("tcp4", nil, dstAddr)
		}
	}
	client.Transport = tr
	return client
}
