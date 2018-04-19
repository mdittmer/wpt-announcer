package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mdittmer/wpt-announcer/announcer"
	"github.com/mdittmer/wpt-announcer/api"

	agit "github.com/mdittmer/wpt-announcer/git"
	log "github.com/sirupsen/logrus"
)

var a announcer.Announcer

func init() {
	go func() {
		log.Info("Initializing announcer")
		var err error
		a, err = announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
			URL:           "https://github.com/w3c/web-platform-tests.git",
			RemoteName:    "origin",
			ReferenceName: "refs/heads/master",
			Git:           agit.GoGit{},
		})
		if err != nil {
			log.Fatalf("Announcer initialization failed: %v", err)
		}
		log.Info("Announcer initialized")
	}()
}

func epochsHandler(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.Marshal(apiEpochs)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Failed to marshal epochs JSON"))
		return
	}
	w.Write(bytes)
}

func latestHandler(w http.ResponseWriter, r *http.Request) {
	if a == nil {
		w.WriteHeader(503)
		w.Write([]byte("Announcer not yet initialized"))
		return
	}

	if len(epochs) == 0 {
		w.WriteHeader(500)
		w.Write([]byte("No epochs"))
		return
	}

	now := time.Now()
	revs, err := a.GetRevisions(latestGetRevisions, announcer.Limits{
		Now:   now,
		Start: now.Add(-2 * epochs[0].GetData().MaxDuration),
	})
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	response, err := api.LatestFromEpochs(revs)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Failed to marshal latest epochal revisions JSON"))
		return
	}

	w.Write(bytes)
}

func main() {
	http.HandleFunc("/api/revisions/epochs", epochsHandler)
	http.HandleFunc("/api/revisions/latest", latestHandler)
}
