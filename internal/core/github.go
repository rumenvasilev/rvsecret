package core

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"sync"

	"github.com/google/go-github/github"
	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/util"
)

// addUser will add a new user to the sess for further scanning and analyzing
func (s *Session) addUser(user *_coreapi.Owner) {
	s.State.Lock()
	defer s.State.Unlock()
	h := md5.New()
	_, _ = io.WriteString(h, *user.Login)                     // TODO handle error
	_, _ = io.WriteString(h, strconv.FormatInt(*user.ID, 10)) // TODO handle error
	userMD5 := fmt.Sprintf("%x", h.Sum(nil))

	for _, o := range s.GithubUsers {
		j := md5.New()
		_, _ = io.WriteString(j, *o.Login)                     // TODO handle error
		_, _ = io.WriteString(h, strconv.FormatInt(*o.ID, 10)) // TODO handle error
		sessMD5 := fmt.Sprintf("%x", h.Sum(nil))

		if userMD5 == sessMD5 {
			return
		}
	}
	s.GithubUsers = append(s.GithubUsers, user)
}

// GatherUsers will generate a list of users from github.com that can then be filtered down to a specific target range
func GatherUsers(sess *Session) error {
	log := sess.Out
	log.Important("Gathering users...")
	ctx := context.Background()
	for _, o := range sess.GithubUserLogins {
		owner, err := sess.Client.GetUserOrganization(ctx, o)
		if err != nil {
			// Should we not skip here?
			log.Error("Unable to collect user %s: %s", o, err)
		}

		// Add the user to the session and increment the user count
		sess.addUser(owner)
		sess.State.Stats.IncrementUsers()
		log.Debug("Added user %s", *owner.Login)
	}
	if len(sess.GithubUsers) == 0 {
		return fmt.Errorf("no Github users found")
	}
	return nil
}

// ValidateUserInput will check for special characters in the strings and make sure we
// have at least one usr/repo/org to scan
func (s *Session) ValidateUserInput() error {
	// Raw user inputs
	// if s.ScanType == api.GithubEnterprise {

	// If no targets are given, fail fast
	if s.Config.Github.UserDirtyRepos == nil && s.Config.Github.UserDirtyOrgs == nil && s.Config.Github.UserDirtyNames == nil {
		return errors.New("you must enter either a user, org or repo[s] to scan")
	}

	// validate the input does not contain any scary characters
	exp := regexp.MustCompile(`[A-Za-z0-9,-_]*$`)

	for _, v := range s.Config.Github.UserDirtyOrgs {
		if exp.MatchString(v) {
			s.GithubUserOrgs = append(s.GithubUserOrgs, v)
		}
	}

	for _, v := range s.Config.Github.UserDirtyRepos {
		if exp.MatchString(v) {
			s.GithubUserRepos = append(s.GithubUserRepos, v)
		}
	}

	for _, v := range s.Config.Github.UserDirtyNames {
		if exp.MatchString(v) {
			s.GithubUserLogins = append(s.GithubUserLogins, v)
		}
	}
	return nil
}

// GatherGithubRepositoriesFromOwner is used gather all the repos associated with a github user
func GatherGithubRepositoriesFromOwner(sess *Session) error {
	log := sess.Out
	var allRepos []*_coreapi.Repository
	ctx := context.Background()

	// The defaults should be fine for a tool like this but if you want to customize
	// settings like repo type (public, private, etc) or the amount of results returned
	// per page this is where you do it.
	// opt := &github.RepositoryListOptions{}

	owner := _coreapi.Owner{Kind: util.StringToPointer(_coreapi.TargetTypeUser)}
	var err error
	// TODO This should be threaded
	for _, ul := range sess.GithubUserLogins {
		// Reset the Page to start for every user
		// opt.Page = 1
		owner.Login = &ul
		allRepos, err = sess.Client.GetRepositoriesFromOwner(ctx, owner)
		if err != nil {
			log.Error("%v", err)
		}
	}

	found := len(allRepos)
	if found == 0 {
		return fmt.Errorf("no repositories have been found for any of the Github owners")
	}
	// Information also available in the summary report
	log.Debug("Found %d repositories", found)

	// If we re only looking for a subset of the repos in an org we do a comparison
	// of the repos gathered for the org and the list of repos that we care about.
	for _, repo := range allRepos {
		// Increment the total number of repos found, regardless if we are cloning them
		sess.State.Stats.IncrementRepositoriesTotal()
		if sess.GithubUserRepos != nil {
			for _, r := range sess.GithubUserRepos {
				log.Debug("current repo: %s, comparing to: %s", r, repo.Name)
				if r == repo.Name {
					log.Debug(" Retrieved repository %s from user %s", repo.FullName, repo.Owner)
					// Add the repo to the sess to be scanned
					sess.AddRepository(repo)
				}
			}
			continue
		}
		log.Debug(" Retrieved repository %s from user %s", repo.FullName, repo.Owner)

		// If we are not doing any filtering and simply grabbing all available repos we add the repos
		// to the session to be scanned
		sess.AddRepository(repo)
	}
	return nil
}

// GatherOrgs will use a client to generate a list of all orgs that the client can see. By default this will include
// orgs that contain both public and private repos
func GatherOrgs(sess *Session, log *log.Logger) error {
	log.Important("Gathering github organizations...")
	ctx := context.Background()

	var orgList []*github.Organization
	// var orgID int64
	// Options necessary for enumerating the orgs. These are not client options such as TLS or auth,
	// these are options such as orgs per page.
	// var opts github.OrganizationsListOptions
	// How many orgs per page
	// opts.PerPage = 40
	// This controls pagination, see below. In order for it to work this gets set to the last org that was found
	// and the next page picks up with the next one in line.
	// opts.Since = -1

	// Used to track the orgID's for the sake of pagination.
	// tmpOrgID := int64(0)
	// TODO SHOULD WE EVEN SEARCH FOR ALL THE AVAILABLE ORGS?
	// If the user did not specify specific orgs then we grab all the orgs the client can see
	// if sess.UserOrgs == nil {
	// 	for opts.Since < tmpOrgID {
	// 		if opts.Since == orgID {
	// 			break
	// 		}
	// 		orgs, _, err := sess.GithubClient.Organizations.ListAll(ctx, &opts)

	// 		if err != nil {
	// 			log.Error("Error gathering Github orgs: %s\n", err)
	// 		}

	// 		for _, org := range orgs {
	// 			orgList = append(orgList, org)
	// 			orgID = *org.ID
	// 		}

	// 		opts.Since = orgID
	// 		tmpOrgID = orgID + 1
	// 	}
	// } else {
	// This will handle orgs passed in via flags
	for _, o := range sess.GithubUserOrgs {
		owner, err := sess.Client.GetUserOrganization(ctx, o)
		if err != nil {
			log.Error("Error gathering the Github org %s: %s", o, err)
		}
		orgList = append(orgList, ownerToOrg(owner))
	}
	// }

	if len(orgList) == 0 {
		return fmt.Errorf("no Github orgs have been found")
	}

	// Add the orgs to the list for later enumeration of repos
	for _, org := range orgList {
		sess.addOrganization(org)
		sess.State.Stats.IncrementOrgs()
		log.Debug("Added org %s", *org.Login)
	}
	return nil
}

func ownerToOrg(owner *_coreapi.Owner) *github.Organization {
	if owner == nil {
		empty := "empty"
		return &github.Organization{
			Login: &empty,
			Name:  &empty,
		}
	}
	// Not all fields are populated, because Owner struct doesn't hold all of them
	return &github.Organization{
		Login:     owner.Login,
		ID:        owner.ID,
		AvatarURL: owner.AvatarURL,
		Name:      owner.Name,
		Company:   owner.Company,
		Blog:      owner.Blog,
		Location:  owner.Location,
		Email:     owner.Email,
		Type:      owner.Type,
		URL:       owner.URL,
	}
}

// addOrganization will add a new organization to the session for further scanning and analyzing
func (s *Session) addOrganization(organization *github.Organization) {
	s.State.Lock()
	defer s.State.Unlock()
	h := md5.New()
	_, _ = io.WriteString(h, *organization.Login) // TODO handle these errors instead of ignoring them explictly
	_, _ = io.WriteString(h, strconv.FormatInt(*organization.ID, 10))
	orgMD5 := fmt.Sprintf("%x", h.Sum(nil))

	for _, o := range s.Organizations {
		j := md5.New()
		_, _ = io.WriteString(j, *o.Login)
		_, _ = io.WriteString(h, strconv.FormatInt(*o.ID, 10))
		sessMD5 := fmt.Sprintf("%x", h.Sum(nil))

		if orgMD5 == sessMD5 {
			return
		}
	}
	s.Organizations = append(s.Organizations, organization)
}

// GatherGithubOrgRepositories will gather all the repositories for a given org.
func GatherGithubOrgRepositories(sess *Session, log *log.Logger) error {
	orgsCnt := len(sess.Organizations)
	// Create a channel for each org in the list
	var ch = make(chan *github.Organization, orgsCnt)
	var wg sync.WaitGroup

	// Calculate the number of threads based on the flag and the number of orgs
	// TODO: implement nice in the threading logic to guard against rate limiting and tripping the
	//  security protections
	threadNum := sess.Config.Global.Threads
	if orgsCnt <= 1 {
		threadNum = 1
	} else if orgsCnt <= sess.Config.Global.Threads {
		threadNum = orgsCnt - 1
	}
	wg.Add(threadNum)
	log.Debug("Threads for repository gathering: %d", threadNum)

	// Start workers
	for i := 0; i < threadNum; i++ {
		go ghWorker(sess, i, &wg, ch, log)
	}

	for _, org := range sess.Organizations {
		ch <- org
	}
	close(ch)
	wg.Wait()
	if len(sess.State.Repositories) == 0 {
		return fmt.Errorf("no repositories have been found for any of the provided Github organizations")
	}
	return nil
}

// GatherOrgsMembersRepositories will gather all orgs members repositories
func GatherOrgsMembersRepositories(sess *Session) {
	var allRepos []*_coreapi.Repository
	log := sess.Out

	sess.Out.Important("Gathering users from orgs...")
	ctx := context.Background()

	// optMember := &github.ListMembersOptions{}
	// optRepo := &github.RepositoryListOptions{}

	// TODO multi thread this
	for _, o := range sess.Organizations {
		members, err := sess.Client.GetOrganizationMembers(ctx, _coreapi.Owner{
			Login: o.Login,
		})
		// Log error and continue with the next org
		if err != nil {
			sess.Out.Warn("%v", err)
		}
		for _, v := range members {
			sess.addUser(v)
			sess.State.Stats.IncrementUsers()
			log.Debug("Added user %s", *v.Login)
			// find all repositories
			allRepos, err = sess.Client.GetRepositoriesFromOwner(ctx, *v)
			if err != nil {
				log.Error("%v", err)
			}
		}
	}
	// FIXME what happens if no repos are recovered

	// If we re only looking for a subset of the repos in an org we do a comparison
	// of the repos gathered for the org and the list of repos that we care about.
	for _, repo := range allRepos {
		// Increment the total number of repos found, regardless if we are cloning them
		sess.State.Stats.IncrementRepositoriesTotal()
		if sess.GithubUserRepos != nil {
			for _, r := range sess.GithubUserRepos {
				if r == repo.Name {
					sess.Out.Debug(" Retrieved repository %s from user %s", repo.FullName, repo.Owner)
					sess.AddRepository(repo)
				}
			}
			continue
		} else {
			sess.Out.Debug(" Retrieved repository %s from user %s", repo.FullName, repo.Owner)
			sess.AddRepository(repo)
		}
	}
}
