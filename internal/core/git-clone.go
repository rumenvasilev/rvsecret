// Package core represents the core functionality of all commands
package core

import (
	"fmt"
	"os"

	"github.com/rumenvasilev/rvsecret/internal/config"
	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// GithubRepository holds the necessary information for a repository,
// this data is specific to Github.
type GithubRepository struct {
	Owner         *string
	ID            *int64
	Name          *string
	FullName      *string
	CloneURL      *string
	URL           *string
	DefaultBranch *string
	Description   *string
	Homepage      *string
}

// CloneConfiguration holds the configurations for cloning a repo
type CloneConfiguration struct { // Alignement of the struct is memory optimized
	URL        string
	Username   string
	Token      string
	Branch     string
	TagMode    git.TagMode
	Depth      int
	InMemClone bool
	Tag        bool
}

// cloneRepository will clone a given repository based upon a configured set or options a user provides.
// This is a catchall for all different types of repos and create a single entry point for cloning a repo.
func cloneRepository(cfg *config.Config, statsIncrementer func(), repo _coreapi.Repository) (*git.Repository, string, error) {
	clone, path, err := cloneRepositoryWrapper(cfg, repo)
	if err != nil {
		switch err.Error() {
		case "remote repository is empty":
			statsIncrementer()
			return nil, "", fmt.Errorf("failed cloning repository %s, it is empty, %w", repo.CloneURL, err)
		default:
			return nil, "", fmt.Errorf("failed cloning repository %s, %w", repo.CloneURL, err)
		}
	}
	statsIncrementer()
	return clone, path, err
}

func cloneRepositoryWrapper(cfg *config.Config, repo _coreapi.Repository) (*git.Repository, string, error) {
	var cloneConfig CloneConfiguration
	var auth = http.BasicAuth{}
	switch cfg.Global.ScanType {
	case api.Github, api.GithubEnterprise:
		cloneConfig = CloneConfiguration{
			URL:        repo.CloneURL,
			Branch:     repo.DefaultBranch,
			Depth:      cfg.Global.CommitDepth,
			InMemClone: cfg.Global.InMemClone,
		}
		auth.Username = "doesn't matter"
		auth.Password = cfg.Github.APIToken
	// case api.UpdateSignatures: This is handled through a different path
	case api.Gitlab:
		cloneConfig = CloneConfiguration{
			URL:        repo.CloneURL,
			Branch:     repo.DefaultBranch,
			Depth:      cfg.Global.CommitDepth,
			InMemClone: cfg.Global.InMemClone,
			// Token:      , // TODO Is this need since we already have a client?
		}
		auth.Username = "oauth2"
		auth.Password = cfg.GitlabAccessToken
	case api.LocalGit:
		cloneConfig = CloneConfiguration{
			URL:        repo.CloneURL,
			Branch:     repo.DefaultBranch,
			Depth:      cfg.Global.CommitDepth,
			InMemClone: cfg.Global.InMemClone,
		}
	default:
		return nil, "", fmt.Errorf("unsupported scantype")
	}
	return CloneRepositoryGeneric(cloneConfig, &auth)
}

// cloneRepositoryGeneric will create either an in memory clone of a given repository or clone to a temp dir.
func CloneRepositoryGeneric(config CloneConfiguration, auth *http.BasicAuth) (repo *git.Repository, dir string, err error) {
	ref := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", config.Branch))
	if config.Tag {
		ref = plumbing.ReferenceName(fmt.Sprintf("refs/tags/%s", config.Branch))
	}
	cloneOptions := &git.CloneOptions{
		URL:           config.URL,
		Depth:         config.Depth,
		ReferenceName: ref,
		SingleBranch:  true,
		Tags:          config.TagMode,
	}

	if auth != nil {
		cloneOptions.Auth = auth
	}

	if config.TagMode == git.InvalidTagMode {
		cloneOptions.Tags = git.NoTags
	}

	if !config.InMemClone {
		dir, err = os.MkdirTemp("", "rvsecret")
		if err != nil {
			return nil, "", err
		}
		repo, err = git.PlainClone(dir, false, cloneOptions)
	} else {
		repo, err = git.Clone(memory.NewStorage(), nil, cloneOptions)
	}
	if err != nil {
		return nil, dir, err
	}
	return repo, dir, nil
}
