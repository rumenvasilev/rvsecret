package core

import (
	"context"
	"errors"
	"net"
	"regexp"
	"strings"
	"time"

	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/session"
	"github.com/rumenvasilev/rvsecret/internal/util"
)

// GatherUsers will generate a list of users from github.com that can then be filtered down to a specific target range
func GatherUserOrOrg(s *session.Session, targetList []string) {
	log := log.Log
	log.Important("Gathering targets...")
	ctx := context.Background()
	for _, o := range targetList {
		owner, err := s.Client.GetUserOrganization(ctx, o)
		if err != nil {
			var dnsErr *net.DNSError
			if errors.As(err, &dnsErr) {
				if dnsErr.IsNotFound {
					log.Error("Encountered DNS resolution error: %s", err)
					return
				} else if dnsErr.Timeout() {
					log.Error("DNS Timeout, will pause for 5 seconds")
					// TODO replace time.Sleep with custom func, so it doesn't block and respond to context cancellation
					time.Sleep(time.Duration(5 * time.Second))
					continue
				}
			}
			log.Error("Unable to collect user %s: %s", o, err)
			continue
		}

		// Add the user/org to the session and increment the target count
		s.State.AddTarget(owner)
		log.Debug("Added %s %q", strings.ToLower(*owner.Kind), *owner.Login)
	}
}

// ValidateUserInput will check for special characters in the strings and make sure we
// have at least one usr/repo/org to scan
func ValidateUserInput(s *session.Session) error {
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

// GatherOrgsMembers will gather all orgs members
// and position them in Targets
func GatherOrgsMembers(sess *session.Session) {
	log := log.Log
	log.Important("Gathering users from orgs...")
	ctx := context.Background()

	for _, o := range sess.Organizations {
		members, err := sess.Client.GetOrganizationMembers(ctx, _coreapi.Owner{
			Login: o.Login,
			Kind:  util.StringToPointer(_coreapi.TargetTypeOrganization),
		})
		// Log the error and continue with the next org
		if err != nil {
			log.Warn("%v", err)
			continue
		}
		sess.State.AddTargets(members)
	}
}
