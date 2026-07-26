// Package config handles loading, saving, and validating SDLAIC configuration files.
//
// Two config types exist:
//   - LocalConfig (.sdlaicrc) — per-project, lives in the project root
//   - GlobalConfig (config.json) — per-user, lives in ~/.sdlaic/
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// LoadLocal reads and parses a .sdlaicrc file from the given path.
// It validates the parsed config and returns an error for invalid values.
func LoadLocal(path string) (domain.LocalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.LocalConfig{}, fmt.Errorf("reading local config %s: %w", path, err)
	}

	var cfg domain.LocalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return domain.LocalConfig{}, fmt.Errorf("parsing local config %s: %w", path, err)
	}

	if err := ValidateLocal(cfg); err != nil {
		return domain.LocalConfig{}, fmt.Errorf("validating local config %s: %w", path, err)
	}

	return cfg, nil
}

// SaveLocal writes a LocalConfig to the given path as formatted JSON.
// It creates parent directories if they don't exist.
func SaveLocal(cfg domain.LocalConfig, path string) error {
	if err := ValidateLocal(cfg); err != nil {
		return fmt.Errorf("validating local config: %w", err)
	}

	if err := ensureDir(path); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling local config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing local config %s: %w", path, err)
	}

	return nil
}

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

// SetActiveChange loads the local config at the given path, updates the active
// change field, and saves it back. Pass an empty string to clear the active change.
func SetActiveChange(path string, changeName string) error {
	cfg, err := LoadLocal(path)
	if err != nil {
		return fmt.Errorf("loading config to set active change: %w", err)
	}

	cfg.ActiveChange = changeName

	if err := SaveLocal(cfg, path); err != nil {
		return fmt.Errorf("saving config after setting active change: %w", err)
	}

	return nil
}

// MergeDefaults applies global config defaults to a local config where local
// values are zero/empty. Local values always take precedence over global defaults.
func MergeDefaults(local domain.LocalConfig, global domain.GlobalConfig) domain.LocalConfig {
	result := local

	if result.Storage == "" {
		result.Storage = global.DefaultStorage
	}
	if result.Workflow == "" {
		result.Workflow = global.DefaultWorkflow
	}

	return result
}

// ValidateLocal checks that a LocalConfig has valid field values.
func ValidateLocal(cfg domain.LocalConfig) error {
	if cfg.Storage != "" {
		if _, err := domain.ParseStorageMode(string(cfg.Storage)); err != nil {
			return fmt.Errorf("invalid storage mode %q: %w", cfg.Storage, err)
		}
	}
	if cfg.Workflow != "" {
		if _, err := domain.ParseWorkflowLevel(string(cfg.Workflow)); err != nil {
			return fmt.Errorf("invalid workflow level %q: %w", cfg.Workflow, err)
		}
	}
	if cfg.ProjectHash == "" {
		return fmt.Errorf("project_hash is required")
	}
	return nil
}

// ValidateGlobal checks that a GlobalConfig has valid field values.
func ValidateGlobal(cfg domain.GlobalConfig) error {
	if cfg.DefaultStorage != "" {
		if _, err := domain.ParseStorageMode(string(cfg.DefaultStorage)); err != nil {
			return fmt.Errorf("invalid default_storage %q: %w", cfg.DefaultStorage, err)
		}
	}
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
