package cmd

import (
	"errors"
	"strings"
)

var (
	ErrInvalidArg = errors.New("invalid arg")
)

func sanitiseTagPrefix(tagPrefix string) string {
	switch tagPrefix := strings.ToLower(strings.TrimSpace(tagPrefix)); {
	case tagPrefix == "":
		return "v"
	case tagPrefix == "v":
		return "v"
	case strings.HasSuffix(tagPrefix, "-"):
		return tagPrefix
	case strings.HasSuffix(tagPrefix, "/v"):
		return tagPrefix
	default:
		return tagPrefix + "-"
	}
}
