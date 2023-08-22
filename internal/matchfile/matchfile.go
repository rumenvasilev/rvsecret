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

// newMatchFile will generate a match object by dissecting a filename
func New(path string) MatchFile {
	_, filename := filepath.Split(path)
	extension := filepath.Ext(path)
	return MatchFile{
		Path:      path,
		Filename:  filename,
		Extension: extension,
	}
}

// isSkippable will check the matched file against a list of extensions or paths either supplied by the user or set by default
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
