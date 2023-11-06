package finding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getCSVHeader(t *testing.T) {
	got := getCSVHeader()
	want := []string{"FilePath", "Line Number", "Action", "Description", "SignatureID", "Finding List", "Repo Owner", "Repo Name", "Commit Hash", "Commit Message", "Commit Author", "File URL", "Secret ID", "App Version", "Signatures Version"}
	assert.Equal(t, want, got)
}

func TestToCSV(t *testing.T) {
	f := makeTestFinding()
	got := f.toCSV()
	want := makeTestFindingStringSlice()
	assert.Equal(t, want, got)
}

func makeTestFinding() *Finding {
	return &Finding{
		"f.Action",
		"f.AppVersion",
		"f.Content",
		"f.CommitAuthor",
		"f.CommitHash",
		"f.CommitMessage",
		"f.CommitURL",
		"f.Description",
		"f.FilePath",
		"f.FileURL",
		"f.Hash",
		"f.LineNumber",
		"f.RepositoryName",
		"f.RepositoryOwner",
		"f.RepositoryURL",
		"f.SignatureID",
		"f.SignatureVersion",
		"f.SecretID",
	}
}

func makeTestFindingStringSlice() []string {
	return []string{
		"f.FilePath",
		"f.LineNumber",
		"f.Action",
		"f.Description",
		"f.SignatureID",
		"f.Content",
		"f.RepositoryOwner",
		"f.RepositoryName",
		"f.CommitHash",
		"f.CommitMessage",
		"f.CommitAuthor",
		"f.FileURL",
		"f.SecretID",
		"f.AppVersion",
		"f.SignatureVersion"}
}
