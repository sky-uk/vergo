package cmd

import (
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	vergo "sky.uk/vergo/git"
)

func AdviseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "advise release",
		Short:     "advises whether changed files between HEAD and the last release includes a relevant change",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"release"},
		RunE: func(cmd *cobra.Command, args []string) error {
			rootFlags, err := readRootFlags(cmd)
			if err != nil {
				return err
			}
			maxLogIteration, err := cmd.Flags().GetInt(maxLogIteration)
			if err != nil {
				return err
			}
			repo, err := git.PlainOpen(rootFlags.repositoryLocation)
			if err != nil {
				return err
			}
			if err = HeadCommitMessageIgnoreHintPresent(repo, rootFlags.tagPrefixRaw); err != nil {
				return err
			}
			if err = vergo.RelevantChanges(repo, rootFlags.tagPrefix, rootFlags.tagPrefixRaw, maxLogIteration); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().Int(maxLogIteration, 100, "how many commits should be searched for the advise")
	return cmd
}

func init() {
	rootCmd.AddCommand(AdviseCmd())
}
