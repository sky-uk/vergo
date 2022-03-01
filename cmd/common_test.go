package cmd_test

import (
	"bytes"
	"errors"
	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/sky-uk/umc-shared/vergo/bump"
	. "github.com/sky-uk/umc-shared/vergo/cmd"
	vergo "github.com/sky-uk/umc-shared/vergo/git"
	. "github.com/sky-uk/umc-shared/vergo/internal-test"
	"github.com/sky-uk/umc-shared/vergo/release"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

var success error = nil

func bumpSuccess(t *testing.T) bump.BumpFunc {
	t.Helper()
	return func(repo *git.Repository, tagPrefix, increment string, versionedBranches []string, dryRun bool) (*semver.Version, error) {
		return NewVersionT(t, "0.1.0"), nil
	}
}

func mockPushTagSuccess(_ *git.Repository, _, _, _ string, _ bool) error {
	return nil
}

func mockPushTagFailure(_ *git.Repository, _, _, _ string, _ bool) error {
	return errors.New("push tag failed")
}

func checkReleaseDependencies(t *testing.T, err1 error, err2 error, err3 error) (release.SkipHintPresentFunc,
	release.ValidateHEADFunc, release.IncrementHintFunc) {
	t.Helper()
	return func(repo *git.Repository, tagPrefixRaw string) error {
			return err1
		}, func(repo *git.Repository, remote string, versionedBranches []string) error {
			return err2
		}, func(repo *git.Repository, tagPrefixRaw string) (string, error) {
			if err3 == nil {
				return "some-increment", nil
			} else {
				return "", err3
			}
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

func makeBumpFunc(t *testing.T, bump bump.BumpFunc) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	cmd := RootCmd()
	cmd.AddCommand(BumpCmd(bump, mockPushTagSuccess))
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
	cmd.AddCommand(CheckCmd(checkReleaseDependencies(t, success, success, success)))
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	return cmd, b
}

func makeCheckFail(t *testing.T, err1 error, err2 error, err3 error) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	cmd := RootCmd()
	cmd.AddCommand(CheckCmd(checkReleaseDependencies(t, err1, err2, err3)))
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
