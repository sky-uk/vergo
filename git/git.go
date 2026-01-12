package git

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/Masterminds/semver/v3"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	log "github.com/sirupsen/logrus"
	cryptossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/sky-uk/vergo/release"
)

type SortDirection string

const (
	refTagPrefix = "refs/tags/"
	semVerRegex  = `v?([0-9]+)(\.[0-9]+)?(\.[0-9]+)?` +
		`(-([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?` +
		`(\+([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?`

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
	ErrUndefinedAuth        = errors.New("no auth has been configured, GITHUB_TOKEN or SSH_AUTH_SOCK must be set")
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
	tag = refTagPrefix + tag
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

type PushTagFunc func(
	repo *gogit.Repository,
	version, prefix, remote string,
	dryRun bool, disableStrictHostChecking bool, tokenEnvVarKey string) error

func PushTag(r *gogit.Repository, version, prefix, remote string, dryRun bool, disableStrictHostChecking bool, tokenEnvVarKey string) error {
	tag := prefix + version

	var auth transport.AuthMethod

	if githubToken, ok := os.LookupEnv(tokenEnvVarKey); ok {
		log.Debug("Using Github Bearer Token Auth")
		auth = &http.BasicAuth{
			Username: "can-be-anything",
			Password: githubToken,
		}
	} else if socket, ok := os.LookupEnv("SSH_AUTH_SOCK"); ok {
		log.Debug("Using SSH Agent Authentication")
		conn, err := net.Dial("unix", socket)
		if err != nil {
			log.WithError(err).Fatalln("Failed to open SSH_AUTH_SOCK")
		}

		agentClient := agent.NewClient(conn)
		defer func() {
			_ = conn.Close()
		}()

		sshAuth := generateSshAuth(agentClient, disableStrictHostChecking)

		auth = sshAuth
	} else {
		return ErrUndefinedAuth
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
		err := r.Push(po)

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

func generateSshAuth(agentClient agent.ExtendedAgent, disableStrictHostChecking bool) *ssh.PublicKeys {
	signers, err := agentClient.Signers()
	if err != nil || len(signers) == 0 {
		log.WithError(err).Fatalln("failed to get signers, make sure to add private key identities to the authentication agent")
	}

	sshAuth := &ssh.PublicKeys{
		User:   "git",
		Signer: signers[0],
	}

	if disableStrictHostChecking {
		sshAuth.HostKeyCallback = cryptossh.InsecureIgnoreHostKey()
	}
	return sshAuth
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
	tagPrefix := refTagPrefix + prefix
	re := regexp.MustCompile("^" + tagPrefix + semVerRegex + "$")

	tagRefs, err := repo.Tags()
	if err != nil {
		return EmptyRefList, err
	}

	var versions []SemverRef
	err = tagRefs.ForEach(func(t *plumbing.Reference) error {
		tag := t.Name().String()
		if re.MatchString(tag) {
			versionString := strings.TrimPrefix(tag, tagPrefix)
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

type GetOptions struct {
	NearestRelease bool
}

type CurrentVersionFunc func(repo *gogit.Repository, prefix string, preRelease release.PreReleaseFunc, options GetOptions) (SemverRef, error)

func CurrentVersion(repo *gogit.Repository, prefix string, preRelease release.PreReleaseFunc, options GetOptions) (SemverRef, error) {
	head, err := repo.Head()
	if err != nil {
		return EmptyRef, err
	}
	sortedTagRefs, err := reversedRefsWithPrefix(repo, prefix)
	if err != nil {
		return EmptyRef, err
	}

	for _, tagRef := range sortedTagRefs {
		switch tagObject, err := repo.TagObject(tagRef.Ref.Hash()); {
		case err == nil && tagObject.Target == head.Hash() && tagObject.TargetType == plumbing.CommitObject:
			// Tag object present
			return SemverRef{
				Version: tagRef.Version,
				Ref:     head,
			}, nil
		case err == plumbing.ErrObjectNotFound && tagRef.Ref.Hash() == head.Hash():
			// Not a tag object
			return tagRef, nil
		case err == nil:
			break
		case err != plumbing.ErrObjectNotFound:
			return EmptyRef, err
		}
	}

	var latest SemverRef
	if options.NearestRelease {
		latest, err = NearestTag(repo, prefix)
	} else {
		latest, err = LatestRef(repo, prefix)
	}
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

func NearestTag(repo *gogit.Repository, prefix string) (SemverRef, error) {
	head, err := repo.Head()
	if err != nil {
		return EmptyRef, err
	}

	// Build a map of commit hash -> matching tags (only once)
	tagMap, err := buildTagMap(repo, prefix)
	if err != nil {
		return EmptyRef, err
	}

	// Check HEAD first
	if tags, exists := tagMap[head.Hash()]; exists {
		return tags[0], nil // Return first matching tag
	}

	// Walk commit history and check map
	commitIter, err := repo.Log(&gogit.LogOptions{From: head.Hash()})
	if err != nil {
		return EmptyRef, fmt.Errorf("failed to get commit log: %w", err)
	}

	var nearestTag SemverRef
	err = commitIter.ForEach(func(commit *object.Commit) error {
		if tags, exists := tagMap[commit.Hash]; exists {
			nearestTag = tags[0] // Take first matching tag
			return storer.ErrStop
		}
		return nil
	})

	if err != nil {
		return EmptyRef, fmt.Errorf("failed to iterate over commits: %w", err)
	}

	if nearestTag.Version == nil {
		return EmptyRef, ErrNoTagFound
	}

	return nearestTag, nil
}

func buildTagMap(repo *gogit.Repository, prefix string) (map[plumbing.Hash][]SemverRef, error) {
	tagMap := make(map[plumbing.Hash][]SemverRef)
	tagPrefix := refTagPrefix + prefix
	re := regexp.MustCompile("^" + tagPrefix + semVerRegex + "$")

	tags, err := repo.Tags()
	if err != nil {
		return nil, err
	}

	err = tags.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().String()
		if re.MatchString(tagName) {
			versionString := strings.TrimPrefix(tagName, tagPrefix)
			if version, err := semver.NewVersion(versionString); err == nil {
				// Handle both lightweight and annotated tags
				var commitHash plumbing.Hash
				var tagRef *plumbing.Reference

				if tagObject, err := repo.TagObject(ref.Hash()); err == nil {
					// Annotated tag - use the target commit hash
					commitHash = tagObject.Target
					// Create a reference that points to the commit, not the tag object
					tagRef = plumbing.NewHashReference(ref.Name(), commitHash)
				} else if err == plumbing.ErrObjectNotFound {
					// Lightweight tag - use the tag reference hash directly
					commitHash = ref.Hash()
					tagRef = ref
				} else {
					return err
				}

				semverRef := SemverRef{Version: version, Ref: tagRef}
				tagMap[commitHash] = append(tagMap[commitHash], semverRef)
			}
		}
		return nil
	})

	// Sort tags for each commit by semantic version (descending)
	for hash := range tagMap {
		sort.Sort(sort.Reverse(SemverRefColl(tagMap[hash])))
	}

	return tagMap, err
}

func BranchExists(repo *gogit.Repository, branchName string) (bool, error) {
	branches, err := repo.Branches()
	if err != nil {
		return false, err
	}
	defer branches.Close()
	branchExists := false
	for {
		branch, err := branches.Next()
		if err != nil {
			break
		}
		if branch.Name().Short() == branchName {
			branchExists = true
		}
	}
	return branchExists, nil
}
