package assets

import (
	"embed"
	"net/http"
	"strings"
)

//go:embed static/*
var Assets embed.FS

// binaryFS  holds a filesystem handle
type BinaryFS struct {
	fs http.FileSystem
}

// Open will return an http file object that refers to a given file
func (b *BinaryFS) Open(name string) (http.File, error) {
	return b.fs.Open(name)
}

// Exists checks if a given file with a given prefix exists and attempts to open it
func (b *BinaryFS) Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		if _, err := b.fs.Open(p); err != nil {
			return false
		}
		return true
	}
	return false
}

func GetStaticFS() *BinaryFS {
	// serverRoot, err := fs.Sub(&Assets, "static")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// static.ServeFileSystem
	return &BinaryFS{
		http.FS(&Assets),
	}
}
