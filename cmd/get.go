package cmd

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5"
	gogit "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
	vergo "sky.uk/vergo/git"
)

func ExactValidArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(n)(cmd, args); err != nil {
			return err
		}
		return OnlyValidArgsAndAliases(cmd, args)
	}
}

// OnlyValidArgs returns an error if any args are not in the list of ValidArgs.
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

type RefFunc func(repo *gogit.Repository, prefix string) (vergo.SemverRef, error)
type CurrentVersionFunc func(repo *gogit.Repository, prefix string, preRelease vergo.PreRelease) (vergo.SemverRef, error)

func GetCmd(latest, previous RefFunc, current CurrentVersionFunc) *cobra.Command {
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

func get(latest, previous RefFunc, current CurrentVersionFunc, rootFlags *RootFlags, modifier string, withMetadata bool) (vergo.SemverRef, error) {
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
		return current(repo, rootFlags.tagPrefix, func(version *semver.Version) (semver.Version, error) {
			pre, err := version.IncMinor().SetPrerelease("SNAPSHOT")
			if err != nil {
				return semver.Version{}, err
			}
			if withMetadata {
				head, err := repo.Head()
				if err != nil {
					return semver.Version{}, err
				}
				return pre.SetMetadata(head.Hash().String()[0:7])
			}
			return pre, nil
		})
	default:
		return vergo.EmptyRef, fmt.Errorf("%w : %s", ErrInvalidArg, modifier)
	}
}
