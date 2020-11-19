package ringbuffer

import (
	"bytes"
	"io"
	"math/rand"
	"testing"
)

func TestRingBufferBasicOp(t *testing.T) {
	rb, err := New(10)
	if err != nil {
		t.Errorf("New should not fail, have: %w\n", err)
	}

	size := rb.Cap()

	if !rb.Empty() {
		t.Errorf("Empty should return true when RingBuffer is empty. want: %v, have: %v\n", true, false)
	}

	src := make([]byte, size+10)
	rand.Read(src)

	written, err := rb.Write(src)
	if err != io.ErrShortWrite {
		t.Errorf("Write too long buffer should fail. want: %v, have: %w\n", io.ErrShortBuffer, err)
	}
	if written != size {
		t.Errorf("Write should return written bytes. want: %v, have: %v\n", size, written)
	}

	if !rb.Full() {
		t.Errorf("Full should return true when RingBuffer is full, want: %v, have: %v\n", true, false)
	}
	if err != io.ErrShortWrite {
		t.Errorf("Write should not fail. want: %v, have: %w\n", nil, err)
	}

	dst := make([]byte, len(src))
	readed, err := rb.Read(dst)
	if err != nil {
		t.Errorf("Read too long buffer should not fail. want: %v, have: %w\n", nil, err)
	}
	if written != readed {
		t.Errorf("Read should return readed bytes. want: %v, have: %v\n", written, readed)
	}

	if bytes.Compare(src[:written], dst[:readed]) != 0 {
		t.Errorf("Read give wrong bytes:\nwant: %x\nhave: %x", src[:written], dst[:readed])
	}

	//t.Errorf("%+v\n", rb)
}
