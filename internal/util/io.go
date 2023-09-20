package util

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/mitchellh/go-homedir"
)

// TODO THIS FUNC HAS TO RETURN ERROR, OTHERWISE WE DO THE SAME CHECK AGAIN LATER
// PathExists will check if a path exists or not and is used to validate user input
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOSPC {
		return false
	}

	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		// logger.Debug("Path does not exist: %s", err.Error())
		return false
	}

	return true
}

// FileExists will check for the existence of a file and return a bool depending
// on if it exists in a given path or not.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		// File does not exist?
		return !errors.Is(err, fs.ErrNotExist)
	}
	return true
}

// SetHomeDir will set the correct homedir.
func SetHomeDir(h string) (string, error) {
	if strings.Contains(h, "$HOME") {
		home, err := homedir.Dir()
		if err != nil {
			return "", err
		}

		h = strings.Replace(h, "$HOME", home, -1)
	}

	if strings.Contains(h, "~") {
		home, err := homedir.Dir()
		if err != nil {
			return "", err
		}
		h = strings.Replace(h, "~", home, -1)
	}
	return h, nil
}

// IsMaxFileSize will determine if the file size is under the max limit set by maxFileSize
func IsMaxFileSize(filename string, maxFileSize int64) (bool, string) {

	fi, err := os.Stat(filename)

	// This error occurs when the file is not found.
	// The source of truth for files traversed comes from:
	// 		git - the commit history (or)
	// 		filepath - walking the filepath.
	//
	// In the case of filepath, it can be safely assumed that the file will always exist because we only check
	// the files that were walked.
	//
	// In the case of git, it can be assumed that the file will exist somewhere in the commit history.
	// Thereforce, we assume that the file size is within the limit and return false.
	if _, ok := err.(*os.PathError); ok {
		return false, "does not exist"
	}

	fileSize := fi.Size()
	mfs := maxFileSize * 1024 * 1024

	if fileSize > mfs {
		return true, "is too large"
	}
	return false, ""
}

// IsTestFileOrPath will run various regex's against a target to determine if it is a test file or contained in a test directory.
func IsTestFileOrPath(fullPath string) bool {
	fName := filepath.Base(fullPath)

	// If the directory contains "test"
	// Ex. foo/test/bar
	r := regexp.MustCompile(`(?i)[/\\]test?[/\\]`)
	if r.MatchString(fullPath) {
		return true
	}

	// If the directory starts with test, the leading slash gets dropped by default
	// Ex. test/foo/bar
	r = regexp.MustCompile(`(?i)test?[/\\]`)
	if r.MatchString(fullPath) {
		return true
	}

	// If the directory path starts with a different root but has the word test in it somewhere
	// Ex. foo/test-secrets/bar
	r = regexp.MustCompile(`/test.*/`)
	if r.MatchString(fullPath) {
		return true
	}

	// A the word Test is in the string, case sensitive
	// Ex. ghTestlk
	// Ex. Testllfhe
	// Ex. Test
	r = regexp.MustCompile(`Test`)
	if r.MatchString(fName) {
		return true
	}

	// A file has a suffix of _test
	// Golang uses this as the default test file naming convention
	//Ex. foo_test.go
	r = regexp.MustCompile(`(?i)_test`)
	if r.MatchString(fName) {
		return true
	}

	// If the pattern _test_ is in the string
	// Ex. foo_test_baz
	r = regexp.MustCompile(`(?i)_test?_`)
	return r.MatchString(fName)
}

func MakeHomeDir(path string) (string, error) {
	dir, err := SetHomeDir(path)
	if err != nil {
		return "", err
	}

	if !PathExists(dir) {
		// create
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return "", err
		}
	}

	return dir, nil
}

// WriteToFile will create a new file or truncate the existing one and write the input byte stream.
func WriteToFile(path string, input []byte) error {
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = fh.Write(input)
	if err != nil {
		return fmt.Errorf("failed writing to configuration file, %w", err)
	}
	return nil
}
