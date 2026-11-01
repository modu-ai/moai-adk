package closure

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/escalation"
)

// Evidence file names inside the card evidence directory (design.md §A).
const (
	SecondReviewFile   = "second-review.jsonl"
	ClosureVerdictFile = "closure-verdict.jsonl"
	ReportMDFile       = "closure-report.md"
	ReportJSONFile     = "closure-report.json"
)

// ReportFileName returns the closure report file names. verdict.md is
// deliberately absent: no A4 code path writes it (spec.md §C.4).

// EvidencePaths is the resolved card evidence layout.
type EvidencePaths struct {
	Home           string // card evidence home (a worktree or the primary checkout)
	Dir            string // <Home>/.moai/reports/<card>
	SecondReview   string
	ClosureVerdict string
	ReportJSON     string
	ReportMD       string
}

// ResolveEvidenceHome applies the §C.7 rule: the linked worktree whose
// directory base name equals card, else the primary checkout. runDir is the
// directory the command runs from. An unreadable worktree list is an error
// (undetermined), never a silent primary-checkout fallback. The result is
// symlink-canonicalized: git reports worktree paths in its own resolved form,
// and writers and readers must build the same evidence paths from one home.
//
// @MX:ANCHOR: [AUTO] the §C.7 evidence-home rule every A4 writer and reader shares.
// @MX:REASON: 4 non-test callers (pushcheck.go, contract_report.go,
// mcp_audit_multi_record.go, contract_verdict.go); a divergent home rule
// would scatter the evidence files across trees.
func ResolveEvidenceHome(runDir, card string) (string, error) {
	wts, err := gitio.WorktreeList(runDir)
	if err != nil {
		return "", fmt.Errorf("closure: list worktrees: %w", err)
	}
	home := ""
	for _, w := range wts {
		if filepath.Base(w.Path) == card {
			home = w.Path
			break
		}
	}
	if home == "" {
		common, err := gitio.CommonDir(runDir)
		if err != nil {
			return "", fmt.Errorf("closure: resolve primary checkout: %w", err)
		}
		home = filepath.Dir(common)
	}
	if resolved, rerr := filepath.EvalSymlinks(home); rerr == nil {
		return resolved, nil
	}
	return home, nil
}

// EvidenceFor returns the resolved evidence paths for a card home.
func EvidenceFor(home, card string) EvidencePaths {
	dir := filepath.Join(home, ".moai", "reports", card)
	return EvidencePaths{
		Home:           home,
		Dir:            dir,
		SecondReview:   filepath.Join(dir, SecondReviewFile),
		ClosureVerdict: filepath.Join(dir, ClosureVerdictFile),
		ReportJSON:     filepath.Join(dir, ReportJSONFile),
		ReportMD:       filepath.Join(dir, ReportMDFile),
	}
}

// CanonicalReportHash returns the SHA-256 of the report's canonical JSON —
// the bytes `json.Marshal` produces with GeneratedAt masked to "". Both the
// verdict recorder and the Human Verdict section use this hash, so a verdict
// stays current across byte-identical rebuilds (the report is deterministic
// apart from generated_at, REQ-CLOSURE-002; hashing the raw file instead
// would stale every verdict on every rebuild).
func CanonicalReportHash(r *Report) string {
	saved := r.GeneratedAt
	r.GeneratedAt = ""
	data, err := json.Marshal(r)
	r.GeneratedAt = saved
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// HashFile returns the SHA-256 of a file's bytes ("" when unreadable).
func HashFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// ─── progress.md parsing (research.md §B.5: no Go parser exists) ───

// progressRowRe matches one §E.2 AC-matrix row: | AC-ID | status | evidence |.
var progressRowRe = regexp.MustCompile(`^\|\s*([A-Za-z0-9-]+)\s*\|\s*([^|]+?)\s*\|\s*(.*?)\s*\|\s*$`)

// ProgressRow is one AC matrix row of progress.md §E.2.
type ProgressRow struct {
	ID       string
	Status   string
	Evidence string
}

// ParseProgressRows extracts the §E.2 AC-matrix rows. It returns the rows in
// file order and whether a §E.2 heading exists at all.
//
// Only rows inside the §E.2 section count: the section opens at a heading line
// naming §E.2 and closes at the next heading of the same or a shallower level
// (deeper sub-headings stay inside it). A body sentence that merely mentions
// §E.2 opens nothing, and AC tables in other sections are not matrix rows.
func ParseProgressRows(progressMD []byte) ([]ProgressRow, bool) {
	var rows []ProgressRow
	found := false
	// level is the heading level of the open §E.2 section; 0 outside it.
	level := 0
	// fenceChar is the opening fence character; 0 outside a fenced block.
	var fenceChar byte
	fenceLen := 0
	for line := range strings.SplitSeq(string(progressMD), "\n") {
		// A fenced code block is quoted text: its "# comment" lines are not
		// headings and its table-looking lines are not matrix rows. It opens at a
		// fence line and closes only at a fence of the same character, at least
		// as long, with no info string — so a longer outer fence may quote a
		// shorter one.
		if ch, n, rest := fenceMarker(line); n > 0 {
			switch {
			case fenceChar == 0:
				fenceChar, fenceLen = ch, n
				continue
			case ch == fenceChar && n >= fenceLen && rest == "":
				fenceChar, fenceLen = 0, 0
				continue
			}
		}
		if fenceChar != 0 {
			continue
		}
		if lv := markdownHeadingLevel(line); lv > 0 {
			switch {
			case strings.Contains(line, "§E.2"):
				// A deeper heading that also names §E.2 (for example "### §E.2.1")
				// stays inside the open section and must not move its depth.
				if level == 0 || lv <= level {
					level = lv
				}
				found = true
			case level > 0 && lv <= level:
				level = 0
			}
			continue
		}
		if level == 0 {
			continue
		}
		if m := progressRowRe.FindStringSubmatch(line); m != nil &&
			strings.HasPrefix(m[1], "AC-") {
			rows = append(rows, ProgressRow{ID: m[1], Status: m[2], Evidence: m[3]})
		}
	}
	return rows, found
}

// fenceMarker reports a Markdown code-fence line: the fence character, the
// length of its run (3 or more), and the text after the run. n is 0 when line
// is not a fence (more than three leading spaces, or fewer than three marks).
func fenceMarker(line string) (ch byte, n int, rest string) {
	t := strings.TrimLeft(line, " ")
	if len(line)-len(t) > 3 || len(t) < 3 || (t[0] != '`' && t[0] != '~') {
		return 0, 0, ""
	}
	for n < len(t) && t[n] == t[0] {
		n++
	}
	if n < 3 {
		return 0, 0, ""
	}
	return t[0], n, strings.TrimSpace(t[n:])
}

// markdownHeadingLevel returns the ATX heading level of line (1-6), or 0 when
// the line is not a heading.
func markdownHeadingLevel(line string) int {
	// Up to three leading spaces are part of the heading; four make a code block.
	t := strings.TrimLeft(line, " ")
	if len(line)-len(t) > 3 {
		return 0
	}
	n := 0
	for n < len(t) && t[n] == '#' {
		n++
	}
	if n == 0 || n > 6 || n >= len(t) || t[n] != ' ' {
		return 0
	}
	return n
}

// ProgressFirstVerdict is the §E.3 fenced YAML block's display fields.
type ProgressFirstVerdict struct {
	RunStatus    string
	ACPassCount  string
	ACFailCount  string
	RunCommitSHA string
	Present      bool // the §E.3 block exists
}

var progressFieldRe = regexp.MustCompile(`^([a-z_]+):\s*(.*?)\s*$`)

// ParseFirstVerdict extracts run_status, ac_pass_count, ac_fail_count, and
// run_commit_sha from the fenced YAML block after the §E.3 heading.
func ParseFirstVerdict(progressMD []byte) ProgressFirstVerdict {
	var out ProgressFirstVerdict
	lines := strings.Split(string(progressMD), "\n")
	in := false
	sawFence := false
	for _, line := range lines {
		if !in {
			if strings.Contains(line, "§E.3") {
				in = true
			}
			continue
		}
		trimmed := strings.TrimSpace(line)
		if !sawFence {
			if strings.HasPrefix(trimmed, "```") {
				sawFence = true
			}
			continue
		}
		if strings.HasPrefix(trimmed, "```") {
			break // closing fence
		}
		if m := progressFieldRe.FindStringSubmatch(line); m != nil {
			switch m[1] {
			case "run_status":
				out.RunStatus, out.Present = m[2], true
			case "ac_pass_count":
				out.ACPassCount, out.Present = m[2], true
			case "ac_fail_count":
				out.ACFailCount, out.Present = m[2], true
			case "run_commit_sha":
				out.RunCommitSHA, out.Present = m[2], true
			}
		}
	}
	return out
}

// ─── plan-audit search (REQ-CLOSURE-011) ───

var (
	planAuditIterRe = regexp.MustCompile(`^plan-audit(?:-iter(\d+))?\.md$`)
	planAuditRevRe  = regexp.MustCompile(`^-review-(\d+)\.md$`)
)

// PlanAuditCandidate is one candidate file of the binding search.
type PlanAuditCandidate struct {
	Path string
	Hash string // recorded input hash to verify ("" when none)
}

// FindPlanAuditReport applies the REQ-CLOSURE-011 search order:
//  1. the receipt-named file (when the receipt carries inputs.plan_audit_report)
//  2. the highest-iteration plan-audit*.md in the card evidence directory
//  3. the highest-numbered .moai/reports/plan-audit/<SPEC-ID>-review-<n>.md
//
// The first existing file wins; the returned hash is the receipt-recorded
// input hash for step 1 (the caller compares it against the file's bytes).
func FindPlanAuditReport(ev EvidencePaths, receiptRef *ReceiptFileRef, specID string) (string, string) {
	if receiptRef != nil && receiptRef.Path != "" {
		p := filepath.Join(ev.Home, filepath.FromSlash(receiptRef.Path))
		if fileExists(p) {
			return p, receiptRef.SHA256
		}
	}
	// 2. highest-iteration plan-audit*.md in the card evidence directory.
	best := ""
	bestN := -1
	entries, err := os.ReadDir(ev.Dir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if m := planAuditIterRe.FindStringSubmatch(e.Name()); m != nil {
				n := 0
				if m[1] != "" {
					n, _ = strconv.Atoi(m[1])
				}
				if n > bestN {
					bestN, best = n, filepath.Join(ev.Dir, e.Name())
				}
			}
		}
	}
	if best != "" {
		return best, ""
	}
	// 3. highest-numbered .moai/reports/plan-audit/<SPEC-ID>-review-<n>.md.
	auditDir := filepath.Join(ev.Home, ".moai", "reports", "plan-audit")
	entries, err = os.ReadDir(auditDir)
	if err != nil {
		return "", ""
	}
	bestN = -1
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, specID+"-review-") {
			continue
		}
		if m := planAuditRevRe.FindStringSubmatch(strings.TrimPrefix(name, specID)); m != nil {
			n, _ := strconv.Atoi(m[1])
			if n > bestN {
				bestN, best = n, filepath.Join(auditDir, name)
			}
		}
	}
	return best, ""
}

// ReceiptFileRef is contract.ReceiptFileRef, re-exported for the display
// layer's signature readability.
type ReceiptFileRef = contract.ReceiptFileRef

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// readFileOrEmpty reads a file; a missing file is empty (the caller renders
// "not observed"), other errors are returned.
func readFileOrEmpty(path string) ([]byte, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return data, true, nil
}

// LoadA2Records loads the card's escalation records from the card evidence
// home through A2's own parser (escalation.ParseRecord — the schema is A2's
// and is never re-implemented here). Records decode in filename order;
// unparseable files are listed, never silently dropped.
func LoadA2Records(home, card string) (records []escalation.Record, unreadable []string, err error) {
	dir := escalation.RecordDir(home, card)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("closure: list records: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		data, rerr := os.ReadFile(filepath.Join(dir, name))
		if rerr != nil {
			unreadable = append(unreadable, name)
			continue
		}
		rec, perr := escalation.ParseRecord(data)
		if perr != nil {
			unreadable = append(unreadable, name)
			continue
		}
		records = append(records, rec)
	}
	return records, unreadable, nil
}
