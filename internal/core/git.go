// Package core represents the core functionality of all commands
package core

import (
	"errors"
	"fmt"
	"os"

	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	_github "github.com/rumenvasilev/rvsecret/internal/core/provider/github"
	_gitlab "github.com/rumenvasilev/rvsecret/internal/core/provider/gitlab"
	"github.com/rumenvasilev/rvsecret/internal/pkg/api"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"gopkg.in/src-d/go-git.v4/utils/merkletrie"
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
type CloneConfiguration struct {
	InMemClone bool
	URL        string
	Username   string
	Token      string
	Branch     string
	Depth      int
}

// EmptyTreeCommit is a dummy commit id used as a placeholder and for testing
const (
	EmptyTreeCommitID = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
)

// GetParentCommit will get the parent commit from a specific point. If the current commit
// has no parents then it will create a dummy commit.
func getParentCommit(commit *object.Commit, repo *git.Repository) (*object.Commit, error) {
	if commit.NumParents() == 0 {
		parentCommit, err := repo.CommitObject(plumbing.NewHash(EmptyTreeCommitID))
		if err != nil {
			return nil, err
		}
		return parentCommit, nil
	}
	parentCommit, err := commit.Parents().Next()
	if err != nil {
		return nil, err
	}
	return parentCommit, nil
}

// GetRepositoryHistory gets the commit history of a repository
func GetRepositoryHistory(repository *git.Repository) ([]*object.Commit, error) {
	var commits []*object.Commit
	ref, err := repository.Head()
	if err != nil {
		return nil, err
	}
	cIter, err := repository.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, err
	}
	_ = cIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})
	return commits, nil
}

// GetChanges will get the changes between to specific commits. It grabs the parent commit of
// the one being passed and uses that to fetch the tree for that commit. If no commit is found,
// it will create a fake on. It then takes that parent tree along with the tree for the commit
// passed in and does a diff
func GetChanges(commit *object.Commit, repo *git.Repository) (object.Changes, error) {
	parentCommit, err := getParentCommit(commit, repo)
	if err != nil {
		return nil, err
	}

	commitTree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	parentCommitTree, err := parentCommit.Tree()
	if err != nil {
		return nil, err
	}

	changes, err := object.DiffTree(parentCommitTree, commitTree)
	if err != nil {
		return nil, err
	}
	return changes, nil
}

// GetChangeAction returns a more condensed and user friendly action for further reference
func GetChangeAction(change *object.Change) string {
	action, err := change.Action()
	if err != nil {
		return "Unknown"
	}

	return action.String()
}

// GetChangePath will set the action of the commit for further action
func GetChangePath(change *object.Change) string {
	action, err := change.Action()
	if err != nil {
		return change.To.Name
	}

	if action == merkletrie.Delete {
		return change.From.Name
	}
	return change.To.Name

}

// GetChangeContent will get the contents of a git change or patch.
func GetChangeContent(change *object.Change) (result string, contentError error) {
	//temporary response to:  https://github.com/sergi/go-diff/issues/89
	defer func() {
		if err := recover(); err != nil {
			contentError = errors.New("panic occurred while retrieving change content: ")
		}
	}()
	patch, err := change.Patch()
	if err != nil {
		return "", err
	}
	for _, filePatch := range patch.FilePatches() {
		if filePatch.IsBinary() {
			continue
		}
		for _, chunk := range filePatch.Chunks() {
			result += chunk.Content()
		}
	}
	return result, nil
}

// InitGitClient will create a new git client of the type given by the input string.
func (s *Session) InitGitClient() error {
	switch s.Config.ScanType {
	case api.Github, api.GithubEnterprise:
		client, err := _github.NewClient(s.Config.GithubAccessToken, "", s.Out)
		if err != nil {
			return err
		}
		if s.Config.ScanType == api.GithubEnterprise {
			if s.Config.GithubEnterpriseURL == "" {
				return fmt.Errorf("github enterprise URL is missing")
			}
			client, err = _github.NewClient(s.Config.GithubAccessToken, s.Config.GithubEnterpriseURL, s.Out)
			if err != nil {
				return err
			}
		}
		s.Client = client
	case api.Gitlab:
		client, err := _gitlab.NewClient(s.Config.GitlabAccessToken, s.Out)
		if err != nil {
			return fmt.Errorf("error initializing GitLab client: %s", err)
		}
		// TODO need to add in the bits to parse the url here as well
		// TODO set this to some sort of consistent client, look to github for ideas
		s.Client = client
	}
	return nil
}

// cloneRepository will clone a given repository based upon a configured set or options a user provides.
// This is a catchall for all different types of repos and create a single entry point for cloning a repo.
func cloneRepository(sess *Session, repo _coreapi.Repository) (*git.Repository, string, error) {
	clone, path, err := cloneRepositoryFunc(sess, repo)
	if err != nil {
		switch err.Error() {
		case "remote repository is empty":
			sess.State.Stats.IncrementRepositoriesCloned()
			return nil, "", fmt.Errorf("failed cloning repository %s, it is empty, %w", repo.CloneURL, err)
		default:
			return nil, "", fmt.Errorf("failed cloning repository %s, %w", repo.CloneURL, err)
		}
	}
	sess.State.Stats.IncrementRepositoriesCloned()
	return clone, path, err
}

func cloneRepositoryFunc(sess *Session, repo _coreapi.Repository) (*git.Repository, string, error) {
	var cloneConfig = CloneConfiguration{}
	var auth = http.BasicAuth{}
	switch sess.Config.ScanType {
	case api.Github, api.GithubEnterprise:
		cloneConfig = CloneConfiguration{
			URL:        repo.CloneURL,
			Branch:     repo.DefaultBranch,
			Depth:      sess.Config.CommitDepth,
			InMemClone: sess.Config.InMemClone,
			// Token:      sess.GithubAccessToken,
		}
		auth.Username = "doesn't matter"
		auth.Password = sess.Config.GithubAccessToken
	case api.Gitlab:
		cloneConfig = CloneConfiguration{
			URL:        repo.CloneURL,
			Branch:     repo.DefaultBranch,
			Depth:      sess.Config.CommitDepth,
			InMemClone: sess.Config.InMemClone,
			// Token:      , // TODO Is this need since we already have a client?
		}
		auth.Username = "oauth2"
		auth.Password = sess.Config.GitlabAccessToken
	case api.LocalGit:
		cloneConfig = CloneConfiguration{
			URL:        repo.CloneURL,
			Branch:     repo.DefaultBranch,
			Depth:      sess.Config.CommitDepth,
			InMemClone: sess.Config.InMemClone,
		}
	}
	return cloneRepositoryGeneric(cloneConfig, &auth)
}

// cloneRepositoryGeneric will create either an in memory clone of a given repository or clone to a temp dir.
func cloneRepositoryGeneric(config CloneConfiguration, auth *http.BasicAuth) (repo *git.Repository, dir string, err error) {
	cloneOptions := &git.CloneOptions{
		URL:           config.URL,
		Depth:         config.Depth,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", config.Branch)),
		SingleBranch:  true,
		Tags:          git.NoTags,
	}

	if auth != nil {
		cloneOptions.Auth = auth
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
