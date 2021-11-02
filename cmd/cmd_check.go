package cmd

import (
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

type CheckReleaseFunc func(repo *git.Repository, tagPrefixRaw string) error

func CheckCmd(checkRelease CheckReleaseFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "check (release)",
		Short:     "performs validations",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"release"},
		RunE: func(cmd *cobra.Command, args []string) error {
			rootFlags, err := readRootFlags(cmd)
			if err != nil {
				return err
			}
			repo, err := git.PlainOpenWithOptions(rootFlags.repositoryLocation, &git.PlainOpenOptions{DetectDotGit: true})
			if err != nil {
				return err
			}

			return checkRelease(repo, rootFlags.tagPrefixRaw)
		},
	}
	return cmd
}
