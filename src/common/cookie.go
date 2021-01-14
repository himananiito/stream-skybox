package common

import (
	"fmt"
	"strings"
	"time"

	"github.com/zellyn/kooky"
	_ "github.com/zellyn/kooky/allbrowsers" // register cookie store finders!
)

func GetBrowserCookie(suffix string, names ...string) string {
	var cookies []string
	for _, name := range names {
		s := func() string {
			cookies := kooky.ReadCookies(kooky.Valid, kooky.DomainHasSuffix(suffix), kooky.Name(name))
			var creation time.Time
			var c string
			for _, cookie := range cookies {
				if cookie.Creation.After(creation) {
					creation = cookie.Creation
					c = cookie.Value
				}
			}
			return c
		}()
		if s != "" {
			cookies = append(cookies, fmt.Sprintf("%s=%s", name, s))
		}
	}
	return strings.Join(cookies, "; ")
}
