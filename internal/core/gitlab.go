package core

import (
	"context"
	"errors"
	"net"
	"time"

	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/session"
	"github.com/rumenvasilev/rvsecret/internal/stats"
)

// GatherTargets will enumerate git targets adding them to a running target list. This will set the targets based
// on the scan type set within the cmd package.
func GatherTargets(sess *session.Session) {
	log := log.Log
	sess.State.Stats.UpdateStatus(stats.StatusGathering)
	log.Important("Gathering targets...")
	ctx := context.Background()

	// var targets []string

	// Based on the type of scan, set in the cmd package, we set a generic
	// variable to the specific targets
	//switch sess.ScanType {
	//case "github":
	//	targets = sess.GithubTargets
	//case "gitlab":
	targets := sess.Config.Gitlab.Targets
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
		//		log.Error(" Error retrieving information on %s: %s\n", loginOption, err)
		//		continue
		//	}
		//} else {
		target, err := sess.Client.GetUserOrganization(ctx, loginOption)
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

			log.Error(" Error retrieving information on %s: %s", loginOption, err)
			continue
		} else if target == nil {
			log.Error(" Error retrieving information on %s: %s", loginOption, err)
			continue
		}

		log.Debug("%s (ID: %d) type: %s", *target.Login, *target.ID, *target.Type)
		sess.State.AddTarget(target)
		// If forking is false AND the target type is an Organization as set above in GetUserOrganization
		if sess.Config.Global.ExpandOrgs && *target.Type == _coreapi.TargetTypeOrganization {
			log.Debug("Gathering members of %s (ID: %d)...", *target.Login, *target.ID)
			members, err := sess.Client.GetOrganizationMembers(ctx, *target)
			if err != nil {
				log.Error(" Error retrieving members of %s: %s", *target.Login, err)
				continue
			}
			// Add organization members gathered above to the target list
			// TODO Do we want to spider this out at some point to enumerate all members of an org?
			for _, member := range members {
				log.Debug("Adding organization member %s (ID: %d) to targets", *member.Login, *member.ID)
				sess.State.AddTarget(member)
			}
		}
	}
}
