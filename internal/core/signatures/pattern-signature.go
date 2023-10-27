package signatures

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	_git "github.com/rumenvasilev/rvsecret/internal/core/git"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/matchfile"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// PatternSignature holds the information about a pattern signature which is a regex used to match content within a file
type PatternSignature struct {
	match           *regexp.Regexp
	comment         string
	description     string
	part            string
	signatureid     string
	enable          int
	entropy         float64
	confidenceLevel int
}

// ExtractMatch will try and find a match within the content of the file.
func (s PatternSignature) ExtractMatch(file matchfile.MatchFile, change *object.Change, scanType api.ScanType, log *log.Logger) (bool, map[string]int) {
	switch s.part {
	case PartPath:
		return s.match.MatchString(file.Path), nil
	case PartFilename:
		return s.match.MatchString(file.Filename), nil
	case PartExtension:
		return s.match.MatchString(file.Extension), nil
	case PartContent:
		return s.partContent(file.Path, change, scanType, log)
	default: // TODO We need to do something with this
		return false, nil
	}
}

func (s PatternSignature) partContent(haystack string, change *object.Change, scanType api.ScanType, log *log.Logger) (bool, map[string]int) {
	var contextMatches []string
	results := make(map[string]int) // the secret and the line number in a map

	if !util.PathExists(haystack) {
		return false, nil
	}

	data, err := os.ReadFile(haystack)
	if err != nil {
		sErrAppend := fmt.Sprintf("ERROR --- Unable to open file for scanning: %q; Reason: %q", haystack, err)
		results[sErrAppend] = 1 // set to zero due to error, we never have a line 0 so we can always ignore that or error on it
		return false, results
	}

	// Check to see if there is a match in the data and if so switch to a Findall that
	// will get a slice of all the individual matches. Doing this ahead of time saves us
	// from looping through if it is not necessary.
	if s.match.Match(data) {
		for _, curRegexMatch := range s.match.FindAll(data, -1) {
			contextMatches = append(contextMatches, string(curRegexMatch))
		}
		if len(contextMatches) > 0 {
			return examineMatchResults(contextMatches, false, s.entropy, string(data))
		}
	}

	if scanType == api.LocalPath {
		return false, results
	}

	content, err := _git.GetChangeContent(change)
	if err != nil {
		log.Error("Error retrieving content in commit %s, change %s: %s", "commit.String()", change.String(), err)
	}

	if s.match.Match([]byte(content)) {
		for _, curRegexMatch := range s.match.FindAll([]byte(content), -1) {
			contextMatches = append(contextMatches, string(curRegexMatch))
		}
		if len(contextMatches) > 0 {
			return examineMatchResults(contextMatches, true, s.entropy, content)
		}
	}

	return false, nil
}

func examineMatchResults(contextMatches []string, dynamicMatch bool, entropy float64, content string) (bool, map[string]int) {
	bResult := true
	var results = make(map[string]int)
	var mn int
	for i, curMatch := range contextMatches {
		thisMatch := string(curMatch[:])
		thisMatch = strings.TrimSuffix(thisMatch, "\n")

		bResult = confirmEntropy(thisMatch, entropy)

		if bResult {
			mn = 0
			if dynamicMatch {
				mn = i
			}
			linesOfScannedFile := strings.Split(content, "\n")
			num := fetchLineNumber(&linesOfScannedFile, thisMatch, mn)
			results[strconv.Itoa(i)+"_"+thisMatch] = num
		}
	}
	return bResult, results
}

// Enable sets whether as signature is active or not
func (s PatternSignature) Enable() int {
	return s.enable
}

// ConfidenceLevel sets the confidence level of the pattern
func (s PatternSignature) ConfidenceLevel() int {
	return s.confidenceLevel
}

// Part sets the part of the file/path that is matched [ filename content extension ]
func (s PatternSignature) Part() string {
	return s.part
}

// Description sets the user comment of the signature
func (s PatternSignature) Description() string {
	return s.description
}

// SignatureID sets the id used to identify the signature. This id is immutable and generated from a has of the
// signature and is changed with every update to a signature.
func (s PatternSignature) SignatureID() string {
	return s.signatureid
}
