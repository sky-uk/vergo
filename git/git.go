package git

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/agent"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
)

const refPrefix = "refs/tags/"

type SortDirection string

const (
	asc  = "asc"
	desc = "desc"
	ASC  = SortDirection(asc)
	DESC = SortDirection(desc)
)

var (
	ErrNoTagFound           = errors.New("no tag found")
	ErrOneTagFound          = errors.New("one tag found")
	ErrInvalidSortDirection = errors.New("invalid sort direction")
	ErrPreReleaseVersion    = errors.New("invalid preReleaseVersion")
)

func ParseSortDirection(str string) (SortDirection, error) {
	str = strings.TrimSpace(strings.ToLower(str))
	switch {
	case str == asc:
		return ASC, nil
	case str == desc:
		return DESC, nil
	default:
		return "", fmt.Errorf("%w : %s", ErrInvalidSortDirection, str)
	}
}

func TagExists(r *gogit.Repository, tag string) (bool, error) {
	tags, err := r.Tags()
	tag = refPrefix + tag
	if err != nil {
		return false, err
	}
	found := false
	err = tags.ForEach(func(r *plumbing.Reference) error {
		if r.Name().String() == tag {
			found = true
		}
		return nil
	})
	return found, err
}

func CreateTagWithMessage(repo *gogit.Repository, version, prefix, message string,
	tagger *object.Signature, dryRun bool) error {
	tag := prefix + version
	found, err := TagExists(repo, tag)
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("%w : %s", gogit.ErrTagExists, tag)
	}
	log.Infof("Set tag %s", tag)
	h, err := repo.Head()
	if err != nil {
		return err
	}
	if dryRun {
		log.Infof("Dry run: create tag %v", tag)
	} else {
		if strings.TrimSpace(message) == "" {
			_, err = repo.CreateTag(tag, h.Hash(), nil)
		} else {
			_, err = repo.CreateTag(tag, h.Hash(), &gogit.CreateTagOptions{
				Tagger:  tagger,
				Message: message,
			})
		}
		if err != nil {
			return fmt.Errorf("%w : %s", err, "create tag error")
		}
	}

	return nil
}

func CreateTag(repo *gogit.Repository, version, prefix string, dryRun bool) error {
	return CreateTagWithMessage(repo, version, prefix, "", nil, dryRun)
}

func PushTag(r *gogit.Repository, socket, version, prefix, remote string, dryRun bool) error {
	tag := prefix + version

	conn, err := net.Dial("unix", socket)
	if err != nil {
		log.WithError(err).Fatalln("Failed to open SSH_AUTH_SOCK")
	}

	agentClient := agent.NewClient(conn)
	defer func() {
		_ = conn.Close()
	}()

	signers, err := agentClient.Signers()
	if err != nil {
		log.WithError(err).Fatalln("failed to get signers")
	}

	auth := &ssh.PublicKeys{
		User:   "git",
		Signer: signers[0],
	}

	log.Debugf("Pushing tag: %v", tag)
	refSpec := config.RefSpec(fmt.Sprintf("refs/tags/%s:refs/tags/%s", tag, tag))
	po := &gogit.PushOptions{
		RemoteName: remote,
		Progress:   os.Stdout,
		RefSpecs:   []config.RefSpec{refSpec},
		Auth:       auth,
	}

	if dryRun {
		log.Infof("Dry run: push tag %v", tag)
	} else {
		err = r.Push(po)

		if err != nil {
			if errors.Is(err, gogit.NoErrAlreadyUpToDate) {
				log.Print("origin remote was up to date, no push done")
				return nil
			}
			log.Infof("push to remote origin error: %s", err)
			return err
		}
	}
	return nil
}

func ListRefs(repo *gogit.Repository, prefix string, direction SortDirection, maxListSize int) ([]SemverRef, error) {
	versions, err := refsWithPrefix(repo, prefix)
	if err != nil {
		return EmptyRefList, err
	}
	less := func(i, j int) bool {
		return versions[i].Version.LessThan(versions[j].Version)
	}

	greater := func(i, j int) bool {
		return versions[i].Version.GreaterThan(versions[j].Version)
	}
	maxRange := func() int {
		_len := len(versions)
		if _len < maxListSize {
			return _len
		}
		return maxListSize
	}
	if direction == ASC {
		sort.Slice(versions, less)
	} else {
		sort.Slice(versions, greater)
	}
	return versions[0:maxRange()], nil
}

func LatestRef(repo *gogit.Repository, prefix string) (SemverRef, error) {
	versions, err := reversedRefsWithPrefix(repo, prefix)
	if err != nil {
		return EmptyRef, err
	}

	latestVersion := versions[0]
	log.Debugf("Latest version: %v\n", latestVersion)
	return latestVersion, nil
}

func PreviousRef(repo *gogit.Repository, prefix string) (SemverRef, error) {
	versions, err := reversedRefsWithPrefix(repo, prefix)
	if err != nil {
		return EmptyRef, err
	}
	if len(versions) == 1 {
		return EmptyRef, ErrOneTagFound
	}
	latestVersion := versions[1]
	log.Debugf("Previous version: %v\n", latestVersion)
	return latestVersion, nil
}

func reversedRefsWithPrefix(repo *gogit.Repository, prefix string) ([]SemverRef, error) {
	versions, err := refsWithPrefix(repo, prefix)
	if err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, ErrNoTagFound
	}
	sort.Sort(sort.Reverse(SemverRefColl(versions)))
	return versions, nil
}

func refsWithPrefix(repo *gogit.Repository, prefix string) ([]SemverRef, error) {
	tagPrefix := refPrefix + prefix
	re := regexp.MustCompile("^" + tagPrefix + semver.SemVerRegex + "$")

	tagRefs, err := repo.Tags()
	if err != nil {
		return EmptyRefList, err
	}

	var versions []SemverRef
	err = tagRefs.ForEach(func(t *plumbing.Reference) error {
		tag := t.Name().String()
		if re.MatchString(tag) {
			versionString := strings.TrimLeft(tag, tagPrefix)
			if version, err := semver.NewVersion(versionString); err == nil {
				versions = append(versions, SemverRef{Version: version, Ref: t})
			} else {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return EmptyRefList, err
	}
	return versions, nil
}

type PreRelease func(version *semver.Version) (semver.Version, error)

func CurrentVersion(repo *gogit.Repository, prefix string, preRelease PreRelease) (SemverRef, error) {
	head, err := repo.Head()
	if err != nil {
		return EmptyRef, err
	}
	sortedTagRefs, err := reversedRefsWithPrefix(repo, prefix)
	if err != nil {
		return EmptyRef, err
	}
	for _, tagRef := range sortedTagRefs {
		obj, err := repo.TagObject(tagRef.Ref.Hash())
		switch err {
		case nil:
			// Tag object present
			if obj.Target == head.Hash() {
				return SemverRef{
					Version: tagRef.Version,
					Ref:     head,
				}, nil
			}
		case plumbing.ErrObjectNotFound:
			// Not a tag object
			if tagRef.Ref.Hash() == head.Hash() {
				return tagRef, nil
			}
		default:
			return EmptyRef, err
		}
	}
	latest, err := LatestRef(repo, prefix)
	if err != nil {
		return EmptyRef, err
	}
	preReleaseVersion, err := preRelease(latest.Version)
	if err != nil {
		return EmptyRef, err
	}
	if preReleaseVersion.Prerelease() == "" {
		return EmptyRef, fmt.Errorf("%w : %s", ErrPreReleaseVersion, "preReleaseVersion must have prerelease part")
	}
	if !preReleaseVersion.GreaterThan(latest.Version) {
		return EmptyRef, fmt.Errorf("%w : %s", ErrPreReleaseVersion, "preReleaseVersion must create a greater version")
	}
	return SemverRef{
		Version: &preReleaseVersion,
		Ref:     head,
	}, nil
}
