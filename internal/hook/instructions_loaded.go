package hook

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"unicode/utf8"
)

// instructionsLoadedHandler processes InstructionsLoaded events.
// It validates character budget compliance for CLAUDE.md and rules files.
type instructionsLoadedHandler struct{}

// NewInstructionsLoadedHandler creates a new InstructionsLoaded event handler.
func NewInstructionsLoadedHandler() Handler {
	return &instructionsLoadedHandler{}
}

// EventType returns EventInstructionsLoaded.
func (h *instructionsLoadedHandler) EventType() EventType {
	return EventInstructionsLoaded
}

// Handle processes an InstructionsLoaded event. It checks that loaded files
// comply with the 40,000 character budget per coding-standards.md.
func (h *instructionsLoadedHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("instructions loaded",
		"session_id", input.SessionID,
		"instruction_file_path", input.InstructionFilePath,
	)

	// Check character budget for the loaded instruction file
	if input.InstructionFilePath != "" {
		if err := h.checkCharacterBudget(input.InstructionFilePath); err != nil {
			return &HookOutput{
				SystemMessage: err.Error(),
			}, nil
		}
	}

	// Also check CLAUDE.md in CWD as a fallback
	if input.CWD != "" {
		claudeMDPath := filepath.Join(input.CWD, "CLAUDE.md")
		if _, err := os.Stat(claudeMDPath); err == nil {
			if budgetErr := h.checkCharacterBudget(claudeMDPath); budgetErr != nil {
				slog.Warn("CLAUDE.md exceeds budget", "path", claudeMDPath, "error", budgetErr)
				// Don't block on CLAUDE.md budget violations, just log
			}
		}
	}

	return &HookOutput{}, nil
}

// checkCharacterBudget verifies that a file does not exceed the 40,000 character limit.
func (h *instructionsLoadedHandler) checkCharacterBudget(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		// File may have been deleted or is inaccessible - log and continue
		slog.Debug("failed to read instruction file", "path", filePath, "error", err)
		return nil
	}

	// Count UTF-8 characters
	charCount := utf8.RuneCount(data)

	// Check budget limit (40,000 characters per coding-standards.md)
	const charBudget = 40000
	if charCount > charBudget {
		return fmt.Errorf("%s exceeds 40,000 char budget at %d; split content per coding-standards.md",
			filePath, charCount)
	}

	return nil
}
