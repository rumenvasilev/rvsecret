// Package core represents the core functionality of all commands
package core

import (
	"os"
	"strconv"
	"strings"
	"sync"

	coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/core/finding"
	"github.com/rumenvasilev/rvsecret/internal/core/git"
	"github.com/rumenvasilev/rvsecret/internal/core/matchfile"
	"github.com/rumenvasilev/rvsecret/internal/core/signatures"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/session"
	"github.com/rumenvasilev/rvsecret/internal/stats"
	"github.com/rumenvasilev/rvsecret/internal/util"
	_git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// AnalyzeRepositories will clone the repos, grab their history for analysis of files and content.
//
//	Before the analysis is done we also check various conditions that can be thought of as filters and
//	are controlled by flags. If a directory, file, or the content pass through all of the filters then
//	it is scanned once per each signature which may lead to a specific secret matching multiple rules
//	and then generating multiple findings.
func AnalyzeRepositories(sess *session.Session, st *stats.Stats) {
	log := log.Log
	st.UpdateStatus(stats.StatusAnalyzing)
	repoCnt := len(sess.State.Repositories)
	if repoCnt == 0 {
		log.Error("No repositories have been gathered.")
	}

	var ch = make(chan coreapi.Repository, repoCnt)
	var wg sync.WaitGroup

	// Calculate the number of threads based on the flag and the number of repos. If the number of repos
	// being scanned is less than the number of threads the user requested, then the thread count is the
	// number of repos.
	threadNum := sess.Config.Global.Threads
	log.Debug("Defaulting threadNum to %d", sess.Config.Global.Threads)
	if repoCnt <= 1 {
		threadNum = 1
	} else if repoCnt <= sess.Config.Global.Threads {
		log.Debug("Setting threadNum to %d", repoCnt)
		threadNum = repoCnt
	}
	log.Debug("Threads for repository analysis: %d", threadNum)
	wg.Add(threadNum)
	log.Important("Analyzing %d %s...", repoCnt, util.Pluralize(repoCnt, "repository", "repositories"))

	// Start analyzer workers
	for i := 0; i < threadNum; i++ {
		go analyzeWorker(i, ch, &wg, sess, st)
	}

	// Feed repos to the analyzer workers
	for _, repo := range sess.State.Repositories {
		ch <- *repo
	}

	// We close the channel to signal to all the for loops to end,
	// once they've finished processing all the scheduled work.
	close(ch)
	wg.Wait()
}

func analyzeWorker(tid int, ch chan coreapi.Repository, wg *sync.WaitGroup, sess *session.Session, st *stats.Stats) {
	log := log.Log
	for {
		log.Debug("[THREAD #%d] Requesting new repository to analyze...", tid)
		repo, ok := <-ch
		if !ok {
			log.Debug("[THREAD #%d] No more tasks, marking WaitGroup done", tid)
			wg.Done()
			return
		}

		// Clone the repository from the remote source or if a local repo from the path
		// The path variable is returning the path that the clone was done to. The repo is cloned directly
		// there.
		log.Debug("[THREAD #%d][%s] Cloning repository...", tid, repo.CloneURL)
		clone, path, err := cloneRepository(sess.Config, st.IncrementRepositoriesCloned, repo)
		if err != nil {
			log.Error("%v", err)
			cleanUpPath(path)
			continue
		}
		log.Debug("[THREAD #%d][%s] Cloned repository to: %s", tid, repo.CloneURL, path)

		analyzeHistory(sess, clone, tid, path, repo)

		log.Debug("[THREAD #%d][%s] Done analyzing commits", tid, repo.CloneURL)
		log.Debug("[THREAD #%d][%s] Deleted %s", tid, repo.CloneURL, path)

		cleanUpPath(path)
		st.IncrementRepositoriesScanned()
	}
}

func cleanUpPath(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Log.Error("Could not remove path from disk: %s", err.Error())
	}
}

func analyzeHistory(sess *session.Session, clone *_git.Repository, tid int, path string, repo coreapi.Repository) {
	stats := sess.State.Stats
	log := log.Log

	// Get the full commit history for the repo
	history, err := git.GetRepositoryHistory(clone)
	if err != nil {
		log.Error("[THREAD #%d][%s] Cannot get full commit history, error: %v", tid, repo.CloneURL, err)
		if err := os.RemoveAll(path); err != nil {
			log.Error("[THREAD #%d][%s] Cannot remove path from disk, error: %v", tid, repo.CloneURL, err)
		}
		return
	}
	log.Debug("[THREAD #%d][%s] Number of commits: %d", tid, repo.CloneURL, len(history))

	// Add in the commits found to the repo into the running total of all commits found
	// sess.Stats.CommitsTotal = sess.Stats.CommitsTotal + len(history)
	stats.IncrementCommitsTotal(len(history))

	// For every commit in the history we want to look through it for any changes
	// there is a known bug in here related to files that have changed paths from the most
	// recent path. The does not do a fetch per history so if a file changes paths from
	// the current one it will throw a file not found error. You can see this by turning
	// on debugging.
	for _, commit := range history {
		log.Debug("[THREAD #%d][%s] Analyzing commit: %s", tid, repo.CloneURL, commit.Hash)

		// Increment the total number of commits. This needs to be used in conjunction with
		// the total number of commits scanned as a commit may have issues and not be scanned once
		// it is found.
		stats.IncrementCommitsScanned()

		if yes := isDirtyCommit(sess, commit, repo, clone, path, tid); yes {
			// Increment the number of commits that were found to be dirty
			stats.IncrementCommitsDirty()
		}
	}
}

// isDirtyCommit will analyze all the changes and return bool if there's a dirty commit
func isDirtyCommit(sess *session.Session, commit *object.Commit, repo coreapi.Repository, clone *_git.Repository, path string, tid int) bool {
	stats := sess.State.Stats
	log := log.Log

	// This will be used to increment the dirty commit stat if any matches are found. A dirty commit
	// means that a secret was found in that commit. This provides an easier way to manually to look
	// through the commit history of a given repo.
	dirtyCommit := false

	changes, _ := git.GetChanges(commit, clone)
	log.Debug("[THREAD #%d][%s] %d changes in %s", tid, repo.CloneURL, len(changes), commit.Hash)

	for _, change := range changes {
		// The total number of files that were evaluated
		stats.IncrementFilesTotal()

		// TODO Is this need for the finding object, why are we saving this?
		changeAction := git.GetChangeAction(change)

		// TODO Add an example of the output from this function
		fPath := git.GetChangePath(change)

		// TODO Add an example of this
		fullFilePath := path + "/" + fPath

		// Break a file name up into its composite pieces including the extension and base name
		mf := matchfile.New(fullFilePath)

		// Check if file has to be ignored
		if ok, msg := ignoredFile(sess.Config.Global.ScanTests, sess.Config.Global.MaxFileSize, fullFilePath, mf, sess.Config.Global.SkippableExt, sess.Config.Global.SkippablePath); ok {
			log.Debug("[THREAD #%d][%s] %s %s", tid, repo.CloneURL, fPath, msg)
			stats.IncrementIgnoredFiles()
			continue
		}

		// We are now finally at the point where we are going to scan a file so we implement
		// that count.
		stats.IncrementScannedFiles()

		// We set this to a default of false and will be used at the end of matching to
		// increment the file count. If we try and do this in the loop it will hit for every
		// signature and give us a false count.
		// dirtyFile := false

		// call signaturesfunc
		dirtyFile, dcommit, ignored, out := signatures.Discover(mf, change, sess.Config, sess.Signatures)
		for _, v := range out {
			fin := &finding.Finding{
				Action:           changeAction,
				Content:          v.Content,
				CommitAuthor:     commit.Author.String(),
				CommitHash:       commit.Hash.String(),
				CommitMessage:    strings.TrimSpace(commit.Message),
				Description:      v.Sig.Description(),
				FilePath:         fPath,
				AppVersion:       sess.Config.Global.AppVersion,
				LineNumber:       strconv.Itoa(v.LineNum),
				RepositoryName:   repo.Name,
				RepositoryOwner:  repo.Owner,
				SignatureID:      v.Sig.SignatureID(),
				SignatureVersion: sess.SignatureVersion,
				SecretID:         util.GenerateID(),
			}
			_ = fin.Initialize(sess.Config.Global.ScanType, sess.Config.Github.GithubEnterpriseURL)
			// Add it to the session
			sess.State.AddFinding(fin)
			log.Debug("[THREAD #%d][%s] Done analyzing changes in %s", tid, repo.CloneURL, commit.Hash)

			// Print realtime data to stdout
			fin.RealtimeOutput(sess.Config.Global)
		}

		if dirtyFile {
			stats.IncrementDirtyFiles()
		}
		if dcommit {
			dirtyCommit = dcommit
		}
		if ignored > 0 {
			stats.IncrementIgnoredFilesWith(ignored)
		}
	}
	return dirtyCommit
}

func ignoredFile(cfgScanTests bool, cfgMaxFileSize int64, fullFilePath string, mf matchfile.MatchFile, cfgSkippableExt, cfgSkippablePath []string) (bool, string) {
	// required as that is a map of interfaces.
	// scanTests := DefaultValues["scan-tests"]
	likelyTestFile := cfgScanTests

	// If we do not want to scan tests we run some checks to see if the file in
	// question is a test file. This will return a true if it is a test file.
	if !cfgScanTests {
		likelyTestFile = util.IsTestFileOrPath(fullFilePath)
	}

	// If the file is likely a test then ignore it. By default this is currently
	// set to false which means we do NOT want to scan tests. This means that we
	// check above and if this returns true because it is likely a test file, we
	// increment the ignored file count and pass through scanning the file and content.
	if likelyTestFile {
		// If we are not scanning the file then by definition we are ignoring it
		return true, "is a test file and is being ignored"
	}

	// Check the file size of the file. If it is greater than the default size then
	// then we increment the ignored file count and pass on through.
	yes, msg := util.IsMaxFileSize(fullFilePath, cfgMaxFileSize)
	if yes {
		return true, msg
	}

	if mf.IsSkippable(cfgSkippableExt, cfgSkippablePath) {
		return true, "is skippable and is being ignored"
	}
	return false, ""
}
