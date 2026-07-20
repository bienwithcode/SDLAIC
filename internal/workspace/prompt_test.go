package workspace

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
