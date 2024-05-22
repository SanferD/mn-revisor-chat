package helpers

import (
	"encoding/base64"
	"strings"
)

func IsLocalhostURL(url string) bool {
	return strings.HasPrefix(url, "localhost") || strings.HasPrefix(url, "http://localhost") || strings.HasPrefix(url, "https://localhost")

}

func Base64Encode(content string) string {
	return base64.StdEncoding.EncodeToString([]byte(content))
}
