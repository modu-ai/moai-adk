package hook

// stale_run_gate.go — the run-state gate behind the stale-run prescription
// (SPEC-STALE-RUN-LABEL-001 REQ-SRL-001..006).
//
// A legacy launch label alone used to decide the hook's whole answer: the
// branch returned the retire prescription before any run-state measurement,
// re-prescribing an already-retired run on every turn (the measured t1345
// loss: 2h40m of session work, 3 lanes locked out of dispatch reception).
// The gate measures the named run through the shared tri-state accessor
// (factorymsg.ProbeRunState — REQ-SRL-003) and answers from the measurement:
//
//	active      → the stale-run prescription, once per session identity
//	              (REQ-SRL-002);
//	not active  → the one-time unbind notice, naming the orphan label and
//	              the measured state, with the re-bind entry only while an
//	              active run exists in the same root (REQ-SRL-005/006);
//	unavailable → the degraded answer — never a prescription, never a hook
//	              error (REQ-SRL-003, fail-open).
//
// The once-per-identity carrier is a marker file under the factory state dir
// keyed by session identity, shared across ALL prescription surfaces
// (SessionStart bootstrap + peer registration, UserPromptSubmit peer path) so
// startup and the first prompt cannot both emit (plan M1.2). It is
// deliberately NOT the kanban session record — a legacy-label session must
// never grow one (SPEC-ROLE-NAMING-CODE-001) — and deliberately NOT the
// run's broker DB, which is the dead run's own store, absent exactly when
// the unbind path needs the carrier most. Marker failures fail open to
// over-informing (emit unmarked): repetition is the defect being repaired,
// so the marker may lose a dedup, never swallow a notice.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// factoryGateBudget bounds the gate's own measurement. It wraps the caller's
// context: the UserPromptSubmit peer path arrives under factoryBindBudget,
// the SessionStart surfaces under the handler context or none at all — the
// tighter of the two budgets wins either way.
//
// It is a variable only so tests can pin it generously (a 200 ms budget made
// the gate tests load-dependent — SPEC-FACTORY-STALE-RUN-HEAL-001 plan §B
// seams); production never assigns it.
var factoryGateBudget = factoryHookInspectionDeadline

const (
	factoryNoticePrescription = "prescription"
	factoryNoticeUnbind       = "unbind"
)

// factoryNoticeMarker is the once-per-session-identity dedup carrier. The
// two kinds are recorded independently: the prescription (REQ-SRL-002, at
// most once) and the unbind notice (REQ-SRL-005, exactly one and final —
// after it nothing emits for the identity again).
//
// State is the third, independent carrier: the last emitted state key of a
// current-vocabulary lane session (`rebound:Y`, `unbound:X`, `ambiguous:<ids>`,
// `refused:Y` — SPEC-FACTORY-STALE-RUN-HEAL-001 DP12). The two legacy fields
// are never read or written by that path, so the legacy cadence — and the
// finality of the legacy unbind notice — is untouched.
type factoryNoticeMarker struct {
	PrescriptionEmittedAt string `json:"prescription_emitted_at,omitempty"`
	UnbindEmittedAt       string `json:"unbind_emitted_at,omitempty"`
	State                 string `json:"state,omitempty"`
}

func (m factoryNoticeMarker) emitted(kind string) bool {
	switch kind {
	case factoryNoticePrescription:
		return m.PrescriptionEmittedAt != ""
	case factoryNoticeUnbind:
		return m.UnbindEmittedAt != ""
	}
	return true
}

func (m *factoryNoticeMarker) mark(kind string, at time.Time) {
	switch kind {
	case factoryNoticePrescription:
		m.PrescriptionEmittedAt = at.Format(time.RFC3339)
	case factoryNoticeUnbind:
		m.UnbindEmittedAt = at.Format(time.RFC3339)
	}
}

func factoryNoticeMarkerPath(dbPath, sessionID string) (string, error) {
	if dbPath == "" {
		return "", errors.New("factory state path unresolved")
	}
	// The notices dir is a sibling of the factory DB inside the factory dir
	// (homestate.FactoryDir == filepath.Dir(FactoryDBPath)) — derived from the
	// already-resolved DB path so one gate answer never re-resolves it.
	return filepath.Join(filepath.Dir(dbPath), "notices", safeNoticeStem(sessionID)+".json"), nil
}

// safeNoticeStem reduces the session identity to a filesystem-safe stem.
// Hook session ids are UUIDs in production; anything outside the safe set is
// hashed so a hostile id cannot escape the notices dir.
func safeNoticeStem(sessionID string) string {
	const safe = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
	if sessionID != "" && len(sessionID) <= 128 &&
		strings.IndexFunc(sessionID, func(r rune) bool { return !strings.ContainsRune(safe, r) }) < 0 {
		return sessionID
	}
	sum := sha256.Sum256([]byte(sessionID))
	return "h-" + hex.EncodeToString(sum[:8])
}

func readFactoryNoticeMarker(dbPath, sessionID string) factoryNoticeMarker {
	var m factoryNoticeMarker
	if sessionID == "" || dbPath == "" {
		return m
	}
	path, err := factoryNoticeMarkerPath(dbPath, sessionID)
	if err != nil {
		return m
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return m
	}
	_ = json.Unmarshal(data, &m)
	return m
}

func markFactoryNotice(dbPath, sessionID, kind string) {
	updateFactoryNoticeMarker(dbPath, sessionID, func(m *factoryNoticeMarker) { m.mark(kind, time.Now()) })
}

// updateFactoryNoticeMarker reads the session identity's marker, applies edit
// and writes it back; fields edit does not touch are preserved. Failures fail
// open exactly as markFactoryNotice documents.
func updateFactoryNoticeMarker(dbPath, sessionID string, edit func(*factoryNoticeMarker)) {
	if sessionID == "" || dbPath == "" {
		// A session without an identity cannot own a carrier; fail open to
		// over-informing (emit unmarked). Degenerate input — production hook
		// inputs always carry the session UUID.
		return
	}
	path, err := factoryNoticeMarkerPath(dbPath, sessionID)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return
	}
	m := readFactoryNoticeMarker(dbPath, sessionID)
	edit(&m)
	data, err := json.Marshal(m)
	if err != nil {
		return
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		slog.Warn("factory gate: notice marker write failed (dedup lost, notice still emitted)", "error", err)
	}
}

// staleRunPrescriptionGate returns the hook's answer for a legacy-label
// session from the measured run state. Surfaces: the peer registration path
// (registerFactoryHookPeer), the SessionStart factory bootstrap, and
// staleRunNoticeFor's factory branch. lang follows the surface's audience.
// The notice is marked before it is returned — at-most-once beats
// at-least-once here, because repetition is the defect being repaired.
func staleRunPrescriptionGate(ctx context.Context, root, sessionID, label, runID, lang string) string {
	if strings.TrimSpace(runID) == "" {
		// No factory run is named — the answer is the relaunch prose, which
		// carries no retire step and needs no gate.
		return legacyFactoryHookNotice(label, runID, lang)
	}
	gateCtx, cancel := context.WithTimeout(ctx, factoryGateBudget)
	defer cancel()
	// The factory DB path is resolved ONCE for the whole answer: every
	// homestate path helper re-runs CanonicalProjectRoot (git subprocesses),
	// and five resolutions were measured exhausting the gate budget before
	// the last measurement ran (the rebind-line test caught exactly that).
	dbPath, err := homestate.FactoryDBPath(root)
	if err != nil {
		return "factory messaging degraded: " + err.Error()
	}
	state, status, err := factorymsg.ProbeRunStateAt(gateCtx, dbPath, runID)
	switch state {
	case factorymsg.RunStateUnavailable:
		// Measurement failure fails open: the degraded answer, never a
		// prescription, never a hook error (REQ-SRL-003).
		if err == nil {
			err = errors.New("factory state unmeasurable")
		}
		return "factory messaging degraded: " + err.Error()
	case factorymsg.RunStateActive:
		carrier := readFactoryNoticeMarker(dbPath, sessionID)
		if carrier.emitted(factoryNoticePrescription) || carrier.emitted(factoryNoticeUnbind) {
			return ""
		}
		markFactoryNotice(dbPath, sessionID, factoryNoticePrescription)
		return legacyFactoryHookNotice(label, runID, lang)
	default: // factorymsg.RunStateNotActive — a measured verdict
		carrier := readFactoryNoticeMarker(dbPath, sessionID)
		if carrier.emitted(factoryNoticeUnbind) {
			return ""
		}
		markFactoryNotice(dbPath, sessionID, factoryNoticeUnbind)
		return unbindFactoryHookNotice(gateCtx, dbPath, label, runID, status, lang)
	}
}

// unbindFactoryHookNotice renders the one-time unbind notice (REQ-SRL-005):
// it names the orphan label and the measured run state, and — only while an
// active run exists in the same root (REQ-SRL-006) — the executable relaunch
// line(s) of rows R7-R9 of the notice-line table (spec.md §D.7). A failed
// listing omits the lines (fail-open).
func unbindFactoryHookNotice(ctx context.Context, dbPath, label, runID, status, lang string) string {
	if !kanban.IsLegacyFactoryRoleValue(strings.TrimSpace(label)) {
		return ""
	}
	m := staleRunMessagesFor(lang)
	notice := fmt.Sprintf(m.laneLabelUnbind, label, runID, status)
	active, err := factorymsg.ActiveRunIDsAt(ctx, dbPath)
	if err != nil || len(active) == 0 {
		return notice
	}
	lines := kanban.RelaunchNoticeFor(kanban.RelaunchNoticeState{
		Provider:   kanban.RelaunchProviderForBackend(os.Getenv(config.EnvMoaiKanbanBackend)),
		Legacy:     true,
		Run:        runID,
		ActiveRuns: active,
	})
	header := m.laneLabelUnbindRebind
	if len(active) > 1 {
		header = m.laneLabelUnbindMany
	}
	notice += "\n" + header + "\n" + strings.Join(lines.Lines, "\n")
	if lines.More > 0 {
		notice += "\n" + fmt.Sprintf(m.laneLabelUnbindMore, lines.More)
	}
	return notice
}

// gatedStaleRunAnswer applies the run-state gate exactly where the answer
// would carry the factory retire step — runID set and the factory fan-out
// env stamped (config.EnvMoaiFactoryWorkers), the same discriminator
// staleRunNotice uses for its factory branch. The kanban relaunch prose
// names no factory run and stays ungated.
func gatedStaleRunAnswer(root, sessionID, label, lang string) string {
	runID := strings.TrimSpace(os.Getenv(config.EnvMoaiKanbanID))
	if runID != "" && os.Getenv(config.EnvMoaiFactoryWorkers) != "" {
		return staleRunPrescriptionGate(context.Background(), root, sessionID, label, runID, lang)
	}
	return staleRunNotice(label, lang)
}
