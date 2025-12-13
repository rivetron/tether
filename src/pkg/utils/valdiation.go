package utils

import (
	"net/url"
	"strings"
)

func IsValidURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	// * Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// ? Check if schema and host are present
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	// ? Check if http or https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	// ? Check for host
	if strings.Contains(parsedURL.Host, " ") {
		return false
	}

	return true
}

func NormalizeURL(urlStr string) string {
	if urlStr == "" {
		return ""
	}

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	return parsedURL.String()
}

func IsBlockedDomain(urlStr string, blockedDomains []string) bool {
	if len(blockedDomains) == 0 {
		return false
	}

	// * Block invalid URLs
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return true
	}

	host := strings.ToLower(parsedURL.Host)

	// ? Check and remove port if present
	if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	for _, blocked := range blockedDomains {
		blocked = strings.ToLower(blocked)

		// * Exact match
		if host == blocked {
			return true
		}

		// * Subdomain match
		if strings.HasPrefix(blocked, "*.") {
			domain := blocked[2:]
			if host == domain || strings.HasPrefix(host, "."+domain) {
				return true
			}
		}

		// ? Suffix match for domain without wildcard
		if strings.HasSuffix(host, "."+blocked) {
			return true
		}
	}

	return false
}

func SanitizeCustomCode(code string) string {
	if code == "" {
		return ""
	}

	// * Trim whitespace
	code = strings.TrimSpace(code)

	// * Convert to lowercase for consistency
	code = strings.ToLower(code)

	return code
}
