// todo_triage.go — `moai todo triage <id>...` (card t943).
//
// The verb answers a question that precedes dispatch: does this card's premise
// still hold in the tree as it stands now? It prints MECHANICAL OBSERVATIONS —
// what a handful of git probes saw — and says, for each one, what it does and
// does not establish. It reaches NO VERDICT. The reader decides.
//
// Everything it renders is a measurement of the current tree at one ref. There
// is no model call, no network beyond git's own local reads, no score and no
// accuracy claim: a judgment layer was measured for this card and did not clear
// its base rate, so what ships is the observation construction alone.
//
// [HARD] READ-ONLY: the verb mutates no card, field, finding, timestamp or
// schema, performs no migration, and takes no queue mutation lock. It reads
// through the observational path (newTodoReadStore + LoadPure), never the
// adopting one. TestTodoTriage_QueueUnchanged asserts the queue file is
// byte-identical across a run.
//
// SUBPROCESS BOUND, per invocation over N cards: at most 2 + 13*N.
// The 2 are cached across the whole run — one positive control and one tree
// listing, taken at most once each. The 13 are per card: one commit search for
// the card id, plus at most 4 symbols x 3 probes (presence, neighborhood,
// provenance). A path-like symbol spends nothing on its neighborhood, because
// the cached tree listing already answers it, so 13 is a ceiling rather than a
// count. Every subprocess goes through todoRunCommand, the package seam the
// census tests count through.
//
// SUBAGENT BOUNDARY: nothing here prompts.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/spf13/cobra"
)

const (
	// todoTriageMaxSymbols bounds the extracted set. Four is what the
	// measured extractor kept, and it bounds the per-card subprocess spend.
	todoTriageMaxSymbols = 4

	// todoTriageControlPattern is the positive control: a string every Go
	// repository carries in quantity. Its job is to fail loudly when the
	// probe itself is broken, so a 0-hit presence result below can be read as
	// a real absence rather than as a broken probe.
	todoTriageControlPattern = "func "

	// todoTriageControlFloor is the count under which the control is treated
	// as evidence the probe is broken rather than as evidence about the tree.
	todoTriageControlFloor = 5

	// Render caps. Each is a ceiling on what is SHOWN; the true count is
	// always stated alongside, because a silently truncated listing reads as
	// a complete one.
	todoTriageDirEntryCap   = 20
	todoTriageNearNameCap   = 10
	todoTriageIdentHitCap   = 5
	todoTriageProvenanceCap = 5

	// todoTriageStemFloor is the shortest filename stem worth searching for
	// near-name siblings. Below it the stem matches half the tree.
	todoTriageStemFloor = 4

	// todoTriageIdentPrefix is how much of an identifier the case-insensitive
	// neighborhood search uses — enough to catch a renamed sibling, short
	// enough that a renamed sibling still matches.
	todoTriageIdentPrefix = 8
)

// The three extraction patterns, applied IN ORDER. They are the ones the
// experiment measured; the order is part of the measurement, because it
// decides which four survive the cap.
var (
	todoTriageBacktickPattern = regexp.MustCompile("`([A-Za-z_][A-Za-z0-9_./-]{3,})`")
	todoTriagePathPattern     = regexp.MustCompile(`\b([a-z_]+/[a-z0-9_./-]+\.(?:go|md|yaml|sh))\b`)
	todoTriageCallPattern     = regexp.MustCompile(`\b([a-z][A-Za-z0-9]{6,})\(`)
)

// newTodoTriageCmd — `moai todo triage <id>...`: render the mechanical
// observations that bear on each named card's premise. Read-only, fail-open.
func newTodoTriageCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "triage <id>...",
		Short: "Render mechanical observations bearing on each card's premise (read-only)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			refs := make([]string, 0, len(args))
			for _, a := range args {
				refs = append(refs, normalizeTodoRef(a))
			}
			return runTodoTriage(cmd, refs, jsonOutput)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false,
		"Emit the observations as JSON on stdout")
	withResolvedLandedRef(cmd, func(landedRef string) {
		cmd.Long = todoTriageLong(landedRef)
	})
	return cmd
}

// todoTriageLong renders the help body against the ref the probes will
// actually be run at. Resolved lazily — see withResolvedLandedRef.
func todoTriageLong(landedRef string) string {
	return `Print, for each named card, the mechanical observations that bear on whether
its premise still holds in the tree at ` + landedRef + `.

The verb OBSERVES. It reaches no verdict, assigns no score, and makes no
accuracy claim — every section states what it does and does not establish, and
the reader decides from the card's own wording.

It changes no card, field, finding, timestamp or schema, performs no migration,
and takes no queue mutation lock.

Five sections per card:

  symbols       identifiers pulled out of the card text by three regexes —
                backtick-quoted, path-like, and call-shaped. An empty result is
                a limit of the extractor, NOT a finding about the card
  control       a probe run before any presence result, so a later 0-hit can be
                read as a real absence rather than as a broken probe
  presence      per symbol, how many files at ` + landedRef + ` carry it. A zero
                hit carries BOTH readings, because nothing here decides which
                one applies
  neighborhood  what sits around an absent symbol: the directory's entries, or
                near-name siblings elsewhere. "absent, and nothing nearby
                delivers it" and "absent, but a sibling under another name sits
                right there" are opposite conclusions
  provenance    commits touching the symbol, and commits naming the card id

Every failure degrades rather than aborts: a section that could not be measured
says so and is never rendered as a zero, the remaining sections still print, a
note goes to stderr, and the exit code stays 0. An id that is not in the queue
prints one line and the next id is processed.`
}

// --- JSON shapes -------------------------------------------------------------

// todoTriageMeasure is the measured/unmeasured discriminator carried by every
// section. It is a distinct field rather than a sentinel value because "could
// not measure" and "measured zero" are different facts, and collapsing them is
// how an unanswerable probe comes to read as an absence.
type todoTriageMeasure struct {
	Measured bool `json:"measured"`
	// Reason is present only when Measured is false.
	Reason string `json:"reason,omitempty"`
}

type todoTriageControl struct {
	todoTriageMeasure
	// Files is the control's file count, meaningful only when Measured.
	Files int `json:"files"`
	// ProbeLooksBroken is set when the control came back at or under the
	// floor — the presence results below cannot be trusted.
	ProbeLooksBroken bool `json:"probe_looks_broken"`
}

type todoTriagePresence struct {
	todoTriageMeasure
	Files int `json:"files"`
}

type todoTriageNeighborhood struct {
	todoTriageMeasure
	// Kind is "path" or "identifier" — which neighborhood question was asked.
	Kind string `json:"kind"`
	// DirExists is meaningful for Kind=="path" only.
	DirExists  bool     `json:"dir_exists"`
	DirEntries []string `json:"dir_entries,omitempty"`
	DirCount   int      `json:"dir_count"`
	// NearNames holds near-name siblings (path kind) or case-insensitive
	// prefix hits (identifier kind).
	NearNames []string `json:"near_names,omitempty"`
	NearCount int      `json:"near_count"`
}

type todoTriageCommit struct {
	SHA     string `json:"sha"`
	Subject string `json:"subject"`
	// AncestorOfRef is true by construction: the listing walked history FROM
	// the ref, so every commit it emitted is reachable from it. Carried as a
	// field so a JSON consumer reads the same fact the text states.
	AncestorOfRef bool `json:"ancestor_of_ref"`
}

type todoTriageProvenance struct {
	todoTriageMeasure
	Commits []todoTriageCommit `json:"commits"`
}

type todoTriageSymbol struct {
	Symbol       string                 `json:"symbol"`
	Kind         string                 `json:"kind"`
	Presence     todoTriagePresence     `json:"presence"`
	Neighborhood todoTriageNeighborhood `json:"neighborhood"`
	Provenance   todoTriageProvenance   `json:"provenance"`
}

type todoTriageCard struct {
	ID    string `json:"id"`
	Ref   string `json:"ref"`
	Found bool   `json:"found"`
	Text  string `json:"text,omitempty"`
	// Symbols is the extracted set, possibly empty.
	Symbols []string             `json:"symbols"`
	Control todoTriageControl    `json:"control"`
	Details []todoTriageSymbol   `json:"details,omitempty"`
	CardID  todoTriageProvenance `json:"card_id_commits"`
}

// --- engine ------------------------------------------------------------------

// todoTriageEngine holds the per-run caches. The control and the tree listing
// are taken AT MOST ONCE for the whole invocation, so N cards do not repeat
// them N times — see the subprocess bound in the file header.
type todoTriageEngine struct {
	ref string

	controlOnce bool
	control     todoTriageControl

	treeOnce     bool
	treePaths    []string
	treeMeasured bool
	treeReason   string
}

// todoTriageRun runs one git subprocess through the package seam.
//
// Exit status 1 is treated as a MEASURED empty result rather than a failure:
// `git grep` and `git ls-files` report "no match" that way, and reading it as
// an error would turn every genuine absence into an unmeasured section — the
// exact collapse this verb exists to avoid. Anything else (128 for a bad ref,
// 127 for no git, a timeout) is unmeasured and says why.
func todoTriageRun(args ...string) (out string, measured bool, reason string) {
	o, err := todoRunCommand("git", args...)
	if err == nil {
		return o, true, ""
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() == 1 {
		return "", true, ""
	}
	return "", false, err.Error()
}

// positiveControl answers, once per run, whether the probe works at all.
func (e *todoTriageEngine) positiveControl() todoTriageControl {
	if e.controlOnce {
		return e.control
	}
	e.controlOnce = true
	out, measured, reason := todoTriageRun("grep", "-c", "--", todoTriageControlPattern, e.ref)
	if !measured {
		e.control = todoTriageControl{todoTriageMeasure: todoTriageMeasure{Reason: reason}}
		return e.control
	}
	n := len(todoTriageLines(out))
	e.control = todoTriageControl{
		todoTriageMeasure: todoTriageMeasure{Measured: true},
		Files:             n,
		ProbeLooksBroken:  n <= todoTriageControlFloor,
	}
	return e.control
}

// tree lists every path at the ref, once per run. Taken lazily: a run whose
// cards extract no path-like symbol never spends the subprocess.
func (e *todoTriageEngine) tree() ([]string, bool, string) {
	if e.treeOnce {
		return e.treePaths, e.treeMeasured, e.treeReason
	}
	e.treeOnce = true
	out, measured, reason := todoTriageRun("ls-tree", "-r", "--name-only", e.ref)
	e.treeMeasured, e.treeReason = measured, reason
	if measured {
		e.treePaths = todoTriageLines(out)
	}
	return e.treePaths, e.treeMeasured, e.treeReason
}

// presence counts the files at the ref carrying sym.
func (e *todoTriageEngine) presence(sym string) todoTriagePresence {
	out, measured, reason := todoTriageRun("grep", "-c", "--", sym, e.ref)
	if !measured {
		return todoTriagePresence{todoTriageMeasure: todoTriageMeasure{Reason: reason}, Files: 0}
	}
	return todoTriagePresence{
		todoTriageMeasure: todoTriageMeasure{Measured: true},
		Files:             len(todoTriageLines(out)),
	}
}

// neighborhood answers "what sits around this symbol" — the discriminator
// between "absent, and nothing nearby delivers it" and "absent, but a sibling
// under another name sits right there".
func (e *todoTriageEngine) neighborhood(sym string) todoTriageNeighborhood {
	if todoTriageIsPathLike(sym) {
		return e.pathNeighborhood(sym)
	}
	return e.identifierNeighborhood(sym)
}

// pathNeighborhood reads the CACHED tree listing — no subprocess of its own.
func (e *todoTriageEngine) pathNeighborhood(sym string) todoTriageNeighborhood {
	n := todoTriageNeighborhood{Kind: "path"}
	paths, measured, reason := e.tree()
	if !measured {
		n.Reason = reason
		return n
	}
	n.Measured = true

	dir := path.Dir(sym)
	prefix := ""
	if dir != "." && dir != "/" {
		prefix = strings.TrimSuffix(dir, "/") + "/"
	}
	stem := todoTriageStem(sym)
	seenDir := map[string]bool{}
	for _, p := range paths {
		if prefix == "" || strings.HasPrefix(p, prefix) {
			rest := strings.TrimPrefix(p, prefix)
			// Direct children only — a deep subtree is a different question.
			if !strings.Contains(rest, "/") && rest != "" {
				if !seenDir[rest] {
					seenDir[rest] = true
					n.DirCount++
					if len(n.DirEntries) < todoTriageDirEntryCap {
						n.DirEntries = append(n.DirEntries, rest)
					}
				}
			}
		}
	}
	n.DirExists = n.DirCount > 0

	if len(stem) >= todoTriageStemFloor {
		lowered := strings.ToLower(stem)
		// SAME EXTENSION ONLY. The stem match is a substring test, so a short
		// common stem drags in everything that merely CONTAINS it: measured
		// live on `internal/cli/update/merge/base.go` against origin/develop
		// (10,887 files), the stem `base` matched 55 paths, of which the ten
		// actually rendered were `codebase-analysis.md`, `supabase.md` and a
		// run of `*-baseline.*` — ten rows of chaff and not one real
		// neighbour, because dot-directories sort first and the cap fell
		// entirely inside them.
		//
		// A sibling delivering the same thing under another name is a sibling
		// of the SAME KIND, so the extension is the discriminator that removes
		// the noise where the noise is (55 -> 19, every dropped row a
		// .md/.json/.txt) while barely touching listings that were already
		// clean (internal/kanban/backlog.go 32 -> 30,
		// internal/glmcred/glmcred.go 1 -> 1).
		//
		// Path-like symbols only — the identifier prefix search is a different
		// question and is unaffected.
		wantExt := strings.ToLower(path.Ext(sym))
		for _, p := range paths {
			if prefix != "" && strings.HasPrefix(p, prefix) {
				continue // already reported as a directory entry
			}
			if strings.ToLower(path.Ext(p)) != wantExt {
				continue
			}
			if !strings.Contains(strings.ToLower(path.Base(p)), lowered) {
				continue
			}
			n.NearCount++
			if len(n.NearNames) < todoTriageNearNameCap {
				n.NearNames = append(n.NearNames, p)
			}
		}
	}
	return n
}

// identifierNeighborhood asks a case-insensitive question about the
// identifier's leading characters — one subprocess.
func (e *todoTriageEngine) identifierNeighborhood(sym string) todoTriageNeighborhood {
	n := todoTriageNeighborhood{Kind: "identifier"}
	prefix := sym
	if len(prefix) > todoTriageIdentPrefix {
		prefix = prefix[:todoTriageIdentPrefix]
	}
	out, measured, reason := todoTriageRun("grep", "-i", "-l", "--", prefix, e.ref)
	if !measured {
		n.Reason = reason
		return n
	}
	n.Measured = true
	hits := todoTriageLines(out)
	n.NearCount = len(hits)
	for _, h := range hits {
		if len(n.NearNames) >= todoTriageIdentHitCap {
			break
		}
		n.NearNames = append(n.NearNames, h)
	}
	return n
}

// provenance lists commits matching one `git log` search at the ref.
func (e *todoTriageEngine) provenance(searchFlag, needle string) todoTriageProvenance {
	out, measured, reason := todoTriageRun("log", "--oneline", "-"+strconv.Itoa(todoTriageProvenanceCap),
		searchFlag+needle, e.ref)
	if !measured {
		return todoTriageProvenance{todoTriageMeasure: todoTriageMeasure{Reason: reason}}
	}
	p := todoTriageProvenance{todoTriageMeasure: todoTriageMeasure{Measured: true}}
	for _, line := range todoTriageLines(out) {
		sha, subject, _ := strings.Cut(strings.TrimSpace(line), " ")
		p.Commits = append(p.Commits, todoTriageCommit{
			SHA: sha, Subject: subject,
			// By construction: the walk started at the ref.
			AncestorOfRef: true,
		})
	}
	return p
}

// --- extraction --------------------------------------------------------------

// todoTriageSymbols pulls candidate identifiers out of a card text.
//
// Three patterns in order, de-duplicated, first todoTriageMaxSymbols kept. The
// extractor is deliberately narrow, and its narrowness is REPORTED rather than
// hidden: an empty result says only that these three shapes found nothing.
func todoTriageSymbols(text string) []string {
	var out []string
	seen := map[string]bool{}
	for _, re := range []*regexp.Regexp{
		todoTriageBacktickPattern, todoTriagePathPattern, todoTriageCallPattern,
	} {
		for _, m := range re.FindAllStringSubmatch(text, -1) {
			s := m[1]
			if seen[s] {
				continue
			}
			seen[s] = true
			out = append(out, s)
			if len(out) == todoTriageMaxSymbols {
				return out
			}
		}
	}
	return out
}

// todoTriageIsPathLike reports whether sym reads as a file path rather than as
// a bare identifier — which decides WHICH neighborhood question is asked.
func todoTriageIsPathLike(sym string) bool {
	return strings.Contains(sym, "/") || strings.Contains(path.Base(sym), ".")
}

// todoTriageStem is the base name with its extension removed.
func todoTriageStem(sym string) string {
	base := path.Base(sym)
	if ext := path.Ext(base); ext != "" {
		base = strings.TrimSuffix(base, ext)
	}
	return base
}

func todoTriageLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

// --- run ---------------------------------------------------------------------

// runTodoTriage renders the observations. Every exit path is exit 0 unless the
// queue itself is unreadable.
func runTodoTriage(cmd *cobra.Command, ids []string, jsonOutput bool) error {
	// REQ-BJD-002 — probed before the read (todo_disclosure.go).
	_ = discloseQueueLayout(cmd, "triage")
	rec, err := newTodoReadStore().LoadPure()
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}

	eng := &todoTriageEngine{ref: todoLandedRef()}
	cards := make([]todoTriageCard, 0, len(ids))
	for _, id := range ids {
		cards = append(cards, eng.observe(cmd, rec, id))
	}

	out := cmd.OutOrStdout()
	if jsonOutput {
		data, err := json.Marshal(cards)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(out, string(data))
		return nil
	}
	for i, c := range cards {
		if i > 0 {
			_, _ = fmt.Fprintln(out)
		}
		renderTodoTriageCard(out, c)
	}
	return nil
}

// observe builds one card's observation block, spending the bounded probes.
func (e *todoTriageEngine) observe(cmd *cobra.Command, rec *kanban.BacklogRecord, id string) todoTriageCard {
	card := todoTriageCard{ID: id, Ref: e.ref, Symbols: []string{}}
	for _, it := range rec.Items {
		if it.ID == id {
			card.Found = true
			card.Text = it.Text
			break
		}
	}
	if !card.Found {
		return card
	}

	card.Symbols = todoTriageSymbols(card.Text)
	card.Control = e.positiveControl()
	if !card.Control.Measured {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"note: the positive control could not be measured at %s (%s); presence results are reported as unmeasured, never as absences\n",
			e.ref, card.Control.Reason)
	}
	for _, sym := range card.Symbols {
		kind := "identifier"
		if todoTriageIsPathLike(sym) {
			kind = "path"
		}
		card.Details = append(card.Details, todoTriageSymbol{
			Symbol:       sym,
			Kind:         kind,
			Presence:     e.presence(sym),
			Neighborhood: e.neighborhood(sym),
			Provenance:   e.provenance("-S", sym),
		})
	}
	card.CardID = e.provenance("--grep=", id)
	return card
}

// --- render ------------------------------------------------------------------

// renderTodoTriageCard writes one card's block. The framing sentences are the
// substance of this verb, not decoration: without them a count is an invitation
// to conclude something the count does not support.
func renderTodoTriageCard(out io.Writer, c todoTriageCard) {
	p := func(format string, a ...any) { _, _ = fmt.Fprintf(out, format+"\n", a...) }

	if !c.Found {
		p("%s\tnot in the queue — nothing to observe", c.ID)
		return
	}
	p("%s\t%s", c.ID, todoPRCell(c.Text))
	p("  ref: %s", c.Ref)

	// 1. Symbols.
	if len(c.Symbols) == 0 {
		p("  symbols: none extracted.")
		p("    This is a LIMIT OF THE EXTRACTOR, which catches only backtick-quoted")
		p("    identifiers, path-like literals and call-shaped tokens. It is NOT evidence")
		p("    that the card names no real defect, and NOT evidence that the defect is")
		p("    dead. Weigh the card's own prose instead.")
		return
	}
	p("  symbols: %s", strings.Join(c.Symbols, " "))

	// 2. Positive control — printed BEFORE any presence result.
	switch {
	case !c.Control.Measured:
		p("  control: could not measure (%s)", c.Control.Reason)
		p("    The presence results below could not be trusted, and are reported as")
		p("    unmeasured rather than as absences.")
	case c.Control.ProbeLooksBroken:
		p("  control: %q matched %d files — the probe looks BROKEN", todoTriageControlPattern, c.Control.Files)
		p("    A near-zero control means the probe itself is not working, so the presence")
		p("    results below cannot be trusted either way.")
	default:
		p("  control: %q matched %d files — the probe works", todoTriageControlPattern, c.Control.Files)
		p("    So a 0-hit presence result below is a real absence, not a broken probe.")
	}

	for _, d := range c.Details {
		p("  %s (%s)", d.Symbol, d.Kind)
		renderTodoTriagePresence(p, d, c.Control)
		renderTodoTriageNeighborhood(p, d)
		renderTodoTriageProvenance(p, "    commits touching it", d.Provenance, c.Ref)
	}

	renderTodoTriageProvenance(p, "  commits naming "+c.ID, c.CardID, c.Ref)
	p("    A delivering commit often carries a SIBLING card's id, so an empty search")
	p("    here is NOT evidence the work was never done. And because the ref is the")
	p("    CURRENT integration ref, a commit naming this card id may be this card's own")
	p("    earlier work rather than a prior sibling's delivery.")
}

func renderTodoTriagePresence(p func(string, ...any), d todoTriageSymbol, ctrl todoTriageControl) {
	if !d.Presence.Measured {
		p("    presence: could not measure (%s) — this is NOT a zero", d.Presence.Reason)
		return
	}
	if d.Presence.Files > 0 {
		p("    presence: %d files (control %d)", d.Presence.Files, ctrl.Files)
		return
	}
	p("    presence: 0 files (control %d)", ctrl.Files)
	p("      Two readings, and nothing here decides between them:")
	p("      - if the card claims this is MISSING (not built, not wired, never")
	p("        delivered), zero hits is consistent with the defect being ALIVE, and is")
	p("        not evidence it is dead;")
	p("      - if the card claims this EXISTS BUT IS WRONG or STALE, zero hits suggests")
	p("        the named thing was renamed or removed, so the card may be pointing at")
	p("        something that no longer exists under that name.")
	p("      This tool does not decide which case applies. You do, from the card's")
	p("      own wording.")
}

func renderTodoTriageNeighborhood(p func(string, ...any), d todoTriageSymbol) {
	n := d.Neighborhood
	if !n.Measured {
		p("    neighborhood: could not measure (%s) — this is NOT an absence", n.Reason)
		return
	}
	if n.Kind == "path" {
		if !n.DirExists {
			p("    neighborhood: the directory %s does not exist at this ref", path.Dir(d.Symbol))
		} else {
			p("    neighborhood: %d entries in %s%s", n.DirCount, path.Dir(d.Symbol), todoTriageShownSuffix(n.DirCount, len(n.DirEntries)))
			for _, e := range n.DirEntries {
				p("      - %s", e)
			}
		}
		if n.NearCount == 0 {
			p("    near names elsewhere: none")
		} else {
			p("    near names elsewhere: %d%s", n.NearCount, todoTriageShownSuffix(n.NearCount, len(n.NearNames)))
			for _, e := range n.NearNames {
				p("      - %s", e)
			}
		}
	} else {
		if n.NearCount == 0 {
			p("    neighborhood: a case-insensitive search on its first %d characters found nothing", todoTriageIdentPrefix)
		} else {
			p("    neighborhood: %d files match its first %d characters, case-insensitively%s",
				n.NearCount, todoTriageIdentPrefix, todoTriageShownSuffix(n.NearCount, len(n.NearNames)))
			for _, e := range n.NearNames {
				p("      - %s", e)
			}
		}
	}
	p("      \"absent, and nothing nearby delivers it\" and \"absent, but a sibling under")
	p("      another name sits right there\" are OPPOSITE conclusions.")
}

func renderTodoTriageProvenance(p func(string, ...any), label string, pv todoTriageProvenance, ref string) {
	if !pv.Measured {
		p("%s: could not measure (%s) — this is NOT an empty history", label, pv.Reason)
		return
	}
	if len(pv.Commits) == 0 {
		p("%s: none", label)
		return
	}
	p("%s: %d (each one an ancestor of %s — the listing walked history from it)", label, len(pv.Commits), ref)
	for _, c := range pv.Commits {
		p("      - %s %s", c.SHA, c.Subject)
	}
}

// todoTriageShownSuffix states the truncation when one happened, so a capped
// listing is never mistaken for a complete one.
func todoTriageShownSuffix(total, shown int) string {
	if shown >= total {
		return ""
	}
	return fmt.Sprintf(" (showing %d)", shown)
}
