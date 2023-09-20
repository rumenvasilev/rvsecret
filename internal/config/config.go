package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/structs"
	"github.com/mitchellh/go-homedir"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"github.com/rumenvasilev/rvsecret/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const configFile = "$HOME/.rvsecret/config.yaml"

// cfg holds the configuration data the commands
var cfg *Config

// defaultIgnoreExtensions is an array of extensions that if they match a file that file will be excluded
var defaultIgnoreExtensions = []string{"jpg", "jpeg", "png", "gif", "bmp", "tiff",
	"tif", "psd", "xcf", "pdf"}

// defaultIgnorePaths is an array of directories that will be excluded from all types of scans.
var defaultIgnorePaths = []string{"node_modules/", "vendor/bundle", "vendor/cache", "/proc/"}

type Config struct {
	Github     Github     `mapstructure:"github" yaml:"github"`
	Local      Local      `mapstructure:"local" yaml:"local"`
	Signatures Signatures `mapstructure:"signatures" yaml:"signatures"`

	Gitlab `mapstructure:"gitlab" yaml:"gitlab"`
	Global Global `mapstructure:"global" yaml:"global"`
}

type Global struct {
	AppVersion      string       `yaml:"-"`
	BindAddress     string       `mapstructure:"bind-address" yaml:"bind-address"`
	ConfigFile      string       `mapstructure:"config-file" yaml:"-"`
	ScanType        api.ScanType `mapstructure:"scan-type" yaml:"-"`
	SkippableExt    []string     `mapstructure:"ignore-extension" yaml:"ignore-extension"`
	SkippablePath   []string     `mapstructure:"ignore-path" yaml:"ignore-path"`
	BindPort        int          `mapstructure:"bind-port" yaml:"bind-port"`
	CommitDepth     int          `mapstructure:"commit-depth" yaml:"commit-depth"`
	ConfidenceLevel int          `mapstructure:"confidence-level" yaml:"confidence-level"`
	MaxFileSize     int64        `mapstructure:"max-file-size" yaml:"max-file-size"`
	Threads         int          `mapstructure:"num-threads" yaml:"num-threads"`
	CSVOutput       bool         `mapstructure:"csv"`
	Debug           bool         `mapstructure:"debug"`
	ExpandOrgs      bool         `mapstructure:"expand-orgs" yaml:"expand-orgs"`
	HideSecrets     bool         `mapstructure:"hide-secrets" yaml:"hide-secrets"`
	InMemClone      bool         `mapstructure:"in-mem-clone" yaml:"in-mem-clone"`
	JSONOutput      bool         `mapstructure:"json"`
	ScanFork        bool         `mapstructure:"scan-forks" yaml:"scan-forks"`
	ScanTests       bool         `mapstructure:"scan-tests" yaml:"scan-tests"`
	Silent          bool         `mapstructure:"silent"`
	WebServer       bool         `mapstructure:"web-server" yaml:"web-server"`
	_               [6]byte
}

type Signatures struct {
	APIToken string `mapstructure:"api-token" yaml:"api-token"`
	File     string `mapstructure:"file"`
	Path     string `mapstructure:"path"`
	URL      string `mapstructure:"url"`
	UserRepo string `mapstructure:"user-repo" yaml:"user-repo"`
	Version  string `mapstructure:"version"`
	Test     bool   `mapstructure:"test"`
	_        [31]byte
}

type Github struct {
	APIToken            string   `mapstructure:"api-token" yaml:"api-token"`
	GithubURL           string   `mapstructure:"url" yaml:"url"`
	GithubEnterpriseURL string   `mapstructure:"enterprise-url" yaml:"enterprise-url"`
	UserDirtyNames      []string `mapstructure:"users" yaml:"users"`
	UserDirtyOrgs       []string `mapstructure:"orgs" yaml:"orgs"`
	UserDirtyRepos      []string `mapstructure:"repos" yaml:"repos"`
	Enterprise          bool     `mapstructure:"enterprise"`
	_                   [7]byte
}

type Gitlab struct {
	GitlabAccessToken string   `mapstructure:"gitlab-api-token"`
	GitlabTargets     []string `mapstructure:"gitlab-targets"`
	_                 [24]byte
}

type Local struct {
	Paths []string `mapstructure:"paths"`
	Repos []string `mapstructure:"repos"`
	_     [16]byte
}

// SetConfig will set the defaults, and load a config file and environment variables if they are present
func SetConfig(cmd *cobra.Command) {
	// Set defaults
	for key, value := range structs.Map(DefaultConfig) {
		viper.SetDefault(key, value)
	}

	// Read from config
	cf := viper.GetString("config-file")
	if cf != "" && cf != configFile {
		viper.SetConfigFile(cf)
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

	// Load config into mapstructure
	err := viper.ReadInConfig() //nolint:errcheck
	if err != nil {
		fmt.Printf("Couldn't load config file; proceeding with defaults and CLI overrides, %v\n", err)
	}
	viper.AutomaticEnv()

	// Load mapstructure config into our config struct
	var c *Config
	err = viper.Unmarshal(&c)
	if err != nil {
		fmt.Println("Failed unmarshaling config to struct")
		os.Exit(1)
	}

	c.Global.SkippablePath = util.AppendToSlice(true, defaultIgnorePaths, c.Global.SkippablePath)
	// add any additional paths the user requested to exclude to the pre-defined slice
	// c.SkippablePath = util.AppendToSlice(true, viper.GetStringSlice("ignore-path"), c.SkippablePath)

	// the default ignorable extensions
	c.Global.SkippableExt = util.AppendToSlice(false, defaultIgnoreExtensions, c.Global.SkippableExt)
	// add any additional extensions the user requested to ignore
	// c.SkippableExt = util.AppendToSlice(true, viper.GetStringSlice("ignore-extension"), c.SkippableExt)

	c.Global.CommitDepth = setCommitDepth(c.Global.CommitDepth)

	c.Global.AppVersion = version.AppVersion()
	cfg = c
}

// TODO detect scanType automatically
func Load(scanType api.ScanType) (*Config, error) {
	// set configuration
	cfg.Global.ScanType = scanType

	switch scanType {
	case api.Github:
		if cfg.Github.APIToken == "" {
			return nil, errors.New("APIToken for Github is not set")
		}
	case api.GithubEnterprise:
		if cfg.Github.GithubEnterpriseURL == "" {
			return nil, errors.New("Github enterprise URL is not set.")
		}
	case api.UpdateSignatures:
		if cfg.Signatures.APIToken == "" {
			return nil, errors.New("APIToken for Github is not set")
		}
	}
	return cfg, nil
}

// setCommitDepth will set the commit depth for the current session. This is an ugly way of doing it
// but for the moment it works fine.
// TODO dynamically acquire the commit depth of a given repo
func setCommitDepth(c int) int {
	if c == -1 {
		return 9999999999
	}
	return c
}

// PrintDebug will print a debug header at the start of the session that displays specific setting
func PrintDebug(signatureVersion string) string {
	maxFileSize := cfg.Global.MaxFileSize * 1024 * 1024
	var out string
	out += "\n"
	out += "Debug Info\n"
	out += fmt.Sprintf("App version..............%v\n", cfg.Global.AppVersion)
	out += fmt.Sprintf("Signatures version.......%v\n", signatureVersion)
	out += fmt.Sprintf("Scanning tests...........%v\n", cfg.Global.ScanTests)
	out += fmt.Sprintf("Max file size............%d\n", maxFileSize)
	out += fmt.Sprintf("JSON output..............%v\n", cfg.Global.JSONOutput)
	out += fmt.Sprintf("CSV output...............%v\n", cfg.Global.CSVOutput)
	out += fmt.Sprintf("Silent output............%v\n", cfg.Global.Silent)
	out += fmt.Sprintf("Web server enabled.......%v\n", cfg.Global.WebServer)
	return out
}
