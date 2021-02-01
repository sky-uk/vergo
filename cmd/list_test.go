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

func makeList() (*cobra.Command, *bytes.Buffer) {
	cmd := RootCmd()
	cmd.AddCommand(ListCmd())
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	return cmd, b
}

//nolint:scopelint,paralleltest
func TestListNoTag(t *testing.T) {
	prefixes := []string{"v", "app-", "apple-"}
	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			_, tempDir := PersistentRepositoryWithDefaultCommit(t)
			{
				cmd, buffer := makeList()
				cmd.SetArgs([]string{"list", "--repository-location", tempDir, "-t", prefix,
					"--log-level", "error", "-p"})
				err := cmd.Execute()
				assert.Nil(t, err)
				out, err := ioutil.ReadAll(buffer)
				assert.Nil(t, err)
				assert.Equal(t, "", string(out))
			}
		})
	}
}

//nolint:funlen,scopelint,paralleltest
func TestList(t *testing.T) {
	prefixes := []string{"v", "app-", "apple-"}
	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r, tempDir := PersistentRepositoryWithDefaultCommit(t)
			{
				cmd, _ := makeBump()
				cmd.SetArgs([]string{"bump", "minor", "--repository-location", tempDir, "-t", prefix,
					"--log-level", "error"})
				err := cmd.Execute()
				assert.Nil(t, err)
			}
			DoCommit(t, r, "jo")
			{
				cmd, _ := makeBump()
				cmd.SetArgs([]string{"bump", "minor", "--repository-location", tempDir, "-t", prefix,
					"--log-level", "error"})
				err := cmd.Execute()
				assert.Nil(t, err)
			}
			t.Run("list-all", func(t *testing.T) {
				{
					cmd, buffer := makeList()
					cmd.SetArgs([]string{"list", "--repository-location", tempDir, "-t", prefix,
						"--log-level", "error", "-p", "--sort-direction", "asc"})
					err := cmd.Execute()
					assert.Nil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Equal(t, fmt.Sprint(prefix, "0.1.0", "\n", prefix, "0.2.0", "\n"), string(out))
				}
				{
					cmd, buffer := makeList()
					cmd.SetArgs([]string{"list", "--repository-location", tempDir, "-t", prefix,
						"--log-level", "error", "-p", "--sort-direction", "desc"})
					err := cmd.Execute()
					assert.Nil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Equal(t, fmt.Sprint(prefix, "0.2.0", "\n", prefix, "0.1.0", "\n"), string(out))
				}
			})
			t.Run("list-1", func(t *testing.T) {
				{
					cmd, buffer := makeList()
					cmd.SetArgs([]string{"list", "--repository-location", tempDir, "-t", prefix,
						"--log-level", "error", "-p", "--sort-direction", "asc", "--max-list-size", "1"})
					err := cmd.Execute()
					assert.Nil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Equal(t, fmt.Sprint(prefix, "0.1.0", "\n"), string(out))
				}
				{
					cmd, buffer := makeList()
					cmd.SetArgs([]string{"list", "--repository-location", tempDir, "-t", prefix,
						"--log-level", "error", "-p", "--sort-direction", "desc", "--max-list-size", "1"})
					err := cmd.Execute()
					assert.Nil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Equal(t, fmt.Sprint(prefix, "0.2.0", "\n"), string(out))
				}
			})

			t.Run("list-defaults", func(t *testing.T) {
				{
					cmd, buffer := makeList()
					cmd.SetArgs([]string{"list", "--repository-location", tempDir, "-t", prefix,
						"--log-level", "error", "-p", "--sort-direction", "desc"})
					err := cmd.Execute()
					assert.Nil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Equal(t, fmt.Sprint(prefix, "0.2.0", "\n", prefix, "0.1.0", "\n"), string(out))
				}
			})
		})
	}
}
