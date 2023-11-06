package github

import (
	"context"
	"fmt"
	"time"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core"
	"github.com/rumenvasilev/rvsecret/internal/core/banner"
	"github.com/rumenvasilev/rvsecret/internal/core/provider"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/output"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/session"
	"github.com/rumenvasilev/rvsecret/internal/webserver"
)

const errMsg = "no targets (%s) to search repositories for have been found"

type Github struct {
	Cfg *config.Config
}

func (g Github) Run() error {
	cfg := g.Cfg
	log := log.Log
	// create session
	sess, err := session.NewWithConfig(cfg)
	if err != nil {
		return err
	}
	// Ensure user input exists and validate it
	err = core.ValidateUserInput(sess)
	if err != nil {
		// log.Warn("No token present. Will proceed scanning only public repositories.")
		return err
	}

	// Start webserver
	if cfg.Global.WebServer && !cfg.Global.Silent {
		ws := webserver.New(*cfg, sess.State)
		go ws.Start()
	}

	// By default we display a header to the user giving basic info about application. This will not be displayed
	// during a silent run which is the default when using this in an automated fashion.
	banner.HeaderInfo(cfg.Global, sess.State.Stats.StartedAt.Format(time.RFC3339), len(sess.Signatures))

	if cfg.Global.Debug {
		log.Debug(config.PrintDebug(sess.SignatureVersion))
	}

	log.Debug("We have these orgs: %s", sess.GithubUserOrgs)
	log.Debug("We have these users: %s", sess.GithubUserLogins)
	log.Debug("We have these repos: %s", sess.GithubUserRepos)

	// Create a github client to be used for the session
	sess.Client, err = provider.InitGitClient(sess.Config)
	if err != nil {
		return err
	}

	switch {
	case sess.GithubUserLogins != nil:
		core.GatherUserOrOrg(sess, sess.GithubUserLogins)
		err = fmt.Errorf(errMsg, "users")
	case sess.GithubUserOrgs != nil:
		if cfg.Global.ExpandOrgs {
			log.Debug("ExpandOrgs is enabled. Searching for members in the organization...")
			core.GatherOrgsMembers(sess)
			err = fmt.Errorf(errMsg, "org members")
		} else {
			core.GatherUserOrOrg(sess, sess.GithubUserOrgs)
			err = fmt.Errorf(errMsg, "orgs")
		}
	default:
		// Catchall for not being able to scan any as either we have no information or
		// we don't have the rights kinds of information
		return fmt.Errorf("please specify an org or user that contains the repo(s)")
	}

	if len(sess.State.Targets) == 0 && err != nil {
		return err
	}

	core.GetAllRepositoriesForTargets(context.TODO(), sess)
	core.AnalyzeRepositories(sess, sess.State.Stats)
	sess.Finish()

	err = output.Summary(sess.State, sess.Config.Global, sess.SignatureVersion)
	if err != nil {
		return err
	}

	if !cfg.Global.Silent && cfg.Global.WebServer {
		log.Important("Press Ctrl+C to stop web server and exit.")
		select {}
	}
	return nil
}

var _ api.Scanner = (*Github)(nil)
