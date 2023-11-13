package session

import (
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"github.com/rumenvasilev/rvsecret/internal/config"
	coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/core/finding"
	providerapi "github.com/rumenvasilev/rvsecret/internal/core/provider/api"
	"github.com/rumenvasilev/rvsecret/internal/core/signatures"
	"github.com/rumenvasilev/rvsecret/internal/log"
	scanapi "github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/stats"
)

// Session contains all the necessary values and parameters used during a scan
type Session struct {
	Client           providerapi.IClient `json:"-"` // Client holds the client for the target git server (github, gitlab)
	Config           *config.Config
	State            *State
	Router           *gin.Engine `json:"-"`
	SignatureVersion string
	GithubUsers      []*coreapi.Owner
	GithubUserLogins []string
	GithubUserOrgs   []string
	GithubUserRepos  []string
	Organizations    []*github.Organization
	Signatures       []signatures.Signature
}

// NewSession is the entry point for starting a new scan session
func New(scanType scanapi.ScanType) (*Session, error) {
	cfg, err := config.Load(scanType)
	if err != nil {
		return nil, err
	}
	return NewWithConfig(cfg)
}

func NewWithConfig(cfg *config.Config) (*Session, error) {
	return initialize(cfg)
}

// Initialize will set the initial values and options used during a scan session
func initialize(cfg *config.Config) (*Session, error) {
	s := new(Session).withConfig(cfg)

	// init state
	s.State = &State{Mutex: &sync.Mutex{}, Findings: make(map[string]*finding.Finding)}

	// init threads
	s.initThreads()

	var err error
	s.Signatures, s.SignatureVersion, err = signatures.Load(cfg.Signatures.File, cfg.Global.ConfidenceLevel)

	return s.start(), err
}

func (s *Session) withConfig(cfg *config.Config) *Session {
	s.Config = cfg
	return s
}

// start will init the statistics, indicating the time we've started the scan
func (s *Session) start() *Session {
	s.State.Stats = stats.Init()
	return s
}

// Finish is called at the end of a scan session and used to generate discrete data points
// for a given scan session including setting the status of a scan to finished.
func (s *Session) Finish() {
	s.State.Stats.FinishedAt = time.Now()
	s.State.Stats.UpdateStatus(stats.StatusFinished)
}

// InitThreads will set the correct number of threads based on the commandline flags
func (s *Session) initThreads() {
	if s.Config.Global.Threads <= 0 {
		numCPUs := runtime.NumCPU()
		s.Config.Global.Threads = numCPUs
		log.Log.Debug("Setting threads to %d", numCPUs)
	}
	runtime.GOMAXPROCS(s.Config.Global.Threads + 2) // thread count + main + web server
}
