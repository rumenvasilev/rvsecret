// Package core represents the core functionality of all commands
package core

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"github.com/rumenvasilev/rvsecret/internal/config"
	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/core/finding"
	providerAPI "github.com/rumenvasilev/rvsecret/internal/core/provider/api"
	"github.com/rumenvasilev/rvsecret/internal/core/signatures"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/stats"
	"github.com/rumenvasilev/rvsecret/internal/util"
)

// Session contains all the necessary values and parameters used during a scan
type Session struct {
	Client           providerAPI.IClient `json:"-"` // Client holds the client for the target git server (github, gitlab)
	Config           *config.Config
	State            *State
	Out              *log.Logger `json:"-"`
	Router           *gin.Engine `json:"-"`
	SignatureVersion string
	GithubUsers      []*_coreapi.Owner
	GithubUserLogins []string
	GithubUserOrgs   []string
	GithubUserRepos  []string
	Organizations    []*github.Organization
	Signatures       []signatures.Signature
}

type State struct {
	*sync.Mutex
	Stats        *stats.Stats
	Findings     []*finding.Finding
	Targets      []*_coreapi.Owner
	Repositories []*_coreapi.Repository
}

// NewSession  is the entry point for starting a new scan session
func NewSession(scanType api.ScanType, log *log.Logger) (*Session, error) {
	cfg, err := config.Load(scanType)
	if err != nil {
		return nil, err
	}
	return NewSessionWithConfig(cfg, log)
}

func NewSessionWithConfig(cfg *config.Config, log *log.Logger) (*Session, error) {
	var session Session
	err := session.Initialize(cfg, log)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Initialize will set the initial values and options used during a scan session
func (s *Session) Initialize(cfg *config.Config, log *log.Logger) error {
	s.Out = log
	s.Config = cfg
	s.State = &State{Mutex: &sync.Mutex{}}
	s.State.Stats = stats.Init()

	s.InitThreads()

	var curSig []signatures.Signature
	var combinedSig []signatures.Signature

	// signaturessss
	// TODO need to catch this error here
	f := cfg.Signatures.File
	f = strings.TrimSpace(f)
	h, err := util.SetHomeDir(f)
	if err != nil {
		return err
	}
	if util.PathExists(h) {
		curSig, s.SignatureVersion, err = signatures.Load(h, cfg.Global.ConfidenceLevel)
		if err != nil {
			return err
		}
		combinedSig = append(combinedSig, curSig...)
	}
	// FIXME: updating global var in another pkg
	// signatures.Signatures = combinedSig
	s.Signatures = combinedSig
	return nil
}

// Finish is called at the end of a scan session and used to generate discrete data points
// for a given scan session including setting the status of a scan to finished.
func (s *Session) Finish() {
	s.State.Stats.FinishedAt = time.Now()
	s.State.Stats.UpdateStatus(_coreapi.StatusFinished)
}

// InitThreads will set the correct number of threads based on the commandline flags
func (s *Session) InitThreads() {
	if s.Config.Global.Threads <= 0 {
		numCPUs := runtime.NumCPU()
		s.Config.Global.Threads = numCPUs
		s.Out.Debug("Setting threads to %d", numCPUs)
	}
	runtime.GOMAXPROCS(s.Config.Global.Threads + 2) // thread count + main + web server
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

// PrintSessionStats will print the performance and sessions stats to stdout at the conclusion of a session scan
func printSessionStats(s *stats.Stats, log *log.Logger, appVersion, signatureVersion string) {
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

// SummaryOutput will spit out the results of the hunt along with performance data
func SummaryOutput(sess *Session) error {
	f := sess.State.GetFindings()
	// alpha sort the findings to make the results idempotent
	if len(f) > 0 {
		sort.Slice(f, func(i, j int) bool {
			return f[i].SecretID < f[j].SecretID
		})
	}

	switch {
	case sess.Config.Global.JSONOutput:
		return writeJSON(f)
	case sess.Config.Global.CSVOutput:
		return writeCSV(f)
	default:
		printSessionStats(sess.State.Stats, sess.Out, sess.Config.Global.AppVersion, sess.SignatureVersion)
		return nil
	}
}

func writeJSON(findings []*finding.Finding) error {
	if len(findings) == 0 {
		fmt.Println("[]")
	}

	b, err := json.MarshalIndent(findings, "", "    ")
	if err != nil {
		return err
	}
	c := string(b)
	if c == "null" {
		fmt.Println("[]")
	} else {
		fmt.Println(c)
	}

	return nil
}

func writeCSV(findings []*finding.Finding) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	err := w.Write(finding.GetFindingsCSVHeader())
	if err != nil {
		return err
	}

	for _, v := range findings {
		err := w.Write(v.ToCSV())
		if err != nil {
			return err
		}
	}

	return nil
}
