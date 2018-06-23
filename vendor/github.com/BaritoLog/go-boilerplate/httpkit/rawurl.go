package httpkit

import (
	"net/url"
	"strconv"
	"strings"
)

type RawUrl string

// PathParameter
func PathParameterOfRawURL(rawUrl, key string) string {
	i := strings.Index(rawUrl, key+"/")
	if i < 0 {
		return ""
	}
	s := rawUrl[i+len(key)+1:]
	j := strings.Index(s, "/")
	if j < 0 {
		return s
	}

	return s[:j]
}

func HostOfRawURL(rawUrl string) (host string, port int) {
	u, _ := url.Parse(rawUrl)

	host = u.Host
	i := strings.Index(host, ":")
	if i >= 0 {
		host = host[0:i]
	}

	port, _ = strconv.Atoi(u.Port())
	return
}
