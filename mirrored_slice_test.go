package ringbuffer

import (
	"bytes"
	"runtime"
	"testing"
)

func TestMirroredSliceBasic(t *testing.T) {
	buf, err := makeMirroredSlice(10)
	if err != nil {
		t.Errorf("MakeMirroredSlice should not fail, have: %w\n", err)
	}

	size := len(buf) / 2
	buf1 := buf[:size]
	buf2 := buf[size:]

	data := []byte("Testazefvavlejzmefafzfdvbzregvfdsgfdgsdfazzev")
	copy(buf1, data)
	if bytes.Compare(buf1, buf2) != 0 {
		t.Errorf("memory should be mirrored properly, have:\nbuf1: %s\nbuf2: %s\n", buf1[:len(data)], buf2[:len(data)])
	}

	data = []byte("b454S5F4fs54Z5DF4F5D4S5V54S5Ve")
	copy(buf2, data)
	if bytes.Compare(buf1, buf2) != 0 {
		t.Errorf("memory should be mirrored properly, have:\nbuf1: %s\nbuf2: %s\n", buf1[:len(data)], buf2[:len(data)])
	}
}

// Dirty test for testing safety
// actually if i remove mmapping part on finalizer, i have segfault
func TestMirroredSliceSafety(t *testing.T) {
	data := []byte("Testazefvavlejzmefafzfdvbzregvfdsgfdgsdfazzev")

	for n := 0; n < 100; n++ {
		buf, err := makeMirroredSlice(10)
		copy(buf, data)

		if err != nil {
			t.Errorf("MakeMirroredSlice should not fail, have %w\n", err)
		}

		runtime.GC()
	}
}
