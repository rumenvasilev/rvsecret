package gitlab

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/version"
	"github.com/xanzy/go-gitlab"
)

// Client holds a gitlab api client instance
type Client struct {
	apiClient *gitlab.Client
}

// NewClient creates a gitlab api client instance using a token
func NewClient(token string) (*Client, error) {
	err := validateAPIToken(token)
	if err != nil {
		return nil, err
	}

	c, err := gitlab.NewClient(token)
	if err != nil {
		return nil, err
	}
	c.UserAgent = version.UserAgent

	client := &Client{
		apiClient: c,
	}
	return client, nil
}

// validateAPIToken will ensure we have a valid github api token
func validateAPIToken(t string) error {
	// check to make sure the length is proper
	if len(t) != 20 {
		return fmt.Errorf("Gitlab token is invalid")
	}
	return nil
}

// GetUserOrganization is used to enumerate the owner in a given org
func (c *Client) GetUserOrganization(ctx context.Context, login string) (*_coreapi.Owner, error) {
	emptyString := gitlab.String("")
	org, orgErr := c.getOrganization(login)
	if orgErr != nil {
		user, userErr := c.getUser(login)
		if userErr != nil {
			return nil, userErr
		}
		id := int64(user.ID)
		return &_coreapi.Owner{
			Login:     gitlab.String(user.Username),
			ID:        &id,
			Type:      gitlab.String(_coreapi.TargetTypeUser),
			Name:      gitlab.String(user.Name),
			AvatarURL: gitlab.String(user.AvatarURL),
			URL:       gitlab.String(user.WebsiteURL),
			Company:   gitlab.String(user.Organization),
			Blog:      emptyString,
			Location:  emptyString,
			Email:     gitlab.String(user.PublicEmail),
			Bio:       gitlab.String(user.Bio),
		}, nil
	}
	id := int64(org.ID)
	return &_coreapi.Owner{
		Login:     gitlab.String(org.Name),
		ID:        &id,
		Type:      gitlab.String(_coreapi.TargetTypeOrganization),
		Name:      gitlab.String(org.Name),
		AvatarURL: gitlab.String(org.AvatarURL),
		URL:       gitlab.String(org.WebURL),
		Company:   gitlab.String(org.FullName),
		Blog:      emptyString,
		Location:  emptyString,
		Email:     emptyString,
		Bio:       gitlab.String(org.Description),
	}, nil

}

// GetOrganizationMembers will gather all the members of a given organization
func (c *Client) GetOrganizationMembers(ctx context.Context, target _coreapi.Owner) ([]*_coreapi.Owner, error) {
	var allMembers []*_coreapi.Owner
	opt := &gitlab.ListGroupMembersOptions{}
	sID := strconv.FormatInt(*target.ID, 10) //safely downcast an int64 to an int
	for {
		members, resp, err := c.apiClient.Groups.ListAllGroupMembers(sID, opt)
		if err != nil {
			return nil, err
		}
		for _, member := range members {
			id := int64(member.ID)
			allMembers = append(allMembers,
				&_coreapi.Owner{
					Login: gitlab.String(member.Username),
					ID:    &id,
					Type:  gitlab.String(_coreapi.TargetTypeUser)})
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allMembers, nil
}

// GetRepositoriesFromOwner is used gather all the repos associated with the org owner or other user
func (c *Client) GetRepositoriesFromOwner(ctx context.Context, target _coreapi.Owner) ([]*_coreapi.Repository, error) {
	if target.ID == nil || target.Type == nil {
		return nil, errors.New("ID or Type fields are not present")
	}
	var allProjects []*_coreapi.Repository
	id := int(*target.ID)
	if *target.Type == _coreapi.TargetTypeUser {
		userProjects, err := c.getUserProjects(id)
		if err != nil {
			return nil, err
		}
		allProjects = append(allProjects, userProjects...)
	} else {
		groupProjects, err := c.getGroupProjects(target)
		if err != nil {
			return nil, err
		}
		allProjects = append(allProjects, groupProjects...)
	}
	return allProjects, nil
}

// getUser will get the necessary info from a given user
func (c *Client) getUser(login string) (*gitlab.User, error) {
	users, _, err := c.apiClient.Users.ListUsers(&gitlab.ListUsersOptions{Username: gitlab.String(login)})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("no GitLab %s or %s %s was found. If you are targeting a GitLab group, be sure to use an ID in place of a name",
			strings.ToLower(_coreapi.TargetTypeUser),
			strings.ToLower(_coreapi.TargetTypeOrganization),
			login)
	}
	return users[0], err
}

// getOrganization will get the necessary info from an org
func (c *Client) getOrganization(login string) (*gitlab.Group, error) {
	id, err := strconv.Atoi(login)
	if err != nil {
		return nil, err
	}
	org, _, err := c.apiClient.Groups.GetGroup(id, nil, nil)
	if err != nil {
		return nil, err
	}
	return org, err
}

// getUserProjects will gather the projects associated with a given user
func (c *Client) getUserProjects(id int) ([]*_coreapi.Repository, error) {
	var allUserProjects []*_coreapi.Repository
	listUserProjectsOps := &gitlab.ListProjectsOptions{}

	var wg sync.WaitGroup
	var mut sync.Mutex

	for {
		projects, response, err := c.apiClient.Projects.ListUserProjects(id, listUserProjectsOps)
		if err != nil {
			return nil, err
		}

		wg.Add(1)

		go func() {
			for _, project := range projects {
				//don't capture forks
				if project.ForkedFromProject == nil {
					id := int64(project.ID)
					p := _coreapi.Repository{
						Owner:         project.Owner.Username,
						ID:            id,
						Name:          project.Name,
						FullName:      project.NameWithNamespace,
						CloneURL:      project.HTTPURLToRepo,
						URL:           project.WebURL,
						DefaultBranch: project.DefaultBranch,
						Description:   project.Description,
						Homepage:      project.WebURL,
					}
					mut.Lock()
					allUserProjects = append(allUserProjects, &p)
					mut.Unlock()
				}
			}
			wg.Done()
		}()

		if response.NextPage == 0 {
			break
		}
		listUserProjectsOps.Page = response.NextPage
	}
	wg.Wait()

	return allUserProjects, nil
}

// getGroupProjects will gather the projects associated with a given group
func (c *Client) getGroupProjects(target _coreapi.Owner) ([]*_coreapi.Repository, error) {
	var allGroupProjects []*_coreapi.Repository
	listGroupProjectsOps := &gitlab.ListGroupProjectsOptions{}
	id := strconv.FormatInt(*target.ID, 10)

	var wg sync.WaitGroup
	var mut sync.Mutex

	for {
		projects, response, err := c.apiClient.Groups.ListGroupProjects(id, listGroupProjectsOps)
		if err != nil {
			return nil, err
		}

		wg.Add(1)

		go func() {
			for _, project := range projects {
				//don't capture forks
				if project.ForkedFromProject == nil {
					id := int64(project.ID)
					p := _coreapi.Repository{
						Owner:         project.Namespace.FullPath,
						ID:            id,
						Name:          project.Name,
						FullName:      project.NameWithNamespace,
						CloneURL:      project.HTTPURLToRepo,
						URL:           project.WebURL,
						DefaultBranch: project.DefaultBranch,
						Description:   project.Description,
						Homepage:      project.WebURL,
					}
					mut.Lock()
					allGroupProjects = append(allGroupProjects, &p)
					mut.Unlock()
				}
			}
			wg.Done()
		}()

		if response.NextPage == 0 {
			break
		}
		listGroupProjectsOps.Page = response.NextPage
	}
	wg.Wait()

	return allGroupProjects, nil
}
