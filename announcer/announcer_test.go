package announcer_test

import (
	"errors"
	"testing"

	"github.com/mdittmer/wpt-announcer/announcer"
	"github.com/mdittmer/wpt-announcer/epoch"
	agit "github.com/mdittmer/wpt-announcer/git"
	"github.com/stretchr/testify/assert"
	billy "gopkg.in/src-d/go-billy.v4"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"gopkg.in/src-d/go-git.v4/storage"
)

type Fake struct{}

func (iter Fake) Next() (ref *plumbing.Reference, err error) {
	return nil, nil
}

func (iter Fake) ForEach(f func(*plumbing.Reference) error) error {
	return nil
}

func (iter Fake) Close() {}

func (Fake) CommitObject(h plumbing.Hash) (*object.Commit, error) {
	return nil, nil
}

func (Fake) Tags() (storer.ReferenceIter, error) {
	return Fake{}, nil
}

func (Fake) Fetch(o *git.FetchOptions) error {
	return nil
}

func (Fake) Clone(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (agit.Repository, error) {
	return nil, nil
}

type CloneErrorProducer struct{}

var cloneError = errors.New("Clone error")

func (cep CloneErrorProducer) Clone(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (agit.Repository, error) {
	return nil, cloneError
}

func TestGitRemoteAnnouncer_Init(t *testing.T) {
	_, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		Git: Fake{},
	})
	assert.True(t, err == nil)
}

func TestGitRemoteAnnouncer_CloneError(t *testing.T) {
	_, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		Git: CloneErrorProducer{},
	})
	assert.True(t, err == cloneError)
}

func TestGitRemoteAnnouncer_GetRevisions_NilRepo(t *testing.T) {
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		Git: Fake{},
	})
	assert.True(t, err == nil)
	_, err = a.GetRevisions([]*epoch.Epoch{}, nil)
	assert.True(t, err == announcer.GetErrNilRepo())
}

type EmptyRepoProducer struct{}

func (erp EmptyRepoProducer) Clone(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (agit.Repository, error) {
	return git.Init(s, worktree)
}

func TestGitRemoteAnnouncer_GetRevisions_VacuousEpochs(t *testing.T) {
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		Git: EmptyRepoProducer{},
	})
	assert.True(t, err == nil)
	_, err = a.GetRevisions([]*epoch.Epoch{}, nil)
	assert.True(t, err == announcer.GetErrVacuousEpochs())
}

type MockRepository struct {
	refs    []*plumbing.Reference
	commits map[string]*object.Commit
	fetch   func(o *git.FetchOptions) error
}

// TODO(markdittmer): Test meaningful GetRevisions() calls, and all behaviours
// for Update().
