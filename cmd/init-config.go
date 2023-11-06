package cmd

import (
	"github.com/rumenvasilev/rvsecret/internal/pkg"
	"github.com/spf13/cobra"
)

var initConfigCmd = &cobra.Command{
	Use:   "init-config",
	Short: "Creates configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pkg.SaveConfig()
	},
}

func init() {
	rootCmd.AddCommand(initConfigCmd)
}
