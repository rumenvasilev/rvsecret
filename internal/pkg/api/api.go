package api

type ScanType string

const (
	Github           ScanType = "github"
	GithubEnterprise ScanType = "github-enterprise"
	Gitlab           ScanType = "gitlab"
	LocalGit         ScanType = "localGit"
	LocalPath        ScanType = "localpath"
	Unknown          ScanType = "unknown" // for testing
	UpdateSignatures ScanType = "update-signatures"
)
