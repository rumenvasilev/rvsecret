package signatures

import (
	"regexp"

	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/matchfile"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// SafeFunctionSignature holds the information about a safe function, that is used to detect and mitigate false positives
type SafeFunctionSignature struct {
	match *regexp.Regexp
	GenericSignature
}

// ExtractMatch is a placeholder to ensure min code complexity and allow the reuse of the functions
func (s SafeFunctionSignature) ExtractMatch(file matchfile.MatchFile, change *object.Change, scanType api.ScanType, log *log.Logger) (bool, map[string]int) {
	// var results map[string]int

	return false, nil
}

// Enable sets whether as signature is active or not
func (s SafeFunctionSignature) Enable() int {
	return s.enable
}

// ConfidenceLevel sets the confidence level of the pattern
func (s SafeFunctionSignature) ConfidenceLevel() int {
	return s.confidenceLevel
}

// Part sets the part of the file/path that is matched [ filename content extension ]
func (s SafeFunctionSignature) Part() string {
	return s.part
}

// Description sets the user comment of the signature
func (s SafeFunctionSignature) Description() string {
	return s.description
}

// SignatureID sets the id used to identify the signature. This id is immutable and generated
// from a has of the signature and is changed with every update to a signature.
func (s SafeFunctionSignature) SignatureID() string {
	return s.signatureid
}
