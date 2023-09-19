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
	"github.com/rumenvasilev/rvsecret/internal/core/provider"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/api"
	"github.com/rumenvasilev/rvsecret/internal/stats"
	"github.com/rumenvasilev/rvsecret/internal/util"
)

// Session contains all the necessary values and parameters used during a scan
type Session struct {
	Client           provider.IClient `json:"-"` // Client holds the client for the target git server (github, gitlab)
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
	Signatures       []*Signature
}

type State struct {
	*sync.Mutex
	Stats        *stats.Stats
	Findings     []*Finding
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

	// if !s.Silent && s.WebServer {
	// 	s.InitRouter()
	// }

	var curSig []Signature
	var combinedSig []Signature

	// signaturessss
	// TODO need to catch this error here
	f := cfg.Signatures.File
	f = strings.TrimSpace(f)
	h, err := util.SetHomeDir(f)
	if err != nil {
		return err
	}
	if util.PathExists(h, s.Out) {
		curSig, err = LoadSignatures(h, cfg.Global.ConfidenceLevel, s)
		if err != nil {
			return err
		}
		combinedSig = append(combinedSig, curSig...)
	}
	Signatures = combinedSig
	return nil
}

// Finish is called at the end of a scan session and used to generate discrete data points
// for a given scan session including setting the status of a scan to finished.
func (s *Session) Finish() {
	s.State.Stats.FinishedAt = time.Now()
	s.State.Stats.UpdateStatus(_coreapi.StatusFinished)
}

// AddTarget will add a new target to a session to be scanned during that session
func (s *Session) AddTarget(target *_coreapi.Owner) {
	s.State.Lock()
	defer s.State.Unlock()
	for _, t := range s.State.Targets {
		if *target.ID == *t.ID {
			return
		}
	}
	s.State.Targets = append(s.State.Targets, target)
	s.State.Stats.IncrementTargets()
}

// AddRepository will add a given repository to be scanned to a session. This counts as
// the total number of repos that have been gathered during a session.
func (s *Session) AddRepository(repository *_coreapi.Repository) {
	s.State.Lock()
	defer s.State.Unlock()
	for _, r := range s.State.Repositories {
		if repository.ID == r.ID {
			return
		}
	}
	s.State.Repositories = append(s.State.Repositories, repository)

}

// AddFinding will add a finding that has been discovered during a session to the list of findings
// for that session
func (s *Session) AddFinding(finding *Finding) {
	s.State.Lock()
	defer s.State.Unlock()
	// const MaxStrLen = 100
	s.State.Findings = append(s.State.Findings, finding)
	s.State.Stats.IncrementFindingsTotal()
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
func SummaryOutput(sess *Session) {

	// alpha sort the findings to make the results idempotent
	if len(sess.State.Findings) > 0 {
		sort.Slice(sess.State.Findings, func(i, j int) bool {
			return sess.State.Findings[i].SecretID < sess.State.Findings[j].SecretID
		})
	}

	if sess.Config.Global.JSONOutput {
		if len(sess.State.Findings) > 0 {
			b, err := json.MarshalIndent(sess.State.Findings, "", "    ")
			if err != nil {
				fmt.Println(err)
				return
			}
			c := string(b)
			if c == "null" {
				fmt.Println("[]")
			} else {
				fmt.Println(c)
			}
		} else {
			fmt.Println("[]")
		}
	}

	if sess.Config.Global.CSVOutput {
		w := csv.NewWriter(os.Stdout)
		defer w.Flush()
		header := []string{
			"FilePath",
			"Line Number",
			"Action",
			"Description",
			"SignatureID",
			"Finding List",
			"Repo Owner",
			"Repo Name",
			"Commit Hash",
			"Commit Message",
			"Commit Author",
			"File URL",
			"Secret ID",
			"App Version",
			"Signatures Version",
		}
		err := w.Write(header)
		if err != nil {
			sess.Out.Error(err.Error())
		}

		for _, v := range sess.State.Findings {
			line := []string{
				v.FilePath,
				v.LineNumber,
				v.Action,
				v.Description,
				v.SignatureID,
				v.Content,
				v.RepositoryOwner,
				v.RepositoryName,
				v.CommitHash,
				v.CommitMessage,
				v.CommitAuthor,
				v.FileURL,
				v.SecretID,
				v.AppVersion,
				v.SignatureVersion,
			}
			err := w.Write(line)
			if err != nil {
				sess.Out.Error(err.Error())
			}
		}
	}

	if !sess.Config.Global.JSONOutput && !sess.Config.Global.CSVOutput {
		printSessionStats(sess.State.Stats, sess.Out, sess.Config.Global.AppVersion, sess.SignatureVersion)
	}
}
