package cmd_test

import (
	"github.com/stretchr/testify/assert"
	"os"
	. "sky.uk/vergo/internal-test"
	"testing"
)

func TestList(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	cmd, buffer := makeList(t)
	cmd.SetArgs([]string{"list", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error"})
	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Equal(t, "0.2.0\n0.1.0\n", readBuffer(t, buffer))
}

func TestListWithPrefix(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	cmd, buffer := makeList(t)
	cmd.SetArgs([]string{"list", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error", "-p"})
	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Equal(t, "some-prefix-0.2.0\nsome-prefix-0.1.0\n", readBuffer(t, buffer))
}

func TestListWithSlashPrefix(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	cmd, buffer := makeList(t)
	cmd.SetArgs([]string{"list", "--repository-location", tempDir, "-t", "prefix/", "--log-level", "error", "-p"})
	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Equal(t, "prefix/0.2.0\nprefix/0.1.0\n", readBuffer(t, buffer))
}

func TestListDetectDotGit(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	tempDirWithInnerFolders := tempDir + "/level1/level2/level3"
	assert.Nil(t, os.MkdirAll(tempDirWithInnerFolders, os.ModePerm))
	cmd, buffer := makeList(t)
	cmd.SetArgs([]string{"list", "--repository-location", tempDirWithInnerFolders, "-t", "some-prefix", "--log-level", "error"})
	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Equal(t, "0.2.0\n0.1.0\n", readBuffer(t, buffer))
}
