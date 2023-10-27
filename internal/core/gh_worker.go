package core

import (
	"context"

	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/util"
)

const retrievedRepo string = " Retrieved repository %s"

// GetAllRepositoriesForOwner will find all repositories for an owner (user or org)
// and populate the repositories state list (s.State.Repositories)
func GetAllRepositoriesForOwner(ctx context.Context, login string, kind string, tid int, sess *Session, log *log.Logger) {
	// Retrieve all the repos in an org regardless of public/private
	repos, err := sess.Client.GetRepositoriesFromOwner(ctx, _coreapi.Owner{
		Login: util.StringToPointer(login),
		Kind:  util.StringToPointer(kind),
	})
	if err != nil {
		// We might get partial result, which is why we only log the error as warning
		// var ghErrResp *github.ErrorResponse
		// if errors.As(err, &ghErrResp) {
		// 	if ghErrResp.Response.StatusCode == 404 {
		// 		log.Warn("[THREAD #%d]: %s", tid, ghErrResp.Error())
		// 	}
		// }
		log.Debug("[THREAD #%d]: GetAllRepositoriesForOwner: %s", tid, err.Error())
	}

	// In the case where all the repos are private
	if len(repos) == 0 {
		log.Debug("[THREAD #%d]: No repositories have been gathered for %s", login)
		return
	}

	// If we re only looking for a subset of the repos in an org we do a comparison
	// of the repos gathered for the org and the list pf repos that we care about.
	for _, repo := range repos {
		// Increment the total number of repos found even if we are not cloning them
		sess.State.Stats.IncrementRepositoriesTotal()

		// Only a subset of repos
		if sess.GithubUserRepos != nil {
			if isFilteredRepo(repo.Name, sess.GithubUserRepos) {
				log.Debug(retrievedRepo, repo.FullName)
				// Add the repo to the sess to be scanned
				sess.State.AddRepository(repo)
			}
			continue
		}

		log.Debug(retrievedRepo, repo.FullName)
		// If we are not doing any filtering and simply grabbing all available repos we add the repos
		// to the session to be scanned
		log.Debug("Adding repo %s", repo.Name)
		sess.State.AddRepository(repo)
	}
}

func isFilteredRepo(name string, in []string) bool {
	for _, r := range in {
		if name == r {
			return true
		}
	}
	return false
}
