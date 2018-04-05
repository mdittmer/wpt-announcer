package github

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
)

var batchSize *int

func init() {
	bs := 100
	batchSize = &bs
}

func getCommitBatch(ctx context.Context, client *github.Client, hashes []string) (commits []*github.Commit, err error) {
	var wg sync.WaitGroup
	wg.Add(len(hashes))
	commits = make([]*github.Commit, len(hashes), len(hashes))
	for i, hash := range hashes {
		go func(i int, hash string) {
			defer wg.Done()
			var commit *github.Commit
			commit, _, goErr := client.Git.GetCommit(ctx, "w3c", "web-platform-tests", hash)
			if goErr != nil {
				err = goErr
				log.Errorf("Error fetching commit %s: %v", hash, err)
				commits[i] = nil
			} else {
				commits[i] = commit
			}
		}(i, hash)
	}
	wg.Wait()
	return commits, err
}

// GetCommits fetches []*github.Commit that corresponds to hashes.
func GetCommits(ctx context.Context, client *github.Client, hashes []string) (commits []*github.Commit, err error) {
	limit, err := GetCoreRateLimit(ctx, client)
	if err != nil {
		return commits, err
	}
	if len(hashes) > limit.Remaining {
		return commits, fmt.Errorf("GetCommits(): %d of %d remaining GitHub API requests insufficient for %d hashes. Limit will reset at %v", limit.Remaining, limit.Limit, len(hashes), limit.Reset.Time)
	}
	commits = make([]*github.Commit, 0, len(hashes))
	bs := *batchSize
	for i := 0; i < len(hashes); i += bs {
		bound := int(math.Min(float64(len(hashes)), float64(i+bs)))
		batch, _ := getCommitBatch(ctx, client, hashes[i:bound])
		commits = append(commits, batch...)
	}
	return commits, err
}
