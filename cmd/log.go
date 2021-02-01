package cmd

import (
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	vergo "sky.uk/vergo/git"
)

func LogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "logs whether changed files between HEAD and last release for a given prefix",
		Args:  cobra.NoArgs,
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
			err = vergo.LogPrefix(repo, rootFlags.tagPrefix, maxLogIteration)
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().Int(maxLogIteration, 100, "how many commits should be searched for the advise")
	return cmd
}

func init() {
	rootCmd.AddCommand(LogCmd())
}
