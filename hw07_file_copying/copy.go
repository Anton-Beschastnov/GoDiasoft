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
		return fmt.Errorf("open source file: %w", err)
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return fmt.Errorf("get file info: %w", err)
	}

	if !info.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	fileSize := info.Size()

	if offset < 0 {
		return fmt.Errorf("offset cannot be negative: %d", offset)
	}
	if offset > fileSize {
		return ErrOffsetExceedsFileSize
	}

	if _, err := src.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("seek to offset: %w", err)
	}

	copySize := fileSize - offset
	if limit > 0 && limit < copySize {
		copySize = limit
	}

	if copySize <= 0 {
		dst, err := os.Create(toPath)
		if err != nil {
			return fmt.Errorf("create destination file: %w", err)
		}
		defer dst.Close()
		return nil
	}

	dst, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer dst.Close()

	progressReader := &ProgressReader{
		reader: src,
		total:  copySize,
	}

	_, err = io.CopyN(dst, progressReader, copySize)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("copy data: %w", err)
	}

	fmt.Println()

	return nil
}

type ProgressReader struct {
	reader  io.Reader
	total   int64
	current int64
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	if pr.current >= pr.total {
		return 0, io.EOF
	}

	remaining := pr.total - pr.current
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}

	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.current += int64(n)
		percent := float64(pr.current) / float64(pr.total) * 100
		fmt.Printf("\rCopying: %.2f%%", percent)
	}

	return n, err
}
