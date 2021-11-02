package cmd

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
	"regexp"
)

// ShowCmd is in incubation
func ShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "commands for querying",
		Long:  "These commands are in incubation",
	}

	cmd.AddCommand(parentsCmd())
	return cmd
}

func parentsCmd() *cobra.Command {
	mergeCommitMessage := regexp.MustCompile(`Merge pull request #\d+ from.*`)
	cmd := &cobra.Command{
		Use:   "parents",
		Short: "lists parents of a commit",
		Args:  cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootFlags, err := readRootFlags(cmd)
			if err != nil {
				return err
			}
			mergeCommits, err := cmd.Flags().GetBool(mergeCommits)
			if err != nil {
				return err
			}
			repo, err := git.PlainOpenWithOptions(rootFlags.repositoryLocation, &git.PlainOpenOptions{DetectDotGit: true})
			if err != nil {
				return err
			}
			commit, err := repo.CommitObject(plumbing.NewHash(args[0]))
			if err != nil {
				return err
			}
			for _, parentHash := range commit.ParentHashes {
				parentCommit, err := repo.CommitObject(parentHash)
				if err != nil {
					return err
				}
				if !mergeCommitMessage.MatchString(parentCommit.Message) || mergeCommits {
					cmd.Println(parentCommit.Hash.String(), " ", parentCommit.Message)
				}

			}
			return nil
		},
	}
	cmd.Flags().Bool(mergeCommits, false, "include merge commits when listing parents")
	return cmd
}
