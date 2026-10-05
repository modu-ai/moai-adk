package cli

// t508 (SPEC-CODEX-ENABLED-FATAL-001) M4 item 8 — the prune verb, EXECUTED.
//
// SPEC §A claims "the prune verb cannot create the fatal shape" and marks that
// claim a carried GAP: it rested on READING pruneCodexSkillEntries and
// judgeCodexSkillEntry on develop, with no fixture ever run against them. This
// file executes the prune and replaces the reading with an observation.
//
// It also records this SPEC's own effect on that other card's code. Card t506
// (SPEC-CODEX-GHOST-SKILLS-PRUNE-001) owns the prune; nothing here modifies it.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// countUnusableEnabled counts entries codex 0.153.4 cannot load: the key absent,
// or declared with a value that is not a bare TOML boolean.
func countUnusableEnabled(content []byte) int {
	n := 0
	for _, e := range codexwiring.ParseSkillEntries(content) {
		switch e.Enabled {
		case codexwiring.SkillEnabledUnspecified, codexwiring.SkillEnabledNonBoolean:
			n++
		}
	}
	return n
}

// TestPruneCodexSkillEntries_DoesNotManufactureFatalShape executes the prune
// against a config carrying every shape at once and asserts the output is not
// worse than the input on the `enabled` axis.
//
// "Not worse" is the right assertion rather than "identical": the prune legally
// REMOVES entries, and removing an entry that was itself unusable lowers the
// count. What must not happen is the count RISING — an entry that codex could
// load before the prune and cannot after.
func TestPruneCodexSkillEntries_DoesNotManufactureFatalShape(t *testing.T) {
	live := liveSkillFile(t)
	absent := absentSkillPath(t, "ghost")

	in := "model = \"gpt-5\"\n\n" +
		// healthy, path resolves — must survive untouched
		"[[skills.config]]\npath = \"" + live + "\"\nenabled = true\n\n" +
		// ghost: path absent, enabled well-formed — the prune's actual target
		"[[skills.config]]\npath = \"" + absent + "\"\nenabled = false\n\n" +
		// fatal shape: no enabled key at all
		"[[skills.config]]\npath = \"" + live + "\"\n\n" +
		// fatal shape: a value codex rejects
		"[[skills.config]]\npath = \"" + live + "\"\nenabled = \"true\"\n"

	before := countUnusableEnabled([]byte(in))
	if before != 2 {
		t.Fatalf("fixture setup: %d unusable entries, want 2 — the fixture does not exercise the claim", before)
	}

	out, verdicts := pruneCodexSkillEntries([]byte(in))

	if after := countUnusableEnabled(out); after > before {
		t.Errorf("the prune manufactured a fatal shape: %d unusable entries before, %d after\n%s", before, after, out)
	}
	if !strings.Contains(string(out), live) {
		t.Errorf("the prune removed the healthy registration:\n%s", out)
	}
	if strings.Contains(string(out), absent) {
		t.Errorf("the prune left the ghost entry behind:\n%s", out)
	}

	if len(verdicts) != 4 {
		t.Fatalf("got %d verdicts, want 4 (one per declared entry)", len(verdicts))
	}
	quoted := verdicts[3]
	if quoted.Eligible {
		t.Errorf("the quoted-`enabled` entry was judged eligible for deletion: %+v", quoted)
	}
}

// TestJudgeCodexSkillEntry_QuotedEnabledDispositionMoved measures this SPEC's
// effect on card t506's deletion predicate, on the one input where the
// DISPOSITION moves rather than only the reason.
//
// The two halves feed the SAME unmodified predicate — judgeCodexSkillEntry is
// t506's code and is read here, never touched. What differs is the SkillEntry
// the parser hands it:
//
//   - pre-t508, `enabled = "true"` matched the lenient matcher, so the line was
//     RECOGNISED (FirstUnrecognizedLine == -1) and read as SkillEnabledTrue.
//     That entry reached the existence test and, with a missing path, was
//     ELIGIBLE — the prune would have deleted it.
//   - post-t508 the line is unrecognised, so FirstUnrecognizedLine is the
//     `enabled` line's index and the deletion guard PRESERVES the entry.
//
// The pre-t508 half is constructed rather than remembered: the struct below is
// exactly what the old parser produced for this input, so the assertion is a
// measurement of the predicate on that input and not a claim about code that no
// longer exists.
//
// A path that RESOLVES would show no disposition change at all — the predicate
// skips it either way, only the reason differs — which is why this case uses a
// missing path.
func TestJudgeCodexSkillEntry_QuotedEnabledDispositionMoved(t *testing.T) {
	missing := absentSkillPath(t, "quoted-enabled")

	asOldParserRead := codexwiring.SkillEntry{
		Path:                  missing,
		Enabled:               codexwiring.SkillEnabledTrue,
		StartLine:             0,
		EndLine:               3,
		FirstUnrecognizedLine: -1,
	}
	if v := judgeCodexSkillEntry(asOldParserRead); !v.Eligible {
		t.Errorf("pre-t508 reading: want ELIGIBLE (the prune would have deleted it), got %+v", v)
	}

	entries := codexwiring.ParseSkillEntries([]byte(
		"[[skills.config]]\npath = \"" + missing + "\"\nenabled = \"true\"\n"))
	if len(entries) != 1 {
		t.Fatalf("ParseSkillEntries returned %d entries, want 1", len(entries))
	}
	v := judgeCodexSkillEntry(entries[0])
	if v.Eligible {
		t.Errorf("post-t508 reading: want PRESERVED, got eligible: %+v", v)
	}
	if !strings.Contains(v.SkipReason, "not recognised") {
		t.Errorf("skip reason = %q, want the unrecognised-line deletion guard", v.SkipReason)
	}
}

// TestPruneCodexSkillEntries_LeavesNoFileBehind is the companion posture check:
// pruneCodexSkillEntries is a pure content transform and writes nothing itself.
// The verb's own file write lives one layer up and is t506's to own.
func TestPruneCodexSkillEntries_LeavesNoFileBehind(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.toml")
	in := "[[skills.config]]\npath = \"" + absentSkillPath(t, "x") + "\"\nenabled = true\n"
	if err := os.WriteFile(cfg, []byte(in), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, verdicts := pruneCodexSkillEntries([]byte(in)); len(verdicts) != 1 {
		t.Fatalf("got %d verdicts, want 1 — the fixture never exercised the prune", len(verdicts))
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("the prune transform touched the directory: %d files, want 1", len(entries))
	}
	got, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != in {
		t.Errorf("the prune transform rewrote the file it was given:\n%s", got)
	}
}
