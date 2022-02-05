package release

import (
	"errors"
	"fmt"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/umc-shared/vergo/git"
	"github.com/thoas/go-funk"
	"regexp"
	"strings"
)

var (
	ErrSkipRelease      = errors.New("skip release hint present")
	ErrInvalidHeadless  = errors.New("invalid headless checkout")
	ErrUnexpectedBranch = errors.New("unexpected branch")
)

func checkSkipHint(aString, tagPrefix string) bool {
	if tagPrefix == "" {
		return regexp.MustCompile("vergo:skip-release").MatchString(aString)
	}
	return regexp.MustCompile("vergo:" + tagPrefix + ":skip-release").MatchString(aString)
}

func SkipHintPresent(repo *gogit.Repository, tagPrefixRaw string) error {
	head, err := repo.Head()
	if err != nil {
		return err
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return err
	}
	if checkSkipHint(commit.Message, tagPrefixRaw) {
		return fmt.Errorf("%w: %s", ErrSkipRelease, tagPrefixRaw)
	}
	return nil
}

func ValidateHEAD(repo *gogit.Repository, versionedBranches []string) error {
	head, err := repo.Head()
	if err != nil {
		return err
	}
	log.Debugf("Current branch:%v, short: %v", head.Name(), head.Name().Short())
	if head.Name() == plumbing.HEAD {
		onMainBranch := false
		for _, branch := range versionedBranches {
			branchRef := plumbing.NewBranchReferenceName(branch)
			if exists, err := git.BranchExists(repo, branch); err == nil && !exists {
				continue
			} else if err != nil {
				return err
			}
			revision, err := repo.ResolveRevision(plumbing.Revision(branchRef))
			if err != nil {
				return err
			}
			if revision.String() == head.Hash().String() {
				onMainBranch = true
			}
		}
		if !onMainBranch {
			return ErrInvalidHeadless
		}
	} else if !funk.ContainsString(versionedBranches, head.Name().Short()) {
		return fmt.Errorf("branch %s is not in main branches list: %s", head.Name().Short(), strings.Join(versionedBranches, ", "))
	}
	return nil
}
