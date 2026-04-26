// Package harness — Layer 3 CLAUDE.md harness marker injector.
package harness

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// markerBlockPattern matches the entire harness block — heading + start marker
// + body + end marker — as one atomic group. Including the heading in the
// pattern is critical for idempotency: replacing only the marker block while
// the heading lives outside causes heading duplication on subsequent runs.
var markerBlockPattern = regexp.MustCompile(
	`(?s)## Project-Specific Configuration \(Harness-Generated\)\n<!-- moai:harness-start[^>]*-->.*?<!-- moai:harness-end -->`,
)

// InjectMarker injects (or replaces) the harness block in the file at
// claudeMdPath. If a block already exists (regardless of its specID), it is
// replaced atomically with the new block built from specID, domain, and
// importPaths. The function is idempotent: re-running with same or different
// specIDs produces exactly one block per file.
func InjectMarker(claudeMdPath, specID, domain string, importPaths []string) error {
	if claudeMdPath == "" {
		return errors.New("InjectMarker: empty path")
	}
	if specID == "" {
		return errors.New("InjectMarker: empty specID")
	}
	data, err := os.ReadFile(claudeMdPath)
	if err != nil {
		return fmt.Errorf("InjectMarker: read %s: %w", claudeMdPath, err)
	}
	block := buildMarkerBlock(specID, domain, importPaths)
	content := string(data)
	if markerBlockPattern.MatchString(content) {
		content = markerBlockPattern.ReplaceAllString(content, block)
	} else {
		// Append with separating newline if file does not end with one.
		sep := ""
		if !strings.HasSuffix(content, "\n") {
			sep = "\n"
		}
		content = content + sep + "\n" + block + "\n"
	}
	if err := os.WriteFile(claudeMdPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("InjectMarker: write %s: %w", claudeMdPath, err)
	}
	return nil
}

func buildMarkerBlock(specID, domain string, importPaths []string) string {
	now := time.Now().UTC()
	var b strings.Builder
	b.WriteString("## Project-Specific Configuration (Harness-Generated)\n")
	fmt.Fprintf(&b, "<!-- moai:harness-start id=%q generated=%q -->\n", specID, now.Format(time.RFC3339))
	fmt.Fprintf(&b, "**Domain**: %s\n", domain)
	b.WriteString("**Harness level**: standard\n")
	fmt.Fprintf(&b, "**Updated**: %s\n", now.Format("2006-01-02"))
	if len(importPaths) > 0 {
		b.WriteString("\n")
		for _, p := range importPaths {
			fmt.Fprintf(&b, "See @%s\n", p)
		}
	}
	b.WriteString("<!-- moai:harness-end -->")
	return b.String()
}
