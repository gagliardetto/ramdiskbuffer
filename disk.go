package ramdiskbuffer

import (
	"bytes"
	"io/ioutil"
	"os"
)

type Buffer struct {
	file *os.File
	buf  bytes.Buffer
}

type CommonInterface interface {
	Write(p []byte) (n int, err error)
	WriteString(s string) (n int, err error)
	Read(p []byte) (n int, err error)

	Len() int
	Close() error

	LenInt64() int64
	Remove() error
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with ErrTooLarge.
func (b *Buffer) Write(p []byte) (n int, err error) {
	if b.file != nil {
		return b.file.Write(p)
	}
	return b.buf.Write(p)
}

// WriteString appends the contents of s to the buffer, growing the buffer as
// needed. The return value n is the length of s; err is always nil. If the
// buffer becomes too large, WriteString will panic with ErrTooLarge.
func (b *Buffer) WriteString(s string) (n int, err error) {
	if b.file != nil {
		return b.file.WriteString(s)
	}
	return b.buf.WriteString(s)
}

// Read reads the next len(p) bytes from the buffer or until the buffer
// is drained. The return value n is the number of bytes read. If the
// buffer has no data to return, err is io.EOF (unless len(p) is zero);
// otherwise it is nil.
func (b *Buffer) Read(p []byte) (n int, err error) {
	if b.file != nil {
		return b.file.Read(p)
	}
	return b.buf.Read(p)
}

func New(toDisk bool) *Buffer {
	if toDisk {
		file, err := ioutil.TempFile("", "ramdiskbuffer")
		if err != nil {
			panic(err)
		}
		return &Buffer{
			file: file,
		}
	}

	return &Buffer{}
}

func (d *Buffer) Remove() error {
	if d.file != nil {
		d.file.Close()
		return os.Remove(d.file.Name())
	}
	// TODO: better remove buffer
	d.buf.Reset()
	return nil
}
func (d *Buffer) Len() int {
	if d.file != nil {
		err := d.file.Sync()
		if err != nil {
			// TODO: not panic??
			panic(err)
		}

		info, err := d.file.Stat()
		if err != nil {
			// TODO: not panic??
			panic(err)
		}
		return int(info.Size())
	}
	return d.buf.Len()
}
func (d *Buffer) LenInt64() int64 {
	return int64(d.Len())
}
func (d *Buffer) Close() error {
	if d.file != nil {
		err := d.file.Sync()
		if err != nil {
			return err
		}
		return d.file.Close()
	}
	return nil
}
func (d *Buffer) PrepareForReading() error {
	if d.file != nil {
		err := d.file.Sync()
		if err != nil {
			return err
		}
		_, err = d.file.Seek(0, 0)
		if err != nil {
			return err
		}
	}
	return nil
}

type BufferArray []*Buffer

func NewArray(length int, toDisk bool) BufferArray {
	buffers := make([]*Buffer, length)
	for i := range buffers {
		buffers[i] = New(toDisk)
	}
	return buffers
}

// PrepareForReading sets all the buffers in read mode
// (i.e. flushes to disk, and seeks to the beginning of the file);
// after PrepareForReading, no writing should be done.
func (ba BufferArray) PrepareForReading() error {
	for _, buf := range ba {
		// prepare for reading
		err := buf.PrepareForReading()
		if err != nil {
			return err
		}
	}
	return nil
}
func (ba BufferArray) Remove() error {
	for _, buf := range ba {
		// remove
		err := buf.Remove()
		if err != nil {
			return err
		}
	}
	return nil
}
