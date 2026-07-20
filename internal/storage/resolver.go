// Package storage handles path resolution and gitignore management
// for SDLAIC change artifacts.
package storage

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// ResolvePath returns the full path to a change directory based on the
// storage mode, project root, optional home directory, and change name.
func ResolvePath(mode domain.StorageMode, projectRoot string, homeDir string, changeName string) (string, error) {
	if changeName == "" {
		return "", fmt.Errorf("change name must not be empty")
	}

	base, err := ChangesBasePath(mode, projectRoot, homeDir)
	if err != nil {
		return "", err
	}

	return filepath.Join(base, changeName), nil
}

// ChangesBasePath returns the base directory for all changes given a storage mode.
func ChangesBasePath(mode domain.StorageMode, projectRoot string, homeDir string) (string, error) {
	switch mode {
	case domain.StorageModeLocal, domain.StorageModeIgnored:
		return filepath.Join(projectRoot, ".sdlaic", "changes"), nil
	case domain.StorageModeGlobal:
		hash := ComputeProjectHash(projectRoot)
		return filepath.Join(homeDir, ".sdlaic", "stores", hash, "changes"), nil
	default:
		return "", fmt.Errorf("unsupported storage mode %q", mode)
	}
}

// ComputeProjectHash returns a short deterministic hash of the project root path.
func ComputeProjectHash(projectRoot string) string {
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		abs = projectRoot
	}
	h := sha256.Sum256([]byte(abs))
	return fmt.Sprintf("%x", h)[:12]
}

// ListChanges returns the names of all change directories in the given base path.
func ListChanges(basePath string) ([]string, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading changes directory %s: %w", basePath, err)
	}

	var changes []string
	for _, entry := range entries {
		if entry.IsDir() {
			changes = append(changes, entry.Name())
		}
	}
	return changes, nil
}
