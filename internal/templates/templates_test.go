package templates

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

func TestGetTemplate_AllTypes(t *testing.T) {
	tests := []struct {
		artifactType domain.ArtifactType
		fileName     string
	}{
		{domain.ArtifactContext, "context.md"},
		{domain.ArtifactProposal, "proposal.md"},
		{domain.ArtifactSpec, "spec.md"},
		{domain.ArtifactDesign, "design.md"},
		{domain.ArtifactTasks, "tasks.md"},
	}

	for _, tt := range tests {
		t.Run(string(tt.artifactType), func(t *testing.T) {
			content, err := GetTemplate(tt.artifactType)
			require.NoError(t, err)
			assert.NotEmpty(t, content, "template content should not be empty")
		})
	}
}

func TestGetTemplate_InvalidType(t *testing.T) {
	_, err := GetTemplate(domain.ArtifactType("invalid"))
	assert.Error(t, err)
}

func TestGetTemplateByName(t *testing.T) {
	tests := []string{"context", "proposal", "spec", "design", "tasks"}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			content, err := GetTemplateByName(name)
			require.NoError(t, err)
			assert.NotEmpty(t, content)
		})
	}
}

func TestGetTemplateByName_InvalidName(t *testing.T) {
	_, err := GetTemplateByName("nonexistent")
	assert.Error(t, err)
}

func TestListTemplates(t *testing.T) {
	templates := ListTemplates()
	assert.Len(t, templates, 5)

	names := make(map[string]bool)
	for _, tmpl := range templates {
		names[tmpl] = true
	}

	assert.True(t, names["context"])
	assert.True(t, names["proposal"])
	assert.True(t, names["spec"])
	assert.True(t, names["design"])
	assert.True(t, names["tasks"])
	assert.False(t, names["rationale"], "rationale template removed")
	assert.False(t, names["specs"], "specs renamed to spec")
}

func TestTemplateContent_HasMarkdownStructure(t *testing.T) {
	// Each template should be valid markdown with at least a heading
	types := domain.OrderedArtifactTypes()
	for _, at := range types {
		t.Run(string(at), func(t *testing.T) {
			content, err := GetTemplate(at)
			require.NoError(t, err)
			// Templates should contain at least one markdown heading
			assert.True(t, strings.Contains(content, "#"),
				"template %s should contain markdown headings", at)
		})
	}
}

func TestTemplateContent_NoPlaceholders(t *testing.T) {
	// Template files themselves should not contain mustache-style placeholders
	// (those would indicate incomplete template work)
	types := domain.OrderedArtifactTypes()
	for _, at := range types {
		t.Run(string(at), func(t *testing.T) {
			content, err := GetTemplate(at)
			require.NoError(t, err)
			// Should not contain {{...}} placeholder syntax
			assert.False(t, strings.Contains(content, "{{"),
				"template %s should not contain {{ placeholders", at)
		})
	}
}
