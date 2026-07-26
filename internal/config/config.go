// Package config handles loading, saving, and validating SDLAIC configuration files.
//
// All state lives in one place: GlobalConfig (config.json) in ~/.sdlaic/.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// LoadGlobal reads and parses a global config.json from the given path.
// It validates the parsed config and returns an error for invalid values.
func LoadGlobal(path string) (domain.GlobalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.GlobalConfig{}, fmt.Errorf("reading global config %s: %w", path, err)
	}

	var cfg domain.GlobalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return domain.GlobalConfig{}, fmt.Errorf("parsing global config %s: %w", path, err)
	}

	if err := ValidateGlobal(cfg); err != nil {
		return domain.GlobalConfig{}, fmt.Errorf("validating global config %s: %w", path, err)
	}

	return cfg, nil
}

// SaveGlobal writes a GlobalConfig to the given path as formatted JSON.
// It creates parent directories if they don't exist.
func SaveGlobal(cfg domain.GlobalConfig, path string) error {
	if err := ValidateGlobal(cfg); err != nil {
		return fmt.Errorf("validating global config: %w", err)
	}

	if err := ensureDir(path); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Ensure projects map is initialized (avoid null in JSON)
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]domain.ProjectEntry)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling global config: %w", err)
	}

	if err := writeFileAtomic(path, data); err != nil {
		return fmt.Errorf("writing global config %s: %w", path, err)
	}

	return nil
}

// LoadOrCreateGlobal reads the global config, returning defaults when the file
// does not exist yet. Any other failure — including a corrupt file — is
// returned, so a malformed config is never silently replaced with defaults.
func LoadOrCreateGlobal(path string) (domain.GlobalConfig, error) {
	cfg, err := LoadGlobal(path)
	if errors.Is(err, os.ErrNotExist) {
		return domain.NewGlobalConfig(), nil
	}
	if err != nil {
		return domain.GlobalConfig{}, err
	}
	return cfg, nil
}

// UpdateProject applies mutate to a single project entry and writes the config
// back. Only the named entry is touched, so a concurrent write from another
// project is limited to a lost update on that project rather than on the file
// as a whole. A file written by an older schema is upgraded on the way out.
func UpdateProject(path string, hash string, mutate func(*domain.ProjectEntry)) error {
	cfg, err := LoadOrCreateGlobal(path)
	if err != nil {
		return fmt.Errorf("loading global config %s: %w", path, err)
	}

	if cfg.Projects == nil {
		cfg.Projects = make(map[string]domain.ProjectEntry)
	}
	entry := cfg.Projects[hash]
	mutate(&entry)
	cfg.Projects[hash] = entry
	cfg.Version = domain.GlobalConfigVersion

	if err := SaveGlobal(cfg, path); err != nil {
		return fmt.Errorf("saving global config %s: %w", path, err)
	}
	return nil
}

// writeFileAtomic writes data to a temp file in the target directory and renames
// it into place, so a failed or interrupted write cannot leave a partial config
// behind. Mirrors the approach in internal/gatestate.
func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file %s: %w", tmpName, err)
	}
	if err := os.Chmod(tmpName, 0644); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("setting permissions on %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming %s to %s: %w", tmpName, path, err)
	}
	return nil
}

// ValidateGlobal checks that a GlobalConfig has valid field values.
func ValidateGlobal(cfg domain.GlobalConfig) error {
	if cfg.DefaultWorkflow != "" {
		if _, err := domain.ParseWorkflowLevel(string(cfg.DefaultWorkflow)); err != nil {
			return fmt.Errorf("invalid default_workflow %q: %w", cfg.DefaultWorkflow, err)
		}
	}
	return nil
}

// ensureDir creates the parent directory for a file path if it doesn't exist.
func ensureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0755)
}
