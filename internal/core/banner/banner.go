// Package core represents the core functionality of all commands
package banner

import (
	_ "embed"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core/signatures"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/version"
)

//go:embed banner.txt
var ASCIIBanner string

func HeaderInfo(cfg config.Global, startTime string, log *log.Logger) {
	if !cfg.JSONOutput && !cfg.CSVOutput {
		log.Warn("%s", ASCIIBanner)
		log.Important("%s v%s started at %s", version.Name, cfg.AppVersion, startTime)
		log.Important("Loaded %d signatures.", len(signatures.Signatures))
		if cfg.WebServer {
			log.Important("Web interface available at http://%s:%d/public", cfg.BindAddress, cfg.BindPort)
		}
	}
}
