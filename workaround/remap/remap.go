package remap

import (
	"reflect"
	"ringbuffer/mman"
	"syscall"
	"unsafe"
)

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

type mapper struct {
	active *mman.SliceSet
}

// Maker is an implementation of ringbuffer.MirroredSliceMaker
var Maker = &mapper{mman.NewSliceSet()}

func (m *mapper) MakeMirroredSlice(size int) ([]byte, error) {
	size = mman.AlignedSize(size)

	fd := -1
	addr, err := mman.Mmap(0, 2*size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_ANONYMOUS, uintptr(fd), 0)
	if err != nil {
		return nil, err
	}

	// munmap on error, munmap should never fail
	defer func() {
		if err != nil {
			panicIf(mman.Munmap(addr, 2*size))
		}
	}()

	err = mman.RemapFilePages(addr+uintptr(size), size, 0, 0, 0)

	if err != nil {
		return nil, err
	}

	sh := reflect.SliceHeader{
		Data: addr,
		Len:  2 * size,
		Cap:  2 * size,
	}

	buf := *(*[]byte)(unsafe.Pointer(&sh))

	m.active.Append(sh)

	return buf, nil
}

func (m *mapper) Free(buf []byte) error {
	sh := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))

	if !m.active.HasAndRemove(sh) {
		return syscall.EINVAL
	}

	return mman.Munmap(sh.Data, sh.Cap)
}
