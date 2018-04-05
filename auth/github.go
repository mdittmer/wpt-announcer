package auth

import (
	"context"
	"flag"
	"io/ioutil"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var githubAccessTokenPath *string
var githubAccessToken *string

func init() {
	githubAccessTokenPath = flag.String("access_token_path", "./.local/github_access_token", "Path to file containing GitHub access token")
	githubAccessToken = flag.String("access_token", "", "GitHub access token")
}

// GetGithubClient gets a GitHub client, authenticated according to flags.
func GetGithubClient(ctx context.Context) (client *github.Client) {
	if *githubAccessToken == "" {
		log.Infof("Loading GitHub access token from %s", *githubAccessTokenPath)
		rawGithubAccessToken, err := ioutil.ReadFile(*githubAccessTokenPath)
		if err != nil {
			log.Errorf("Failed to load GitHub access token from %s", *githubAccessTokenPath)
		} else {
			strGithubAccessToken := string(rawGithubAccessToken)
			githubAccessToken = &strGithubAccessToken
		}
	}

	if *githubAccessToken == "" {
		client = github.NewClient(nil)
	} else {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *githubAccessToken},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}

	return client
}
