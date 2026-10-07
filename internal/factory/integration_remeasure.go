// integration_remeasure.go — the re-measure record and its one verifier
// (card t1479, SPEC-MERGE-WINDOW-QUEUE-001, REQ-MWQ-014/015).
//
// The re-measure is the expensive verification the SPEC moves OUTSIDE the
// window: it runs against the candidate tree before the lane joins the
// queue, and its record is keyed by that tree's SHA. The merge verb
// (REQ-MWQ-017) later requires the record keyed by the tree the merge will
// produce, which is what makes the in-window step seconds-long rather than
// a second re-measure.
//
// Two record forms exist and ONE verifier decides both (REQ-MWQ-014): the
// local form written by `moai integration remeasure`, and the candidate-CI
// form (SPEC-CANDIDATE-CI-001's candidate run id whose verdict is green).
// Until that SPEC lands, workflow.candidate_ci.enabled is absent and reads
// false — the local form governs and the candidate form is a no-op seam.
//
// The trust model is the one spec.md §D states: the record carries its
// build identity, but a hand-written record file is not distinguishable by
// the verifier. Actors are cooperative; forgery is a residual risk, not a
// defended boundary.
package factory

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/pkg/version"
)

// RemeasureRecord is the local-form re-measure record (REQ-MWQ-014/015).
// Tree is the key: the candidate tree SHA the commands ran against. Base is
// the absorbed integration-branch commit. StructuredCount is true only when
// the command's tool emitted a recognized structured test report — it is
// what separates "no tests ran" from "the tool reports no structure", the
// distinction REQ-MWQ-015 exists to enforce.
type RemeasureRecord struct {
	Tree     string `json:"tree"`
	Base     string `json:"base"`
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
	// StructuredRequired records that the command's tool SUPPORTS a
	// recognized structured report (go test): the verifier then demands
	// one. A tool with no recognized report leaves it false and is valid
	// on exit code zero (REQ-MWQ-015's last clause).
	StructuredRequired bool   `json:"structured_required"`
	HasStructured      bool   `json:"structured_count"`
	TestCount          int    `json:"test_count,omitempty"`
	BuildIdentity      string `json:"build_identity"`
	RecordedAt         string `json:"recorded_at"`
}

// remeasureDir resolves the re-measure store under the project's state
// directory, keyed by tree SHA — one file per candidate tree, overwritten by
// the next re-measure of the same tree.
func remeasureDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "state", "remeasure")
}

// WriteRemeasureRecord stores the record keyed by its tree.
func WriteRemeasureRecord(projectRoot, treeSHA string, rec RemeasureRecord) error {
	rec.Tree = treeSHA
	if err := os.MkdirAll(remeasureDir(projectRoot), 0o755); err != nil {
		return fmt.Errorf("re-measure store: %w", err)
	}
	return atomicWriteFile(filepath.Join(remeasureDir(projectRoot), treeSHA+".json"), rec)
}

// ReadRemeasureRecord returns the record keyed by treeSHA, or an error
// naming the absence. A missing record is a DIFFERENT state from an invalid
// one, and callers need to tell them apart: complete's step 3 refuses on a
// missing record with its own message, while the merge verb's cause 1
// refuses on an invalid one.
func ReadRemeasureRecord(projectRoot, treeSHA string) (*RemeasureRecord, error) {
	data, err := readFileNoRename(filepath.Join(remeasureDir(projectRoot), treeSHA+".json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no re-measure record for tree %s", treeSHA)
	}
	if err != nil {
		return nil, fmt.Errorf("read re-measure record: %w", err)
	}
	var rec RemeasureRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("re-measure record for %s is unreadable: %w", treeSHA, err)
	}
	// rec.Tree reads AS WRITTEN, never forced to the key: the merge step's
	// tree-identity check (REQ-MWQ-018 cause 4) compares the pinned tree
	// against the record's tree, and forcing the field here would turn that
	// check into dead code — a forged or corrupted record is exactly what
	// the check exists to refuse.
	return &rec, nil
}

// readFileNoRename is a small error-transparency wrapper around os.ReadFile:
// os.ErrNotExist must survive to the caller's errors.Is check.
func readFileNoRename(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ValidateRemeasureRecord decides whether a local-form record satisfies the
// re-measure gate (REQ-MWQ-014/015). It is the ONE verifier both verbs
// (remeasure's self-check and merge/complete's gate read) call.
//
// Invalid when:
//   - the command's exit code is non-zero — the observed exit is recorded
//     as observed, never rewound to green;
//   - the command's tool supports structured output and the record carries
//     no structured count (go test without -json — REQ-MWQ-015);
//   - the structured count is zero or the tool reported an empty sweep
//     (go test -json with zero tests — REQ-MWQ-015);
//   - any identity field the gate needs is missing (tree key, base, build
//     identity, command).
//
// A merge stand-in (factoryWriteMergeRecord's text) never satisfies this
// verifier (REQ-MWQ-020): it is a different file on a different path, and
// its command carries no recognized structure.
//
// @MX:ANCHOR: [AUTO] the single re-measure verifier every merge path consults
// @MX:REASON: complete's gate, the merge verb, the merge step, and the merge-readiness check all judge a record through it — weakening it here would let an invalid or empty-sweep record stand for a re-measure on every path at once. Measured fan-in: 5 production call sites across 4 files.
// @MX:SPEC: SPEC-MERGE-WINDOW-QUEUE-001
func ValidateRemeasureRecord(rec *RemeasureRecord) error {
	switch {
	case rec == nil:
		return errors.New("re-measure record is missing")
	case rec.Tree == "":
		return errors.New("re-measure record carries no tree key")
	case rec.Base == "":
		return errors.New("re-measure record carries no absorbed base commit")
	case rec.Command == "":
		return errors.New("re-measure record carries no command")
	case rec.BuildIdentity == "":
		return errors.New("re-measure record carries no build identity")
	case rec.ExitCode != 0:
		return fmt.Errorf("re-measure command exited %d (recorded as observed)", rec.ExitCode)
	case rec.StructuredRequired && !rec.HasStructured:
		return fmt.Errorf("command %q supports structured test output but none was requested or recognized", rec.Command)
	case rec.StructuredRequired && rec.TestCount <= 0:
		return fmt.Errorf("structured report carries %d tests — an empty sweep cannot stand for a re-measure", rec.TestCount)
	}
	return nil
}

// ClassifyStructuredOutput judges one command's output for REQ-MWQ-015: it
// returns the recognized per-test count, whether a structured report was
// recognized, and an error when the command's tool supports structured
// output but the caller did not request it.
//
// The recognized runner, for this repository, is `go test -json`: the count
// is the number of per-test pass events (a Test field present; the
// package-level pass event has none). An unstructured `go test` run is
// refused outright — the tool supports the structure, so a bare run cannot
// stand for a re-measure. A tool with no recognized report (true, cat, a
// lint) is valid on exit code zero and records command + exit only — that
// residual is the non-test-command risk spec.md §D names, removed only
// where the candidate-CI form is enabled.
func ClassifyStructuredOutput(command string, output io.Reader) (count int, structured bool, err error) {
	data, readErr := io.ReadAll(output)
	if readErr != nil {
		return 0, false, fmt.Errorf("read command output: %w", readErr)
	}
	text := string(data)
	if isGoTestCommand(command) {
		if !requestsGoTestJSON(command) {
			return 0, false, fmt.Errorf("go test supports structured output (-json) but the command did not request it")
		}
		return countGoTestJSONTests(text)
	}
	return 0, false, nil
}

// isGoTestCommand reports whether the command runs go test — the only tool
// with a recognized structured report in this repository. The tool is
// recognized INSIDE a compound command: the env-scrub form AGENTS.md
// prescribes (`unset VARS && go test ...`) is one invocation whose tool is
// go test, and classifying it by its first token read the scrubbed sweep as
// a non-test command whose zero-test record passed as valid (t1576 review
// round 1). Segments split on && / || / ; verbatim — a separator inside a
// quoted argument would over-split, which only ever widens recognition to a
// segment that still must name `go test` as its own first two words.
func isGoTestCommand(command string) bool {
	for _, segment := range shellSegments(command) {
		fields := stripLeadingEnv(shellFields(segment))
		// t1576 review round 4: the subcommand can sit behind go's global
		// flags (`go -C . test -json ...`) — recognize the tool behind them,
		// so the zero-test refusal reaches the -C form.
		if sub := goSubcommand(fields); len(sub) > 0 && sub[0] == "test" {
			return true
		}
	}
	return false
}

// shellSegments splits a command into its compound segments at the shell's
// command separators (&&, ||, ;) OUTSIDE quoted words — a separator inside
// quotes is data, not a second command (t1576 review round 12: the quoted
// argument 'note; go test -json is useful' must not mint a phantom go test
// segment).
func shellSegments(command string) []string {
	var segments []string
	var cur strings.Builder
	var quote byte
	flush := func() {
		segments = append(segments, cur.String())
		cur.Reset()
	}
	for i := 0; i < len(command); i++ {
		c := command[i]
		if quote == '\'' {
			// Single quotes honor no escapes; only the closing quote ends
			// them.
			if c == '\'' {
				quote = 0
			}
			cur.WriteByte(c)
			continue
		}
		if quote == '"' {
			// Inside double quotes a backslash escapes the next byte — the
			// embedded \" must not close the quotes (t1576 card-review).
			if c == '\\' && i+1 < len(command) {
				cur.WriteByte(c)
				cur.WriteByte(command[i+1])
				i++
				continue
			}
			if c == '"' {
				quote = 0
			}
			cur.WriteByte(c)
			continue
		}
		// Outside quotes a backslash escapes the next byte: the `'\''`
		// embed shellJoinArgs generates is a literal quote, not a boundary
		// (t1576 card-review).
		if c == '\\' && i+1 < len(command) {
			cur.WriteByte(c)
			cur.WriteByte(command[i+1])
			i++
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			cur.WriteByte(c)
			continue
		}
		if c == ';' {
			flush()
			continue
		}
		if (c == '&' || c == '|') && i+1 < len(command) && command[i+1] == c {
			flush()
			i++
			continue
		}
		cur.WriteByte(c)
	}
	flush()
	return segments
}

// shellFields splits a segment into fields the way sh splits words: quotes
// group characters into one field and are themselves dropped
// (`GOFLAGS='-count=1 -v'` is ONE field), so the env-prefix scan sees the
// tool behind a quoted value (t1576 review round 10). Backslash escapes are
// not honored — the classifier needs the word boundaries, not the exact
// bytes.
func shellFields(segment string) []string {
	var fields []string
	var cur strings.Builder
	inField := false
	var quote byte
	for i := 0; i < len(segment); i++ {
		c := segment[i]
		if quote == '\'' {
			// Single quotes honor no escapes; only the closing quote ends
			// them.
			if c == '\'' {
				quote = 0
			} else {
				cur.WriteByte(c)
			}
			continue
		}
		if quote == '"' {
			// Inside double quotes a backslash escapes the next byte — the
			// embedded \" must not close the quotes (t1576 card-review).
			if c == '\\' && i+1 < len(segment) {
				cur.WriteByte(c)
				cur.WriteByte(segment[i+1])
				i++
				continue
			}
			if c == '"' {
				quote = 0
			} else {
				cur.WriteByte(c)
			}
			continue
		}
		// Outside quotes a backslash escapes the next byte: the `'\''`
		// embed is a literal quote, not a boundary (t1576 card-review).
		if c == '\\' && i+1 < len(segment) {
			cur.WriteByte(c)
			cur.WriteByte(segment[i+1])
			i++
			inField = true
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			inField = true
			continue
		}
		if c == ' ' || c == '\t' || c == '\n' {
			if inField {
				fields = append(fields, cur.String())
				cur.Reset()
				inField = false
			}
			continue
		}
		cur.WriteByte(c)
		inField = true
	}
	if inField {
		fields = append(fields, cur.String())
	}
	return fields
}

// goSubcommand strips the leading `go` and the global flags that precede
// the subcommand, returning the subcommand and the rest. Only flags with a
// documented argument are consumed (`-C dir`); anything else ends the scan,
// so an unrecognized shape stays unrecognized rather than over-matching.
func goSubcommand(fields []string) []string {
	if len(fields) == 0 || fields[0] != "go" {
		return nil
	}
	rest := fields[1:]
	if len(rest) >= 2 && rest[0] == "-C" {
		rest = rest[2:]
	}
	return rest
}

// stripLeadingEnv drops the env-assignment prefix and any `env` invocation
// head a command may carry (`GOMAXPROCS=2 go test ...`, `env GOMAXPROCS=2
// go test ...`, `env -u NAME go test ...`) so the tool behind the prefix is
// what classifies (t1576 review rounds 2-3). Shell semantics: assignments
// are only assignments before the command word, so the scan stops at the
// first token without `=`; env's own option forms (-u NAME, -i,
// --ignore-environment, --) are consumed with their arguments.
func stripLeadingEnv(fields []string) []string {
	for len(fields) > 0 {
		switch {
		case fields[0] == "env":
			fields = fields[1:]
		case fields[0] == "-u" && len(fields) >= 2:
			fields = fields[2:]
		case fields[0] == "-i" || fields[0] == "--ignore-environment" || fields[0] == "--":
			fields = fields[1:]
		case strings.Index(fields[0], "=") > 0:
			fields = fields[1:]
		default:
			return fields
		}
	}
	return fields
}

// requestsGoTestJSON reports whether the go test command carries the -json
// flag.
func requestsGoTestJSON(command string) bool {
	for _, f := range strings.Fields(command) {
		if f == "-json" || strings.HasPrefix(f, "-json=") {
			return true
		}
	}
	return false
}

// emptySweepMarkers are the runner-reported empty-sweep tokens REQ-MWQ-015
// refuses: go test -json surfaces them inside output events when a package
// has no tests at all.
const emptySweepMarker = "[no test files]"

// countGoTestJSONTests counts the per-test pass events in a go test -json
// stream and refuses an empty sweep. A stream that parses but carries zero
// per-test events is a count of zero — the caller records it and the
// verifier refuses it as an empty sweep (the distinction between "no tests
// to run" and "no structure" is carried by structured=true).
func countGoTestJSONTests(text string) (count int, structured bool, err error) {
	if strings.Contains(text, emptySweepMarker) {
		return 0, false, fmt.Errorf("runner reported %s — an empty sweep cannot stand for a re-measure", emptySweepMarker)
	}
	structured = true
	decoder := json.NewDecoder(strings.NewReader(text))
	// t1576 review round 4: a pipe into `head` truncates the stream — EOF
	// alone reads as completion and the passes seen so far stand for a
	// finished sweep while later failures never arrive. Every package whose
	// start event is seen must also report its terminal event before the
	// stream may stand for a re-measure.
	started := map[string]bool{}
	finished := map[string]bool{}
	for {
		var event struct {
			Action  string `json:"Action"`
			Test    string `json:"Test"`
			Package string `json:"Package"`
		}
		if decErr := decoder.Decode(&event); decErr != nil {
			if errors.Is(decErr, io.EOF) {
				break
			}
			// A malformed stream (tool output interleaved with the JSON) is
			// not a recognized report: refuse rather than count the prefix.
			return 0, true, fmt.Errorf("go test -json stream is not a recognized report: %v", decErr)
		}
		// t1576 review round 1: `go test -json ./... | cat` reports the
		// pipe's exit 0 while a test failed — the stream is the verdict the
		// exit code cannot carry. Any fail event (per-test, or package-level
		// — a build or setup failure behind which no per-test event may
		// exist) invalidates the record.
		if event.Action == "fail" {
			name := event.Test
			if name == "" {
				name = event.Package + " (package-level)"
			}
			return 0, true, fmt.Errorf("go test -json stream reports a failing test %q — a failing sweep cannot stand for a re-measure", name)
		}
		switch {
		case event.Action == "start" && event.Package != "":
			started[event.Package] = true
			// t1576 review round 8: a package the stream runs AGAIN (a `;`
			// compound of two go test invocations) must re-arm its
			// completion — the first run's finished entry otherwise masks
			// the second run's truncation.
			delete(finished, event.Package)
		case event.Action == "pass" && event.Test != "":
			count++
		// A package without tests ends in a package-level skip, not a pass —
		// the terminal event all the same (t1576 review round 12), or every
		// mixed ./... sweep with one no-test package reads as truncated.
		case (event.Action == "pass" || event.Action == "skip") && event.Test == "" && event.Package != "":
			finished[event.Package] = true
		}
	}
	if missing := len(started) - len(finished); missing > 0 {
		return 0, true, fmt.Errorf("go test -json stream ended with %d started package(s) unreported — a truncated capture (a pipe into head/tail, a killed runner) cannot stand for a re-measure", missing)
	}
	return count, structured, nil
}

// integrationWorktreeCleanError is the refusal RunRemeasure returns when the
// clean-tree or unchanged-HEAD check fails (REQ-MWQ-016). The verb renders
// it to stderr and exits non-zero; no record is ever written on it.
type integrationWorktreeCleanError struct{ msg string }

func (e *integrationWorktreeCleanError) Error() string { return e.msg }

// IsIntegrationWorktreeCleanError reports whether err is a REQ-MWQ-016
// refusal.
func IsIntegrationWorktreeCleanError(err error) bool {
	var clean *integrationWorktreeCleanError
	return errors.As(err, &clean)
}

// gitIntegrationWorktreeClean runs `git status --porcelain
// --untracked-files=all` in dir and reports whether it is empty. The
// --untracked-files=all flag overrides any status.showUntrackedFiles config
// (O2): the check's promise is "no untracked byte", so the config cannot be
// allowed to hide one.
func gitIntegrationWorktreeClean(dir string) (bool, string, error) {
	cmd := exec.Command("git", "status", "--porcelain", "--untracked-files=all")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false, "", fmt.Errorf("git status in %s: %v: %s", dir, err, out)
	}
	text := strings.TrimSpace(string(out))
	return text == "", text, nil
}

// gitHeadSHA reads the HEAD commit of dir.
func gitHeadSHA(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("read HEAD of %s: %v", dir, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// gitMergeBase reads the merge-base of baseRef and HEAD in dir — the
// absorbed integration-branch commit the record names.
func gitMergeBase(dir, baseRef string) (string, error) {
	cmd := exec.Command("git", "merge-base", baseRef, "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("merge-base %s..HEAD in %s: %v", baseRef, dir, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// RunRemeasure is the re-measure body (REQ-MWQ-014/015/016): it verifies the
// worktree is clean and HEAD is recorded, runs the command with its output
// classified, re-verifies the tree stayed clean and HEAD unchanged, and
// writes the record keyed by HEAD's tree — or refuses without writing one.
//
// projectRoot is where the record is stored; worktree is the tree the
// command runs in; baseBranch is the integration branch whose absorbed tip
// the record names; command is the verbatim command line the verb was
// given. An invalid measurement (non-zero exit, empty sweep) still writes
// its record — the observed exit is recorded as observed, and the VERIFIER
// refuses it — so a later reader can see what ran rather than only that it
// failed.
func RunRemeasure(projectRoot, worktree, baseBranch, command string) (*RemeasureRecord, error) {
	clean, status, err := gitIntegrationWorktreeClean(worktree)
	if err != nil {
		return nil, err
	}
	if !clean {
		return nil, &integrationWorktreeCleanError{msg: fmt.Sprintf("integration remeasure: the worktree is not clean at start (%d status lines); commit, stash or clean it, then re-measure", strings.Count(status, "\n")+1)}
	}
	headBefore, err := gitHeadSHA(worktree)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = worktree
	outBytes, runErr := cmd.Output()
	exitCode := 0
	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("integration remeasure: run %q: %v", command, runErr)
		}
	}
	count, structured, classifyErr := ClassifyStructuredOutput(command, bytes.NewReader(outBytes))

	// The finish checks (REQ-MWQ-016): the tree must be clean and HEAD
	// unchanged across the run, or no record is written.
	clean, status, err = gitIntegrationWorktreeClean(worktree)
	if err != nil {
		return nil, err
	}
	if !clean {
		return nil, &integrationWorktreeCleanError{msg: fmt.Sprintf("integration remeasure: the command left the worktree dirty (%d status lines); no record written", strings.Count(status, "\n")+1)}
	}
	headAfter, err := gitHeadSHA(worktree)
	if err != nil {
		return nil, err
	}
	if headAfter != headBefore {
		return nil, &integrationWorktreeCleanError{msg: "integration remeasure: the command moved HEAD; no record written"}
	}

	tree := headTreeSHA(worktree, headAfter)
	base, err := gitMergeBase(worktree, baseBranch)
	if err != nil {
		return nil, err
	}
	rec := &RemeasureRecord{
		Tree:               tree,
		Base:               base,
		Command:            command,
		ExitCode:           exitCode,
		StructuredRequired: isGoTestCommand(command),
		HasStructured:      structured,
		TestCount:          count,
		BuildIdentity:      moaiBuildIdentity(),
		RecordedAt:         WindowClock().Format(time.RFC3339),
	}
	if err := WriteRemeasureRecord(projectRoot, tree, *rec); err != nil {
		return nil, err
	}
	if exitCode != 0 {
		return rec, fmt.Errorf("integration remeasure: %q exited %d (recorded as observed; the record is invalid until a green re-measure)", command, exitCode)
	}
	if classifyErr != nil {
		return rec, fmt.Errorf("integration remeasure: %q is not a valid re-measure (%v); the record is written and the verifier refuses it", command, classifyErr)
	}
	return rec, nil
}

// headTreeSHA resolves treeSHA for commit in dir.
func headTreeSHA(dir, commit string) string {
	cmd := exec.Command("git", "rev-parse", commit+"^{tree}")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// moaiBuildIdentity renders the producing binary's identity the record
// carries (REQ-MWQ-014): the ldflags-stamped version, commit, and build id.
func moaiBuildIdentity() string {
	id := fmt.Sprintf("moai %s (%s)", version.Version, version.Commit)
	if version.BuildID != "" {
		id += " build " + version.BuildID
	}
	return id
}
