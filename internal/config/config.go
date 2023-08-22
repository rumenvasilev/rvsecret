package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/rumenvasilev/rvsecret/internal/pkg/api"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"github.com/rumenvasilev/rvsecret/version"
	"github.com/spf13/viper"
)

// cfg holds the configuration data the commands
var cfg *viper.Viper

// defaultIgnoreExtensions is an array of extensions that if they match a file that file will be excluded
var defaultIgnoreExtensions = []string{"jpg", "jpeg", "png", "gif", "bmp", "tiff",
	"tif", "psd", "xcf"}

// defaultIgnorePaths is an array of directories that will be excluded from all types of scans.
var defaultIgnorePaths = []string{"node_modules/", "vendor/bundle", "vendor/cache", "/proc/"}

// DefaultValues is a map of all flag default values and other mutable variables
var defaultValues = map[string]interface{}{
	"bind-address":       "127.0.0.1",
	"bind-port":          9393,
	"commit-depth":       -1,
	"config-file":        "$HOME/.rvsecret/config.yaml",
	"csv":                false,
	"debug":              false,
	"ignore-extension":   nil,
	"ignore-path":        nil,
	"in-mem-clone":       false,
	"json":               false,
	"max-file-size":      10,
	"num-threads":        -1,
	"local-paths":        nil,
	"scan-forks":         false,
	"scan-tests":         false,
	"scan-type":          "",
	"silent":             false,
	"confidence-level":   3,
	"signature-file":     "$HOME/.rvsecret/signatures/default.yaml",
	"signature-path":     "$HOME/.rvsecret/signatures/",
	"signatures-path":    "$HOME/.rvsecret/signatures/",
	"signatures-url":     "https://github.com/rumenvasilev/rvsecret-signatures",
	"signatures-version": "",
	"scan-dir":           nil,
	"scan-file":          nil,
	"hide-secrets":       false,
	"rules-url":          "",
	"test-signatures":    false,
	"web-server":         false,
	// Github
	"add-org-members":       false,
	"github-enterprise-url": "",
	"github-url":            "https://api.github.com",
	"github-api-token":      "",
	"github-orgs":           nil,
	"github-repos":          nil,
	"github-users":          nil,
	// Gitlab
	"gitlab-targets": nil,
	//"gitlab-url":                 "", // TODO set the default
	"gitlab-api-token": "",
}

// SetConfig will set the defaults, and load a config file and environment variables if they are present
func SetConfig() {
	for key, value := range defaultValues {
		viper.SetDefault(key, value)
	}

	configFile := viper.GetString("config-file")

	if configFile != defaultValues["config-file"] {
		viper.SetConfigFile(configFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home + "/.rvsecret/")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Couldn't load Viper config.", err)
		// os.Exit(1)
	}

	viper.AutomaticEnv()

	cfg = viper.GetViper()
}

type Config struct {
	AppVersion          string
	BindAddress         string
	BindPort            int
	CommitDepth         int
	ConfidenceLevel     int
	CSVOutput           bool
	Debug               bool
	ExpandOrgs          bool
	GithubAccessToken   string
	GithubEnterpriseURL string
	GitlabAccessToken   string
	GitlabTargets       []string
	GitlabURL           string
	HideSecrets         bool
	InMemClone          bool
	JSONOutput          bool
	LocalPaths          []string
	MaxFileSize         int64
	ScanFork            bool
	ScanTests           bool
	ScanType            api.ScanType
	Silent              bool
	SkippableExt        []string
	SkippablePath       []string
	SignatureFiles      []string
	Threads             int
	UserDirtyNames      []string
	UserDirtyOrgs       []string
	UserDirtyRepos      []string
	WebServer           bool
}

// TODO detect scanType automatically
func Load(scanType api.ScanType) (*Config, error) {
	c := Config{
		BindAddress:     cfg.GetString("bind-address"),
		BindPort:        cfg.GetInt("bind-port"),
		CommitDepth:     setCommitDepth(cfg.GetFloat64("commit-depth")),
		CSVOutput:       cfg.GetBool("csv"),
		Debug:           cfg.GetBool("debug"),
		ExpandOrgs:      cfg.GetBool("expand-orgs"),
		HideSecrets:     cfg.GetBool("hide-secrets"),
		InMemClone:      cfg.GetBool("in-mem-clone"),
		JSONOutput:      cfg.GetBool("json"),
		MaxFileSize:     cfg.GetInt64("max-file-size"),
		ConfidenceLevel: cfg.GetInt("confidence-level"),
		ScanFork:        cfg.GetBool("scan-forks"),
		ScanTests:       cfg.GetBool("scan-tests"),
		ScanType:        scanType,
		Silent:          cfg.GetBool("silent"),
		SignatureFiles:  cfg.GetStringSlice("signature-file"),
		Threads:         cfg.GetInt("num-threads"),
		AppVersion:      version.AppVersion(),
		WebServer:       cfg.GetBool("web-server"),
	}

	switch scanType {
	case api.LocalGit:
		c.LocalPaths = cfg.GetStringSlice("local-repos")
	case api.LocalPath:
		c.LocalPaths = cfg.GetStringSlice("paths")
	case api.Gitlab:
		c.GitlabAccessToken = cfg.GetString("gitlab-api-token")
		c.GitlabTargets = cfg.GetStringSlice("gitlab-targets")
	case api.Github, api.GithubEnterprise:
		c.GithubAccessToken = cfg.GetString("github-api-token")
		c.UserDirtyRepos = cfg.GetStringSlice("github-repos")
		c.UserDirtyOrgs = cfg.GetStringSlice("github-orgs")
		c.UserDirtyNames = cfg.GetStringSlice("github-users")
		if scanType == api.GithubEnterprise {
			c.GithubEnterpriseURL = cfg.GetString("github-enterprise-url")
			if c.GithubEnterpriseURL == "" {
				return nil, errors.New("Github enterprise URL is not set.")
			}
		}
	}

	// Add the default directories to the sess if they don't already exist
	c.SkippablePath = util.AppendToSlice(true, defaultIgnorePaths, c.SkippablePath)
	// add any additional paths the user requested to exclude to the pre-defined slice
	c.SkippablePath = util.AppendToSlice(true, cfg.GetStringSlice("ignore-path"), c.SkippablePath)

	// the default ignorable extensions
	c.SkippableExt = util.AppendToSlice(false, defaultIgnoreExtensions, c.SkippableExt)
	// add any additional extensions the user requested to ignore
	c.SkippableExt = util.AppendToSlice(true, cfg.GetStringSlice("ignore-extension"), c.SkippableExt)

	return &c, nil
}

// setCommitDepth will set the commit depth for the current session. This is an ugly way of doing it
// but for the moment it works fine.
// TODO dynamically acquire the commit depth of a given repo
func setCommitDepth(c float64) int {
	if c == -1 {
		return 9999999999
	}
	return int(c)
}
