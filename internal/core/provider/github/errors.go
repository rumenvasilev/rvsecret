package github

import (
	"errors"
	"net/http"

	"github.com/google/go-github/github"
)

func IsCredentialsError(err error) bool {
	var gherr *github.ErrorResponse
	if errors.As(err, &gherr) {
		if gherr.Response.StatusCode == http.StatusUnauthorized {
			return true
		}
	}
	return false
}
