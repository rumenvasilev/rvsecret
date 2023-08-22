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

// scanLocalGitRepoCmd represents the scanLocalGitRepo command
var scanLocalGitRepoCmd = &cobra.Command{
	Use:   "local-git-repo",
	Short: "Scan a git repo on a local machine",
	Run: func(cmd *cobra.Command, args []string) {
		log := log.NewLogger(viper.GetBool("debug"), viper.GetBool("silent"))
		err := pkg.Scan(api.LocalGit, log)
		if err != nil {
			log.Fatal("%v", err)
		}
	},
}

func init() {
	ScanCmd.AddCommand(scanLocalGitRepoCmd)
	// scanLocalGitRepoCmd.Flags().Float64("commit-depth", -1, "Set the commit depth to scan")
	scanLocalGitRepoCmd.Flags().StringSlice("local-repos", nil, "List of local git repos to scan")
	viper.BindPFlags(scanLocalGitRepoCmd.Flags()) //nolint:errcheck
}
