package pathmatch

import (
	"os"
	"strings"
)

func PathMatch(path, pattern string) bool {
	sep := string(os.PathSeparator)

	if !strings.HasPrefix(path, sep) {
		path = sep + path
	}

	if !strings.HasSuffix(path, sep) {
		path = path + sep
	}

	if !strings.HasPrefix(pattern, sep) {
		pattern = sep + pattern
	}

	if !strings.HasSuffix(pattern, sep) {
		pattern = pattern + sep
	}

	return strings.Contains(path, pattern)
}

func PathMatchAny(path string, patterns ...string) bool {
	for _, pattern := range patterns {
		if PathMatch(path, pattern) {
			return true
		}
	}
	return false
}
