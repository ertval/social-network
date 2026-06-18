package localstorage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	dirPermission  = 0o750
	filePermission = 0o640
	savingDir      = "frontend/static/images/uploads"
)

var (
	ErrPathIsEmpty = errors.New("path is required")
	ErrInvalidPath = errors.New("invalid storage path")
)

type LocalStorage struct {
	// we will maybe need the base path as a dependency
	// and the const above would be passed here as dependencies from the config
	// but ill leave it for later
}

func NewLocalStorage() *LocalStorage {
	return &LocalStorage{}
}

func (s *LocalStorage) Upload(ctx context.Context, file []byte, filename string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fileName, err := cleanPath(filename)
	if err != nil {
		return err
	}

	storagePath := filepath.Join(savingDir, fileName)

	if err = os.MkdirAll(filepath.Dir(storagePath), dirPermission); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}

	if err = ctx.Err(); err != nil {
		return err
	}

	if err = os.WriteFile(storagePath, file, filePermission); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func (s *LocalStorage) Delete(ctx context.Context, filename string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fileName, err := cleanPath(filename)
	if err != nil {
		return err
	}

	storagePath := filepath.Join(savingDir, fileName)

	err = os.Remove(storagePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}

func cleanPath(path string) (string, error) {
	if path == "" {
		return "", ErrPathIsEmpty
	}

	storagePath := filepath.Clean(path)
	if storagePath == "." || storagePath == string(filepath.Separator) {
		return "", ErrInvalidPath
	}

	if filepath.IsAbs(storagePath) {
		return "", ErrInvalidPath
	}

	if storagePath == ".." || strings.HasPrefix(storagePath, ".."+string(filepath.Separator)) {
		return "", ErrInvalidPath
	}

	return storagePath, nil
}
