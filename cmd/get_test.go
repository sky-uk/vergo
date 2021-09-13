package cmd_test

import (
	"github.com/stretchr/testify/assert"
	"os"
	. "sky.uk/vergo/internal-test"
	"testing"
)

func TestGetAllValidArgsAndAliases(t *testing.T) {
	_, tempDir := PersistentRepository(t)

	args := []string{"latest-release", "previous-release", "current-version"}
	aliases := []string{"lr", "cv", "pr"}
	for _, arg := range append(args, aliases...) {
		cmd, buffer := makeGet(t)
		cmd.SetArgs([]string{"get", arg, "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error"})
		err := cmd.Execute()
		assert.Nil(t, err)
		assert.Equal(t, "0.1.0", readBuffer(t, buffer))
	}
}

func TestGetAllValidArgsAndAliasesWithPrefix(t *testing.T) {
	_, tempDir := PersistentRepository(t)

	args := []string{"latest-release", "previous-release", "current-version"}
	aliases := []string{"lr", "cv", "pr"}
	for _, arg := range append(args, aliases...) {
		cmd, buffer := makeGet(t)
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
		cmd, buffer := makeGet(t)
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
	cmd, buffer := makeGet(t)
	cmd.SetArgs([]string{"get", "lr", "--repository-location", tempDirWithInnerFolders, "-t", "some-prefix", "--log-level", "error"})
	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Equal(t, "0.1.0", readBuffer(t, buffer))
}
