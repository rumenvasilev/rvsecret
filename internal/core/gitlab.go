package core

import (
	"context"
	"sync"

	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/util"
)

// GatherTargets will enumerate git targets adding them to a running target list. This will set the targets based
// on the scan type set within the cmd package.
func GatherTargets(sess *Session) {
	sess.State.Stats.UpdateStatus(_coreapi.StatusGathering)
	sess.Out.Important("Gathering targets...")
	ctx := context.Background()

	// var targets []string

	// Based on the type of scan, set in the cmd package, we set a generic
	// variable to the specific targets
	//switch sess.ScanType {
	//case "github":
	//	targets = sess.GithubTargets
	//case "gitlab":
	targets := sess.Config.GitlabTargets
	//}

	//var target *Owner

	// For each target that the user provided, we use the client set in the session
	// initialization to enumerate the target. There are flag that be used here to
	// decide if forks are followed the scope of a target can be increased a lot. This
	// could be useful as some developers may keep secrets in their forks, yet purge
	// them before creating a pull request. Developers may also keep a specific environment
	// file within their repo that is not set to be ignored so they can more easily develop
	// on multiple boxes or collaborate with multiple people.
	for _, loginOption := range targets {

		//if sess.ScanType == "github" || sess.ScanType == "github-enterprise" {
		//	target, err := sess.GithubClient.GetUserOrganization(loginOption)
		//	if err != nil || target == nil {
		//		sess.Out.Error(" Error retrieving information on %s: %s\n", loginOption, err)
		//		continue
		//	}
		//} else {
		target, err := sess.Client.GetUserOrganization(ctx, loginOption)
		if err != nil || target == nil {
			sess.Out.Error(" Error retrieving information on %s: %s", loginOption, err)
			continue
		}

		sess.Out.Debug("%s (ID: %d) type: %s", *target.Login, *target.ID, *target.Type)
		sess.AddTarget(target)
		// If forking is false AND the target type is an Organization as set above in GetUserOrganization
		if sess.Config.Global.ExpandOrgs && *target.Type == _coreapi.TargetTypeOrganization {
			sess.Out.Debug("Gathering members of %s (ID: %d)...", *target.Login, *target.ID)
			members, err := sess.Client.GetOrganizationMembers(ctx, *target)
			if err != nil {
				sess.Out.Error(" Error retrieving members of %s: %s", *target.Login, err)
				continue
			}
			// Add organization members gathered above to the target list
			// TODO Do we want to spider this out at some point to enumerate all members of an org?
			for _, member := range members {
				sess.Out.Debug("Adding organization member %s (ID: %d) to targets", *member.Login, *member.ID)
				sess.AddTarget(member)
			}
		}
	}
}

// GatherGitlabRepositories will gather all repositories associated with a given target during a scan session.
// This is done using threads, whose count is set via commandline flag. Care much be taken to avoid rate
// limiting associated with suspected DOS attacks.
func GatherGitlabRepositories(sess *Session) {
	log := sess.Out
	ctx := context.Background()
	var ch = make(chan *_coreapi.Owner, len(sess.State.Targets))
	log.Debug("Number of targets: %d", len(sess.State.Targets))
	var wg sync.WaitGroup
	var threadNum int
	if len(sess.State.Targets) == 1 {
		threadNum = 1
	} else if len(sess.State.Targets) <= sess.Config.Global.Threads {
		threadNum = len(sess.State.Targets) - 1
	} else {
		threadNum = sess.Config.Global.Threads
	}
	wg.Add(threadNum)
	log.Debug("Threads for repository gathering: %d", threadNum)
	for i := 0; i < threadNum; i++ {
		go func() {
			for {
				target, ok := <-ch
				if !ok {
					wg.Done()
					return
				}
				repos, err := sess.Client.GetRepositoriesFromOwner(ctx, *target)
				if err != nil {
					log.Error(" Failed to retrieve repositories from %s: %s", *target.Login, err)
				}
				if len(repos) == 0 {
					continue
				}
				for _, repo := range repos {
					log.Debug(" Retrieved repository: %s", repo.CloneURL)
					sess.AddRepository(repo)
				}
				log.Info(" Retrieved %d %s from %s", len(repos), util.Pluralize(len(repos), "repository", "repositories"), *target.Login)
			}
		}()
	}

	for _, target := range sess.State.Targets {
		ch <- target
	}
	close(ch)
	wg.Wait()
}
