package web

import (
	"bytes"
	"context"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// SPEC-WEB-CONSOLE-011 M5 — READ-ONLY SPEC board (REQ-WC11-040..046).
//
// The board is a purely observational dashboard: it renders the SPEC status
// distribution, an implemented-not-completed "close debt" section (the primary
// value), and the MUST-FIX SyncStatusDrift findings with a COPYABLE remediation
// string. It sources data exclusively from the pure-FS scanners spec.ListDocs +
// spec.Audit — the git-dependent drift-scan path is deliberately never invoked
// (REQ-WC11-045). It has NO write path, executes NO command server-side
// (REQ-WC11-044), and performs NO status transition (REQ-WC11-046). The route is
// GET-only.

// boardStatusOrder is the canonical 8-value status enum render order
// (spec-frontmatter-schema.md § Status Enum). Statuses outside this set are
// appended after, sorted, under their literal value (empty → "(unknown)").
var boardStatusOrder = []string{
	"draft", "planned", "in-progress", "implemented",
	"completed", "superseded", "archived", "rejected",
}

// boardStatusCount is one row of the status distribution summary.
type boardStatusCount struct {
	Status string
	Count  int
}

// boardSpec is one SPEC row in the close-debt section.
type boardSpec struct {
	ID      string
	Tier    string // "" → no badge (REQ-WC11-042: tier is OPTIONAL)
	Title   string
	Updated string
}

// boardFinding is one MUST-FIX drift finding with its copyable remediation.
type boardFinding struct {
	SpecID      string
	FindingType string
	Remediation string
}

// boardView is the typed view-model for the READ-ONLY SPEC board page.
type boardView struct {
	BindAddr     string
	TotalSpecs   int
	StatusCounts []boardStatusCount
	CloseDebt    []boardSpec
	MustFix      []boardFinding

	// Banner is an optional error message (BannerKind "error"); the board never
	// emits a success banner (there is nothing to save).
	Banner     string
	BannerKind string
}

// handleBoard serves GET /specs — the READ-ONLY SPEC dashboard. Non-GET methods
// are rejected with 405 (there is no write path — REQ-WC11-044/046).
func (a *app) handleBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	view, err := a.buildBoardView()
	if err != nil {
		a.renderError(w, http.StatusInternalServerError, "could not read SPEC board: "+err.Error())
		return
	}
	a.renderBoard(w, http.StatusOK, view)
}

// buildBoardView assembles the board view-model from the two pure-FS scanners:
// spec.ListDocs (per-SPEC frontmatter) + spec.Audit (drift findings). Both take
// the project root as their base dir (Audit appends ".moai/specs" internally;
// ListDocs does the same), so Config.ProjectRoot feeds both. The git-dependent
// drift-scan path is deliberately never invoked (REQ-WC11-045).
func (a *app) buildBoardView() (boardView, error) {
	records, err := spec.ListDocs(a.cfg.ProjectRoot)
	if err != nil {
		return boardView{}, err
	}
	auditRes, err := spec.Audit(spec.AuditOptions{BaseDir: a.cfg.ProjectRoot})
	if err != nil {
		return boardView{}, err
	}

	view := boardView{
		BindAddr:   a.resolveBindAddr(),
		TotalSpecs: len(records),
	}

	// Status distribution + close-debt (implemented-not-completed) list.
	counts := make(map[string]int, len(boardStatusOrder))
	for _, rec := range records {
		status := rec.Frontmatter.Status
		counts[status]++
		if status == "implemented" {
			view.CloseDebt = append(view.CloseDebt, boardSpec{
				ID:      boardSpecID(rec),
				Tier:    rec.Frontmatter.Tier,
				Title:   rec.Frontmatter.Title,
				Updated: rec.Frontmatter.Updated,
			})
		}
	}
	view.StatusCounts = orderedStatusCounts(counts)

	// MUST-FIX drift findings, in the deterministic order Audit returns them.
	for _, f := range auditRes.DriftFindings {
		if f.Severity == "MUST-FIX" {
			view.MustFix = append(view.MustFix, boardFinding{
				SpecID:      f.SpecID,
				FindingType: f.FindingType,
				Remediation: f.Remediation,
			})
		}
	}

	return view, nil
}

// boardSpecID returns the SPEC identifier for a record, falling back to the
// directory name when the frontmatter ID is absent (a malformed spec.md still
// contributes a stable identifier to the board).
func boardSpecID(rec spec.DocRecord) string {
	if rec.Frontmatter.ID != "" {
		return rec.Frontmatter.ID
	}
	return filepath.Base(filepath.Dir(rec.Path))
}

// orderedStatusCounts flattens the status→count map into the canonical enum
// order, appending any out-of-enum statuses (sorted) at the end. Empty status
// (e.g. an unparsed spec.md) is labeled "(unknown)".
func orderedStatusCounts(counts map[string]int) []boardStatusCount {
	out := make([]boardStatusCount, 0, len(counts))
	seen := make(map[string]bool, len(counts))
	for _, s := range boardStatusOrder {
		if c := counts[s]; c > 0 {
			out = append(out, boardStatusCount{Status: s, Count: c})
			seen[s] = true
		}
	}
	var extras []string
	for s := range counts {
		if !seen[s] {
			extras = append(extras, s)
		}
	}
	sort.Strings(extras)
	for _, s := range extras {
		label := s
		if label == "" {
			label = "(unknown)"
		}
		out = append(out, boardStatusCount{Status: label, Count: counts[s]})
	}
	return out
}

// boardCount renders an int count as a decimal string for the Templ board
// component (Templ interpolates only strings).
func boardCount(n int) string { return strconv.Itoa(n) }

// renderBoard renders the board Templ component into a buffer first, so a render
// error surfaces as a readable inline error (REQ-WC-010 discipline) rather than a
// half-written 200.
func (a *app) renderBoard(w http.ResponseWriter, status int, view boardView) {
	var buf bytes.Buffer
	if err := boardPage(view).Render(context.Background(), &buf); err != nil {
		a.renderError(w, http.StatusInternalServerError, "internal error: board render failed: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}
