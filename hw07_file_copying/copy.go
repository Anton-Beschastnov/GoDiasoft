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
	return prepareAndCopy(src, toPath, offset, limit, info.Size())
}

func prepareAndCopy(src *os.File, toPath string, off, lim, size int64) error {
	if _, err := src.Seek(off, io.SeekStart); err != nil {
		return fmt.Errorf("seek: %w", err)
	}
	copySize := size - off
	if lim > 0 && lim < copySize {
		copySize = lim
	}
	if copySize <= 0 && lim > 0 {
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
	if total <= 0 {
		_, err := io.Copy(dst, src)
		return err
	}
	var current int64
	buf := make([]byte, 32*1024)
	for {
		n, rErr := src.Read(buf)
		if n > 0 {
			if err := writeChunk(dst, buf[:n], &current, total); err != nil {
				return err
			}
			if current >= total {
				fmt.Println()
				return nil
			}
		}
		if rErr != nil {
			return handleReadError(rErr, current)
		}
	}
}

func writeChunk(dst io.Writer, b []byte, current *int64, total int64) error {
	toWrite := int64(len(b))
	if *current+toWrite > total {
		toWrite = total - *current
	}
	if _, err := dst.Write(b[:toWrite]); err != nil {
		return err
	}
	*current += toWrite
	fmt.Printf("\rCopying: %.2f%%", float64(*current)/float64(total)*100)
	return nil
}

func handleReadError(err error, current int64) error {
	if errors.Is(err, io.EOF) {
		if current > 0 {
			fmt.Println()
		}
		return nil
	}
	return err
}
