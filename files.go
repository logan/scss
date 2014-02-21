package sass

import (
	"io"
	"net/http" // for filesystem interfaces
)

type File struct {
	http.FileSystem
	http.File
	Name  string
	Bytes []byte
}

func (f *File) OpenRelative(name string) (*File, error) {
	rel, err := f.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	return openFile(f.FileSystem, rel, name)
}

func openFile(fs http.FileSystem, httpFile http.File, name string) (*File, error) {
	fileInfo, err := httpFile.Stat()
	if err != nil {
		return nil, err
	}
	f := &File{
		FileSystem: fs,
		File:       httpFile,
		Name:       name,
		Bytes:      make([]byte, fileInfo.Size()),
	}
	n, err := f.Read(f.Bytes)
	if err != nil && err != io.EOF {
		f.Close()
		return nil, err
	}
	f.Bytes = f.Bytes[:n]
	return f, nil
}
