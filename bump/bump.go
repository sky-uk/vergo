package bump

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/umc-shared/vergo/git"
	"github.com/sky-uk/umc-shared/vergo/release"
	"strings"
)

var (
	ErrUnknownIncrementor = errors.New("unknown incrementor")
)

func NextVersion(increment string, version semver.Version) (incrementedVersion semver.Version, err error) {
	switch strings.ToLower(increment) {
	case "patch":
		incrementedVersion = version.IncPatch()
	case "minor":
		incrementedVersion = version.IncMinor()
	case "major":
		incrementedVersion = version.IncMajor()
	default:
		err = fmt.Errorf("%w : %s", ErrUnknownIncrementor, increment)
	}
	return
}

const (
	firstVersion  = "0.1.0"
	defaultRemote = "origin"
)

type BumpFunc func(
	repo *gogit.Repository,
	tagPrefix, increment string,
	versionedBranches []string,
	dryRun bool) (*semver.Version, error)

func Bump(repo *gogit.Repository, tagPrefix, increment string, versionedBranches []string, dryRun bool) (*semver.Version, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	if err := release.ValidateHEAD(repo, defaultRemote, versionedBranches); err != nil {
		return nil, err
	}

	latest, err := git.LatestRef(repo, tagPrefix)
	if errors.Is(err, git.ErrNoTagFound) {
		newVersion, err := semver.NewVersion(firstVersion)
		if err != nil {
			return nil, err
		}
		if err := git.CreateTag(repo, newVersion.String(), tagPrefix, dryRun); err != nil {
			log.WithError(err).Errorln("Failed to create tag", tagPrefix, newVersion.String())
			return nil, err
		}
		return newVersion, nil
	}
	switch tagObject, err := repo.TagObject(latest.Ref.Hash()); {
	case err == nil && tagObject.Target == head.Hash() && tagObject.TargetType == plumbing.CommitObject:
		// Tag object present
		return latest.Version, nil
	case err == plumbing.ErrObjectNotFound && latest.Ref.Hash() == head.Hash():
		// Not a tag object
		return latest.Version, nil
	case err == nil:
		break
	case err != plumbing.ErrObjectNotFound:
		return nil, err
	}
	newVersion, err := NextVersion(increment, *latest.Version)
	if err != nil {
		return nil, err
	}
	if err := git.CreateTag(repo, newVersion.String(), tagPrefix, dryRun); err != nil {
		return nil, err
	}
	return &newVersion, nil
}
