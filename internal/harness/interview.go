// Package harness implements the Socratic interview buffer and result writer
// for the Phase 5 Harness Activation workflow (SPEC-V3R3-PROJECT-HARNESS-001).
package harness

import (
	"errors"
	"time"
)

// Answer represents a single question-answer pair captured during the
// 16-question Socratic interview (REQ-PH-002).
type Answer struct {
	// QuestionID is the canonical question identifier (Q01..Q16).
	QuestionID string
	// Round is the interview round (1..4).
	Round int
	// QuestionText is the full question text in conversation_language.
	QuestionText string
	// AnswerText is the user-selected answer (verbatim, including Korean text).
	AnswerText string
	// RecordedAt is the UTC timestamp when the answer was recorded.
	RecordedAt time.Time
}

// Buffer is an in-memory, append-only accumulator for interview Answers.
// No disk I/O is performed until Commit() is called; Abort() guarantees
// zero disk writes and clears the in-memory state (REQ-PH-010).
type Buffer struct {
	answers []Answer
	frozen  bool
}

// NewBuffer creates and returns an empty, mutable Buffer.
func NewBuffer() *Buffer {
	return &Buffer{}
}

// Append adds an Answer to the buffer. Returns an error if the buffer
// is already frozen (i.e., Commit or Abort has been called).
func (b *Buffer) Append(ans Answer) error {
	if b.frozen {
		return errors.New("harness: Append: buffer is frozen")
	}
	b.answers = append(b.answers, ans)
	return nil
}

// Abort discards all in-memory answers and marks the buffer as frozen.
// It MUST NOT touch the disk in any way (REQ-PH-010).
func (b *Buffer) Abort() {
	b.answers = nil
	b.frozen = true
}

// Commit marks the buffer as frozen, preventing further Append calls.
// Returns an error if the buffer is already frozen.
func (b *Buffer) Commit() error {
	if b.frozen {
		return errors.New("harness: Commit: buffer is already frozen")
	}
	b.frozen = true
	return nil
}

// Frozen reports whether the buffer has been committed or aborted.
func (b *Buffer) Frozen() bool {
	return b.frozen
}

// Answers returns a shallow copy of the accumulated answers, providing
// an immutable view of the buffer contents.
func (b *Buffer) Answers() []Answer {
	if len(b.answers) == 0 {
		return []Answer{}
	}
	result := make([]Answer, len(b.answers))
	copy(result, b.answers)
	return result
}

// Len returns the number of answers currently in the buffer.
func (b *Buffer) Len() int {
	return len(b.answers)
}
