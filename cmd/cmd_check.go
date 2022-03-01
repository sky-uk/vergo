package cmd

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/sky-uk/umc-shared/vergo/release"
	"github.com/spf13/cobra"
)

func CheckCmd(
	skipHintPresent release.SkipHintPresentFunc,
	validateHEAD release.ValidateHEADFunc,
	incrementHint release.IncrementHintFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "performs validations",
	}
	cmd.AddCommand(checkReleaseCmd(skipHintPresent, validateHEAD))
	cmd.AddCommand(checkIncrementHintCmd(skipHintPresent, incrementHint))
	return cmd
}

func checkReleaseCmd(skipHintPresent release.SkipHintPresentFunc, validateHEAD release.ValidateHEADFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "performs release validations",
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
			if err := validateHEAD(repo, rootFlags.remote, rootFlags.versionedBranches); err != nil {
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

func checkIncrementHintCmd(skipHintPresent release.SkipHintPresentFunc, incrementHint release.IncrementHintFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "increment-hint",
		Short: "checks if commit message includes certain keywords",
		RunE: func(cmd *cobra.Command, args []string) error {
			rootFlags, err := readRootFlags(cmd)
			if err != nil {
				return err
			}
			repo, err := git.PlainOpenWithOptions(rootFlags.repositoryLocation, &git.PlainOpenOptions{DetectDotGit: true})
			if err != nil {
				return err
			}

			err = skipHintPresent(repo, rootFlags.tagPrefixRaw)
			if errors.Is(err, release.ErrSkipRelease) {
				return nil
			}
			if err != nil {
				return err
			}
			_, err = incrementHint(repo, rootFlags.tagPrefixRaw)
			return err
		},
	}
	return cmd
}
