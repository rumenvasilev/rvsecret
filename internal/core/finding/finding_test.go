package finding

import (
	"testing"

	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/stretchr/testify/assert"
)

func TestFinding_Initialize(t *testing.T) {
	type args struct {
		finding  *Finding
		scanType api.ScanType
		gheURL   string
	}
	tests := []struct {
		name    string
		args    args
		want    *Finding
		wantErr bool
	}{
		{"github", args{&Finding{}, api.Github, "doesnt-matter"}, &Finding{CommitURL: "https://github.com///commit/", FileURL: "https://github.com///blob//", RepositoryURL: "https://github.com//"}, false},
		{"github_enterprise", args{&Finding{}, api.GithubEnterprise, "gitfake.enterprise"}, &Finding{CommitURL: "gitfake.enterprise///commit/", FileURL: "gitfake.enterprise///blob//", RepositoryURL: "gitfake.enterprise//"}, false},
		{"gitlab", args{&Finding{}, api.Gitlab, "doesnt-matter"}, &Finding{CommitURL: "https://gitlab.com///commit/", FileURL: "https://gitlab.com///blob//", RepositoryURL: "https://gitlab.com//"}, false},
		{"unsupported", args{&Finding{}, api.Unknown, ""}, &Finding{}, false},
		{"uninitialized", args{nil, "random string", ""}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.args.finding
			err := f.Initialize(tt.args.scanType, tt.args.gheURL)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, *tt.want, *f)
			}
		})
	}
}

func TestGetFindingsCSVHeader(t *testing.T) {
	got := GetFindingsCSVHeader()
	want := []string{"FilePath", "Line Number", "Action", "Description", "SignatureID", "Finding List", "Repo Owner", "Repo Name", "Commit Hash", "Commit Message", "Commit Author", "File URL", "Secret ID", "App Version", "Signatures Version"}
	assert.Equal(t, want, got)
}

func TestToCSV(t *testing.T) {
	f := makeTestFinding()
	got := f.ToCSV()
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
