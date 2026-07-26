package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstructions_AllTypes(t *testing.T) {
	types := []string{"context", "proposal", "spec", "design", "tasks"}
	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			output, err := ExecuteCommand(rootCmd, "instructions", typ)
			require.NoError(t, err)
			assert.NotEmpty(t, output)
			assert.Contains(t, output, "#")
		})
	}
}

func TestInstructions_InvalidType(t *testing.T) {
	_, err := ExecuteCommand(rootCmd, "instructions", "invalid")
	assert.Error(t, err)
}

func TestInstructions_NoArgs(t *testing.T) {
	_, err := ExecuteCommand(rootCmd, "instructions")
	assert.Error(t, err)
}
