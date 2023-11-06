package github

import (
	"context"
	"testing"

	"github.com/google/go-github/github"
	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	type args struct {
		token  string
		gheURL string
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		{"no token", args{"", ""}, &Client{}, true},
		{"invalid token", args{"invalid-token", ""}, &Client{}, true},
		{"valid token length, unsupported characters", args{"invalid-token-must-be-40-so-we$add_strin", ""}, &Client{}, true},
		{"valid token length, github client", args{"invalid_token_must_be_40_so_we_add_strin", ""}, &Client{}, false},
		{"valid token length, gh enterprise client", args{"invalid_token_must_be_40_so_we_add_strin", "fake-url"}, &Client{}, false},
		{"invalid enterprise url", args{"invalid_token_must_be_40_so_we_add_strin", "https\\://f@ke-url"}, &Client{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.token, tt.args.gheURL)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.NotNil(t, got.apiClient)
				assert.IsType(t, *tt.want, *got)
			}
		})
	}
}

func Test_githubToAPIRepos(t *testing.T) {
	type args struct {
		repos []*github.Repository
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"two repos", args{makeRepository(2)}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := githubToAPIRepos(tt.args.repos)
			assert.Len(t, got, tt.want)
		})
	}
}

func makeRepository(cnt int) []*github.Repository {
	r := github.Repository{
		ID:            util.Int64ToPointer(5),
		Owner:         &github.User{Login: util.StringToPointer("petko")},
		Name:          util.StringToPointer("repo.Name"),
		FullName:      util.StringToPointer("repo.FullName"),
		CloneURL:      util.StringToPointer("repo.CloneURL"),
		URL:           util.StringToPointer("repo.HTMLURL"),
		DefaultBranch: util.StringToPointer("repo.DefaultBranch"),
		Description:   util.StringToPointer("repo.Description"),
		Homepage:      util.StringToPointer("repo.Homepage"),
		Fork:          github.Bool(false),
	}

	var out []*github.Repository
	for i := 0; i < cnt; i++ {
		out = append(out, &r)
	}
	return out
}

func Test_GetUserOrganization(t *testing.T) {
	t.Run("getUser", func(t *testing.T) {
		want := _coreapi.Owner{
			Name: util.StringToPointer("foobar"),
		}
		c := newStandardClient(mockedURL)
		owner, err := c.GetUserOrganization(context.TODO(), "dummy")
		assert.NoError(t, err)
		assert.IsType(t, _coreapi.Owner{}, *owner)
		assert.Equal(t, *want.Name, *owner.Name)
	})

	t.Run("getOrg", func(t *testing.T) {
		want := _coreapi.Owner{
			Name: util.StringToPointer("foobar123thisorgwasmocked"),
		}
		c := newStandardClient(mockedURLErrorUserOrg)
		owner, err := c.GetUserOrganization(context.TODO(), "dummy")
		assert.NoError(t, err)
		assert.IsType(t, _coreapi.Owner{}, *owner)
		assert.Equal(t, *want.Name, *owner.Name)
	})
}

func Test_getOrgE(t *testing.T) {
	c := newStandardClient(mockedURLError)

	owner, err := c.getOrg(context.TODO(), "dummy")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404 getOrg failed []")
	assert.Nil(t, owner)
}

func Test_GetRepositoriesFromOwner(t *testing.T) {
	tests := []struct {
		name    string
		kind    string
		want    []*_coreapi.Repository
		wantErr string
	}{
		{"org-type", _coreapi.TargetTypeOrganization, []*_coreapi.Repository{
			{Name: "petko"},
			{Name: "schmetko"},
		}, ""},
		{"user-type", _coreapi.TargetTypeUser, []*_coreapi.Repository{
			{Name: "petko"},
			{Name: "schmetko"},
		}, ""},
		{"error", "unknown", nil, "unsupported target type \"unknown\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newStandardClient(mockedURL)
			q := _coreapi.Owner{
				Kind:  util.StringToPointer(tt.kind),
				Login: util.StringToPointer("random-input"),
			}

			repos, err := c.GetRepositoriesFromOwner(context.TODO(), q)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr)
				assert.Nil(t, repos)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, []*_coreapi.Repository{}, repos)
				assert.Len(t, repos, len(tt.want))
				assert.Equal(t, tt.want[0].Name, repos[0].Name)
				assert.Equal(t, tt.want[1].Name, repos[1].Name)
			}
		})
	}
}

func Test_getOrgRepositoriesE(t *testing.T) {
	c := newStandardClient(mockedURLError)

	repos, err := c.getOrgRepositories(context.TODO(), "dummy")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404 getOrgRepositories failed []")
	assert.Nil(t, repos)
}

func Test_getUserRepositoriesE(t *testing.T) {
	c := newStandardClient(mockedURLError)

	repos, err := c.getUserRepositories(context.TODO(), "dummy")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404 getUserRepositories failed []")
	assert.Nil(t, repos)
}

func Test_GetOrganizationMembers(t *testing.T) {
	c := newStandardClient(mockedURL)
	want := []*_coreapi.Owner{
		{Login: util.StringToPointer("faking-it")},
	}

	q := _coreapi.Owner{
		Login: util.StringToPointer("foobar123thisorgwasmocked"),
	}

	owners, err := c.GetOrganizationMembers(context.TODO(), q)
	assert.NoError(t, err)
	assert.IsType(t, []*_coreapi.Owner{}, owners)
	assert.Len(t, owners, 2)
	assert.Equal(t, want[0].Login, owners[0].Login)
}

func Test_GetOrganizationMembersE(t *testing.T) {
	c := newStandardClient(mockedURLError)
	q := _coreapi.Owner{
		Login: util.StringToPointer("foobar123thisorgwasmocked"),
	}

	owners, err := c.GetOrganizationMembers(context.TODO(), q)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404 mock response not found for /orgs/foobar123thisorgwasmocked/members []")
	assert.Nil(t, owners)
}

func Test_GetLatestRelease(t *testing.T) {
	c := newStandardClient(mockedURL)
	want := github.RepositoryRelease{Name: util.StringToPointer("my-fake-release")}

	got, err := c.GetLatestRelease(context.TODO(), "i-am", "fepo")
	assert.NoError(t, err)
	assert.IsType(t, github.RepositoryRelease{}, *got)
	assert.Equal(t, want, *got)
}

func Test_GetLatestReleaseE(t *testing.T) {
	c := newStandardClient(mockedURLError)
	got, err := c.GetLatestRelease(context.TODO(), "i-am", "fepo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404 GetLatestRelease failed []")
	assert.Nil(t, got)
}

func Test_GetReleaseByTag(t *testing.T) {
	c := newStandardClient(mockedURL)
	want := github.RepositoryRelease{Name: util.StringToPointer("my-fake-release")}

	got, err := c.GetReleaseByTag(context.TODO(), "i-am", "fepo", "rnadom")
	assert.NoError(t, err)
	assert.IsType(t, github.RepositoryRelease{}, *got)
	assert.Equal(t, want, *got)
}

func Test_GetReleaseByTagE(t *testing.T) {
	c := newStandardClient(mockedURLError)
	got, err := c.GetReleaseByTag(context.TODO(), "i-am", "fepo", "rnadom")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404 GetReleaseByTag failed []")
	assert.Nil(t, got)
}
