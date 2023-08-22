package localpath

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rumenvasilev/rvsecret/internal/core"
	"github.com/rumenvasilev/rvsecret/internal/matchfile"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"github.com/rumenvasilev/rvsecret/version"

	"golang.org/x/sync/errgroup"
)

// DoFileScan with create a match object and then test for various criteria necessary in order to determine if it should be scanned. This includes if it should be skipped due to a default or user supplied extension, if it matches a test regex, or is in a protected directory or is itself protected. This will only run when doing scanLocalPath.
func doFileScan(filename string, sess *core.Session) {

	// Set default values for all pre-requisites for a file scan
	likelyTestFile := false

	// This is the total number of files that we know exist in out path. This does not care about the scan, it is simply the total number of files found
	sess.State.Stats.IncrementFilesTotal()

	mf := matchfile.New(filename)
	if mf.IsSkippable(sess.Config.SkippableExt, sess.Config.SkippablePath) {
		sess.Out.Debug("%s is listed as skippable and is being ignored", filename)
		sess.State.Stats.IncrementFilesIgnored()
		return
	}

	// If we are not scanning tests then drop all files that match common test file patterns
	// If we do not want to scan any test files or paths we check for them and then exclude them if they are found
	// The default is to not scan test files or common test paths
	if !sess.Config.ScanTests {
		likelyTestFile = util.IsTestFileOrPath(filename)
	}

	if likelyTestFile {
		// We want to know how many files have been ignored
		sess.State.Stats.IncrementFilesIgnored()
		sess.Out.Debug("%s is a test file and being ignored", filename)
		return
	}

	// Check the file size of the file. If it is greater than the default size
	// then we increment the ignored file count and pass on through.
	val, msg := util.IsMaxFileSize(filename, sess.Config.MaxFileSize)
	if val {
		sess.State.Stats.IncrementFilesIgnored()
		sess.Out.Debug("%s %s", filename, msg)
		return
	}

	if sess.Config.Debug {
		// Print the filename of every file being scanned
		sess.Out.Debug("Analyzing %s", filename)
	}

	// Increment the number of files scanned
	sess.State.Stats.IncrementFilesScanned()
	var content string
	// Scan the file for know signatures
	for _, signature := range core.Signatures {
		bMatched, matchMap := signature.ExtractMatch(mf, sess, nil)
		// for every instance of the secret that matched the specific rule create a new finding
		for k, v := range matchMap {
			cleanK := strings.SplitAfterN(k, "_", 2)
			content = cleanK[1]

			// destroy the secret if the flag is set
			if sess.Config.HideSecrets {
				content = ""
			}

			if bMatched {
				finding := &core.Finding{
					FilePath:         filename,
					Action:           `File Scan`,
					Description:      signature.Description(),
					SignatureID:      signature.SignatureID(),
					Content:          content,
					RepositoryOwner:  `not-a-repo`,
					RepositoryName:   `not-a-repo`,
					CommitHash:       ``,
					CommitMessage:    ``,
					CommitAuthor:     ``,
					LineNumber:       strconv.Itoa(v),
					SecretID:         util.GenerateID(),
					AppVersion:       version.AppVersion(),
					SignatureVersion: sess.SignatureVersion,
				}

				// Add a new finding and increment the total
				finding.Initialize(sess)
				sess.AddFinding(finding)

				// print the current finding to stdout
				finding.RealtimeOutput(sess)
			}
		}
	}
}

// ScanDir will scan a directory for all the files and then kick a file scan on each of them
func scanDir(path string, sess *core.Session) {
	ctx, cancel := context.WithTimeout(context.Background(), 3600*time.Second)
	defer cancel()

	// get an slice of of all paths
	files, err := search(ctx, path, sess.Config.SkippablePath, sess)
	if err != nil {
		sess.Out.Error("There is an error scanning %s: %s", path, err.Error())
	}

	maxThreads := 100
	sem := make(chan struct{}, maxThreads)

	var wg sync.WaitGroup

	wg.Add(len(files))
	for _, file := range files {
		sem <- struct{}{}
		go func(f string) {
			defer wg.Done()

			// scan the specific file if it is found to be a valid candidate
			doFileScan(f, sess)
			<-sem
		}(file)
	}

	wg.Wait()
}

// Search will walk the path or a given directory and append each viable path to an array
func search(ctx context.Context, root string, skippablePath []string, sess *core.Session) ([]string, error) {
	sess.Out.Important("Enumerating Paths")
	g, ctx := errgroup.WithContext(ctx)
	paths := make(chan string, 20)

	// get all the paths within a tree
	g.Go(func() error {
		defer close(paths)

		return filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
			// This will check against the combined list of directories that we want to exclude
			// There is the stock list that we pre-defined and then user have the ability to add to this list via the commandline
			for _, p := range skippablePath {
				if strings.HasPrefix(path, p) {
					return nil
				}
			}

			if os.IsPermission(err) {
				return nil
			}
			if !fi.Mode().IsRegular() {
				return nil
			}

			select {
			case paths <- path:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	})

	var result []string
	for r := range paths {
		result = append(result, r)
	}
	return result, g.Wait()
}
