package github

import (
	"net/http"

	"github.com/google/go-github/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

var mockedHTTPClient = mock.NewMockedHTTPClient(
	mock.WithRequestMatch(
		mock.GetUsersByUsername,
		github.User{
			Name: github.String("foobar"),
		},
	),
	mock.WithRequestMatch(
		mock.GetOrgsByOrg,
		github.Organization{
			Name: github.String("foobar123thisorgwasmocked"),
		},
	),
	mock.WithRequestMatch(
		mock.GetOrgsReposByOrg,
		[]*github.Repository{
			&github.Repository{
				ID: github.Int64(5),
				Owner: &github.User{
					Login: github.String("petko"),
				},
				Name:          github.String("petko"),
				FullName:      github.String("whatever/petko"),
				CloneURL:      github.String("github.fake/clone/whatever/petko"),
				HTMLURL:       github.String("github.fake/clone/whatever/petko"),
				DefaultBranch: github.String("maine"),
				Description:   github.String("no description now"),
				Homepage:      github.String("www.homepage.fake"),
				Fork:          github.Bool(false)},
			&github.Repository{
				ID: github.Int64(5),
				Owner: &github.User{
					Login: github.String("schmetko"),
				},
				Name:          github.String("schmetko"),
				FullName:      github.String("whatever/schmetko"),
				CloneURL:      github.String("github.fake/clone/whatever/schmetko"),
				HTMLURL:       github.String("github.fake/clone/whatever/schmetko"),
				DefaultBranch: github.String("maine"),
				Description:   github.String("no description now"),
				Homepage:      github.String("www.homepage.fake"),
				Fork:          github.Bool(false)},
		},
	),
	mock.WithRequestMatch(
		mock.GetUsersReposByUsername,
		[]*github.Repository{
			&github.Repository{
				ID: github.Int64(5),
				Owner: &github.User{
					Login: github.String("petko"),
				},
				Name:          github.String("petko"),
				FullName:      github.String("whatever/petko"),
				CloneURL:      github.String("github.fake/clone/whatever/petko"),
				HTMLURL:       github.String("github.fake/clone/whatever/petko"),
				DefaultBranch: github.String("maine"),
				Description:   github.String("no description now"),
				Homepage:      github.String("www.homepage.fake"),
				Fork:          github.Bool(false)},
			&github.Repository{
				ID: github.Int64(5),
				Owner: &github.User{
					Login: github.String("schmetko"),
				},
				Name:          github.String("schmetko"),
				FullName:      github.String("whatever/schmetko"),
				CloneURL:      github.String("github.fake/clone/whatever/schmetko"),
				HTMLURL:       github.String("github.fake/clone/whatever/schmetko"),
				DefaultBranch: github.String("maine"),
				Description:   github.String("no description now"),
				Homepage:      github.String("www.homepage.fake"),
				Fork:          github.Bool(false)},
		},
	),
	mock.WithRequestMatch(
		mock.GetOrgsMembersByOrg,
		[]*github.User{
			&github.User{
				Login: github.String("faking-it"),
				ID:    github.Int64(7),
				Type:  github.String("unknown"),
			},
		},
	),
	mock.WithRequestMatch(
		mock.GetReposReleasesLatestByOwnerByRepo,
		&github.RepositoryRelease{
			Name: github.String("my-fake-release"),
		},
	),
	mock.WithRequestMatch(
		mock.GetReposReleasesTagsByOwnerByRepoByTag,
		&github.RepositoryRelease{
			Name: github.String("my-fake-release"),
		},
	),
	mock.WithRequestMatchHandler(
		mock.GetOrgsProjectsByOrg,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write(mock.MustMarshal([]github.Project{
				{
					Name: github.String("mocked-proj-1"),
				},
				{
					Name: github.String("mocked-proj-2"),
				},
			}))
		}),
	),
)
