package helpers

import (
	"os"
	"strings"
)

// EnforceHTTP enforces HTTP prototcol
func EnforceHTTP(url string) string {
	if url[:4] != "http" {
		return "http://" + url
	}
	return url
}

// Validates domain and return false if anything wrong is flagged 
func RemoveDomainError(url string) bool {
	// Removing url prefix like "http//:", "https//:" 
	if url == os.Getenv("DOMAIN") {
		return false
	}
	newURL := strings.Replace(url, "http://", "", 1)
	newURL = strings.Replace(newURL, "https://", "", 1)
	newURL = strings.Replace(newURL, "www.", "", 1)
	newURL = strings.Split(newURL, "/")[0]

	if newURL == os.Getenv("DOMAIN") {
		return false
	}
	return true 
}