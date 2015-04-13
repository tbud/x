package ioutil

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CopyFlags uint

const (
	FlagDefault  CopyFlags = 0         // default flag
	FlagCopyHide CopyFlags = 1 << iota // if set this flag, whill copy hide dir ('.' prefix dir)
)

type CopyFilter func(dest, src string, srcInfo os.FileInfo) (skiped bool, err error)

func Copy(destDir, srcDir string, copyFlags CopyFlags, copyFilter CopyFilter) (err error) {
	if srcDir, err = filepath.Abs(srcDir); err != nil {
		return err
	}

	if destDir, err = filepath.Abs(destDir); err != nil {
		return err
	}

	var srcInfo os.FileInfo
	if srcInfo, err = os.Stat(srcDir); !os.IsNotExist(err) {
		if err = os.MkdirAll(destDir, srcInfo.Mode()); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	srcDirLen := len(srcDir)

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) (errr error) {
		relSrcPath := strings.TrimLeft(path[srcDirLen:], string(os.PathSeparator))
		destPath := filepath.Join(destDir, relSrcPath)

		if strings.HasPrefix(relSrcPath, ".") {
			if info.IsDir() && (copyFlags&FlagCopyHide != FlagCopyHide) {
				return filepath.SkipDir
			}
		}

		skiped := false
		if copyFilter != nil {
			if skiped, errr = copyFilter(destPath, path, info); errr != nil {
				return errr
			}
		}

		if !skiped {
			if info.IsDir() {
				errr = os.MkdirAll(destPath, info.Mode())
				if !os.IsNotExist(errr) {
					return errr
				}
				return nil
			} else {
				return copyFile(destPath, path)
			}
		}

		return nil
	})
}

func CopyFile(destFile, srcFile string) (err error) {
	if destFile, err = filepath.Abs(destFile); err != nil {
		return err
	}

	if srcFile, err = filepath.Abs(srcFile); err != nil {
		return err
	}

	if err = os.MkdirAll(destFile, 0777); err != nil {
		return err
	}

	return copyFile(destFile, srcFile)
}

func copyFile(destFile, srcFile string) (err error) {
	var dst, src *os.File
	if dst, err = os.Create(destFile); err != nil {
		return err
	}
	defer dst.Close()

	if src, err = os.Open(srcFile); err != nil {
		return err
	}
	defer src.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return
	}

	return nil
}
