package signatures

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/matchfile"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/yaml.v2"
)

// These are the various items that we are attempting to match against using either regex's or simple pattern matches.
const (
	PartExtension = "extension" // file extension
	PartFilename  = "filename"  // file name
	PartPath      = "path"      // the path to the file
	PartContent   = "content"   // the content of the file
)

type signatureKind int

const (
	_ signatureKind = iota
	simpleKind
	patternKind
	safeFunctionKind
)

// WARNING, GLOBAL VAR!
// Signatures holds a list of all signatures used during the session
var Signatures []Signature

// SafeFunctionSignatures is a collection of safe function sigs
var SafeFunctionSignatures []SafeFunctionSignature

// loadSignatureSet will read in the defined signatures from an external source
func loadSignatureSet(filename string) (SignatureConfig, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return SignatureConfig{}, err
	}

	var c SignatureConfig
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		return SignatureConfig{}, err
	}

	return c, nil
}

// Signature is an expression that we are looking for in a file
type Signature interface {
	Description() string
	Enable() int
	ExtractMatch(file matchfile.MatchFile, change *object.Change, scanType api.ScanType, log *log.Logger) (bool, map[string]int)
	ConfidenceLevel() int
	Part() string
	SignatureID() string // TODO change id -> ID
}

// SignaturesMetaData is used by updateSignatures to determine if/how to update the signatures
type SignaturesMetaData struct {
	Date    string
	Version string
	Time    int
}

// SignatureDef maps to a signature within the yaml file
type SignatureDef struct {
	Comment         string  `yaml:"comment"`
	Description     string  `yaml:"description"`
	Match           string  `yaml:"match"`
	Part            string  `yaml:"part"`
	SignatureID     string  `yaml:"signatureid"`
	Enable          int     `yaml:"enable"`
	Entropy         float64 `yaml:"entropy"`
	ConfidenceLevel int     `yaml:"confidence-level"`
}

// SignatureConfig holds the base file structure for the signatures file
type SignatureConfig struct {
	Meta                   SignaturesMetaData `yaml:"Meta"`
	PatternSignatures      []SignatureDef     `yaml:"PatternSignatures"`
	SimpleSignatures       []SignatureDef     `yaml:"SimpleSignatures"`
	SafeFunctionSignatures []SignatureDef     `yaml:"SafeFunctionSignatures"`
}

// IsSafeText check against known "safe" (aka not a password) list
func IsSafeText(sMatchString *string) bool {
	bResult := false
	for _, safeSig := range SafeFunctionSignatures {
		if safeSig.match.MatchString(*sMatchString) {
			bResult = true
		}
	}
	return bResult
}

// confirmEntropy will determine correct entrophy of the string and decide if we move forward with the match
func confirmEntropy(thisMatch string, iSessionEntropy float64) bool {
	bResult := false

	iEntropy := util.GetEntropyInt(thisMatch)

	if (iSessionEntropy == 0) || (iEntropy >= iSessionEntropy) {
		if !IsSafeText(&thisMatch) {
			bResult = true
		}
	}

	return bResult
}

// fetchLineNumber will read a file line by line and when the match is found, save the line number.
// It manages multiple matches in a file by way of the count and an index
func fetchLineNumber(input *[]string, thisMatch string, idx int) int {
	linesOfScannedFile := *input
	lineNumIndexMap := make(map[int]int)

	count := 0

	for i, line := range linesOfScannedFile {
		if strings.Contains(line, thisMatch) {

			// We need to add 1 here as the index starts at zero so every line number would be line -1 normally
			lineNumIndexMap[count] = i + 1
			count = count + 1
		}
	}
	return lineNumIndexMap[idx]
}

// Load will load all known signatures for the various match types into the session
func Load(filePath string, mLevel int) ([]Signature, string, error) { // TODO we don't need to bring in session here
	// ensure that we have the proper home directory
	fp, err := util.SetHomeDir(filePath)
	if err != nil {
		return []Signature{}, "", err
	}

	c, err := loadSignatureSet(fp)
	if err != nil {
		return []Signature{}, "", fmt.Errorf("failed to load signatures file %s: %w", filePath, err)
	}
	signaturesVersion := c.Meta.Version
	// signaturesMetaData := SignaturesMetaData{
	// 	Version: c.Meta.Version,
	// 	Date:    c.Meta.Date,
	// 	Time:    c.Meta.Time,
	// }

	// sess.SignatureVersion = signaturesMetaData.Version

	var SimpleSignatures []SimpleSignature
	var PatternSignatures []PatternSignature
	for _, curSig := range c.SimpleSignatures {
		res, ok := processSignatures(curSig, mLevel, simpleKind).(SimpleSignature)
		if res == (SimpleSignature{}) || !ok {
			continue
		}
		SimpleSignatures = append(SimpleSignatures, res)
	}

	for _, curSig := range c.PatternSignatures {
		res, ok := processSignatures(curSig, mLevel, patternKind).(PatternSignature)
		if res == (PatternSignature{}) || !ok {
			continue
		}
		PatternSignatures = append(PatternSignatures, res)
	}

	for _, curSig := range c.SafeFunctionSignatures {
		res, ok := processSignatures(curSig, mLevel, safeFunctionKind).(SafeFunctionSignature)
		if res == (SafeFunctionSignature{}) || !ok {
			continue
		}
		SafeFunctionSignatures = append(SafeFunctionSignatures, res)
	}

	idx := len(PatternSignatures) + len(SimpleSignatures)

	Signatures := make([]Signature, idx)
	jdx := 0
	for _, v := range SimpleSignatures {
		Signatures[jdx] = v
		jdx++
	}

	for _, v := range PatternSignatures {
		Signatures[jdx] = v
		jdx++
	}

	// TODO are we loading the safe ones somewhere

	return Signatures, signaturesVersion, nil
}

func processSignatures(curSig SignatureDef, mLevel int, kind signatureKind) interface{} {
	if curSig.Enable > 0 && curSig.ConfidenceLevel >= mLevel {
		switch kind {
		case simpleKind:
			return SimpleSignature{
				match:           curSig.Match,
				comment:         curSig.Comment,
				description:     curSig.Description,
				part:            getPart(curSig),
				signatureid:     curSig.SignatureID,
				enable:          curSig.Enable,
				entropy:         curSig.Entropy,
				confidenceLevel: curSig.ConfidenceLevel,
			}
		case patternKind:
			return PatternSignature{
				match:           regexp.MustCompile(curSig.Match),
				comment:         curSig.Comment,
				description:     curSig.Description,
				part:            getPart(curSig),
				signatureid:     curSig.SignatureID,
				enable:          curSig.Enable,
				entropy:         curSig.Entropy,
				confidenceLevel: curSig.ConfidenceLevel,
			}
		case safeFunctionKind:
			return SafeFunctionSignature{
				match:           regexp.MustCompile(curSig.Match),
				comment:         curSig.Comment,
				description:     curSig.Description,
				part:            getPart(curSig),
				signatureid:     curSig.SignatureID,
				enable:          curSig.Enable,
				entropy:         curSig.Entropy,
				confidenceLevel: curSig.ConfidenceLevel,
			}
		}
	}
	return nil
}

func getPart(sigDef SignatureDef) string {
	switch strings.ToLower(sigDef.Part) {
	case "partpath":
		return PartPath
	case "partfilename":
		return PartFilename
	case "partextension":
		return PartExtension
	case "partcontent":
		return PartContent
	default:
		return PartContent
	}
}

type DiscoverOutput struct {
	Sig     Signature
	Content string
	LineNum int
}

func Discover(mf matchfile.MatchFile, change *object.Change, cfg *config.Config, log *log.Logger) (dirtyFile bool, dirtyCommit bool, results []DiscoverOutput) {
	var content string
	// for each signature that is loaded scan the file as a whole and generate a map of
	// the match and the line number the match was found on
	for _, sig := range Signatures {
		ok, matchMap := sig.ExtractMatch(mf, change, cfg.Global.ScanType, log)
		if !ok {
			continue
		}

		dirtyFile = true
		dirtyCommit = true

		// For every instance of the secret that matched the specific signatures
		// create a new finding. This will produce dupes as the file may exist
		// in multiple commits.
		for k, v := range matchMap {
			// Default to no content, only publish information if explicitly allowed to
			content = ""
			if matchMap != nil && !cfg.Global.HideSecrets {
				// This sets the content for the finding, in this case the actual secret
				// is the content. This can be removed and hidden via a commandline flag.
				cleanK := strings.SplitAfterN(k, "_", 2)
				content = cleanK[1]
			}

			results = append(results, DiscoverOutput{Content: content, Sig: sig, LineNum: v})
		}
	}
	return //dirtyFile, dirtyCommit, results
}
