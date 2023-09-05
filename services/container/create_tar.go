package container

import (
	"archive/tar"
	"bytes"
	"code-runner/model"
	"compress/gzip"
	"errors"
	"os"
	"time"
)

var errorNil = errors.New("passed value is nil")

func createSourceTar(sourceFiles []*model.SourceFile) ([]byte, error) {
	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	defer w.Close()
	for _, f := range sourceFiles {
		hdr := new(tar.Header)
		hdr.Uid = 0
		hdr.Gid = 0
		hdr.Mode = 0755
		hdr.Name = f.Filename
		hdr.Size = int64(len(f.Content))
		hdr.AccessTime = time.Now()
		hdr.ModTime = time.Now()
		hdr.ChangeTime = time.Now()
		if err := w.WriteHeader(hdr); err != nil {

			return nil, err
		}
		if _, err := w.Write([]byte(f.Content)); err != nil {
			return nil, err
		}
		w.Flush()
	}
	return buf.Bytes(), nil
}

func createTar(filenames []string) ([]byte, error) {
	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	defer w.Close()
	for _, f := range filenames {
		file, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		hdr := new(tar.Header)
		hdr.Uid = 0
		hdr.Gid = 0
		hdr.Mode = 0755
		hdr.Name = f
		hdr.Size = int64(len(file))
		hdr.AccessTime = time.Now()
		hdr.ModTime = time.Now()
		hdr.ChangeTime = time.Now()
		if err := w.WriteHeader(hdr); err != nil {

			return nil, err
		}
		if _, err := w.Write(file); err != nil {
			return nil, err
		}
		w.Flush()
	}
	return buf.Bytes(), nil
}

func gzipTar(tar []byte) ([]byte, error) {
	if tar == nil {
		return nil, errorNil
	}
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	defer writer.Close()
	_, err := writer.Write(tar)
	if err != nil {
		return nil, err
	}
	writer.Flush()
	return buf.Bytes(), nil

}