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

// scanGithubCmd represents the scanGithub command that will enumerate and scan github.com
var scanGithubCmd = &cobra.Command{
	Use:     "github",
	Aliases: []string{"gh"},
	Short:   "Scan one or more github.com orgs or users for secrets.",
	Run: func(cmd *cobra.Command, args []string) {
		scanType := api.Github
		if cmd.Flags().Changed("enterprise") {
			scanType = api.GithubEnterprise
		}
		log := log.NewLogger(viper.GetBool("debug"), viper.GetBool("silent"))
		err := pkg.Scan(scanType, log)
		if err != nil {
			log.Fatal("%v", err)
		}
	},
}

func init() {
	ScanCmd.AddCommand(scanGithubCmd)
	scanGithubCmd.Flags().Bool("add-org-members", false, "Add members to targets when processing organizations")
	scanGithubCmd.Flags().StringP("github-api-token", "t", "", "API token for github access, see documentation for necessary scope")
	scanGithubCmd.MarkFlagRequired("github-api-token") //nolint:errcheck
	scanGithubCmd.Flags().StringSliceP("github-orgs", "o", nil, "List of github orgs to scan")
	scanGithubCmd.Flags().StringSliceP("github-users", "u", nil, "List of github.com users to scan")
	scanGithubCmd.MarkFlagsMutuallyExclusive("github-orgs", "github-users")
	scanGithubCmd.Flags().StringSliceP("github-repos", "r", nil, "List of github repositories to scan")
	scanGithubCmd.Flags().BoolP("enterprise", "e", false, "Enterprise Github instance")
	scanGithubCmd.Flags().String("github-enterprise-url", "", "Github instance address. Update this if you're using GHE with different address.")
	viper.BindPFlags(scanGithubCmd.Flags()) //nolint:errcheck
}
