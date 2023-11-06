package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core/provider/api"
	"github.com/rumenvasilev/rvsecret/internal/core/provider/github"
	"github.com/rumenvasilev/rvsecret/internal/core/provider/gitlab"
	pkgapi "github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func TestInitGitClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		want    api.IClient
		wantErr string
	}{
		{"github", config.Config{Global: config.Global{ScanType: pkgapi.Github}, Github: config.Github{APIToken: alphabet[:40]}}, &github.Client{}, ""},
		{"github, error, no token", config.Config{Global: config.Global{ScanType: pkgapi.Github}}, &github.Client{}, "cannot create new Github client, The token is invalid. Please use a valid Github token."},
		{"github enterprise, error, url missing", config.Config{Global: config.Global{ScanType: pkgapi.GithubEnterprise}}, &github.Client{}, "github enterprise URL is missing"},
		{"github enterprise, error, no token", config.Config{Global: config.Global{ScanType: pkgapi.GithubEnterprise}, Github: config.Github{GithubEnterpriseURL: "fake"}}, &github.Client{}, "cannot create new Github client, The token is invalid. Please use a valid Github token."},
		{"github enterprise", config.Config{Global: config.Global{ScanType: pkgapi.GithubEnterprise}, Github: config.Github{GithubEnterpriseURL: "fake", APIToken: alphabet[:40]}}, &github.Client{}, ""},
		{"gitlab", config.Config{Global: config.Global{ScanType: pkgapi.Gitlab}, Gitlab: config.Gitlab{APIToken: alphabet[:20]}}, &gitlab.Client{}, ""},
		{"gitlab, error, no token", config.Config{Global: config.Global{ScanType: pkgapi.Gitlab}}, &gitlab.Client{}, "Gitlab token is invalid"},
		{"UpdateSignatures", config.Config{Global: config.Global{ScanType: pkgapi.UpdateSignatures}, Signatures: config.Signatures{APIToken: alphabet[:40]}}, &github.Client{}, ""},
		{"UpdateSignatures, error, no token", config.Config{Global: config.Global{ScanType: pkgapi.UpdateSignatures}}, &github.Client{}, "cannot create new Github client, The token is invalid. Please use a valid Github token."},
		{"error, unknown scan type", config.Config{}, &github.Client{}, "unknown scan type provided"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitGitClient(&tt.cfg)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, tt.want, got)
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("InitGitClient() = %v, want %v", got, tt.want)
			// }
		})
	}
}
