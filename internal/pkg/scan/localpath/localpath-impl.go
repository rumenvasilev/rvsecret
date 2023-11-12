package localpath

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rumenvasilev/rvsecret/internal/core"
	"github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/session"

	"golang.org/x/sync/errgroup"
)

// ScanDir will scan a directory for all the files and then kick a file scan on each of them
func scanDir(path string, sess *session.Session) {
	ctx, cancel := context.WithTimeout(context.Background(), 3600*time.Second)
	defer cancel()

	// get an slice of of all paths
	files, err := search(ctx, path, sess.Config.Global.SkippablePath, sess)
	if err != nil {
		log.Log.Error("There is an error scanning %s: %s", path, err.Error())
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
			core.AnalyzeObject(ctx, sess, nil, nil, f, api.Repository{})
			<-sem
		}(file)
	}

	wg.Wait()
}

// Search will walk the path or a given directory and append each viable path to an array
func search(ctx context.Context, root string, skippablePath []string, sess *session.Session) ([]string, error) {
	log.Log.Important("Enumerating Paths")
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
