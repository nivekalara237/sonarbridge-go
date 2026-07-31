package utils

import (
	"net/url"
	"strings"
)

func EncodeURIComponent(str string) string {
	return strings.ReplaceAll(url.QueryEscape(str), "+", "%20")
}
