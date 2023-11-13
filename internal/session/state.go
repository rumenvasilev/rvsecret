package session

import (
	"sync"

	coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/core/finding"
	"github.com/rumenvasilev/rvsecret/internal/stats"
)

type State struct {
	*sync.Mutex
	Stats        *stats.Stats
	Findings     map[string]*finding.Finding
	Targets      []*coreapi.Owner
	Repositories []*coreapi.Repository
}

// AddTargets would iterate over the list and call AddTarget to append each one to the state list
func (st *State) AddTargets(targets []*coreapi.Owner) {
	for _, v := range targets {
		st.AddTarget(v)
	}
}

// AddTarget will add a new target to a session to be scanned during that session
func (st *State) AddTarget(target *coreapi.Owner) {
	st.Lock()
	defer st.Unlock()
	for _, t := range st.Targets {
		if *target.ID == *t.ID {
			return
		}
	}
	st.Targets = append(st.Targets, target)

	// Update statistics
	switch *target.Kind {
	case coreapi.TargetTypeOrganization:
		st.Stats.IncrementOrgs()
	case coreapi.TargetTypeUser:
		st.Stats.IncrementUsers()
	}
	st.Stats.IncrementTargets()
}

// AddRepository will add a given repository to be scanned to a session. This counts as
// the total number of repos that have been gathered during a session.
func (st *State) AddRepository(repository *coreapi.Repository) {
	st.Lock()
	defer st.Unlock()
	for _, r := range st.Repositories {
		if repository.ID == r.ID {
			return
		}
	}
	st.Repositories = append(st.Repositories, repository)

}

// AddFinding will add a finding that has been discovered during a session to the list of findings
// for that session
func (st *State) AddFinding(finding *finding.Finding) bool {
	st.Lock()
	defer st.Unlock()
	// const MaxStrLen = 100
	// st.Findings = append(st.Findings, finding)
	// No need to append another finding of the same
	// TODO perhaps make the rules that matched a list
	if _, ok := st.Findings[finding.SecretID]; ok {
		return false
	}
	st.Findings[finding.SecretID] = finding
	st.Stats.IncrementFindingsTotal()
	return true
}

func (st *State) GetFindings() []*finding.Finding {
	var res []*finding.Finding
	for _, f := range st.Findings {
		res = append(res, f)
	}
	return res
}
