package git

import (
	"flag"
	"io"
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

var gitChunkSize *int

func init() {
	flag.Int("git_size_chunk", 100, "Number of Git objects to fetch in chunks to extend depth-limited fetches")
}

type FilteredReferenceIter struct {
	filter func(*plumbing.Reference) bool
	iter   storer.ReferenceIter
}

func (iter FilteredReferenceIter) Next() (ref *plumbing.Reference, err error) {
	for ref, err = iter.iter.Next(); err == nil && !iter.filter(ref); ref, err = iter.iter.Next() {
	}
	return ref, err
}

func (iter FilteredReferenceIter) ForEach(f func(*plumbing.Reference) error) error {
	return iter.iter.ForEach(func(ref *plumbing.Reference) error {
		if iter.filter(ref) {
			return f(ref)
		}
		return nil
	})
}

func (iter FilteredReferenceIter) Close() {
	iter.iter.Close()
}

func NewMergedPRIter(iter storer.ReferenceIter) storer.ReferenceIter {
	return FilteredReferenceIter{
		filter: func(ref *plumbing.Reference) bool {
			if ref == nil {
				return false
			}
			return strings.HasPrefix(string(ref.Name()), "refs/tags/merged_pr_")
		},
		iter: iter,
	}
}

type StopReferenceIter struct {
	stopAt *plumbing.Reference
	iter   storer.ReferenceIter
}

func (iter StopReferenceIter) Next() (ref *plumbing.Reference, err error) {
	ref, err = iter.iter.Next()
	if err != nil {
		return ref, err
	}
	if iter.stopAt.Hash() == ref.Hash() {
		iter.Close()
		return nil, io.EOF
	}
	return ref, err
}

func (iter StopReferenceIter) ForEach(f func(*plumbing.Reference) error) error {
	return iter.iter.ForEach(func(ref *plumbing.Reference) error {
		if iter.stopAt.Hash() == ref.Hash() {
			return io.EOF
		}
		return f(ref)
	})
}

func (iter StopReferenceIter) Close() {
	iter.iter.Close()
}

func NewStopReferenceIter(iter storer.ReferenceIter, stopAt *plumbing.Reference) storer.ReferenceIter {
	return StopReferenceIter{
		stopAt,
		iter,
	}
}
