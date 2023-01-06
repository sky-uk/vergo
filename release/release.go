package release

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"regexp"
	"strings"
)

var (
	ErrNoIncrement     = errors.New("increment hint not present")
	ErrSkipRelease     = errors.New("skip release hint present")
	ErrHEADValidation  = errors.New("HEAD validation")
	ErrInvalidHeadless = fmt.Errorf("%w: %s", ErrHEADValidation, "invalid headless checkout")
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
		re = regexp.MustCompile("vergo:(major|minor|patch)-release")
	} else {
		re = regexp.MustCompile("vergo:" + tagPrefixRaw + ":(major|minor|patch)-release")
	}
	match := re.FindStringSubmatch(aString)
	if len(match) != 2 {
		return "", fmt.Errorf("%w: %s", ErrNoIncrement, tagPrefixRaw)
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
				if isCommitOnBranch(repo, head.Hash().String(), branchRef) {
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
		return fmt.Errorf("%w: branch %s is not in main branches list: %s", ErrHEADValidation,
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

func isCommitOnBranch(repo *gogit.Repository, hash string, branch plumbing.ReferenceName) bool {
	//todo handle errors
	commit := plumbing.NewHash(hash)
	reference, _ := repo.Reference(branch, true)
	memo := make(map[plumbing.Hash]bool)
	reaches, _ := reaches(repo, reference.Hash(), commit, memo)
	return reaches
}

func reaches(r *gogit.Repository, start, c plumbing.Hash, memo map[plumbing.Hash]bool) (bool, error) {
	if v, ok := memo[start]; ok {
		return v, nil
	}
	if start == c {
		memo[start] = true
		return true, nil
	}
	co, err := r.CommitObject(start)
	if err != nil {
		return false, err
	}
	for _, p := range co.ParentHashes {
		v, err := reaches(r, p, c, memo)
		if err != nil {
			return false, err
		}
		if v {
			memo[start] = true
			return true, nil
		}
	}
	memo[start] = false
	return false, nil
}
