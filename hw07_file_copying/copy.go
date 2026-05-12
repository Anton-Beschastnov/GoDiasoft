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
	src, err := os.Open(fromPath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	if !info.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	if offset > info.Size() {
		return ErrOffsetExceedsFileSize
	}

	if _, err := src.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("seek: %w", err)
	}

	copySize := info.Size() - offset
	if limit > 0 && limit < copySize {
		copySize = limit
	}

	if copySize <= 0 {
		dst, err := os.Create(toPath)
		if err != nil {
			return fmt.Errorf("create destination: %w", err)
		}
		defer dst.Close()
		return nil
	}

	dst, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer dst.Close()

	return copyWithProgress(src, dst, copySize)
}

func copyWithProgress(src io.Reader, dst io.Writer, total int64) error {
	var current int64
	buf := make([]byte, 32*1024)

	for current < total {
		remaining := total - current
		toRead := int64(len(buf))
		if toRead > remaining {
			toRead = remaining
		}

		n, err := src.Read(buf[:toRead])
		if n > 0 {
			if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			current += int64(n)
			fmt.Printf("\rCopying: %.2f%%", float64(current)/float64(total)*100)
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println()
				return nil
			}
			return err
		}
	}

	fmt.Println()
	return nil
}
