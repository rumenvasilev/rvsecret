package api

import (
	"context"

	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
)

// IClient interface is used with the api clients to hold the repo and org specific info.
type IClient interface {
	GetUserOrganization(ctx context.Context, login string) (*_coreapi.Owner, error)
	GetRepositoriesFromOwner(ctx context.Context, target _coreapi.Owner) ([]*_coreapi.Repository, error)
	GetOrganizationMembers(ctx context.Context, target _coreapi.Owner) ([]*_coreapi.Owner, error)
}
