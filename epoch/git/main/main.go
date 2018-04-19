package main

import (
	"encoding/hex"
	"time"

	log "github.com/sirupsen/logrus"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

func doStuff(repo *git.Repository) {
	itr, err := repo.Tags()
	var ref *plumbing.Reference
	count := 0
	commitTime := time.Now()
	for ref, err = itr.Next(); ref != nil && err == nil; ref, err = itr.Next() {
		log.Infof("Reference: %v", ref)
		hash := ref.Hash()
		log.Infof("Reference data: %d %s %s %s", ref.Type(), string(ref.Name()), hex.EncodeToString(hash[0:]), string(ref.Target()))

		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			log.Errorf("Failed to get commit from ref: %v", err)
			continue
		}
		nextCommitTime := commit.Committer.When
		log.Infof("Commit time: %v", commit.Committer.When)
		if commitTime.Before(nextCommitTime) {
			log.Errorf("Next tag occurs after previous tag  :-( prev: %v next: %v", commitTime, nextCommitTime)
		} else {
			log.Infof("Next tag occurs before previous tag :-) prev: %v next: %v", commitTime, nextCommitTime)
		}
		commitTime = nextCommitTime
		count++
	}
	log.Infof("Reference count: %d", count)
	// log.Infof("Post-loop tag: %v", tag)
	// log.Infof("Post-loop error: %v", err)
}

func main() {
	cloneStart := time.Now()
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           "https://github.com/w3c/web-platform-tests.git",
		RemoteName:    "origin",
		ReferenceName: "refs/heads/master",
		Depth:         1,
		SingleBranch:  false,
		Tags:          git.TagFollowing,
	})
	if err != nil {
		log.Fatalf("Failed to clone: %v", err)
	}
	cloneEnd := time.Now()
	log.Infof("Finish complete clone %v", cloneEnd.Sub(cloneStart))
	doStuff(repo)

	/*
		pullStart := time.Now()
		if err = repo.Fetch(&git.FetchOptions{
			RemoteName: "origin",
			RefSpecs:   []config.RefSpec{"+refs/heads/master:refs/remotes/origin/master"},
			Depth:      100,
		}); err != nil {
			log.Errorf("Failed to pull: %v", err)
		}
		pullEnd := time.Now()
		log.Infof("Finish complete pull %v", pullEnd.Sub(pullStart))
		// doStuff()
	*/
}
