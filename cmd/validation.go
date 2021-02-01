package cmd

import (
	"github.com/spf13/cobra"
)

type Validation struct {
	skipLatestTagOnTheHead bool
	firstVersion           string
}

func readValidationFlags(cmd *cobra.Command) (*Validation, error) {
	var defaultValidation = Validation{firstVersion: firstVersion}
	tagOnTheHead, err := cmd.Flags().GetBool(skipValidationLatestTagOnTheHead)
	if err != nil {
		return &defaultValidation, err
	}
	firstVersion, err := cmd.Flags().GetString(skipValidationFirstVersion)
	if err != nil {
		return &defaultValidation, err
	}
	return &Validation{
		skipLatestTagOnTheHead: tagOnTheHead,
		firstVersion:           firstVersion,
	}, nil
}
