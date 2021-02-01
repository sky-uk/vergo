package cmd

import (
	"github.com/go-git/go-git/v5"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	. "sky.uk/vergo/bump"
	vergo "sky.uk/vergo/git"
)

func BumpCmd() *cobra.Command {
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
			validationFlags, err := readValidationFlags(cmd)
			if err != nil {
				return err
			}
			pushTag, err := cmd.Flags().GetBool(pushTag)
			if err != nil {
				return err
			}
			socket, err := checkAuthSocket(pushTag)
			if err != nil {
				return err
			}
			repo, err := git.PlainOpen(rootFlags.repositoryLocation)
			if err != nil {
				return err
			}
			options := Options{
				VersionedBranches:      rootFlags.versionedBranches,
				FirstVersionIfNoTag:    validationFlags.firstVersion,
				SkipLatestTagOnTheHead: validationFlags.skipLatestTagOnTheHead,
				DryRun:                 rootFlags.dryRun,
			}
			version, err := Bump(repo, rootFlags.tagPrefix, increment, options)
			if err != nil {
				return err
			}
			if pushTag {
				err = vergo.PushTag(repo, socket, version.String(), rootFlags.tagPrefix, rootFlags.remote, rootFlags.dryRun)
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
	cmd.Flags().BoolP(pushTag, "u", false, "push the new tag")
	cmd.Flags().Bool(skipValidationLatestTagOnTheHead, true, "skips tagging if head has the latest tag")
	cmd.Flags().StringP(skipValidationFirstVersion, "f", firstVersion, "first version to be used if no tag found")
	return cmd
}

func init() {
	rootCmd.AddCommand(BumpCmd())
}
