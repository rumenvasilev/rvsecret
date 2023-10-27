// Package cmd represents the specific commands that the user will execute. Only specific code related to the command
// should be in these files. As much of the code as possible should be pushed to other packages.
package scan

import (
	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// scanGitlabCmd represents the scanGitlab command
var scanGitlabCmd = &cobra.Command{
	Use:     "gitlab",
	Aliases: []string{"gl"},
	Short:   "Scan one or more gitlab groups or users for secrets",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(api.Gitlab)
		if err != nil {
			return err
		}
		log := log.NewLogger(cfg.Global.Debug, cfg.Global.Silent)
		return scan.New(cfg, log).Do()
	},
}

func init() {
	ScanCmd.AddCommand(scanGitlabCmd)
	scanGitlabCmd.Flags().StringP("api-token", "t", "", "API token for access to gitlab, see doc for necessary scope")
	viper.BindPFlag("gitlab.api-token", scanGitlabCmd.Flags().Lookup("api-token")) //nolint:errcheck
	scanGitlabCmd.Flags().StringSlice("projects", config.DefaultConfig.Gitlab.Targets, "List of Gitlab projects or users to scan")
	viper.BindPFlag("gitlab.projects", scanGitlabCmd.Flags().Lookup("projects")) //nolint:errcheck
}
