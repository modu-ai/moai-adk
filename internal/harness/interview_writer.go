// Package harness — interview_writer.go provides the interview results writer
// that serializes a committed Buffer to the interview-results.md schema
// (SPEC-V3R3-PROJECT-HARNESS-001 plan.md §4.2).
package harness

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// roundHeaders maps round numbers to their canonical section headings.
var roundHeaders = map[int]string{
	1: "## Round 1: Domain & Technology Foundation",
	2: "## Round 2: Methodology & Design",
	3: "## Round 3: Security, Performance, Deployment",
	4: "## Round 4: Customization & Final Confirmation",
}

// errWriter wraps an io.Writer to track the first write error, removing the
// need to check Fprintf return values on every call. The first non-nil error
// is preserved; subsequent writes are no-ops.
type errWriter struct {
	w   io.Writer
	err error
}

// Printf writes formatted output to the underlying writer. If a previous
// write returned an error, the call is a no-op so the original error is
// preserved.
func (ew *errWriter) Printf(format string, args ...any) {
	if ew.err != nil {
		return
	}
	_, ew.err = fmt.Fprintf(ew.w, format, args...)
}

// WriteResults serializes the committed Buffer to w in the interview-results.md
// schema. It returns an error if:
//   - the buffer is not frozen (Commit not yet called)
//   - the buffer contains fewer than 16 answers
//   - the underlying writer returns an error during serialization
func WriteResults(buffer *Buffer, projectRoot, specID, conversationLang string, w io.Writer) error {
	if !buffer.Frozen() {
		return errors.New("harness: WriteResults: buffer must be committed before writing")
	}

	answers := buffer.Answers()
	if len(answers) < 16 {
		return fmt.Errorf("harness: WriteResults: expected 16 answers, got %d", len(answers))
	}

	generatedAt := time.Now().UTC().Format(time.RFC3339)
	ew := &errWriter{w: w}

	// --- YAML frontmatter ---
	ew.Printf("---\n")
	ew.Printf("spec_id: %s\n", specID)
	ew.Printf("generated_at: %s\n", generatedAt)
	ew.Printf("project_root: %s\n", projectRoot)
	ew.Printf("conversation_language: %s\n", conversationLang)
	ew.Printf("---\n\n")

	// --- Document heading ---
	ew.Printf("# Interview Results\n\n")

	// --- Per-round sections ---
	currentRound := 0
	for _, ans := range answers {
		if ans.Round != currentRound {
			currentRound = ans.Round
			header, ok := roundHeaders[currentRound]
			if !ok {
				header = fmt.Sprintf("## Round %d", currentRound)
			}
			if currentRound > 1 {
				ew.Printf("\n")
			}
			ew.Printf("%s\n", header)
		}
		recordedAt := ans.RecordedAt.UTC().Format(time.RFC3339)
		ew.Printf("\n- %s: %s\n", ans.QuestionID, ans.QuestionText)
		ew.Printf("  - Answer: %s\n", ans.AnswerText)
		ew.Printf("  - Recorded at: %s\n", recordedAt)
	}

	if ew.err != nil {
		return fmt.Errorf("harness: WriteResults: write: %w", ew.err)
	}
	return nil
}

// WriteResultsToFile serializes the committed Buffer to a file at path,
// creating parent directories if they do not exist. It is a convenience
// wrapper around WriteResults.
func WriteResultsToFile(buffer *Buffer, path, projectRoot, specID, lang string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("harness: WriteResultsToFile: mkdir: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("harness: WriteResultsToFile: create: %w", err)
	}
	defer f.Close() //nolint:errcheck

	if err := WriteResults(buffer, projectRoot, specID, lang, f); err != nil {
		return fmt.Errorf("harness: WriteResultsToFile: write: %w", err)
	}
	return nil
}
