package cmd

import (
	"github.com/go-git/go-git/v5"
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/umc-shared/vergo/bump"
	vergo "github.com/sky-uk/umc-shared/vergo/git"
	"github.com/sky-uk/umc-shared/vergo/release"
	"github.com/spf13/cobra"
)

func BumpCmd(bump bump.BumpFunc, pushTag vergo.PushTagFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "release (patch|minor|major|auto)",
		Short:     "increments the version numbers",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"patch", "minor", "major", "auto"},
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
			repo, err := git.PlainOpenWithOptions(rootFlags.repositoryLocation, &git.PlainOpenOptions{DetectDotGit: true})
			if err != nil {
				return err
			}
			if err := release.SkipHintPresent(repo, rootFlags.tagPrefixRaw); err != nil {
				return err
			}
			if increment == "auto" {
				if increment, err = release.IncrementHint(repo, rootFlags.tagPrefixRaw); err != nil {
					return err
				}
			}
			version, err := bump(repo, rootFlags.tagPrefix, increment, rootFlags.versionedBranches, rootFlags.dryRun)
			if err != nil {
				return err
			}
			if pushTagParam {
				err = pushTag(repo, version.String(), rootFlags.tagPrefix, rootFlags.remote, rootFlags.dryRun)
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
