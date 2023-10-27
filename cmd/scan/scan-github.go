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

// scanGithubCmd represents the scanGithub command that will enumerate and scan github.com
var scanGithubCmd = &cobra.Command{
	Use:     "github",
	Aliases: []string{"gh"},
	Short:   "Scan one or more github.com orgs or users for secrets.",
	RunE: func(cmd *cobra.Command, args []string) error {
		scanType := api.Github
		if cmd.Flags().Changed("enterprise") {
			scanType = api.GithubEnterprise
		}
		cfg, err := config.Load(scanType)
		if err != nil {
			return err
		}
		log := log.NewLogger(cfg.Global.Debug, cfg.Global.Silent)
		return scan.New(cfg, log).Do()
	},
}

func init() {
	ScanCmd.AddCommand(scanGithubCmd)
	scanGithubCmd.Flags().StringP("api-token", "t", "", "API token for github access, see documentation for necessary scope")
	viper.BindPFlag("github.api-token", scanGithubCmd.Flags().Lookup("api-token")) //nolint:errcheck
	scanGithubCmd.Flags().StringSliceP("orgs", "o", config.DefaultConfig.Github.UserDirtyOrgs, "List of github orgs to scan")
	viper.BindPFlag("github.orgs", scanGithubCmd.Flags().Lookup("orgs")) //nolint:errcheck
	scanGithubCmd.Flags().Bool("expand-orgs", false, "Set to true if you want to discover all the users under given organization and enumerate and thus scan all of the combined reporistories.")
	viper.BindPFlag("global.expand-orgs", scanGithubCmd.Flags().Lookup("expand-orgs")) //nolint:errcheck
	scanGithubCmd.Flags().StringSliceP("users", "u", config.DefaultConfig.Github.UserDirtyNames, "List of github.com users to scan")
	scanGithubCmd.MarkFlagsMutuallyExclusive("orgs", "users")
	viper.BindPFlag("github.users", scanGithubCmd.Flags().Lookup("users")) //nolint:errcheck
	scanGithubCmd.Flags().StringSliceP("repos", "r", config.DefaultConfig.Github.UserDirtyRepos, "List of github repositories to scan")
	viper.BindPFlag("github.repos", scanGithubCmd.Flags().Lookup("repos")) //nolint:errcheck
	scanGithubCmd.Flags().BoolP("enterprise", "e", config.DefaultConfig.Github.Enterprise, "Enterprise Github instance")
	viper.BindPFlag("github.enterprise", scanGithubCmd.Flags().Lookup("enterprise")) //nolint:errcheck
	scanGithubCmd.Flags().String("enterprise-url", config.DefaultConfig.Github.GithubEnterpriseURL, "Github instance address. Update this if you're using GHE with different address.")
	viper.BindPFlag("github.enterprise-url", scanGithubCmd.Flags().Lookup("enterprise-url")) //nolint:errcheck
}
