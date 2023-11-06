package gitlab

import (
	"net/http"

	"github.com/rumenvasilev/go-gitlab-mock/mock"
	"github.com/xanzy/go-gitlab"
)

var mockedURL = mock.NewMockedHTTPServer(
	mock.WithRequestMatch(
		mock.GetApiV4Users,
		[]*gitlab.User{
			{
				ID:   5,
				Name: "john",
			},
		},
		[]*gitlab.User{},
	),
	mock.WithRequestMatch(
		mock.GetApiV4GroupsById,
		&gitlab.Group{
			ID:   1,
			Name: "foobar123thisorgwasmocked",
		},
	),
	mock.WithRequestMatchPages(
		mock.GetApiV4UsersProjectsByUserId,
		[]*gitlab.Project{
			{
				Name:              "mocked-proj-1",
				NameWithNamespace: "john/mocked-proj-1",
				HTTPURLToRepo:     "http://bla.repo",
				WebURL:            "http://bla.repo",
				DefaultBranch:     "main",
				Description:       "sample description for the test",
				Owner:             &gitlab.User{Username: "john"},
			},
		},
		[]*gitlab.Project{
			{
				Name:              "mocked-proj-2",
				NameWithNamespace: "john/mocked-proj-2",
				HTTPURLToRepo:     "http://bla.repo2",
				WebURL:            "http://bla.repo2",
				DefaultBranch:     "main",
				Description:       "sample description for the test",
				Owner:             &gitlab.User{Username: "ben"},
			},
		},
	),
	mock.WithRequestMatchPages(
		mock.GetApiV4GroupsProjectsById,
		[]*gitlab.Project{
			{
				Name:              "mocked-proj-1",
				NameWithNamespace: "john/mocked-proj-1",
				Namespace:         &gitlab.ProjectNamespace{FullPath: "john/mocked-proj-1"},
				HTTPURLToRepo:     "http://bla.repo",
				WebURL:            "http://bla.repo",
				DefaultBranch:     "main",
				Description:       "sample description for the test",
				Owner:             &gitlab.User{Username: "john"},
			},
		},
		[]*gitlab.Project{
			{
				Name:              "mocked-proj-2",
				NameWithNamespace: "john/mocked-proj-2",
				Namespace:         &gitlab.ProjectNamespace{FullPath: "john/mocked-proj-2"},
				HTTPURLToRepo:     "http://bla.repo2",
				WebURL:            "http://bla.repo2",
				DefaultBranch:     "main",
				Description:       "sample description for the test",
				Owner:             &gitlab.User{Username: "ben"},
			},
		},
	),
	mock.WithRequestMatchPages(
		mock.GetApiV4GroupsMembersAllById,
		[]*gitlab.GroupMember{
			{ID: 7, Username: "jack"},
		},
		[]*gitlab.GroupMember{
			{ID: 2, Username: "patrick"},
		},
	),
)

var mockedURLError = mock.NewMockedHTTPServer(
	mock.WithRequestMatchHandler(
		mock.GetApiV4GroupsMembersAllById,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mock.WriteError(
				w,
				http.StatusNotFound,
				"test error we've defined",
			)
		}),
	),
	mock.WithRequestMatchHandler(
		mock.GetApiV4UsersProjectsByUserId,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mock.WriteError(
				w,
				http.StatusNotFound,
				"getUserProjects failed",
			)
		}),
	),
	mock.WithRequestMatchHandler(
		mock.GetApiV4GroupsProjectsById,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mock.WriteError(
				w,
				http.StatusNotFound,
				"getGroupProjects failed",
			)
		}),
	),
	mock.WithRequestMatchHandler(
		mock.GetApiV4Users,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mock.WriteError(
				w,
				http.StatusNotFound,
				"getUser failed",
			)
		}),
	),
	mock.WithRequestMatchHandler(
		mock.GetApiV4GroupsById,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mock.WriteError(
				w,
				http.StatusNotFound,
				"getOrganization failed",
			)
		}),
	),
)
