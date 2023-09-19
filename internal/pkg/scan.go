package pkg

import (
	"fmt"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/api"
	"github.com/rumenvasilev/rvsecret/internal/pkg/ghe"
	"github.com/rumenvasilev/rvsecret/internal/pkg/github"
	"github.com/rumenvasilev/rvsecret/internal/pkg/gitlab"
	"github.com/rumenvasilev/rvsecret/internal/pkg/localgit"
	"github.com/rumenvasilev/rvsecret/internal/pkg/localpath"
)

func Scan(cfg *config.Config, log *log.Logger) error {
	switch cfg.Global.ScanType {
	case api.LocalPath:
		return localpath.Scan(cfg, log)
	case api.LocalGit:
		return localgit.Scan(cfg, log)
	case api.Github:
		return github.Scan(cfg, log)
	case api.GithubEnterprise:
		return ghe.Scan(cfg, log)
	case api.Gitlab:
		return gitlab.Scan(cfg, log)
	default:
		return fmt.Errorf("unsupported scan type")
	}
}
