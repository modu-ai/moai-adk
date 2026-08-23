package feedback

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// maskLogFileName is the mask log's name under .moai/logs.
const maskLogFileName = "feedback-mask.log"

// maskLogPerm is 0o600 rather than the 0o644 the config log uses: the subject
// here is secret-adjacent, and a log about credential masking that is
// world-readable is a worse artefact than no log at all.
const maskLogPerm os.FileMode = 0o600

// maskLogDirPerm is the directory mode, matching the other .moai/logs writers.
const maskLogDirPerm os.FileMode = 0o755

// MaskLogPathForRoot returns the mask log's canonical location under a project
// root. One join, so no two callers can disagree about where the log lives.
func MaskLogPathForRoot(root string) string {
	return filepath.Join(root, defs.MoAIDir, defs.LogsSubdir, maskLogFileName)
}

// appendMaskLog records one scrub's findings.
//
// [HARD] fail-open. Logging is an observation of the scrub, never a
// precondition of it: a caller that cannot write the log still gets its masked
// text, because refusing to mask when the log is unavailable would push the
// user back onto the unmasked manual path — the opposite of what the control
// is for. Every failure degrades to slog.Warn and returns.
//
// [HARD] The entry carries kind, location and count only. Findings hold no raw
// value by construction (see Finding), and nothing here reintroduces one: the
// log is the only artefact that outlives the scrub, so a value written here
// would make the control the leak path (AP-6).
//
// The line-oriented append is deliberate, and deliberately unlike the retry
// queue's locked JSON document: an interleaved write from a concurrent scrub
// costs one garbled line, whereas a lock would let a stuck holder block the
// scrub itself.
func appendMaskLog(root string, findings []Finding, at time.Time) {
	if root == "" || len(findings) == 0 {
		return
	}

	path := MaskLogPathForRoot(root)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, maskLogDirPerm); err != nil {
		slog.Warn("feedback: cannot create mask log directory", "dir", dir, "error", err)
		return
	}

	f, err := os.OpenFile(filepath.Clean(path), os.O_APPEND|os.O_CREATE|os.O_WRONLY, maskLogPerm)
	if err != nil {
		slog.Warn("feedback: cannot open mask log", "file", path, "error", err)
		return
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(maskLogEntry(findings, at)); err != nil {
		slog.Warn("feedback: cannot write mask log", "file", path, "error", err)
	}
}

// maskLogEntry renders one line: an RFC3339 timestamp, the total, then one
// field group per finding.
//
//	2026-08-23T18:00:00+09:00 | total=3 | kind=env where=title count=1 | kind=secret where=body count=2
func maskLogEntry(findings []Finding, at time.Time) string {
	total := 0
	parts := make([]string, 0, len(findings)+2)
	parts = append(parts, at.Format(time.RFC3339), "")
	for _, f := range findings {
		total += f.Count
		parts = append(parts, fmt.Sprintf("kind=%s where=%s count=%d", f.Kind, f.Where, f.Count))
	}
	parts[1] = fmt.Sprintf("total=%d", total)
	return strings.Join(parts, " | ") + "\n"
}
