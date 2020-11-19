package mmapshm

import (
	"bytes"
	"syscall"
	"testing"
)

func TestMirroredSliceBasic(t *testing.T) {
	var err error
	buf, err := Maker.MakeMirroredSlice(10)
	if err != nil {
		t.Errorf("MakeMirroredSlice should not fail, have: %w\n", err)
	}

	size := len(buf) / 2
	buf1 := buf[:size]
	buf2 := buf[size:]

	data := []byte("Testazefvavlejzmefafzfdvbzregvfdsgfdgsdfazzev")
	copy(buf1, data)
	if bytes.Compare(buf1, buf2) != 0 {
		Maker.Free(buf)
		t.Errorf("memory should be mirrored properly, have:\nbuf1: %s\nbuf2: %s\n", buf1[:len(data)], buf2[:len(data)])
	}

	data = []byte("b454S5F4fs54Z5DF4F5D4S5V54S5Ve")
	copy(buf2, data)
	if bytes.Compare(buf1, buf2) != 0 {
		Maker.Free(buf)
		t.Errorf("memory should be mirrored properly, have:\nbuf1: %s\nbuf2: %s\n", buf1[:len(data)], buf2[:len(data)])
	}

	err = Maker.Free(buf)
	if err != nil {
		t.Errorf("Free should not fail, have: %w\n", err)
	}

	err = Maker.Free(buf)
	if err != syscall.EINVAL {
		t.Errorf("Double free should fail with EINVAL, have: %w\n", err)
	}
}
