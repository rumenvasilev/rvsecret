package config

var signatures = Signatures{
	URL:     "https://github.com/rumenvasilev/rvsecret-signatures",
	Version: "latest",
	Test:    true,
	File:    "$HOME/.rvsecret/signatures/default.yaml",
	Path:    "$HOME/.rvsecret/signatures",
}

var github = Github{
	GithubURL:      "https://api.github.com",
	UserDirtyOrgs:  nil,
	UserDirtyNames: nil,
	UserDirtyRepos: nil,
	Enterprise:     false,
}

var gitlab = Gitlab{
	Targets: nil,
}

var global = Global{
	BindAddress: "127.0.0.1",
	// ConfigFile:      "none", // default defined as const in config.go
	BindPort:        9393,
	CommitDepth:     -1,
	ConfidenceLevel: 3,
	MaxFileSize:     10,
	Threads:         -1,
	HideSecrets:     false,
	InMemClone:      false,
	ScanFork:        false,
	ScanTests:       false,
	Silent:          false,
	WebServer:       false,
}

var DefaultConfig = Config{
	Signatures: signatures,
	Github:     github,
	Gitlab:     gitlab,
	Global:     global,
}
