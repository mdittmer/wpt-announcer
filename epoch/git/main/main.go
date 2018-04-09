package main

import (
	"time"

	log "github.com/sirupsen/logrus"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

/*
func doStuff(repo *git.Repository) {
	itr, err := repo.Tags()
	var tag *plumbing.Reference
	for tag, err = itr.Next(); tag != nil && err == nil; tag, err = itr.Next() {
		// log.Infof("Tag: %v", tag)
		commit, err := repo.CommitObject(tag.Hash())
		if err != nil {
			log.Errorf("Failed to get commit from tag: %v", err)
			continue
		}
		// log.Infof("Commit time: %v", commit.Committer.When)
	}
	// log.Infof("Post-loop tag: %v", tag)
	// log.Infof("Post-loop error: %v", err)
}
*/

func main() {
	cloneStart := time.Now()
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           "https://github.com/w3c/web-platform-tests.git",
		RemoteName:    "origin",
		ReferenceName: "refs/tags/merge_pr_8470",
		Depth:         100,
	})
	if err != nil {
		log.Fatalf("Failed to clone: %v", err)
	}
	cloneEnd := time.Now()
	log.Infof("Finish complete clone %v", cloneEnd.Sub(cloneStart))
	// doStuff(repo)

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
}
