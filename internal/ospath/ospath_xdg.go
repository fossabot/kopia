// +build !windows,!darwin

package ospath

import (
	"os"
	"path/filepath"
)

func init() {
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		userSettingsDir = os.Getenv("XDG_CONFIG_HOME")
	} else {
		userSettingsDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	if os.Getenv("XDG_CACHE_HOME") != "" {
		userCacheDir = os.Getenv("XDG_CACHE_HOME")
	} else {
		userCacheDir = filepath.Join(os.Getenv("HOME"), ".cache")
	}
}
