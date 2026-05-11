package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	source, err := os.Open(fromPath)
	if err != nil {
		return fmt.Errorf("cannot open source file: %w", err)
	}
	defer source.Close()

	info, err := source.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat source file: %w", err)
	}

	if !info.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	if offset > info.Size() {
		return ErrOffsetExceedsFileSize
	}

	if _, err := source.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("cannot seek: %w", err)
	}

	copySize := info.Size() - offset
	if limit > 0 && limit < copySize {
		copySize = limit
	}

	if copySize <= 0 && limit > 0 {
		return nil
	}

	dest, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("cannot create destination file: %w", err)
	}
	defer dest.Close()

	return copyWithProgress(source, dest, copySize)
}

func copyWithProgress(src io.Reader, dst io.Writer, total int64) error {
	if total <= 0 {
		_, err := io.Copy(dst, src)
		return err
	}

	var current int64
	buf := make([]byte, 32*1024)

	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			toWrite := int64(n)
			if current+toWrite > total {
				toWrite = total - current
			}

			if _, err := dst.Write(buf[:toWrite]); err != nil {
				return err
			}

			current += toWrite
			percent := float64(current) / float64(total) * 100
			fmt.Printf("\rCopying: %.2f%%", percent)

			if current >= total {
				fmt.Println()
				return nil
			}
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				if current > 0 {
					fmt.Println()
				}
				return nil
			}
			return readErr
		}
	}
}
