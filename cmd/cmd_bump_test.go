package cmd_test

import (
	"fmt"
	"github.com/sky-uk/vergo/bump"
	. "github.com/sky-uk/vergo/internal-test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	prefixes   = []string{"", "app", "application", "app/v"}
	increments = []string{"prerelease", "patch", "minor", "major"}
)

//nolint:scopelint,paralleltest
func TestBumpShouldWorkWithIncrements(t *testing.T) {
	for _, increment := range increments {
		t.Run(increment, func(t *testing.T) {
			_, tempDir := PersistentRepository(t)
			cmd, buffer := makeBump(t)
			cmd.SetArgs([]string{"bump", increment, "--repository-location", tempDir})
			err := cmd.Execute()
			assert.Nil(t, err)
			assert.Equal(t, "0.1.0", readBuffer(t, buffer))
		})
	}
}

//nolint:scopelint,paralleltest
func TestBumpShouldWorkWithAutoIncrement(t *testing.T) {
	nextVersion := []string{"0.1.0-alpha1", "0.1.1", "0.2.0", "1.0.0"}
	for _, prefix := range prefixes {
		for i, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				repo, tempDir := PersistentRepository(t)

				{
					DoCommit(t, repo, "init")
					cmd, buffer := makeBumpFunc(t, bump.Bump)
					cmd.SetArgs([]string{"bump", increment, "--repository-location", tempDir, "-t", prefix})
					err := cmd.Execute()
					assert.Nil(t, err)
					assert.Equal(t, "0.1.0", readBuffer(t, buffer))
				}

				{
					hint := increment
					if increment == "prerelease" {
						hint = "pre"
					}
					if prefix == "" {
						DoCommitWithMessage(t, repo, "some file", fmt.Sprintf("vergo:%s-release", hint))
					} else {
						DoCommitWithMessage(t, repo, "some file", fmt.Sprintf("vergo:%s:%s-release", prefix, hint))
					}

					cmd, buffer := makeBumpFunc(t, bump.Bump)
					cmd.SetArgs([]string{"bump", "auto", "--repository-location", tempDir, "-t", prefix})
					err := cmd.Execute()
					assert.Nil(t, err)
					assert.Equal(t, nextVersion[i], readBuffer(t, buffer))
				}
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestBumpShouldWorkWithIncrementsAndPrefix(t *testing.T) {
	for _, increment := range increments {
		t.Run(increment, func(t *testing.T) {
			_, tempDir := PersistentRepository(t)
			cmd, buffer := makeBump(t)
			cmd.SetArgs([]string{"bump", increment, "--repository-location", tempDir, "-t", "some-prefix", "-p"})
			err := cmd.Execute()
			assert.Nil(t, err)
			assert.Equal(t, "some-prefix-0.1.0", readBuffer(t, buffer))
		})
	}
}

func TestBumpShouldWorkWithIncrementsAndSlashPrefix(t *testing.T) {
	for _, increment := range increments {
		t.Run(increment, func(t *testing.T) {
			_, tempDir := PersistentRepository(t)
			cmd, buffer := makeBump(t)
			cmd.SetArgs([]string{"bump", increment, "--repository-location", tempDir, "-t", "prefix/v", "-p"})
			err := cmd.Execute()
			assert.Nil(t, err)
			assert.Equal(t, "prefix/v0.1.0", readBuffer(t, buffer))
		})
	}
}

func TestBumpFailWhenCalledWithUnknownIncrement(t *testing.T) {
	cmd, _ := makeBump(t)
	cmd.SetArgs([]string{"bump", "unknown-increment", "--repository-location", "some-location", "-t", "some-prefix"})
	err := cmd.Execute()
	assert.NotNil(t, err)
	assert.Equal(t, `invalid argument "unknown-increment" for "vergo release"`, err.Error())
}

func TestBumpFailWhenPushTagFails(t *testing.T) {
	cmd, _ := pushTagFail(t)
	_, tempDir := PersistentRepository(t)
	cmd.SetArgs([]string{"bump", "minor", "--repository-location", tempDir, "-t", "some-prefix", "--push-tag"})
	err := cmd.Execute()
	assert.NotNil(t, err)
	assert.Equal(t, `push tag failed`, err.Error())
}

func TestBumpDetectDotGit(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	tempDirWithInnerFolders := tempDir + "/level1/level2/level3"
	assert.Nil(t, os.MkdirAll(tempDirWithInnerFolders, os.ModePerm))
	cmd, buffer := makeBump(t)
	cmd.SetArgs([]string{"bump", "minor", "--repository-location", tempDirWithInnerFolders})
	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Equal(t, "0.1.0", readBuffer(t, buffer))
}
