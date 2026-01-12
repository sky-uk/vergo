package bump

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/vergo/git"
	"github.com/sky-uk/vergo/release"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrUnknownIncrementor = errors.New("unknown incrementor")
)

func NextVersion(increment string, version semver.Version) (incrementedVersion semver.Version, err error) {
	switch strings.ToLower(increment) {
	case "prerelease":
		incrementedVersion = NextPreRelease(version)
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

func NextPreRelease(version semver.Version) (incrementedVersion semver.Version) {
	var currentPreRelease = version.Prerelease()
	if currentPreRelease == "" {
		incrementedVersion, _ = version.SetPrerelease("alpha1")
	} else {
		suffix, num := ParsePreRelease(currentPreRelease)
		incrementedVersion, _ = version.SetPrerelease(fmt.Sprintf(suffix+"%d", num+1))
	}
	return
}

func ParsePreRelease(preRelease string) (suffix string, number int) {
	re := regexp.MustCompile(`([a-zA-Z]+)(\d+)`)
	match := re.FindStringSubmatch(preRelease)
	suffix = match[1]
	number, _ = strconv.Atoi(match[2])
	return
}

const (
	firstVersion = "0.1.0"
)

type Options struct {
	TagPrefix         string
	Remote            string
	VersionedBranches []string
	DryRun            bool
	NearestRelease    bool
}

type Func func(repo *gogit.Repository, increment string, options Options) (*semver.Version, error)

func Bump(repo *gogit.Repository, increment string, options Options) (*semver.Version, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	if err := release.ValidateHEAD(repo, options.Remote, options.VersionedBranches); err != nil {
		return nil, err
	}

	var latest git.SemverRef
	if options.NearestRelease {
		latest, err = git.NearestTag(repo, options.TagPrefix)
	} else {
		latest, err = git.LatestRef(repo, options.TagPrefix)
	}

	if errors.Is(err, git.ErrNoTagFound) {
		newVersion, err := semver.NewVersion(firstVersion)
		if err != nil {
			return nil, err
		}
		if err := git.CreateTag(repo, newVersion.String(), options.TagPrefix, options.DryRun); err != nil {
			log.WithError(err).Errorln("Failed to create tag", options.TagPrefix, newVersion.String())
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
	if err := git.CreateTag(repo, newVersion.String(), options.TagPrefix, options.DryRun); err != nil {
		return nil, err
	}
	return &newVersion, nil
}
