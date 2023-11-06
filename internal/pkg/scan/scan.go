package scan

import (
	"errors"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/github"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/gitlab"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/localgit"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/localpath"
)

func New(cfg *config.Config) api.Scanner {
	switch cfg.Global.ScanType {
	case api.LocalPath:
		return localpath.Localpath{Cfg: cfg}
	case api.LocalGit:
		return localgit.LocalGit{Cfg: cfg}
	case api.GithubEnterprise, api.Github:
		return github.Github{Cfg: cfg}
	case api.Gitlab:
		return gitlab.Gitlab{Cfg: cfg}
	default:
		return Unsupported{}
	}
}

type Unsupported struct{}

func (u Unsupported) Run() error {
	return errors.New("this type of scan is unsupported")
}
