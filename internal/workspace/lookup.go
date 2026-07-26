package workspace

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// Lookup returns the hash and entry of the registered project that dir belongs
// to, or ErrWorkspaceNotFound when it belongs to none.
//
// SDLAIC keeps no marker file in the project, so this replaces walking up the
// tree: dir is matched against the project paths in the global config and the
// longest match wins, which reproduces the "nearest workspace" behaviour for
// nested and monorepo layouts. Entries whose path no longer exists on disk are
// skipped rather than matched.
func Lookup(cfg domain.GlobalConfig, dir string) (string, domain.ProjectEntry, error) {
	target, err := resolveDir(dir)
	if err != nil {
		return "", domain.ProjectEntry{}, fmt.Errorf("resolving directory %s: %w", dir, err)
	}

	var bestHash string
	var best domain.ProjectEntry
	bestLen := -1

	for hash, entry := range cfg.Projects {
		root, err := resolveDir(entry.Path)
		if err != nil {
			continue // stale entry: the project directory is gone
		}
		if !isWithin(target, root) {
			continue
		}
		if len(root) > bestLen {
			bestLen, bestHash, best = len(root), hash, entry
		}
	}

	if bestHash == "" {
		return "", domain.ProjectEntry{}, domain.ErrWorkspaceNotFound
	}
	return bestHash, best, nil
}

// resolveDir returns an absolute, symlink-free form of path. Both sides of every
// comparison go through it: on macOS a temp dir reported as /var/... resolves to
// /private/var/..., and comparing the two forms would silently never match.
func resolveDir(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

// isWithin reports whether dir is root or sits underneath it. The comparison is
// segment-aware via filepath.Rel, so /a/foobar is not inside /a/foo.
func isWithin(dir, root string) bool {
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return false
	}
	return rel == "." || !strings.HasPrefix(rel, "..")
}
