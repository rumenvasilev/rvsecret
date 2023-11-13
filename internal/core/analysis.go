// Package core represents the core functionality of all commands
package core

import (
	"context"
	"fmt"
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

type threadID int

const TID threadID = 0

// AnalyzeRepositories will clone the repos, grab their history for analysis of files and content.
//
//	Before the analysis is done we also check various conditions that can be thought of as filters and
//	are controlled by flags. If a directory, file, or the content pass through all of the filters then
//	it is scanned once per each signature which may lead to a specific secret matching multiple rules
//	and then generating multiple findings.
func AnalyzeRepositories(ctx context.Context, sess *session.Session, st *stats.Stats) {
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
		go analyzeWorker(ctx, i, &wg, ch, sess, st)
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

func analyzeWorker(ctx context.Context, workerID int, wg *sync.WaitGroup, ch chan coreapi.Repository, sess *session.Session, st *stats.Stats) {
	log := log.Log
	ctxworker := context.WithValue(ctx, TID, workerID)
	for {
		select {
		case <-ctx.Done():
			log.Info("[THREAD #%d] Job cancellation requested.", workerID)
			wg.Done()
			return
		case repo, ok := <-ch:
			log.Debug("[THREAD #%d] Requesting new repository to analyze...", workerID)
			if !ok {
				wg.Done()
				return
			}

			// Clone the repository from the remote source or if a local repo from the path
			// The path variable is returning the path that the clone was done to. The repo is cloned directly
			// there.
			log.Debug("[THREAD #%d][%s] Cloning repository...", workerID, repo.CloneURL)
			clone, path, err := cloneRepository(sess.Config, st.IncrementRepositoriesCloned, repo)
			if err != nil {
				log.Error("%v", err)
				cleanUpPath(path)
				continue
			}

			analyzeHistory(ctxworker, sess, clone, path, repo)
			cleanUpPath(path)
			st.IncrementRepositoriesScanned()
		}
	}
}

func cleanUpPath(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Log.Error("Could not remove path from disk: %s", err.Error())
	}
}

func analyzeHistory(ctx context.Context, sess *session.Session, clone *_git.Repository, path string, repo coreapi.Repository) {
	stats := sess.State.Stats
	log := log.Log
	tid := ctx.Value(TID)
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

		if yes := isDirtyCommit(ctx, sess, commit, repo, clone, path); yes {
			// Increment the number of commits that were found to be dirty
			stats.IncrementCommitsDirty()
		}
	}
}

// isDirtyCommit will analyze all the changes and return bool if there's a dirty commit
func isDirtyCommit(ctx context.Context, sess *session.Session, commit *object.Commit, repo coreapi.Repository, clone *_git.Repository, path string) bool {
	// stats := sess.State.Stats
	log := log.Log
	tid := ctx.Value(TID)
	// This will be used to increment the dirty commit stat if any matches are found. A dirty commit
	// means that a secret was found in that commit. This provides an easier way to manually to look
	// through the commit history of a given repo.
	dirtyCommit := false

	changes, _ := git.GetChanges(commit, clone)
	log.Debug("[THREAD #%d][%s] %d changes in %s", tid, repo.CloneURL, len(changes), commit.Hash)

	for _, change := range changes {
		if AnalyzeObject(ctx, sess, change, commit, path, repo) {
			dirtyCommit = true
		}
	}
	return dirtyCommit
}

func AnalyzeObject(ctx context.Context, sess *session.Session, change *object.Change, commit *object.Commit, filepath string, repo coreapi.Repository) bool {
	log := log.Log
	tid := ctx.Value(TID)
	cfg := sess.Config
	fPath := filepath

	// The total number of files that were evaluated
	sess.State.Stats.IncrementFilesTotal()

	var changeAction string
	if change != nil {
		changeAction = git.GetChangeAction(change)
		fPath = git.GetChangePath(change)
		filepath += fmt.Sprintf("/%s", fPath)
	}

	// Break a file name up into its composite pieces including the extension and base name
	mf := matchfile.New(filepath)

	// Check if file has to be ignored
	if ok, msg := isIgnoredFile(cfg.Global.ScanTests, cfg.Global.MaxFileSize, filepath, mf, cfg.Global.SkippableExt, cfg.Global.SkippablePath); ok {
		if change != nil {
			log.Debug("[THREAD #%d][%s] %s %s", tid, repo.CloneURL, fPath, msg)
		} else {
			log.Debug("[THREAD #%d] %s %s", tid, fPath, msg)
		}
		sess.State.Stats.IncrementIgnoredFiles()
		return false
	}

	// We are now finally at the point where we are going to scan a file so we implement
	// that count.
	sess.State.Stats.IncrementScannedFiles()

	dirtyFile, dcommit, ignored, results := signatures.Discover(mf, change, cfg, sess.Signatures)

	// Create template finding, so we won't need to pass all the parameters to the generateFindings func
	tpl := finding.Finding{
		Action:           changeAction,
		FilePath:         fPath,
		AppVersion:       sess.Config.Global.AppVersion,
		RepositoryName:   ``, // TODO do we need to set these 2 lines to nothing?
		RepositoryOwner:  ``,
		SignatureVersion: sess.SignatureVersion,
	}
	if commit != nil {
		tpl.CommitAuthor = commit.Author.String()
		tpl.CommitHash = commit.Hash.String()
		tpl.CommitMessage = strings.TrimSpace(commit.Message)
	}
	if repo != (coreapi.Repository{}) {
		tpl.RepositoryName = repo.Name
		tpl.RepositoryOwner = repo.Owner
	}
	for _, v := range results {
		generateFindings(sess, v, tpl)
	}

	if dirtyFile {
		sess.State.Stats.IncrementDirtyFiles()
	}

	if ignored > 0 {
		sess.State.Stats.IncrementIgnoredFilesWith(ignored)
	}

	return dcommit
}

func isIgnoredFile(cfgScanTests bool, cfgMaxFileSize int64, fullFilePath string, mf matchfile.MatchFile, cfgSkippableExt, cfgSkippablePath []string) (bool, string) {
	// Check if file exist before moving on
	if !util.PathExists(fullFilePath) {
		return true, "file does not exist"
	}

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
	if yes, msg := util.IsMaxFileSize(fullFilePath, cfgMaxFileSize); yes {
		return true, msg
	}

	// Check if it is a binary file
	yes, err := util.IsBinaryFile(fullFilePath)
	if yes || err != nil {
		return true, "is a binary file, ignoring"
	}

	if mf.IsSkippable(cfgSkippableExt, cfgSkippablePath) {
		return true, "is skippable, ignoring"
	}

	return false, ""
}

func generateFindings(sess *session.Session, data signatures.DiscoverOutput, template finding.Finding) {
	fin := template
	fin.Content = data.Content
	fin.Description = data.Sig.Description()
	fin.LineNumber = strconv.Itoa(data.LineNum)
	fin.SignatureID = data.Sig.SignatureID()

	// SecretID is used for dedup later under AddFinding()
	params := []string{fin.RepositoryName, fin.FilePath, fin.LineNumber, fin.Content}
	fin.SecretID = util.GenerateSecretIDWithParams(params...)

	_ = fin.Initialize(sess.Config.Global.ScanType, sess.Config.Github.GithubEnterpriseURL)

	// Add it to the session
	if sess.State.AddFinding(&fin) {
		// Print realtime data to stdout if finding was not a dup
		fin.RealtimeOutput(sess.Config.Global)
	}
}
