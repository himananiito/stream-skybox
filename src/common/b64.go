package common

import "encoding/base64"

func B64Dec(s string) string {
	bs, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return string(bs)
}

func B64Enc(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}
