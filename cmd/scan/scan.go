// Package cmd represents the specific commands that the user will execute. Only specific code related to the command
// should be in these files. As much of the code as possible should be pushed to other packages.
package scan

import (
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
	// Global flags under `scan` command
	ScanCmd.PersistentFlags().String("bind-address", "127.0.0.1", "The IP address for the webserver")
	viper.BindPFlag("global.bind-address", ScanCmd.PersistentFlags().Lookup("bind-address")) //nolint:errcheck
	ScanCmd.PersistentFlags().Int("bind-port", 9393, "The port for the webserver")
	viper.BindPFlag("global.bind-port", ScanCmd.PersistentFlags().Lookup("bind-port")) //nolint:errcheck
	ScanCmd.PersistentFlags().Int("confidence-level", 3, "The confidence level of the expressions used to find matches")
	viper.BindPFlag("global.confidence-level", ScanCmd.PersistentFlags().Lookup("confidence-level")) //nolint:errcheck
	ScanCmd.PersistentFlags().Bool("csv", false, "Output csv format")
	viper.BindPFlag("global.csv", ScanCmd.PersistentFlags().Lookup("csv")) //nolint:errcheck
	ScanCmd.PersistentFlags().Bool("json", false, "Output json format")
	viper.BindPFlag("global.json", ScanCmd.PersistentFlags().Lookup("json")) //nolint:errcheck
	ScanCmd.PersistentFlags().Bool("hide-secrets", false, "Do not print secrets to any supported output")
	viper.BindPFlag("global.hide-secrets", ScanCmd.PersistentFlags().Lookup("hide-secrets")) //nolint:errcheck
	ScanCmd.PersistentFlags().StringSlice("ignore-extension", nil, "List of file extensions to ignore")
	viper.BindPFlag("global.ignore-extension", ScanCmd.PersistentFlags().Lookup("ignore-extension")) //nolint:errcheck
	ScanCmd.PersistentFlags().StringSlice("ignore-path", nil, "List of file paths to ignore")
	viper.BindPFlag("global.ignore-path", ScanCmd.PersistentFlags().Lookup("ignore-path")) //nolint:errcheck
	ScanCmd.PersistentFlags().Int("max-file-size", 10, "Max file size to scan (in MB)")
	viper.BindPFlag("global.max-file-size", ScanCmd.PersistentFlags().Lookup("max-file-size")) //nolint:errcheck
	ScanCmd.PersistentFlags().Int("num-threads", -1, "Number of execution threads")
	viper.BindPFlag("global.num-threads", ScanCmd.PersistentFlags().Lookup("num-threads")) //nolint:errcheck
	ScanCmd.PersistentFlags().Float64("commit-depth", -1, "Set the commit depth to scan")
	viper.BindPFlag("global.commit-depth", ScanCmd.PersistentFlags().Lookup("commit-depth")) //nolint:errcheck
	ScanCmd.PersistentFlags().Bool("scan-tests", false, "Scan suspected test files")
	viper.BindPFlag("global.scan-tests", ScanCmd.PersistentFlags().Lookup("scan-tests")) //nolint:errcheck
	ScanCmd.PersistentFlags().Bool("silent", false, "Suppress all output. An alternative output will need to be configured")
	viper.BindPFlag("global.silent", ScanCmd.PersistentFlags().Lookup("silent")) //nolint:errcheck
	ScanCmd.PersistentFlags().Bool("web-server", false, "Enable the web interface for scan output")
	viper.BindPFlag("global.web-server", ScanCmd.PersistentFlags().Lookup("web-server")) //nolint:errcheck
}
