package cmd_test

import (
	"bytes"
	"fmt"
	. "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	. "sky.uk/vergo/cmd"
	. "sky.uk/vergo/internal"
	"testing"
)

func makeBump() (*cobra.Command, *bytes.Buffer) {
	cmd := RootCmd()
	cmd.AddCommand(BumpCmd())
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	return cmd, b
}

//nolint:scopelint,paralleltest
func TestNoTagNoCommit(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	increments := []string{"patch", "minor", "major"}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				_, tempDir := PersistentRepository(t)
				cmd, buffer := makeBump()
				cmd.SetArgs([]string{"bump", increment, "--repository-location", tempDir, "-t", prefix})
				err := cmd.Execute()
				assert.NotNil(t, err)
				out, err := ioutil.ReadAll(buffer)
				assert.Nil(t, err)
				assert.Regexp(t, "Error: reference not found\n", string(out))
			})
		}
	}
}

//nolint:paralleltest
func TestBumpNoTag(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	increments := []string{"patch", "minor", "major"}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				_, tempDir := PersistentRepositoryWithDefaultCommit(t)
				cmd, buffer := makeBump()
				cmd.SetArgs([]string{"bump", increment, "--repository-location", tempDir, "-t", prefix, "--skip-validation-first-version", ""}) //nolint
				err := cmd.Execute()
				assert.NotNil(t, err)
				out, err := ioutil.ReadAll(buffer)
				assert.Nil(t, err)
				assert.Equal(t, "Error: no tag found\n", string(out))
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestCreateFirstTag(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	increments := []string{"patch", "minor", "major"}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				_, tempDir := PersistentRepositoryWithDefaultCommit(t)
				cmd, buffer := makeBump()
				cmd.SetArgs([]string{"bump", increment, "--repository-location", tempDir, "-t", prefix, "--log-level", "error"})
				err := cmd.Execute()
				assert.Nil(t, err)
				out, err := ioutil.ReadAll(buffer)
				assert.Nil(t, err)
				assert.Regexp(t, "0.1.0", string(out))
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestRefOnTheHeadValidation(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	increments := []string{"patch", "minor", "major"}
	scenarios := []struct {
		name           string
		skipValidation bool
		output         string
	}{
		{name: "DontSkipRefOnTheHead", skipValidation: false, output: "ref is on the head"},
		{name: "SkipRefOnTheHead", skipValidation: true, output: "0.1.0"},
	}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			for _, scenario := range scenarios {
				t.Run(prefix+"-"+increment+"-"+scenario.name, func(t *testing.T) {
					_, tempDir := PersistentRepositoryWithDefaultCommit(t)
					{
						cmd, buffer := makeBump()
						cmd.SetArgs([]string{"bump", increment, "--repository-location", tempDir, "-t", prefix})
						err := cmd.Execute()
						assert.Nil(t, err)
						out, err := ioutil.ReadAll(buffer)
						assert.Nil(t, err)
						assert.Regexp(t, "0.1.0", string(out))
					}
					{
						cmd, buffer := makeBump()
						validationFlag := fmt.Sprint("--skip-validation-latest-tag-on-the-head=", scenario.skipValidation)
						cmd.SetArgs([]string{"bump", increment, "--repository-location", tempDir, "-t", prefix, validationFlag})
						err := cmd.Execute()
						if scenario.skipValidation {
							assert.Nil(t, err)
						} else {
							assert.NotNil(t, err)
						}
						out, err := ioutil.ReadAll(buffer)
						assert.Nil(t, err)
						assert.Regexp(t, scenario.output, string(out))
					}
				})
			}
		}
	}
}

//nolint:scopelint,paralleltest
func TestNotOnMainBranch(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	increments := []string{"patch", "minor", "major"}
	for _, prefix := range prefixes {
		for _, increment := range increments {
			t.Run(prefix+"-"+increment, func(t *testing.T) {
				r, tempDir := PersistentRepositoryWithDefaultCommit(t)

				wt, err := r.Worktree()
				assert.Nil(t, err)
				err = wt.Checkout(&CheckoutOptions{Branch: plumbing.NewBranchReferenceName("apple"), Create: true})
				assert.Nil(t, err)

				cmd, buffer := makeBump()
				cmd.SetArgs([]string{"bump", increment, "--repository-location", tempDir, "-t", prefix})
				err = cmd.Execute()
				assert.NotNil(t, err)
				out, err := ioutil.ReadAll(buffer)
				assert.Nil(t, err)
				assert.Regexp(t, "command disabled for branches", string(out))
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestBump(t *testing.T) {
	prefixes := []string{"v", "app-", "apple-"}
	versions := []struct {
		increment, pre, post string
	}{
		{
			increment: "patch",
			pre:       "0.1.0",
			post:      "0.1.1",
		},
		{
			increment: "minor",
			pre:       "0.1.0",
			post:      "0.2.0",
		},
		{
			increment: "major",
			pre:       "0.1.0",
			post:      "1.0.0",
		},
	}
	for _, prefix := range prefixes {
		for _, version := range versions {
			t.Run(prefix+"-"+version.increment, func(t *testing.T) {
				r, tempDir := PersistentRepositoryWithDefaultCommit(t)
				head, _ := r.Head()

				ref, err := r.CreateTag(prefix+version.pre, head.Hash(), nil)
				assert.Nil(t, err)
				assert.NotNil(t, ref)

				DoCommit(t, r, "bar")

				{
					cmd, buffer := makeBump()
					cmd.SetArgs([]string{"bump", version.increment, "--repository-location", tempDir, "-t", prefix,
						"--log-level", "error"})
					err = cmd.Execute()
					assert.Nil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Equal(t, version.post, string(out))
				}
				{
					cmd, buffer := makeBump()
					cmd.SetArgs([]string{"bump", version.increment, "--repository-location", tempDir, "-t", prefix,
						"--log-level", "error", "-p"})
					err = cmd.Execute()
					assert.Nil(t, err)
					out, err := ioutil.ReadAll(buffer)
					assert.Nil(t, err)
					assert.Equal(t, prefix+version.post, string(out))
				}
			})
		}
	}
}
