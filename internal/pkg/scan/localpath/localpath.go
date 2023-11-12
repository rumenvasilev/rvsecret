package localpath

import (
	"context"
	"time"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core"
	coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/core/banner"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/output"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/session"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"github.com/rumenvasilev/rvsecret/internal/webserver"
)

type Localpath struct {
	Cfg *config.Config
}

func (l Localpath) Run() error {
	cfg := l.Cfg
	log := log.Log
	ctx := context.Background()
	// exclude the .git directory from local scans as it is not handled properly here
	cfg.Global.SkippablePath = util.AppendIfMissing(cfg.Global.SkippablePath, ".git/")

	// create session
	sess, err := session.NewWithConfig(cfg)
	if err != nil {
		return err
	}

	// Start webserver
	if cfg.Global.WebServer && !cfg.Global.Silent {
		ws := webserver.New(ctx, *cfg, sess.State)
		go ws.Start()
	}

	log.Debug("We have these paths: %s", cfg.Local.Paths)

	if cfg.Global.Debug {
		log.Debug(config.PrintDebug(sess.SignatureVersion))
	}

	// By default we display a header to the user giving basic info about application. This will not be displayed
	// during a silent run which is the default when using this in an automated fashion.
	banner.HeaderInfo(cfg.Global, sess.State.Stats.StartedAt.Format(time.RFC3339), len(sess.Signatures))
	ctxworker := context.WithValue(ctx, core.TID, 0)
	for _, p := range cfg.Local.Paths {
		if util.PathExists(p) {
			last := p[len(p)-1:]
			if last == "/" {
				scanDir(p, sess)
			} else {
				core.AnalyzeObject(ctxworker, sess, nil, nil, p, coreapi.Repository{})
			}
		}
	}

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

var _ api.Scanner = (*Localpath)(nil)
