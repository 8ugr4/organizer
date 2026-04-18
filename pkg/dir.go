package pkg

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

func DirSize(path string) (int64, error) {
	var size int64
	var mu sync.Mutex

	var calculateSize func(string) error
	calculateSize = func(p string) error {
		fileInfo, err := os.Lstat(p)
		if err != nil {
			return err
		}

		// skip symbolic links to avoid counting them multiple times
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		if fileInfo.IsDir() {
			entries, err := os.ReadDir(p)
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if err := calculateSize(filepath.Join(p, entry.Name())); err != nil {
					return err
				}
			}
		} else {
			mu.Lock()
			size += fileInfo.Size()
			mu.Unlock()
		}

		return nil
	}

	if err := calculateSize(path); err != nil {
		return 0, err
	}

	return size, nil
}

var oneGb = math.Pow(10, 9) //nolint:mnd

func ValidateDir(dirp string) error {
	fp, err := os.Stat(dirp)
	if err != nil {
		return err
	}
	if fp.IsDir() {
		dirSize, err := DirSize(dirp)
		if err != nil {
			return err
		}
		dirSizeGb := float64(dirSize) / oneGb
		slog.Info("", slog.String(dirp, "is a directory"), slog.String("size", fmt.Sprint(dirSizeGb, " bytes")))
	} else {
		return errors.New("path is not dir")
	}

	return nil
}

func createDirectory(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if errCreateDirectory := os.MkdirAll(path, syscall.O_CREAT|syscall.O_EXCL|syscall.O_WRONLY); errCreateDirectory != nil {
				return errCreateDirectory
			}
		} else {
			return err
		}
	}

	return nil
}
