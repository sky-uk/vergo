package cmd

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"os"
	"regexp"
	"strings"
)

var (
	ErrInvalidArg        = errors.New("invalid arg")
	ErrMessageIgnoreHint = errors.New("ignore hint")
)

type IgnoreRelease func(repo *git.Repository, tagPrefixRaw string) error

func IgnoreHintPresent(string, tagPrefix string) bool {
	re := regexp.MustCompile("vergo.?" + tagPrefix + ".?ignore")
	return re.MatchString(string)
}

func HeadCommitMessageIgnoreHintPresent(repo *git.Repository, tagPrefixRaw string) error {
	head, err := repo.Head()
	if err != nil {
		return err
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return err
	}
	if IgnoreHintPresent(commit.Message, tagPrefixRaw) {
		return fmt.Errorf("%w : %s", ErrMessageIgnoreHint, commit.Message)
	}
	return nil
}

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
