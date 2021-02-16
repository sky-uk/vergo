package cmd_test

import (
	"github.com/stretchr/testify/assert"
	. "sky.uk/vergo/internal"
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
