package cmd

import (
	"github.com/go-git/go-git/v5"
	vergo "github.com/sky-uk/vergo/git"
	"github.com/spf13/cobra"
)

func PushCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "push the latest tag to a remote",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rootFlags, err := readRootFlags(cmd)
			if err != nil {
				return err
			}
			repo, err := git.PlainOpenWithOptions(rootFlags.repositoryLocation, &git.PlainOpenOptions{DetectDotGit: true})
			if err != nil {
				return err
			}
			ref, err := vergo.LatestRef(repo, rootFlags.tagPrefix)
			if err != nil {
				return err
			}
			err = vergo.PushTag(repo, ref.Version.String(), rootFlags.tagPrefix, rootFlags.remote, rootFlags.dryRun, rootFlags.disableStrictHostChecking, rootFlags.tokenEnvVarKey)
			if err != nil {
				return err
			}
			return nil
		},
	}
}
