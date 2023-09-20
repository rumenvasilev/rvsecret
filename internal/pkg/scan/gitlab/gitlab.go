package gitlab

import (
	"time"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core"
	"github.com/rumenvasilev/rvsecret/internal/core/banner"
	"github.com/rumenvasilev/rvsecret/internal/core/provider"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/webserver"
	"github.com/spf13/viper"
)

type Gitlab struct {
	Cfg *config.Config
	Log *log.Logger
}

func (g Gitlab) Do() error {
	cfg := g.Cfg
	log := g.Log
	// create session
	sess, err := core.NewSessionWithConfig(cfg, log)
	if err != nil {
		return err
	}

	// Start webserver
	if cfg.Global.WebServer && !cfg.Global.Silent {
		ws := webserver.New(*cfg, sess.State, log)
		go ws.Start()
	}

	// By default we display a header to the user giving basic info about application. This will not be displayed
	// during a silent run which is the default when using this in an automated fashion.
	banner.HeaderInfo(cfg.Global, sess.State.Stats.StartedAt.Format(time.RFC3339), log)

	cfg.GitlabAccessToken = viper.GetString("gitlab-api-token")

	sess.Client, err = provider.InitGitClient(sess.Config, log)
	if err != nil {
		return err
	}

	core.GatherTargets(sess)
	core.GatherGitlabRepositories(sess)
	core.AnalyzeRepositories(sess, sess.State.Stats, log)
	sess.Finish()

	core.SummaryOutput(sess)

	if cfg.Global.WebServer && !cfg.Global.Silent {
		log.Important("%s", banner.ASCIIBanner)
		log.Important("Press Ctrl+C to stop web server and exit.")
		select {}
	}
	return nil
}

var _ api.Scanner = (*Gitlab)(nil)
