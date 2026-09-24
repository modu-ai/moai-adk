// Resolution: UPGRADE — CLAUDE.md 40,000-char budget check per coding-standards.md.
package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
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
	// Official stdin field for InstructionsLoaded is file_path;
	// instruction_file_path is the legacy MoAI field name (kept as fallback).
	instructionPath := input.FilePath
	if instructionPath == "" {
		instructionPath = input.InstructionFilePath
	}

	slog.Info("instructions loaded",
		"session_id", input.SessionID,
		"file_path", instructionPath,
		"load_reason", input.LoadReason,
		"globs", input.Globs,
		"trigger_file_path", input.TriggerFilePath,
	)

	// The slog record above does not reach a reader HERE: resolveLoggingDecision
	// in internal/cli/logging.go keeps every `moai hook` invocation off
	// stdout/stderr, unconditionally, so both stay clean for the hook JSON
	// contract. It is an Info record, so it also sits below the sink's level
	// gate (.moai/logs/hook-runtime.log admits warn and above) — it is written
	// nowhere at all.
	//
	// The persistent record is therefore an audit row, written here — without it
	// the three host-supplied fields are observable nowhere.
	appendRuleLoadAudit(input.CWD, RuleLoadAuditRecord{
		SessionID:       input.SessionID,
		FilePath:        instructionPath,
		LoadReason:      input.LoadReason,
		Globs:           input.Globs,
		TriggerFilePath: input.TriggerFilePath,
	})

	// Check character budget for the loaded instruction file
	if instructionPath != "" {
		if err := h.checkCharacterBudget(instructionPath); err != nil {
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

// ruleLoadAuditFileName is the InstructionsLoaded observation log under
// <projectRoot>/.moai/logs/ — one JSONL row per rule/instruction load,
// shaped like agent-stop-audit.jsonl.
const ruleLoadAuditFileName = "rule-load-audit.jsonl"

// RuleLoadAuditRecord is one observed instruction-load event. LoadReason,
// Globs, and TriggerFilePath are the host-supplied fields (HookInput, v2.1.69+)
// that the handler previously dropped, which left `paths:` glob matching
// unobservable at runtime.
type RuleLoadAuditRecord struct {
	Timestamp       string   `json:"timestamp"`
	SessionID       string   `json:"session_id"`
	FilePath        string   `json:"file_path"`
	LoadReason      string   `json:"load_reason,omitempty"`
	Globs           []string `json:"globs,omitempty"`
	TriggerFilePath string   `json:"trigger_file_path,omitempty"`
}

// appendRuleLoadAudit appends one record to <projectRoot>/.moai/logs/.
// Every failure path is silent-and-continue, matching the agent-stop-audit
// precedent: an audit failure may never fail the observed hook event.
func appendRuleLoadAudit(projectRoot string, rec RuleLoadAuditRecord) {
	if projectRoot == "" {
		return
	}
	if rec.Timestamp == "" {
		rec.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	line, err := json.Marshal(rec)
	if err != nil {
		return
	}
	logsDir := filepath.Join(projectRoot, ".moai", "logs")
	if err := os.MkdirAll(logsDir, 0o750); err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(logsDir, ruleLoadAuditFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.Write(append(line, '\n'))
}
