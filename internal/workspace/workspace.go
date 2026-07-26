// Package workspace identifies which registered project a directory belongs to.
//
// SDLAIC keeps no marker file in the project: Lookup matches the working
// directory against the projects registered in ~/.sdlaic/config.json.
package workspace

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

// ProjectHash computes a short, deterministic hash for a project directory.
//
// The hash is based on the absolute, symlink-free path, matching how Lookup
// compares paths. Without that agreement the same project reached through a
// symlink would hash differently and register twice — on macOS a temp dir is
// reported as /var/... but resolves to /private/var/.... Directories that do not
// exist yet fall back to the plain absolute path.
func ProjectHash(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	h := sha256.Sum256([]byte(abs))
	return fmt.Sprintf("%x", h)[:12], nil
}

// defaultHomeDir returns the user's home directory.
func defaultHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp"
	}
	return home
}
