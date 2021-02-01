package bump_test

import (
	"errors"
	"github.com/Masterminds/semver"
	"github.com/go-git/go-billy/v5/memfs"
	. "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	. "sky.uk/vergo/bump"
	. "sky.uk/vergo/internal"
	"testing"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

const noFirstVersion = ""
const firstVersion = "0.1.0"
const dryRun = false
const DontSkipLatestTagOnTheHead = false
const skipLatestTagOnTheHead = true

//nolint:scopelint,paralleltest
func TestShouldIncrementPatch(t *testing.T) {
	versions := []struct {
		increment string
		pre       *semver.Version
		post      *semver.Version
	}{
		{
			increment: "patch",
			pre:       NewVersionT(t, "v0.1.0"),
			post:      NewVersionT(t, "v0.1.1"),
		},
		{
			increment: "minor",
			pre:       NewVersionT(t, "v0.1.0"),
			post:      NewVersionT(t, "v0.2.0"),
		},
		{
			increment: "major",
			pre:       NewVersionT(t, "v0.1.0"),
			post:      NewVersionT(t, "v1.0.0"),
		},
	}
	for _, version := range versions {
		t.Run(version.increment, func(t *testing.T) {
			actual, err := NextVersion(version.increment, *version.pre)
			assert.Nil(t, err)
			assert.Equal(t, actual, *version.post)
		})
	}
}

//nolint:scopelint,paralleltest
func TestNoTagNoCommit(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	increments := []string{"patch", "minor", "major"}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				fs := memfs.New()
				r, err := Init(memory.NewStorage(), fs)
				assert.Nil(t, err)
				options := Options{
					VersionedBranches:      []string{"master"},
					FirstVersionIfNoTag:    noFirstVersion,
					SkipLatestTagOnTheHead: DontSkipLatestTagOnTheHead,
					DryRun:                 dryRun,
				}
				_, err = Bump(r, prefix, increment, options)
				assert.Regexp(t, "reference not found", err)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestNoTag(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	increments := []string{"patch", "minor", "major"}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r := InMemoryRepositoryWithDefaultCommit(t)
				options := Options{
					VersionedBranches:      []string{"master"},
					FirstVersionIfNoTag:    noFirstVersion,
					SkipLatestTagOnTheHead: DontSkipLatestTagOnTheHead,
					DryRun:                 dryRun,
				}
				_, err := Bump(r, prefix, increment, options)
				assert.Regexp(t, "no tag found", err)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestCreateFirstTag(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	increments := []string{"patch", "minor", "major"}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r := InMemoryRepositoryWithDefaultCommit(t)
				options := Options{
					VersionedBranches:      []string{"master"},
					FirstVersionIfNoTag:    firstVersion,
					SkipLatestTagOnTheHead: DontSkipLatestTagOnTheHead,
					DryRun:                 dryRun,
				}
				newVersion, err := Bump(r, prefix, increment, options)
				assert.Nil(t, err)
				assert.Equal(t, NewVersionT(t, firstVersion), newVersion)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestRefOnTheHead(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	increments := []string{"patch", "minor", "major"}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r := InMemoryRepositoryWithDefaultCommit(t)
				{
					options := Options{
						VersionedBranches:      []string{"master"},
						FirstVersionIfNoTag:    firstVersion,
						SkipLatestTagOnTheHead: DontSkipLatestTagOnTheHead,
						DryRun:                 dryRun,
					}
					newVersion, err := Bump(r, prefix, increment, options)
					assert.Nil(t, err)
					assert.Equal(t, NewVersionT(t, firstVersion), newVersion)
				}
				{
					options := Options{
						VersionedBranches:      []string{"master"},
						FirstVersionIfNoTag:    firstVersion,
						SkipLatestTagOnTheHead: DontSkipLatestTagOnTheHead,
						DryRun:                 dryRun,
					}
					_, err := Bump(r, prefix, increment, options)
					assert.True(t, errors.Is(err, ErrRefOnTheHead))
				}
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestSkipRefOnTheHead(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	increments := []string{"patch", "minor", "major"}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r := InMemoryRepositoryWithDefaultCommit(t)
				{
					options := Options{
						VersionedBranches:      []string{"master"},
						FirstVersionIfNoTag:    firstVersion,
						SkipLatestTagOnTheHead: DontSkipLatestTagOnTheHead,
						DryRun:                 dryRun,
					}
					newVersion, err := Bump(r, prefix, increment, options)
					assert.Nil(t, err)
					assert.Equal(t, NewVersionT(t, firstVersion), newVersion)
				}
				{
					options := Options{
						VersionedBranches:      []string{"master"},
						FirstVersionIfNoTag:    firstVersion,
						SkipLatestTagOnTheHead: skipLatestTagOnTheHead,
						DryRun:                 dryRun,
					}
					newVersion, err := Bump(r, prefix, increment, options)
					assert.Nil(t, err)
					assert.Equal(t, NewVersionT(t, firstVersion), newVersion)
				}
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestNotOnMainBranch(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	increments := []string{"patch", "minor", "major"}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r := InMemoryRepositoryWithDefaultCommit(t)
				wt, err := r.Worktree()
				assert.Nil(t, err)
				branchName := "apple"
				err = wt.Checkout(&CheckoutOptions{Branch: plumbing.NewBranchReferenceName(branchName), Create: true})
				if err != nil {
					assert.FailNow(t, "checkout failed", err)
				}

				branches, err := r.Branches()
				assert.Nil(t, err)
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
				assert.True(t, branchExists)
				branches.Close()
				head, err := r.Head()
				assert.Nil(t, err)
				assert.Equal(t, branchName, head.Name().Short())
				options := Options{
					VersionedBranches:      []string{"master"},
					FirstVersionIfNoTag:    noFirstVersion,
					SkipLatestTagOnTheHead: DontSkipLatestTagOnTheHead,
					DryRun:                 dryRun,
				}
				_, err = Bump(r, prefix, increment, options)
				assert.Regexp(t, "command disabled for branches", err)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestHeadlessCheckout(t *testing.T) {
	prefixes := []string{""}
	increments := []string{"patch"}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r := InMemoryRepositoryWithDefaultCommit(t)
				wt, err := r.Worktree()
				assert.Nil(t, err)
				head, err := r.Head()
				assert.Nil(t, err)
				err = wt.Checkout(&CheckoutOptions{Hash: head.Hash()})
				assert.Nil(t, err)
				head, err = r.Head()
				assert.Nil(t, err)
				assert.Equal(t, "HEAD", head.Name().Short())
				options := Options{
					VersionedBranches:      []string{"HEAD"},
					FirstVersionIfNoTag:    firstVersion,
					SkipLatestTagOnTheHead: DontSkipLatestTagOnTheHead,
					DryRun:                 dryRun,
				}
				_, err = Bump(r, prefix, increment, options)
				assert.Nil(t, err)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestBump(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	versions := []struct {
		increment         string
		versionedBranches []string
		pre               *semver.Version
		post              *semver.Version
	}{
		{
			increment:         "patch",
			versionedBranches: []string{"master"},
			pre:               NewVersionT(t, "0.1.0"),
			post:              NewVersionT(t, "0.1.1"),
		},
		{
			increment:         "minor",
			versionedBranches: []string{"master"},
			pre:               NewVersionT(t, "0.1.0"),
			post:              NewVersionT(t, "0.2.0"),
		},
		{
			increment:         "major",
			versionedBranches: []string{"master"},
			pre:               NewVersionT(t, "0.1.0"),
			post:              NewVersionT(t, "1.0.0"),
		},
	}
	for _, prefix := range prefixes {
		for _, version := range versions {
			t.Run(prefix+"-"+version.increment, func(t *testing.T) {
				r := InMemoryRepositoryWithDefaultCommit(t)
				head, _ := r.Head()

				ref, err := r.CreateTag(prefix+version.pre.String(), head.Hash(), nil)
				assert.Nil(t, err)
				assert.NotNil(t, ref)

				DoCommit(t, r, "bar")

				options := Options{
					VersionedBranches:      version.versionedBranches,
					FirstVersionIfNoTag:    noFirstVersion,
					SkipLatestTagOnTheHead: DontSkipLatestTagOnTheHead,
					DryRun:                 dryRun,
				}
				actualTag, err := Bump(r, prefix, version.increment, options)
				assert.Nil(t, err)
				assert.Equal(t, *version.post, *actualTag)
				_, err = Bump(r, prefix, version.increment, options)
				assert.NotNil(t, err)
			})
		}
	}
}
