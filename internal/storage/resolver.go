// Package storage resolves the on-disk locations of SDLAIC change artifacts
// from a project's configured changes directory.
package storage

import (
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
