package sass

import (
	"errors"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type stringFile struct {
	name     string
	contents []byte
	modTime  time.Time
	offset   int64
}

func (f *stringFile) Stat() (os.FileInfo, error) {
	if string(f.contents) == "err:stat" {
		return nil, errors.New("nostat")
	}
	return f, nil
}
func (f *stringFile) Name() string       { return f.name }
func (f *stringFile) Size() int64        { return int64(len(f.contents)) }
func (f *stringFile) Mode() os.FileMode  { return os.FileMode(0) }
func (f *stringFile) ModTime() time.Time { return f.modTime }
func (f *stringFile) IsDir() bool        { return false }
func (f *stringFile) Sys() interface{}   { return nil }
func (f *stringFile) Close() error       { return nil }

func (f *stringFile) Readdir(int) ([]os.FileInfo, error) {
	return nil, errors.New("not a directory")
}

func (f *stringFile) Read(b []byte) (int, error) {
	if string(f.contents) == "err:read" {
		return 0, errors.New("noread")
	}
	x := int64(len(f.contents)) - f.offset
	if x > int64(len(b)) {
		x = int64(len(b))
	}
	if x > 0 {
		copy(b, f.contents[f.offset:f.offset+x])
		f.offset += x
		return int(x), nil
	}
	return 0, io.EOF
}

func (f *stringFile) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64
	switch whence {
	case 0:
		newOffset = offset
	case 1:
		newOffset = f.offset + offset
	case 2:
		newOffset = int64(len(f.contents)) + offset
	}
	if newOffset < 0 {
		return newOffset, errors.New("invalid offset")
	}
	f.offset = newOffset
	return f.offset, nil
}

type mapFS map[string]string

func (fs mapFS) Open(name string) (http.File, error) {
	f, ok := fs[name]
	if ok {
		return &stringFile{name, []byte(f), time.Now(), 0}, nil
	} else {
		return nil, errors.New("file not found")
	}
}

func TestOpenRelative(t *testing.T) {
	fs := mapFS{"A": "aaa", "B": "bbbb", "C": "c", "nostat": "err:stat", "noread": "err:read"}
	httpFile, err := fs.Open("A")
	if err != nil {
		t.Fatal("failed to open fake file")
	}
	f, err := openFile(fs, httpFile, "A")
	if err != nil {
		t.Fatal("failed to open fake file")
	}

	Convey("should open relative file", t, func() {
		g, err := f.OpenRelative("B")
		So(err, ShouldBeNil)
		So(g.Name, ShouldEqual, "B")
		So(string(g.Bytes), ShouldEqual, fs["B"])

		h, err := g.OpenRelative("C")
		So(h.Name, ShouldEqual, "C")
		So(string(h.Bytes), ShouldEqual, fs["C"])
	})

	Convey("should pass error through", t, func() {
		_, err := f.OpenRelative("doesn't exist")
		So(err.Error(), ShouldEqual, "file not found")
		_, err = f.OpenRelative("nostat")
		So(err.Error(), ShouldEqual, "nostat")
		_, err = f.OpenRelative("noread")
		So(err.Error(), ShouldEqual, "noread")
	})
}
