package cmd_test

import (
	"github.com/stretchr/testify/assert"
	. "sky.uk/vergo/internal"
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
