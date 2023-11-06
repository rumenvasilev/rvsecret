package github

import (
	"net/http"

	"github.com/google/go-github/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

var mockedURL = mock.NewMockedHTTPClient(
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
	mock.WithRequestMatchPages(
		mock.GetOrgsReposByOrg,
		[]*github.Repository{
			{
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
				Fork:          github.Bool(false),
			},
		},
		[]*github.Repository{
			{
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
				Fork:          github.Bool(false),
			},
		},
	),
	mock.WithRequestMatchPages(
		mock.GetUsersReposByUsername,
		[]*github.Repository{
			{
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
				Fork:          github.Bool(false),
			},
		},
		[]*github.Repository{
			{
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
				Fork:          github.Bool(false),
			},
		},
	),
	mock.WithRequestMatchPages(
		mock.GetOrgsMembersByOrg,
		[]*github.User{
			{
				Login: github.String("faking-it"),
				ID:    github.Int64(7),
				Type:  github.String("unknown"),
			},
		},
		[]*github.User{
			{
				Login: github.String("faking-it-2"),
				ID:    github.Int64(5),
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

var mockedURLErrorUserOrg = mock.NewMockedHTTPClient( // Test_GetUserOrganization
	mock.WithRequestMatchHandler(
		mock.GetUsersByUsername,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mock.WriteError(
				w,
				http.StatusNotFound,
				"getUser failed",
			)
		}),
	),

	mock.WithRequestMatch(
		mock.GetOrgsByOrg,
		github.Organization{
			Name: github.String("foobar123thisorgwasmocked"),
		},
	),
)

var mockedURLError = mock.NewMockedHTTPClient(
	mock.WithRequestMatchHandler(
		mock.GetOrgsByOrg,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mock.WriteError(
				w,
				http.StatusNotFound,
				"getOrg failed",
			)
		}),
	),
	mock.WithRequestMatchHandler(
		mock.GetOrgsReposByOrg,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mock.WriteError(
				w,
				http.StatusNotFound,
				"getOrgRepositories failed",
			)
		}),
	),
	mock.WithRequestMatchHandler(
		mock.GetUsersReposByUsername,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mock.WriteError(
				w,
				http.StatusNotFound,
				"getUserRepositories failed",
			)
		}),
	),
	mock.WithRequestMatchHandler(
		mock.GetReposReleasesLatestByOwnerByRepo,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mock.WriteError(
				w,
				http.StatusNotFound,
				"GetLatestRelease failed",
			)
		}),
	),
	mock.WithRequestMatchHandler(
		mock.GetReposReleasesTagsByOwnerByRepoByTag,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mock.WriteError(
				w,
				http.StatusNotFound,
				"GetReleaseByTag failed",
			)
		}),
	),
)
