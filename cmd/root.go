// Package cmd represents the specific commands that the user will execute. Only specific code related to the command
// should be in these files. As much of the code as possible should be pushed to other packages.
package cmd

import (
	"fmt"
	"os"

	"github.com/rumenvasilev/rvsecret/cmd/scan"
	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	signaturesFile string = "signatures-file"
	signaturesPath string = "signatures-path"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "rvsecret",
		Short: "A tool to scan for secrets in various digital hiding spots",
		Long:  "A tool to scan for secrets in various digital hiding spots - v" + version.AppVersion(),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			config.SetConfig(cmd)
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(scan.ScanCmd)
	rootCmd.PersistentFlags().Bool("debug", false, "Print available debugging information to stdout")
	viper.BindPFlag("global.debug", rootCmd.PersistentFlags().Lookup("debug")) //nolint:errcheck
	rootCmd.PersistentFlags().String("config-file", config.DefaultConfig.Global.ConfigFile, "Config file location")
	viper.BindPFlag("global.config-file", rootCmd.PersistentFlags().Lookup("config-file")) //nolint:errcheck
	rootCmd.PersistentFlags().String(signaturesFile, config.DefaultConfig.Signatures.File, "file(s) containing detection signatures.")
	viper.BindPFlag("signatures.file", rootCmd.PersistentFlags().Lookup(signaturesFile)) //nolint:errcheck
	rootCmd.PersistentFlags().String(signaturesPath, config.DefaultConfig.Signatures.Path, "path containing detection signatures.")
	rootCmd.MarkFlagsMutuallyExclusive(signaturesFile, signaturesPath)
	viper.BindPFlag("signatures.path", rootCmd.PersistentFlags().Lookup(signaturesPath)) //nolint:errcheck
}
