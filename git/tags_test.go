package git_test

import (
	"errors"
	"io"
	"testing"

	"encoding/hex"

	agit "github.com/mdittmer/wpt-announcer/git"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type MockReferenceIter struct {
	refs     []*plumbing.Reference
	idx      int
	isClosed bool
}

func (iter *MockReferenceIter) Next() (ref *plumbing.Reference, err error) {
	if iter.isClosed {
		return nil, errors.New("iter.Next() after iter.Close()")
	}
	if iter.idx >= len(iter.refs) {
		return nil, io.EOF
	}
	ref = iter.refs[iter.idx]
	iter.idx++
	return ref, err
}

func (iter *MockReferenceIter) ForEach(f func(*plumbing.Reference) error) error {
	if iter.isClosed {
		return errors.New("iter.ForEach() after iter.Close()")
	}
	if iter.idx >= len(iter.refs) {
		return io.EOF
	}

	refs := iter.refs[iter.idx:]
	for _, ref := range refs {
		err := f(ref)
		if err != nil {
			return err
		}
		iter.idx++
	}

	return nil
}

func (iter *MockReferenceIter) Close() {
	iter.isClosed = true
}

func NewMockIter(refs []*plumbing.Reference) MockReferenceIter {
	return MockReferenceIter{refs, 0, false}
}

func NewHashRef(str string) *plumbing.Reference {
	refName := plumbing.ReferenceName(str)
	hashSlice, err := hex.DecodeString(str)
	if err != nil {
		log.Fatalf("NewHashRef() expects hex string, but got %s", str)
	}
	if len(hashSlice) != 20 {
		log.Fatalf("NewHashRef() expects hex string constituting 20 bytes but got %d bytes from %s", len(hashSlice), str)
	}
	var fixedHash [20]byte
	for i := range fixedHash {
		fixedHash[i] = hashSlice[i]
	}
	return plumbing.NewHashReference(refName, plumbing.Hash(fixedHash))
}

func NewTagRef(name, target plumbing.ReferenceName) *plumbing.Reference {
	return plumbing.NewSymbolicReference("refs/tags/"+name, target)
}

func TestFilteredReferenceIter_Simple(t *testing.T) {
	prs := []*plumbing.Reference{
		NewTagRef("merged_pr_4", "0000000000000000000000000000000000000004"),
		NewTagRef("merged_pr_6", "0000000000000000000000000000000000000006"),
	}
	allRefs := []*plumbing.Reference{
		NewHashRef("0000000000000000000000000000000000000001"),
		NewHashRef("0000000000000000000000000000000000000002"),
		NewHashRef("0000000000000000000000000000000000000003"),
		NewTagRef("not_a_mergedpr_3", "0000000000000000000000000000000000000003"),
		NewHashRef("0000000000000000000000000000000000000004"),
		prs[0],
		NewHashRef("0000000000000000000000000000000000000005"),
		NewTagRef("not_a_mergedpr", "0000000000000000000000000000000000000003"),
		NewHashRef("0000000000000000000000000000000000000006"),
		prs[1],
	}

	baseIter := NewMockIter(allRefs)
	filteredIter := agit.NewMergedPRIter(&baseIter)
	i := 0
	for ref, err := filteredIter.Next(); err == nil; ref, err = filteredIter.Next() {
		assert.True(t, ref == prs[i])
		i++
	}
	assert.True(t, i == len(prs))

	baseIter = NewMockIter(allRefs)
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
	stopAt := NewHashRef("0000000000000000000000000000000000000004")
	includedRefs := []*plumbing.Reference{
		NewHashRef("0000000000000000000000000000000000000001"),
		NewHashRef("0000000000000000000000000000000000000002"),
		NewHashRef("0000000000000000000000000000000000000003"),
		NewTagRef("not_a_mergedpr_3", "0000000000000000000000000000000000000003"),
	}
	var allRefs []*plumbing.Reference
	allRefs = append(allRefs, includedRefs...)
	allRefs = append(allRefs, []*plumbing.Reference{
		stopAt,
		NewHashRef("0000000000000000000000000000000000000005"),
		NewTagRef("not_a_mergedpr", "0000000000000000000000000000000000000003"),
		NewHashRef("0000000000000000000000000000000000000006"),
		NewTagRef("merged_pr_6", "0000000000000000000000000000000000000006"),
	}...)

	baseIter := NewMockIter(allRefs)
	filteredIter := agit.NewStopReferenceIter(&baseIter, stopAt)
	i := 0
	for ref, err := filteredIter.Next(); err == nil; ref, err = filteredIter.Next() {
		assert.True(t, ref == includedRefs[i])
		i++
	}
	assert.True(t, i == len(includedRefs))

	baseIter = NewMockIter(allRefs)
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
	stopAt := NewHashRef("0000000000000000000000000000000000000005")
	includedPrs := []*plumbing.Reference{
		NewTagRef("merged_pr_4", "0000000000000000000000000000000000000004"),
	}
	prs := []*plumbing.Reference{
		includedPrs[0],
		NewTagRef("merged_pr_6", "0000000000000000000000000000000000000006"),
	}
	allRefs := []*plumbing.Reference{
		NewHashRef("0000000000000000000000000000000000000001"),
		NewHashRef("0000000000000000000000000000000000000002"),
		NewHashRef("0000000000000000000000000000000000000003"),
		NewTagRef("not_a_mergedpr_3", "0000000000000000000000000000000000000003"),
		NewHashRef("0000000000000000000000000000000000000004"),
		prs[0],
		stopAt,
		NewTagRef("not_a_mergedpr", "0000000000000000000000000000000000000003"),
		NewHashRef("0000000000000000000000000000000000000006"),
		prs[1],
	}

	baseIter := NewMockIter(allRefs)
	stopIter := agit.NewStopReferenceIter(&baseIter, stopAt)
	filteredIter := agit.NewMergedPRIter(stopIter)
	i := 0
	for ref, err := filteredIter.Next(); err == nil; ref, err = filteredIter.Next() {
		assert.True(t, ref == includedPrs[i])
		i++
	}
	assert.True(t, i == len(includedPrs))

	baseIter = NewMockIter(allRefs)
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
