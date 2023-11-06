package localgit

import (
	"time"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core"
	"github.com/rumenvasilev/rvsecret/internal/core/banner"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/output"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/session"
	"github.com/rumenvasilev/rvsecret/internal/webserver"
)

type LocalGit struct {
	Cfg *config.Config
}

func (l LocalGit) Run() error {
	cfg := l.Cfg
	log := log.Log
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

	log.Debug("We have these repo paths: %s", cfg.Local.Repos)

	if cfg.Global.Debug {
		log.Debug(config.PrintDebug(sess.SignatureVersion))
	}

	// By default we display a header to the user giving basic info about application. This will not be displayed
	// during a silent run which is the default when using this in an automated fashion.
	banner.HeaderInfo(cfg.Global, sess.State.Stats.StartedAt.Format(time.RFC3339), len(sess.Signatures))

	err = core.GatherLocalRepositories(sess)
	if err != nil {
		return err
	}
	core.AnalyzeRepositories(sess, sess.State.Stats)
	sess.Finish()

	err = output.Summary(sess.State, sess.Config.Global, sess.SignatureVersion)
	if err != nil {
		return err
	}

	if cfg.Global.WebServer && !cfg.Global.Silent {
		log.Important("Press Ctrl+C to stop web server and exit.")
		select {}
	}
	return nil
}

var _ api.Scanner = (*LocalGit)(nil)
