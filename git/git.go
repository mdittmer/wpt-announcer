package git

import (
	billy "gopkg.in/src-d/go-billy.v4"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"gopkg.in/src-d/go-git.v4/storage"
)

// Repository is a handfule of git.Repository functions reified as an interface to facilitate testing.
type Repository interface {
	CommitObject(h plumbing.Hash) (*object.Commit, error)
	Tags() (storer.ReferenceIter, error)
	Fetch(o *git.FetchOptions) error
}

// Git is a handfule of git functions reified as an interface to facilitate testing.
type Git interface {
	Clone(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (Repository, error)
}
