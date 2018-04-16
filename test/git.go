package test

import (
	"encoding/hex"
	"errors"
	"io"

	log "github.com/sirupsen/logrus"
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

func NewTagRef(name, hashStr string) *plumbing.Reference {
	refName := plumbing.ReferenceName(name)
	hashSlice, err := hex.DecodeString(hashStr)
	if err != nil {
		log.Fatalf("NewHashRef() expects hex string, but got %s", hashStr)
	}
	if len(hashSlice) != 20 {
		log.Fatalf("NewHashRef() expects hex string constituting 20 bytes but got %d bytes from %s", len(hashSlice), hashStr)
	}
	var fixedHash [20]byte
	for i := range fixedHash {
		fixedHash[i] = hashSlice[i]
	}
	return plumbing.NewHashReference("refs/tags/"+refName, plumbing.Hash(fixedHash))
}
