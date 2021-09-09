package cmd_test

import (
	"github.com/stretchr/testify/assert"
	"os"
	. "sky.uk/vergo/internal-test"
	"testing"
)

//nolint:scopelint,paralleltest
func TestBumpShouldWorkWithIncrements(t *testing.T) {
	increments := []string{"patch", "minor", "major"}
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
func TestBumpShouldWorkWithIncrementsAndPrefix(t *testing.T) {
	increments := []string{"patch", "minor", "major"}
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
	increments := []string{"patch", "minor", "major"}
	for _, increment := range increments {
		t.Run(increment, func(t *testing.T) {
			_, tempDir := PersistentRepository(t)
			cmd, buffer := makeBump(t)
			cmd.SetArgs([]string{"bump", increment, "--repository-location", tempDir, "-t", "prefix/", "-p"})
			err := cmd.Execute()
			assert.Nil(t, err)
			assert.Equal(t, "prefix/0.1.0", readBuffer(t, buffer))
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
