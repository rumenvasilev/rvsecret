package localpath

import (
	"time"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"github.com/rumenvasilev/rvsecret/internal/webserver"
)

func Scan(cfg *config.Config, log *log.Logger) error {
	// exclude the .git directory from local scans as it is not handled properly here
	cfg.Global.SkippablePath = util.AppendIfMissing(cfg.Global.SkippablePath, ".git/")

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

	log.Debug("We have these paths: %s", cfg.Local.Paths)

	if cfg.Global.Debug {
		log.Debug(config.PrintDebug(sess.SignatureVersion))
	}

	// By default we display a header to the user giving basic info about application. This will not be displayed
	// during a silent run which is the default when using this in an automated fashion.
	core.HeaderInfo(*cfg, sess.State.Stats.StartedAt.Format(time.RFC3339), log)

	for _, p := range cfg.Local.Paths {
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

	if cfg.Global.WebServer && !cfg.Global.Silent {
		log.Important("Press Ctrl+C to stop web server and exit.")
		select {}
	}

	return nil
}
