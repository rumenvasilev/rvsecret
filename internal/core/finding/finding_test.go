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
