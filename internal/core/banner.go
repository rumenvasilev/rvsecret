// Package core represents the core functionality of all commands
package core

import (
	_ "embed"
	"time"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/stats"
	"github.com/rumenvasilev/rvsecret/version"
)

//go:embed resources/banner.txt
var ASCIIBanner string

func HeaderInfo(cfg config.Config, stats *stats.Stats, log *log.Logger) {
	if !cfg.JSONOutput && !cfg.CSVOutput {
		log.Warn("%s", ASCIIBanner)
		log.Important("%s v%s started at %s", version.Name, cfg.AppVersion, stats.StartedAt.Format(time.RFC3339))
		log.Important("Loaded %d signatures.", len(Signatures))
		if cfg.WebServer {
			log.Important("Web interface available at http://%s:%d/public", cfg.BindAddress, cfg.BindPort)
		}
	}
}
