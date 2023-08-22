package ghe

import (
	"fmt"

	"github.com/rumenvasilev/rvsecret/internal/core"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/api"
)

func Scan(log *log.Logger) error {
	sess, err := core.NewSession(api.GithubEnterprise, log)
	if err != nil {
		return err
	}
	// Ensure user input exists and validate it
	err = sess.ValidateUserInput()
	if err != nil {
		// log.Warn("No token present. Will proceed scanning only public repositories.")
		return err
	}

	// By default we display a header to the user giving basic info about application. This will not be displayed
	// during a silent run which is the default when using this in an automated fashion.
	core.HeaderInfo(*sess.Config, sess.State.Stats, sess.Out)

	log.Debug("We have these orgs: %s", sess.GithubUserOrgs)
	log.Debug("We have these users: %s", sess.GithubUserLogins)
	log.Debug("We have these repos: %s", sess.GithubUserRepos)

	// Create a github client to be used for the session
	err = sess.InitGitClient()
	if err != nil {
		return err
	}

	// If we have github users and no orgs or repos then we default to scan
	// the visible repos of that user.
	if sess.GithubUserLogins != nil {
		if sess.GithubUserOrgs == nil && sess.GithubUserRepos == nil {
			err = core.GatherUsers(sess)
			if err != nil {
				return err
			}
		}
	}

	// If the user has only given orgs then we grab all te repos from those orgs
	if sess.GithubUserOrgs != nil {
		if sess.GithubUserLogins == nil && sess.GithubUserRepos == nil {
			err = core.GatherOrgs(sess, log)
			if err != nil {
				return err
			}
		}
	}

	// If we have repo(s) given we need to ensure that we also have orgs or users. rvsecret will then
	// look for the repo in the user or login lists and scan it.
	if sess.GithubUserRepos != nil {
		if sess.GithubUserOrgs != nil {
			err = core.GatherOrgs(sess, log)
			if err != nil {
				return err
			}
			err = core.GatherGithubOrgRepositories(sess, log)
			if err != nil {
				return err
			}
		} else if sess.GithubUserLogins != nil {
			err = core.GatherUsers(sess)
			if err != nil {
				return err
			}
			err = core.GatherGithubRepositoriesFromOwner(sess)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("You need to specify an org or user that contains the repo(s).")
		}
	}

	core.AnalyzeRepositories(sess, sess.State.Stats, log)
	sess.Finish()

	core.SummaryOutput(sess)

	if !sess.Config.Silent && sess.Config.WebServer {
		log.Important("Press Ctrl+C to stop web server and exit.")
		select {}
	}
	return nil
}
