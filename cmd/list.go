package cmd

import (
	"github.com/go-git/go-git/v5"
	gogit "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	vergo "sky.uk/vergo/git"
)

type ListRefs func(
	repo *gogit.Repository,
	prefix string,
	direction vergo.SortDirection,
	maxListSize int) ([]vergo.SemverRef, error)

func ListCmd(listRefs ListRefs) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists the tags",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rootFlags, err := readRootFlags(cmd)
			if err != nil {
				return err
			}
			sortDirectionString, err := cmd.Flags().GetString(sortDirection)
			if err != nil {
				return err
			}
			direction, err := vergo.ParseSortDirection(sortDirectionString)
			if err != nil {
				return err
			}
			maxListSize, err := cmd.Flags().GetInt(maxListSize)
			if err != nil {
				return err
			}
			repo, err := git.PlainOpenWithOptions(rootFlags.repositoryLocation, &git.PlainOpenOptions{DetectDotGit: true})
			if err != nil {
				return err
			}
			refs, err := listRefs(repo, rootFlags.tagPrefix, direction, maxListSize)
			if err != nil {
				return err
			}
			for _, ref := range refs {
				if rootFlags.withPrefix {
					cmd.Print(rootFlags.tagPrefix, ref.Version.String())
					cmd.Println()
				} else {
					cmd.Print(ref.Version.String())
					cmd.Println()
				}
			}
			return nil
		},
	}
	cmd.Flags().String(sortDirection, "desc", "sort direction [asc,desc]")
	cmd.Flags().Int(maxListSize, 10, "maximum size of the list returned")
	return cmd
}
