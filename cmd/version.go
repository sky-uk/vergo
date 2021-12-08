package cmd

import (
	"fmt"
	"github.com/thoas/go-funk"

	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals
var (
	version  string
	commit   string
	date     string
	builtBy  string
	snapshot string
)

func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "version",
		Short:     "Print the version number of command",
		Long:      `All software has versions.`,
		ValidArgs: []string{"simple"},
		Args:      cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if snapshot == "true" {
				fmt.Println("This is a SNAPSHOT build")
			}
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
}
