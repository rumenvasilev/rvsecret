package output

import (
	"sort"
	"time"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core/finding"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/session"
	"github.com/rumenvasilev/rvsecret/internal/stats"
)

// Summary will spit out the results of the hunt along with performance data
func Summary(st *session.State, cfg config.Global, sigVersion string) error {
	f := st.GetFindings()
	// alpha sort the findings to make the results idempotent
	if len(f) > 0 {
		sort.Slice(f, func(i, j int) bool {
			return f[i].SecretID < f[j].SecretID
		})
	}

	switch {
	case cfg.JSONOutput:
		return finding.WriteJSON(f)
	case cfg.CSVOutput:
		return finding.WriteCSV(f)
	default:
		printSessionStats(st.Stats, cfg.AppVersion, sigVersion)
		return nil
	}
}

// printSessionStats will print the performance and sessions stats to stdout at the conclusion of a session scan
func printSessionStats(s *stats.Stats, appVersion, signatureVersion string) {
	log := log.Log
	log.Important("\n--------Results--------")
	log.Important("")
	log.Important("-------Findings------")
	log.Info("Total Findings......: %d", s.Findings)
	log.Important("")
	log.Important("--------Files--------")
	log.Info("Total Files.........: %d", s.FilesTotal)
	log.Info("Files Scanned.......: %d", s.FilesScanned)
	log.Info("Files Ignored.......: %d", s.FilesIgnored)
	log.Info("Files Dirty.........: %d", s.FilesDirty)
	log.Important("")
	log.Important("---------SCM---------")
	log.Info("Repos Found.........: %d", s.RepositoriesTotal)
	log.Info("Repos Cloned........: %d", s.RepositoriesCloned)
	log.Info("Repos Scanned.......: %d", s.RepositoriesScanned)
	log.Info("Commits Total.......: %d", s.CommitsTotal)
	log.Info("Commits Scanned.....: %d", s.CommitsScanned)
	log.Info("Commits Dirty.......: %d", s.CommitsDirty)
	log.Important("")
	log.Important("-------General-------")
	log.Info("App Version.........: %s", appVersion)
	log.Info("Signatures Version..: %s", signatureVersion)
	log.Info("Elapsed Time........: %s", time.Since(s.StartedAt))
	log.Info("")
}

// SaveToFile will save a json representation of the session output to a file
// func (s *Session) SaveToFile(location string) error {
// 	sessionJSON, err := json.Marshal(s)
// 	if err != nil {
// 		return err
// 	}
// 	err = os.WriteFile(location, sessionJSON, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
