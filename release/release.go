package release

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"regexp"
)

var (
	ErrSkipRelease = errors.New("skip release hint present")
)

func skipReleaseHintPresent(aString, tagPrefix string) bool {
	if tagPrefix == "" {
		return regexp.MustCompile("vergo:skip-release").MatchString(aString)
	}
	return regexp.MustCompile("vergo:" + tagPrefix + ":skip-release").MatchString(aString)
}

func CheckRelease(repo *git.Repository, tagPrefixRaw string) error {
	head, err := repo.Head()
	if err != nil {
		return err
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return err
	}
	if skipReleaseHintPresent(commit.Message, tagPrefixRaw) {
		return fmt.Errorf("%w: %s", ErrSkipRelease, tagPrefixRaw)
	}
	return nil
}
