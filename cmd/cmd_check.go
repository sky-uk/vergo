package cmd

import (
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

type SkipHintPresentFunc func(repo *git.Repository, tagPrefixRaw string) error
type ValidateHEADFunc func(repo *git.Repository, versionedBranches []string) error

func CheckCmd(skipHintPresent SkipHintPresentFunc, validateHEAD ValidateHEADFunc) *cobra.Command {
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

			var errs errs
			if err := skipHintPresent(repo, rootFlags.tagPrefixRaw); err != nil {
				errs = append(errs, err)
			}
			if err := validateHEAD(repo, rootFlags.versionedBranches); err != nil {
				errs = append(errs, err)
			}
			if len(errs) > 0 {
				return errs
			}
			return nil
		},
	}
	return cmd
}
