package core

import (
	"context"
	"sync"

	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/session"
	"github.com/rumenvasilev/rvsecret/internal/util"
)

// GatherRepositories will gather all repositories associated with a given target during a scan session.
// This is done using threads, whose count is set via commandline flag. Care much be taken to avoid rate
// limiting associated with suspected DOS attacks.
func GatherRepositories(ctx context.Context, sess *session.Session) {
	log := log.Log

	var ch = make(chan *_coreapi.Owner, len(sess.State.Targets))
	log.Debug("Number of targets: %d", len(sess.State.Targets))

	threadNum := setThreadNum(len(sess.State.Targets), sess.Config.Global.Threads)
	log.Debug("Threads for repository gathering: %d", threadNum)

	var wg sync.WaitGroup
	wg.Add(threadNum)
	for i := 0; i < threadNum; i++ {
		go retrieveReposWorker(ctx, i, &wg, ch, sess)
	}

	for _, target := range sess.State.Targets {
		ch <- target
	}
	close(ch)
	wg.Wait()
}

func retrieveReposWorker(ctx context.Context, workerID int, wg *sync.WaitGroup, ch <-chan *_coreapi.Owner, sess *session.Session) {
	log := log.Log
	for {
		select {
		case <-ctx.Done():
			log.Info("[THREAD #%d] Job cancellation requested.", workerID)
			wg.Done()
			return
		case target, ok := <-ch:
			if !ok {
				log.Debug("[THREAD #%d]: No more targets to retrieve", workerID)
				wg.Done()
				return
			}
			repos, err := sess.Client.GetRepositoriesFromOwner(ctx, *target)
			if err != nil {
				log.Error("[THREAD #%d]: Failed to retrieve repositories from %s: %s", workerID, *target.Login, err)
			}
			if len(repos) == 0 {
				log.Debug("[THREAD #%d]: No repositories have been gathered for %s", workerID, *target.Login)
				continue
			}
			for _, repo := range repos {
				log.Debug("[THREAD #%d]: Retrieved repository: %s", workerID, repo.CloneURL)
				// Increment the total number of repos found even if we are not cloning them
				sess.State.Stats.IncrementRepositoriesTotal()
				// Only a subset of repos
				if isGithub(sess, repo) {
					continue
				}
				sess.State.AddRepository(repo)
			}
			log.Info("[THREAD #%d]: Retrieved %d %s from %s", workerID, len(repos), util.Pluralize(len(repos), "repository", "repositories"), *target.Login)
		}
	}
}

func isGithub(sess *session.Session, repo *_coreapi.Repository) bool {
	result := false
	switch sess.Config.Global.ScanType {
	case api.Github, api.GithubEnterprise:
		result = true
		if sess.GithubUserRepos != nil {
			if isFilteredRepo(repo.Name, sess.GithubUserRepos) {
				log.Log.Debug("Retrieved repository %s", repo.FullName)

				sess.State.AddRepository(repo)
			}
		}
	}

	return result
}

func isFilteredRepo(name string, in []string) bool {
	for _, r := range in {
		if name == r {
			return true
		}
	}
	return false
}

func setThreadNum(stateTargets, configTargets int) int {
	var threadNum int
	if stateTargets == 1 {
		threadNum = 1
	} else if stateTargets <= configTargets {
		threadNum = stateTargets - 1
	} else {
		threadNum = configTargets
	}
	return threadNum
}
