package test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mikeschinkel/go-pathvars"
)

// TestFuzzCorpus runs all discovered corpus files from fuzzing to ensure
// they continue to pass without panics or infinite loops. This test runs
// on every CI/CD build and is much faster than full fuzzing.
func TestFuzzCorpus(t *testing.T) {
	corpusDir := "testdata/fuzz/FuzzParseTemplate"

	entries, err := os.ReadDir(corpusDir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("No corpus directory found (run fuzzing first to generate corpus)")
			return
		}
		t.Fatalf("Failed to read corpus directory: %v", err)
	}

	if len(entries) == 0 {
		t.Skip("Corpus directory is empty")
		return
	}

	var infiniteLoops, panics, parseErrors, successes int

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Read corpus file
		corpusPath := filepath.Join(corpusDir, entry.Name())
		data, err := os.ReadFile(corpusPath)
		if err != nil {
			t.Errorf("Failed to read corpus file %s: %v", entry.Name(), err)
			continue
		}

		// Extract input string from corpus file format
		input := extractCorpusInput(string(data))
		if input == "" {
			t.Logf("Skipping empty or unparseable corpus file: %s", entry.Name())
			continue
		}

		// Test with timeout
		done := make(chan struct{})
		var parseErr error
		var panicked bool

		go func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
					t.Errorf("PANIC in corpus file %s with input %q: %v", entry.Name(), input, r)
				}
				close(done)
			}()

			_, parseErr = pathvars.ParseTemplate(input)
		}()

		select {
		case <-done:
			// Completed successfully
			if panicked {
				panics++
			} else if parseErr != nil {
				parseErrors++
			} else {
				successes++
			}

		case <-time.After(1 * time.Second):
			infiniteLoops++
			t.Errorf("INFINITE LOOP detected in corpus file %s: input=%q", entry.Name(), input)
		}
	}

	// Log summary
	total := successes + parseErrors + panics + infiniteLoops
	t.Logf("Corpus test summary: %d total cases", total)
	t.Logf("  - %d successes (parsed without error)", successes)
	t.Logf("  - %d parse errors (expected for malformed input)", parseErrors)
	t.Logf("  - %d panics (MUST FIX)", panics)
	t.Logf("  - %d infinite loops (MUST FIX)", infiniteLoops)

	// Fail test if we found critical issues
	if panics > 0 {
		t.Errorf("Found %d corpus files causing panics", panics)
	}
	if infiniteLoops > 0 {
		t.Errorf("Found %d corpus files causing infinite loops", infiniteLoops)
	}
}

// extractCorpusInput parses a Go fuzz corpus file and extracts the string input.
// Corpus file format: "go test fuzz v1\nstring(\"value\")"
func extractCorpusInput(corpusData string) string {
	lines := strings.Split(corpusData, "\n")
	if len(lines) < 2 {
		return ""
	}

	// Skip the header line "go test fuzz v1"
	// Parse the second line which should be: string("value") or string("multi\nline")
	line := strings.TrimSpace(lines[1])
	if !strings.HasPrefix(line, "string(") {
		return ""
	}

	// Extract the quoted string - it's a Go string literal
	// Find the opening quote
	start := strings.Index(line, `"`)
	if start == -1 {
		return ""
	}

	// The rest of the file might be the multi-line string value
	// We need to parse it as a Go string literal, which might span multiple lines
	remainder := line[start:]

	// Try to unquote as a single-line string first
	if end := strings.LastIndex(line, `"`) + 1; end > start {
		candidate := line[start:end]
		if unescaped, err := strconv.Unquote(candidate); err == nil {
			return unescaped
		}
	}

	// If that failed, the string might span multiple lines
	// Rebuild the full string literal from all remaining lines
	var fullLiteral strings.Builder
	fullLiteral.WriteString(remainder)

	for i := 2; i < len(lines); i++ {
		fullLiteral.WriteString("\n")
		fullLiteral.WriteString(lines[i])
	}

	literalStr := fullLiteral.String()

	// Find the end of the string literal (closing quote followed by closing paren)
	// This is a simplified parser - Go's string literals can be complex
	inString := false
	escaped := false

	for i, ch := range literalStr {
		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' && inString {
			escaped = true
			continue
		}

		if ch == '"' {
			if !inString {
				inString = true
			} else {
				// Found closing quote
				// Extract everything up to and including this quote
				quotedStr := literalStr[:i+1]
				if unescaped, err := strconv.Unquote(quotedStr); err == nil {
					return unescaped
				}
				return ""
			}
		}
	}

	return ""
}
