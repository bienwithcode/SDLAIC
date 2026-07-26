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
