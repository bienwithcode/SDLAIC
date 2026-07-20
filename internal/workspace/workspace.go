// Package workspace handles discovering and initializing SDLAIC workspaces.
//
// A workspace is any directory containing a .sdlaicrc config file.
// FindWorkspace walks up from the current directory to locate the nearest one.
package workspace

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bienwithcode/SDLAIC/internal/config"
	"github.com/bienwithcode/SDLAIC/internal/domain"
)

const configFile = ".sdlaicrc"

// FindWorkspace walks up from startDir to find the nearest .sdlaicrc file.
// Returns the directory containing the config file, or ErrWorkspaceNotFound.
func FindWorkspace(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path: %w", err)
	}

	for {
		cfgPath := filepath.Join(dir, configFile)
		if _, err := os.Stat(cfgPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", domain.ErrWorkspaceNotFound
		}
		dir = parent
	}
}

// InitWorkspace creates a new SDLAIC workspace in the given directory.
// It writes a .sdlaicrc config file and creates the appropriate directory
// structure based on the storage mode.
func InitWorkspace(projectDir string, storage domain.StorageMode, workflow domain.WorkflowLevel, projectHash string) (domain.LocalConfig, error) {
	return initWorkspace(projectDir, defaultHomeDir(), storage, workflow, projectHash)
}

// InitWorkspaceWithHome is like InitWorkspace but accepts an explicit home directory.
// Useful for testing where the real home directory shouldn't be used.
func InitWorkspaceWithHome(projectDir string, homeDir string, storage domain.StorageMode, workflow domain.WorkflowLevel, projectHash string) (domain.LocalConfig, error) {
	return initWorkspace(projectDir, homeDir, storage, workflow, projectHash)
}

func initWorkspace(projectDir string, homeDir string, storage domain.StorageMode, workflow domain.WorkflowLevel, projectHash string) (domain.LocalConfig, error) {
	// Check if workspace already exists
	cfgPath := filepath.Join(projectDir, configFile)
	if _, err := os.Stat(cfgPath); err == nil {
		return domain.LocalConfig{}, domain.ErrWorkspaceExists
	}

	cfg := domain.NewLocalConfig(storage, workflow, projectHash)

	// Write config file
	if err := config.SaveLocal(cfg, cfgPath); err != nil {
		return domain.LocalConfig{}, fmt.Errorf("saving .sdlaicrc: %w", err)
	}

	// Create directory structure based on storage mode
	switch storage {
	case domain.StorageModeLocal, domain.StorageModeIgnored:
		changesDir := filepath.Join(projectDir, ".sdlaic", "changes")
		if err := os.MkdirAll(changesDir, 0755); err != nil {
			return domain.LocalConfig{}, fmt.Errorf("creating changes directory: %w", err)
		}
	case domain.StorageModeGlobal:
		storeDir := filepath.Join(homeDir, ".sdlaic", "stores", projectHash, "changes")
		if err := os.MkdirAll(storeDir, 0755); err != nil {
			return domain.LocalConfig{}, fmt.Errorf("creating global store directory: %w", err)
		}
	}

	return cfg, nil
}

// ProjectHash computes a short, deterministic hash for a project directory.
// The hash is based on the absolute path of the directory.
func ProjectHash(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path: %w", err)
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
