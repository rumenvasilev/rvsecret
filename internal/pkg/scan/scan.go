package scan

import (
	"errors"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/github"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/gitlab"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/localgit"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/localpath"
)

func New(cfg *config.Config, log *log.Logger) api.Scanner {
	switch cfg.Global.ScanType {
	case api.LocalPath:
		return localpath.Localpath{Cfg: cfg, Log: log}
	case api.LocalGit:
		return localgit.LocalGit{Cfg: cfg, Log: log}
	case api.GithubEnterprise, api.Github:
		return github.Github{Cfg: cfg, Log: log}
	case api.Gitlab:
		return gitlab.Gitlab{Cfg: cfg, Log: log}
	default:
		return Unsupported{}
	}
}

type Unsupported struct{}

func (u Unsupported) Do() error {
	return errors.New("this type of scan is unsupported")
}
