// Copyright 2018 The Chubao Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package metanode

import (
	"github.com/chubaofs/chubaofs/util/btree"
	"sync"
)

const defaultBTreeDegree = 32

type (
	// BtreeItem type alias google btree Item
	BtreeItem = btree.Item
)

var _ Snapshot = &BTreeSnapShot{}

type BTreeSnapShot struct {
	inode     *InodeBTree
	dentry    *DentryBTree
	extend    *ExtendBTree
	multipart *MultipartBTree
}

func (b *BTreeSnapShot) Range(tp TreeType, cb func(v []byte) (bool, error)) error {
	switch tp {
	case InodeType:
		return b.inode.Range(&Inode{}, nil, cb)
	case DentryType:
		return b.dentry.Range(&Dentry{}, nil, cb)
	case ExtendType:
		return b.extend.Range(&Extend{}, nil, cb)
	case MultipartType:
		return b.multipart.Range(&Multipart{}, nil, cb)
	}
	panic("out of type")
}

func (b *BTreeSnapShot) Close() {}

func (b *BTreeSnapShot) Count(tp TreeType) (uint64, error) {
	switch tp {
	case InodeType:
		return uint64(b.inode.Len()), nil
	case DentryType:
		return uint64(b.dentry.Len()), nil
	case ExtendType:
		return uint64(b.extend.Len()), nil
	case MultipartType:
		return uint64(b.multipart.Len()), nil
	}
	panic("out of type")
}

var _ InodeTree = &InodeBTree{}
var _ DentryTree = &DentryBTree{}
var _ ExtendTree = &ExtendBTree{}
var _ MultipartTree = &MultipartBTree{}

type InodeBTree struct {
	*BTree
}

type DentryBTree struct {
	*BTree
}
type ExtendBTree struct {
	*BTree
}
type MultipartBTree struct {
	*BTree
}

//get
func (i *InodeBTree) Get(ino uint64) (*Inode, error) {
	return i.BTree.CopyGet(&Inode{Inode: ino}).(*Inode), nil
}
func (i *DentryBTree) Get(pid uint64, name string) (*Dentry, error) {
	return i.BTree.CopyGet(&Dentry{ParentId: pid, Name: name}).(*Dentry), nil
}
func (i *ExtendBTree) Get(ino uint64) (*Extend, error) {
	return i.BTree.CopyGet(&Extend{inode: ino}).(*Extend), nil
}
func (i *MultipartBTree) Get(key, id string) (*Multipart, error) {
	return i.BTree.CopyGet(&Multipart{key: key, id: id}).(*Multipart), nil
}

//put
func (i *InodeBTree) Put(inode *Inode) error {
	i.BTree.ReplaceOrInsert(inode, true)
	return nil
}
func (i *DentryBTree) Put(dentry *Dentry) error {
	i.BTree.ReplaceOrInsert(dentry, true)
	return nil
}
func (i *ExtendBTree) Put(extend *Extend) error {
	i.BTree.ReplaceOrInsert(extend, true)
	return nil
}
func (i *MultipartBTree) Put(multipart *Multipart) error {
	i.BTree.ReplaceOrInsert(multipart, true)
	return nil
}

//create
func (i *InodeBTree) Create(inode *Inode) error {
	_, ok := i.BTree.ReplaceOrInsert(inode, false)
	if ok {
		return nil
	}
	return existsError
}
func (i *DentryBTree) Create(dentry *Dentry) error {
	_, ok := i.BTree.ReplaceOrInsert(dentry, false)
	if ok {
		return nil
	}
	return existsError
}
func (i *ExtendBTree) Create(extend *Extend) error {
	_, ok := i.BTree.ReplaceOrInsert(extend, false)
	if ok {
		return nil
	}
	return existsError
}
func (i *MultipartBTree) Create(mul *Multipart) error {
	_, ok := i.BTree.ReplaceOrInsert(mul, false)
	if ok {
		return nil
	}
	return existsError
}

//delete
func (i *InodeBTree) Delete(ino uint64) error {
	i.BTree.Delete(&Inode{Inode: ino})
	return nil
}
func (i *DentryBTree) Delete(pid uint64, name string) error {
	i.BTree.Delete(&Dentry{ParentId: pid, Name: name})
	return nil
}
func (i *ExtendBTree) Delete(ino uint64) error {
	i.BTree.Delete(&Extend{inode: ino})
	return nil
}
func (i *MultipartBTree) Delete(key, id string) error {
	i.BTree.Delete(&Multipart{key: key, id: id})
	return nil
}

//range
func (i *InodeBTree) Range(start, end *Inode, cb func(v []byte) (bool, error)) error {
	var (
		err  error
		bs   []byte
		next bool
	)

	callback := func(i BtreeItem) bool {
		bs, err = i.(*Inode).Marshal()
		if err != nil {
			return false
		}
		next, err = cb(bs)
		if err != nil {
			return false
		}
		return next
	}

	if end == nil {
		i.BTree.AscendGreaterOrEqual(start, callback)
	} else {
		i.BTree.AscendRange(start, end, callback)
	}

	return err
}
func (i *DentryBTree) Range(start, end *Dentry, cb func(v []byte) (bool, error)) error {
	var (
		err  error
		bs   []byte
		next bool
	)
	callback := func(i BtreeItem) bool {
		bs, err = i.(*Dentry).Marshal()
		if err != nil {
			return false
		}
		next, err = cb(bs)
		if err != nil {
			return false
		}
		return next
	}

	if end == nil {
		i.BTree.AscendGreaterOrEqual(start, callback)
	} else {
		i.BTree.AscendRange(start, end, callback)
	}
	return err
}
func (i *ExtendBTree) Range(start, end *Extend, cb func(v []byte) (bool, error)) error {
	var (
		err  error
		bs   []byte
		next bool
	)
	callback := func(i BtreeItem) bool {
		bs, err = i.(*Extend).Bytes()
		if err != nil {
			return false
		}
		next, err = cb(bs)
		if err != nil {
			return false
		}
		return next
	}

	if end == nil {
		i.BTree.AscendGreaterOrEqual(start, callback)
	} else {
		i.BTree.AscendRange(start, end, callback)
	}

	return err
}
func (i *MultipartBTree) Range(start, end *Multipart, cb func(v []byte) (bool, error)) error {
	var (
		err  error
		bs   []byte
		next bool
	)
	callback := func(i BtreeItem) bool {
		bs, err = i.(*Multipart).Bytes()
		if err != nil {
			return false
		}
		next, err = cb(bs)
		if err != nil {
			return false
		}
		return next
	}

	if end == nil {
		i.BTree.AscendGreaterOrEqual(start, callback)
	} else {
		i.BTree.AscendRange(start, end, callback)
	}
	return err
}

// BTree is the wrapper of Google's btree.
type BTree struct {
	sync.RWMutex
	tree *btree.BTree
}

// NewBtree creates a new btree.
func NewBtree() *BTree {
	return &BTree{
		tree: btree.New(defaultBTreeDegree),
	}
}

// Get returns the object of the given key in the btree.
func (b *BTree) Get(key BtreeItem) (item BtreeItem) {
	b.RLock()
	item = b.tree.Get(key)
	b.RUnlock()
	return
}

func (b *BTree) CopyGet(key BtreeItem) (item BtreeItem) {
	b.Lock()
	item = b.tree.CopyGet(key)
	b.Unlock()
	return
}

// Find searches for the given key in the btree.
func (b *BTree) Find(key BtreeItem, fn func(i BtreeItem)) {
	b.RLock()
	item := b.tree.Get(key)
	b.RUnlock()
	if item == nil {
		return
	}
	fn(item)
}

func (b *BTree) CopyFind(key BtreeItem, fn func(i BtreeItem)) {
	b.Lock()
	item := b.tree.CopyGet(key)
	fn(item)
	b.Unlock()
}

// Has checks if the key exists in the btree.
func (b *BTree) Has(key BtreeItem) (ok bool) {
	b.RLock()
	ok = b.tree.Has(key)
	b.RUnlock()
	return
}

// Delete deletes the object by the given key.
func (b *BTree) Delete(key BtreeItem) (item BtreeItem) {
	b.Lock()
	item = b.tree.Delete(key)
	b.Unlock()
	return
}

func (b *BTree) Execute(fn func(tree *btree.BTree) interface{}) interface{} {
	b.Lock()
	defer b.Unlock()
	return fn(b.tree)
}

// ReplaceOrInsert is the wrapper of google's btree ReplaceOrInsert.
func (b *BTree) ReplaceOrInsert(key BtreeItem, replace bool) (item BtreeItem, ok bool) {
	b.Lock()
	if replace {
		item = b.tree.ReplaceOrInsert(key)
		b.Unlock()
		ok = true
		return
	}

	item = b.tree.Get(key)
	if item == nil {
		item = b.tree.ReplaceOrInsert(key)
		b.Unlock()
		ok = true
		return
	}
	ok = false
	b.Unlock()
	return
}

// Ascend is the wrapper of the google's btree Ascend.
// This function scans the entire btree. When the data is huge, it is not recommended to use this function online.
// Instead, it is recommended to call GetTree to obtain the snapshot of the current btree, and then do the scan on the snapshot.
func (b *BTree) Ascend(fn func(i BtreeItem) bool) {
	b.RLock()
	b.tree.Ascend(fn)
	b.RUnlock()
}

// AscendRange is the wrapper of the google's btree AscendRange.
func (b *BTree) AscendRange(greaterOrEqual, lessThan BtreeItem, iterator func(i BtreeItem) bool) {
	b.RLock()
	b.tree.AscendRange(greaterOrEqual, lessThan, iterator)
	b.RUnlock()
}

// AscendGreaterOrEqual is the wrapper of the google's btree AscendGreaterOrEqual
func (b *BTree) AscendGreaterOrEqual(pivot BtreeItem, iterator func(i BtreeItem) bool) {
	b.RLock()
	b.tree.AscendGreaterOrEqual(pivot, iterator)
	b.RUnlock()
}

// GetTree returns the snapshot of a btree.
func (b *BTree) GetTree() *BTree {
	b.Lock()
	t := b.tree.Clone()
	b.Unlock()
	nb := NewBtree()
	nb.tree = t
	return nb
}

// Reset resets the current btree.
func (b *BTree) Reset() {
	b.Lock()
	b.tree.Clear(true)
	b.Unlock()
}

func (i *BTree) Release() {
	i.Reset()
}

func (i *BTree) SetApplyID(index uint64) {
}

func (i *BTree) Flush() error {
	panic("implement me")
}

func (i *BTree) Count() uint64 {
	return uint64(i.Len())
}

// Len returns the total number of items in the btree.
func (b *BTree) Len() (size int) {
	b.RLock()
	size = b.tree.Len()
	b.RUnlock()
	return
}

// MaxItem returns the largest item in the btree.
func (b *BTree) MaxItem() BtreeItem {
	b.RLock()
	item := b.tree.Max()
	b.RUnlock()
	return item
}