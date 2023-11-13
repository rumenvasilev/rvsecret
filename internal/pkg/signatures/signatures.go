package signatures

import (
	"fmt"
	"os"
	"regexp"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core/provider/github"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/session"
	"github.com/rumenvasilev/rvsecret/internal/util"
	whilp "github.com/whilp/git-urls"
	"gopkg.in/yaml.v3"
)

// signature version options
// `latest` => last release in github
// `current` => main branch latest commit
// `semver` => specific version

func Update(cfg *config.Config) error {
	log := log.Log
	var dir string

	// create session
	sess, err := session.NewWithConfig(cfg)
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
		if github.IsCredentialsError(err) {
			log.Debug(err.Error())
			return fmt.Errorf("github token is not authorized, please update its permissions or generate a new one")
		}
		log.Warn("Couldn't fetch the signatures from Github REST API, falling back to git clone method")
		dir, err = fetchSignaturesWithGit(cfg.Signatures.Version, sess)
		if err != nil {
			return fmt.Errorf("couldn't fetch the signatures with git clone either, reason: %w", err)
		}
	}

	// install the signatures
	if updateSignatures(dir, sess) {
		log.Info("The signatures have been successfully updated at: %s", cfg.Signatures.Path)
	} else {
		return fmt.Errorf("the signatures were not updated")
	}
	return nil
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
func updateSignatures(rRepo string, sess *session.Session) bool {
	log := log.Log
	// create a temp directory to hold the signatures we pull
	// TODO put this in /tmp via a real library
	tempSignaturesDir := rRepo + "/signatures"

	// ensure we have the proper home directory
	// if the signatures path does not exist then we create it
	home, err := util.MakeHomeDir(sess.Config.Signatures.Path)
	if err != nil {
		log.Error("couldn't create signatures directory %q, error %s", sess.Config.Signatures.Path, err.Error())
		return false
	}

	// if we want to test the signatures before we install them
	if sess.Config.Signatures.Test {
		if !executeTests(tempSignaturesDir) {
			log.Error("Signature tests have failed. Files are available for inspection in the temporary directory: %q", tempSignaturesDir)
			return false
		}
	}

	// copy the files from the temp directory to the signatures directory
	err = util.CopyFiles(tempSignaturesDir, home)
	if err != nil {
		log.Error(err.Error())
		log.Error("The signature files are available for inspection in the temporary directory: %q", tempSignaturesDir)
		return false
	}
	cleanUp(rRepo)
	return true
}

// executeTests will run any tests associated with the expressions
func executeTests(dir string) bool {
	log := log.Log
	log.Debug("Running tests on acquired signature files...")
	sigFiles, err := util.GetYamlFiles(dir)
	if err != nil {
		log.Error("Failed to get signature files from target path %q, error: %q", dir, err.Error())
		return false
	}

	// Run tests:
	return isYaml(sigFiles)
}

// cleanUp after ourselves and remove any temp garbage
func cleanUp(path string) {
	if err := os.RemoveAll(path); err != nil {
		log.Log.Error(err.Error())
	}
}

// runYamlTest would try to marshal the input file
// if it can => true
// if it cannot => false
func isYaml(files []string) bool {
	log := log.Log
	var tmp = make(map[string]interface{})
	for _, v := range files {
		// read file
		f, err := os.ReadFile(v)
		if err != nil {
			log.Error("Failed to read file %q, error: %q", f, err.Error())
			return false
		}
		// try unmarshalling
		err = yaml.Unmarshal(f, tmp)
		if err != nil {
			log.Error("YAML unmarshalling test for file %q failed. Error: %q", f, err.Error())
			return false
		}
	}
	return true
}
