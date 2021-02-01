package cmd_test

import (
	"bytes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	. "sky.uk/vergo/cmd"
	. "sky.uk/vergo/internal"
	"testing"
)

func makeGet() (*cobra.Command, *bytes.Buffer) {
	cmd := RootCmd()
	cmd.AddCommand(GetCmd())
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	return cmd, b
}

//nolint:funlen,scopelint,paralleltest
func TestGet(t *testing.T) {
	prefixes := []string{"v", "app-", "apple-"}
	versions := []struct{ increment, next string }{
		{increment: "patch", next: "0.1.1"},
		{increment: "minor", next: "0.2.0"},
		{increment: "major", next: "1.0.0"},
	}
	for _, prefix := range prefixes {
		for _, version := range versions {
			t.Run(prefix+"-"+version.increment, func(t *testing.T) {
				r, tempDir := PersistentRepositoryWithDefaultCommit(t)
				{
					cmd, buffer := makeBump()
					cmd.SetArgs([]string{"bump", version.increment, "--repository-location", tempDir, "-t", prefix,
						"--log-level", "error"})
					err := cmd.Execute()
					assert.Nil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Equal(t, "0.1.0", string(out))
				}
				{
					cmd, buffer := makeGet()
					cmd.SetArgs([]string{"get", "latest-release", "--repository-location", tempDir, "-t", prefix,
						"--log-level", "error", "-p"})
					err := cmd.Execute()
					assert.Nil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Equal(t, prefix+"0.1.0", string(out))
				}
				DoCommit(t, r, "bar")
				{
					cmd, buffer := makeBump()
					cmd.SetArgs([]string{"bump", version.increment, "--repository-location", tempDir, "-t", prefix,
						"--log-level", "error", "-p"})
					err := cmd.Execute()
					assert.Nil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Equal(t, prefix+version.next, string(out))
				}
				{
					cmd, buffer := makeGet()
					cmd.SetArgs([]string{"get", "previous-release", "--repository-location", tempDir, "-t", prefix,
						"--log-level", "error", "-p"})
					err := cmd.Execute()
					assert.Nil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Equal(t, prefix+"0.1.0", string(out))
				}
				{
					cmd, buffer := makeGet()
					cmd.SetArgs([]string{"get", "latest-release", "--repository-location", tempDir, "-t", prefix,
						"--log-level", "error", "-p"})
					err := cmd.Execute()
					assert.Nil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Equal(t, prefix+version.next, string(out))
				}
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestGetNoTag(t *testing.T) {
	prefixes := []string{"v", "app-", "apple-"}
	args := []string{"latest-release", "previous-release", "current-version", "lr", "pr", "cv"}
	for _, prefix := range prefixes {
		for _, arg := range args {
			t.Run(prefix+"-"+arg, func(t *testing.T) {
				_, tempDir := PersistentRepositoryWithDefaultCommit(t)
				{
					cmd, buffer := makeGet()
					cmd.SetArgs([]string{"get", arg, "--repository-location", tempDir, "-t", prefix, "--log-level", "error", "-p"})
					err := cmd.Execute()
					assert.NotNil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Equal(t, "Error: no tag found\n", string(out))
				}
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestGetOneNoTag(t *testing.T) {
	prefixes := []string{"v", "app-", "apple-"}
	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			_, tempDir := PersistentRepositoryWithDefaultCommit(t)
			{
				cmd, buffer := makeBump()
				cmd.SetArgs([]string{"bump", "minor", "--repository-location", tempDir, "-t", prefix, "--log-level", "error"})
				err := cmd.Execute()
				assert.Nil(t, err)
				out, err := ioutil.ReadAll(buffer)
				assert.Nil(t, err)
				assert.Equal(t, "0.1.0", string(out))
			}
			{
				cmd, buffer := makeGet()
				cmd.SetArgs([]string{"get", "latest-release", "--repository-location", tempDir, "-t", prefix, "--log-level", "error", "-p"}) //nolint
				err := cmd.Execute()
				assert.Nil(t, err)
				out, err := ioutil.ReadAll(buffer)
				assert.Nil(t, err)
				assert.Equal(t, prefix+"0.1.0", string(out))
			}
			{
				cmd, buffer := makeGet()
				cmd.SetArgs([]string{"get", "current-version", "--repository-location", tempDir, "-t", prefix, "--log-level", "error", "-p"}) //nolint
				err := cmd.Execute()
				assert.Nil(t, err)
				out, err := ioutil.ReadAll(buffer)
				assert.Nil(t, err)
				assert.Equal(t, prefix+"0.1.0", string(out))
			}
			{
				cmd, buffer := makeGet()
				cmd.SetArgs([]string{"get", "previous-release", "--repository-location", tempDir, "-t", prefix, "--log-level", "error", "-p"}) //nolint
				err := cmd.Execute()
				assert.NotNil(t, err)
				out, err := ioutil.ReadAll(buffer)
				assert.Nil(t, err)
				assert.Equal(t, "Error: one tag found\n", string(out))
			}
		})
	}
}

//nolint:scopelint,paralleltest
func TestGetCurrentVersion(t *testing.T) {
	prefixes := []string{"v", "app-", "apple-"}
	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r, tempDir := PersistentRepositoryWithDefaultCommit(t)
			{
				cmd, buffer := makeBump()
				cmd.SetArgs([]string{"bump", "minor", "--repository-location", tempDir, "-t", prefix, "--log-level", "error"})
				err := cmd.Execute()
				assert.Nil(t, err)
				out, err := ioutil.ReadAll(buffer)
				assert.Nil(t, err)
				assert.Equal(t, "0.1.0", string(out))
			}
			DoCommit(t, r, "jo")
			head, err := r.Head()
			assert.Nil(t, err)
			{
				cmd, buffer := makeGet()
				cmd.SetArgs([]string{"get", "current-version", "--repository-location", tempDir, "-t", prefix, "--log-level", "error"}) //nolint
				err := cmd.Execute()
				assert.Nil(t, err)
				out, err := ioutil.ReadAll(buffer)
				assert.Nil(t, err)
				assert.Equal(t, "0.2.0-SNAPSHOT", string(out))
			}
			{
				cmd, buffer := makeGet()
				cmd.SetArgs([]string{"get", "current-version", "--repository-location", tempDir, "-t", prefix, "--log-level", "error", "-m"}) //nolint
				err := cmd.Execute()
				assert.Nil(t, err)
				out, err := ioutil.ReadAll(buffer)
				assert.Nil(t, err)
				assert.Equal(t, "0.2.0-SNAPSHOT+"+head.Hash().String()[0:7], string(out))
			}
		})
	}
}
