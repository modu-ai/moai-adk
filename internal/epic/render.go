package epic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// RenderJSON serializes the EpicStatus into the frozen-shape JSON document
// (REQ-ES-008 §B.1). The output is a single JSON document on stdout with
// stable key ordering (struct field declaration order). Callers MUST tolerate
// unknown extra fields (forward-compat parse rule — AC-ES-009).
func RenderJSON(s *EpicStatus) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return nil, fmt.Errorf("encode epic status: %w", err)
	}
	// json.Encoder appends a trailing newline; trim it for byte-stable output
	// (the CLI layer adds its own newline via fmt.Fprintln).
	out := buf.Bytes()
	if n := len(out); n > 0 && out[n-1] == '\n' {
		out = out[:n-1]
	}
	return out, nil
}

// localeProgressLabel returns the localized "Epic progress" label per the
// 4-locale table at moai.md:593-599. Falls back to English.
func localeProgressLabel(locale string) string {
	switch strings.ToLower(locale) {
	case "ko":
		return "에픽 진행"
	case "ja":
		return "エピック進行"
	case "zh":
		return "史诗进度"
	default:
		return "Epic progress"
	}
}

// statusIcon maps a milestone status to its Progress Board icon
// (moai.md:660-668 legend): done=🟢, in-progress=🟡, planned=⬜, absent=⬜.
func statusIcon(status string) string {
	switch status {
	case "done":
		return "🟢"
	case "in-progress":
		return "🟡"
	default:
		return "⬜"
	}
}

// renderBar draws the 10-cell Progress Board bar: round(done/total*10) ▓ +
// remainder ░.
func renderBar(done, total int) string {
	if total <= 0 {
		return strings.Repeat("░", 10)
	}
	filled := int(math.Round(float64(done) / float64(total) * 10))
	if filled > 10 {
		filled = 10
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("▓", filled) + strings.Repeat("░", 10-filled)
}

// RenderHuman emits the Progress Board grammar (REQ-ES-007, AC-ES-008b):
//
//	🎯 <epic> ▓▓░░░░░░░░ done/total (pct%)
//	Epic progress:   <label>
//	  🟢 M0 <label>            SPEC-XXX (completed)
//	  🟡 M1 <label>            SPEC-YYY (in-progress)
//	  ⬜ M2 <label>            orphan
//
// Translated per the 4-locale table when locale ∈ {ko, ja, zh}.
func RenderHuman(s *EpicStatus, locale string) (string, error) {
	var b strings.Builder
	label := localeProgressLabel(locale)

	if s.Total == 0 {
		// AC-ES-003b: empty epic — "no SPECs matched prefix". Still emit the
		// locale label so a non-English session sees its own script.
		fmt.Fprintf(&b, "🎯 %s — no SPECs matched prefix '%s'\n", s.Epic, s.Epic)
		fmt.Fprintf(&b, "%s:   %s\n", label, s.Epic)
		return b.String(), nil
	}

	// Header line: 🎯 epic ▓▓░░ done/total (pct%)
	fmt.Fprintf(&b, "🎯 %s %s %d/%d (%d%%)\n", s.Epic, renderBar(s.Done, s.Total), s.Done, s.Total, s.Pct)
	fmt.Fprintf(&b, "%s:   %s\n", label, s.Epic)

	// Per-milestone lines.
	for _, m := range s.Milestones {
		icon := statusIcon(m.Status)
		line := fmt.Sprintf("  %s %s %s", icon, m.ID, m.Label)
		// Pad to a column for the SPEC suffix.
		if len(line) < 40 {
			line += strings.Repeat(" ", 40-len(line))
		} else {
			line += " "
		}
		switch {
		case !m.Covered:
			line += "orphan"
		case m.SpecID != "":
			line += m.SpecID
			if m.SpecStatus != "" {
				line += " (" + m.SpecStatus + ")"
			}
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}

	// Orphan summary line when orphans exist.
	if len(s.OrphanMx) > 0 {
		fmt.Fprintf(&b, "orphan_mx: %s\n", strings.Join(s.OrphanMx, ", "))
	}
	return b.String(), nil
}
