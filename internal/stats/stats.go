package stats

import (
	"sync"
	"time"

	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
)

// Stats hold various runtime statistics used for perf data as well generating various reports
// They are accessed from the web server as well
type Stats struct {
	*sync.Mutex
	StartedAt           time.Time       // The time we started the scan
	FinishedAt          time.Time       // The time we finished the scan
	Status              _coreapi.Status // The running status of a scan for the web interface
	Progress            float64         // The running progress for the bar on the web interface
	RepositoriesTotal   int             // The toatal number of repos discovered
	RepositoriesScanned int             // The total number of repos scanned (not excluded, errors, empty)
	RepositoriesCloned  int             // The total number of repos cloned (excludes errors and excluded, includes empty)
	Organizations       int             // The number of github orgs
	CommitsScanned      int             // The number of commits scanned in a repo
	CommitsDirty        int             // The number of commits in a repo found to have secrets
	FilesScanned        int             // The number of files actually scanned
	FilesIgnored        int             // The number of files ignored (tests, extensions, paths)
	FilesTotal          int             // The total number of files that were processed
	FilesDirty          int
	FindingsTotal       int // The total number of findings. There can be more than one finding per file and more than one finding of the same type in a file
	Users               int // Github users
	Targets             int // The number of dirs, people, orgs, etc on the command line or config file (what do you want rvsecret to enumerate on)
	Repositories        int // This will point to RepositoriesScanned
	CommitsTotal        int // This will point to commits scanned
	Findings            int // This will point to findings total
	Files               int // This will point to FilesScanned
	Commits             int // This will point to CommitsScanned
}

// InitStats will set the initial values for a session
func Init() *Stats {
	return &Stats{
		Mutex:         &sync.Mutex{},
		FilesIgnored:  0,
		FilesScanned:  0,
		FindingsTotal: 0,
		Organizations: 0,
		Progress:      0.0,
		StartedAt:     time.Now(),
		Status:        _coreapi.StatusFinished,
		Users:         0,
		Targets:       0,
		Repositories:  0,
		CommitsTotal:  0,
		Findings:      0,
		Files:         0,
	}
}

func (s *Stats) UpdateStatus(to _coreapi.Status) {
	s.Status = to
}

// IncrementTargets will add one to the running target count during the target discovery phase of a session
func (s *Stats) IncrementTargets() {
	s.Lock()
	defer s.Unlock()
	s.Targets++
}

// IncrementRepositories will add one to the running repository count during the target discovery phase of a session
func (s *Stats) IncrementRepositories() {
	s.Lock()
	defer s.Unlock()
	s.Repositories++
}

// IncrementCommitsTotal will add one to the running count of commits during the target discovery phase of a session
func (s *Stats) IncrementCommitsTotal(with int) {
	s.Lock()
	defer s.Unlock()
	if with != 0 {
		s.CommitsTotal = s.CommitsTotal + with
	} else {
		s.CommitsTotal++
	}
}

// IncrementFiles will add one to the running count of files during the target discovery phase of a session
func (s *Stats) IncrementFiles() {
	s.Lock()
	defer s.Unlock()
	s.Files++
}

// IncrementFindings will add one to the running count of findings during the target discovery phase of a session
func (s *Stats) IncrementFindings() {
	s.Lock()
	defer s.Unlock()
	s.Findings++
}

// UpdateProgress will update the progress percentage
func (s *Stats) updateProgress(current int, total int) {
	//s.Lock() TODO REMOVE ME
	//defer s.Unlock() TODO REMOVE ME
	progress := 100.0
	if current < total {
		progress = (float64(current) * float64(100)) / float64(total)
	}
	s.Progress = progress
}

// IncrementFilesTotal will bump the count of files that have been discovered. This does not reflect
// if the file was scanned/skipped. It is simply a count of files that were found.
func (s *Stats) IncrementFilesTotal() {
	s.Lock()
	defer s.Unlock()
	s.FilesTotal++
}

// IncrementFilesDirty will bump the count of files that have been discovered. This does not reflect
// if the file was scanned/skipped. It is simply a count of files that were found.
func (s *Stats) IncrementFilesDirty() {
	s.Lock()
	defer s.Unlock()
	s.FilesDirty++
}

// IncrementFilesScanned will bump the count of files that have been scanned successfully.
func (s *Stats) IncrementFilesScanned() {
	s.Lock()
	defer s.Unlock()
	s.FilesScanned++
	s.Files++
}

// IncrementFilesIgnored will bump the number of files that have been ignored for various reasons.
func (s *Stats) IncrementFilesIgnored() {
	s.Lock()
	defer s.Unlock()
	s.FilesIgnored++
}

// IncrementFilesIgnoredWith will bump the number of files that have been ignored with a number.
func (s *Stats) IncrementFilesIgnoredWith(amount int) {
	s.Lock()
	defer s.Unlock()
	s.FilesIgnored += amount
}

// IncrementFindingsTotal will bump the total number of findings that have been matched. This does
// exclude any other documented criteria.
func (s *Stats) IncrementFindingsTotal() {
	s.Lock()
	defer s.Unlock()
	s.FindingsTotal++
	s.Findings++
}

// IncrementRepositoriesTotal will bump the total number of repositories that have been discovered.
// This will include empty ones as well as those that had errors
func (s *Stats) IncrementRepositoriesTotal() {
	s.Lock()
	defer s.Unlock()
	s.RepositoriesTotal++
}

// IncrementRepositoriesCloned will bump the number of repositories that have been cloned with errors but may be empty
func (s *Stats) IncrementRepositoriesCloned() {
	s.Lock()
	defer s.Unlock()
	s.RepositoriesCloned++
	s.updateProgress(s.RepositoriesCloned, s.RepositoriesCloned)
}

// IncrementRepositoriesScanned will bump the total number of repositories that have been scanned and are not empty
func (s *Stats) IncrementRepositoriesScanned() {
	s.Lock()
	defer s.Unlock()
	s.RepositoriesScanned++
	s.Repositories++
	s.updateProgress(s.RepositoriesScanned, s.RepositoriesScanned)
}

// IncrementUsers will bump the total number of users that have been enumerated
func (s *Stats) IncrementUsers() {
	s.Lock()
	defer s.Unlock()
	s.Users++
}

// IncrementCommitsScanned will bump the number of commits that have been scanned.
// This is scan wide and not on a per repo/org basis
func (s *Stats) IncrementCommitsScanned() {
	s.Lock()
	defer s.Unlock()
	s.CommitsScanned++
	s.Commits++
}

// IncrementOrgs will bump the number of orgs that have been gathered.
// This is scan wide and not on a per repo/org basis
func (s *Stats) IncrementOrgs() {
	s.Lock()
	defer s.Unlock()
	s.Organizations++
}

// IncrementCommitsDirty will bump the number of commits that have been found to be dirty,
// as in they contain one of more findings
func (s *Stats) IncrementCommitsDirty() {
	s.Lock()
	defer s.Unlock()
	s.CommitsDirty++
}
