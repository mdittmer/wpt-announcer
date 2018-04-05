package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/go-github/github"
	"github.com/mdittmer/wpt-announcer/auth"
	gh "github.com/mdittmer/wpt-announcer/github"
	log "github.com/sirupsen/logrus"
	models "github.com/w3c/wptdashboard/shared"
)

type testRun struct {
	commit *github.Commit

	models.TestRun
}

func addCommits(ctx context.Context, client *github.Client, runs []models.TestRun) (testRuns []testRun, err error) {
	hashesMap := make(map[string]*github.Commit)
	for _, run := range runs {
		hashesMap[run.Revision] = nil
	}
	hashes := make([]string, 0, len(hashesMap))
	for hash := range hashesMap {
		hashes = append(hashes, hash)
	}
	commits, err := gh.GetCommits(ctx, client, hashes)
	if err != nil {
		return testRuns, err
	}
	for i := range commits {
		hashesMap[hashes[i]] = commits[i]
	}
	for _, run := range runs {
		testRuns = append(testRuns, testRun{
			commit:  hashesMap[run.Revision],
			TestRun: run,
		})
	}
	return testRuns, err
}

func main() {
	ctx := context.Background()
	resp, err := http.Get(fmt.Sprintf("https://wpt.fyi/api/runs?max-count=%d", int(math.Floor(500/4))))
	if err != nil {
		log.Fatalf("Failed to load runs: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("Unexpected response status code loading runs: %d", resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	var runs []models.TestRun
	if err := decoder.Decode(&runs); err != nil {
		log.Fatalf("Failed to decode runs JSON")
	}
	log.Infof("Fetched %d runs", len(runs))

	client := auth.GetGithubClient(ctx)
	commits, err := addCommits(ctx, client, runs)
	if err != nil {
		log.Fatalf("Failed to load commit data for runs: %v", err)
	}

	times := make(map[string][]time.Time)
	deltas := make(map[string][]time.Duration)
	for _, commit := range commits {
		key := commit.TestRun.BrowserName
		_, ok := times[key]
		if !ok {
			times[key] = make([]time.Time, 0)
			deltas[key] = make([]time.Duration, 0)
		}
		next := commit.CreatedAt
		if ok {
			prev := times[key][len(times[key])-1]
			deltas[key] = append(deltas[key], prev.Sub(next))
		}
		times[key] = append(times[key], next)
	}
	bytes, err := json.Marshal(&times)
	log.Infof("Commit times: %s", string(bytes))
	bytes, err = json.Marshal(&deltas)
	log.Infof("Time deltas: %s", string(bytes))

	outputFilePath := "./deltas.csv"
	log.Infof("Writing deltas to %s", outputFilePath)
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()
	for browserName, deltaSlice := range deltas {
		slice := make([]string, len(deltaSlice)+1, len(deltaSlice)+1)
		slice[0] = browserName
		for i, delta := range deltaSlice {
			slice[i+1] = strconv.FormatInt(int64(delta), 10)
		}
		writer.Write(slice)
	}
}
