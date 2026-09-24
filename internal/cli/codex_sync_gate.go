package cli

// codex_sync_gate.go — the sync-phase quality gate's decision core in Go, for
// the Codex Stop chain (SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2d, design §D2,
// §D3.3; plan.md Q4).
//
// Q4 decision (M2d): the judgment logic moves to a Go entry point, and the
// Claude script is NOT turned into a shim over it. The operator's M2d
// constraint keeps .claude/hooks/moai/sync-phase-quality-gate.sh byte-identical,
// so the two are parallel implementations of one decision, and the AC-HPR-002
// goldens — which run the unmodified script and this file on the same fixture
// — are the equivalence proof. Every predicate below names the script line it
// mirrors; a change to either side must keep those goldens green.
//
// Two halves:
//   - syncGateApplies: the script's self-gates (:207–218 subject, :227–231
//     language marker, :233–253 code delta), evaluated in-hook before any
//     receipt is read.
//   - produceSyncGateReceipt: the checks (:561–674) run out of hook by
//     `moai verify sync-gate`, recorded as a verify-snapshot receipt (operator
//     M2d decision (a): .moai/state/sync-quality-gate.last stays Claude-only).

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/modu-ai/moai-adk/internal/verify"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// syncGateCheckID names the sync gate's receipt entry.
const syncGateCheckID = "sync-gate"

// Receipt ExitCode bits: which slot failed. The Stop-side mode resolution
// needs them (the automatic tier blocks only a build failure, script :414–418).
const (
	syncGateC1Failed = 1
	syncGateC2Failed = 2
)

// syncCheck is one check slot (c1 vet/lint, c2 build) for a language.
type syncCheck struct {
	Label string
	Tool  string   // probed on PATH; absent → recorded as a skip (exit 0)
	Argv  []string // run in the project root
}

// syncGateChecks mirrors the script's per-language case (:561–674). A pipe to
// `head` after run_step in the script applies to run_step's own (empty)
// stdout, so it does not change the command and is not carried here.
var syncGateChecks = map[string][2]*syncCheck{
	"go":     {{"go vet", "go", []string{"go", "vet", "./..."}}, {"go build", "go", []string{"go", "build", "./..."}}},
	"python": {{"ruff", "ruff", []string{"ruff", "check", "."}}, nil},
	"node":   {{"eslint", "eslint", []string{"eslint", "."}}, nil},
	"rust":   {{"cargo check", "cargo", []string{"cargo", "check"}}, nil},
	"java": {{"javac compile check", "javac", []string{"sh", "-c",
		`out=$(find . -name "*.java" -exec javac -cp "$(find . -name "*.jar" -printf "{}:")" {} + 2>&1); rc=$?; printf "%s\n" "$out" | head -20; exit $rc`}}, nil},
	"kotlin": {{"kotlinc", "kotlinc", []string{"sh", "-c",
		`out=$(find . -name "*.kt" -exec kotlinc -cp "$(find . -name "*.jar" -printf "{}:")" {} + 2>&1); rc=$?; printf "%s\n" "$out" | head -20; exit $rc`}}, nil},
	"csharp": {{"dotnet build", "dotnet", []string{"dotnet", "build", "--no-restore"}}, nil},
	"ruby": {{"ruby syntax", "ruby", []string{"find", ".", "-name", "*.rb", "-exec", "sh", "-c",
		`rc=0; for f do ruby -c "$f" 2>&1 || rc=1; done; exit $rc`, "sh", "{}", "+"}}, nil},
	"php": {{"php syntax", "php", []string{"find", ".", "-name", "*.php", "-exec", "sh", "-c",
		`rc=0; for f do php -l "$f" 2>&1 || rc=1; done; exit $rc`, "sh", "{}", "+"}}, nil},
	"elixir": {{"mix compile", "mix", []string{"mix", "compile", "--no-start"}}, nil},
	"cpp": {{"g++ syntax check", "g++", []string{"sh", "-c",
		`out=$(find . \( -name "*.cpp" -o -name "*.cc" -o -name "*.cxx" \) -print -exec g++ -fsyntax-only -std=c++17 {} + 2>&1); rc=$?; if [ -z "$out" ]; then echo "0 C++ files checked: no *.cpp/*.cc/*.cxx found (not a passing check)"; else echo "$out"; fi; exit $rc`}}, nil},
	"scala": {{"scalac", "scalac", []string{"sh", "-c",
		`out=$(find . -name "*.scala" -exec scalac -cp "$(find . -name "*.jar" -printf "{}:")" {} + 2>&1); rc=$?; printf "%s\n" "$out" | head -20; exit $rc`}}, nil},
	"r": {{"R syntax", "R", []string{"sh", "-c", `files=$(find . -name "*.R" -o -name "*.r" | head -5); rc=0
while IFS= read -r f; do
    [ -n "$f" ] || continue
    Rscript -e "parse(\"$f\")" 2>&1 || rc=1
done <<EOF
$files
EOF
exit $rc`}}, nil},
	"flutter": {{"dart analyze", "dart", []string{"dart", "analyze"}}, nil},
	"swift":   {{"swift build", "swift", []string{"swift", "build"}}, nil},
}

// syncGateDeltaPatterns mirrors code_delta_pattern (:169–188).
var syncGateDeltaPatterns = map[string]*regexp.Regexp{
	"go":      regexp.MustCompile(`\.go$`),
	"python":  regexp.MustCompile(`\.py$`),
	"node":    regexp.MustCompile(`\.(js|ts|jsx|tsx|mjs|cjs)$`),
	"rust":    regexp.MustCompile(`\.rs$`),
	"java":    regexp.MustCompile(`\.java$`),
	"kotlin":  regexp.MustCompile(`\.(kt|kts)$`),
	"csharp":  regexp.MustCompile(`\.cs$`),
	"ruby":    regexp.MustCompile(`\.rb$`),
	"php":     regexp.MustCompile(`\.php$`),
	"elixir":  regexp.MustCompile(`\.ex$|\.exs$`),
	"cpp":     regexp.MustCompile(`\.(cpp|cc|cxx|h|hpp|hxx)$`),
	"scala":   regexp.MustCompile(`\.scala$`),
	"r":       regexp.MustCompile(`\.r$|\.R$`),
	"flutter": regexp.MustCompile(`\.dart$`),
	"swift":   regexp.MustCompile(`\.swift$`),
}

// syncGatePruneDirs mirrors the script's find prune set (:112–118).
var syncGatePruneDirs = map[string]bool{
	".git": true, ".hg": true, ".svn": true, "node_modules": true, "vendor": true,
	".venv": true, "venv": true, "site-packages": true, "__pycache__": true,
	".tox": true, ".nox": true, ".mypy_cache": true, ".ruff_cache": true, ".pytest_cache": true,
	"dist": true, "build": true, "target": true, ".next": true, ".output": true,
}

var kotlinBuildMarker = regexp.MustCompile(`(?i)kotlin\(|org\.jetbrains\.kotlin|kotlin-dsl|libs\.plugins\.kotlin`)

// detectSyncGateLanguages mirrors detect_languages (:90–157): every language
// evidenced by a manifest or a source suffix, in the script's order. One walk
// collects the suffixes the script probes one `find -quit` at a time.
func detectSyncGateLanguages(root string) []string {
	suffixes := map[string]bool{}
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != root && syncGatePruneDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if ext := filepath.Ext(d.Name()); ext != "" {
			suffixes[ext] = true
		}
		return nil
	})
	has := func(exts ...string) bool {
		for _, e := range exts {
			if suffixes[e] {
				return true
			}
		}
		return false
	}
	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(root, name))
		return err == nil
	}
	var langs []string
	add := func(l string) { langs = append(langs, l) }
	if exists("go.mod") || has(".go") {
		add("go")
	}
	if exists("pyproject.toml") || exists("requirements.txt") || has(".py") {
		add("python")
	}
	if exists("package.json") || has(".js", ".ts", ".jsx", ".tsx") {
		add("node")
	}
	if exists("Cargo.toml") || has(".rs") {
		add("rust")
	}
	kts, _ := os.ReadFile(filepath.Join(root, "build.gradle.kts"))
	switch {
	case (len(kts) > 0 && kotlinBuildMarker.Match(kts)) || has(".kt"):
		add("kotlin")
	case exists("pom.xml") || exists("build.gradle") || exists("build.gradle.kts") || has(".java"):
		add("java")
	}
	if exists("Gemfile") || has(".rb") {
		add("ruby")
	}
	if exists("composer.json") || has(".php") {
		add("php")
	}
	if exists("mix.exs") || has(".ex", ".exs") {
		add("elixir")
	}
	if exists("CMakeLists.txt") || exists("Makefile") || has(".cpp", ".cc", ".h") {
		add("cpp")
	}
	if exists("build.sbt") || has(".scala") {
		add("scala")
	}
	if exists("DESCRIPTION") || exists("renv.lock") || has(".R", ".r") {
		add("r")
	}
	if exists("pubspec.yaml") || has(".dart") {
		add("flutter")
	}
	if exists("Package.swift") || has(".swift") {
		add("swift")
	}
	if exists(".vs") || has(".csproj", ".cs") {
		add("csharp")
	}
	return langs
}

// isSyncPhaseSubject mirrors the script's case over the last commit subject
// (:209–218): *"docs("*"): sync-phase"*, *"chore("*"): sync-phase"*,
// *"docs: sync"*, *"chore: sync"*.
func isSyncPhaseSubject(subject string) bool {
	inOrder := func(first, then string) bool {
		i := strings.Index(subject, first)
		return i >= 0 && strings.Contains(subject[i+len(first):], then)
	}
	return inOrder("docs(", "): sync-phase") || inOrder("chore(", "): sync-phase") ||
		strings.Contains(subject, "docs: sync") || strings.Contains(subject, "chore: sync")
}

// syncGateApplies evaluates the script's three self-gates in its order. It
// returns the detected languages when the gate applies, and the reason it does
// not otherwise. Any git failure reads as "not a sync-phase commit", which is
// the script's own reading (:208 `|| echo ""`).
func syncGateApplies(ctx context.Context, root string) (bool, []string, string) {
	subject, err := gitOutputIn(ctx, root, "log", "-1", "--format=%s")
	if err != nil || !isSyncPhaseSubject(strings.TrimSpace(subject)) {
		return false, nil, "HEAD is not a sync-phase commit"
	}
	langs := detectSyncGateLanguages(root)
	if len(langs) == 0 {
		return false, nil, "no recognised language marker"
	}
	diffRange := "HEAD~1..HEAD"
	if _, err := gitOutputIn(ctx, root, "rev-parse", "--verify", "-q", "HEAD~1"); err != nil {
		// Initial commit: diff against the empty tree (:238).
		diffRange = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
	}
	names, _ := gitOutputIn(ctx, root, "diff", "--name-only", diffRange)
	delta := 0
	for _, l := range langs {
		re := syncGateDeltaPatterns[l]
		for _, n := range strings.Split(names, "\n") {
			if n != "" && re != nil && re.MatchString(n) {
				delta++
			}
		}
	}
	if delta == 0 {
		return false, nil, "no code-file delta in the sync-phase commit"
	}
	return true, langs, ""
}

// syncGateMode mirrors resolve_gate_mode (:390–424) for a failing outcome
// whose failed slots are c1Failed / c2Failed. It returns "blocking" or
// "advisory".
func syncGateMode(c1Failed, c2Failed bool) string {
	mode := "blocking"
	switch os.Getenv("MOAI_SYNC_GATE_BLOCKING") {
	case "0", "off", "false", "advisory", "no":
		mode = "advisory"
	}
	switch strings.ToLower(os.Getenv("MOAI_AUTONOMY_TIER")) {
	case "fully-autonomous":
		mode = "advisory"
	case "automatic":
		// The script's DECISION is block for any failing slot; a c1-only
		// failure (c2 passing) is advisory under the automatic tier.
		if c1Failed && !c2Failed {
			mode = "advisory"
		}
	}
	return mode
}

// syncGateReceiptState computes the five receipt fields for the current tree
// (design §D3.6), shared by the producer and the Stop chain. The config digest
// binds what the checks depend on — the detected languages and the check
// commands; the blocking switch is resolved at read time, like the script's
// re-delivery (:476–483), so it is not part of it. The tool version binds the
// check tools on PATH (path, size, modification time — no process is spawned
// in the hook) and the moai build that carries the check table.
func syncGateReceiptState(key string, langs []string) (verify.ReceiptState, [2]*syncCheck) {
	head, digest, _ := strings.Cut(key, ":")
	var checks [2]*syncCheck
	if len(langs) > 0 {
		checks = syncGateChecks[langs[0]]
	}
	var cmdText, tools []string
	for _, c := range checks {
		if c == nil {
			continue
		}
		cmdText = append(cmdText, c.Label+"="+strings.Join(c.Argv, " "))
		tools = append(tools, toolFingerprint(c.Tool))
	}
	return verify.ReceiptState{
		Head:       head,
		TreeDigest: digest,
		ConfigDigest: verify.ConfigDigest(map[string]string{
			"gate":      syncGateCheckID,
			"languages": strings.Join(langs, ","),
			"checks":    strings.Join(cmdText, "\n"),
		}),
		Command:     codexwiring.SyncGateReceiptCommand,
		ToolVersion: "moai " + version.GetVersion() + "; " + strings.Join(tools, "; "),
	}, checks
}

// toolFingerprint identifies a tool on PATH without running it.
func toolFingerprint(tool string) string {
	p, err := exec.LookPath(tool)
	if err != nil {
		return tool + "=absent"
	}
	st, err := os.Stat(p)
	if err != nil {
		return tool + "=" + p
	}
	return fmt.Sprintf("%s=%s:%d:%d", tool, p, st.Size(), st.ModTime().Unix())
}

// @MX:ANCHOR: [AUTO] sync-gate receipt producer — the out-of-hook half of the Codex sync gate (Q4); `moai verify sync-gate` and the AC-HPR-002 goldens call it
// @MX:REASON: the Stop chain reads only what this writes; a field it binds differently from syncGateReceiptState makes every receipt stale and every sync commit continue until the §D3.8 cap

// produceSyncGateReceipt runs the sync gate's checks for the current tree out
// of hook and records the outcome as a receipt. It runs regardless of the
// self-gates (an explicit invocation), and records the outcome under the key
// measured before the checks ran, so a check that writes into the tree leaves
// a receipt the Stop chain reads as stale rather than a pass for a tree it did
// not check.
func produceSyncGateReceipt(ctx context.Context, root string) (verify.Receipt, error) {
	key, err := verify.Key(ctx, root)
	if err != nil {
		return verify.Receipt{}, fmt.Errorf("sync gate receipt: %w", err)
	}
	langs := detectSyncGateLanguages(root)
	state, checks := syncGateReceiptState(key, langs)
	bits := 0
	for i, c := range checks {
		if c == nil {
			continue
		}
		if _, lerr := exec.LookPath(c.Tool); lerr != nil {
			continue // absent tool: graceful skip, exit 0 (:549–552)
		}
		cmd := exec.CommandContext(ctx, c.Argv[0], c.Argv[1:]...)
		cmd.Dir = root
		if out, rerr := cmd.CombinedOutput(); rerr != nil {
			bits |= 1 << i
			_, _ = fmt.Fprintf(os.Stderr, "sync gate: %s failed: %s\n", c.Label, strings.TrimSpace(tailString(string(out), 400)))
		}
	}
	verdict := "pass"
	if bits != 0 {
		verdict = "fail"
	}
	r := verify.Receipt{
		CheckID:      syncGateCheckID,
		Head:         state.Head,
		TreeDigest:   state.TreeDigest,
		ConfigDigest: state.ConfigDigest,
		Command:      state.Command,
		ToolVersion:  state.ToolVersion,
		ExitCode:     bits,
		Verdict:      verdict,
		RecordedAt:   time.Now(),
	}
	if err := verify.RecordReceipt(root, r); err != nil {
		return r, fmt.Errorf("sync gate receipt: %w", err)
	}
	return r, nil
}

// syncGateFailedLabel names the failing slot, as the script's BLOCKED_REASON
// does (:712–718).
func syncGateFailedLabel(checks [2]*syncCheck, bits int) string {
	for i, c := range checks {
		if bits&(1<<i) != 0 && c != nil {
			return c.Label + " failed"
		}
	}
	return "check failed (exit bits " + strconv.Itoa(bits) + ")"
}

func tailString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

// gitOutputIn runs one git subcommand in dir under ctx.
func gitOutputIn(ctx context.Context, dir string, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...).Output()
	return string(out), err
}
