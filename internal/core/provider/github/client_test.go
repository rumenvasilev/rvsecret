package github

import (
	"context"
	"testing"

	"github.com/google/go-github/github"
	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	type args struct {
		token  string
		gheURL string
		logger *log.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		{"no token", args{"", "", &log.Logger{}}, &Client{}, true},
		{"invalid token", args{"invalid-token", "", &log.Logger{}}, &Client{}, true},
		{"valid token length, unsupported characters", args{"invalid-token-must-be-40-so-we$add_strin", "", &log.Logger{}}, &Client{}, true},
		{"valid token length, github client", args{"invalid_token_must_be_40_so_we_add_strin", "", &log.Logger{}}, &Client{}, false},
		{"valid token length, gh enterprise client", args{"invalid_token_must_be_40_so_we_add_strin", "fake-url", &log.Logger{}}, &Client{}, false},
		{"invalid enterprise url", args{"invalid_token_must_be_40_so_we_add_strin", "https\\://f@ke-url", &log.Logger{}}, &Client{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.token, tt.args.gheURL, tt.args.logger)
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

func Test_getUser(t *testing.T) {
	c := newStandardClient(mockedHTTPClient, &log.Logger{})
	want := _coreapi.Owner{
		Name: util.StringToPointer("foobar"),
	}

	owner, err := c.getUser(context.TODO(), "dummy")
	assert.NoError(t, err)
	assert.IsType(t, _coreapi.Owner{}, *owner)
	assert.Equal(t, *want.Name, *owner.Name)
}

func Test_getOrg(t *testing.T) {
	c := newStandardClient(mockedHTTPClient, &log.Logger{})
	want := _coreapi.Owner{
		Name: util.StringToPointer("foobar123thisorgwasmocked"),
	}

	owner, err := c.getOrg(context.TODO(), "dummy")
	assert.NoError(t, err)
	assert.IsType(t, _coreapi.Owner{}, *owner)
	assert.Equal(t, *want.Name, *owner.Name)
}

func Test_GetRepositoriesFromOwner(t *testing.T) {
	tests := []struct {
		name string
		kind string
		want []*_coreapi.Repository
	}{
		{"org-type", _coreapi.TargetTypeOrganization, []*_coreapi.Repository{
			{Name: "petko"},
			{Name: "schmetko"},
		}},
		{"user-type", _coreapi.TargetTypeUser, []*_coreapi.Repository{
			{Name: "petko"},
			{Name: "schmetko"},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newStandardClient(mockedHTTPClient, &log.Logger{})
			q := _coreapi.Owner{
				Kind:  util.StringToPointer(tt.kind),
				Login: util.StringToPointer("random-input"),
			}

			repos, err := c.GetRepositoriesFromOwner(context.TODO(), q)
			assert.NoError(t, err)
			assert.IsType(t, []*_coreapi.Repository{}, repos)
			assert.Len(t, repos, len(tt.want))
			assert.Equal(t, tt.want[0].Name, repos[0].Name)
			assert.Equal(t, tt.want[1].Name, repos[1].Name)
		})
	}
}

func Test_getOrgRepositories(t *testing.T) {
	t.Skip() // covered by GetRepositoriesFromOwner
	c := newStandardClient(mockedHTTPClient, &log.Logger{})
	want := []*_coreapi.Repository{
		{Name: "petko"},
		{Name: "schmetko"},
	}

	repos, err := c.getOrgRepositories(context.TODO(), "dummy")
	assert.NoError(t, err)
	assert.IsType(t, []*_coreapi.Repository{}, repos)
	assert.Len(t, repos, len(want))
	assert.Equal(t, want[0].Name, repos[0].Name)
	assert.Equal(t, want[1].Name, repos[1].Name)
}

func Test_getUserRepositories(t *testing.T) {
	t.Skip() // covered by GetRepositoriesFromOwner
	c := newStandardClient(mockedHTTPClient, &log.Logger{})
	want := []*_coreapi.Repository{
		{Name: "petko"},
		{Name: "schmetko"},
	}

	repos, err := c.getUserRepositories(context.TODO(), "dummy")
	assert.NoError(t, err)
	assert.IsType(t, []*_coreapi.Repository{}, repos)
	assert.Len(t, repos, 2)
	assert.Equal(t, want[0].Name, repos[0].Name)
	assert.Equal(t, want[1].Name, repos[1].Name)
}

func Test_GetOrganizationMembers(t *testing.T) {
	c := newStandardClient(mockedHTTPClient, &log.Logger{})
	want := []*_coreapi.Owner{
		&_coreapi.Owner{Login: util.StringToPointer("faking-it")},
	}

	q := _coreapi.Owner{
		Login: util.StringToPointer("foobar123thisorgwasmocked"),
	}

	owners, err := c.GetOrganizationMembers(context.TODO(), q)
	assert.NoError(t, err)
	assert.IsType(t, []*_coreapi.Owner{}, owners)
	assert.Len(t, owners, 1)
	assert.Equal(t, want[0].Login, owners[0].Login)
}

func Test_GetLatestRelease(t *testing.T) {
	c := newStandardClient(mockedHTTPClient, &log.Logger{})
	want := github.RepositoryRelease{Name: util.StringToPointer("my-fake-release")}

	got, err := c.GetLatestRelease(context.TODO(), "i-am", "fepo")
	assert.NoError(t, err)
	assert.IsType(t, github.RepositoryRelease{}, *got)
	assert.Equal(t, want, *got)
}

func Test_GetReleaseByTag(t *testing.T) {
	c := newStandardClient(mockedHTTPClient, &log.Logger{})
	want := github.RepositoryRelease{Name: util.StringToPointer("my-fake-release")}

	got, err := c.GetReleaseByTag(context.TODO(), "i-am", "fepo", "rnadom")
	assert.NoError(t, err)
	assert.IsType(t, github.RepositoryRelease{}, *got)
	assert.Equal(t, want, *got)
}
