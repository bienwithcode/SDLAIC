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

// Prompter asks a series of questions on one input stream.
//
// It exists because a bufio.Scanner reads ahead: giving each question its own
// scanner lets the first one swallow the answers to the rest, so the second
// question sees EOF and silently takes its default.
type Prompter struct {
	scanner *bufio.Scanner
	out     io.Writer
}

// NewPrompter returns a Prompter reading from in and writing prompts to out.
func NewPrompter(in io.Reader, out io.Writer) *Prompter {
	return &Prompter{scanner: bufio.NewScanner(in), out: out}
}

// read returns the next trimmed line, or ok=false at EOF.
func (p *Prompter) read() (string, bool, error) {
	if !p.scanner.Scan() {
		if err := p.scanner.Err(); err != nil {
			return "", false, err
		}
		return "", false, nil
	}
	return strings.TrimSpace(p.scanner.Text()), true, nil
}

// Choice asks until the answer matches one of options (case-insensitively).
// Empty input or EOF returns defaultVal.
func (p *Prompter) Choice(prompt string, options []string, defaultVal string) (string, error) {
	optStr := strings.Join(options, "/")

	for {
		fmt.Fprintf(p.out, "%s [%s] (default %s): ", prompt, optStr, defaultVal)

		input, ok, err := p.read()
		if err != nil {
			return "", err
		}
		if !ok || input == "" {
			return defaultVal, nil
		}

		for _, opt := range options {
			if strings.EqualFold(input, opt) {
				return opt, nil
			}
		}
		fmt.Fprintf(p.out, "Invalid choice: %q. Please try again.\n", input)
	}
}

// Line asks for a free-text value, used where the answer is not one of a fixed
// set — a filesystem path, for instance. Empty input or EOF returns defaultVal.
func (p *Prompter) Line(prompt string, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Fprintf(p.out, "%s (default %s): ", prompt, defaultVal)
	} else {
		fmt.Fprintf(p.out, "%s: ", prompt)
	}

	input, ok, err := p.read()
	if err != nil {
		return "", err
	}
	if !ok || input == "" {
		return defaultVal, nil
	}
	return input, nil
}

// AskChoice asks a single choice question. Use a Prompter when asking more than
// one question on the same stream.
func AskChoice(in io.Reader, out io.Writer, prompt string, options []string, defaultVal string) (string, error) {
	return NewPrompter(in, out).Choice(prompt, options, defaultVal)
}

// AskLine asks a single free-text question. Use a Prompter when asking more
// than one question on the same stream.
func AskLine(in io.Reader, out io.Writer, prompt string, defaultVal string) (string, error) {
	return NewPrompter(in, out).Line(prompt, defaultVal)
}
