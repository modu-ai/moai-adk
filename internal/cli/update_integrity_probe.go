// update_integrity_probe.go — the managed-surface integrity probe for the
// version-match skip path (SPEC-UPDATE-MIGRATION-FIX-001 M3; REQ-UMF-001..003).
//
// On a version-matched `moai update` the template sync is skipped, so nothing
// re-checks the managed project surface: a damaged file stays damaged through
// every non-force update. The probe closes that gap with a fixed, constant-cost
// observation of two representative paths. It is a canary for gross
// structural loss of the managed project core. It does NOT detect the damage
// classes of the mo.ai.kr production run; a damage-class-targeted set is an
// operator decision (decision-index Q2).
//
// The probe is warn-and-continue. It never fails the update and never changes
// the update's exit code (REQ-UMF-002).
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// probeCheck is the test a representative path must pass to count as intact.
type probeCheck int

const (
	// probeJSON: the path exists and parses as JSON.
	probeJSON probeCheck = iota
)

// integrityProbeEntry is one representative managed path, relative to the
// project root.
type integrityProbeEntry struct {
	rel   string
	check probeCheck
	// claudeOnly marks a member deployed only with the Claude surfaces. A codex-only
	// project never receives .claude/**, so the member is not managed there.
	claudeOnly bool
}

// appliesTo reports whether the entry is a managed member of the project rooted at
// projectRoot. A claude-only member is managed unless the project is codex-only:
// the codex-only deployers hide .claude/** (internal/template/harness_fs.go), so its
// absence there is the deploy's intended shape, not damage. The predicate is the one
// update_template_sync.go switches on: config.ReadHarness == "gpt".
func (e integrityProbeEntry) appliesTo(projectRoot string) bool {
	return !e.claudeOnly || config.ReadHarness(projectRoot) != "gpt"
}

// managedSurfaceProbeSet is the fixed representative set. It is a named slice so
// that a revision is one edit (plan.md M3). Each entry is a stable project file:
//   - .claude/settings.json: the Claude Code settings merge target; must parse as
//     JSON. Claude-harness projects only (claudeOnly, F2).
//   - .moai/manifest.json: the managed-file manifest; must parse as JSON.
//
// Every member must stay observable on the version-matched path, the only path the
// probe runs on (F3; TestIntegrityProbeSet_EveryMemberObservableOnVersionMatchedPath).
// system.yaml is therefore not a member: it carries the template-version stamp the
// skip predicate reads, so damage that empties or removes it makes the predicate
// false and the probe never runs.
var managedSurfaceProbeSet = []integrityProbeEntry{
	{rel: ".claude/settings.json", check: probeJSON, claudeOnly: true},
	{rel: ".moai/manifest.json", check: probeJSON},
}

// damageReason returns the reason the entry is damaged, or "" when it is intact.
// Every read failure is classified here; none escapes the probe.
//
// A member is read only after it is confirmed to be a regular file, and the read
// is bounded by plan.MaxConfigSize. A named pipe, socket, or device where a
// representative file belongs would otherwise block the read with no writer, and
// the recover in runManagedSurfaceIntegrityProbe cannot catch a block
// (SPEC-UPDATE-MIGRATION-FIX-001 F1).
func (e integrityProbeEntry) damageReason(projectRoot string) string {
	abs := filepath.Join(projectRoot, filepath.FromSlash(e.rel))
	info, err := os.Stat(abs)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "missing"
	case err != nil:
		return "unreadable"
	case !info.Mode().IsRegular():
		return "not a file"
	}
	data, reason := readProbeFile(abs)
	if reason != "" {
		return reason
	}
	switch e.check {
	case probeJSON:
		if !json.Valid(data) {
			return "unparseable"
		}
	}
	return ""
}

// readProbeFile reads the member at abs through a checked handle. The open is
// non-blocking, so a path swapped for a named pipe between the stat and the open
// cannot wedge the probe. The handle must then be a regular file, and the read is
// capped at plan.MaxConfigSize, the same bound the version-stamp reader applies.
// A file over that bound is reported unreadable: the probe does not check what it
// cannot read whole.
func readProbeFile(abs string) ([]byte, string) {
	f, err := os.OpenFile(abs, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, "missing"
		}
		return nil, "unreadable"
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return nil, "unreadable"
	}
	if !info.Mode().IsRegular() {
		return nil, "not a file"
	}
	if info.Size() > plan.MaxConfigSize {
		return nil, "unreadable"
	}
	data, err := io.ReadAll(io.LimitReader(f, plan.MaxConfigSize+1))
	if err != nil || len(data) > plan.MaxConfigSize {
		return nil, "unreadable"
	}
	return data, ""
}

// runManagedSurfaceIntegrityProbe checks the representative set under projectRoot
// and writes one warn row per damaged entry through the standard update channel.
// A fault inside the probe degrades to one warn row and is never returned to the
// caller, so it can never fail the update (REQ-UMF-002).
func runManagedSurfaceIntegrityProbe(out io.Writer, projectRoot string) {
	th := resolveTheme()
	defer func() {
		if r := recover(); r != nil {
			bugreport.Capture(bugreport.KindPanic, nil, "", nil)
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Integrity", "probe internal error", fmt.Sprint(r), &th))
		}
	}()
	for _, e := range managedSurfaceProbeSet {
		if !e.appliesTo(projectRoot) {
			continue
		}
		if reason := e.damageReason(projectRoot); reason != "" {
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Integrity", fmt.Sprintf("%s (%s)", e.rel, reason), "", &th))
		}
	}
}
