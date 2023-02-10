package cmd_test

import (
	"fmt"
	. "github.com/sky-uk/vergo/internal-test"
	"github.com/sky-uk/vergo/release"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckReleaseSuccess(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	cmd, buffer := makeCheck(t)
	cmd.SetArgs([]string{"check", "release", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error"})
	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Equal(t, "", readBuffer(t, buffer))
}

func TestCheckIncrementHintSuccess(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	cmd, buffer := makeCheck(t)
	cmd.SetArgs([]string{"check", "increment-hint", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error"})
	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Equal(t, "", readBuffer(t, buffer))
}

func TestCheckIncrementHintSkipHint(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	expectedError := fmt.Errorf("commit %s is not on a versioned branch: %s", "blah", "blah")
	cmd, buffer := makeCheckFail(t, release.ErrSkipRelease, expectedError, release.ErrNoIncrement)
	cmd.SetArgs([]string{"check", "increment-hint", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error"})
	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Equal(t, "", readBuffer(t, buffer))
}

func TestCheckReleaseFailure(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	cmd, buffer := makeCheckFail(t, release.ErrSkipRelease, success, success)
	cmd.SetArgs([]string{"check", "release", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error"})
	err := cmd.Execute()
	assert.NotNil(t, err)
	assert.Equal(t, "Error: skip release hint present\n", readBuffer(t, buffer))
}

func TestCheckIncrementHintFailure(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	cmd, buffer := makeCheckFail(t, success, success, release.ErrNoIncrement)
	cmd.SetArgs([]string{"check", "increment-hint", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error"})
	err := cmd.Execute()
	assert.NotNil(t, err)
	assert.Equal(t, "Error: increment hint not present\n", readBuffer(t, buffer))
}

func TestCheckReleaseManyFailures(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	expectedError := fmt.Errorf("commit %s is not on a versioned branch: %s", "blah", "blah")
	cmd, buffer := makeCheckFail(t, release.ErrSkipRelease, expectedError, success)
	cmd.SetArgs([]string{"check", "release", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error"})
	err := cmd.Execute()
	assert.NotNil(t, err)
	assert.Equal(t, `Error: skip release hint present
commit blah is not on a versioned branch: blah
`, readBuffer(t, buffer))
}

func TestCheckReleaseInvalidHeadless(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	expectedError := fmt.Errorf("commit %s is not on a versioned branch: %s", "blah", "blah")
	cmd, buffer := makeCheckFail(t, success, expectedError, success)
	cmd.SetArgs([]string{"check", "release", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error"})
	err := cmd.Execute()
	assert.NotNil(t, err)
	assert.Equal(t, "Error: commit blah is not on a versioned branch: blah\n", readBuffer(t, buffer))
}
