package cmd

import (
	"fmt"
	"github.com/thoas/go-funk"

	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals
var (
	version string
	commit  string
	date    string
	builtBy string
)

func init() {
	var versionCmd = &cobra.Command{
		Use:       "version",
		Short:     "Print the version number of command",
		Long:      `All software has versions.`,
		ValidArgs: []string{"simple"},
		Args:      cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if funk.ContainsString(args, "simple") {
				fmt.Print(version) //nolint
			} else {
				fmt.Printf("version: %s\n", version) //nolint
				fmt.Printf("commit : %s\n", commit)  //nolint
				fmt.Printf("date: %s\n", date)       //nolint
				fmt.Printf("builtBy: %s\n", builtBy) //nolint
			}
		},
	}
	rootCmd.AddCommand(versionCmd)
}
