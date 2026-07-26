package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bienwithcode/SDLAIC/internal/config"
	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/storage"
	"github.com/bienwithcode/SDLAIC/internal/workspace"
)

// projectContext is everything a command needs to know about the project it is
// running in. It is the single place state is read from, so commands never care
// whether that state came from the global config or the legacy local file.
type projectContext struct {
	Root         string
	Hash         string
	ChangesDir   string
	Workflow     domain.WorkflowLevel
	ActiveChange string

	// fromLocalConfig marks a project that has not been migrated yet and is
	// still being served from .sdlaicrc.
	// TEMPORARY: removed in T17 along with the fallback itself.
	fromLocalConfig bool
}

// resolveProject determines the project containing the current directory.
//
// The global config is authoritative. A project with no entry there falls back
// to the legacy .sdlaicrc so the binary keeps working while commands are
// migrated one at a time.
// TEMPORARY: the fallback is removed in T17.
func resolveProject() (projectContext, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return projectContext{}, fmt.Errorf("getting current directory: %w", err)
	}

	cfg, err := config.LoadOrCreateGlobal(globalConfigPath())
	if err != nil {
		return projectContext{}, fmt.Errorf("loading global config: %w", err)
	}

	hash, entry, err := workspace.Lookup(cfg, cwd)
	if err == nil {
		workflowLevel := entry.Workflow
		if workflowLevel == "" {
			workflowLevel = cfg.DefaultWorkflow
		}
		return projectContext{
			Root:         entry.Path,
			Hash:         hash,
			ChangesDir:   entry.ChangesDir,
			Workflow:     workflowLevel,
			ActiveChange: entry.ActiveChange,
		}, nil
	}
	if !errors.Is(err, domain.ErrWorkspaceNotFound) {
		return projectContext{}, err
	}

	return resolveProjectFromLocalConfig(cwd)
}

// resolveProjectFromLocalConfig serves a project that still only has a
// .sdlaicrc, deriving ChangesDir from its storage mode.
// TEMPORARY: removed in T17.
func resolveProjectFromLocalConfig(cwd string) (projectContext, error) {
	root, err := workspace.FindWorkspace(cwd)
	if err != nil {
		return projectContext{}, err
	}

	local, err := config.LoadLocal(filepath.Join(root, ".sdlaicrc"))
	if err != nil {
		return projectContext{}, fmt.Errorf("loading local config: %w", err)
	}

	changesDir, err := storage.ChangesBasePath(local.Storage, root, resolveHome())
	if err != nil {
		return projectContext{}, err
	}

	return projectContext{
		Root:            root,
		Hash:            local.ProjectHash,
		ChangesDir:      changesDir,
		Workflow:        local.Workflow,
		ActiveChange:    local.ActiveChange,
		fromLocalConfig: true,
	}, nil
}

// changesDir returns the configured changes directory, or ErrChangesDirNotSet
// when the project has not been configured. Callers prompt rather than guess.
func (p projectContext) changesDir() (string, error) {
	return storage.ChangesBase(p.ChangesDir)
}

// changePath returns the directory of a single change in this project.
func (p projectContext) changePath(changeName string) (string, error) {
	return storage.ChangePath(p.ChangesDir, changeName)
}

// resolveChange returns the explicit change name if given, otherwise the
// project's active change.
func (p projectContext) resolveChange(flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if p.ActiveChange != "" {
		return p.ActiveChange, nil
	}
	return "", domain.ErrNoActiveChange
}

// setActiveChange persists the active change, writing back to whichever source
// this context was resolved from. Pass an empty string to clear it.
func (p projectContext) setActiveChange(changeName string) error {
	if p.fromLocalConfig {
		// TEMPORARY: removed in T17.
		return config.SetActiveChange(filepath.Join(p.Root, ".sdlaicrc"), changeName)
	}
	return config.UpdateProject(globalConfigPath(), p.Hash, func(e *domain.ProjectEntry) {
		e.ActiveChange = changeName
	})
}

// workspaceHash is the project-hash function used for both the global config
// key and the gate-state directory, so the two always agree.
func workspaceHash(dir string) (string, error) {
	return workspace.ProjectHash(dir)
}

// registerProjectEntry writes a project's configuration to the global config,
// after refusing a changes directory another project already owns. Shared by
// init and open so the two cannot drift apart.
func registerProjectEntry(cwd string, changesDir string, workflowLevel domain.WorkflowLevel) error {
	hash, err := workspace.ProjectHash(cwd)
	if err != nil {
		return fmt.Errorf("computing project hash: %w", err)
	}

	cfgPath := globalConfigPath()
	if err := ensureChangesDirUnclaimed(cfgPath, hash, changesDir); err != nil {
		return err
	}

	if err := os.MkdirAll(changesDir, 0755); err != nil {
		return fmt.Errorf("creating changes directory: %w", err)
	}

	if err := config.UpdateProject(cfgPath, hash, func(e *domain.ProjectEntry) {
		e.Path = cwd
		e.ChangesDir = changesDir
		e.Workflow = workflowLevel
	}); err != nil {
		return fmt.Errorf("registering project: %w", err)
	}

	// TEMPORARY: commands not yet migrated still read .sdlaicrc. Only written for
	// the in-project default, so an external changes dir stays zero-footprint.
	// Removed in T17.
	if changesDir == storage.DefaultChangesDir(cwd) {
		local := domain.NewLocalConfig(domain.StorageModeLocal, workflowLevel, hash)
		if err := config.SaveLocal(local, filepath.Join(cwd, ".sdlaicrc")); err != nil {
			return fmt.Errorf("writing legacy local config: %w", err)
		}
	}
	return nil
}

// ensureChangesDirUnclaimed rejects a directory already registered to a
// different project. One changes directory belongs to exactly one project:
// sharing one would make `list` mix projects, let `archive` overwrite another
// project's tarball, and leave the same change carrying a different gate state
// depending on which project you stand in.
func ensureChangesDirUnclaimed(cfgPath string, hash string, changesDir string) error {
	cfg, err := config.LoadOrCreateGlobal(cfgPath)
	if err != nil {
		return fmt.Errorf("loading global config: %w", err)
	}
	for otherHash, entry := range cfg.Projects {
		if otherHash == hash || entry.ChangesDir != changesDir {
			continue
		}
		return fmt.Errorf("changes directory %s is already used by project %s", changesDir, entry.Path)
	}
	return nil
}
