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

// scanLocalPathCmd represents the scanLocalFiles command
var scanLocalPathCmd = &cobra.Command{
	TraverseChildren: true,
	Use:              "localpath",
	Short:            "Scan local files and directories",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(api.LocalPath)
		if err != nil {
			return err
		}
		log := log.NewLogger(cfg.Global.Debug, cfg.Global.Silent)
		return pkg.Scan(cfg, log)
	},
}

func init() {
	ScanCmd.AddCommand(scanLocalPathCmd)
	scanLocalPathCmd.Flags().StringSliceP("paths", "p", nil, "List of local paths to scan")
	viper.BindPFlag("local.paths", scanLocalPathCmd.Flags().Lookup("paths")) //nolint:errcheck
}
