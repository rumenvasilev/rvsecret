package gitlab

import (
	"context"
	"time"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core"
	"github.com/rumenvasilev/rvsecret/internal/core/banner"
	"github.com/rumenvasilev/rvsecret/internal/core/provider"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/output"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/session"
	"github.com/rumenvasilev/rvsecret/internal/webserver"
)

type Gitlab struct {
	Cfg *config.Config
}

func (g Gitlab) Run() error {
	cfg := g.Cfg
	log := log.Log
	ctx := context.Background()
	// create session
	sess, err := session.NewWithConfig(cfg)
	if err != nil {
		return err
	}

	// Start webserver
	if cfg.Global.WebServer && !cfg.Global.Silent {
		ws := webserver.New(*cfg, sess.State)
		go ws.Start()
	}

	// By default we display a header to the user giving basic info about application. This will not be displayed
	// during a silent run which is the default when using this in an automated fashion.
	banner.HeaderInfo(cfg.Global, sess.State.Stats.StartedAt.Format(time.RFC3339), len(sess.Signatures))

	sess.Client, err = provider.InitGitClient(sess.Config)
	if err != nil {
		return err
	}

	core.GatherTargets(sess)
	core.GatherRepositories(ctx, sess)
	core.AnalyzeRepositories(sess, sess.State.Stats)
	sess.Finish()

	err = output.Summary(sess.State, sess.Config.Global, sess.SignatureVersion)
	if err != nil {
		return err
	}

	if cfg.Global.WebServer && !cfg.Global.Silent {
		log.Important("%s", banner.ASCIIBanner)
		log.Important("Press Ctrl+C to stop web server and exit.")
		select {}
	}
	return nil
}

var _ api.Scanner = (*Gitlab)(nil)
