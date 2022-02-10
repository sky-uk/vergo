package release_test

import (
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	. "github.com/sky-uk/umc-shared/vergo/internal-test"
	"github.com/sky-uk/umc-shared/vergo/release"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	prefixes   = []string{"", "app", "application", "app/v"}
	increments = []string{"patch", "minor", "major"}
	mainBranch = []string{"master", "main"}
	remoteName = "origin"
)

//nolint:scopelint,paralleltest
func TestShouldVerifySkipReleaseHint(t *testing.T) {
	testCases := []struct {
		tagPrefix string
		messages  []string
	}{
		{
			tagPrefix: "",
			messages: []string{
				"[vergo:skip-release] doc update",
				"@vergo:skip-release@ doc update",
			},
		},
		{
			tagPrefix: "app",
			messages: []string{
				"[vergo:app:skip-release] doc update",
				"@vergo:app:skip-release@ doc update",
			},
		},
	}
	for _, testCase := range testCases {
		for _, message := range testCase.messages {
			t.Run(testCase.tagPrefix+message, func(t *testing.T) {
				r := NewTestRepo(t)
				assert.Nil(t, release.SkipHintPresent(r.Repo, testCase.tagPrefix))

				DoCommitWithMessage(t, r.Repo, "some content 1", message)
				assert.ErrorIs(t, release.SkipHintPresent(r.Repo, testCase.tagPrefix), release.ErrSkipRelease)

				DoCommitWithMessage(t, r.Repo, "some content 2", "another commit")
				assert.Nil(t, release.SkipHintPresent(r.Repo, testCase.tagPrefix))

				DoCommitWithMessage(t, r.Repo, "some content 3", "another commit"+message)
				assert.ErrorIs(t, release.SkipHintPresent(r.Repo, testCase.tagPrefix), release.ErrSkipRelease)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestShouldFailWhenNotOnMainBranch(t *testing.T) {
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r := NewTestRepo(t)
				branchName := "apple"
				err := r.Worktree().Checkout(&gogit.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(branchName), Create: true})
				assert.Nil(t, err)
				r.BranchExists(branchName)
				assert.Equal(t, branchName, r.Head().Name().Short())
				err = release.ValidateHEAD(r.Repo, remoteName, mainBranch)
				assert.Regexp(t, "branch apple is not in main branches list: master, main", err)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestShouldWorkWhenHeadlessCheckoutOfMainBranch(t *testing.T) {
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r := NewTestRepo(t)
				err := r.Worktree().Checkout(&gogit.CheckoutOptions{Hash: r.Head().Hash()})
				assert.Nil(t, err)
				assert.Equal(t, plumbing.HEAD.String(), r.Head().Name().Short())

				err = release.ValidateHEAD(r.Repo, remoteName, mainBranch)
				assert.Nil(t, err)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestShouldNOTWorkWhenHeadlessCheckoutOfOtherBranch(t *testing.T) {
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

				err = release.ValidateHEAD(r.Repo, remoteName, mainBranch)
				assert.Regexp(t, "invalid headless checkout", err)
			})
		}
	}
}
