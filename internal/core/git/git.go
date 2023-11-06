package git

import (
	"errors"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/utils/merkletrie"
)

// EmptyTreeCommit is a dummy commit id used as a placeholder and for testing
const (
	EmptyTreeCommitID = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
)

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

// GetParentCommit will get the parent commit from a specific point. If the current commit
// has no parents then it will create a dummy commit.
func getParentCommit(commit *object.Commit, repo *git.Repository) (*object.Commit, error) {
	if commit.NumParents() == 0 {
		return repo.CommitObject(plumbing.NewHash(EmptyTreeCommitID))
	}

	return commit.Parents().Next()
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
