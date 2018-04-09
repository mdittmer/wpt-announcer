package git

import (
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

func Bootstrap(opts *git.CloneOptions) (itr storer.ReferenceIter, err error) {
	repo, err := git.Clone(memory.NewStorage(), nil, opts)
	if err != nil {
		return itr, err
	}
	itr, err = repo.Tags()
	return itr, err
}
