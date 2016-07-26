package bundle

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nitrous-io/rise-cli-go/pkg/pathmatch"
	"github.com/nitrous-io/rise-cli-go/progressbar"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"

	log "github.com/Sirupsen/logrus"
)

var (
	// From http://docs.aws.amazon.com/AmazonS3/latest/dev/UsingMetadata.html#object-keys
	FilenamePatternRe = regexp.MustCompile("[^0-9A-Za-z,!_'()\\.\\*\\-]+")

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
func (b *Bundle) Assemble(ignoreList []string, showWarnings bool) (count int, size int64, err error) {
	b.fileList = []string{}

	basePath, err := filepath.Abs(b.path)
	if err != nil {
		return 0, 0, err
	}

	walkFn := func(path string, fi os.FileInfo, err error) error {
		// if there is an error lstat-ing a file, just skip it
		if err != nil {
			return nil
		}

		incl, fileSize, err := shouldInclude(path, basePath, ignoreList, fi, showWarnings)
		if err != nil {
			// SkipDir error should fall through
			if err == filepath.SkipDir {
				return err
			}
			return nil
		}

		if incl {
			relPath, err := filepath.Rel(basePath, path)
			if err != nil {
				return err
			}

			b.fileList = append(b.fileList, relPath)
			size += fileSize
		}
		return nil
	}

	if err := filepath.Walk(basePath, walkFn); err != nil {
		b.fileList = []string{}
		return 0, 0, err
	}

	return len(b.fileList), size, nil
}

func shouldInclude(path, basePath string, ignoreList []string, fi os.FileInfo, showWarnings bool) (incl bool, fileSize int64, err error) {
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

	logWarn := func(m string) {
		if showWarnings {
			relPath, _ := filepath.Rel(basePath, path)
			if relPath == "" {
				relPath = path
			}
			log.Warnf(tr.T("ignore_file_reason"), relPath, m)
		}
	}

	// follow symlink
	if mode&os.ModeSymlink == os.ModeSymlink {
		var err error
		fi, err = os.Stat(path)
		// if there is an error following the symlink, skip
		if err != nil {
			logWarn(tr.T("symlink_error"))
			return skip()
		}

		// if symlink points to a directory, skip
		if fi.IsDir() {
			logWarn(tr.T("symlink_to_dir"))
			return skip()
		}

		// if symlink points to a non-regular file, skip
		if !fi.Mode().IsRegular() {
			logWarn(tr.T("special_mode_bits"))
			return skip()
		}
	}

	// if file name starts with ".", "#" or ends with "~", skip
	if base[0] == '.' {
		logWarn(tr.T("name_has_dot_prefix"))
		return skip()
	}

	if base[0] == '#' {
		logWarn(tr.T("name_has_hash_prefix"))
		return skip()
	}

	if base[len(base)-1] == '~' {
		logWarn(tr.T("name_has_tilde_suffix"))
		return skip()
	}

	if FilenamePatternRe.MatchString(base) {
		logWarn(tr.T("name_has_non_safe_character"))
		return skip()
	}

	// if file is in the ignore list, skip
	if ignoreList != nil && pathmatch.PathMatchAny(path, ignoreList...) {
		logWarn(tr.T("name_in_ignore_list"))
		return skip()
	}

	// if file is not a regular file or symlink to a regular file (tested earlier), skip
	if mode&os.ModeSymlink != os.ModeSymlink && !isDir && !mode.IsRegular() {
		logWarn(tr.T("special_mode_bits"))
		return skip()
	}

	// let this directory to be scanned
	if isDir {
		return false, 0, nil
	}

	// if the file can't be read, skip
	f, err := os.Open(path)
	if err != nil {
		logWarn(tr.T("file_unreadable"))
		return false, 0, nil
	}
	f.Close()

	return true, fi.Size(), nil
}

func (b *Bundle) FileList() []string {
	return b.fileList
}

func (b *Bundle) Pack(tarballPath string, showError, showProgress bool) error {
	logErr := func(m string) {
		if showError {
			log.Error(m)
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

	basePath, err := filepath.Abs(b.path)
	if err != nil {
		return err
	}

	pb := progressbar.NewCounter(tui.Out, len(b.fileList))

	for _, path := range b.fileList {
		absPath := filepath.Join(basePath, path)

		fi, err := os.Stat(absPath)
		if err != nil {
			logErr(fmt.Sprintf(tr.T("stat_failed"), absPath))
			return err
		}

		unixPath := path

		if filepath.Separator != '/' {
			unixPath = strings.Replace(path, string(filepath.Separator), "/", -1)
		}

		hdr, err := tar.FileInfoHeader(fi, unixPath)
		hdr.Name = unixPath
		if err != nil {
			logErr(fmt.Sprintf(tr.T("stat_failed"), absPath))
			return err
		}

		if err := tw.WriteHeader(hdr); err != nil {
			logErr(fmt.Sprintf(tr.T("write_failed"), tarballPath))
			return err
		}

		ff, err := os.Open(absPath)
		if err != nil {
			logErr(fmt.Sprintf(tr.T("write_failed"), tarballPath))
			return err
		}

		n, err := io.Copy(tw, ff)
		ff.Close()

		if err != nil {
			logErr(fmt.Sprintf(tr.T("write_failed"), tarballPath))
			return err
		}

		if n != hdr.Size {
			logErr(fmt.Sprintf(tr.T("file_size_changed"), tarballPath))
			return ErrFileChanged
		}

		if showProgress {
			pb.Next()
		}
	}

	return nil
}

func Sha256Sum(path string) (string, error) {
	hasher := sha256.New()

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
