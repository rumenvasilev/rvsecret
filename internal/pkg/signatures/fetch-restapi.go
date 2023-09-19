package signatures

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/rumenvasilev/rvsecret/internal/core"
	_github "github.com/rumenvasilev/rvsecret/internal/core/provider/github"
	"github.com/rumenvasilev/rvsecret/version"
)

// fetchSignaturesFromGithubAPI will only download a version of the signatures file from Github REST API
func fetchSignaturesFromGithubAPI(version string, sess *core.Session) (string, error) {
	ctx := context.Background()
	if sess.Config.Signatures.UserRepo == "" {
		return "", fmt.Errorf("please provide -signatures-user-repo value")
	}

	res := strings.Split(sess.Config.Signatures.UserRepo, "/")
	if len(res) != 2 {
		return "", fmt.Errorf("user/repo doesn't have matching format, %s", sess.Config.Signatures.UserRepo)
	}
	owner := res[0]
	repo := res[1]

	client, err := _github.NewClient(sess.Config.Signatures.APIToken, "", sess.Out)
	if err != nil {
		return "", fmt.Errorf("failed instantiation of Github client, %w", err)
	}

	var assets *github.RepositoryRelease
	if version == "latest" {
		assets, err = client.GetLatestRelease(ctx, owner, repo)
	} else {
		assets, err = client.GetReleaseByTag(ctx, owner, repo, version)
	}
	if err != nil {
		// TODO: handle 404 not found
		return "", fmt.Errorf("error while fetching release information, %w", err)
	}

	assetURL, err := getAssetURL(assets.Assets)
	if err != nil {
		return "", err
	}

	return downloadAsset(assetURL, sess)
}

func getAssetURL(assets []github.ReleaseAsset) (string, error) {
	var download string
	for _, v := range assets {
		if v.GetName() == "default.yaml" {
			download = v.GetURL()
			break
		}
	}
	if download == "" {
		return "", fmt.Errorf("couldn't find the release asset default.yaml")
	}
	return download, nil
}

func downloadAsset(url string, sess *core.Session) (string, error) {
	// Create tmp dir
	path, err := os.MkdirTemp("", "rvsecret")
	if err != nil {
		return "", err
	}

	err = os.Mkdir(fmt.Sprintf("%s/signatures", path), 0700)
	if err != nil {
		return "", err
	}

	// fetch from URL
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("token %s", sess.Config.Signatures.APIToken))
	req.Header.Add("User-Agent", version.UserAgent)
	req.Header.Add("Accept", "application/octet-stream")

	// call github
	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint:errcheck

	switch resp.StatusCode {
	case 200:
	default:
		return "", &github.ErrorResponse{Response: resp}
	}

	// store file
	filename := fmt.Sprintf("%s/signatures/default.yaml", path)
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	defer f.Close() //nolint:staticcheck
	if err != nil {
		return "", err
	}

	b := make([]byte, 4096)
	var i int
	for err == nil {
		i, err = resp.Body.Read(b)
		f.Write(b[:i]) //nolint:errcheck
	}

	return path, nil
}
