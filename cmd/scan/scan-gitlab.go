// Package cmd represents the specific commands that the user will execute. Only specific code related to the command
// should be in these files. As much of the code as possible should be pushed to other packages.
package scan

import (
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
	Run: func(cmd *cobra.Command, args []string) {
		log := log.NewLogger(viper.GetBool("debug"), viper.GetBool("silent"))
		err := pkg.Scan(api.Gitlab, log)
		if err != nil {
			log.Fatal("%v", err)
		}
	},
}

func init() {
	ScanCmd.AddCommand(scanGitlabCmd)
	scanGitlabCmd.Flags().Bool("add-org-members", false, "Add members to targets when processing organizations")
	// scanGitlabCmd.Flags().Float64("commit-depth", -1, "Set the commit depth to scan")
	scanGitlabCmd.Flags().StringP("gitlab-api-token", "t", "", "API token for access to gitlab, see doc for necessary scope")
	scanGitlabCmd.Flags().StringSlice("gitlab-projects", nil, "List of Gitlab projects or users to scan")
	viper.BindPFlags(scanGithubCmd.Flags()) //nolint:errcheck
}
