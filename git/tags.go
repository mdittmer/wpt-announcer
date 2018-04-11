package git

import (
	"flag"
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
	for ref, err = iter.iter.Next(); err != nil && iter.filter(ref); ref, err = iter.iter.Next() {
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
			return strings.HasPrefix(string(ref.Name()), "merged_pr_")
		},
		iter: iter,
	}
}

/*

type TagIterConfig struct {
	URL           string
	ReferenceName plumbing.ReferenceName
	gitChunkSize  *int
}

type tagIter struct {
	repo *git.Repository
	iter storer.ReferenceIter
	prev *plumbing.Reference
}

func (iter tagIter) Next() (*plumbing.Reference, error) {
	ref, err := iter.iter.Next()
	if err == io.EOF {
		return iter.getNext()
	}
	iter.prev = ref
	return ref, err
}

func (iter tagIter) ForEach(f func(*plumbing.Reference) error) error {
	// TODO(markdittmer): This is not right. Should complete; check error, then maybe fetch more and do more forEaching
	wrapper := func(r *plumbing.Reference) error {
		err := f(r)
		if err == io.EOF {
			return iter.continueForEach(f)
		}
		return err
	}
	return iter.iter.ForEach(wrapper)
}

func (iter tagIter) Close() {
	iter.iter.Close()
}

func NewTagIter(cfg TagIterConfig) (*tagIter, error) {
	var depth int
	if cfg.gitChunkSize != nil {
		depth = *cfg.gitChunkSize
	} else {
		depth = *gitChunkSize
	}
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           cfg.URL,
		RemoteName:    "origin",
		ReferenceName: cfg.ReferenceName,
		Depth:         depth,
	})
	if err != nil {
		log.Errorf("Error creating git clone: %v", err)
		return nil, err
	}
	baseIter, err := repo.Tags()
	if err != nil {
		log.Errorf("Error creating fetching tags from clone: %v", err)
		return nil, err
	}
	return &tagIter{
		repo: repo,
		iter: baseIter,
	}, nil
}
*/
