package pkg

import (
	"fmt"

	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/api"
	"github.com/rumenvasilev/rvsecret/internal/pkg/ghe"
	"github.com/rumenvasilev/rvsecret/internal/pkg/github"
	"github.com/rumenvasilev/rvsecret/internal/pkg/gitlab"
	"github.com/rumenvasilev/rvsecret/internal/pkg/localgit"
	"github.com/rumenvasilev/rvsecret/internal/pkg/localpath"
)

func Scan(scanType api.ScanType, log *log.Logger) error {
	switch scanType {
	case api.LocalPath:
		return localpath.Scan(log)
	case api.LocalGit:
		return localgit.Scan(log)
	case api.Github:
		return github.Scan(log)
	case api.GithubEnterprise:
		return ghe.Scan(log)
	case api.Gitlab:
		return gitlab.Scan(log)
	default:
		return fmt.Errorf("unsupported scan type")
	}
}
