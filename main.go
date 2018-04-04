package main

import (
	"context"
	"flag"
	"io/ioutil"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var githubAccessTokenPath *string
var githubAccessToken *string
var numPRs *int

type EpochalPredicate func(prev *time.Time, next *time.Time) bool

type Epoch struct {
	Prev        *Epoch
	MaxDuration time.Duration
	isEpochal   EpochalPredicate
}

func (e Epoch) IsEpochal(prev *time.Time, next *time.Time) bool {
	return (e.Prev != nil && e.Prev.IsEpochal(prev, next)) || e.isEpochal(prev, next)
}

func init() {
	githubAccessTokenPath = flag.String("access_token_path", "./.local/github_access_token", "Path to file containing GitHub access token")
	githubAccessToken = flag.String("access_token", "", "GitHub access token")
	numPRs = flag.Int("num_prs", 500, "Number of merged PRs to fetch for epoch finding")

	flag.Parse()
}

func GetMonthly() Epoch {
	return Epoch{
		MaxDuration: time.Hour * 24 * 31,
		isEpochal: func(prev *time.Time, next *time.Time) bool {
			if prev.Year() != next.Year() {
				return true
			}
			return prev.Month() != next.Month()
		},
	}
}

func GetWeekly() Epoch {
	return Epoch{
		MaxDuration: time.Hour * 24 * 31,
		isEpochal: func(prev *time.Time, next *time.Time) bool {
			if next.Sub(*prev).Hours() >= 24*7 {
				return true
			}
			return prev.Weekday() > next.Weekday()
		},
	}
}

func GetDaily() Epoch {
	return Epoch{
		MaxDuration: time.Hour * 24,
		isEpochal: func(prev *time.Time, next *time.Time) bool {
			if next.Sub(*prev).Hours() >= 24 {
				return true
			}
			return prev.Day() != next.Day()
		},
	}
}

func GetHourly() Epoch {
	return Epoch{
		MaxDuration: time.Hour,
		isEpochal: func(prev *time.Time, next *time.Time) bool {
			if next.Sub(*prev).Hours() >= 1 {
				return true
			}
			return prev.Hour() != next.Hour()
		},
	}
}

func GetEpochs() epochSlice {
	return LinkEpochs([]Epoch{GetMonthly(), GetWeekly(), GetDaily(), GetHourly()})
}

type epochSlice []Epoch

func LinkEpochs(epochs []Epoch) epochSlice {
	for i, e := range epochs {
		if i != 0 {
			e.Prev = &epochs[i-1]
		}
	}
	return epochSlice(epochs)
}

func getEpochal(prs []*github.PullRequest, epoch Epoch) *github.PullRequest {
	for i, next := range prs {
		if i == 0 {
			continue
		}
		prev := prs[i-1]
		if epoch.IsEpochal(prev.MergedAt, next.MergedAt) {
			return prev
		}
	}
	return nil
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

func getMergedPRs(ctx context.Context, client *github.Client, page int) (prs []*github.PullRequest, err error) {
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

func GetMergedPRs(ctx context.Context, client *github.Client, num int) (prs []*github.PullRequest, err error) {
	numPages := int(math.Ceil(float64(num)/100) + 1)
	prs = make([]*github.PullRequest, 0, num)
	for len(prs) < num {
		pages := make([][]*github.PullRequest, numPages, numPages)
		var wg sync.WaitGroup
		wg.Add(numPages)
		for i := 0; i < numPages; i++ {
			go func(idx int) {
				defer wg.Done()
				pages[idx], err = getMergedPRs(ctx, client, idx+1)
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

func main() {
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

	ctx := context.Background()
	var client *github.Client
	if *githubAccessToken == "" {
		client = github.NewClient(nil)
	} else {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *githubAccessToken},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}

	prs, err := GetMergedPRs(ctx, client, *numPRs)
	if err != nil {
		log.Fatalf("Failed to fetch PRs: %v", err)
	}
	sort.Sort(byMergedAt(prs))
	log.Infof("Fetched %d PRs", len(prs))
	epochs := GetEpochs()
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
