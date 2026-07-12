// Package harness — Layer 3 CLAUDE.md harness marker injector.
package harness

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness/curator"
)

// InjectMarker injects (or replaces) the harness block in the file at
// claudeMdPath. If a block already exists (regardless of its specID), it is
// replaced atomically with the new block built from specID, domain, and
// importPaths. The function is idempotent: re-running with same or different
// specIDs produces exactly one block per file.
//
// The marker-block locate/replace/append mechanics are delegated to the
// typed writer curator.WriteManagedBlock (BlockTypeHarnessGenerated). This
// function is preserved as a thin backward-compatible wrapper so existing
// callers — notably internal/cli/harness/install.go RunInstall (the sole
// production caller) — continue to work byte-identical through the typed
// writer (REQ-HEV2-002, AP-HEV2-004).
func InjectMarker(claudeMdPath, specID, domain string, importPaths []string) error {
	if claudeMdPath == "" {
		return errors.New("InjectMarker: empty path")
	}
	if specID == "" {
		return errors.New("InjectMarker: empty specID")
	}
	startAttrs, body := buildHarnessBody(specID, domain, importPaths)
	return curator.WriteManagedBlock(claudeMdPath, curator.BlockTypeHarnessGenerated,
		curator.BlockContent{RawBody: body, StartAttrs: startAttrs})
}

// buildHarnessBody renders the structured body + start-marker attributes for
// the Harness-Generated managed block. The body format (Domain / Harness
// level / Updated / import paths) is preserved byte-identical from the
// original buildMarkerBlock so the production output of InjectMarker is
// unchanged after the curator generalization (D1 load-bearing constraint).
func buildHarnessBody(specID, domain string, importPaths []string) (startAttrs, body string) {
	now := time.Now().UTC()
	startAttrs = fmt.Sprintf(` id=%q generated=%q`, specID, now.Format(time.RFC3339))
	var b strings.Builder
	fmt.Fprintf(&b, "**Domain**: %s\n", domain)
	b.WriteString("**Harness level**: standard\n")
	fmt.Fprintf(&b, "**Updated**: %s\n", now.Format("2006-01-02"))
	if len(importPaths) > 0 {
		b.WriteString("\n")
		for _, p := range importPaths {
			fmt.Fprintf(&b, "See @%s\n", p)
		}
	}
	return startAttrs, b.String()
}
