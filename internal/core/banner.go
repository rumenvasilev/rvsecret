// Package core represents the core functionality of all commands
package core

import (
	_ "embed"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/version"
)

//go:embed resources/banner.txt
var ASCIIBanner string

func HeaderInfo(cfg config.Config, startTime string, log *log.Logger) {
	if !cfg.Global.JSONOutput && !cfg.Global.CSVOutput {
		log.Warn("%s", ASCIIBanner)
		log.Important("%s v%s started at %s", version.Name, cfg.Global.AppVersion, startTime)
		log.Important("Loaded %d signatures.", len(Signatures))
		if cfg.Global.WebServer {
			log.Important("Web interface available at http://%s:%d/public", cfg.Global.BindAddress, cfg.Global.BindPort)
		}
	}
}
