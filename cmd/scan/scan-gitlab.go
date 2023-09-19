// Package cmd represents the specific commands that the user will execute. Only specific code related to the command
// should be in these files. As much of the code as possible should be pushed to other packages.
package scan

import (
	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg"
	"github.com/rumenvasilev/rvsecret/internal/pkg/api"
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
		return pkg.Scan(cfg, log)
	},
}

func init() {
	ScanCmd.AddCommand(scanGitlabCmd)
	// scanGitlabCmd.Flags().Bool("add-org-members", false, "Add members to targets when processing organizations")
	scanGitlabCmd.Flags().StringP("gitlab-api-token", "t", "", "API token for access to gitlab, see doc for necessary scope")
	scanGitlabCmd.Flags().StringSlice("gitlab-projects", config.DefaultConfig.Gitlab.GitlabTargets, "List of Gitlab projects or users to scan")
	viper.BindPFlags(scanGitlabCmd.Flags()) //nolint:errcheck
}
