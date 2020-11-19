package mmapshm

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
	addr, err := mman.Mmap(0, 2*size, syscall.PROT_NONE, syscall.MAP_SHARED|syscall.MAP_ANONYMOUS, uintptr(fd), 0)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			panicIf(mman.Munmap(addr, 2*size))
		}
	}()

	// 0600 means read and write autorisation for user, like -rw------
	flag := uintptr(0600)
	shmid, err := mman.Shmget(mman.IPC_PRIVATE, size, mman.IPC_CREAT|flag)
	if err != nil {
		return nil, err
	}

	// Mark the segment to be destroyed after all addr are detached with shmdt
	defer func() {
		panicIf(mman.Shmctl(shmid, mman.IPC_RMID, 0))
	}()

	// should shmdt on error maybe
	_, err = mman.Shmat(shmid, addr, mman.SHM_REMAP)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			panicIf(mman.Shmdt(addr))
		}
	}()

	_, err = mman.Shmat(shmid, addr+uintptr(size), mman.SHM_REMAP)
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

	err1 := mman.Shmdt(sh.Data)
	err2 := mman.Shmdt(sh.Data + uintptr(sh.Cap/2))
	if err1 != nil {
		return err1
	}
	return err2
}
