package cmd_test

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	. "sky.uk/vergo/cmd"
	. "sky.uk/vergo/internal"
	"testing"
)

func makeCheck() (*cobra.Command, *bytes.Buffer) {
	cmd := RootCmd()
	cmd.AddCommand(BumpCmd())
	cmd.AddCommand(AdviseCmd())
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	return cmd, b
}

//nolint:scopelint,paralleltest
func TestCheckIgnoreHint(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	formats := []string{"vergo:%s:ignore"}
	increments := []string{"patch", "minor", "major"}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			for _, format := range formats {
				t.Run(prefix+"-"+increment, func(t *testing.T) {
					r, tempDir := PersistentRepositoryWithDefaultCommit(t)
					DoCommitWithMessage(t, r, "bar", fmt.Sprintf(format, prefix))
					cmd, buffer := makeCheck()
					cmd.SetArgs([]string{"advise", "release", "--repository-location", tempDir, "-t", prefix})
					err := cmd.Execute()
					assert.NotNil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Regexp(t, "ignore hint", string(out))
				})
			}
		}
	}
}

//nolint:scopelint,paralleltest
func TestCheckNoRelevantChanges(t *testing.T) {
	prefixes := []string{"app", "apple"}
	increments := []string{"patch", "minor", "major"}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r, tempDir := PersistentRepositoryWithDefaultCommit(t)
				{
					cmd, buffer := makeBump()
					cmd.SetArgs([]string{"bump", increment, "--repository-location", tempDir, "-t", prefix})
					err := cmd.Execute()
					assert.Nil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Regexp(t, "0.1.0", string(out))
				}
				DoCommit(t, r, "bar")
				cmd, buffer := makeCheck()
				cmd.SetArgs([]string{"advise", "release", "--repository-location", tempDir, "-t", prefix})
				err := cmd.Execute()
				assert.NotNil(t, err)
				out, err := ioutil.ReadAll(buffer)
				assert.Nil(t, err)
				assert.Regexp(t, "no relevant change", string(out))
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestCheckRelevantChanges(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	increments := []string{"patch", "minor", "major"}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r, tempDir := PersistentRepositoryWithDefaultCommit(t)
				{
					cmd, buffer := makeBump()
					cmd.SetArgs([]string{"bump", increment, "--repository-location", tempDir, "-t", prefix})
					err := cmd.Execute()
					assert.Nil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Regexp(t, "0.1.0", string(out))
				}
				DoCommit(t, r, prefix+".txt")
				cmd, _ := makeCheck()
				cmd.SetArgs([]string{"advise", "release", "--repository-location", tempDir, "-t", prefix})
				err := cmd.Execute()
				assert.Nil(t, err)
			})
		}
	}
}
