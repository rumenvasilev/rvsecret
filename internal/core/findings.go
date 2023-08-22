// Package core represents the core functionality of all commands
package core

import (
	"fmt"
	"strings"

	"github.com/rumenvasilev/rvsecret/internal/pkg/api"
	"github.com/rumenvasilev/rvsecret/internal/util"
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

// setupUrls will set the urls used to search through either github or gitlab for inclusion in the finding data
func (f *Finding) setupUrls(sess *Session) {
	var baseURL string
	switch sess.Config.ScanType {
	// case api.GithubEnterprise:
	// baseURL = sess.GithubEnterpriseURL
	// SHOULD THIS BE THIS WAY?
	// f.RepositoryURL = fmt.Sprintf("%s/%s/%s", baseURL, f.RepositoryOwner, f.RepositoryName)
	// f.FileURL = fmt.Sprintf("%s/blob/%s/%s", f.RepositoryURL, f.CommitHash, f.FilePath)
	// f.CommitURL = fmt.Sprintf("%s/commit/%s", f.RepositoryURL, f.CommitHash)
	case api.Github, api.GithubEnterprise:
		baseURL = "https://github.com"
		// TODO: IS THIS CORRECT??
		if sess.Config.ScanType == api.GithubEnterprise {
			baseURL = sess.Config.GithubEnterpriseURL
		}
		f.RepositoryURL = fmt.Sprintf("%s/%s/%s", baseURL, f.RepositoryOwner, f.RepositoryName)
		f.FileURL = fmt.Sprintf("%s/blob/%s/%s", f.RepositoryURL, f.CommitHash, f.FilePath)
		f.CommitURL = fmt.Sprintf("%s/commit/%s", f.RepositoryURL, f.CommitHash)
	case "gitlab":
		baseURL = "https://gitlab.com"
		results := util.CleanURLSpaces(f.RepositoryOwner, f.RepositoryName)
		f.RepositoryURL = fmt.Sprintf("%s/%s/%s", baseURL, results[0], results[1])
		f.FileURL = fmt.Sprintf("%s/blob/%s/%s", f.RepositoryURL, f.CommitHash, f.FilePath)
		f.CommitURL = fmt.Sprintf("%s/commit/%s", f.RepositoryURL, f.CommitHash)
	}
}

// Initialize will set the urls and create an ID for inclusion within the finding
func (f *Finding) Initialize(sess *Session) {
	f.setupUrls(sess)
}

func (f *Finding) RealtimeOutput(sess *Session) {
	if !sess.Config.Silent && !sess.Config.CSVOutput && !sess.Config.JSONOutput {
		log := sess.Out
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
