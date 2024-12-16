package cmd_test

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	vergo "github.com/sky-uk/vergo/git"
	. "github.com/sky-uk/vergo/internal-test"
	"github.com/sky-uk/vergo/release"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var useTestDefaultForCurrentVersion vergo.CurrentVersionFunc = nil

func TestGetAllValidArgsAndAliases(t *testing.T) {
	_, tempDir := PersistentRepository(t)

	args := []string{"latest-release", "previous-release", "current-version"}
	aliases := []string{"lr", "cv", "pr"}
	for _, arg := range append(args, aliases...) {
		cmd, buffer := makeGet(t, useTestDefaultForCurrentVersion)
		cmd.SetArgs([]string{"get", arg, "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error"})
		err := cmd.Execute()
		assert.Nil(t, err)
		assert.Equal(t, "0.1.0", readBuffer(t, buffer))
	}
}

func TestGetCurrentVersionShouldReturnDefaultVersionWhenRepoIsEmpty(t *testing.T) {
	_, tempDir := PersistentRepository(t)

	args := []string{"current-version"}
	aliases := []string{"cv"}
	for _, arg := range append(args, aliases...) {
		cmd, buffer := makeGet(t, func(_ *git.Repository, _ string, _ release.PreReleaseFunc, _ vergo.GetOptions) (vergo.SemverRef, error) {
			return vergo.EmptyRef, plumbing.ErrReferenceNotFound
		})
		cmd.SetArgs([]string{"get", arg, "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error"})
		err := cmd.Execute()
		assert.Nil(t, err)
		assert.Equal(t, "0.0.0-SNAPSHOT", readBuffer(t, buffer))
	}
}

func TestGetCurrentVersionShouldReturnDefaultVersionWhenNoTagFound(t *testing.T) {
	_, tempDir := PersistentRepository(t)

	args := []string{"current-version"}
	aliases := []string{"cv"}
	for _, arg := range append(args, aliases...) {
		cmd, buffer := makeGet(t, func(_ *git.Repository, _ string, _ release.PreReleaseFunc, _ vergo.GetOptions) (vergo.SemverRef, error) {
			return vergo.EmptyRef, vergo.ErrNoTagFound
		})
		cmd.SetArgs([]string{"get", arg, "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error"})
		err := cmd.Execute()
		assert.Nil(t, err)
		assert.Equal(t, "0.0.0-SNAPSHOT", readBuffer(t, buffer))
	}
}

func TestGetAllValidArgsAndAliasesWithPrefix(t *testing.T) {
	_, tempDir := PersistentRepository(t)

	args := []string{"latest-release", "previous-release", "current-version"}
	aliases := []string{"lr", "cv", "pr"}
	for _, arg := range append(args, aliases...) {
		cmd, buffer := makeGet(t, useTestDefaultForCurrentVersion)
		cmd.SetArgs([]string{"get", arg, "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error", "-p"})
		err := cmd.Execute()
		assert.Nil(t, err)
		assert.Equal(t, "some-prefix-0.1.0", readBuffer(t, buffer))
	}
}

func TestGetAllValidArgsAndAliasesWithSlashPrefix(t *testing.T) {
	_, tempDir := PersistentRepository(t)

	args := []string{"latest-release", "previous-release", "current-version"}
	aliases := []string{"lr", "cv", "pr"}
	for _, arg := range append(args, aliases...) {
		cmd, buffer := makeGet(t, useTestDefaultForCurrentVersion)
		cmd.SetArgs([]string{"get", arg, "--repository-location", tempDir, "-t", "prefix/v", "--log-level", "error", "-p"})
		err := cmd.Execute()
		assert.Nil(t, err)
		assert.Equal(t, "prefix/v0.1.0", readBuffer(t, buffer))
	}
}

func TestGetDetectDotGit(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	tempDirWithInnerFolders := tempDir + "/level1/level2/level3"
	assert.Nil(t, os.MkdirAll(tempDirWithInnerFolders, os.ModePerm))
	cmd, buffer := makeGet(t, useTestDefaultForCurrentVersion)
	cmd.SetArgs([]string{"get", "lr", "--repository-location", tempDirWithInnerFolders, "-t", "some-prefix", "--log-level", "error"})
	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Equal(t, "0.1.0", readBuffer(t, buffer))
}
