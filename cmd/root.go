package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/vergo/bump"
	vergo "github.com/sky-uk/vergo/git"
	"github.com/sky-uk/vergo/release"
	"github.com/spf13/cobra"
	"os"
)

func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "vergo",
		Short:        "vergo [command]",
		SilenceUsage: true,
	}
	rootCmd.SetOut(os.Stdout)
	rootCmd.PersistentFlags().StringP(remoteName, "r", "origin", "remote name for push")
	rootCmd.PersistentFlags().StringP(tagPrefix, "t", "", "version prefix")
	rootCmd.PersistentFlags().StringP(repositoryLocation, "l", ".", "repository location")
	rootCmd.PersistentFlags().String(logLevel, "Info", "set log level")
	rootCmd.PersistentFlags().BoolP(strictHostChecking, "d", false, "disable strict host checking for git. should only be enabled on ci.")
	rootCmd.PersistentFlags().StringP(tokenEnvVarKey, "k", "GH_TOKEN", "environment variable key to use for lookup when deciding if token based git auth should be used")
	rootCmd.PersistentFlags().Bool(dryRun, false, "dry run")
	rootCmd.PersistentFlags().Bool(firstTagEncountered, false, "use first tag matched in the commit history, default use highest tag")
	rootCmd.PersistentFlags().StringSlice(versionedBranchNames, []string{"master", "main"},
		"names of the main working branches")
	rootCmd.PersistentFlags().BoolP(withPrefix, "p", false, "returns version with prefix")
	return rootCmd
}

type RootFlags struct {
	remote, tagPrefix, tagPrefixRaw, repositoryLocation                string
	tokenEnvVarKey                                                     string
	logLevel                                                           log.Level
	withPrefix, dryRun, firstTagEncountered, disableStrictHostChecking bool
	versionedBranches                                                  []string
}

func readRootFlags(cmd *cobra.Command) (*RootFlags, error) {
	remote, err := cmd.Flags().GetString(remoteName)
	if err != nil {
		return nil, err
	}
	versionedBranches, err := cmd.Flags().GetStringSlice(versionedBranchNames)
	if err != nil {
		return nil, err
	}
	dryRun, err := cmd.Flags().GetBool(dryRun)
	if err != nil {
		return nil, err
	}
	firstTagEncountered, err := cmd.Flags().GetBool(firstTagEncountered)
	if err != nil {
		return nil, err
	}
	withPrefix, err := cmd.Flags().GetBool(withPrefix)
	if err != nil {
		return nil, err
	}
	prefix, err := cmd.Flags().GetString(tagPrefix)
	if err != nil {
		return nil, err
	}
	repositoryLocation, err := cmd.Flags().GetString(repositoryLocation)
	if err != nil {
		return nil, err
	}
	logLevelParam, err := cmd.Flags().GetString(logLevel)
	if err != nil {
		return nil, err
	}
	disableStrictHostChecking, err := cmd.Flags().GetBool(strictHostChecking)
	if err != nil {
		return nil, err
	}
	tokenEnvVarKey, err := cmd.Flags().GetString(tokenEnvVarKey)
	if err != nil {
		return nil, err
	}
	logLevel, err := log.ParseLevel(logLevelParam)
	if err != nil {
		log.WithError(err).Errorln("invalid log level, using INFO instead")
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(logLevel)
	}
	return &RootFlags{
		remote:                    remote,
		versionedBranches:         versionedBranches,
		tagPrefix:                 sanitiseTagPrefix(prefix),
		tagPrefixRaw:              prefix,
		repositoryLocation:        repositoryLocation,
		logLevel:                  logLevel,
		dryRun:                    dryRun,
		firstTagEncountered:       firstTagEncountered,
		withPrefix:                withPrefix,
		disableStrictHostChecking: disableStrictHostChecking,
		tokenEnvVarKey:            tokenEnvVarKey,
	}, nil
}

// Execute executes the root command.
func Execute() error {
	var rootCmd = RootCmd()
	rootCmd.AddCommand(BumpCmd(bump.Bump, vergo.PushTag))
	rootCmd.AddCommand(GetCmd(vergo.LatestRef, vergo.PreviousRef, vergo.CurrentVersion))
	rootCmd.AddCommand(PushCmd())
	rootCmd.AddCommand(ListCmd(vergo.ListRefs))
	rootCmd.AddCommand(CheckCmd(release.SkipHintPresent, release.ValidateHEAD, release.IncrementHint))
	rootCmd.AddCommand(ShowCmd())
	rootCmd.AddCommand(VersionCmd())
	return rootCmd.Execute()
}
