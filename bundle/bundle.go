package bundle

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nitrous-io/rise-cli-go/pkg/pathmatch"
)

var (
	ErrFileChanged = errors.New("file changed while processing")
)

type Bundle struct {
	path     string
	fileList []string
}

func New(path string) *Bundle {
	return &Bundle{path: path}
}

// Walks the path and forms a list of files that should be included in the bundle
func (b *Bundle) Assemble(ignoreList []string, verbose bool) (count int, size int64, err error) {
	b.fileList = []string{}

	absPath, err := filepath.Abs(b.path)
	if err != nil {
		return 0, 0, err
	}

	walkFn := func(path string, fi os.FileInfo, err error) error {
		// if there is an error lstat-ing a file, just skip it
		if err != nil {
			return nil
		}

		incl, fileSize, err := shouldInclude(path, ignoreList, fi, verbose)
		if err != nil {
			// SkipDir error should fall through
			if err == filepath.SkipDir {
				return err
			}
			return nil
		}

		if incl {
			relPath, err := filepath.Rel(absPath, path)
			if err != nil {
				return err
			}

			b.fileList = append(b.fileList, relPath)
			size += fileSize
		}
		return nil
	}

	if err := filepath.Walk(absPath, walkFn); err != nil {
		b.fileList = []string{}
		return 0, 0, err
	}

	return len(b.fileList), size, nil
}

func shouldInclude(path string, ignoreList []string, fi os.FileInfo, verbose bool) (incl bool, fileSize int64, err error) {
	var (
		isDir = fi.IsDir()
		mode  = fi.Mode()
		base  = filepath.Base(path)
	)

	// if path is a directory, skip the entire directory by returning SkipDir
	skip := func() (bool, int64, error) {
		if isDir {
			return false, 0, filepath.SkipDir
		}
		return false, 0, nil
	}

	log := func(m string) {
		if verbose {
			fmt.Printf("Warning: Ignoring \"%s\", %s\n", path, m)
		}
	}

	// follow symlink
	if mode&os.ModeSymlink == os.ModeSymlink {
		var err error
		fi, err = os.Stat(path)
		// if there is an error following the symlink, skip
		if err != nil {
			log("could not follow symlink")
			return skip()
		}

		// if symlink points to a directory, skip
		if fi.IsDir() {
			log("symlink points to a directory")
			return skip()
		}

		// if symlink points to a non-regular file, skip
		if !fi.Mode().IsRegular() {
			log("file has special mode bits set")
			return skip()
		}
	}

	// if file name starts with ".", "#" or ends with "~", skip
	if base[0] == '.' {
		log(`name begins with "."`)
		return skip()
	}

	if base[0] == '#' {
		log(`name begins with "#"`)
		return skip()
	}

	if base[len(base)-1] == '~' {
		log(`name ends with "~"`)
		return skip()
	}

	// if file is in the ignore list, skip
	if ignoreList != nil && pathmatch.PathMatchAny(path, ignoreList...) {
		log("name is in ignore list")
		return skip()
	}

	// if file is not a regular file or symlink to a regular file (tested earlier), skip
	if mode&os.ModeSymlink != os.ModeSymlink && !isDir && !mode.IsRegular() {
		log("file has special mode bits set")
		return skip()
	}

	// let this directory to be scanned
	if isDir {
		return false, 0, nil
	}

	// if the file can't be read, skip
	f, err := os.Open(path)
	if err != nil {
		log("file can't be read")
		return false, 0, nil
	}
	f.Close()

	return true, fi.Size(), nil
}

func (b *Bundle) FileList() []string {
	return b.fileList
}

func (b *Bundle) Pack(tarballPath string, verbose bool) error {
	log := func(m string) {
		if verbose {
			fmt.Printf("Error: %s\n", m)
		}
	}

	f, err := os.OpenFile(tarballPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer func() {
		gw.Flush()
		gw.Close()
	}()

	tw := tar.NewWriter(gw)
	defer func() {
		tw.Flush()
		tw.Close()
	}()

	for _, path := range b.fileList {
		fi, err := os.Stat(path)
		if err != nil {
			log(fmt.Sprintf("Could not get file info for \"%s\", aborting!", path))
			return err
		}

		hdr, err := tar.FileInfoHeader(fi, path)
		hdr.Name = path
		if err != nil {
			log(fmt.Sprintf("Could not get file info for \"%s\", aborting!", path))
			return err
		}

		if err := tw.WriteHeader(hdr); err != nil {
			log(fmt.Sprintf("Failed to write to \"%s\", aborting!", tarballPath))
			return err
		}

		ff, err := os.Open(path)
		if err != nil {
			log(fmt.Sprintf("Failed to write to \"%s\", aborting!", tarballPath))
			return err
		}

		n, err := io.Copy(tw, ff)
		ff.Close()

		if err != nil {
			log(fmt.Sprintf("Failed to write to \"%s\", aborting!", tarballPath))
			return err
		}

		if n != hdr.Size {
			log(fmt.Sprintf("File size of \"%s\" changed while packing, aborting!", tarballPath))
			return ErrFileChanged
		}
	}

	return nil
}
