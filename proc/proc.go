package proc

import (
	"os"
)

const unknown = "unknown"

// Hostname returns current hostname
func Hostname() string {
	if h, _ := os.Hostname(); h != "" {
		return h
	}

	return unknown
}
