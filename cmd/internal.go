package cmd

import (
	"errors"
	"os"
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
	default:
		return tagPrefix + "-"
	}
}

var ErrUndefinedSSHAuthSock = errors.New("SSH_AUTH_SOCK is not defined")

func checkAuthSocket(pushTag bool) (string, error) {
	socket, found := os.LookupEnv("SSH_AUTH_SOCK")
	if pushTag && !found {
		return "", ErrUndefinedSSHAuthSock
	}
	return socket, nil
}
