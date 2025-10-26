package base

import (
	"path/filepath"
	"strings"
)

func IsAllowedPhotoType(t string) bool {
	t = strings.ToLower(strings.TrimSpace(t))
	switch t {
	case "original", "generated", "send", "demo":
		return true
	default:
		return false
	}
}

func IsAllowedVideoType(t string) bool {
	t = strings.ToLower(strings.TrimSpace(t))
	switch t {
	case "send", "demo":
		return true
	default:
		return false
	}
}

func NormalizeVideoType(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	if IsAllowedVideoType(t) {
		return t
	}
	return "send"
}

func canonicalVideoExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	switch ext {
	case ".mp4", ".mov", ".avi", ".mkv", ".webm", ".m4v", ".mpg", ".mpeg", ".3gp":
		return ext
	}
	return ""
}

func NormalizeVideoExt(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return ".bin"
	}

	if idx := strings.IndexAny(name, "?#"); idx >= 0 {
		name = strings.TrimSpace(name[:idx])
	}

	if ext := canonicalVideoExt(filepath.Ext(name)); ext != "" {
		return ext
	}

	if !strings.ContainsAny(name, "./") {
		if ext := canonicalVideoExt("." + name); ext != "" {
			return ext
		}
	}

	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		subtype := strings.TrimSpace(parts[len(parts)-1])
		if idx := strings.Index(subtype, ";"); idx >= 0 {
			subtype = strings.TrimSpace(subtype[:idx])
		}
		if subtype != "" {
			if ext := canonicalVideoExt("." + strings.TrimPrefix(subtype, ".")); ext != "" {
				return ext
			}
		}
	}

	return ".bin"
}

func ResolveStoragePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return path
	}
	trimmed := strings.TrimLeft(path, "/")
	if trimmed == "" {
		return ""
	}
	cleaned := filepath.Clean(filepath.FromSlash(trimmed))
	if strings.HasPrefix(cleaned, "..") {
		return ""
	}
	return filepath.Join("data", cleaned)
}
