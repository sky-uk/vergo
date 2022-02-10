package cmd_test

import (
	. "github.com/sky-uk/umc-shared/vergo/internal-test"
	"github.com/sky-uk/umc-shared/vergo/release"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckReleaseSuccess(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	cmd, buffer := makeCheck(t)
	cmd.SetArgs([]string{"check", "release", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error", "-p"})
	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Equal(t, "", readBuffer(t, buffer))
}

func TestCheckReleaseFailure(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	cmd, buffer := makeCheckFail(t, release.ErrSkipRelease, success)
	cmd.SetArgs([]string{"check", "release", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error", "-p"})
	err := cmd.Execute()
	assert.NotNil(t, err)
	assert.Equal(t, "Error: skip release hint present\n", readBuffer(t, buffer))
}

func TestCheckRelease2Failures(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	cmd, buffer := makeCheckFail(t, release.ErrSkipRelease, release.ErrInvalidHeadless)
	cmd.SetArgs([]string{"check", "release", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error", "-p"})
	err := cmd.Execute()
	assert.NotNil(t, err)
	assert.Equal(t, `Error: skip release hint present
invalid headless checkout
`, readBuffer(t, buffer))
}

func TestCheckReleaseInvalidHeadless(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	cmd, buffer := makeCheckFail(t, success, release.ErrInvalidHeadless)
	cmd.SetArgs([]string{"check", "release", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error", "-p"})
	err := cmd.Execute()
	assert.NotNil(t, err)
	assert.Equal(t, "Error: invalid headless checkout\n", readBuffer(t, buffer))
}
