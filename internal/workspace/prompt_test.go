package workspace

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAskChoice_DefaultValue(t *testing.T) {
	in := strings.NewReader("\n")
	var out bytes.Buffer

	prompt := "Choose style"
	options := []string{"classic", "modern", "minimal"}
	defaultVal := "modern"

	val, err := AskChoice(in, &out, prompt, options, defaultVal)
	assert.NoError(t, err)
	assert.Equal(t, "modern", val)
	assert.Contains(t, out.String(), "Choose style [classic/modern/minimal] (default modern): ")
}

func TestAskChoice_ValidChoice(t *testing.T) {
	in := strings.NewReader("classic\n")
	var out bytes.Buffer

	prompt := "Choose style"
	options := []string{"classic", "modern", "minimal"}
	defaultVal := "modern"

	val, err := AskChoice(in, &out, prompt, options, defaultVal)
	assert.NoError(t, err)
	assert.Equal(t, "classic", val)
}

func TestAskChoice_CaseInsensitive(t *testing.T) {
	in := strings.NewReader("MINIMAL\n")
	var out bytes.Buffer

	prompt := "Choose style"
	options := []string{"classic", "modern", "minimal"}
	defaultVal := "modern"

	val, err := AskChoice(in, &out, prompt, options, defaultVal)
	assert.NoError(t, err)
	assert.Equal(t, "minimal", val) // returns original option case
}

func TestAskChoice_InvalidThenValid(t *testing.T) {
	in := strings.NewReader("invalid_choice\nmodern\n")
	var out bytes.Buffer

	prompt := "Choose style"
	options := []string{"classic", "modern", "minimal"}
	defaultVal := "minimal"

	val, err := AskChoice(in, &out, prompt, options, defaultVal)
	assert.NoError(t, err)
	assert.Equal(t, "modern", val)
	assert.Contains(t, out.String(), "Invalid choice: \"invalid_choice\". Please try again.")
}

func TestAskChoice_EOF(t *testing.T) {
	in := strings.NewReader("") // immediate EOF
	var out bytes.Buffer

	prompt := "Choose style"
	options := []string{"classic", "modern", "minimal"}
	defaultVal := "minimal"

	val, err := AskChoice(in, &out, prompt, options, defaultVal)
	assert.NoError(t, err)
	assert.Equal(t, "minimal", val)
}

func TestAskLine_ReturnsTypedValue(t *testing.T) {
	var out bytes.Buffer
	got, err := AskLine(strings.NewReader("/work/openspec/changes\n"), &out, "Changes directory", "")
	require.NoError(t, err)
	assert.Equal(t, "/work/openspec/changes", got)
}

func TestAskLine_EmptyInputUsesDefault(t *testing.T) {
	var out bytes.Buffer
	got, err := AskLine(strings.NewReader("\n"), &out, "Changes directory", "/default/changes")
	require.NoError(t, err)
	assert.Equal(t, "/default/changes", got)
}

func TestAskLine_EOFUsesDefault(t *testing.T) {
	var out bytes.Buffer
	got, err := AskLine(strings.NewReader(""), &out, "Changes directory", "/default/changes")
	require.NoError(t, err)
	assert.Equal(t, "/default/changes", got)
}

func TestPrompter_SharesOneScannerAcrossQuestions(t *testing.T) {
	// A scanner reads ahead. With one scanner per question, the first would
	// swallow the second answer and the second question would silently take its
	// default — which is exactly how a custom path prompt loses the path.
	var out bytes.Buffer
	p := NewPrompter(strings.NewReader("custom\n/work/openspec/changes\nlight\n"), &out)

	choice, err := p.Choice("Where", []string{"default", "custom"}, "default")
	require.NoError(t, err)
	path, err := p.Line("Path", "/fallback")
	require.NoError(t, err)
	workflow, err := p.Choice("Workflow", []string{"strict", "light", "free"}, "strict")
	require.NoError(t, err)

	assert.Equal(t, "custom", choice)
	assert.Equal(t, "/work/openspec/changes", path)
	assert.Equal(t, "light", workflow)
}
