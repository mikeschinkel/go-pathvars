package test

import (
	"strings"
)

// parsePathSpec parses a path specification in the format "METHOD /path" or "/path"
// If no space is found, it returns empty method and the full string as path
func parsePathSpec(pathSpec string) (method, path string) {
	if before, after, found := strings.Cut(pathSpec, " "); found {
		return before, after
	}
	return "", pathSpec
}
