// backlog_issuance.go — SPEC-TODO-CARD-ISSUANCE-001 M1: the read-only
// issuance presentation lookup. Deliberately separate from the duplicate
// classifier (backlog_analysis.go): the classifier skips dropped cards
// because a dropped card is not a refusal subject, while the presentation
// WANTS dropped and archived cards as notice targets. Only the measures
// (NormalizeCardText, TokenSetJaccard) are shared — ClassifyCardText is
// never called here (REQ-TCI-006; AC-TCI-007 pins the classifier file
// byte-unchanged).
//
// Every function here is pure over its inputs: the caller reads a LoadPure
// snapshot OUTSIDE the queue's cross-process write lock and supplies the
// completed-SPEC entries and lane-branch probes through parameters, so this
// layer never touches the filesystem and never waits on the lock
// (REQ-TCI-003).
package factory

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Values read from the tracked baseline record (REQ-TCI-001):
// .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/baseline.md, M0
// remeasurement 2026-10-05. The SPEC prose fixes no numbers (D7) — when the
// baseline is remeasured, these constants are the edit.
const (
	// IssuanceNeighborLimit — baseline threshold table, similar-card cap
	// (card-fixed value 3).
	IssuanceNeighborLimit = 3
	// IssuanceDisplayFloor — baseline threshold table, token-set Jaccard
	// display floor 0.30 (SB03: alerts 8.1% overall / 4.1% recent window;
	// recorded-relation recall at 0.3 is 5%, SB05 — the floor trades recall
	// for noise, deliberately).
	IssuanceDisplayFloor = 0.30
	// IssuanceComponentDepth — baseline threshold table: a component key is
	// the first two segments of a repo path (depth-3 keys shared 0 pairs
	// across 741 open-card pairs, SB07).
	IssuanceComponentDepth = 2
	// IssuanceProbeTimeBound — per-probe wall-clock cap for reads outside the
	// queue (lane-branch git diffs, the SPEC directory). The card's tree
	// measured ~0.25s per lane-branch diff (SB08); one order of magnitude of
	// headroom keeps a slow probe from delaying admission while still
	// finishing inside any human-noticeable window. A probe that exceeds the
	// bound degrades its item to `unmeasured (time bound)` (REQ-TCI-003).
	IssuanceProbeTimeBound = 2 * time.Second
)

// IssuanceDropPrefix is the drop convention the queue writes: the stored
// text of a dropped card begins with "[DROPPED — <reason>] ".
const IssuanceDropPrefix = "[DROPPED — "

// IssuanceNeighbor is one vocabulary-similarity notice target: a live,
// dropped or archived card whose token-set Jaccard score against the
// candidate reaches the display floor, or whose normalized text equals the
// candidate exactly (Exact — shown regardless of the floor; the most certain
// item the presentation has, and one ClassifyCardText cannot see for
// non-live states).
type IssuanceNeighbor struct {
	ID     string
	State  string // "live" | "dropped" | "archived"
	Score  float64
	Text   string // stored text with the drop prefix stripped
	Reason string // drop reason for a dropped card, "" otherwise
	Exact  bool   // normalized text equal — always shown, rendered label "exact"
}

// IssuanceStripDropPrefix splits a stored card text into the comparison body
// and the drop reason. A text without the prefix is returned unchanged with
// dropped=false.
func IssuanceStripDropPrefix(text string) (stripped, reason string, dropped bool) {
	if !strings.HasPrefix(text, IssuanceDropPrefix) {
		return text, "", false
	}
	rest := text[len(IssuanceDropPrefix):]
	close := strings.Index(rest, "]")
	if close < 0 {
		return text, "", false
	}
	return strings.TrimSpace(rest[close+1:]), strings.TrimSpace(rest[:close]), true
}

// IssuanceNeighbors returns up to IssuanceNeighborLimit vocabulary neighbors
// of candidate across live, dropped (prefix stripped, reason kept) and
// archived cards at the baseline display floor, sorted most-similar first
// with exact matches outranking scored ones.
func IssuanceNeighbors(candidate string, items []BacklogItem, archived []BacklogArchiveEntry) []IssuanceNeighbor {
	return IssuanceNeighborsFloor(candidate, items, archived, IssuanceDisplayFloor)
}

// IssuanceNeighborsFloor is IssuanceNeighbors with an explicit floor; a
// floor below 0 lifts it (the --dry-run view shows the top-3 neighbors
// regardless of score, design §3.2). Below-floor cards are not notice
// targets; a normalized-equal card always is.
func IssuanceNeighborsFloor(candidate string, items []BacklogItem, archived []BacklogArchiveEntry, floor float64) []IssuanceNeighbor {
	cnorm := NormalizeCardText(candidate)
	type scored struct {
		n     IssuanceNeighbor
		exact bool
	}
	var found []scored
	consider := func(id, state, text string) {
		stripped, reason, _ := IssuanceStripDropPrefix(text)
		other := NormalizeCardText(stripped)
		exact := cnorm != "" && cnorm == other
		score := TokenSetJaccard(cnorm, other)
		if !exact && score < floor {
			return
		}
		found = append(found, scored{
			n:     IssuanceNeighbor{ID: id, State: state, Score: score, Text: stripped, Reason: reason, Exact: exact},
			exact: exact,
		})
	}
	for _, it := range items {
		state := "live"
		if it.State == BacklogStateDropped {
			state = "dropped"
		}
		consider(it.ID, state, it.Text)
	}
	for _, entry := range archived {
		consider(entry.Item.ID, "archived", entry.Item.Text)
	}
	sort.SliceStable(found, func(i, j int) bool {
		if found[i].exact != found[j].exact {
			return found[i].exact
		}
		return found[i].n.Score > found[j].n.Score
	})
	if len(found) > IssuanceNeighborLimit {
		found = found[:IssuanceNeighborLimit]
	}
	out := make([]IssuanceNeighbor, 0, len(found))
	for _, s := range found {
		out = append(out, s.n)
	}
	return out
}

// issuancePathToken matches a whole whitespace-delimited token that spells a
// repo path with a known source extension. The match is over the WHOLE
// token, so "internal/cli/todo.go.orig" is one token that fails the
// extension test and never yields "internal/cli/todo.go" — the full-path
// equality contract of AC-TCI-003 (h) starts at extraction, not at
// comparison.
var issuancePathToken = regexp.MustCompile(
	`^[A-Za-z0-9_\-]+(?:/[A-Za-z0-9_.\-]+)+\.(?:go|md|yaml|yml|sh|js|json|toml|templ|txt)$`)

// IssuanceExtractPaths extracts the repo paths a card body names, read-only
// and in first-appearance order (D13: an inference that is never stored).
func IssuanceExtractPaths(text string) []string {
	seen := map[string]bool{}
	var out []string
	for _, tok := range strings.FieldsFunc(text, func(r rune) bool {
		keep := r == '/' || r == '.' || r == '-' || r == '_' ||
			(r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
		return !keep
	}) {
		if !issuancePathToken.MatchString(tok) || seen[tok] {
			continue
		}
		seen[tok] = true
		out = append(out, tok)
	}
	return out
}

// issuanceComponentKey returns the first IssuanceComponentDepth segments of
// a repo path ("internal/cli/todo.go" -> "internal/cli"), or "" when the
// path is shallower than the key depth.
func issuanceComponentKey(p string) string {
	parts := strings.SplitN(p, "/", IssuanceComponentDepth+1)
	if len(parts) < IssuanceComponentDepth {
		return ""
	}
	return strings.Join(parts[:IssuanceComponentDepth], "/")
}

// IssuanceComponent is an OPEN card (queued, picked, hold) sharing a
// component key with the candidate. Dropped cards are not component targets
// (open-only, like the classifier's refusal doctrine but for notice
// coverage) and a card naming no path has no key to share.
type IssuanceComponent struct {
	ID    string
	State string
	Key   string
}

// IssuanceSameComponent returns the open cards whose body paths share a
// depth-2 component key with any path the candidate names, in queue order.
func IssuanceSameComponent(candidate string, items []BacklogItem) []IssuanceComponent {
	keys := map[string]bool{}
	for _, p := range IssuanceExtractPaths(candidate) {
		if k := issuanceComponentKey(p); k != "" {
			keys[k] = true
		}
	}
	if len(keys) == 0 {
		return nil
	}
	var out []IssuanceComponent
	for _, it := range items {
		switch it.State {
		case BacklogStateQueued, BacklogStatePicked, BacklogStateHold:
		default:
			continue
		}
		for _, p := range IssuanceExtractPaths(it.Text) {
			k := issuanceComponentKey(p)
			if k != "" && keys[k] {
				out = append(out, IssuanceComponent{ID: it.ID, State: string(it.State), Key: k})
				break
			}
		}
	}
	return out
}

// IssuanceCompletedSpec is one parsed SPEC frontmatter entry the caller
// read. The factory layer never touches the filesystem — the CLI supplies
// the entries from its time-bounded directory read (REQ-TCI-003).
type IssuanceCompletedSpec struct {
	ID     string
	Status string
	Module string
	Title  string
	Tags   string
}

// IssuanceSpecMatch is a completed SPEC whose recorded surface (title, tags,
// module) measures similar to the candidate. The renderer labels it
// measure=spec-heuristic with the heuristic marker — this is a heuristic
// coverage hint, never a verdict.
type IssuanceSpecMatch struct {
	ID     string
	Status string
}

// IssuanceCompletedSpecCoverage returns the COMPLETED specs whose title,
// tags and module measure at or above the display floor against the
// candidate, or whose module path the candidate names. Draft and every other
// status is excluded.
func IssuanceCompletedSpecCoverage(candidate string, specs []IssuanceCompletedSpec) []IssuanceSpecMatch {
	cnorm := NormalizeCardText(candidate)
	candidatePaths := IssuanceExtractPaths(candidate)
	var out []IssuanceSpecMatch
	for _, s := range specs {
		if s.Status != "completed" {
			continue
		}
		score := TokenSetJaccard(cnorm, NormalizeCardText(s.Title+" "+s.Tags+" "+s.Module))
		namesModule := false
		for _, p := range candidatePaths {
			if p == s.Module {
				namesModule = true
				break
			}
		}
		if score >= IssuanceDisplayFloor || namesModule {
			out = append(out, IssuanceSpecMatch{ID: s.ID, Status: s.Status})
		}
	}
	return out
}

// IssuanceInFlightCard is one picked card holding a lane assignment together
// with its expected files: the body-named paths plus whatever the lane
// probe returned. Empty Files means the card carries NO input — it is
// excluded from the comparison and reported nowhere (the verdict is
// set-wise, REQ-TCI-002(c)'s terminology note).
type IssuanceInFlightCard struct {
	ID    string
	Lane  string
	Files []string
}

// IssuanceOverlapItem is one shared path between the candidate and one
// in-flight card.
type IssuanceOverlapItem struct {
	CardID string
	Lane   string
	Path   string
}

// IssuanceOverlap is the set-wise file-overlap verdict:
//   - Items — one entry per shared path (the positive case);
//   - None   — inputs existed on both sides and nothing was shared (a
//     measured no; the renderer writes NO line for it);
//   - Unmeasured — a reason string when the comparison could not be made
//     (candidate carries no files, no in-flight card carries files, or the
//     probe budget ran out: rendered as `unmeasured (<reason>)`).
type IssuanceOverlap struct {
	Items      []IssuanceOverlapItem
	None       bool
	Unmeasured string
}

// IssuanceInFlightOverlap compares the candidate's expected files against
// every in-flight card that carries input. Path equality is over the whole
// normalized path — file basenames and string prefixes are not overlaps
// (AC-TCI-003 (h)).
func IssuanceInFlightOverlap(candidateFiles []string, inFlight []IssuanceInFlightCard) IssuanceOverlap {
	// No in-flight cards at all — nothing the presentation could advise
	// about: silent (design §2 — the MCP parity fixture's empty queue keeps
	// an empty presentation). The unmeasured verdict below is for lanes that
	// EXIST but carry no comparable input.
	if len(inFlight) == 0 {
		return IssuanceOverlap{}
	}
	if len(candidateFiles) == 0 {
		return IssuanceOverlap{Unmeasured: "new card carries no expected files"}
	}
	cand := map[string]bool{}
	var candOrder []string
	for _, p := range candidateFiles {
		if !cand[p] {
			cand[p] = true
			candOrder = append(candOrder, p)
		}
	}
	comparable := 0
	var items []IssuanceOverlapItem
	for _, f := range inFlight {
		if len(f.Files) == 0 {
			continue
		}
		comparable++
		fset := map[string]bool{}
		for _, p := range f.Files {
			fset[p] = true
		}
		for _, p := range candOrder {
			if fset[p] {
				items = append(items, IssuanceOverlapItem{CardID: f.ID, Lane: f.Lane, Path: p})
			}
		}
	}
	if comparable == 0 {
		return IssuanceOverlap{Unmeasured: "no in-flight card carries expected files"}
	}
	if len(items) == 0 {
		return IssuanceOverlap{None: true}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CardID != items[j].CardID {
			return items[i].CardID < items[j].CardID
		}
		return items[i].Path < items[j].Path
	})
	return IssuanceOverlap{Items: items}
}

// LaneFilesProbe returns one in-flight card's lane-branch changed files. The
// production seam runs a time-bounded git read; tests inject fixed answers.
// ok=false means the probe produced no input (the card is excluded from the
// comparison).
type LaneFilesProbe func(cardID, lane string) (files []string, ok bool)

// CardClosedAt derives a card's closing time from the stamps the record
// already carries (D9 — the closing time is never stored): the archive
// stamp when the caller read one, else the drop stamp, else "unknown". One
// accessor, three answers.
func CardClosedAt(item BacklogItem, archivedAt *string) string {
	if archivedAt != nil {
		return *archivedAt
	}
	if item.DroppedAt != nil {
		return *item.DroppedAt
	}
	return "unknown"
}

// IssuanceDispositionValues is the closed disposition set a finding may
// carry (REQ-TCI-011): recording-only, operator verb only.
var IssuanceDispositionValues = []string{"accept", "merge", "reject"}

// RecordFindingDisposition sets the disposition of the finding naming the
// given pair (either order), through the locked whole-record write.
// Recording changes no card, no relation and no queue order (D11); a pair
// with no finding, or a value outside the closed set, is refused with
// nothing written.
func RecordFindingDisposition(store *BacklogStore, subject, related, disposition string) error {
	valid := false
	for _, v := range IssuanceDispositionValues {
		if v == disposition {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("disposition must be one of %s (got %q)",
			strings.Join(IssuanceDispositionValues, ", "), disposition)
	}
	return store.Mutate(func(rec *BacklogRecord) error {
		matched := 0
		for i := range rec.Findings {
			f := &rec.Findings[i]
			if f.Names(subject) && f.Names(related) {
				matched++
				v := disposition
				f.Disposition = &v
			}
		}
		if matched == 0 {
			return fmt.Errorf("no finding names %s and %s", subject, related)
		}
		return nil
	})
}

// IssuancePresentation is everything the renderer prints for one candidate.
// Empty means no item fired anywhere — the CLI prints nothing and the MCP
// result text stays the bare "<id> <pos>" line.
type IssuancePresentation struct {
	Neighbors  []IssuanceNeighbor
	Components []IssuanceComponent
	Overlap    IssuanceOverlap
	Specs      []IssuanceSpecMatch
}

// Empty reports whether the presentation carries no item at all.
func (p IssuancePresentation) Empty() bool {
	return len(p.Neighbors) == 0 && len(p.Components) == 0 &&
		len(p.Overlap.Items) == 0 && !p.Overlap.None && p.Overlap.Unmeasured == "" &&
		len(p.Specs) == 0
}

// BuildIssuancePresentation assembles the presentation at the baseline
// display floor, from a queue snapshot the caller already read outside the
// lock. leased carries the lane labels the factory record leases cards to
// (card id → lane) — the assignments the queue's runtime table carries do
// not include a factory lease (card t1454 card-review r2 finding 9).
func BuildIssuancePresentation(candidate string, rec *BacklogRecord, specs []IssuanceCompletedSpec, probe LaneFilesProbe, leased map[string]string) IssuancePresentation {
	return BuildIssuancePresentationFloor(candidate, rec, specs, probe, leased, IssuanceDisplayFloor)
}

// BuildIssuancePresentationFloor assembles the presentation with an explicit
// display floor (floor < 0 lifts it for the dry-run view). The probe runs
// per in-flight card under IssuanceProbeTimeBound; a probe that outlives the
// bound degrades the overlap verdict to `unmeasured (time bound)` — even a
// computed `none` was decided on partial input once a probe was lost — and
// admission is unaffected (REQ-TCI-003 (d)).
func BuildIssuancePresentationFloor(candidate string, rec *BacklogRecord, specs []IssuanceCompletedSpec, probe LaneFilesProbe, leased map[string]string, floor float64) IssuancePresentation {
	p := IssuancePresentation{
		Neighbors:  IssuanceNeighborsFloor(candidate, rec.Items, rec.Archived, floor),
		Components: IssuanceSameComponent(candidate, rec.Items),
	}
	// In-flight lane cards: picked with a lane-labelled runtime assignment
	// or a factory lease. The leader label is not a lane.
	laneOf := map[string]string{}
	for _, a := range rec.Runtime.Assignments {
		if a.CardID == "" || a.OwnerLabel == "" || a.OwnerLabel == "leader" {
			continue
		}
		laneOf[a.CardID] = a.OwnerLabel
	}
	for cardID, lane := range leased {
		if cardID == "" || lane == "" || lane == "leader" {
			continue
		}
		// The live lease is the current holder.
		laneOf[cardID] = lane
	}
	var inFlight []IssuanceInFlightCard
	probeTimedOut := false
	for _, it := range rec.Items {
		if it.State != BacklogStatePicked {
			continue
		}
		lane, ok := laneOf[it.ID]
		if !ok {
			continue
		}
		files := IssuanceExtractPaths(it.Text)
		if it.Issuance != nil {
			// The explicitly recorded expected files are comparison input
			// too (card t1454 card-review r2 finding 8) — the operator's
			// statement, ahead of any body-derived inference.
			files = append(files, it.Issuance.Files...)
		}
		if probe != nil && !probeTimedOut {
			done := make(chan struct{})
			var got []string
			var pok bool
			go func() {
				defer close(done)
				got, pok = probe(it.ID, lane)
			}()
			select {
			case <-done:
				if pok {
					files = append(files, got...)
				}
			case <-time.After(IssuanceProbeTimeBound):
				probeTimedOut = true
			}
		}
		inFlight = append(inFlight, IssuanceInFlightCard{ID: it.ID, Lane: lane, Files: files})
	}
	p.Overlap = IssuanceInFlightOverlap(IssuanceExtractPaths(candidate), inFlight)
	if probeTimedOut && len(p.Overlap.Items) == 0 {
		p.Overlap = IssuanceOverlap{Unmeasured: "time bound"}
	}
	p.Specs = IssuanceCompletedSpecCoverage(candidate, specs)
	return p
}
