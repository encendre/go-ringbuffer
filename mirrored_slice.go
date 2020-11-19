package ringbuffer

import (
	"runtime"
	"syscall"
	"unsafe"

	"ringbuffer/mman"
)

func panicIf(errs ...error) {
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
}

func mmapFixed(addr uintptr, size int) error {
	_, err := mman.Mmap(
		addr,
		size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_PRIVATE|syscall.MAP_FIXED|syscall.MAP_ANONYMOUS,
		^uintptr(0), // == uintptr(-1)
		0,
	)
	return err
}

func makeMirroredSlice(size int) ([]byte, error) {
	size = mman.AlignedSize(size)
	buf := mman.AlignedMake(2 * size)
	addr := uintptr(unsafe.Pointer(&buf[0]))

	shmid, err := mman.Shmget(mman.IPC_PRIVATE, size, mman.IPC_CREAT|0600) // 0600 is read and write autorisation -rw------
	if err != nil {
		return nil, err
	}

	// Mark the segment to be destroyed after all addr are detached with shmdt
	defer func() {
		// should never fail
		panicIf(mman.Shmctl(shmid, mman.IPC_RMID, 0))
	}()

	_, err = mman.Shmat(shmid, addr, mman.SHM_REMAP)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			// should never fail
			panicIf(
				mman.Shmdt(addr),
				mmapFixed(addr, size), // remapping deleted segment data
			)
		}
	}()

	_, err = mman.Shmat(shmid, addr+uintptr(size), mman.SHM_REMAP)
	if err != nil {
		return nil, err
	}

	runtime.SetFinalizer(&buf[0], func(p *byte) {
		addr := uintptr(unsafe.Pointer(p))
		// should never fail
		panicIf(
			mman.Shmdt(addr),
			mman.Shmdt(addr+uintptr(size)),
			// this is evil
			// remapping deleted segment data
			// maybe race condition if GC access the memory area between shmdt and mmap
			mmapFixed(addr, 2*size),
		)
	})

	return buf, nil
}
