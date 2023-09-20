package finding

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/util"
)

const (
	githubURL string = "https://github.com"
	gitlabURL string = "https://gitlab.com"
)

// Finding is a secret that has been discovered within a target by a discovery method
type Finding struct {
	Action           string
	AppVersion       string
	Content          string
	CommitAuthor     string
	CommitHash       string
	CommitMessage    string
	CommitURL        string
	Description      string
	FilePath         string
	FileURL          string
	Hash             string
	LineNumber       string
	RepositoryName   string
	RepositoryOwner  string
	RepositoryURL    string
	SignatureID      string
	SignatureVersion string
	SecretID         string
}

// Initialize will set the urls and create an ID for inclusion within the finding
func (f *Finding) Initialize(scanType api.ScanType, gheURL string) error {
	if f == nil {
		return errors.New("finding is uninitialized")
	}
	f.setupUrls(scanType, gheURL)
	return nil
}

// setupUrls will set the urls used to search through either github or gitlab for inclusion in the finding data
func (f *Finding) setupUrls(scanType api.ScanType, gheURL string) {
	var baseURL string
	switch scanType {
	// case api.GithubEnterprise:
	// baseURL = sess.GithubEnterpriseURL
	// SHOULD THIS BE THIS WAY?
	// f.RepositoryURL = fmt.Sprintf("%s/%s/%s", baseURL, f.RepositoryOwner, f.RepositoryName)
	// f.FileURL = fmt.Sprintf("%s/blob/%s/%s", f.RepositoryURL, f.CommitHash, f.FilePath)
	// f.CommitURL = fmt.Sprintf("%s/commit/%s", f.RepositoryURL, f.CommitHash)
	case api.Github, api.GithubEnterprise:
		baseURL = githubURL
		// TODO: IS THIS CORRECT??
		if scanType == api.GithubEnterprise {
			baseURL = gheURL
		}
		f.RepositoryURL = fmt.Sprintf("%s/%s/%s", baseURL, f.RepositoryOwner, f.RepositoryName)
		f.FileURL = fmt.Sprintf("%s/blob/%s/%s", f.RepositoryURL, f.CommitHash, f.FilePath)
		f.CommitURL = fmt.Sprintf("%s/commit/%s", f.RepositoryURL, f.CommitHash)
	case "gitlab":
		baseURL = gitlabURL
		results := util.CleanURLSpaces(f.RepositoryOwner, f.RepositoryName)
		f.RepositoryURL = fmt.Sprintf("%s/%s/%s", baseURL, results[0], results[1])
		f.FileURL = fmt.Sprintf("%s/blob/%s/%s", f.RepositoryURL, f.CommitHash, f.FilePath)
		f.CommitURL = fmt.Sprintf("%s/commit/%s", f.RepositoryURL, f.CommitHash)
	}
}

func (f *Finding) RealtimeOutput(cfg config.Global, log *log.Logger) {
	if !cfg.Silent && !cfg.CSVOutput && !cfg.JSONOutput {
		log.Warn(" %s", strings.ToUpper(f.Description))
		log.Info("  SignatureID..........: %s", f.SignatureID)
		log.Info("  Repo.................: %s", f.RepositoryName)
		log.Info("  File Path............: %s", f.FilePath)
		log.Info("  Line Number..........: %s", f.LineNumber)
		log.Info("  Message..............: %s", util.TruncateString(f.CommitMessage, 100))
		log.Info("  Commit Hash..........: %s", util.TruncateString(f.CommitHash, 100))
		log.Info("  Author...............: %s", f.CommitAuthor)
		log.Info("  SecretID.............: %v", f.SecretID)
		log.Info("  App Version..........: %s", f.AppVersion)
		log.Info("  Signatures Version...: %v", f.SignatureVersion)
		if len(f.Content) > 0 {
			issues := "\n\t" + f.Content
			log.Info("  Issues..........: %s", issues)
		}

		log.Info(" ------------------------------------------------")
		log.Info("")
	}
}
