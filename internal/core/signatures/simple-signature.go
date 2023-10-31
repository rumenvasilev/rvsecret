package signatures

import (
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/matchfile"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// SimpleSignature holds the information about a simple signature which is used to match a path or filename
type SimpleSignature struct {
	match string
	GenericSignature
}

// ExtractMatch will attempt to match a path or file name of the given file
func (s SimpleSignature) ExtractMatch(file matchfile.MatchFile, change *object.Change, scanType api.ScanType, log *log.Logger) (bool, map[string]int) {
	var haystack string

	switch s.part {
	case PartPath:
		haystack = file.Path
	case PartFilename:
		haystack = file.Filename
	case PartExtension:
		haystack = file.Extension
	default:
		return false, nil
	}

	return s.match == haystack, nil
}

// Enable sets whether as signature is active or not
func (s SimpleSignature) Enable() int {
	return s.enable
}

// ConfidenceLevel sets the confidence level of the pattern
func (s SimpleSignature) ConfidenceLevel() int {
	return s.confidenceLevel
}

// Part sets the part of the file/path that is matched [ filename content extension ]
func (s SimpleSignature) Part() string {
	return s.part
}

// Description sets the user comment of the signature
func (s SimpleSignature) Description() string {
	return s.description
}

// SignatureID sets the id used to identify the signature. This id is immutable and generated from a
// has of the signature and is changed with every update to a signature.
func (s SimpleSignature) SignatureID() string {
	return s.signatureid
}
