package ringbuffer

import (
	"io"
)

type op int

const (
	readOp op = iota
	writeOp
)

// RingBuffer a simple ringbuffer
type RingBuffer struct {
	data   []byte
	size   int
	r, w   int
	lastOp op
	mapper MirroredSliceMaker
}

// MirroredSliceMaker the interface used by the ring buffer to create mirrored memory
// Several implementations can be found in `workaround` folder of this repository
type MirroredSliceMaker interface {
	MakeMirroredSlice(int) ([]byte, error)
	Free([]byte) error
}

// NewWithMaker create new RingBuffer with the specified maker
func NewWithMaker(size int, msp MirroredSliceMaker) (*RingBuffer, error) {
	data, err := msp.MakeMirroredSlice(size)
	if err != nil {
		return nil, err
	}

	rb := &RingBuffer{
		data:   data,
		size:   cap(data) / 2,
		mapper: msp,
	}

	return rb, nil
}

// Free a RingBuffer invalidate all slice
// Using RingBuffer after Free lead to undefined behavior
// (panic, segfault, etc.)
func (rb *RingBuffer) Free() (err error) {
	if rb.mapper == nil {
		return
	}
	err = rb.mapper.Free(rb.data)

	// panic is better than segfault
	rb.data = rb.data[:0:0]
	return
}

// New create a RingBuffer using default MirroredSlice
func New(size int) (*RingBuffer, error) {
	data, err := makeMirroredSlice(size)
	if err != nil {
		return nil, err
	}

	return &RingBuffer{
		data: data,
		size: cap(data) / 2,
	}, nil
}

// Empty return true if buffer is empty
func (rb *RingBuffer) Empty() bool {
	return rb.r == rb.w && rb.lastOp == readOp
}

// Full return true if buffer is full
func (rb *RingBuffer) Full() bool {
	return rb.r == rb.w && rb.lastOp == writeOp
}

// Cap return buffer capacity
func (rb *RingBuffer) Cap() int {
	return rb.size
}

// Len return the amount of byte to be read
func (rb *RingBuffer) Len() int {
	r, w := rb.r2w()
	return w - r
}

// Reset reset all data
func (rb *RingBuffer) Reset() {
	rb.w = 0
	rb.r = 0
	rb.lastOp = readOp
}

// read to write offsets, use for reading purpose
func (rb *RingBuffer) r2w() (r, w int) {
	r = rb.r
	w = rb.w
	if r > w || rb.Full() {
		w += rb.size
	}
	return
}

// write to read offsets, use for writing purpose
func (rb *RingBuffer) w2r() (r, w int) {
	r = rb.r
	w = rb.w
	if w > r || rb.Empty() {
		r += rb.size
	}
	return
}

// Write
func (rb *RingBuffer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	r, w := rb.w2r()
	n = copy(rb.data[w:r], p)
	if n < len(p) {
		err = io.ErrShortWrite
	}

	rb.w = (rb.w + n) % rb.size

	rb.lastOp = writeOp
	return
}

// Read
func (rb *RingBuffer) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if rb.Empty() {
		return 0, io.EOF
	}
	r, w := rb.r2w()
	n = copy(p, rb.data[r:w])

	rb.r = (rb.r + n) % rb.size

	rb.lastOp = readOp
	return
}

func min(a, b int) int {
	if b < a {
		a = b
	}
	return a
}

// Peek return a slice containing the next n bytes without
// moving the read cursor
// Returned slice is valid until next Write, ReadFromN or Free
// Read or write in the slice after Free may segfault
func (rb *RingBuffer) Peek(n int) []byte {
	r, w := rb.r2w()
	return rb.data[r:min(r+n, w)]
}

// Skip move the cursor forward by n bytes
func (rb *RingBuffer) Skip(n int) error {
	if rb.Empty() {
		return io.EOF
	}

	r, w := rb.r2w()
	rb.r = min(r+n, w) % rb.size
	rb.lastOp = readOp
	return nil
}

// Next return a slice containing the next n bytes and
// move the cursor forward
// Returned slice is valid until next Write, ReadFromN or Free
// Read or write in the slice after Free may segfault
func (rb *RingBuffer) Next(n int) ([]byte, error) {
	return rb.Peek(n), rb.Skip(n)
}

// WriteToN write up to N bytes to the io.Writer
func (rb *RingBuffer) WriteToN(writer io.Writer, n int) (written int, err error) {
	if rb.Empty() {
		return 0, io.EOF
	}

	r, w := rb.r2w()

	p := rb.data[r:min(r+n, w)]

	written, err = writer.Write(p)
	rb.r = (rb.r + written) % rb.size
	rb.lastOp = readOp
	return
}

// ReadFromN read up to N bytes from the io.Reader
func (rb *RingBuffer) ReadFromN(reader io.Reader, n int) (readed int, err error) {
	if rb.Full() {
		return 0, io.ErrShortWrite
	}

	r, w := rb.w2r()

	p := rb.data[w:min(w+n, r)]

	readed, err = reader.Read(p)
	rb.w = (rb.w + readed) % rb.size
	if err == nil && readed < n {
		err = io.ErrShortWrite
	}
	rb.lastOp = writeOp
	return
}
