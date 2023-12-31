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
	Owner         string
	Name          string
	FullName      string
	CloneURL      string
	URL           string
	DefaultBranch string
	Description   string // WHY DO WE NEED THIS FIELD???
	Homepage      string // WHY DO WE NEED THIS FIELD???
	ID            int64
}
