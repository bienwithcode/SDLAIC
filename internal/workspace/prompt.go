package workspace

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// IsTerminal returns true if os.Stdin is connected to a TTY/terminal.
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// AskChoice prompts the user via the provided reader/writer and returns a validated option.
// It will trim spaces, check for an empty input (returning defaultVal), and loop until
// a valid case-insensitive match from options is entered.
func AskChoice(in io.Reader, out io.Writer, prompt string, options []string, defaultVal string) (string, error) {
	scanner := bufio.NewScanner(in)
	optStr := strings.Join(options, "/")

	for {
		fmt.Fprintf(out, "%s [%s] (default %s): ", prompt, optStr, defaultVal)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return defaultVal, nil // EOF returns defaultVal
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return defaultVal, nil
		}

		// Validate option
		for _, opt := range options {
			if strings.EqualFold(input, opt) {
				return opt, nil
			}
		}
		fmt.Fprintf(out, "Invalid choice: %q. Please try again.\n", input)
	}
}
