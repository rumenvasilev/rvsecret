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
			b := FileExists(f)

			Convey("The function should return true", func() {
				So(b, ShouldEqual, true)
			})
			Convey("The function should not return false", func() {
				So(b, ShouldNotEqual, false)
			})
		})

		Convey("When the file does not exist", func() {
			f := "../NOPE.md"
			b := FileExists(f)

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
