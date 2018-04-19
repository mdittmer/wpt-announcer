package main

import (
	"context"
	"flag"

	"github.com/google/go-github/github"
	"github.com/mdittmer/wpt-announcer/auth"
	"github.com/mdittmer/wpt-announcer/epoch"
	gh "github.com/mdittmer/wpt-announcer/github"
	log "github.com/sirupsen/logrus"
)

var numPRs *int

func init() {
	numPRs = flag.Int("num_prs", 500, "Number of merged PRs to fetch for epoch finding")
}

func getEpochal(prs []*github.PullRequest, epoch epoch.Epoch) *github.PullRequest {
	for i, next := range prs {
		if i == 0 {
			continue
		}
		prev := prs[i-1]
		if epoch.IsEpochal(*prev.MergedAt, *next.MergedAt) {
			return prev
		}
	}
	return nil
}

func main() {
	flag.Parse()

	ctx := context.Background()
	client := auth.GetGithubClient(ctx)

	prs, err := gh.GetMergedPRs(ctx, client, *numPRs)
	if err != nil {
		log.Fatalf("Failed to fetch PRs: %v", err)
	}
	log.Infof("Fetched %d PRs", len(prs))
	epochs := []epoch.Epoch{
		epoch.Monthly{},
		epoch.Weekly{},
		epoch.Daily{},
		epoch.Hourly{},
	}
	chosenPRs := make([]*github.PullRequest, 0, len(epochs))
	for _, epoch := range epochs {
		pr := getEpochal(prs, epoch)
		if pr == nil {
			log.Fatal("Failed to find epochal PR")
		}
		chosenPRs = append(chosenPRs, pr)
	}
	for _, pr := range chosenPRs {
		log.Infof("Chose PR %d merged at %v with %v", *pr.Number, *pr.MergedAt, *pr.MergeCommitSHA)
	}
}
