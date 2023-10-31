package github

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sync"

	"github.com/google/go-github/github"
	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"github.com/rumenvasilev/rvsecret/version"
)

// Client holds a github api client instance
type Client struct {
	apiClient *github.Client
	logger    *log.Logger
}

// NewClient creates a gitlab api client instance using a token
func NewClient(token, gheURL string, logger *log.Logger) (*Client, error) {
	err := validateAPIToken(token)
	if err != nil {
		return nil, fmt.Errorf("cannot create new Github client, %w", err)
	}
	// Get OAuth client
	oauth := getOauthClient(token)
	if gheURL != "" {
		return newEnterpriseClient(oauth, gheURL, logger)
	}
	return newStandardClient(oauth, logger), nil
}

func newStandardClient(oauth *http.Client, logger *log.Logger) *Client {
	c := github.NewClient(oauth)
	c.UserAgent = version.UserAgent

	return &Client{
		apiClient: c,
		logger:    logger,
	}
}

func newEnterpriseClient(oauth *http.Client, gheURL string, logger *log.Logger) (*Client, error) {
	baseURL := fmt.Sprintf("%s/api/v3", gheURL)
	uploadURL := fmt.Sprintf("%s/api/uploads", gheURL)

	var c *github.Client
	c, err := github.NewEnterpriseClient(baseURL, uploadURL, oauth)
	if err != nil {
		return nil, fmt.Errorf("unable to parse --github-enterprise-url: %q, %w", gheURL, err)
	}

	c.UserAgent = version.UserAgent

	return &Client{
		apiClient: c,
		logger:    logger,
	}, nil
}

// validateAPIToken will ensure we have a valid github api token
func validateAPIToken(token string) error {
	errmsg := "The token is invalid. Please use a valid Github token."
	// check to make sure the length is proper
	if len(token) != 40 {
		return fmt.Errorf(errmsg)
	}
	// match only letters and numbers and ensure you match 40
	exp1 := regexp.MustCompile(`[A-Za-z0-9\_]{40}`)
	if !exp1.MatchString(token) {
		return fmt.Errorf(errmsg)
	}
	return nil
}

// GetUserOrganization is used to enumerate the owner in a given org
func (c *Client) GetUserOrganization(ctx context.Context, name string) (*_coreapi.Owner, error) {
	res, err := c.getUser(ctx, name)
	// Couldn't find user by that name, try organization instead
	// TODO perhaps we should pass config argument which one are we searching for
	if err != nil {
		c.logger.Warn("Couldn't find user under that name %q. Will search for org instead.", name)
		c.logger.Debug("Wrapped error: %q", err.Error())
		return c.getOrg(ctx, name)
	}
	return res, nil
}

func (c *Client) getUser(ctx context.Context, name string) (*_coreapi.Owner, error) {
	user, _, err := c.apiClient.Users.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	return &_coreapi.Owner{
		Login:     user.Login,
		ID:        user.ID,
		Type:      user.Type,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		URL:       user.HTMLURL,
		Company:   user.Company,
		Blog:      user.Blog,
		Location:  user.Location,
		Email:     user.Email,
		Bio:       user.Bio,
		Kind:      util.StringToPointer(_coreapi.TargetTypeUser),
	}, nil
}

func (c *Client) getOrg(ctx context.Context, name string) (*_coreapi.Owner, error) {
	org, _, err := c.apiClient.Organizations.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	return &_coreapi.Owner{
		Login:     org.Login,
		ID:        org.ID,
		Type:      org.Type,
		Name:      org.Name,
		AvatarURL: org.AvatarURL,
		URL:       org.HTMLURL,
		Company:   org.Company,
		Blog:      org.Blog,
		Location:  org.Location,
		Email:     org.Email,
		Kind:      util.StringToPointer(_coreapi.TargetTypeOrganization),
		// Bio:       org.Bio,
	}, nil
}

// GetRepositoriesFromOwner is used gather all the repos associated with the org owner or other user.
// This is only used by the gitlab client. The github client use a github specific function.
func (c *Client) GetRepositoriesFromOwner(ctx context.Context, target _coreapi.Owner) ([]*_coreapi.Repository, error) {
	switch *target.Kind {
	case _coreapi.TargetTypeOrganization:
		c.logger.Debug("We're searching all org repositories...")
		return c.getOrgRepositories(ctx, *target.Login)
	case _coreapi.TargetTypeUser:
		c.logger.Debug("We're searching all user repositories...")
		return c.getUserRepositories(ctx, *target.Login)
	default:
		return nil, fmt.Errorf("unsupported target type %q", *target.Kind)
	}
}

func (c *Client) getOrgRepositories(ctx context.Context, login string) ([]*_coreapi.Repository, error) {
	var allRepos []*_coreapi.Repository
	opt := &github.RepositoryListByOrgOptions{
		Type: "sources",
	}
	for {
		repos, resp, err := c.apiClient.Repositories.ListByOrg(ctx, login, opt)
		if err != nil {
			return allRepos, err
		}

		allRepos = append(allRepos, githubToAPIRepos(repos)...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func (c *Client) getUserRepositories(ctx context.Context, login string) ([]*_coreapi.Repository, error) {
	var allRepos []*_coreapi.Repository
	opt := &github.RepositoryListOptions{
		Type: "sources",
	}
	for {
		repos, resp, err := c.apiClient.Repositories.List(ctx, login, opt)
		if err != nil {
			return allRepos, err
		}

		allRepos = append(allRepos, githubToAPIRepos(repos)...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func githubToAPIRepos(repos []*github.Repository) []*_coreapi.Repository {
	var allRepos []*_coreapi.Repository
	for _, repo := range repos {
		if !*repo.Fork {
			r := _coreapi.Repository{
				Owner:         util.PointerToString(repo.Owner.Login),
				ID:            util.PointerToInt64(repo.ID),
				Name:          util.PointerToString(repo.Name),
				FullName:      util.PointerToString(repo.FullName),
				CloneURL:      util.PointerToString(repo.CloneURL),
				URL:           util.PointerToString(repo.HTMLURL),
				DefaultBranch: util.PointerToString(repo.DefaultBranch),
				Description:   util.PointerToString(repo.Description),
				Homepage:      util.PointerToString(repo.Homepage),
			}
			allRepos = append(allRepos, &r)
		}
	}
	return allRepos
}

// GetOrganizationMembers will gather all the members of a given organization
func (c *Client) GetOrganizationMembers(ctx context.Context, target _coreapi.Owner) ([]*_coreapi.Owner, error) {
	var allMembers []*_coreapi.Owner
	opt := &github.ListMembersOptions{}

	var wg sync.WaitGroup
	var mut sync.Mutex

	for {
		members, resp, err := c.apiClient.Organizations.ListMembers(ctx, *target.Login, opt)
		if err != nil {
			return nil, fmt.Errorf("unable to get github organization members, %w", err)
		}

		wg.Add(1)
		go func(members []*github.User) {
			for _, member := range members {
				mut.Lock()
				allMembers = append(allMembers, &_coreapi.Owner{Login: member.Login, ID: member.ID, Type: member.Type})
				mut.Unlock()
			}
			wg.Done()
		}(members)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	wg.Wait()

	return allMembers, nil
}

func (c *Client) GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, error) {
	release, _, err := c.apiClient.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	return release, nil
}

func (c *Client) GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, error) {
	release, _, err := c.apiClient.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	if err != nil {
		return nil, err
	}

	return release, nil
}
