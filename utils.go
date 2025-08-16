package gonertia

import (
	crypto "crypto/md5"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
)

func setOf[T comparable](data []T) map[T]struct{} {
	if len(data) == 0 {
		return nil
	}

	set := make(map[T]struct{}, len(data))
	for _, v := range data {
		set[v] = struct{}{}
	}

	return set
}

func firstOr[T any](items []T, fallback T) T {
	if len(items) > 0 {
		return items[0]
	}
	return fallback
}

func md5(str string) string {
	hash := crypto.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

func md5FileFromFileFS(file fs.File) (string, error) {
	hash := crypto.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func md5FileFromFS(fs fs.FS, path string) (string, error) {
	f, err := fs.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	return md5FileFromFileFS(f)
}

func md5File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	return md5FileFromFileFS(f)
}
