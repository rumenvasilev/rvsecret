package signatures

import (
	"fmt"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
	"github.com/rumenvasilev/rvsecret/internal/core"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/api"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"github.com/spf13/viper"
	whilp "github.com/whilp/git-urls"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func Update(log *log.Logger) error {
	sess, err := core.NewSession(api.UpdateSignatures, log)
	if err != nil {
		return err
	}

	// get the signatures version or if blank, set it to latest
	// TODO this should be in the default values from the session
	signatureVersion := "latest"
	if viper.GetString("signatures-path") != "" {
		signatureVersion = viper.GetString("signatures-version")
	}

	// fetch the signatures from the remote location
	// git clone
	rRepo, err := fetchSignatures(signatureVersion, sess, log)
	if err != nil {
		return err
	}

	// install the signatures
	if updateSignatures(rRepo, sess, log) {
		// TODO set this in the session so we have a single location for everything
		fmt.Printf("The signatures have been successfully updated at: %s\n", viper.GetString("signatures-path"))
	} else {
		log.Warn("The signatures were not updated")
	}
	return nil
}

// fetchSignatures will download the signatures from a remote location to a temp location
func fetchSignatures(signatureVersion string, sess *core.Session, log *log.Logger) (string, error) {

	// TODO if this is not set then pull from the stock place, that should be the default url set in the session
	rURL := viper.GetString("signatures-url")

	// set the remote url that we will fetch
	// TODO need to look into this more
	remoteURL, err := cleanInput(rURL)
	if err != nil {
		return "", err
	}

	// TODO document this
	dir, err := os.MkdirTemp("", "rvsecret")
	if err != nil {
		return "", err
	}

	// for now we only pull from a given version at some point we can look at pulling the latest
	// TODO be able to pass in a commit or version string
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:           remoteURL,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", "stable")),
		SingleBranch:  true,
		Tags:          git.AllTags,
	})
	if err != nil {
		defer os.RemoveAll(dir)
		return "", fmt.Errorf("Failed to clone signatures repository, %w", err)
	}

	// TODO give a valid error if the version is not REMOVE ME
	if signatureVersion != "" {
		// Get the working tree so we can change refs
		// TODO figure this out REMOVE ME
		tree, err := repo.Worktree()
		if err != nil {
			log.Error(err.Error())
		}

		// Set the tag to the signatures version that we want to use
		// TODO fix this REMOVE ME
		tagName := string(signatureVersion)

		// Checkout our tag
		// TODO way are we using a tag here is we only checkout master
		// TODO fix this
		err = tree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/tags/" + tagName),
		})
		if err != nil {
			return "", fmt.Errorf("Requested version not available. Please enter a valid version")
		}
	}
	return dir, nil
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

	// final resting place for the signatures
	rPath := viper.GetString("signatures-path")

	// ensure we have the proper home directory
	var err error
	rPath, err = util.SetHomeDir(rPath)
	if err != nil {
		// TODO-RV: Do something more?
		log.Error(err.Error())
	}

	// if the signatures path does not exist then we create it
	if !util.PathExists(rPath, log) {

		err := os.MkdirAll(rPath, 0700)
		if err != nil {
			log.Error(err.Error())
		}
	}

	// if we want to test the signatures before we install them
	// TODO need to implement something here
	if viper.GetBool("test-signatures") {

		// if the tests pass then we install the signatures
		if executeTests(rRepo) {

			// copy the files from the temp directory to the signatures directory
			if err := cp.Copy(tempSignaturesDir, rPath); err != nil {
				log.Error(err.Error())
				return false
			}

			// get all the files in the signatures directory
			files, err := os.ReadDir(rPath)
			if err != nil {
				log.Error(err.Error())
				return false
			}

			// set them to the current user and the proper permissions
			for _, f := range files {
				if err := os.Chmod(rPath+"/"+f.Name(), 0644); err != nil {
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
	if err := cp.Copy(tempSignaturesDir, rPath); err != nil {
		log.Error(err.Error())
		return false
	}

	// get all the files in the signatures directory
	files, err := os.ReadDir(rPath)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	// set them to the current user and the proper permissions
	// TODO ensure these are .yaml somehow
	for _, f := range files {
		sFileExt := filepath.Ext(rPath + "/" + f.Name())
		if sFileExt == "yml" || sFileExt == "yaml" {
			if err := os.Chmod(rPath+"/"+f.Name(), 0644); err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}
	// TODO why is the commented out
	// TODO Cleanup after ourselves and remove any temp garbage
	// os.RemoveAll(tempSignaturesDir)
	return true
}

// executeTests will run any tests associated with the expressions
// TODO deal with this
func executeTests(dir string) bool {

	// run some tests here and return a true/false depending on the outcome
	return true
}
