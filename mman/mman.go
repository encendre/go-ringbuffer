//Package for Memory MANagement utility
package mman

import (
	"os"
	"reflect"
	"sync"
	"unsafe"
)

type SliceSet struct {
	sync.Mutex
	set map[reflect.SliceHeader]struct{}
}

func NewSliceSet() *SliceSet {
	return &SliceSet{
		set: make(map[reflect.SliceHeader]struct{}),
	}
}

func (m *SliceSet) HasAndRemove(sh reflect.SliceHeader) (ok bool) {
	m.Lock()
	_, ok = m.set[sh]
	delete(m.set, sh)
	m.Unlock()
	return
}

func (m *SliceSet) Append(sh reflect.SliceHeader) {
	m.Lock()
	m.set[sh] = struct{}{}
	m.Unlock()
}

func AlignedSize(size int) int {
	pageSize := os.Getpagesize()
	return (size/pageSize)*pageSize + pageSize
}

func AlignedMake(size int) []byte {
	pageSize := os.Getpagesize()

	// alloc buffer
	buf := make([]byte, size+pageSize)
	addr := uintptr(unsafe.Pointer(&buf[0]))

	// align buffer
	offset := pageSize - int(addr%uintptr(pageSize))
	buf = buf[offset : offset+size]

	return buf
}
