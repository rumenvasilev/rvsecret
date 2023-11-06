package matchfile

import (
	"fmt"
	"path/filepath"
	"strings"
)

// MatchFile holds the various parts of a file that will be matched using either regex's or simple pattern matches.
type MatchFile struct {
	Path      string
	Filename  string
	Extension string
}

// New will generate a match object by dissecting a filename
// It doesn't check if the file exist or not!
// Input string is not sanitized, it comes from git, so we treat it as safe.
// Might be a wrong assumption, but this app should always be run in a safe,
// patched environment as well as in a docker scratch container with mounted
// source code.
func New(path string) MatchFile {
	_, filename := filepath.Split(path)
	extension := filepath.Ext(path)
	return MatchFile{
		Path:      path,
		Filename:  filename,
		Extension: extension,
	}
}

// IsSkippable will check the matched file against a list of extensions or paths either supplied by the user or set by default
func (f *MatchFile) IsSkippable(skippableExt, skippablePath []string) bool {
	ext := strings.ToLower(f.Extension)
	path := strings.ToLower(f.Path)
	for _, v := range skippableExt {
		if ext == fmt.Sprintf(".%s", v) {
			return true
		}
	}
	for _, v := range skippablePath {
		if strings.Contains(path, v) {
			return true
		}
	}
	return false
}
