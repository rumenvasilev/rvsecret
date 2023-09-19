package github

import (
	"fmt"
	"time"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/webserver"
)

func Scan(cfg *config.Config, log *log.Logger) error {
	// create session
	sess, err := core.NewSessionWithConfig(cfg, log)
	if err != nil {
		return err
	}
	// Ensure user input exists and validate it
	err = sess.ValidateUserInput()
	if err != nil {
		// log.Warn("No token present. Will proceed scanning only public repositories.")
		return err
	}

	// Start webserver
	if cfg.Global.WebServer && !cfg.Global.Silent {
		ws := webserver.New(*cfg, sess.State, log)
		go ws.Start()
	}

	// By default we display a header to the user giving basic info about application. This will not be displayed
	// during a silent run which is the default when using this in an automated fashion.
	core.HeaderInfo(*cfg, sess.State.Stats.StartedAt.Format(time.RFC3339), sess.Out)

	if cfg.Global.Debug {
		fmt.Println("Global debug set to", cfg.Global.Debug)
		log.Debug(config.PrintDebug(sess.SignatureVersion))
	}

	log.Debug("We have these orgs: %s", sess.GithubUserOrgs)
	log.Debug("We have these users: %s", sess.GithubUserLogins)
	log.Debug("We have these repos: %s", sess.GithubUserRepos)

	//Create a github client to be used for the session
	err = sess.InitGitClient()
	if err != nil {
		return err
	}

	if sess.GithubUserLogins != nil {
		err = core.GatherUsers(sess)
		if err != nil {
			return err
		}
		err = core.GatherGithubRepositoriesFromOwner(sess)
		if err != nil {
			return err
		}
	} else if cfg.Global.ExpandOrgs && sess.GithubUserOrgs != nil {
		// FIXME: this should be from --add-org-members
		core.GatherOrgsMembersRepositories(sess)
	} else if sess.GithubUserOrgs != nil {
		err = core.GatherOrgs(sess, log)
		if err != nil {
			return err
		}
		err = core.GatherGithubOrgRepositories(sess, log)
		if err != nil {
			return err
		}
	} else {
		// Catchall for not being able to scan any as either we have no information or
		// we don't have the rights kinds of information
		return fmt.Errorf("please specify an org or user that contains the repo(s)")
	}

	core.AnalyzeRepositories(sess, sess.State.Stats, log)
	sess.Finish()

	core.SummaryOutput(sess)

	if !cfg.Global.Silent && cfg.Global.WebServer {
		log.Important("Press Ctrl+C to stop web server and exit.")
		select {}
	}
	return nil
}
