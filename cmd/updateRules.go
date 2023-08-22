package cmd

import (
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/signatures"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// updateSignaturesCmd represents the updateSignatures command
var updateSignaturesCmd = &cobra.Command{
	Use:   "updateSignatures",
	Short: "Update the signatures to the latest version available",
	Long:  "Update the signatures to the latest version available",
	Run: func(cmd *cobra.Command, args []string) {
		log := log.NewLogger(viper.GetBool("debug"), viper.GetBool("silent"))
		err := signatures.Update(log)
		if err != nil {
			log.Fatal("%v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateSignaturesCmd)

	updateSignaturesCmd.Flags().String("signatures-path", "$HOME/.rvsecret/signatures/", "path where the signatures will be installed")
	updateSignaturesCmd.Flags().String("signatures-url", "https://github.com/rumenvasilev/rvsecret-signatures", "url where the signatures can be found")
	updateSignaturesCmd.Flags().String("signatures-version", "", "specific version of the signatures to install")
	updateSignaturesCmd.Flags().Bool("test-signatures", false, "run any tests associated with the signatures and display the output")
	viper.BindPFlags(updateSignaturesCmd.Flags()) //nolint:errcheck
}
