package cmd

import (
	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	gogit "github.com/go-git/go-git/v5"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type BumpFunc func(
	repo *gogit.Repository,
	tagPrefix, increment string,
	versionedBranches []string,
	dryRun bool) (*semver.Version, error)

type PushTagFunc func(
	repo *gogit.Repository,
	socket, version, prefix, remote string,
	dryRun bool) error

func BumpCmd(bump BumpFunc, pushTag PushTagFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "release (patch|minor|major)",
		Short:     "increments the version numbers",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"patch", "minor", "major"},
		Aliases:   []string{"bump"},
		RunE: func(cmd *cobra.Command, args []string) error {
			increment := args[0]
			rootFlags, err := readRootFlags(cmd)
			if err != nil {
				return err
			}
			pushTagParam, err := cmd.Flags().GetBool(pushTagParam)
			if err != nil {
				return err
			}
			socket, err := checkAuthSocket(pushTagParam)
			if err != nil {
				return err
			}
			repo, err := git.PlainOpenWithOptions(rootFlags.repositoryLocation, &git.PlainOpenOptions{DetectDotGit: true})
			if err != nil {
				return err
			}
			version, err := bump(repo, rootFlags.tagPrefix, increment, rootFlags.versionedBranches, rootFlags.dryRun)
			if err != nil {
				return err
			}
			if pushTagParam {
				err = pushTag(repo, socket, version.String(), rootFlags.tagPrefix, rootFlags.remote, rootFlags.dryRun)
				if err != nil {
					return err
				}
			} else {
				log.Trace("Push not enabled")
			}
			if rootFlags.withPrefix {
				cmd.Print(rootFlags.tagPrefix, version.String())
			} else {
				cmd.Print(version.String())
			}
			return nil
		},
	}
	cmd.Flags().BoolP(pushTagParam, "u", false, "push the new tag")
	return cmd
}
