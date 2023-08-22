// Package cmd represents the specific commands that the user will execute. Only specific code related to the command
// should be in these files. As much of the code as possible should be pushed to other packages.
package cmd

import (
	"fmt"
	"os"

	"github.com/rumenvasilev/rvsecret/cmd/scan"
	"github.com/rumenvasilev/rvsecret/version"

	"github.com/spf13/cobra"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "rvsecret",
		Short: "A tool to scan for secrets in various digital hiding spots",
		Long:  "A tool to scan for secrets in various digital hiding spots - v" + version.AppVersion(), // TODO write a better long description
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
}
