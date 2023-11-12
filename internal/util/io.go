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
	"unicode/utf8"

	"github.com/mitchellh/go-homedir"
	cp "github.com/otiai10/copy"
)

var magicNumbers = [][]byte{
	{0x1F, 0x8B, 0x08, 0x00},                         // GZip
	{0x42, 0x5A, 0x68, 0x32},                         // BZip2
	{0x50, 0x4B, 0x03, 0x04},                         // ZIP
	{0x89, 0x50, 0x4E, 0x47},                         // PNG
	{0x4D, 0x5A},                                     // Windows EXE
	{0x7F, 'E', 'L', 'F'},                            // Linux ELF Executable
	{0xFE, 0xED, 0xFA, 0xCE, 0xCE, 0xFA, 0xED, 0xFE}, // macOS Mach-O Binary
	{0xFE, 0xED, 0xFA, 0xCF, 0x0C, 0x00, 0x00, 0x01}, // Mach-O 64-bit (x86_64)
}

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

// CopyFiles will copy files from src to dest directory and attempt to set correct permissions
func CopyFiles(src, dest string) error {
	if err := cp.Copy(src, dest); err != nil {
		return err
	}

	sigs, err := GetSignatureFiles(dest)
	if err != nil {
		return err
	}

	// set them to the current user and the proper permissions
	for _, f := range sigs {
		if err := os.Chmod(f, 0644); err != nil {
			return err
		}
	}
	return nil
}

// GetSignatureFiles will find all the yaml files in the signatures directory
func GetSignatureFiles(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var sigs []string
	var ext string
	for _, f := range files {
		ext = filepath.Ext(dir + "/" + f.Name())
		if ext == "yml" || ext == "yaml" {
			sigs = append(sigs, fmt.Sprintf("%s/%s", dir, f.Name()))
		}
	}
	return sigs, nil
}

func IsBinaryFile(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read the first 4 bytes to identify file type
	buffer := make([]byte, 4)
	_, err = file.Read(buffer)
	if err != nil {
		return false, err
	}

	// Check for common binary file magic numbers
	for _, magic := range magicNumbers {
		if bytesMatch(buffer, magic) {
			return true, nil
		}
	}

	// The implementation above doesn't catch all binaries, one example being go compiled
	// binaries for darwin (macos), but unicode test does.
	// https://groups.google.com/g/golang-nuts/c/YeLL7L7SwWs/m/LGlsc9GIJlUJ
	// if the encoding is invalid, it returns (RuneError, 1)
	runerr, p := utf8.DecodeRune(buffer)
	if runerr == utf8.RuneError {
		if p == 0 || p == 1 {
			return true, nil
		}
	}

	return false, nil
}

func bytesMatch(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
