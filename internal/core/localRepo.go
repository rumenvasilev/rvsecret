package core

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/session"
	"github.com/rumenvasilev/rvsecret/internal/stats"
	"github.com/rumenvasilev/rvsecret/internal/util"

	"gopkg.in/src-d/go-git.v4"
)

// GatherLocalRepositories will grab all the local repos from the user input and generate a repository
// object, putting dummy or generated values in where necessary.
func GatherLocalRepositories(sess *session.Session) error {
	log := log.Log
	// This is the number of targets as we don't do forks or anything else.
	// It will contain directories, that will then be added to the repo count
	// if they contain a .git directory
	// TODO - must use a method, to be memsafe like everything else
	sess.State.Stats.Targets = len(sess.Config.Local.Repos)
	sess.State.Stats.UpdateStatus(stats.StatusGathering)
	log.Important("Gathering Local Repositories...")

	for _, pth := range sess.Config.Local.Repos {

		if !util.PathExists(pth) {
			return fmt.Errorf("[*] <%s> does not exist! Quitting", pth)
		}
		err := pathWalker(pth, sess)
		if err != nil {
			log.Error("something went wrong with filepath walk: %v", err)
			return err
		}
	}
	return nil
}

// pathWalker will walk all the paths and for each path it finds a repository
// will append to the state repository list for later inspection
func pathWalker(pth string, sess *session.Session) error {
	log := log.Log
	// Gather all paths in the tree
	return filepath.Walk(pth, func(path string, f os.FileInfo, err1 error) error {
		if err1 != nil {
			log.Error("Failed to enumerate the path: %s", err1.Error())
			return nil
		}

		// If it is not a directory, exit
		if !f.IsDir() {
			return nil
		}

		// If there is a .git directory then we have a repo
		if filepath.Ext(path) != ".git" {
			return nil
		}

		// Set the url to the relative path of the repo based on the execution path of rvsecret
		repoURL, _ := filepath.Split(path)
		// This is used to id the owner, fullname, and description of the repo. It is ugly but effective. It is the relative path to the repo, for example ../foo
		gitProjName, _ := filepath.Split(repoURL)

		openRepo, err1 := git.PlainOpen(repoURL)
		if err1 != nil {
			// if err1 == git.ErrRepositoryNotExists {
			log.Error(err1.Error())
			// }
			return nil
		}

		ref, err1 := openRepo.Head()
		if err1 != nil {
			log.Error("Failed to open the repo HEAD: %s", err1.Error())
			return nil
		}

		// Get the name of the branch we are working on
		// s := ref.Strings()
		// branchPath := fmt.Sprintf("%s", s[0])
		branchPathParts := strings.Split(ref.Strings()[0], string("refs/heads/"))
		branchName := branchPathParts[len(branchPathParts)-1]

		commit, _ := openRepo.CommitObject(ref.Hash())
		var commitHash = commit.Hash[:]

		// TODO make this a generic function at some point
		// Generate a uid for the repo
		h := sha1.New()
		repoID := fmt.Sprintf("%x", h.Sum(commitHash))

		intRepoID, _ := strconv.ParseInt(repoID, 10, 64)
		// var pRepoID *int64
		// pRepoID = &intRepoID

		// Set the url to the relative path of the repo based on the execution path of rvsecret
		// pRepoURL := &parent

		// pGitProjName := &gitProjName

		// The project name is simply the parent directory in the case of a local scan with all other path bits removed for example ../foo -> foo.
		projectPathParts := strings.Split(gitProjName, string(os.PathSeparator))
		projectName := projectPathParts[len(projectPathParts)-2]

		sessR := _coreapi.Repository{
			Owner:         gitProjName,
			ID:            intRepoID,
			Name:          projectName,
			FullName:      gitProjName,
			CloneURL:      repoURL,
			URL:           repoURL,
			DefaultBranch: branchName,
			Description:   gitProjName,
			Homepage:      repoURL,
		}

		// Add the repo to the sess to be cloned and scanned
		sess.State.AddRepository(&sessR)
		return nil
	})
}
