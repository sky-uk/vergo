package cmd

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	vergo "github.com/sky-uk/umc-shared/vergo/git"
	"github.com/sky-uk/umc-shared/vergo/release"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
)

func ExactValidArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(n)(cmd, args); err != nil {
			return err
		}
		return OnlyValidArgsAndAliases(cmd, args)
	}
}

// OnlyValidArgsAndAliases returns an error if any args are not in the list of ValidArgs.
func OnlyValidArgsAndAliases(cmd *cobra.Command, args []string) error {
	if err := cobra.OnlyValidArgs(cmd, args); err != nil {
		if len(cmd.ArgAliases) > 0 {
			for _, v := range args {
				if !funk.ContainsString(cmd.ArgAliases, v) {
					return err
				}
			}
		} else {
			return err
		}
	}

	return nil
}

type RefFunc func(repo *git.Repository, prefix string) (vergo.SemverRef, error)

func GetCmd(latest, previous RefFunc, current vergo.CurrentVersionFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:        "get (latest-release|previous-release|current-version)",
		Short:      "gets the latest release or current version",
		Args:       ExactValidArgs(1),
		ValidArgs:  []string{"latest-release", "previous-release", "current-version"},
		ArgAliases: []string{"lr", "cv", "pr"},
		RunE: func(cmd *cobra.Command, args []string) error {
			modifier := args[0]
			rootFlags, err := readRootFlags(cmd)
			if err != nil {
				return err
			}
			withMetadata, err := cmd.Flags().GetBool(withMetadata)
			if err != nil {
				return err
			}
			ref, err := get(latest, previous, current, rootFlags, modifier, withMetadata)
			if err != nil {
				return err
			}
			if rootFlags.withPrefix {
				cmd.Print(rootFlags.tagPrefix, ref.Version.String())
			} else {
				cmd.Print(ref.Version.String())
			}
			return nil
		},
	}
	cmd.Flags().BoolP(withMetadata, "m", false, "returns current version with commit hash as metadata")
	return cmd
}

func get(latest, previous RefFunc, current vergo.CurrentVersionFunc, rootFlags *RootFlags, modifier string, withMetadata bool) (vergo.SemverRef, error) {
	repo, err := git.PlainOpenWithOptions(rootFlags.repositoryLocation, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return vergo.EmptyRef, err
	}

	switch modifier {
	case "lr", "latest-release":
		return latest(repo, rootFlags.tagPrefix)
	case "pr", "previous-release":
		return previous(repo, rootFlags.tagPrefix)
	case "cv", "current-version":
		ref, err := current(repo, rootFlags.tagPrefix, release.PreRelease(repo, release.PreReleaseOptions{WithMetadata: withMetadata}))
		if errors.Is(err, plumbing.ErrReferenceNotFound) || errors.Is(err, vergo.ErrNoTagFound) {
			return vergo.SemverRef{Version: semver.MustParse("0.0.0-SNAPSHOT")}, nil
		}
		return ref, err
	default:
		return vergo.EmptyRef, fmt.Errorf("%w : %s", ErrInvalidArg, modifier)
	}
}
