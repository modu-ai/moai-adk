// Package measure is a pure, dependency-free leaf package holding the project-health
// parsers (go test JSON, coverage profile, lint line count) shared by both the loop
// feedback generator (internal/loop) and the harness regression gate (internal/harness).
//
// SPEC-HARNESS-REGRESSION-GATE-001 REQ-RG-002 / DD-1: these parsers were extracted
// verbatim from internal/loop/go_feedback.go so that internal/harness can reuse them
// without importing internal/loop (which pulls in internal/lsp + internal/lsp/gopls).
// This package MUST import none of internal/lsp, internal/lsp/gopls, internal/harness,
// or internal/loop — it depends only on the Go standard library (C9, proven by
// `go list -deps`).
package measure

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"strings"
)

// goTestEvent represents a single JSON event from `go test -json`.
type goTestEvent struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"`
}

// ParseGoTestJSON parses go test -json output and returns (passed, failed) counts.
// Only top-level test results are counted (events with a non-empty Test field);
// package-level events (empty Test) are ignored.
//
// @MX:ANCHOR: [AUTO] ParseGoTestJSON is the shared test-pass parser.
// @MX:REASON: [AUTO] fan_in >= 3: measure_test.go, internal/loop go_feedback.go, internal/harness regression collector.
func ParseGoTestJSON(data []byte) (passed, failed int) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		var ev goTestEvent
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			continue
		}
		// Only count top-level test results (Test field non-empty, package-level events have empty Test).
		if ev.Test == "" {
			continue
		}
		switch ev.Action {
		case "pass":
			passed++
		case "fail":
			failed++
		}
	}
	return
}

// ParseCoverageFile reads a Go coverage profile and returns the total coverage percentage.
// An absent or unreadable file yields 0.
//
// @MX:ANCHOR: [AUTO] ParseCoverageFile is the shared coverage parser.
// @MX:REASON: [AUTO] fan_in >= 3: measure_test.go, internal/loop go_feedback.go, internal/harness regression collector.
func ParseCoverageFile(path string) float64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	var totalStatements, coveredStatements int
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		// Format: file:startLine.startCol,endLine.endCol numStatements count
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		// parts[1] = numStatements, parts[2] = count
		var stmts, count int
		if _, err := parseIntFromString(parts[1]); err == nil {
			stmts = mustParseInt(parts[1])
		}
		if _, err := parseIntFromString(parts[2]); err == nil {
			count = mustParseInt(parts[2])
		}
		totalStatements += stmts
		if count > 0 {
			coveredStatements += stmts
		}
	}

	if totalStatements == 0 {
		return 0
	}
	return float64(coveredStatements) / float64(totalStatements) * 100.0
}

// CountNonEmptyLines counts the number of non-empty lines in byte data.
// Used to count go vet lint output lines.
//
// @MX:ANCHOR: [AUTO] CountNonEmptyLines is the shared lint-line parser.
// @MX:REASON: [AUTO] fan_in >= 3: measure_test.go, internal/loop go_feedback.go, internal/harness regression collector.
func CountNonEmptyLines(data []byte) int {
	count := 0
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	return count
}

// parseIntFromString is a helper that validates a string is a valid integer.
func parseIntFromString(s string) (int, error) {
	return mustParseIntErr(s)
}

func mustParseInt(s string) int {
	v, _ := mustParseIntErr(s)
	return v
}

func mustParseIntErr(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, &json.InvalidUnmarshalError{}
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
