package git_test

import (
	"io"
	"log"
	"testing"
	"time"

	agit "github.com/mdittmer/wpt-announcer/git"
	"github.com/mdittmer/wpt-announcer/test"
	"github.com/stretchr/testify/assert"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

func nilFetchImpl(mr *test.MockRepository, o *git.FetchOptions) error {
	return nil
}

func TestTimeOrderedReferenceIter_Simple(t *testing.T) {
	refs := []*plumbing.Reference{
		test.NewTagRef("not_a_mergedpr_1", "0000000000000000000000000000000000000001"),
		test.NewTagRef("not_a_mergedpr_2", "0000000000000000000000000000000000000002"),
		test.NewTagRef("not_a_mergedpr_3", "0000000000000000000000000000000000000003"),
		test.NewTagRef("merged_pr_4", "0000000000000000000000000000000000000004"),
		test.NewTagRef("not_a_mergedpr_5", "0000000000000000000000000000000000000005"),
		test.NewTagRef("merged_pr_6", "0000000000000000000000000000000000000006"),
	}
	commits := map[string]*object.Commit{
		"0000000000000000000000000000000000000001": test.NewCommit("0000000000000000000000000000000000000001", time.Date(2018, 4, 3, 0, 0, 0, 0, time.UTC)),
		"0000000000000000000000000000000000000002": test.NewCommit("0000000000000000000000000000000000000002", time.Date(2018, 4, 2, 0, 0, 0, 0, time.UTC)),
		"0000000000000000000000000000000000000003": test.NewCommit("0000000000000000000000000000000000000003", time.Date(2018, 4, 6, 0, 0, 0, 0, time.UTC)),
		"0000000000000000000000000000000000000004": test.NewCommit("0000000000000000000000000000000000000004", time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)),
		"0000000000000000000000000000000000000005": test.NewCommit("0000000000000000000000000000000000000005", time.Date(2018, 4, 5, 0, 0, 0, 0, time.UTC)),
		"0000000000000000000000000000000000000006": test.NewCommit("0000000000000000000000000000000000000006", time.Date(2018, 4, 4, 0, 0, 0, 0, time.UTC)),
	}
	iterFactory := func(t *testing.T) storer.ReferenceIter {
		baseIter := test.NewMockIter(refs)
		iter, err := agit.NewTimeOrderedReferenceIter(&baseIter, test.NewMockRepository(refs, commits, nilFetchImpl))
		assert.True(t, err == nil)
		return iter
	}

	sortedIdxs := []int{2, 4, 5, 0, 1, 3}
	var ref *plumbing.Reference
	var err error

	iter := iterFactory(t)
	i := 0
	for ref, err = iter.Next(); ref != nil && err == nil; ref, err = iter.Next() {
		assert.True(t, refs[sortedIdxs[i]] == ref)
		i++
	}
	assert.True(t, err == io.EOF)

	iter = iterFactory(t)
	i = 0
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		assert.True(t, i < len(refs))
		assert.True(t, refs[sortedIdxs[i]] == ref)
		i++
		return nil
	})
	assert.True(t, err == nil)
}

func TestFilteredReferenceIter_Custom(t *testing.T) {
	refs := []*plumbing.Reference{
		test.NewTagRef("tag_1", "0000000000000000000000000000000000000001"),
		test.NewTagRef("tag_2", "0000000000000000000000000000000000000002"),
		test.NewTagRef("tag_3", "0000000000000000000000000000000000000003"),
	}

	// Include every other commit.
	newFilter := func() agit.ReferencePredicate {
		include := false
		return func(ref *plumbing.Reference) bool {
			include = !include
			return include
		}
	}

	iter := test.NewMockIter(refs)
	stopIter := agit.NewFilteredReferenceIter(&iter, newFilter())
	firstRef, err := stopIter.Next()
	assert.True(t, err == nil)
	assert.True(t, firstRef == refs[0])
	secondRef, err := stopIter.Next()
	assert.True(t, err == nil)
	assert.True(t, secondRef == refs[2])
	_, err = stopIter.Next()
	assert.True(t, err == io.EOF)

	iter = test.NewMockIter(refs)
	stopIter = agit.NewFilteredReferenceIter(&iter, newFilter())
	i := 0
	err = stopIter.ForEach(func(ref *plumbing.Reference) error {
		assert.True(t, ref == refs[i])
		i += 2
		return nil
	})
	assert.True(t, err == nil)
}

// TODO(markdittmer): Rename TestMergedPRIter TestMergedPRIter.
// Test custom filtered reference iter.

func TestMergedPRIter_Simple(t *testing.T) {
	prs := []*plumbing.Reference{
		test.NewTagRef("merged_pr_6", "0000000000000000000000000000000000000006"),
		test.NewTagRef("merged_pr_4", "0000000000000000000000000000000000000004"),
	}
	allRefs := []*plumbing.Reference{
		test.NewTagRef("not_a_mergedpr_1", "0000000000000000000000000000000000000001"),
		test.NewTagRef("not_a_mergedpr_2", "0000000000000000000000000000000000000002"),
		test.NewTagRef("not_a_mergedpr_3", "0000000000000000000000000000000000000003"),
		prs[1],
		test.NewTagRef("not_a_mergedpr_5", "0000000000000000000000000000000000000005"),
		prs[0],
	}
	commits := map[string]*object.Commit{
		"0000000000000000000000000000000000000001": &object.Commit{
			Hash: allRefs[0].Hash(),
			Committer: object.Signature{
				When: time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		"0000000000000000000000000000000000000002": &object.Commit{
			Hash: allRefs[1].Hash(),
			Committer: object.Signature{
				When: time.Date(2018, 4, 2, 0, 0, 0, 0, time.UTC),
			},
		},
		"0000000000000000000000000000000000000003": &object.Commit{
			Hash: allRefs[2].Hash(),
			Committer: object.Signature{
				When: time.Date(2018, 4, 3, 0, 0, 0, 0, time.UTC),
			},
		},
		"0000000000000000000000000000000000000004": &object.Commit{
			Hash: allRefs[3].Hash(),
			Committer: object.Signature{
				When: time.Date(2018, 4, 4, 0, 0, 0, 0, time.UTC),
			},
		},
		"0000000000000000000000000000000000000005": &object.Commit{
			Hash: allRefs[4].Hash(),
			Committer: object.Signature{
				When: time.Date(2018, 4, 5, 0, 0, 0, 0, time.UTC),
			},
		},
		"0000000000000000000000000000000000000006": &object.Commit{
			Hash: allRefs[5].Hash(),
			Committer: object.Signature{
				When: time.Date(2018, 4, 6, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	repo := test.NewMockRepository(allRefs, commits, nilFetchImpl)
	baseIter := test.NewMockIter(allRefs)
	filteredIter, err := agit.NewMergedPRIter(&baseIter, repo)
	assert.True(t, err == nil)
	i := 0
	for ref, err := filteredIter.Next(); err == nil; ref, err = filteredIter.Next() {
		assert.True(t, ref == prs[i])
		i++
	}
	assert.True(t, i == len(prs))

	baseIter = test.NewMockIter(allRefs)
	filteredIter, err = agit.NewMergedPRIter(&baseIter, repo)
	assert.True(t, err == nil)
	i = 0
	filteredIter.ForEach(func(ref *plumbing.Reference) error {
		assert.True(t, ref == prs[i])
		i++
		return nil
	})
	assert.True(t, i == len(prs))
}

func stopAtHash(h plumbing.Hash) agit.ReferencePredicate {
	return func(ref *plumbing.Reference) bool {
		if ref == nil {
			log.Fatal("Unexpected nil reference in test stopAtHash function")
		}
		return ref.Hash() == h
	}
}

func TestStopReferenceIter_Simple(t *testing.T) {
	stopAt := test.NewHash("0000000000000000000000000000000000000004")
	includedRefs := []*plumbing.Reference{
		test.NewTagRef("some_tag_1", "0000000000000000000000000000000000000001"),
		test.NewTagRef("some_tag_2", "0000000000000000000000000000000000000002"),
		test.NewTagRef("some_tag_3", "0000000000000000000000000000000000000003"),
	}
	var allRefs []*plumbing.Reference
	allRefs = append(allRefs, includedRefs...)
	allRefs = append(allRefs, []*plumbing.Reference{
		test.NewTagRef("some_tag_4", "0000000000000000000000000000000000000004"),
		test.NewTagRef("some_tag_5", "0000000000000000000000000000000000000005"),
		test.NewTagRef("some_tag_6", "0000000000000000000000000000000000000006"),
	}...)

	baseIter := test.NewMockIter(allRefs)
	filteredIter := agit.NewStopReferenceIter(&baseIter, stopAtHash(stopAt))
	i := 0
	for ref, err := filteredIter.Next(); err == nil; ref, err = filteredIter.Next() {
		assert.True(t, ref == includedRefs[i])
		i++
	}
	assert.True(t, i == len(includedRefs))

	baseIter = test.NewMockIter(allRefs)
	filteredIter = agit.NewStopReferenceIter(&baseIter, stopAtHash(stopAt))
	i = 0
	filteredIter.ForEach(func(ref *plumbing.Reference) error {
		assert.True(t, ref == includedRefs[i])
		i++
		return nil
	})
	assert.True(t, i == len(includedRefs))
}

func TestMergedPRIter_StopReferenceIter_Compose(t *testing.T) {
	stopAt := test.NewHash("0000000000000000000000000000000000000006")
	includedPrs := []*plumbing.Reference{
		test.NewTagRef("merged_pr_7", "0000000000000000000000000000000000000007"),
	}
	prs := []*plumbing.Reference{
		includedPrs[0],
		test.NewTagRef("merged_pr_6", "0000000000000000000000000000000000000006"),
		test.NewTagRef("merged_pr_4", "0000000000000000000000000000000000000004"),
	}
	// Time order matches 1..6 hash values, but listed out of order to test composition with TimeOrderedReferenceIter.
	allRefs := []*plumbing.Reference{
		prs[0],
		test.NewTagRef("not_a_mergedpr_1", "0000000000000000000000000000000000000001"),
		prs[2],
		test.NewTagRef("not_a_mergedpr_3", "0000000000000000000000000000000000000003"),
		prs[1],
		test.NewTagRef("not_a_mergedpr_2", "0000000000000000000000000000000000000002"),
		test.NewTagRef("not_a_mergedpr_5", "0000000000000000000000000000000000000005"),
	}
	commits := map[string]*object.Commit{
		"0000000000000000000000000000000000000001": &object.Commit{
			Hash: allRefs[0].Hash(),
			Committer: object.Signature{
				When: time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		"0000000000000000000000000000000000000002": &object.Commit{
			Hash: allRefs[1].Hash(),
			Committer: object.Signature{
				When: time.Date(2018, 4, 2, 0, 0, 0, 0, time.UTC),
			},
		},
		"0000000000000000000000000000000000000003": &object.Commit{
			Hash: allRefs[2].Hash(),
			Committer: object.Signature{
				When: time.Date(2018, 4, 3, 0, 0, 0, 0, time.UTC),
			},
		},
		"0000000000000000000000000000000000000004": &object.Commit{
			Hash: allRefs[3].Hash(),
			Committer: object.Signature{
				When: time.Date(2018, 4, 4, 0, 0, 0, 0, time.UTC),
			},
		},
		"0000000000000000000000000000000000000005": &object.Commit{
			Hash: allRefs[4].Hash(),
			Committer: object.Signature{
				When: time.Date(2018, 4, 5, 0, 0, 0, 0, time.UTC),
			},
		},
		"0000000000000000000000000000000000000006": &object.Commit{
			Hash: allRefs[5].Hash(),
			Committer: object.Signature{
				When: time.Date(2018, 4, 6, 0, 0, 0, 0, time.UTC),
			},
		},
		"0000000000000000000000000000000000000007": &object.Commit{
			Hash: allRefs[6].Hash(),
			Committer: object.Signature{
				When: time.Date(2018, 4, 7, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	repo := test.NewMockRepository(allRefs, commits, nilFetchImpl)
	baseIter := test.NewMockIter(allRefs)
	filteredIter, err := agit.NewMergedPRIter(&baseIter, repo)
	assert.True(t, err == nil)
	iter := agit.NewStopReferenceIter(filteredIter, stopAtHash(stopAt))
	i := 0
	for ref, err := iter.Next(); err == nil; ref, err = iter.Next() {
		assert.True(t, ref == includedPrs[i])
		i++
	}
	assert.True(t, i == len(includedPrs))

	baseIter = test.NewMockIter(allRefs)
	filteredIter, err = agit.NewMergedPRIter(&baseIter, repo)
	assert.True(t, err == nil)
	iter = agit.NewStopReferenceIter(filteredIter, stopAtHash(stopAt))
	assert.True(t, err == nil)
	i = 0
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		assert.True(t, ref == includedPrs[i])
		i++
		return nil
	})
	assert.True(t, i == len(includedPrs))
}
