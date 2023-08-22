package api

// Set easier names to refer to
const (
	TargetTypeUser         = "User"
	TargetTypeOrganization = "Organization"
)

// Owner holds the info that we want for a repo owner
type Owner struct {
	ID        *int64
	Kind      *string // holds information if this is an org or a user
	Login     *string
	Type      *string
	Name      *string
	AvatarURL *string
	URL       *string
	Company   *string
	Blog      *string
	Location  *string
	Email     *string
	Bio       *string
}

// Repository holds the info we want for a repo itself
type Repository struct {
	ID            int64
	Owner         string
	Name          string
	FullName      string
	CloneURL      string
	URL           string
	DefaultBranch string
	// WHY DO WE NEED THESE FIELDS AT ALL???
	Description string
	Homepage    string
}

type Status string

// These are various environment variables and tool statuses used in auth and displaying messages
const (
	StatusInitializing Status = "initializing"
	StatusGathering    Status = "gathering"
	StatusAnalyzing    Status = "analyzing"
	StatusFinished     Status = "finished"
)
