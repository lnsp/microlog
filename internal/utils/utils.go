package utils

import (
	"net"
	"net/http"
)

func RemoteHost(r *http.Request) string {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}
