package app

import (
	"net/url"
	"strconv"
)

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

func redirectFlash(path, msg string) string {
	return path + "?flash=" + url.QueryEscape(msg)
}
