// Package cmd represents the specific commands that the user will execute. Only specific code related to the command
// should be in these files. As much of the code as possible should be pushed to other packages.
package scan

import (
	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ScanCmd = &cobra.Command{
		Use:   "scan",
		Short: "Use this command to initiate secrets scan",
		Long:  "Use this command to initiate secrets scan - v" + version.AppVersion(),
	}
)

func init() {
	cobra.OnInitialize(config.SetConfig)
	// Global flags under `scan` command
	ScanCmd.PersistentFlags().String("bind-address", "127.0.0.1", "The IP address for the webserver")
	ScanCmd.PersistentFlags().Int("bind-port", 9393, "The port for the webserver")
	ScanCmd.PersistentFlags().Int("confidence-level", 3, "The confidence level of the expressions used to find matches")
	ScanCmd.PersistentFlags().String("config-file", "$HOME/.rvsecret/config.yaml", "config file")
	ScanCmd.PersistentFlags().Bool("csv", false, "Output csv format")
	ScanCmd.PersistentFlags().Bool("json", false, "Output json format")
	ScanCmd.PersistentFlags().Bool("debug", false, "Print available debugging information to stdout")
	ScanCmd.PersistentFlags().Bool("hide-secrets", false, "Do not print secrets to any supported output")
	ScanCmd.PersistentFlags().StringSlice("ignore-extension", nil, "List of file extensions to ignore")
	ScanCmd.PersistentFlags().StringSlice("ignore-path", nil, "List of file paths to ignore")
	ScanCmd.PersistentFlags().Int("max-file-size", 10, "Max file size to scan (in MB)")
	ScanCmd.PersistentFlags().Int("num-threads", -1, "Number of execution threads")
	ScanCmd.PersistentFlags().Float64("commit-depth", -1, "Set the commit depth to scan")
	ScanCmd.PersistentFlags().Bool("scan-tests", false, "Scan suspected test files")
	ScanCmd.PersistentFlags().String("signature-file", "$HOME/.rvsecret/signatures/default.yaml", "file(s) containing detection signatures.")
	ScanCmd.PersistentFlags().String("signature-path", "$HOME/.rvsecret/signatures", "path containing detection signatures.")
	ScanCmd.PersistentFlags().Bool("silent", false, "Suppress all output. An alternative output will need to be configured")
	ScanCmd.PersistentFlags().Bool("web-server", false, "Enable the web interface for scan output")
	viper.BindPFlags(ScanCmd.PersistentFlags()) //nolint:errcheck
}
