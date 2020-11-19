# ringbuffer

## ⚠ Disclaimer
This package is not portable and uses a lot of unsafe, moreover it plays with the memory provided by the runtime and therefore the slightest bug can be catastrophic in production.

It has only been tested on my machine (`x86_64 Linux 5.9.8-zen1-1-zen`) and uses several linux-specific system calls, some of which are deprecated. 

## Description

This package provides several proofs of concept of mirrored memory as well as an implementation of a ringbuffer using it.

The default RingBuffer is supposed to work without having to free it manually, however the implementation is not POSIX and take advantage of several linux specific system calls.
For the POSIX way see `workaround/tmpfile`.

## Examples

### Using default maker

```go
// No need to explicitly free buffer he is garbage collected when unreachable
rb, _ := ringbuffer.New(10)

rb.Write([]byte("test\n"))
rb.WriteToN(os.Stdout, rb.Len()) // Print "test\n"

rb.Write([]byte("test2"))
buf := rb.Peek(n)                // Return slice to underlying memory

rb = nil                         // Dropping ringbuffer
runtime.GC()
runtime.GC()

// Since we still have a slice to underlying memory,
// slice is not garbage collected, so we can access it safely
fmt.Printf("%s\n", buf)

buf = nil                        // Underlying memory is garbage collected
```

### Using custom maker

```go
rb, _ := ringbuffer.NewWithMaker(10, tmpfile.Maker)
rb.Write([]byte("test2"))
buf := rb.Peek(n)                // return slice to underlying memory

// This time underlying memory is not in the GC area so there
// is no garbage collection. We must explicitly free underlying
// memory calling (*RingBuffer).Free.
rb.Free()

// On the memory is freed, underlying memory is not valid
fmt.Printf("%s\n", buf)          // Undefined behavior, may segfault
```

## Explanations

The principle is to map the same physical memory to two adjacent memory areas, all this can only be done by the kernel using syscalls.

### Default maker

The default slice maker get memory with `make()`, we use the `shma` syscall with the same shmid and the `SHM_REMAP` flag (linux specific) on the two adjacent areas. With `runtime.SetFinalizer`, when the memory area becomes unreachable, the segments are detached with `shmdt`, but since the memory area is provided by the runtime it may want to use it afterwards, so for preveting undefined behavior such as segfault, we alloc again the memory with `mmap` syscall with `MAP_FIXED` flag.

Note: the `runtime.SetFinalizer` works only because the memory is provided by the runtime

### Maker using temporary file (POSIX way)

We create a temporary file, that we map to two adjacent memory areas using `mmap` syscall with `MAP_FIXED` flag. When we're done, we simply release memory manually with `munmap`.

### Maker using mmap and shm

It is mainly the same as default maker, the main the difference is that the memory is given by `mmap` and is freed manually with `munmap`.

### Maker using remap_page_files

Memory is allocated with `mmap` and we use `remap_page_files` (linux specific) syscall to mirrored them. When we're done we the memory with `munmap`.
