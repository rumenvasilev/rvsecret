package provider

import (
	"fmt"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core/provider/api"
	"github.com/rumenvasilev/rvsecret/internal/core/provider/github"
	"github.com/rumenvasilev/rvsecret/internal/core/provider/gitlab"
	scanAPI "github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
)

// InitGitClient will create a new git client of the type given by the input string.
func InitGitClient(cfg *config.Config) (api.IClient, error) {
	switch cfg.Global.ScanType {
	case scanAPI.Github:
		return github.NewClient(cfg.Github.APIToken, "")
	case scanAPI.GithubEnterprise:
		if cfg.Github.GithubEnterpriseURL == "" {
			return nil, fmt.Errorf("github enterprise URL is missing")
		}
		return github.NewClient(cfg.Github.APIToken, cfg.Github.GithubEnterpriseURL)
	case scanAPI.Gitlab:
		// TODO need to add in the bits to parse the url here as well
		// TODO set this to some sort of consistent client, look to github for ideas
		return gitlab.NewClient(cfg.Gitlab.APIToken)
	case scanAPI.UpdateSignatures:
		return github.NewClient(cfg.Signatures.APIToken, "")
	default:
		return nil, fmt.Errorf("unknown scan type provided")
	}
}
