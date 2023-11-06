package signatures

import (
	"fmt"
	"os"

	"github.com/rumenvasilev/rvsecret/internal/core"
	"github.com/rumenvasilev/rvsecret/internal/core/provider"
	"github.com/rumenvasilev/rvsecret/internal/session"
	"gopkg.in/src-d/go-git.v4"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

// fetchSignaturesWithGit will clone the signatures repository and return the local path to it
func fetchSignaturesWithGit(version string, sess *session.Session) (string, error) {
	var err error
	branch := version
	tag := true
	if version == "latest" {
		branch = "stable"
		tag = false
	}
	sess.Client, err = provider.InitGitClient(sess.Config)
	if err != nil {
		return "", err
	}

	// build the URL
	url := sess.Config.Signatures.URL
	if sess.Config.Signatures.UserRepo != "" {
		// TODO this address has to be a const perhaps?
		url = fmt.Sprintf("https://github.com/%s", sess.Config.Signatures.UserRepo)
	}
	// sanitize checks
	cURL, err := cleanInput(url)
	if err != nil {
		return "", err
	}

	cloneCfg := core.CloneConfiguration{
		URL:        cURL,
		Branch:     branch,
		Depth:      sess.Config.Global.CommitDepth,
		InMemClone: sess.Config.Global.InMemClone,
		Tag:        tag,
		// Should we?
		TagMode: git.AllTags,
	}
	auth := &githttp.BasicAuth{
		Username: "egal",
		Password: sess.Config.Signatures.APIToken,
	}

	// If we're gonna use git clone to get a specific tag, we need to pass git.AllTags as parameter here.
	_, dir, err := core.CloneRepositoryGeneric(cloneCfg, auth)
	if err != nil {
		// cleanup dir
		_ = os.RemoveAll(dir)
		return "", err
	}

	return dir, nil
}
