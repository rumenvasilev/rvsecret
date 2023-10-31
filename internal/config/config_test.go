package config

import (
	"os"
	"testing"

	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"github.com/rumenvasilev/rvsecret/version"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// func TestSetConfig(t *testing.T) {
// 	type args struct {
// 		cmd *cobra.Command
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want Config
// 	}{
// 		{"default", args{cmd: &cobra.Command{}}, getTestDefaultConfig(t, false)},
// 		{"new-config-file", args{cmd: &cobra.Command{}}, getTestDefaultConfig(t, true)},
// 	}

// 	e := os.Getenv("HOME")
// 	c := DefaultConfig
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			defer resetConfig(e, c)
// 			w := tt.want
// 			SetConfig(tt.args.cmd)
// 			assert.Equal(t, w, *cfg)
// 			os.Setenv("HOME", e)
// 		})
// 	}
// }

func TestSetConfigDefault(t *testing.T) {
	e := os.Getenv("HOME")
	c := DefaultConfig
	defer resetConfig(e, c)
	w := getTestDefaultConfig(t, false)
	SetConfig(&cobra.Command{})
	assert.Equal(t, w, *cfg)
	os.Setenv("HOME", e)
}

func TestSetConfigWithConfig(t *testing.T) {
	e := os.Getenv("HOME")
	c := DefaultConfig
	defer resetConfig(e, c)
	w := getTestDefaultConfig(t, true)
	SetConfig(&cobra.Command{})
	assert.Equal(t, w, *cfg)
	os.Setenv("HOME", e)
}

func getTestDefaultConfig(t *testing.T, cfgFile bool) Config {
	t.Setenv("HOME", "/tmp")
	if cfgFile {
		// create tmp config file
		f, err := os.CreateTemp("", "*.yaml")
		if err != nil {
			// can't create tmp dir for test
			panic(err)
		}

		tmp := DefaultConfig
		tmp.Global.ConfigFile = f.Name()
		tmp.Github.APIToken = "testing"
		// populate with settings from default
		f.Write([]byte(tmp.toYaml()))
		f.Close()
		DefaultConfig.Global.ConfigFile = f.Name()
		DefaultConfig.Github.APIToken = "testing"
	}
	c := DefaultConfig
	c.Global.SkippablePath = util.AppendToSlice(true, defaultIgnorePaths, global.SkippablePath)
	c.Global.SkippableExt = util.AppendToSlice(false, defaultIgnoreExtensions, global.SkippableExt)
	c.Global.CommitDepth = setCommitDepth(global.CommitDepth)
	c.Global.AppVersion = version.AppVersion()
	return c
}

func resetConfig(e string, c Config) {
	DefaultConfig = c
	os.Setenv("HOME", e)
}

func TestLoad(t *testing.T) {
	type args struct {
		scanType api.ScanType
		cfg      Config
	}
	tests := []struct {
		name    string
		args    args
		want    Config
		wantErr string
	}{
		{"LocalGit", args{api.LocalGit, Config{}}, Config{Global: Global{ScanType: api.LocalGit}}, ""},
		{"LocalPath", args{api.LocalPath, Config{}}, Config{Global: Global{ScanType: api.LocalPath}}, ""},
		{"Github", args{api.Github, Config{Github: Github{APIToken: "la"}}}, Config{Global: Global{ScanType: api.Github}, Github: Github{APIToken: "la"}}, ""},
		{"Github_E", args{api.Github, Config{Github: Github{APIToken: ""}}}, Config{}, "APIToken for Github is not set"},
		{"GithubEnterprise", args{api.GithubEnterprise, Config{Github: Github{APIToken: "la", GithubEnterpriseURL: "random"}}}, Config{Global: Global{ScanType: api.GithubEnterprise}, Github: Github{APIToken: "la", GithubEnterpriseURL: "random"}}, ""},
		{"GithubEnterprise_E", args{api.GithubEnterprise, Config{Github: Github{APIToken: "la", GithubEnterpriseURL: ""}}}, Config{}, "Github enterprise URL is not set"},
		{"Gitlab", args{api.Gitlab, Config{Gitlab: Gitlab{APIToken: "bla"}}}, Config{Global: Global{ScanType: api.Gitlab}, Gitlab: Gitlab{APIToken: "bla"}}, ""},
		{"Gitlab_E", args{api.Gitlab, Config{Gitlab: Gitlab{APIToken: ""}}}, Config{}, "APIToken for Gitlab is not set"},
		{"UpdateSignatures", args{api.UpdateSignatures, Config{Signatures: Signatures{APIToken: "bla"}}}, Config{Global: Global{ScanType: api.UpdateSignatures}, Signatures: Signatures{APIToken: "bla"}}, ""},
		{"UpdateSignatures_E", args{api.UpdateSignatures, Config{Signatures: Signatures{APIToken: ""}}}, Config{}, "APIToken for Github is not set"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg = &tt.args.cfg
			got, err := Load(tt.args.scanType)
			if tt.wantErr != "" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, *got, tt.want)
			}
		})
	}
}

func Test_setCommitDepth(t *testing.T) {
	type args struct {
		c int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"5", args{c: 5}, 5},
		{"-800", args{c: -800}, 9999999999},
		{"0", args{c: 0}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := setCommitDepth(tt.args.c)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrintDebug(t *testing.T) {
	type args struct {
		signatureVersion string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"14", args{signatureVersion: "14"}, "\nDebug Info\nApp version..............\nSignatures version.......14\nScanning tests...........false\nMax file size............0\nJSON output..............false\nCSV output...............false\nSilent output............false\nWeb server enabled.......false\n"},
		{"seven", args{signatureVersion: "seven"}, "\nDebug Info\nApp version..............\nSignatures version.......seven\nScanning tests...........false\nMax file size............0\nJSON output..............false\nCSV output...............false\nSilent output............false\nWeb server enabled.......false\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrintDebug(tt.args.signatureVersion)
			assert.Equal(t, tt.want, got)
		})
	}
}
