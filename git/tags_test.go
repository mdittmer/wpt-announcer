package git_test

import (
	"testing"

	agit "github.com/mdittmer/wpt-announcer/git"
	"github.com/mdittmer/wpt-announcer/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func TestFilteredReferenceIter_Simple(t *testing.T) {
	prs := []*plumbing.Reference{
		test.NewTagRef("merged_pr_4", "0000000000000000000000000000000000000004"),
		test.NewTagRef("merged_pr_6", "0000000000000000000000000000000000000006"),
	}
	allRefs := []*plumbing.Reference{
		test.NewTagRef("not_a_mergedpr_1", "0000000000000000000000000000000000000001"),
		test.NewTagRef("not_a_mergedpr_2", "0000000000000000000000000000000000000002"),
		test.NewTagRef("not_a_mergedpr_3", "0000000000000000000000000000000000000003"),
		prs[0],
		test.NewTagRef("not_a_mergedpr_4", "0000000000000000000000000000000000000005"),
		prs[1],
	}

	baseIter := test.NewMockIter(allRefs)
	filteredIter := agit.NewMergedPRIter(&baseIter)
	i := 0
	for ref, err := filteredIter.Next(); err == nil; ref, err = filteredIter.Next() {
		assert.True(t, ref == prs[i])
		i++
	}
	assert.True(t, i == len(prs))

	baseIter = test.NewMockIter(allRefs)
	filteredIter = agit.NewMergedPRIter(&baseIter)
	i = 0
	filteredIter.ForEach(func(ref *plumbing.Reference) error {
		assert.True(t, ref == prs[i])
		i++
		return nil
	})
	assert.True(t, i == len(prs))
}

func TestStopReferenceIter_Simple(t *testing.T) {
	stopAt := test.NewTagRef("some_tag_4", "0000000000000000000000000000000000000004")
	includedRefs := []*plumbing.Reference{
		test.NewTagRef("some_tag_1", "0000000000000000000000000000000000000001"),
		test.NewTagRef("some_tag_2", "0000000000000000000000000000000000000002"),
		test.NewTagRef("some_tag_3", "0000000000000000000000000000000000000003"),
	}
	var allRefs []*plumbing.Reference
	allRefs = append(allRefs, includedRefs...)
	allRefs = append(allRefs, []*plumbing.Reference{
		stopAt,
		test.NewTagRef("some_tag_5", "0000000000000000000000000000000000000005"),
		test.NewTagRef("some_tag_6", "0000000000000000000000000000000000000006"),
	}...)

	baseIter := test.NewMockIter(allRefs)
	filteredIter := agit.NewStopReferenceIter(&baseIter, stopAt)
	i := 0
	for ref, err := filteredIter.Next(); err == nil; ref, err = filteredIter.Next() {
		assert.True(t, ref == includedRefs[i])
		i++
	}
	assert.True(t, i == len(includedRefs))

	baseIter = test.NewMockIter(allRefs)
	filteredIter = agit.NewStopReferenceIter(&baseIter, stopAt)
	i = 0
	filteredIter.ForEach(func(ref *plumbing.Reference) error {
		assert.True(t, ref == includedRefs[i])
		i++
		return nil
	})
	assert.True(t, i == len(includedRefs))
}

func TestFilteredReferenceIter_StopReferenceIter_Compose(t *testing.T) {
	stopAt := test.NewTagRef("some_tag_5", "0000000000000000000000000000000000000005")
	includedPrs := []*plumbing.Reference{
		test.NewTagRef("merged_pr_4", "0000000000000000000000000000000000000004"),
	}
	prs := []*plumbing.Reference{
		includedPrs[0],
		test.NewTagRef("merged_pr_6", "0000000000000000000000000000000000000006"),
	}
	allRefs := []*plumbing.Reference{
		test.NewTagRef("not_a_mergedpr_1", "0000000000000000000000000000000000000001"),
		test.NewTagRef("not_a_mergedpr_2", "0000000000000000000000000000000000000002"),
		test.NewTagRef("not_a_mergedpr_3", "0000000000000000000000000000000000000003"),
		prs[0],
		stopAt,
		prs[1],
	}

	baseIter := test.NewMockIter(allRefs)
	stopIter := agit.NewStopReferenceIter(&baseIter, stopAt)
	filteredIter := agit.NewMergedPRIter(stopIter)
	i := 0
	for ref, err := filteredIter.Next(); err == nil; ref, err = filteredIter.Next() {
		assert.True(t, ref == includedPrs[i])
		i++
	}
	assert.True(t, i == len(includedPrs))

	baseIter = test.NewMockIter(allRefs)
	stopIter = agit.NewStopReferenceIter(&baseIter, stopAt)
	filteredIter = agit.NewMergedPRIter(stopIter)
	i = 0
	filteredIter.ForEach(func(ref *plumbing.Reference) error {
		assert.True(t, ref == includedPrs[i])
		i++
		return nil
	})
	assert.True(t, i == len(includedPrs))
}
