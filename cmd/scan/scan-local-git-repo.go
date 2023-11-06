// Package cmd represents the specific commands that the user will execute. Only specific code related to the command
// should be in these files. As much of the code as possible should be pushed to other packages.
package scan

import (
	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// scanLocalGitRepoCmd represents the scanLocalGitRepo command
var scanLocalGitRepoCmd = &cobra.Command{
	Use:   "local-git-repo",
	Short: "Scan a git repo on a local machine",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(api.LocalGit)
		if err != nil {
			return err
		}
		return scan.New(cfg).Run()
	},
}

func init() {
	ScanCmd.AddCommand(scanLocalGitRepoCmd)
	scanLocalGitRepoCmd.Flags().StringSliceP("paths", "p", nil, "List of local git repos to scan")
	viper.BindPFlag("local.repos", scanLocalGitRepoCmd.Flags().Lookup("paths")) //nolint:errcheck
}
