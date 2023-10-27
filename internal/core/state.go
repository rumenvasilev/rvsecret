package core

import (
	_coreapi "github.com/rumenvasilev/rvsecret/internal/core/api"
	"github.com/rumenvasilev/rvsecret/internal/core/finding"
)

// AddTargets would iterate over the list and call AddTarget to append each one to the state list
func (st *State) AddTargets(targets []*_coreapi.Owner) {
	for _, v := range targets {
		st.AddTarget(v)
	}
}

// AddTarget will add a new target to a session to be scanned during that session
func (st *State) AddTarget(target *_coreapi.Owner) {
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
	case _coreapi.TargetTypeOrganization:
		st.Stats.IncrementOrgs()
	case _coreapi.TargetTypeUser:
		st.Stats.IncrementUsers()
	}
	st.Stats.IncrementTargets()
}

// AddRepository will add a given repository to be scanned to a session. This counts as
// the total number of repos that have been gathered during a session.
func (st *State) AddRepository(repository *_coreapi.Repository) {
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
func (st *State) AddFinding(finding *finding.Finding) {
	st.Lock()
	defer st.Unlock()
	// const MaxStrLen = 100
	st.Findings = append(st.Findings, finding)
	st.Stats.IncrementFindingsTotal()
}

func (st *State) GetFindings() []*finding.Finding {
	return st.Findings
}
