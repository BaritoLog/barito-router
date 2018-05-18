package httpkit

import (
	"net/url"
	"strconv"
	"strings"
)

// PathParameter
func PathParameter(path, key string) string {
	i := strings.Index(path, key+"/")
	if i < 0 {
		return ""
	}
	s := path[i+len(key)+1:]
	j := strings.Index(s, "/")
	if j < 0 {
		return s
	} else {
		return s[:j]
	}
}

func Host(rawurl string) (host string, port int) {
	u, _ := url.Parse(rawurl)

	host = u.Host
	i := strings.Index(host, ":")
	if i >= 0 {
		host = host[0:i]
	}

	port, _ = strconv.Atoi(u.Port())

	return

}
