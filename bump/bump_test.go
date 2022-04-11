package bump_test

import (
	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	. "github.com/sky-uk/vergo/bump"
	. "github.com/sky-uk/vergo/git"
	. "github.com/sky-uk/vergo/internal-test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const firstVersion = "0.1.0"

var (
	prefixes   = []string{"", "app", "application", "app/v"}
	increments = []string{"patch", "minor", "major"}
	mainBranch = []string{"master", "main"}
)

//nolint:scopelint,paralleltest
func TestShouldIncrementVersion(t *testing.T) {
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
func TestBumpShouldFailWhenThereIsNoCommit(t *testing.T) {
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				fs := memfs.New()
				r, err := gogit.Init(memory.NewStorage(), fs)
				assert.Nil(t, err)
				_, err = Bump(r, increment, Options{TagPrefix: prefix, VersionedBranches: mainBranch})
				assert.Regexp(t, "reference not found", err)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestBumpShouldCreateFirstTag(t *testing.T) {
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r := NewTestRepo(t)
				newVersion, err := Bump(r.Repo, increment, Options{TagPrefix: prefix, VersionedBranches: mainBranch})
				assert.Nil(t, err)
				assert.Equal(t, NewVersionT(t, firstVersion), newVersion)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestShouldBeAbleToCallBumpMultipleTimes(t *testing.T) {
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r := NewTestRepo(t)

				firstCall, err := Bump(r.Repo, increment, Options{TagPrefix: prefix, VersionedBranches: mainBranch})
				assert.Nil(t, err)
				assert.Equal(t, NewVersionT(t, firstVersion), firstCall)

				secondCall, err := Bump(r.Repo, increment, Options{TagPrefix: prefix, VersionedBranches: mainBranch})
				assert.Nil(t, err)
				assert.Equal(t, NewVersionT(t, firstVersion), secondCall)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestBumpShouldFailWhenNotOnMainBranch(t *testing.T) {
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r := NewTestRepo(t)
				branchName := "apple"
				err := r.Worktree().Checkout(&gogit.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(branchName), Create: true})
				assert.Nil(t, err)
				r.BranchExists(branchName)
				assert.Equal(t, branchName, r.Head().Name().Short())
				_, err = Bump(r.Repo, increment, Options{TagPrefix: prefix, VersionedBranches: mainBranch})
				assert.Regexp(t, "branch apple is not in main branches list: master, main", err)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestBumpShouldWorkWhenHeadlessCheckoutOfMainBranch(t *testing.T) {
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r := NewTestRepo(t)
				err := r.Worktree().Checkout(&gogit.CheckoutOptions{Hash: r.Head().Hash()})
				assert.Nil(t, err)
				assert.Equal(t, plumbing.HEAD.String(), r.Head().Name().Short())

				_, err = Bump(r.Repo, increment, Options{TagPrefix: prefix, VersionedBranches: mainBranch})
				assert.Nil(t, err)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestBumpShouldNOTWorkWhenHeadlessCheckoutOfOtherBranch(t *testing.T) {
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r := NewTestRepo(t)
				branchName := "apple"
				err := r.Worktree().Checkout(&gogit.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(branchName), Create: true})
				assert.Nil(t, err)
				r.BranchExists(branchName)
				assert.Equal(t, branchName, r.Head().Name().Short())
				r.DoCommit("foo")
				latestHashOnApple := r.Head().Hash()

				err = r.Worktree().Checkout(&gogit.CheckoutOptions{Branch: plumbing.Master})
				assert.Nil(t, err)
				assert.Equal(t, plumbing.Master, r.Head().Name())

				err = r.Worktree().Checkout(&gogit.CheckoutOptions{Hash: latestHashOnApple})
				assert.Nil(t, err)
				assert.Equal(t, plumbing.HEAD.String(), r.Head().Name().Short())

				_, err = Bump(r.Repo, increment, Options{TagPrefix: prefix, VersionedBranches: mainBranch})
				assert.Regexp(t, "invalid headless checkout", err)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestBumpWithAnnotatedTags(t *testing.T) {
	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := NewTestRepo(t)
			tagger := &object.Signature{
				Name:  "test",
				Email: "test@test.com",
				When:  time.Now(),
			}
			err := CreateTagWithMessage(r.Repo, "0.0.1", prefix, "test message", tagger, false)
			assert.Nil(t, err)

			{
				tag, err := Bump(r.Repo, "patch", Options{TagPrefix: prefix, VersionedBranches: mainBranch})
				assert.Nil(t, err)
				assert.Equal(t, NewVersionT(t, "0.0.1"), tag)
			}

			r.DoCommit("foo")
			assert.Nil(t, CreateTag(r.Repo, "1.0.0", prefix, false))
			{
				tag, err := Bump(r.Repo, "patch", Options{TagPrefix: prefix, VersionedBranches: mainBranch})
				assert.Nil(t, err)
				assert.Equal(t, NewVersionT(t, "1.0.0"), tag)
			}

			r.DoCommit("bar")
			{
				tag, err := Bump(r.Repo, "patch", Options{TagPrefix: prefix, VersionedBranches: mainBranch})
				assert.Nil(t, err)
				assert.Equal(t, NewVersionT(t, "1.0.1"), tag)
			}
		})
	}
}

//nolint:scopelint,paralleltest
func TestBumpAllIncrements(t *testing.T) {
	versions := []struct {
		increment         string
		versionedBranches []string
		pre               *semver.Version
		post              *semver.Version
	}{
		{
			increment:         "patch",
			versionedBranches: mainBranch,
			pre:               NewVersionT(t, "0.1.0"),
			post:              NewVersionT(t, "0.1.1"),
		},
		{
			increment:         "minor",
			versionedBranches: mainBranch,
			pre:               NewVersionT(t, "0.1.0"),
			post:              NewVersionT(t, "0.2.0"),
		},
		{
			increment:         "major",
			versionedBranches: mainBranch,
			pre:               NewVersionT(t, "0.1.0"),
			post:              NewVersionT(t, "1.0.0"),
		},
	}
	for _, prefix := range prefixes {
		for _, version := range versions {
			t.Run(prefix+"-"+version.increment, func(t *testing.T) {
				r := NewTestRepo(t)
				r.CreateTag(prefix+version.pre.String(), r.Head().Hash())
				r.DoCommit("bar")

				tag, err := Bump(r.Repo, version.increment, Options{TagPrefix: prefix, VersionedBranches: version.versionedBranches})
				assert.Nil(t, err)
				assert.Equal(t, *version.post, *tag)
			})
		}
	}
}
