package container

import (
	"archive/tar"
	"bytes"
	"code-runner/model"
	"time"
)

func createTar(sourceFiles []*model.SourceFile) ([]byte, error) {
	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	defer w.Close()
	for _, f := range sourceFiles {
		hdr := new(tar.Header)
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
	}
	return buf.Bytes(), nil
}
