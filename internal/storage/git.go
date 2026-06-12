package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AppendToGitignore adds an entry to the .gitignore file in the given directory.
// If the entry already exists, it is not added again (idempotent).
// If .gitignore doesn't exist, it is created.
func AppendToGitignore(dir string, entry string) error {
	gitignorePath := filepath.Join(dir, ".gitignore")

	// Check if entry already exists
	if entryExists(gitignorePath, entry) {
		return nil
	}

	// Open for append (create if needed)
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening .gitignore: %w", err)
	}
	defer f.Close()

	line := entry
	// Ensure we start on a new line if file has content
	info, err := f.Stat()
	if err == nil && info.Size() > 0 {
		line = "\n" + entry
	}

	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("writing to .gitignore: %w", err)
	}

	return nil
}

// RemoveFromGitignore removes an entry from the .gitignore file in the given directory.
// If the entry doesn't exist or the file doesn't exist, it returns nil (no error).
func RemoveFromGitignore(dir string, entry string) error {
	gitignorePath := filepath.Join(dir, ".gitignore")

	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		return fmt.Errorf("reading .gitignore: %w", err)
	}

	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != strings.TrimSpace(entry) {
			lines = append(lines, scanner.Text())
		}
	}

	// Write back
	content := strings.Join(lines, "\n")
	if err := os.WriteFile(gitignorePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing .gitignore: %w", err)
	}

	return nil
}

// entryExists checks if a specific entry exists in the gitignore file.
func entryExists(gitignorePath string, entry string) bool {
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		return false
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == strings.TrimSpace(entry) {
			return true
		}
	}
	return false
}
