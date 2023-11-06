package matchfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		path string
		want MatchFile
	}{
		{"ok", "/tmp/file.exe", MatchFile{Path: "/tmp/file.exe", Filename: "file.exe", Extension: ".exe"}},
		{"special chars also allowed and not escaped", "/tmp/&fil!.exe", MatchFile{Path: "/tmp/&fil!.exe", Filename: "&fil!.exe", Extension: ".exe"}},
		{"shellshock", "env x='() { :;}; echo vulnerable'", MatchFile{Path: "env x='() { :;}; echo vulnerable'", Filename: "env x='() { :;}; echo vulnerable'", Extension: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.path)
			assert.NotEmpty(t, got)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMatchFile_IsSkippable(t *testing.T) {
	type args struct {
		skippableExt  []string
		skippablePath []string
	}
	tests := []struct {
		name string
		mf   *MatchFile
		args args
		want bool
	}{
		{"extension skippable", &MatchFile{Path: "/path/to/s3cr!t.bin", Filename: "s3cr1t.bin", Extension: ".bin"}, args{skippableExt: []string{"bin"}}, true},
		{"path skippable", &MatchFile{Path: "/path/to/s3cr!t.bin", Filename: "s3cr1t.bin", Extension: ".bin"}, args{skippablePath: []string{"/path/to/s3cr!t.bin"}}, true},
		{"no extension", &MatchFile{Path: "/some/random/path", Filename: "path"}, args{skippableExt: []string{"path"}}, false},
		{"extension doesn't match", &MatchFile{Path: "/some/random/path.file", Filename: "path", Extension: "file"}, args{skippableExt: []string{"path"}}, false},
		{"broken path, contains part of it, it's a match", &MatchFile{Path: "&broken^path/haha", Filename: "haha"}, args{skippablePath: []string{"haha"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mf.IsSkippable(tt.args.skippableExt, tt.args.skippablePath)
			assert.Equal(t, tt.want, got)
		})
	}
}
