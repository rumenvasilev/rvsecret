package cmd

import (
	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/pkg/signatures"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	userRepo string = "user-repo"
	url      string = "url"
)

// updateSignaturesCmd represents the updateSignatures command
var updateSignaturesCmd = &cobra.Command{
	Use:   "updateSignatures",
	Short: "Update the signatures to the latest version available",
	Long:  "Update the signatures to the latest version available",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(api.UpdateSignatures)
		if err != nil {
			return err
		}
		return signatures.Update(cfg)
	},
}

func init() {
	rootCmd.AddCommand(updateSignaturesCmd)
	updateSignaturesCmd.Flags().StringP("api-token", "t", "", "API token for github access, see documentation for necessary scope")
	viper.BindPFlag("signatures.api-token", updateSignaturesCmd.Flags().Lookup("api-token")) //nolint:errcheck
	updateSignaturesCmd.Flags().String(userRepo, "", "user/repo where signatures can be found, example: rumenvasilev/rvsecret-signatures")
	viper.BindPFlag("signatures.user-repo", updateSignaturesCmd.Flags().Lookup(userRepo)) //nolint:errcheck
	updateSignaturesCmd.Flags().String(url, config.DefaultConfig.Signatures.URL, "url where the signatures can be found")
	viper.BindPFlag("signatures.url", updateSignaturesCmd.Flags().Lookup(url)) //nolint:errcheck
	updateSignaturesCmd.MarkFlagsMutuallyExclusive(userRepo, url)
	updateSignaturesCmd.Flags().String("signatures-version", config.DefaultConfig.Signatures.Version, "specific version of the signatures to install (latest, v1.2.0)")
	viper.BindPFlag("signatures.version", updateSignaturesCmd.Flags().Lookup("signatures-version")) //nolint:errcheck
	updateSignaturesCmd.Flags().Bool("test", config.DefaultConfig.Signatures.Test, "run any tests associated with the signatures and display the output")
	viper.BindPFlag("signatures.test", updateSignaturesCmd.Flags().Lookup("test")) //nolint:errcheck
}
