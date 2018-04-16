package announcer_test

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/mdittmer/wpt-announcer/announcer"
	"github.com/mdittmer/wpt-announcer/epoch"
	agit "github.com/mdittmer/wpt-announcer/git"
	"github.com/mdittmer/wpt-announcer/test"
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

func (mr *MockRepository) CommitObject(h plumbing.Hash) (*object.Commit, error) {
	hashStr := hex.EncodeToString(h[0:])
	commit, ok := mr.commits[hashStr]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Unable to locate commit for hash %s", hashStr))
	}
	return commit, nil
}

func (mr *MockRepository) Tags() (storer.ReferenceIter, error) {
	iter := test.NewMockIter(mr.refs)
	return &iter, nil
}

func (mr *MockRepository) Fetch(o *git.FetchOptions) error {
	return mr.fetch(o)
}

func (mr *MockRepository) Clone(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (agit.Repository, error) {
	return mr, nil
}

func NewCommit(hashStr string, commitTime time.Time) *object.Commit {
	hashSlice, err := hex.DecodeString(hashStr)
	if err != nil {
		log.Fatalf("Failed to decode hash string %s", hashStr)
	}
	var fixedHash [20]byte
	for i := range fixedHash {
		fixedHash[i] = hashSlice[i]
	}
	return &object.Commit{
		Hash: fixedHash,
		Committer: object.Signature{
			When: commitTime,
		},
	}
}

func TestGitRemoteAnnouncer_GetRevisions_AllPrs(t *testing.T) {
	dailyCommit := NewCommit("0000000000000000000000000000000000000002", time.Date(2018, 4, 9, 0, 0, 0, 0, time.UTC))
	prs := []*plumbing.Reference{
		test.NewTagRef("merged_pr_4", "0000000000000000000000000000000000000004"),
		test.NewTagRef("merged_pr_6", "0000000000000000000000000000000000000006"),
	}
	allRefs := []*plumbing.Reference{
		test.NewTagRef("not_a_mergedpr_1", "0000000000000000000000000000000000000001"),
		test.NewTagRef("not_a_mergedpr_2", "0000000000000000000000000000000000000002"),
		test.NewTagRef("not_a_mergedpr_3", "0000000000000000000000000000000000000003"),
		prs[0],
		test.NewTagRef("not_a_mergedpr_5", "0000000000000000000000000000000000000005"),
		prs[1],
	}
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		Git: &MockRepository{
			refs: allRefs,
			commits: map[string]*object.Commit{
				"0000000000000000000000000000000000000001": NewCommit("0000000000000000000000000000000000000001", time.Date(2018, 4, 10, 0, 0, 0, 0, time.UTC)),
				"0000000000000000000000000000000000000002": dailyCommit,
				"0000000000000000000000000000000000000003": NewCommit("0000000000000000000000000000000000000003", time.Date(2018, 4, 8, 0, 0, 0, 0, time.UTC)),
				"0000000000000000000000000000000000000004": NewCommit("0000000000000000000000000000000000000004", time.Date(2018, 4, 7, 0, 0, 0, 0, time.UTC)),
				"0000000000000000000000000000000000000005": NewCommit("0000000000000000000000000000000000000005", time.Date(2018, 4, 6, 0, 0, 0, 0, time.UTC)),
				"0000000000000000000000000000000000000006": NewCommit("0000000000000000000000000000000000000006", time.Date(2018, 4, 5, 0, 0, 0, 0, time.UTC)),
			},
			fetch: func(o *git.FetchOptions) error {
				return nil
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to instantiate announcer: %v", err)
	}
	daily := epoch.GetDaily()
	basis := &epoch.Basis{}
	rs, err := a.GetRevisions([]*epoch.Epoch{daily}, basis)
	if err != nil {
		log.Fatalf("Unexpected GetRevisions() error: %v", err)
	}
	assert.True(t, len(rs) == 1)
	assert.True(t, rs[0].GetEpoch() == daily)
	assert.True(t, rs[0].GetEpochBasis() == basis)
	assert.True(t, rs[0].GetHash() == dailyCommit.Hash)
	assert.True(t, rs[0].GetTime() == dailyCommit.Committer.When)
}
