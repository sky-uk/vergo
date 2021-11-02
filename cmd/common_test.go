package cmd_test

import (
	"bytes"
	"errors"
	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	. "github.com/sky-uk/umc-shared/vergo/cmd"
	vergo "github.com/sky-uk/umc-shared/vergo/git"
	. "github.com/sky-uk/umc-shared/vergo/internal-test"
	"github.com/sky-uk/umc-shared/vergo/release"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func bumpSuccess(t *testing.T) BumpFunc {
	t.Helper()
	return func(repo *git.Repository, tagPrefix, increment string, versionedBranches []string, dryRun bool) (*semver.Version, error) {
		return NewVersionT(t, "0.1.0"), nil
	}
}

func mockPushTagSuccess(repo *git.Repository, socket, version, prefix, remote string, dryRun bool) error {
	return nil
}

func mockPushTagFailure(repo *git.Repository, socket, version, prefix, remote string, dryRun bool) error {
	return errors.New("push tag failed")
}

func checkReleaseSuccess(t *testing.T) CheckReleaseFunc {
	t.Helper()
	return func(repo *git.Repository, tagPrefixRaw string) error {
		return nil
	}
}

func checkReleaseFail(t *testing.T) CheckReleaseFunc {
	t.Helper()
	return func(repo *git.Repository, tagPrefixRaw string) error {
		return release.ErrSkipRelease
	}
}

func makeBump(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	cmd := RootCmd()
	cmd.AddCommand(BumpCmd(bumpSuccess(t), mockPushTagSuccess))
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	return cmd, b
}

func pushTagFail(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	cmd := RootCmd()
	cmd.AddCommand(BumpCmd(bumpSuccess(t), mockPushTagFailure))
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	return cmd, b
}

func makeCheck(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	cmd := RootCmd()
	cmd.AddCommand(CheckCmd(checkReleaseSuccess(t)))
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	return cmd, b
}

func makeCheckFail(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	cmd := RootCmd()
	cmd.AddCommand(CheckCmd(checkReleaseFail(t)))
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	return cmd, b
}

func makeGet(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	latest := func(repo *git.Repository, prefix string) (vergo.SemverRef, error) {
		return vergo.SemverRef{
			Version: NewVersionT(t, "0.1.0"),
		}, nil
	}
	previous := func(repo *git.Repository, prefix string) (vergo.SemverRef, error) {
		return vergo.SemverRef{
			Version: NewVersionT(t, "0.1.0"),
		}, nil
	}
	current := func(repo *git.Repository, prefix string, preRelease vergo.PreRelease) (vergo.SemverRef, error) {
		return vergo.SemverRef{
			Version: NewVersionT(t, "0.1.0"),
		}, nil
	}

	cmd := RootCmd()
	cmd.AddCommand(GetCmd(latest, previous, current))
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	return cmd, b
}

func makeList(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	var emptyListRef = func(repo *git.Repository, prefix string, direction vergo.SortDirection, maxListSize int) ([]vergo.SemverRef, error) {
		return []vergo.SemverRef{
			{Version: NewVersionT(t, "0.2.0")},
			{Version: NewVersionT(t, "0.1.0")},
		}, nil
	}
	cmd := RootCmd()
	cmd.AddCommand(ListCmd(emptyListRef))
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	return cmd, b
}

func readBuffer(t *testing.T, buffer *bytes.Buffer) string {
	t.Helper()
	out, err := ioutil.ReadAll(buffer)
	assert.Nil(t, err)
	return string(out)
}
