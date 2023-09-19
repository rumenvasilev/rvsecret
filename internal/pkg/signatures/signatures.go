package signatures

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/go-github/github"
	cp "github.com/otiai10/copy"
	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/util"
	whilp "github.com/whilp/git-urls"
)

// https://raw.githubusercontent.com/N0MoreSecr3ts/wraith-signatures/develop/signatures/default.yaml

// signature version options
// `latest` => last release in github
// `current` => main branch latest commit
// `semver` => specific version

func Update(cfg *config.Config, log *log.Logger) error {
	var dir string

	// create session
	sess, err := core.NewSessionWithConfig(cfg, log)
	if err != nil {
		return err
	}

	switch cfg.Signatures.Version {
	case "latest":
		log.Debug("Fetching latest release")
	default:
		log.Debug("Fetching a specific version: %q", cfg.Signatures.Version)
		semver := regexp.MustCompile(`^[0-2].[0-9]+.[0-9]+$`)
		if !semver.MatchString(cfg.Signatures.Version) {
			return fmt.Errorf("something went wrong, %w", err)
		}
	}

	// try from Github REST API first
	dir, err = fetchSignaturesFromGithubAPI(cfg.Signatures.Version, sess)
	if err != nil {
		if isCredentialsError(err) {
			log.Debug(err.Error())
			return fmt.Errorf("github token is not authorized, please update its permissions or generate a new one")
		}
		log.Warn("Couldn't fetch the signatures from Github REST API, falling back to git method")
		dir, err = fetchSignaturesWithGit(cfg.Signatures.Version, sess)
		if err != nil {
			return fmt.Errorf("couldn't fetch the signatures with git clone either, reason: %w", err)
		}
	}

	// install the signatures
	if updateSignatures(dir, sess, log) {
		log.Info("The signatures have been successfully updated at: %s", cfg.Signatures.Path)
	} else {
		return fmt.Errorf("the signatures were not updated")
	}
	return nil
}

func isCredentialsError(err error) bool {
	var gherr *github.ErrorResponse
	if errors.As(err, &gherr) {
		if gherr.Response.StatusCode == http.StatusUnauthorized {
			// log.Debug(err.Error())
			return true
		}
	}
	return false
}

// cleanInput will ensure that any user supplied git url is in the proper format
func cleanInput(u string) (string, error) {
	_, err := whilp.Parse(u)
	if err != nil {
		return "", err
	}
	return u, nil
}

// updateSignatures will install the new signatures into the specified location, changing the name of the previous set
func updateSignatures(rRepo string, sess *core.Session, log *log.Logger) bool {

	// create a temp directory to hold the signatures we pull
	// TODO put this in /tmp via a real library
	tempSignaturesDir := rRepo + "/signatures"

	// ensure we have the proper home directory
	home, err := util.SetHomeDir(sess.Config.Signatures.Path)
	if err != nil {
		// TODO-RV: Do something more?
		log.Error(err.Error())
	}

	// if the signatures path does not exist then we create it
	if !util.PathExists(home, log) {
		err := os.MkdirAll(home, 0700)
		if err != nil {
			log.Error(err.Error())
		}
	}

	// if we want to test the signatures before we install them
	// TODO need to implement something here
	if sess.Config.Signatures.Test {

		// if the tests pass then we install the signatures
		if executeTests(rRepo) {

			// copy the files from the temp directory to the signatures directory
			if err := cp.Copy(tempSignaturesDir, home); err != nil {
				log.Error(err.Error())
				return false
			}

			// get all the files in the signatures directory
			files, err := os.ReadDir(home)
			if err != nil {
				log.Error(err.Error())
				return false
			}

			// set them to the current user and the proper permissions
			for _, f := range files {
				if err := os.Chmod(home+"/"+f.Name(), 0644); err != nil {
					log.Error(err.Error())
					return false
				}
			}
			err = os.RemoveAll(rRepo)
			if err != nil {
				log.Error(err.Error())
			}
			return true

		}
		err := os.RemoveAll(rRepo)
		if err != nil {
			log.Error(err.Error())
		}
		return false

	}

	// copy the files from the temp directory to the signatures directory
	if err := cp.Copy(tempSignaturesDir, home); err != nil {
		log.Error(err.Error())
		return false
	}

	// get all the files in the signatures directory
	files, err := os.ReadDir(home)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	// set them to the current user and the proper permissions
	// TODO ensure these are .yaml somehow
	for _, f := range files {
		sFileExt := filepath.Ext(home + "/" + f.Name())
		if sFileExt == "yml" || sFileExt == "yaml" {
			if err := os.Chmod(home+"/"+f.Name(), 0644); err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}
	// Cleanup after ourselves and remove any temp garbage
	_ = os.RemoveAll(rRepo)
	return true
}

// executeTests will run any tests associated with the expressions
// TODO deal with this
func executeTests(dir string) bool {

	// run some tests here and return a true/false depending on the outcome
	return true
}
