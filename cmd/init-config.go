package cmd

import (
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initConfigCmd = &cobra.Command{
	Use:   "init-config",
	Short: "Creates configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := log.NewLogger(viper.GetBool("debug"), viper.GetBool("silent"))
		return pkg.SaveConfig(log)
	},
}

func init() {
	rootCmd.AddCommand(initConfigCmd)
}
