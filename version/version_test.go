package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppVersion(t *testing.T) {
	got := AppVersion()
	want := "0.1.0"
	assert.Equal(t, want, got)
}
