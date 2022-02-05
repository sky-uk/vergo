package cmd_test

import (
	. "github.com/sky-uk/umc-shared/vergo/internal-test"
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
	cmd, buffer := makeCheckFail(t)
	cmd.SetArgs([]string{"check", "release", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error", "-p"})
	err := cmd.Execute()
	assert.NotNil(t, err)
	assert.Equal(t, "Error: skip release hint present\n", readBuffer(t, buffer))
}

func TestCheckRelease2Failures(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	cmd, buffer := makeCheckWithCheckList(t, checkReleaseFail(t), checkReleaseFailForAnotherReason(t))
	cmd.SetArgs([]string{"check", "release", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error", "-p"})
	err := cmd.Execute()
	assert.NotNil(t, err)
	assert.Equal(t, `Error: skip release hint present
it does not look right
`, readBuffer(t, buffer))
}

func TestCheckReleaseSuccessfulAndFailingCheck(t *testing.T) {
	_, tempDir := PersistentRepository(t)
	cmd, buffer := makeCheckWithCheckList(t, checkReleaseFail(t), checkReleaseSuccess(t))
	cmd.SetArgs([]string{"check", "release", "--repository-location", tempDir, "-t", "some-prefix", "--log-level", "error", "-p"})
	err := cmd.Execute()
	assert.NotNil(t, err)
	assert.Equal(t, "Error: skip release hint present\n", readBuffer(t, buffer))
}
