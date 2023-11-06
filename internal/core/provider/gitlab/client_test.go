package gitlab

import (
	"context"
	"testing"

	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
)

func Test_NewClient(t *testing.T) {
	var token = "abcdefghijklmnopqrstuvwxyz"
	t.Run("ok", func(t *testing.T) {
		client, err := NewClient(token[:20])
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("less chars, error", func(t *testing.T) {
		client, err := NewClient(token[:18])
		assert.Error(t, err)
		assert.EqualError(t, err, "Gitlab token is invalid")
		assert.Nil(t, client)
	})

	t.Run("more chars, error", func(t *testing.T) {
		client, err := NewClient(token)
		assert.Error(t, err)
		assert.EqualError(t, err, "Gitlab token is invalid")
		assert.Nil(t, client)
	})
}

func Test_GetUserOrganization(t *testing.T) {
	glClient, err := gitlab.NewClient("", gitlab.WithBaseURL(mockedURL))
	assert.NoError(t, err)
	c := &Client{apiClient: glClient}

	t.Run("Get User", func(t *testing.T) {
		owner, err := c.GetUserOrganization(context.TODO(), "barbara")
		assert.NoError(t, err)
		assert.NotNil(t, owner)
		assert.IsType(t, &_coreapi.Owner{}, owner)

		assert.Equal(t, "john", *owner.Name)
		assert.Equal(t, int64(5), *owner.ID)
		assert.Equal(t, "User", *owner.Type)
	})

	t.Run("Get User 2, error", func(t *testing.T) {
		owner, err := c.GetUserOrganization(context.TODO(), "monica")
		assert.Error(t, err)
		assert.Nil(t, owner)
		assert.EqualError(t, err, "no GitLab user or organization monica was found. If you are targeting a GitLab group, be sure to use an ID in place of a name")
	})

	t.Run("Get Organization", func(t *testing.T) {
		owner, err := c.GetUserOrganization(context.TODO(), "1")
		assert.NoError(t, err)
		assert.NotNil(t, owner)
		assert.IsType(t, &_coreapi.Owner{}, owner)

		assert.Equal(t, "foobar123thisorgwasmocked", *owner.Name)
		assert.Equal(t, int64(1), *owner.ID)
		assert.Equal(t, "Organization", *owner.Type)
	})
}

func Test_GetOrganizationMembers(t *testing.T) {
	glClient, err := gitlab.NewClient("", gitlab.WithBaseURL(mockedURL))
	assert.NoError(t, err)
	c := &Client{apiClient: glClient}

	owners, err := c.GetOrganizationMembers(context.TODO(), _coreapi.Owner{ID: util.Int64ToPointer(12)})
	assert.NoError(t, err)
	assert.NotNil(t, owners)
	assert.Len(t, owners, 2)
	assert.IsType(t, []*_coreapi.Owner{}, owners)

	assert.Equal(t, "jack", *owners[0].Login)
	assert.Equal(t, int64(7), *owners[0].ID)
	assert.Equal(t, "User", *owners[0].Type)

	assert.Equal(t, "patrick", *owners[1].Login)
	assert.Equal(t, int64(2), *owners[1].ID)
	assert.Equal(t, "User", *owners[1].Type)
}

func Test_GetOrganizationMembersE(t *testing.T) {
	glClient, err := gitlab.NewClient("", gitlab.WithBaseURL(mockedURLError))
	assert.NoError(t, err)
	c := &Client{apiClient: glClient}
	owners, err := c.GetOrganizationMembers(context.TODO(), _coreapi.Owner{ID: util.Int64ToPointer(12)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404 test error we've defined")
	assert.Nil(t, owners)

}

func Test_GetRepositoriesFromOwner(t *testing.T) {
	glClient, err := gitlab.NewClient("", gitlab.WithBaseURL(mockedURL))
	assert.NoError(t, err)
	c := &Client{apiClient: glClient}

	t.Run("getUserProjects", func(t *testing.T) {
		repos, err := c.GetRepositoriesFromOwner(
			context.TODO(),
			_coreapi.Owner{
				ID:   util.Int64ToPointer(5),
				Type: util.StringToPointer("User"),
			},
		)
		assert.NoError(t, err)
		assert.NotNil(t, repos)
		assert.Len(t, repos, 2)
		assert.IsType(t, []*_coreapi.Repository{}, repos)
		assert.Equal(t, repos[0].FullName, "john/mocked-proj-1")
		assert.Equal(t, repos[1].FullName, "john/mocked-proj-2")
	})

	t.Run("getGroupProjects", func(t *testing.T) {
		repos, err := c.GetRepositoriesFromOwner(
			context.TODO(),
			_coreapi.Owner{
				ID:   util.Int64ToPointer(17),
				Type: util.StringToPointer("Group"),
			},
		)
		assert.NoError(t, err)
		assert.NotNil(t, repos)
		assert.Len(t, repos, 2)
		assert.IsType(t, []*_coreapi.Repository{}, repos)
		assert.Equal(t, repos[0].FullName, "john/mocked-proj-1")
		assert.Equal(t, repos[1].FullName, "john/mocked-proj-2")
	})

	t.Run("incomplete input", func(t *testing.T) {
		repos, err := c.GetRepositoriesFromOwner(
			context.TODO(),
			_coreapi.Owner{},
		)
		assert.Error(t, err)
		assert.EqualError(t, err, "ID or Type fields are not present")
		assert.Nil(t, repos)
	})
}

func Test_GetRepositoriesFromOwnerE(t *testing.T) {
	glClient, err := gitlab.NewClient("", gitlab.WithBaseURL(mockedURLError))
	assert.NoError(t, err)
	c := &Client{apiClient: glClient}

	t.Run("getUserProjects", func(t *testing.T) {
		repos, err := c.GetRepositoriesFromOwner(
			context.TODO(),
			_coreapi.Owner{
				ID:   util.Int64ToPointer(5),
				Type: util.StringToPointer("User"),
			},
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404 getUserProjects failed")
		assert.Nil(t, repos)
	})

	t.Run("getGroupProjects", func(t *testing.T) {
		repos, err := c.GetRepositoriesFromOwner(
			context.TODO(),
			_coreapi.Owner{
				ID:   util.Int64ToPointer(17),
				Type: util.StringToPointer("Group"),
			},
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404 getGroupProjects failed")
		assert.Nil(t, repos)
	})
}

func Test_getUserE(t *testing.T) {
	glClient, err := gitlab.NewClient("", gitlab.WithBaseURL(mockedURLError))
	assert.NoError(t, err)
	c := &Client{apiClient: glClient}

	user, err := c.getUser("john")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404 getUser failed")
	assert.Nil(t, user)
}

func Test_getOrganization(t *testing.T) {
	glClient, err := gitlab.NewClient("", gitlab.WithBaseURL(mockedURLError))
	assert.NoError(t, err)
	c := &Client{apiClient: glClient}

	group, err := c.getOrganization("1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404 getOrganization failed")
	assert.Nil(t, group)
}
