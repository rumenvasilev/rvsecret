package api

type ScanType string

const (
	Github           ScanType = "github"
	GithubEnterprise ScanType = "github-enterprise"
	Gitlab           ScanType = "gitlab"
	LocalGit         ScanType = "localGit"
	LocalPath        ScanType = "localpath"
	UpdateSignatures ScanType = "update-signatures"
)
