package mman

// Use man for documentation of syscall
import "syscall"

func Mmap(addr uintptr, length, prot, flag int, fd uintptr, offset int64) (uintptr, error) {
	r1, _, e1 := syscall.Syscall6(
		syscall.SYS_MMAP,
		uintptr(addr),
		uintptr(length),
		uintptr(prot),
		uintptr(flag),
		uintptr(fd),
		uintptr(offset),
	)

	if e1 != 0 {
		return 0, syscall.Errno(e1)
	}
	return r1, nil
}

func Munmap(addr uintptr, length int) error {
	_, _, e1 := syscall.Syscall(
		syscall.SYS_MUNMAP,
		uintptr(addr),
		uintptr(length),
		uintptr(0),
	)

	if e1 != 0 {
		return syscall.Errno(e1)
	}
	return nil
}

const (
	IPC_PRIVATE = 0
	IPC_CREAT   = 512
	IPC_RMID    = 0

	// linux specific
	SHM_REMAP = 16384
)

func Shmget(key uintptr, size int, flag uintptr) (int, error) {
	r1, _, e1 := syscall.Syscall(
		syscall.SYS_SHMGET,
		key,
		uintptr(size),
		flag,
	)

	if e1 != 0 {
		return 0, syscall.Errno(e1)
	}
	return int(r1), nil
}

func Shmat(id int, addr, flag uintptr) (uintptr, error) {
	r1, _, e1 := syscall.Syscall(
		syscall.SYS_SHMAT,
		uintptr(id),
		addr,
		flag,
	)

	if e1 != 0 {
		return 0, syscall.Errno(e1)
	}
	return r1, nil
}

func Shmctl(id, cmd int, buf uintptr) error {
	_, _, e1 := syscall.Syscall(
		syscall.SYS_SHMCTL,
		uintptr(id),
		uintptr(cmd),
		buf,
	)

	if e1 != 0 {
		return syscall.Errno(e1)
	}
	return nil
}

func Shmdt(addr uintptr) error {
	_, _, e1 := syscall.Syscall(
		syscall.SYS_SHMDT,
		addr, 0, 0,
	)
	if e1 != 0 {
		return syscall.Errno(e1)
	}
	return nil
}

// linux specific
func RemapFilePages(addr uintptr, size, prot, pgoff, flags int) error {
	_, _, e1 := syscall.Syscall6(
		syscall.SYS_REMAP_FILE_PAGES,
		addr,
		uintptr(size),
		uintptr(prot),
		uintptr(pgoff),
		uintptr(flags),
		0,
	)

	if e1 != 0 {
		return syscall.Errno(e1)
	}
	return nil
}
