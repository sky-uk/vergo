package release

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"regexp"
	"strings"
)

var (
	ErrNoIncrement = errors.New("increment hint not present")
	ErrSkipRelease = errors.New("skip release hint present")
)

func checkSkipHint(aString, tagPrefix string) bool {
	if tagPrefix == "" {
		return regexp.MustCompile("vergo:skip-release").MatchString(aString)
	}
	return regexp.MustCompile("vergo:" + tagPrefix + ":skip-release").MatchString(aString)
}

type SkipHintPresentFunc func(repo *gogit.Repository, tagPrefixRaw string) error

func SkipHintPresent(repo *gogit.Repository, tagPrefixRaw string) error {
	head, err := repo.Head()
	switch {
	case errors.Is(err, plumbing.ErrReferenceNotFound):
		return nil
	case err != nil:
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

func checkIncrementHint(aString, tagPrefixRaw string) (string, error) {
	var re *regexp.Regexp
	if tagPrefixRaw == "" {
		re = regexp.MustCompile("vergo:(major|minor|patch|pre)-release")
	} else {
		re = regexp.MustCompile("vergo:" + tagPrefixRaw + ":(major|minor|patch|pre)-release")
	}
	match := re.FindStringSubmatch(aString)
	if len(match) != 2 {
		return "", fmt.Errorf("%w: %s", ErrNoIncrement, tagPrefixRaw)
	}
	if match[1] == "pre" {
		return "prerelease", nil
	}
	return match[1], nil
}

type IncrementHintFunc func(repo *gogit.Repository, tagPrefixRaw string) (string, error)

func IncrementHint(repo *gogit.Repository, tagPrefixRaw string) (string, error) {
	head, err := repo.Head()
	switch {
	case errors.Is(err, plumbing.ErrReferenceNotFound):
		return "minor", nil
	case err != nil:
		return "", err
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return "", err
	}
	return checkIncrementHint(commit.Message, tagPrefixRaw)
}

type ValidateHEADFunc func(repo *gogit.Repository, remoteName string, versionedBranches []string) error

func ValidateHEAD(repo *gogit.Repository, remoteName string, versionedBranches []string) error {
	head, err := repo.Head()
	if err != nil {
		return err
	}
	log.Debugf("Current branch:%v, short: %v", head.Name(), head.Name().Short())
	isHeadlessCheckout := head.Name() == plumbing.HEAD
	if isHeadlessCheckout {
		validRef := false
		for _, mainBranchName := range versionedBranches {
			remote, err := repo.Remote(remoteName)
			if err != nil && !errors.Is(err, gogit.ErrRemoteNotFound) {
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
				commitOnVersionedBranch, err := isCommitOnBranch(repo, head.Hash(), branchRef)
				if err != nil {
					log.WithError(err).Errorf("Failed to check if commit %s is on branch %s\n",
						head.Hash().String(), branchRef.String())
				}

				if commitOnVersionedBranch {
					validRef = true
					break
				} else {
					log.Warnf("Commit not found on branch [branch: %s, head: %s, ref: %s]\n",
						branchRef.String(), head.Hash().String(), revision.String())
				}
			}
		}
		if !validRef {
			return fmt.Errorf("commit %s is not on a versioned branch: %s",
				head.Hash(), strings.Join(versionedBranches, ", "))
		}
	} else if !funk.ContainsString(versionedBranches, head.Name().Short()) {
		return fmt.Errorf("branch %s is not in versioned branches list: %s",
			head.Name().Short(), strings.Join(versionedBranches, ", "))
	}
	return nil
}

type PreReleaseFunc func(version *semver.Version) (semver.Version, error)
type PreReleaseOptions struct {
	WithMetadata bool
}

func PreRelease(repo *gogit.Repository, options PreReleaseOptions) PreReleaseFunc {
	return func(version *semver.Version) (semver.Version, error) {
		pre, err := version.IncMinor().SetPrerelease("SNAPSHOT")
		if err != nil {
			return semver.Version{}, err
		}
		if options.WithMetadata {
			head, err := repo.Head()
			if err != nil {
				return semver.Version{}, err
			}
			return pre.SetMetadata(head.Hash().String()[0:7])
		}
		return pre, nil
	}
}

func isCommitOnBranch(repo *gogit.Repository, commit plumbing.Hash, branch plumbing.ReferenceName) (bool, error) {
	branchRef, err := repo.Reference(branch, true)
	if err != nil {
		return false, err
	}

	branchGitLog, err := repo.Log(&gogit.LogOptions{
		From: branchRef.Hash(),
	})

	if err != nil {
		return false, err
	}

	reaches := false
	err = branchGitLog.ForEach(func(branchCommit *object.Commit) error {
		if branchCommit.Hash == commit {
			reaches = true
			return storer.ErrStop
		}
		return nil
	})

	if err != nil {
		return false, err
	}

	return reaches, nil
}
