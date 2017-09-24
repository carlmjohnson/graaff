package fileutil

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// CopyTempDir takes a path and copies its files into a temporary directory
func CopyTempDir(src string) (tempdir string, err error) {
	dest, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}

	return dest, CopyRecursive(src, dest)
}

type copyPath struct{ src, dest string }

type errors []error

func (es errors) Error() string {
	ss := make([]string, 0, len(es))
	for _, e := range es {
		ss = append(ss, e.Error())
	}
	return strings.Join(ss, "; ")
}

// CopyRecursive recusively copies its source directory into a destination
// directory. It does not copy files or directories beginning with ".".
func CopyRecursive(src, dest string) error {
	var queue []copyPath

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if path == src {
			return nil
		}
		dir, name := filepath.Split(path)
		if strings.HasPrefix(name, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		destPath := filepath.Join(dest, dir, name)
		if info.IsDir() {
			return os.MkdirAll(destPath, 0700)
		}
		queue = append(queue, copyPath{src: path, dest: destPath})
		return err
	})

	if err != nil {
		return err
	}

	var (
		numWorkers = runtime.GOMAXPROCS(-1)
		copyC      = make(chan copyPath)
		errC       = make(chan error)
		wg         sync.WaitGroup
	)

	go func() {
		for _, c := range queue {
			copyC <- c
		}
		close(copyC)
	}()

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for c := range copyC {
				err := CopyFile(c.src, c.dest)
				if err != nil {
					errC <- err
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errC)
	}()

	var errs errors
	for err := range errC {
		errs = append(errs, err)
	}
	if len(errs) > 1 {
		return errs
	}
	return nil
}

// CopyFile copies a single source path to its destination. The enclosing
// folder for the destination is assumed to exist.
func CopyFile(src, dest string) error {
	fi, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fi.Close()

	fo, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer fo.Close()

	_, err = io.Copy(fo, fi)
	return err
}
