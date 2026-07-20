// Package templates provides embedded markdown templates for SDLAIC artifacts.
//
// Templates are embedded at build time using go:embed and can be accessed
// by artifact type name.
package templates

import (
	"embed"
	"fmt"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

//go:embed data/*.md
var templateFS embed.FS

// templateFileNames maps artifact types to their embedded file names.
var templateFileNames = map[domain.ArtifactType]string{
	domain.ArtifactContext:   "data/context.md",
	domain.ArtifactRationale: "data/rationale.md",
	domain.ArtifactProposal: "data/proposal.md",
	domain.ArtifactSpecs:    "data/specs.md",
	domain.ArtifactDesign:   "data/design.md",
	domain.ArtifactTasks:    "data/tasks.md",
}

// GetTemplate returns the template content for the given artifact type.
func GetTemplate(artifactType domain.ArtifactType) (string, error) {
	fileName, ok := templateFileNames[artifactType]
	if !ok {
		return "", fmt.Errorf("no template for artifact type %q", artifactType)
	}

	data, err := templateFS.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("reading template %s: %w", fileName, err)
	}

	return string(data), nil
}

// GetTemplateByName returns the template content for a given artifact type name string.
func GetTemplateByName(name string) (string, error) {
	artifactType, err := domain.ParseArtifactType(name)
	if err != nil {
		return "", fmt.Errorf("invalid template name %q: %w", name, err)
	}
	return GetTemplate(artifactType)
}

// ListTemplates returns the names of all available templates.
func ListTemplates() []string {
	types := domain.OrderedArtifactTypes()
	names := make([]string, len(types))
	for i, t := range types {
		names[i] = string(t)
	}
	return names
}
