package helpers

import (
	"encoding/base64"
)

func Base64Encode(content string) string {
	return base64.StdEncoding.EncodeToString([]byte(content))
}
