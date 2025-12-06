package utils

import (
	"net"
	"net/http"
	"strings"
)

func GetClientIP(r *http.Request) string {
	// ? Check for X-Forwarded-For header (comma-separated list, first is original client)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// ? Check X-Real-IP Header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// ? Check CF-Connecting-IP header (Cloudflare)
	if cfip := r.Header.Get("CCF-Connecting-IP"); cfip != "" {
		if net.ParseIP(cfip) != nil {
			return cfip
		}
	}

	// ? Check X-Forwarded header
	if xf := r.Header.Get("X-Forwarded"); xf != "" {
		parts := strings.SplitSeq(xf, ",")
		for part := range parts {
			part = strings.TrimSpace(part)
			if after, ok := strings.CutPrefix(part, "for="); ok {
				ip := after
				if net.ParseIP(ip) != nil {
					return ip
				}
			}
		}
	}

	// * Fallback to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
