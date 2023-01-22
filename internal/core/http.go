package core

import (
	"net"
	"net/http"
)

func GetRemoteIP(req *http.Request) string {
	if ip := req.Header.Get("CF-Connecting-IP"); len(ip) != 0 {
		return ip
	}
	if ip := req.Header.Get("X-Forwarded-For"); len(ip) != 0 {
		return ip
	}
	if ip := req.Header.Get("X-Real-IP"); len(ip) != 0 {
		return ip
	}
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}

func GetUserAgent(req *http.Request) string {
	return req.Header.Get("User-Agent")
}
