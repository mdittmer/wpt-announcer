package github

import (
	"context"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
)

func GetCoreRateLimit(ctx context.Context, client *github.Client) (rate *github.Rate, err error) {
	limits, _, err := client.RateLimits(ctx)
	if err != nil {
		return rate, err
	}
	limit := limits.Core
	log.Infof("%d of %d GitHub API requests remaining. Limit will reset at %v", limit.Remaining, limit.Limit, limit.Reset.Time)
	return limit, err
}
