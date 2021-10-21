package cmd

import (
	"github.com/go-git/go-git/v5"
	vergo "github.com/sky-uk/umc-shared/vergo/git"
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
			socket, err := checkAuthSocket(true)
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
			err = vergo.PushTag(repo, socket, ref.Version.String(), rootFlags.tagPrefix, rootFlags.remote, rootFlags.dryRun)
			if err != nil {
				return err
			}
			return nil
		},
	}
}
