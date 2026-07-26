// Package storage handles path resolution and gitignore management
// for SDLAIC change artifacts.
package storage

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// DefaultChangesDir returns the in-project location used when a project does
// not override it: <project-root>/.sdlaic/changes.
func DefaultChangesDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".sdlaic", "changes")
}

// NormalizeChangesDir turns user input into the absolute, cleaned form that gets
// persisted. A leading ~ is expanded against home and a relative path is
// resolved against cwd.
//
// Normalisation happens once, at the moment the value is set — never lazily at
// read time — so a stored path cannot change meaning when a later command runs
// from a different directory.
func NormalizeChangesDir(input string, cwd string, home string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", fmt.Errorf("changes directory must not be empty")
	}

	if trimmed == "~" {
		trimmed = home
	} else if strings.HasPrefix(trimmed, "~"+string(filepath.Separator)) {
		trimmed = filepath.Join(home, trimmed[2:])
	}

	if !filepath.IsAbs(trimmed) {
		trimmed = filepath.Join(cwd, trimmed)
	}
	return filepath.Clean(trimmed), nil
}

// ChangesBase returns the cleaned base directory holding all changes, or
// ErrChangesDirNotSet when the project has no configured location yet.
func ChangesBase(changesDir string) (string, error) {
	if strings.TrimSpace(changesDir) == "" {
		return "", domain.ErrChangesDirNotSet
	}
	return filepath.Clean(changesDir), nil
}

// ChangePath returns the directory of a single change inside changesDir.
func ChangePath(changesDir string, changeName string) (string, error) {
	base, err := ChangesBase(changesDir)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(changeName) == "" {
		return "", fmt.Errorf("change name must not be empty")
	}
	return filepath.Join(base, changeName), nil
}

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
