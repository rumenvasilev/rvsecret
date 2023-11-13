package util

import (
	"fmt"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileExists(t *testing.T) {

	// Only pass t into top-level Convey calls
	Convey("Given a filename", t, func() {

		Convey("When the file exists", func() {
			f := "../../README.md"
			b := PathExists(f)

			Convey("The function should return true", func() {
				So(b, ShouldEqual, true)
			})
			Convey("The function should not return false", func() {
				So(b, ShouldNotEqual, false)
			})
		})

		Convey("When the file does not exist", func() {
			f := "../NOPE.md"
			b := PathExists(f)

			Convey("The function should return false", func() {
				So(b, ShouldEqual, false)
			})
			Convey("The function should not return true", func() {
				So(b, ShouldNotEqual, true)
			})
		})
	})
}

func Test_IsBinaryFile(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr string
	}{
		{"strings.go", "./strings.go", false, ""},
		{"compiled binary", findBinary(t), true, ""},
		{"non-existing file", "some/unexisting/file", false, "open some/unexisting/file: no such file or directory"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEmpty(t, tt.input, "Make sure you've run `make build` before running this test. It relies on the binary being present in bin directory.")
			got, err := IsBinaryFile(tt.input)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// finds the pre-built binary under the name <root>/bin/rvsecret*
func findBinary(t *testing.T) string {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	sp := strings.Split(cwd, "/")
	root := strings.Join(sp[:len(sp)-2], "/")
	binPath := fmt.Sprintf("%s/bin", root)
	dir, err := os.ReadDir(binPath)
	require.NoError(t, err, "Please run `make build` before running this test.")
	binaryName := ""
	for _, entry := range dir {
		if entry.IsDir() {
			continue
		} else if entry.Type().IsRegular() {
			if strings.Contains(entry.Name(), "rvsecret") {
				t.Logf("Found binary name %q in path %q", entry.Name(), binPath)
				binaryName = fmt.Sprintf("%s/%s", binPath, entry.Name())
				break
			}
		}
	}

	return binaryName
}

// go test -run XXX -bench=. -benchmem
// goos: darwin
// goarch: amd64
// pkg: github.com/rumenvasilev/rvsecret/internal/util
// cpu: Intel(R) Core(TM) i7-5557U CPU @ 3.10GHz
// BenchmarkIsBinaryFile-4   	   92380	     12687 ns/op	     136 B/op	       3 allocs/op
// PASS
// ok  	github.com/rumenvasilev/rvsecret/internal/util	1.323s
func BenchmarkIsBinaryFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		IsBinaryFile("../../bin/rvsecret-darwin")
	}
}

func TestGetYamlFiles(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		dir := "../../testfixtures/yamlfiles"
		want := []string{dir + "/file-2.yml", dir + "/file1.yaml"}
		got, err := GetYamlFiles(dir)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("no files found", func(t *testing.T) {
		dir := "./"
		got, err := GetYamlFiles(dir)
		assert.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("err", func(t *testing.T) {
		dir := "none/existing/dir"
		got, err := GetYamlFiles(dir)
		assert.Error(t, err)
		assert.Empty(t, got)
	})
}

func TestCopyFiles(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		src := "../../testfixtures/yamlfiles"
		dst, err := os.MkdirTemp("", "TestCopyFilesOK*")
		defer os.RemoveAll(dst)
		require.NoError(t, err)

		err = CopyFiles(src, dst)
		assert.NoError(t, err)
	})

	t.Run("err", func(t *testing.T) {
		src := "none/existing/dir"
		dst, err := os.MkdirTemp("", "TestCopyFilesErr*")
		defer os.RemoveAll(dst)
		require.NoError(t, err)

		err = CopyFiles(src, dst)
		assert.Error(t, err)
	})
}

func TestIsMaxFileSize(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		got, msg := IsMaxFileSize("./io.go", 1)
		assert.False(t, got)
		assert.Empty(t, msg)
	})

	t.Run("file does not exist", func(t *testing.T) {
		got, msg := IsMaxFileSize("./no-file", 1)
		assert.False(t, got)
		assert.Equal(t, "stat ./no-file: no such file or directory", msg)
	})

	t.Run("negative input, too large file", func(t *testing.T) {
		got, msg := IsMaxFileSize("./io.go", -1)
		assert.True(t, got)
		assert.Equal(t, "is too large", msg)
	})

}

func TestWriteToFile(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		dst, err := os.MkdirTemp("", "TestWriteToFileOK*")
		defer os.RemoveAll(dst)
		require.NoError(t, err)

		fh := dst + "/sample.txt"
		err = WriteToFile(fh, []byte("just-testing-meh"))
		assert.NoError(t, err)
	})

	t.Run("no input", func(t *testing.T) {
		dst, err := os.MkdirTemp("", "TestWriteToFileOKNoInput*")
		defer os.RemoveAll(dst)
		require.NoError(t, err)

		fh := dst + "/sample.txt"
		err = WriteToFile(fh, []byte{})
		assert.NoError(t, err)
	})

	t.Run("incorrect path", func(t *testing.T) {
		err := WriteToFile("", []byte("just-testing-meh"))
		assert.Error(t, err)
		assert.EqualError(t, err, "open : no such file or directory")
	})
}

func TestSetHomeDir(t *testing.T) {
	t.Skip()
}

func TestMakeHomeDir(t *testing.T) {
	t.Skip()
}

func Test_IsTestFileOrPath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"foo/test/bar", true},
		{"test/foo/bar", true},
		{"foo/test-secrets/bar", true},
		{"ghTestlk", true},
		{"Testllfhe", true},
		{"Test", true},
		{"foo_test.go", true},
		{"foo_test_baz", true},
		{"../../testfixtures/yamlfiles", true},
		{"/testfixtures/yamlfiles", true},
		{"thisShouldnt_Match.zip", false},
		{"/root/user/must/not.match", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsTestFileOrPath(tt.input)
			if tt.want {
				assert.True(t, got)
			} else {
				assert.False(t, got)
			}
		})
	}
}
