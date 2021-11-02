package cmd_test

import (
	"bytes"
	"errors"
	"github.com/Masterminds/semver/v3"
	gogit "github.com/go-git/go-git/v5"
	. "github.com/sky-uk/umc-shared/vergo/cmd"
	vergo "github.com/sky-uk/umc-shared/vergo/git"
	. "github.com/sky-uk/umc-shared/vergo/internal-test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func bumpSuccess(t *testing.T) BumpFunc {
	t.Helper()
	return func(repo *gogit.Repository, tagPrefix, increment string, versionedBranches []string, dryRun bool) (*semver.Version, error) {
		return NewVersionT(t, "0.1.0"), nil
	}
}

func mockPushTagSuccess(repo *gogit.Repository, socket, version, prefix, remote string, dryRun bool) error {
	return nil
}

func mockPushTagFailure(repo *gogit.Repository, socket, version, prefix, remote string, dryRun bool) error {
	return errors.New("push tag failed")
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

func makeGet(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	latest := func(repo *gogit.Repository, prefix string) (vergo.SemverRef, error) {
		return vergo.SemverRef{
			Version: NewVersionT(t, "0.1.0"),
		}, nil
	}
	previous := func(repo *gogit.Repository, prefix string) (vergo.SemverRef, error) {
		return vergo.SemverRef{
			Version: NewVersionT(t, "0.1.0"),
		}, nil
	}
	current := func(repo *gogit.Repository, prefix string, preRelease vergo.PreRelease) (vergo.SemverRef, error) {
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
	var emptyListRef = func(repo *gogit.Repository, prefix string, direction vergo.SortDirection, maxListSize int) ([]vergo.SemverRef, error) {
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
