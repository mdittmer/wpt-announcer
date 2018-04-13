package announcer

import (
	"errors"
	"fmt"
	"time"

	"github.com/mdittmer/wpt-announcer/epoch"
	agit "github.com/mdittmer/wpt-announcer/git"
	log "github.com/sirupsen/logrus"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

var errNotAllEpochsConsumed = errors.New("Not all epochs consumed")
var errNilRepo = errors.New("GetRemoteAnnouncer.repo is nil")
var errVacuousEpochs = errors.New("[]epoch.Epoch slice is vacuous: contains no epochs")

// GetErrNotAllEpochsConsumed produces the canonical error for failing to consume all input epochs.
func GetErrNotAllEpochsConsumed() error {
	return errNotAllEpochsConsumed
}

// GetErrNilRepo produces the canonical error for a nil repo value that was expected to be non-nil.
func GetErrNilRepo() error {
	return errNilRepo
}

// GetErrVacuousEpochs the canonical error for a vacuous computation over epochs; i.e., passing an empty slice of epochs which would yield an empty output.
func GetErrVacuousEpochs() error {
	return errVacuousEpochs
}

// Revision constitutes a announcer data related to a git revision.
type Revision interface {
	// GetHash returns the SHA hash associated with this revision.
	GetHash() [20]byte

	// GetTime returns the UTC commit time associated with this revision.
	GetTime() time.Time

	// GetEpoch returns the epoch associated with this revision.
	GetEpoch() *epoch.Epoch

	// GetEpochBasis returns the epoch basis associated with this revision.
	GetEpochBasis() *epoch.Basis
}

// RevisionData is a basic struct implementation of Revision.
type RevisionData struct {
	Hash       [20]byte
	CommitTime time.Time
	Epoch      *epoch.Epoch
	Basis      *epoch.Basis
}

func (r RevisionData) GetHash() [20]byte {
	return r.Hash
}

func (r RevisionData) GetTime() time.Time {
	return r.CommitTime
}

func (r RevisionData) GetEpoch() *epoch.Epoch {
	return r.Epoch
}

func (r RevisionData) GetEpochBasis() *epoch.Basis {
	return r.Basis
}

// Announcer constitutes the top-level component for implementing a revisions-of-interest announcer.
type Announcer interface {
	// GetRevisions computes epochal revisions based on current local announcer state.
	GetRevisions([]*epoch.Epoch, *epoch.Basis) ([]Revision, error)

	// Update applies an incremental update to announcer state; e.g., an Announcer bound to a repository may have a local clone and perform an incremental fetch.
	Update() error

	// Reset abandons current announcer state and reloads a valid initial announcer state.
	Reset() error
}

// GitRemoteAnnouncerConfig configures the git operations performed by a GitRemoteAnnouncer.
type GitRemoteAnnouncerConfig struct {
	URL           string
	RemoteName    string
	ReferenceName plumbing.ReferenceName
	Depth         int

	agit.Git
}

type gitRemoteAnnouncer struct {
	repo agit.Repository
	cfg  *GitRemoteAnnouncerConfig
}

// NewGitRemoteAnnouncer produces an Announcer that is bound to an agit.Repository.
func NewGitRemoteAnnouncer(cfg GitRemoteAnnouncerConfig) (a Announcer, err error) {
	a = &gitRemoteAnnouncer{
		nil,
		&cfg,
	}
	err = a.Reset()
	return a, err
}

func (a *gitRemoteAnnouncer) GetRevisions(epochs []*epoch.Epoch, basis *epoch.Basis) (rs []Revision, err error) {
	if a.repo == nil {
		err = GetErrNilRepo()
	} else if len(epochs) == 0 {
		err = GetErrVacuousEpochs()
	}
	if err != nil {
		log.Error(err)
		return rs, err
	}

	iter, err := a.repo.Tags()
	if err != nil {
		log.Errorf("Error loading repository tags: %v", err)
		return rs, err
	}
	iter = agit.NewMergedPRIter(iter)

	var prev *object.Commit
	prev = nil
	for next, err := iter.Next(); err == nil && len(epochs) > 0; next, err = iter.Next() {
		epoch := epochs[0]
		nextCommit, err := a.repo.CommitObject(next.Hash())
		if err != nil {
			log.Warnf("Failed to locate commit for PR tag: %s", string(next.Name()))
			continue
		}

		if epoch.IsEpochal(&prev.Committer.When, &nextCommit.Committer.When, basis) {
			rs = append(rs, RevisionData{
				Hash:       nextCommit.Hash,
				CommitTime: nextCommit.Committer.When,
				Epoch:      epoch,
				Basis:      basis,
			})
			epochs = epochs[1:]
		}

		prev = nextCommit
	}

	if len(epochs) > 0 {
		err = GetErrNotAllEpochsConsumed()
		log.Warnf("Not all epochs consumed: %v", err)
		return rs, err
	}

	return rs, err
}

func (a *gitRemoteAnnouncer) Update() (err error) {
	if a.repo == nil {
		err = GetErrNilRepo()
		log.Error(err)
		return err
	}

	name := a.cfg.ReferenceName
	refSpec := config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/remotes/origin/%s", name, name))
	if err = a.repo.Fetch(&git.FetchOptions{
		RemoteName: a.cfg.RemoteName,
		RefSpecs:   []config.RefSpec{refSpec},
		Depth:      100,
	}); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (a *gitRemoteAnnouncer) Reset() error {
	cfg := a.cfg
	repo, err := cfg.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           cfg.URL,
		RemoteName:    cfg.RemoteName,
		ReferenceName: cfg.ReferenceName,
		Depth:         cfg.Depth,
	})
	if err != nil {
		log.Errorf("Error creating git clone: %v", err)
		return err
	}
	a.repo = repo
	return nil
}
