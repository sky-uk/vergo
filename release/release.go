package release

import (
	"errors"
	"fmt"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"regexp"
	"strings"
)

var (
	ErrSkipRelease     = errors.New("skip release hint present")
	ErrInvalidHeadless = errors.New("invalid headless checkout")
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

func ValidateHEAD(repo *gogit.Repository, remoteName string, versionedBranches []string) error {
	head, err := repo.Head()
	if err != nil {
		return err
	}
	log.Debugf("Current branch:%v, short: %v", head.Name(), head.Name().Short())
	if head.Name() == plumbing.HEAD {
		validRef := false
		for _, mainBranchName := range versionedBranches {
			remote, err := repo.Remote(remoteName)
			if err != nil && err != gogit.ErrRemoteNotFound {
				return err
			}
			var branchRef plumbing.ReferenceName
			if remote != nil {
				branchRef = plumbing.NewRemoteReferenceName(remoteName, mainBranchName)
			} else {
				branchRef = plumbing.NewBranchReferenceName(mainBranchName)
			}
			revision, err := repo.ResolveRevision(plumbing.Revision(branchRef))
			if err != nil {
				log.WithError(err).Debugf("branchRef could not be resolved: %s\n", branchRef.String())
			} else {
				if revision.String() == head.Hash().String() {
					validRef = true
					break
				} else {
					log.Tracef("Invalid ref [branch: %s, head: %s, ref: %s]\n",
						branchRef.String(), head.Hash().String(), revision.String())
				}
			}
		}
		if !validRef {
			return ErrInvalidHeadless
		}
	} else if !funk.ContainsString(versionedBranches, head.Name().Short()) {
		return fmt.Errorf("branch %s is not in main branches list: %s", head.Name().Short(), strings.Join(versionedBranches, ", "))
	}
	return nil
}
