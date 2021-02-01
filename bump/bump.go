package bump

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	. "github.com/go-git/go-git/v5"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"sky.uk/vergo/git"
	"strings"
)

var ErrUnknownIncrementor = errors.New("unknown incrementor")

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

var (
	ErrRefOnTheHead = errors.New("ref is on the head")
	ErrBump         = errors.New("bump failed")
)

type Options struct {
	VersionedBranches      []string
	FirstVersionIfNoTag    string
	SkipLatestTagOnTheHead bool
	DryRun                 bool
}

func Bump(repo *Repository, tagPrefix, increment string, options Options) (*semver.Version, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	log.Debugf("Current branch:%v, short: %v", head.Name(), head.Name().Short())
	if !funk.ContainsString(options.VersionedBranches, head.Name().Short()) {
		return nil, fmt.Errorf("%w : %s", ErrBump, "command disabled for branches")
	}
	latest, err := git.LatestRef(repo, tagPrefix)
	switch {
	case errors.Is(err, git.ErrNoTagFound) && options.FirstVersionIfNoTag == "":
		return nil, err
	case errors.Is(err, git.ErrNoTagFound) && options.FirstVersionIfNoTag != "":
		newVersion, err := semver.NewVersion(options.FirstVersionIfNoTag)
		if err != nil {
			return nil, err
		}
		if err := git.CreateTag(repo, newVersion.String(), tagPrefix, options.DryRun); err != nil {
			log.WithError(err).Errorln("Failed to create tag", tagPrefix, newVersion.String())
			return nil, err
		}
		return newVersion, nil
	case latest.Ref.Hash() == head.Hash() && !options.SkipLatestTagOnTheHead:
		return nil, ErrRefOnTheHead
	case latest.Ref.Hash() == head.Hash() && options.SkipLatestTagOnTheHead:
		return latest.Version, nil
	default:
		newVersion, err := NextVersion(increment, *latest.Version)
		if err != nil {
			return nil, err
		}
		if err := git.CreateTag(repo, newVersion.String(), tagPrefix, options.DryRun); err != nil {
			return nil, err
		}
		return &newVersion, nil
	}
}
