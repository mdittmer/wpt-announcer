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
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

var errNotAllEpochsConsumed = errors.New("Not all epochs consumed")
var errNilRepo = errors.New("GetRemoteAnnouncer.repo is nil")

func GetErrNotAllEpochsConsumed() error {
	return errNotAllEpochsConsumed
}

func GetErrNilRepo() error {
	return errNilRepo
}

type Revision interface {
	GetHash() [20]byte
	GetTime() time.Time
	GetEpoch() *epoch.Epoch
	GetEpochBasis() *epoch.Basis
}

type RevisionData struct {
	hash       [20]byte
	commitTime time.Time
	epoch      *epoch.Epoch
	basis      *epoch.Basis
}

func (r RevisionData) GetHash() [20]byte {
	return r.hash
}

func (r RevisionData) GetTime() time.Time {
	return r.commitTime
}

func (r RevisionData) GetEpoch() *epoch.Epoch {
	return r.epoch
}

func (r RevisionData) GetEpochBasis() *epoch.Basis {
	return r.basis
}

type Announcer interface {
	GetInitialRevisions([]epoch.Epoch, epoch.Basis) ([]Revision, error)
	GetUpdatedRevisions([]Revision) ([]Revision, error)

	getInitialPRs() (storer.ReferenceIter, error)
	getNewPRs() (storer.ReferenceIter, error)
	cloneRepo() error
	updateRepo() error
	resetRepo() error
}

type GitRemoteAnnouncer struct {
	repo *git.Repository
	cfg  *GitRemoteAnnouncerConfig
}

type GitRemoteAnnouncerConfig struct {
	URL           string
	RemoteName    string
	ReferenceName plumbing.ReferenceName
	Depth         int
}

func (a GitRemoteAnnouncer) cloneRepo() error {
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           a.cfg.URL,
		RemoteName:    a.cfg.RemoteName,
		ReferenceName: a.cfg.ReferenceName,
		Depth:         a.cfg.Depth,
	})
	if err != nil {
		log.Errorf("Error creating git clone: %v", err)
		return err
	}
	a.repo = repo
	return nil
}

func NewGitRemoteAnnouncer(cfg GitRemoteAnnouncerConfig) (a *GitRemoteAnnouncer, err error) {
	a = &GitRemoteAnnouncer{}
	if err = a.cloneRepo(); err != nil {
		return nil, err
	}
	return a, err
}

func (a GitRemoteAnnouncer) getRevisions(epochs []*epoch.Epoch, basis *epoch.Basis, iter storer.ReferenceIter) (rs []Revision, err error) {
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
				hash:       nextCommit.Hash,
				commitTime: nextCommit.Committer.When,
				epoch:      epoch,
				basis:      basis,
			})
			epochs = epochs[1:]
		}

		prev = nextCommit
	}
	if len(epochs) > 0 {
		return rs, GetErrNotAllEpochsConsumed()
	}

	return rs, err
}

func (a GitRemoteAnnouncer) GetInitialRevisions(epochs []*epoch.Epoch, basis *epoch.Basis) (rs []Revision, err error) {
	if a.repo == nil {
		err = GetErrNilRepo()
		log.Errorf("GitRemoteAnnouncer.GetInitialRevisions: %v", err)
		return rs, err
	}

	iter, err := a.repo.Tags()
	if err != nil {
		log.Errorf("Error loading repository tags: %v", err)
		return rs, err
	}
	iter = agit.NewMergedPRIter(iter)

	return a.getRevisions(epochs, basis, iter)
}

func (a GitRemoteAnnouncer) GetUpdatedRevisions(epochs []*epoch.Epoch, basis *epoch.Basis) (rs []Revision, err error) {
	if a.repo == nil {
		err = GetErrNilRepo()
		log.Errorf("GitRemoteAnnouncer.GetUpdatedRevisions: %v", err)
		return rs, err
	}

	name := a.cfg.ReferenceName
	refSpec := config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/remotes/origin/%s", name, name))
	if err = a.repo.Fetch(&git.FetchOptions{
		RemoteName: a.cfg.RemoteName,
		RefSpecs:   []config.RefSpec{refSpec},
		Depth:      100,
	}); err != nil {
		log.Errorf("GitRemoteAnnouncer.GetUpdatedRevisions: %v", err)
		return rs, err
	}

	return a.GetInitialRevisions(epochs, basis)
}
