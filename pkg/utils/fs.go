package utils

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func CopyDir(src, dst string) error {
	return fs.WalkDir(os.DirFS(src), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, path), os.ModePerm)
		}

		out, err := os.Create(filepath.Join(dst, path))
		if err != nil {
			return err
		}
		defer out.Close()

		in, err := os.Open(filepath.Join(src, path))
		if err != nil {
			return err
		}
		defer in.Close()

		_, err = io.Copy(out, in)
		return err
	})
}
