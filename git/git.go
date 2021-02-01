package git

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	. "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
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
	ErrEndOfIteration       = errors.New("end of iteration")
	ErrNoRelevantChange     = errors.New("no relevant change")
)

func ParseSortDirection(str string) (SortDirection, error) {
	str = strings.TrimSpace(strings.ToLower(str))
	switch {
	case strings.HasPrefix(str, asc):
		return ASC, nil
	case strings.HasPrefix(str, desc):
		return DESC, nil
	default:
		return "", fmt.Errorf("%w : %s", ErrInvalidSortDirection, str)
	}
}

func TagExists(r *Repository, tag string) (bool, error) {
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

func CreateTag(repo *Repository, version string, prefix string, dryRun bool) error {
	tag := prefix + version
	found, err := TagExists(repo, tag)
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("%w : %s", ErrTagExists, tag)
	}
	log.Infof("Set tag %s", tag)
	h, err := repo.Head()
	if err != nil {
		return err
	}
	if dryRun {
		log.Infof("Dry run: create tag %v", tag)
	} else {
		_, err = repo.CreateTag(tag, h.Hash(), nil)

		if err != nil {
			return fmt.Errorf("%w : %s", err, "create tag error")
		}
	}

	return nil
}

func PushTag(r *Repository, socket, version, prefix, remote string, dryRun bool) error {
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
	po := &PushOptions{
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
			if errors.Is(err, NoErrAlreadyUpToDate) {
				log.Print("origin remote was up to date, no push done")
				return nil
			}
			log.Infof("push to remote origin error: %s", err)
			return err
		}
	}
	return nil
}

func ListRefs(repo *Repository, prefix string, direction SortDirection, maxListSize int) ([]SemverRef, error) {
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

func LatestRef(repo *Repository, prefix string) (SemverRef, error) {
	versions, err := refsWithPrefix(repo, prefix)
	if err != nil {
		return EmptyRef, err
	}
	switch {
	case len(versions) == 0:
		return EmptyRef, ErrNoTagFound
	default:
		sort.Sort(SemverRefColl(versions))
		latestVersion := versions[len(versions)-1]
		log.Debugf("Latest version: %v\n", latestVersion)
		return latestVersion, nil
	}
}
func PreviousRef(repo *Repository, prefix string) (SemverRef, error) {
	versions, err := refsWithPrefix(repo, prefix)
	if err != nil {
		return EmptyRef, err
	}
	switch {
	case len(versions) == 0:
		return EmptyRef, ErrNoTagFound
	case len(versions) == 1:
		return EmptyRef, ErrOneTagFound
	default:
		sort.Sort(SemverRefColl(versions))
		latestVersion := versions[len(versions)-2]
		log.Debugf("Previous version: %v\n", latestVersion)
		return latestVersion, nil
	}
}
func refsWithPrefix(repo *Repository, prefix string) ([]SemverRef, error) {
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

func CurrentVersion(repo *Repository, prefix string, preRelease PreRelease) (SemverRef, error) {
	latest, err := LatestRef(repo, prefix)
	if err != nil {
		return EmptyRef, err
	}
	head, err := repo.Head()
	if err != nil {
		return EmptyRef, err
	}
	if latest.Ref.Hash() == head.Hash() {
		return latest, nil
	}
	fnVersion, err := preRelease(latest.Version)
	if err != nil {
		return EmptyRef, err
	}
	if fnVersion.Prerelease() == "" {
		return EmptyRef, fmt.Errorf("%w : %s", ErrPreReleaseVersion, "preReleaseVersion must have prerelease part")
	}
	if !fnVersion.GreaterThan(latest.Version) {
		return EmptyRef, fmt.Errorf("%w : %s", ErrPreReleaseVersion, "preReleaseVersion must create a greater version")
	}
	return SemverRef{
		Version: &fnVersion,
		Ref:     head,
	}, nil
}

const LogSeparator = "--------"

// RelevantChanges prints whether changed files between HEAD and the last release for a given prefix
func LogPrefix(repo *Repository, prefix string, maxLogIteration int) error {
	head, err := repo.Head()
	if err != nil {
		return err
	}
	latestTag, err := LatestRef(repo, prefix)
	if err != nil {
		return err
	}
	tag := refPrefix + prefix + latestTag.Version.String()
	latestReleaseHash, err := repo.ResolveRevision(plumbing.Revision(tag))
	if err != nil {
		return err
	}
	log.Tracef("Latest ref tag %s with commit ref %s", tag, latestReleaseHash.String())
	iter, err := repo.Log(&LogOptions{From: head.Hash()})
	if err != nil {
		return err
	}
	previous, err := repo.CommitObject(head.Hash())
	if err != nil {
		return err
	}
	commitCounter := 0
	err = iter.ForEach(func(commit *object.Commit) error {
		commitCounter++
		if commit.Hash == head.Hash() || *latestReleaseHash == head.Hash() {
			return nil
		}
		commit, err := repo.CommitObject(commit.Hash)
		if err != nil {
			return err
		}
		patch, err := previous.Patch(commit)
		if err != nil {
			return err
		}
		fmt.Println(previous.Hash.String()[0:7], "..", commit.Hash.String()[0:7], " ", previous.Message) //nolint
		for _, filePatch := range patch.FilePatches() {
			from, to := filePatch.Files()
			fmt.Println(filePath(from), ",", filePath(to)) //nolint
		}
		fmt.Println(LogSeparator) //nolint
		if commit.Hash == *latestReleaseHash {
			return ErrEndOfIteration
		}
		if commitCounter == maxLogIteration {
			log.Infof("Reached max log iteration %d", maxLogIteration)
			return ErrEndOfIteration
		}
		previous = commit
		return nil
	})
	if errors.Is(err, ErrEndOfIteration) {
		return nil
	}
	return err
}

func filePath(file diff.File) string {
	if file == nil {
		return "N/A"
	}
	return file.Path()
}

// RelevantChanges checks whether changed files between HEAD and the last release includes a relevant change
func RelevantChanges(repo *Repository, prefix, change string, maxLogIteration int) error {
	head, err := repo.Head()
	if err != nil {
		return err
	}
	latestTag, err := LatestRef(repo, prefix)
	if err != nil {
		return err
	}
	tag := refPrefix + prefix + latestTag.Version.String()
	latestReleaseHash, err := repo.ResolveRevision(plumbing.Revision(tag))
	if err != nil {
		return err
	}
	log.Tracef("Latest ref tag %s with commit ref %s", tag, latestReleaseHash.String())
	iter, err := repo.Log(&LogOptions{From: head.Hash()})
	if err != nil {
		return err
	}
	previous, err := repo.CommitObject(head.Hash())
	if err != nil {
		return err
	}
	commitCounter := 0
	anyChange := false
	_ = iter.ForEach(func(commit *object.Commit) error {
		commitCounter++
		if commit.Hash == head.Hash() || *latestReleaseHash == head.Hash() {
			return nil
		}
		commit, err := repo.CommitObject(commit.Hash)
		if err != nil {
			return err
		}
		patch, err := previous.Patch(commit)
		if err != nil {
			return err
		}
		anyChange = anyChange || patchedFilesContainsTheChange(patch, change)
		if anyChange {
			log.Infof("Message: %s\nHelper: git diff --name-only %s..%s", previous.Message,
				previous.Hash.String()[0:7], commit.Hash.String()[0:7])
		}
		if commit.Hash == *latestReleaseHash {
			return ErrEndOfIteration
		}
		if commitCounter == maxLogIteration {
			log.Infof("Reached max log iteration %d", maxLogIteration)
			return ErrEndOfIteration
		}
		previous = commit
		return nil
	})
	if anyChange {
		return nil
	}
	return ErrNoRelevantChange
}

func patchedFilesContainsTheChange(patch *object.Patch, change string) bool {
	for _, filePatch := range patch.FilePatches() {
		from, to := filePatch.Files()
		if from != nil && strings.Contains(from.Path(), change) {
			return true
		}
		if to != nil && strings.Contains(to.Path(), change) {
			return true
		}
	}
	return false
}
