package storage

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FS struct {
	baseDir string
	prefix  string
}

func NewFS() *FS {
	return &FS{baseDir: strings.TrimRight("data", string(os.PathSeparator)), prefix: strings.Trim("photos", "/")}
}
func (f *FS) OnStart(_ context.Context) error {
	return nil
}

func (f *FS) OnStop(_ context.Context) error {
	return nil
}

func (f *FS) Save(ctx context.Context, r io.Reader, ext string) (string, error) {
	if f == nil || f.baseDir == "" {
		return "", errors.New("storage FS is not configured")
	}
	ext = normalizeExt(ext)

	absDir := filepath.Join(f.baseDir, f.prefix)

	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return "", err
	}

	filename := uuid.New().String() + ext
	tmpPath := filepath.Join(absDir, filename+".tmp")
	finalPath := filepath.Join(absDir, filename)

	dst, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return "", err
	}
	copyErr := copyWithContext(ctx, dst, r)
	closeErr := dst.Close()

	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return "", copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return "", closeErr
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}

	relPath := filepath.ToSlash(filepath.Join(f.prefix, filename))
	return relPath, nil
}

func normalizeExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		return ".bin"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	switch ext {
	case ".jpeg":
		return ".jpg"
	case ".jpg", ".png", ".webp", ".bin":
		return ext
	default:

		return ext
	}
}

func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader) error {
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		n, rerr := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if rerr != nil {
			if errors.Is(rerr, io.EOF) {
				return nil
			}
			return rerr
		}
	}
}

func (f *FS) Remove(ctx context.Context, rel string) error {
	if f == nil || f.baseDir == "" {
		return errors.New("storage FS is not configured")
	}
	// защита от traversal
	if strings.Contains(rel, "..") || strings.HasPrefix(rel, "/") {
		return errors.New("invalid relative path")
	}
	abs := filepath.Join(f.baseDir, filepath.FromSlash(rel))

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := os.Remove(abs); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
