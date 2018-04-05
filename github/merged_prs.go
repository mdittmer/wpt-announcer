package github

import (
	"context"
	"flag"
	"math"
	"sort"
	"sync"

	"github.com/google/go-github/github"
)

var slackMergedPRs *int

func init() {
	slackMergedPRs = flag.Int("slack_merged_rs", 20, "Number of extra merged PRs sorted by creation date to ensure correct ordering by merged date")
}

type byMergedAt []*github.PullRequest

func (prs byMergedAt) Len() int {
	return len(prs)
}
func (prs byMergedAt) Swap(i, j int) {
	prs[i], prs[j] = prs[j], prs[i]
}
func (prs byMergedAt) Less(i, j int) bool {
	if prs[i].MergedAt == nil {
		return false
	}
	if prs[j].MergedAt == nil {
		return true
	}
	return prs[i].MergedAt.After(*prs[j].MergedAt)
}

func filterMerged(inPrs []*github.PullRequest) (outPrs []*github.PullRequest) {
	outPrs = inPrs[:0]
	for _, pr := range inPrs {
		if pr.MergedAt != nil {
			outPrs = append(outPrs, pr)
		}
	}
	return outPrs
}

func getPageOfMergedPRs(ctx context.Context, client *github.Client, page int) (prs []*github.PullRequest, err error) {
	prs, _, err = client.PullRequests.List(ctx, "w3c", "web-platform-tests", &github.PullRequestListOptions{
		State:     "closed",
		Base:      "master",
		Sort:      "created",
		Direction: "desc",
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: 100,
		},
	})
	if err != nil {
		return prs, err
	}
	prs = filterMerged(prs)
	return prs, err
}

func getMergedPRs(ctx context.Context, client *github.Client, num int) (prs []*github.PullRequest, err error) {
	numPages := int(math.Ceil(float64(num)/100) + 1)
	prs = make([]*github.PullRequest, 0, num)
	for len(prs) < num {
		pages := make([][]*github.PullRequest, numPages, numPages)
		var wg sync.WaitGroup
		wg.Add(numPages)
		for i := 0; i < numPages; i++ {
			go func(idx int) {
				defer wg.Done()
				pages[idx], err = getPageOfMergedPRs(ctx, client, idx+1)
			}(i)
		}
		wg.Wait()
		if err != nil {
			break
		}

		for i := 0; i < numPages && len(prs) < num; i++ {
			page := pages[i]
			l1, l2 := num-len(prs), len(page)
			var limit int
			if l1 <= l2 {
				limit = l1
			} else {
				limit = l2
			}
			prs = append(prs, page[:limit]...)
		}
		numPages = 1
	}
	return prs, err
}

// GetMergedPRs attempts to fetch a []*github.PullRequest of length num where all PRs are merged, and are interpreted in descending github.PullRequest.MergedAt.
func GetMergedPRs(ctx context.Context, client *github.Client, num int) (prs []*github.PullRequest, err error) {
	totalNum := num + *slackMergedPRs
	prs, err = getMergedPRs(ctx, client, totalNum)
	if err != nil {
		return prs, err
	}
	sort.Sort(byMergedAt(prs))
	return prs[:num], err
}
