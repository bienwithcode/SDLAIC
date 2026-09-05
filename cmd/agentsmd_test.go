package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureAgentsMdBlock_CreatesFile(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, ensureAgentsMdBlock(root))

	content, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	require.NoError(t, err)
	assert.Equal(t, agentsMdBlock+"\n", string(content))
	assert.Contains(t, string(content), agentsMdBeginMarker)
	assert.Contains(t, string(content), agentsMdEndMarker)
}

func TestEnsureAgentsMdBlock_PreservesUserContent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "AGENTS.md")
	require.NoError(t, os.WriteFile(path, []byte("# My project\n\nHouse rules here.\n"), 0644))

	require.NoError(t, ensureAgentsMdBlock(root))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(content), "# My project\n\nHouse rules here.\n\n<!-- sdlaic:begin -->"))
	assert.Contains(t, string(content), agentsMdEndMarker)
}

func TestEnsureAgentsMdBlock_ReplacesStaleBlock(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "AGENTS.md")
	stale := "# My project\n\n" + agentsMdBeginMarker + "\nold instructions\n" + agentsMdEndMarker + "\n\n# Trailing section\n"
	require.NoError(t, os.WriteFile(path, []byte(stale), 0644))

	require.NoError(t, ensureAgentsMdBlock(root))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "old instructions")
	assert.Contains(t, string(content), agentsMdBlock)
	assert.True(t, strings.HasSuffix(string(content), agentsMdEndMarker+"\n\n# Trailing section\n"))
}

func TestEnsureAgentsMdBlock_Idempotent(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, ensureAgentsMdBlock(root))
	first, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	require.NoError(t, err)

	require.NoError(t, ensureAgentsMdBlock(root))
	second, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	require.NoError(t, err)

	assert.Equal(t, string(first), string(second), "re-running must be a no-op diff")
}

func TestChangesDirInsideProject(t *testing.T) {
	tests := []struct {
		name       string
		changesDir string
		root       string
		want       bool
	}{
		{name: "default inside", changesDir: filepath.Join("/proj", ".sdlaic", "changes"), root: "/proj", want: true},
		{name: "root itself", changesDir: "/proj", root: "/proj", want: true},
		{name: "sibling outside", changesDir: "/elsewhere/changes", root: "/proj", want: false},
		{name: "parent outside", changesDir: "/changes", root: "/proj/nested", want: false},
		{name: "dotted prefix inside", changesDir: "/project2/changes", root: "/proj", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, changesDirInsideProject(tt.changesDir, tt.root))
		})
	}
}
