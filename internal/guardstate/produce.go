package guardstate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/guardliveness"
)

// runListLimit caps how many run records a per-subject query asks for.
//
// It bounds the request rather than the answer: a subject firing on every push
// can have thousands of records and none of them older than the newest hundred
// changes any decision, because every row of the entry axis reads either the
// most recent qualifying run or whether any exists at all.
const runListLimit = 100

// Producer decides every declared entry and every undeclared workflow file into
// exactly one classification, and emits the result in the shape the surfacing
// half consumes.
//
// It is the implementation of guardliveness.Producer, which is the seam between
// the two SPECs: SPEC-GUARD-STATE-MODEL-001 owns everything that PRODUCES a
// result, SPEC-GUARD-LIVENESS-001 owns how one reaches a reader. The direction
// of the import says the same thing — this package knows the contract it
// satisfies, and the consuming package names no classification value.
type Producer struct {
	// querierFor builds the run-history reader for a root. A field rather than
	// a direct call so a test measures the classifier instead of a network.
	querierFor func(root string) RunQuerier

	// enumeratorFor builds the disk side for a root.
	enumeratorFor func(root string) Enumerator

	// now is the moment the measurement is taken against. Every window
	// comparison reads it, so it is an input rather than an annotation.
	now func() time.Time
}

// ProducerOption configures a Producer.
type ProducerOption func(*Producer)

// WithQuerierFactory replaces the run-history reader.
func WithQuerierFactory(f func(root string) RunQuerier) ProducerOption {
	return func(p *Producer) { p.querierFor = f }
}

// WithClock fixes the moment the measurement is taken against.
func WithClock(f func() time.Time) ProducerOption {
	return func(p *Producer) { p.now = f }
}

// NewProducer returns the production producer: the forge-backed run query and
// the on-disk workflow enumeration.
func NewProducer(opts ...ProducerOption) *Producer {
	p := &Producer{
		querierFor:    func(root string) RunQuerier { return newGHQuerier(root) },
		enumeratorFor: func(root string) Enumerator { return newDiskEnumerator(root) },
		now:           time.Now,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// ErrNoRoot reports an activation that named no tree to evaluate.
var ErrNoRoot = errors.New("guardstate: the activation names no root to evaluate")

// Produce evaluates the tree the activation names.
//
// A failure to READ the inputs is returned as an error rather than as an empty
// result, and the difference is the whole of this SPEC's own subject at its own
// layer: an empty result partitions to nothing non-clean, which the consumer
// renders as silence — an all-clear about a set nobody evaluated. The consumer
// persists only a refresh that answered, so an error here leaves the previous
// verdict standing with its age growing where the reader can see it.
func (p *Producer) Produce(ctx context.Context, act guardliveness.Activation) (guardliveness.Result, error) {
	ev, err := p.Evaluate(ctx, act.Root)
	if err != nil {
		return guardliveness.Result{}, err
	}
	return ev.ToResult(), nil
}

// Evaluate runs the state model over a tree and returns the full evaluation,
// counts included.
//
// Produce narrows this to the consumed contract, which carries entries and a
// designation and no counts. The counts are not dropped — they are read by this
// package's own guards (Refusals) before the narrowing, which is what
// AC-GSM-013 clause (b) requires of every one of them.
func (p *Producer) Evaluate(ctx context.Context, root string) (Evaluation, error) {
	if root == "" {
		return Evaluation{}, ErrNoRoot
	}

	manifest, err := LoadManifest(filepath.Join(root, filepath.FromSlash(ManifestPath)))
	if err != nil {
		return Evaluation{}, err
	}

	return Evaluate(ctx, manifest, p.enumeratorFor(root), p.querierFor(root), p.now()), nil
}

// ToResult narrows an evaluation to the published contract.
//
// Three clauses are discharged here and each is a shape the consumer refuses
// loudly rather than mis-partitioning: exactly one classification per entry,
// exactly one designated clean value, and the designation carried in
// machine-readable form so a consumer that has never heard of this vocabulary
// can still partition.
func (e Evaluation) ToResult() guardliveness.Result {
	entries := make([]guardliveness.Entry, 0, len(e.Decisions))
	for _, d := range e.Decisions {
		entries = append(entries, guardliveness.Entry{
			Subject:         d.Subject,
			Classifications: []string{string(d.Class)},
			Surface:         string(d.Surface),
			Expectation:     d.Expectation.published(),
		})
	}
	return guardliveness.Result{
		Entries: entries,
		Clean:   &guardliveness.Designation{Values: []string{string(cleanValue())}},
	}
}

// cleanValue is the single value the vocabulary designates clean, derived from
// the vocabulary rather than written down twice.
//
// Derived because a second literal is a second place to be wrong: designating a
// value here that IsClean does not agree with would break the seam silently —
// every criterion in the consuming SPEC would still pass while its advisory
// under-fired.
func cleanValue() Classification {
	for _, c := range Classifications() {
		if c.IsClean() {
			return c
		}
	}
	// Unreachable while exactly one value is clean. Returning the empty value
	// designates nothing that any entry carries, so the consumer reports every
	// entry rather than silently reporting none.
	return ""
}

// published converts a declared expectation to its carried form, or nil when
// the entry missed nothing. nil rather than a blank struct: "nothing was
// missed" must not be readable as "the expectation was blank".
func (e *Expectation) published() *guardliveness.Expectation {
	if e == nil {
		return nil
	}
	return &guardliveness.Expectation{Window: e.Window, Measure: string(e.Measure)}
}

// ---------------------------------------------------------------------------
// The disk side.
// ---------------------------------------------------------------------------

// workflowDir is where a github-workflow subject's file lives, relative to the
// repository root.
const workflowDir = ".github/workflows"

// diskEnumerator answers the two disk questions the state table asks, and
// answers them INDEPENDENTLY.
//
// Enumerate walks a pattern; Exists tests one named path. The independence is
// the whole of the enumeration-integrity gate: an Exists that consulted the
// enumeration would inherit whatever the pattern got wrong, and the
// corroboration would agree with the defect it exists to catch.
type diskEnumerator struct{ root string }

func newDiskEnumerator(root string) *diskEnumerator { return &diskEnumerator{root: root} }

// Enumerate returns every workflow file under the root, as repository-relative
// locators — the form the manifest declares. Returning absolute paths would
// make every declared entry read as absent from the enumeration.
func (d *diskEnumerator) Enumerate() ([]string, error) {
	dir := filepath.Join(d.root, filepath.FromSlash(workflowDir))
	names, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("guardstate: enumerate %s: %w", workflowDir, err)
	}

	var out []string
	for _, n := range names {
		if n.IsDir() {
			continue
		}
		ext := filepath.Ext(n.Name())
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		out = append(out, workflowDir+"/"+n.Name())
	}
	sort.Strings(out)
	return out, nil
}

// Exists is the per-subject point test: one stat of one named locator, reaching
// neither the enumeration nor its pattern.
func (d *diskEnumerator) Exists(locator string) (bool, error) {
	full, err := d.resolve(locator)
	if err != nil {
		return false, err
	}
	switch _, statErr := os.Stat(full); {
	case statErr == nil:
		return true, nil
	case errors.Is(statErr, os.ErrNotExist):
		return false, nil
	default:
		return false, fmt.Errorf("guardstate: point test %q: %w", locator, statErr)
	}
}

// resolve keeps a locator inside the tree it was given.
//
// A locator is data read from a file, so it is not trusted to stay within the
// root: a point test that resolved one outside would answer about a subject the
// enumeration could never have seen, and its agreement would mean nothing.
func (d *diskEnumerator) resolve(locator string) (string, error) {
	root, err := filepath.Abs(d.root)
	if err != nil {
		return "", fmt.Errorf("guardstate: resolve the root: %w", err)
	}
	full := filepath.Clean(filepath.Join(root, filepath.FromSlash(locator)))
	if full != root && !strings.HasPrefix(full, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("guardstate: locator %q resolves outside the evaluated tree", locator)
	}
	return full, nil
}

// ---------------------------------------------------------------------------
// The forge side.
// ---------------------------------------------------------------------------

// ghQuerier reads run history through the gh CLI.
//
// Every call it makes is a LISTING. Nothing on this type creates, comments,
// closes, dispatches, re-runs or cancels anything: the evaluator reads and
// reports (REQ-GSM-011), and the argument vectors below are where that is
// decidable by inspection rather than by trust.
type ghQuerier struct{ root string }

func newGHQuerier(root string) *ghQuerier { return &ghQuerier{root: root} }

// runListArgs is the per-subject query, built as a function so its shape is
// testable without executing anything.
//
// The subject is named in the REQUEST (--workflow) rather than filtered out of
// a global answer. The difference is not stylistic: a repository-global listing
// is measurably incapable of answering for a low-frequency subject, because a
// busy repository's most recent N runs may contain none of that subject's — and
// the incapacity is invisible from inside the listing, which returns rows and
// looks like it worked.
func runListArgs(locator string) []string {
	return []string{
		"run", "list",
		"--workflow", filepath.Base(locator),
		"--limit", fmt.Sprint(runListLimit),
		"--json", "conclusion,createdAt",
	}
}

// allRunsArgs is the repository-global listing. It is DELIBERATELY reachable
// and is never called by the evaluator: keeping it callable is what lets
// AC-GSM-006 (b) be a measured call count rather than a source grep, and a
// mutant assembling the same request from string fragments defeats a grep but
// not a counter.
func allRunsArgs() []string {
	return []string{"run", "list", "--limit", fmt.Sprint(runListLimit), "--json", "conclusion,createdAt"}
}

// ghRun is one row of the listing.
type ghRun struct {
	Conclusion string    `json:"conclusion"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (g *ghQuerier) RunsForSubject(ctx context.Context, locator string) (RunHistory, error) {
	return g.list(ctx, runListArgs(locator))
}

func (g *ghQuerier) AllRuns(ctx context.Context) (RunHistory, error) {
	return g.list(ctx, allRunsArgs())
}

// list runs one listing and decodes it.
//
// A failed listing is an ERROR rather than an empty history, and the
// distinction is row 2's whole reason for existing: an empty history means the
// subject has not run, a failed query means nobody knows — and reporting the
// second as the first hands the operator "look again with a longer window" for
// what may be an expired credential.
func (g *ghQuerier) list(ctx context.Context, args []string) (RunHistory, error) {
	out, err := runGH(ctx, g.root, args)
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && len(exit.Stderr) > 0 {
			return RunHistory{}, fmt.Errorf("guardstate: gh %s: %w: %s",
				strings.Join(args, " "), err, strings.TrimSpace(string(exit.Stderr)))
		}
		return RunHistory{}, fmt.Errorf("guardstate: gh %s: %w", strings.Join(args, " "), err)
	}
	return decodeRunListing(out)
}

// runGH executes one gh invocation and returns its stdout.
//
// A package variable rather than an inline call so a test can exercise what
// list() DOES with an answer — and, more to the point, with a failure — without
// a subprocess. What stays outside every test is the two lines below: that gh
// is the binary, and that it runs in the evaluated tree.
var runGH = func(ctx context.Context, dir string, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = dir
	return cmd.Output()
}

// decodeRunListing maps a listing to run history.
//
// An unreadable listing is an ERROR rather than an empty history, and the
// distinction is row 2's whole reason for existing: an empty history means the
// subject has not run, an unreadable answer means nobody knows — and reporting
// the second as the first hands the operator "look again with a longer window"
// for what may be an expired credential.
func decodeRunListing(payload []byte) (RunHistory, error) {
	var rows []ghRun
	if err := json.Unmarshal(payload, &rows); err != nil {
		return RunHistory{}, fmt.Errorf("guardstate: decode the run listing: %w", err)
	}

	history := RunHistory{Runs: make([]Run, 0, len(rows))}
	for _, r := range rows {
		history.Runs = append(history.Runs, Run{Conclusion: r.Conclusion, At: r.CreatedAt})
	}
	return history, nil
}
