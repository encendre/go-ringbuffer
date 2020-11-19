package tmpfile

import (
	"io/ioutil"
	"os"
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

	nofd := ^uintptr(0) // == uintptr(-1)
	addr, err := mman.Mmap(0, 2*size, syscall.PROT_NONE, syscall.MAP_PRIVATE|syscall.MAP_ANONYMOUS, nofd, 0)
	if err != nil {
		return nil, err
	}

	// munmap on error, munmap should never fail
	defer func() {
		if err != nil {
			panicIf(mman.Munmap(addr, 2*size))
		}
	}()

	f, err := ioutil.TempFile("/tmp", "goringbuffer")
	if err != nil {
		return nil, err
	}

	// its ok to close file after being mmaped
	defer f.Close()

	// its ok to delete tmp file
	if err = os.Remove(f.Name()); err != nil {
		return nil, err
	}

	fd := f.Fd()
	if err = syscall.Ftruncate(int(fd), int64(size)); err != nil {
		return nil, err
	}

	_, err = mman.Mmap(addr, size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_FIXED, fd, 0)
	if err != nil {
		return nil, err
	}

	_, err = mman.Mmap(addr+uintptr(size), size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_FIXED, fd, 0)
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
