package localpath

import (
	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/api"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"github.com/rumenvasilev/rvsecret/internal/webserver"
)

func Scan(log *log.Logger) error {
	// load config
	cfg, err := config.Load(api.LocalPath)
	if err != nil {
		return err
	}
	// exclude the .git directory from local scans as it is not handled properly here
	cfg.SkippablePath = util.AppendIfMissing(cfg.SkippablePath, ".git/")

	// create session
	sess, err := core.NewSessionWithConfig(cfg, log)
	if err != nil {
		return err
	}

	// Start webserver
	if cfg.WebServer && !cfg.Silent {
		ws := webserver.New(*cfg, sess.State, log)
		go ws.Start()
	}

	log.Debug("We have these paths: %s", cfg.LocalPaths)

	if cfg.Debug {
		core.PrintDebug(sess)
	}

	// By default we display a header to the user giving basic info about application. This will not be displayed
	// during a silent run which is the default when using this in an automated fashion.
	core.HeaderInfo(*cfg, sess.State.Stats, sess.Out)

	for _, p := range cfg.LocalPaths {
		if util.PathExists(p, log) {
			last := p[len(p)-1:]
			if last == "/" {
				scanDir(p, sess)
			} else {
				doFileScan(p, sess)
			}
		}
	}

	sess.Finish()

	core.SummaryOutput(sess)

	if cfg.WebServer && !cfg.Silent {
		log.Important("Press Ctrl+C to stop web server and exit.")
		select {}
	}

	return nil
}
