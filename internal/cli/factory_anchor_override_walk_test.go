package cli

// factory_anchor_override_walk_test.go — SPEC-LAUNCHER-ARGV-001 REQ-LA-005 and
// REQ-LA-001 (sync audit iteration 3 on card t1748: F1, F2 and the
// test-independence finding).
//
// The override flag counts only where it stands — never as another option's
// VALUE and never after the launcher's own separator. "Where it stands" is
// decided on the argv the BACKEND receives: the launcher removes its own options
// (`-b`, `--bypass`, `-c`, `--continue`, `--model X`, `-m X`, `--permission-mode
// X` and the `=` forms) before the backend reads anything, so a token's
// neighbour in the typed argv is not its neighbour in the exec argv. Three
// earlier rounds patched one spelling at a time; these tests pin the rule itself:
//
//   - literal rows: every expectation below is written by hand from the rule
//     "the launcher strips first, the backend reads what is left", never derived
//     with the implementation's condition;
//   - the entry cells drive the real cc and glm leader entries;
//   - the property test enumerates every short argv over the launcher-consumed
//     forms, the backend value options, the flag and the separator, against an
//     oracle that restates the rule independently, and an AST check keeps the
//     oracle's form table complete: a case added to the consumption walk that the
//     table does not know fails the test instead of being missed.

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// sameArgv compares two argvs, reading nil and empty as the same argv.
func sameArgv(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}

// walkRow is one hand-written expectation: the typed launcher argv, whether the
// override flag stands as an override in it, the argv with the override
// removed, and the argv the backend receives (resolveLaunchView of that rest).
type walkRow struct {
	name        string
	args        []string
	wantAllowed bool
	wantRest    []string
	wantFinal   []string
	skipFinal   bool
}

// keepRow: the flag does NOT stand as an override; the argv is untouched.
func keepRow(name string, args, final []string) walkRow {
	return walkRow{name: name, args: args, wantAllowed: false, wantRest: args, wantFinal: final}
}

// overRow: the flag stands as an override and is removed.
func overRow(name string, args, rest, final []string) walkRow {
	return walkRow{name: name, args: args, wantAllowed: true, wantRest: rest, wantFinal: final}
}

// launcherConsumedSamples are launcher-consumed token groups, each with a value
// where the form takes one. The launcher removes the whole group before the
// backend reads the argv.
var launcherConsumedSamples = []struct {
	name string
	toks []string
}{
	{"-b", []string{"-b"}},
	{"--bypass", []string{"--bypass"}},
	{"-c", []string{"-c"}},
	{"--continue", []string{"--continue"}},
	{"--model opus", []string{"--model", "opus"}},
	{"-m opus", []string{"-m", "opus"}},
	{"--permission-mode plan", []string{"--permission-mode", "plan"}},
	{"--model=opus", []string{"--model=opus"}},
	{"--permission-mode=plan", []string{"--permission-mode=plan"}},
}

// backendValueSamples are backend options that take the next token as a value.
var backendValueSamples = []string{"--append-system-prompt", "-n", "--name", "-cn"}

func overrideWalkRows() []walkRow {
	flag := allowCardAnchorFlag
	rows := []walkRow{
		// F1: `--model=--` and `--permission-mode=--` store `--` as the value, yet
		// the NEXT `--` is still the launcher's separator, so the flag after it is
		// the backend's.
		keepRow("F1 --model=-- then the separator", []string{"--model=--", "--", flag, "-w", "WT-card"}, []string{flag, "-w", "WT-card"}),
		keepRow("F1 --permission-mode=-- then the separator", []string{"--permission-mode=--", "--", flag, "-w", "WT-card"}, []string{flag, "-w", "WT-card"}),
		keepRow("F1 --model=-- then only the flag after the separator", []string{"--model=--", "--", flag}, []string{flag}),

		// The separator as a launcher-consumed option's value, and not.
		overRow("--model takes the separator as its value", []string{"--model", "--", flag, "-w", "WT-card"}, []string{"--model", "--", "-w", "WT-card"}, []string{"-w", "WT-card"}),
		overRow("-m takes the separator as its value", []string{"-m", "--", flag}, []string{"-m", "--"}, nil),
		overRow("--permission-mode takes the separator as its value", []string{"--permission-mode", "--", flag}, []string{"--permission-mode", "--"}, nil),
		keepRow("the second separator after --model is the launcher's", []string{"--model", "--", "--", flag}, []string{flag}),
		keepRow("-b then the separator", []string{"-b", "--", flag}, []string{flag}),
		keepRow("--bypass then the separator", []string{"--bypass", "--", flag}, []string{flag}),
		keepRow("-c then the separator", []string{"-c", "--", flag}, []string{flag}),
		keepRow("--continue then the separator", []string{"--continue", "--", flag}, []string{flag}),
		keepRow("--model=opus then the separator", []string{"--model=opus", "--", flag}, []string{flag}),
		keepRow("--permission-mode=plan then the separator", []string{"--permission-mode=plan", "--", flag}, []string{flag}),
		keepRow("a backend value option then the separator", []string{"--name", "--", flag}, []string{"--name", flag}),
		keepRow("a cluster then the separator", []string{"-cn", "--", flag}, []string{"-cn", flag}),

		// F2, the other direction: the typed neighbour is a launcher-consumed token
		// that the backend never sees, so the flag IS standing.
		overRow("two value options once the token between them is stripped",
			[]string{"--append-system-prompt", "-b", "--append-system-prompt", flag},
			[]string{"--append-system-prompt", "-b", "--append-system-prompt"},
			[]string{"--append-system-prompt", "--append-system-prompt"}),
		overRow("a stripped option pair between two value options",
			[]string{"--name", "--model", "opus", "--name", flag},
			[]string{"--name", "--model", "opus", "--name"},
			[]string{"--name", "--name"}),
		overRow("a stripped -c between two clusters",
			[]string{"-cn", "-c", "-cn", flag},
			[]string{"-cn", "-c", "-cn"},
			[]string{"-cn", "-cn"}),

		// The launcher's own --branch pair is removed before the backend reads the
		// argv, so it too moves the flag's neighbour.
		keepRow("a removed --branch pair leaves the flag as the value of --name", []string{"--name", "--branch", "main", flag}, []string{"--name", flag}),
		overRow("a --branch pair then the flag", []string{"--branch", "main", flag}, []string{"--branch", "main"}, nil),
		// A flag-shaped --branch value is a parse error the launch refuses; nothing
		// is overridden, so the error is not cleared by removing the flag.
		{name: "a flag-shaped --branch value is not an override", args: []string{"--branch", flag, "main"}, wantAllowed: false, wantRest: []string{"--branch", flag, "main"}, skipFinal: true},

		// -w normalization: a bare -w followed by the flag stays bare; the flag is
		// standing and, once removed, the name after it is -w's value.
		overRow("a bare -w, the flag, then the name", []string{"-w", flag, "WT-card"}, []string{"-w", "WT-card"}, []string{"--worktree", "WT-card"}),

		// Iteration 4 F1: `--worktree=<flag>` and `-w=<flag>` are rewritten into two
		// tokens, and the second is the flag text — but nobody typed it as a flag, so
		// it never stands, and the typed standalone flag keeps its own reading.
		keepRow("F1 --worktree=<flag> as a synthesized token, the typed flag is a prompt value",
			[]string{"-b", "--worktree=" + flag, "--append-system-prompt", flag},
			[]string{"--worktree", flag, "--append-system-prompt", flag}),
		keepRow("F1 -w=<flag> as a synthesized token, the typed flag is a prompt value",
			[]string{"-b", "-w=" + flag, "--append-system-prompt", flag},
			[]string{"--worktree", flag, "--append-system-prompt", flag}),
		keepRow("F1 --worktree=<flag> alone", []string{"--worktree=" + flag}, []string{"--worktree", flag}),
		keepRow("F1 -w=<flag> then a name", []string{"-w=" + flag, "WT-card"}, []string{"--worktree", flag, "WT-card"}),
		overRow("F1 --worktree=<flag> then a typed standalone flag",
			[]string{"--worktree=" + flag, flag},
			[]string{"--worktree=" + flag},
			[]string{"--worktree", flag}),
		overRow("F1 -w=<flag> then a typed standalone flag and a prompt",
			[]string{"-w=" + flag, flag, "hello"},
			[]string{"-w=" + flag, "hello"},
			[]string{"--worktree", flag, "hello"}),
		overRow("F1 a typed standalone flag before --worktree=<flag>",
			[]string{flag, "--worktree=" + flag},
			[]string{"--worktree=" + flag},
			[]string{"--worktree", flag}),
		keepRow("F1 --worktree=<flag> after the separator is the backend's", []string{"--", "--worktree=" + flag, flag}, []string{"--worktree=" + flag, flag}),
	}

	// F2, the matrix: a backend value option, a launcher-consumed token, the flag
	// and a prompt. The launcher strips the token first, so the backend reads
	// `<option> <flag> <prompt>`: the flag is the option's value.
	for _, opt := range backendValueSamples {
		for _, group := range launcherConsumedSamples {
			args := append(append([]string{opt}, group.toks...), flag, "hello")
			rows = append(rows, keepRow("F2 "+opt+" "+group.name+" then the flag", args, []string{opt, flag, "hello"}))
		}
	}
	return rows
}

func TestLauncherAnchorOverrideWalk(t *testing.T) {
	pinSnapshotModel(t)

	// The consumption function and the view, in process.
	t.Run("literal_rows", func(t *testing.T) {
		for _, row := range overrideWalkRows() {
			t.Run(row.name, func(t *testing.T) {
				rest, allowed := consumeAllowCardAnchor(row.args)
				if allowed != row.wantAllowed || !sameArgv(rest, row.wantRest) {
					t.Errorf("consumeAllowCardAnchor(%q) = (%q, %v), want (%q, %v)", row.args, rest, allowed, row.wantRest, row.wantAllowed)
				}
				if row.skipFinal {
					return
				}
				if final := resolveLaunchView(rest).FinalArgv; !sameArgv(final, row.wantFinal) {
					t.Errorf("%q: the backend receives %q, want %q", row.args, final, row.wantFinal)
				}
			})
		}
	})

	// Through the real leader entry, both arms.
	for _, arm := range leaderArms {
		refused := func(t *testing.T, args []string) {
			t.Helper()
			fx := newLeaderFixture(t)
			obs := driveLeaderLaunch(t, fx, arm, args)
			assertRefusedBeforeMutation(t, arm, obs)
			refusalNames(t, arm, obs, fx.card, "WT-card")
			if _, err := os.Stat(filepath.Join(fx.root, config.DefaultLeaderAnchorAuditRelPath)); err == nil {
				t.Errorf("%s: an audit row exists, so a backend-side flag was recorded as an override", arm)
			}
		}

		// Control: no override flag at all.
		t.Run(arm+"/control_anchored_launch_is_refused", func(t *testing.T) {
			refused(t, []string{"-f", "-w", "WT-card", "hello"})
		})

		// F1: the flag after the launcher's separator belongs to the backend.
		t.Run(arm+"/F1_model_equals_separator", func(t *testing.T) {
			refused(t, []string{"-f", "--model=--", argSeparator, allowCardAnchorFlag, "-w", "WT-card"})
		})
		t.Run(arm+"/F1_permission_mode_equals_separator", func(t *testing.T) {
			refused(t, []string{"-f", "--permission-mode=--", argSeparator, allowCardAnchorFlag, "-w", "WT-card"})
		})

		// Iteration 4 F1: the flag-equal token the -w normalization synthesizes out
		// of `--worktree=<flag>` / `-w=<flag>` is not a typed flag. It is no override,
		// and it must not shift the reading of a typed token.
		t.Run(arm+"/F1_worktree_equals_flag_is_no_override", func(t *testing.T) {
			refused(t, []string{"-f", "--worktree=" + allowCardAnchorFlag, "-w", "WT-card"})
		})
		t.Run(arm+"/F1_w_equals_flag_is_no_override", func(t *testing.T) {
			refused(t, []string{"-f", "-w=" + allowCardAnchorFlag, "-w", "WT-card"})
		})
		t.Run(arm+"/F1_synthesized_token_does_not_shift_the_typed_flag", func(t *testing.T) {
			refused(t, []string{"-f", "--worktree=" + allowCardAnchorFlag, "-w", "WT-card", "--append-system-prompt", allowCardAnchorFlag})
		})
		t.Run(arm+"/F1_typed_flag_after_the_equals_form_is_honored", func(t *testing.T) {
			fx := newLeaderFixture(t)
			obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "--worktree=" + allowCardAnchorFlag, "-w", "WT-card", allowCardAnchorFlag})
			if obs.err != nil || !obs.launched {
				t.Fatalf("%s: override launch: err=%v launched=%v, want the typed override honored", arm, obs.err, obs.launched)
			}
			if _, err := os.Stat(filepath.Join(fx.root, config.DefaultLeaderAnchorAuditRelPath)); err != nil {
				t.Errorf("%s: no audit row for an honored override: %v", arm, err)
			}
		})

		// F2: the launcher strips the token between a value option and the flag, so
		// the flag is that option's value on the backend.
		for _, group := range []struct {
			name string
			toks []string
		}{
			{"-b", []string{"-b"}},
			{"--model opus", []string{"--model", "opus"}},
			{"--permission-mode=plan", []string{"--permission-mode=plan"}},
		} {
			t.Run(arm+"/F2_append_system_prompt_"+group.name, func(t *testing.T) {
				args := append(append([]string{"-f", "-w", "WT-card", "--append-system-prompt"}, group.toks...), allowCardAnchorFlag, "hello")
				refused(t, args)
			})
		}
		for _, opt := range []string{"-n", "--name", "-cn"} {
			t.Run(arm+"/F2_"+opt+"_-b", func(t *testing.T) {
				refused(t, []string{"-f", "-w", "WT-card", opt, "-b", allowCardAnchorFlag, "hello"})
			})
		}

		// F2, the other direction: the stripped token between two value options
		// leaves the flag standing, so the override is honored and recorded.
		t.Run(arm+"/F2_stripped_token_leaves_the_flag_standing", func(t *testing.T) {
			fx := newLeaderFixture(t)
			obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "WT-card", "--append-system-prompt", "-b", "--append-system-prompt", allowCardAnchorFlag})
			if obs.err != nil || !obs.launched {
				t.Fatalf("%s: override launch: err=%v launched=%v, want the override honored", arm, obs.err, obs.launched)
			}
			if _, err := os.Stat(filepath.Join(fx.root, config.DefaultLeaderAnchorAuditRelPath)); err != nil {
				t.Errorf("%s: no audit row for an honored override: %v", arm, err)
			}
			for _, a := range obs.execArgs {
				if a == allowCardAnchorFlag {
					t.Errorf("%s: the override flag reached the backend argv: %q", arm, obs.execArgs)
				}
			}
		})

		// A plain tree: the flag is the value of --append-system-prompt and the
		// backend receives it untouched.
		t.Run(arm+"/F2_flag_value_reaches_the_backend", func(t *testing.T) {
			fx := newLeaderFixture(t)
			obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "plain", "--append-system-prompt", "-b", allowCardAnchorFlag, "hello"})
			if obs.err != nil || !obs.launched {
				t.Fatalf("%s: plain-tree launch: err=%v launched=%v", arm, obs.err, obs.launched)
			}
			want := []string{"--append-system-prompt", allowCardAnchorFlag, "hello"}
			found := false
			for i := 0; i+len(want) <= len(obs.execArgs); i++ {
				found = found || reflect.DeepEqual(obs.execArgs[i:i+len(want)], want)
			}
			if !found {
				t.Errorf("%s: backend argv %q lacks the untouched run %q", arm, obs.execArgs, want)
			}
		})
	}
}

// oracleForm is one launcher-consumed form of the oracle table.
type oracleForm struct {
	prefix     bool // `--model=<v>`: the value follows the prefix in the same token
	takesValue bool // `--model <v>`: the next token is the value, `--` included
}

// launcherFormOracle restates, by hand, what the launcher consumes in its own
// right. TestLauncherAnchorOverrideWalkProperty/oracle_table_is_complete keeps it
// in step with the consumption walk's cases.
var launcherFormOracle = map[string]oracleForm{
	"-b": {}, "--bypass": {}, "-c": {}, "--continue": {},
	"--permission-mode": {takesValue: true}, "--model": {takesValue: true}, "-m": {takesValue: true},
	"--model=": {prefix: true}, "--permission-mode=": {prefix: true},
}

// backendTakesValue restates the backend model for the options the property's
// alphabet uses: each takes the next token as its value.
var backendTakesValue = map[string]bool{"--append-system-prompt": true, "--name": true, "-n": true, "-cn": true}

func lookupLauncherForm(tok string) (oracleForm, bool) {
	if f, ok := launcherFormOracle[tok]; ok && !f.prefix {
		return f, true
	}
	for name, f := range launcherFormOracle {
		if f.prefix && strings.HasPrefix(tok, name) {
			return f, true
		}
	}
	return oracleForm{}, false
}

// worktreeFormOracle restates, by hand, the -w / --branch spellings the launcher
// rewrites before its consumption walk reads anything (normalizeWorktreeFlag,
// splitWorktreeBranchFlag). The value says whether the spelling carries its value
// in the same token (`--worktree=<v>`); TestLauncherAnchorOverrideWalkProperty/
// worktree_form_table_is_complete keeps the table in step with the code.
var worktreeFormOracle = map[string]bool{
	"-w": false, "--worktree": false, "--worktree=": true, "-w=": true,
	"--branch": false, "--branch=": true,
}

// flagBearingWorktreeForms returns every `=` spelling of the table carrying the
// override flag text as its value: the typed tokens that, once rewritten, put a
// flag-equal token into the argv without anyone typing a flag.
func flagBearingWorktreeForms() []string {
	var out []string
	for form, joined := range worktreeFormOracle {
		if joined {
			out = append(out, form+allowCardAnchorFlag)
		}
	}
	sort.Strings(out)
	return out
}

// argvSpace is the number of sequences of 1..maxLen tokens over an alphabet.
func argvSpace(alphabet, maxLen int) int {
	total, pow := 0, 1
	for n := 1; n <= maxLen; n++ {
		pow *= alphabet
		total += pow
	}
	return total
}

// launchTok is one token of the argv the consumption walk reads.
type launchTok struct {
	tok string
	at  int // index in the typed argv; -1 for a token the launcher synthesized
}

// launchTokens restates the rewrite that precedes the walk: before the first typed
// separator the launcher removes the --branch tokens and splits `--worktree=<v>` /
// `-w=<v>` into `--worktree` and a synthesized `<v>`; the tail after that separator
// is verbatim backend input.
func launchTokens(g []string) []launchTok {
	firstSep := len(g)
	for i, tok := range g {
		if tok == argSeparator {
			firstSep = i
			break
		}
	}
	var out []launchTok
	for i, tok := range g {
		switch {
		case i < firstSep && strings.HasPrefix(tok, "--branch="):
			// removed before the walk
		case i < firstSep && (strings.HasPrefix(tok, "--worktree=") || strings.HasPrefix(tok, "-w=")):
			out = append(out, launchTok{"--worktree", i})
			if value := tok[strings.Index(tok, "=")+1:]; value != "" {
				out = append(out, launchTok{value, -1})
			}
		default:
			out = append(out, launchTok{tok, i})
		}
	}
	return out
}

// walkExpect is the oracle's answer for one typed launcher argv.
type walkExpect struct {
	allowed bool
	rest    []string // the argv with the standing flag tokens removed
	final   []string // what the backend receives (before launcher-injected flags)
}

// expectWalk applies the rule directly: rewrite the -w / --branch spellings, strip
// what the launcher consumes (up to the first separator it does not eat as a
// value), read the remaining sequence by the backend value model, and take the
// flag where it stands — a TYPED flag token only, never one the rewrite made.
func expectWalk(g []string) walkExpect {
	items := launchTokens(g)
	var kept []int // indices into items of the tokens the launcher passes through
	sep := len(items)
	for i := 0; i < len(items); {
		tok := items[i].tok
		if tok == argSeparator {
			sep = i
			break
		}
		if f, ok := lookupLauncherForm(tok); ok {
			if f.takesValue && i+1 < len(items) {
				i += 2
			} else {
				i++
			}
			continue
		}
		kept = append(kept, i)
		i++
	}
	standingAt := map[int]bool{}   // typed index of a standing flag
	standingItem := map[int]bool{} // item index of a standing flag
	for k := 0; k < len(kept); k++ {
		it := items[kept[k]]
		if it.tok == allowCardAnchorFlag && it.at >= 0 {
			standingAt[it.at] = true
			standingItem[kept[k]] = true
			continue
		}
		if backendTakesValue[it.tok] {
			k++
		}
	}
	var e walkExpect
	e.allowed = len(standingAt) > 0
	for i, tok := range g {
		if !standingAt[i] {
			e.rest = append(e.rest, tok)
		}
	}
	for _, idx := range kept {
		if !standingItem[idx] {
			e.final = append(e.final, items[idx].tok)
		}
	}
	for _, it := range items[min(sep+1, len(items)):] {
		e.final = append(e.final, it.tok)
	}
	return e
}

// eachArgv calls fn for every sequence of 1..maxLen tokens over alphabet. The
// slice handed to fn is reused: fn must not retain it.
func eachArgv(alphabet []string, maxLen int, fn func([]string)) {
	var rec func(cur []string)
	rec = func(cur []string) {
		if len(cur) > 0 {
			fn(cur)
		}
		if len(cur) == maxLen {
			return
		}
		for _, tok := range alphabet {
			rec(append(cur, tok))
		}
	}
	rec(make([]string, 0, maxLen))
}

// tokenLiterals returns the string literals fn matches the identifier tok
// against: `switch tok` cases, `tok == "..."` operands and the prefix argument
// of strings.HasPrefix / TrimPrefix on tok. Literals compared against anything
// else (a flag value, an empty-string guard) are not token forms.
func tokenLiterals(fn *ast.FuncDecl, tok string) []string {
	var out []string
	add := func(e ast.Expr) {
		if lit, ok := e.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			if v, err := strconv.Unquote(lit.Value); err == nil {
				out = append(out, v)
			}
		}
	}
	isTok := func(e ast.Expr) bool {
		id, ok := e.(*ast.Ident)
		return ok && id.Name == tok
	}
	ast.Inspect(fn, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SwitchStmt:
			if isTok(x.Tag) {
				for _, stmt := range x.Body.List {
					for _, e := range stmt.(*ast.CaseClause).List {
						add(e)
					}
				}
			}
		case *ast.BinaryExpr:
			if x.Op == token.EQL || x.Op == token.NEQ {
				if isTok(x.X) {
					add(x.Y)
				}
				if isTok(x.Y) {
					add(x.X)
				}
			}
		case *ast.CallExpr:
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok && len(x.Args) == 2 {
				if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "strings" &&
					(sel.Sel.Name == "HasPrefix" || sel.Sel.Name == "TrimPrefix") && isTok(x.Args[0]) {
					add(x.Args[1])
				}
			}
		}
		return true
	})
	return out
}

func countToken(argv []string, tok string) int {
	n := 0
	for _, a := range argv {
		if a == tok {
			n++
		}
	}
	return n
}

func TestLauncherAnchorOverrideWalkProperty(t *testing.T) {
	pinSnapshotModel(t)

	// The oracle table must name every form the consumption walk matches: a case
	// added to the walk and not to the table would otherwise be silently absent
	// from the enumeration below.
	t.Run("oracle_table_is_complete", func(t *testing.T) {
		walked := map[string]bool{}
		found := false
		ast.Inspect(parseGoFile(t, "launch_view.go"), func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || (fn.Name.Name != "consumeLaunchArgs" && fn.Name.Name != "walkLaunchArgs") {
				return true
			}
			found = true
			for _, lit := range tokenLiterals(fn, "arg") {
				walked[lit] = true
			}
			return false
		})
		if !found {
			t.Fatal("launch_view.go defines neither consumeLaunchArgs nor walkLaunchArgs: the completeness check has nothing to read")
		}
		delete(walked, argSeparator)
		for lit := range walked {
			if _, ok := launcherFormOracle[lit]; !ok {
				t.Errorf("the consumption walk matches %q, which launcherFormOracle does not know: add the form (and its value rule) to the oracle table", lit)
			}
		}
		for lit := range launcherFormOracle {
			if !walked[lit] {
				t.Errorf("launcherFormOracle names %q, which the consumption walk no longer matches", lit)
			}
		}
		if len(walked) < 5 {
			t.Errorf("the walk matched only %d literals %v: the AST read is not seeing the cases", len(walked), walked)
		}
	})

	// The same completeness check for the rewrite that precedes the walk: a -w or
	// --branch spelling added to the normalization or the branch split and not to
	// worktreeFormOracle would otherwise be silently absent from the enumerations.
	t.Run("worktree_form_table_is_complete", func(t *testing.T) {
		walked := map[string]bool{}
		read := map[string]bool{}
		ast.Inspect(parseGoFile(t, "launch_view.go"), func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			switch fn.Name.Name {
			case "normalizeWorktreeFlag", "normalizeWorktreeFlagFrom", "splitWorktreeBranchFlag", "splitWorktreeBranchFlagFrom":
				read[fn.Name.Name] = true
				for _, lit := range tokenLiterals(fn, "arg") {
					walked[lit] = true
				}
			}
			return false
		})
		for _, name := range []string{"normalizeWorktreeFlag", "splitWorktreeBranchFlag"} {
			if !read[name] {
				t.Fatalf("launch_view.go does not define %s: the completeness check has nothing to read", name)
			}
		}
		delete(walked, argSeparator)
		for lit := range walked {
			if _, ok := worktreeFormOracle[lit]; !ok {
				t.Errorf("the -w / --branch rewrite matches %q, which worktreeFormOracle does not know: add the spelling to the table", lit)
			}
		}
		for lit := range worktreeFormOracle {
			if !walked[lit] {
				t.Errorf("worktreeFormOracle names %q, which the -w / --branch rewrite no longer matches", lit)
			}
		}
		if len(walked) < 5 {
			t.Errorf("the rewrite matched only %d literals %v: the AST read is not seeing the cases", len(walked), walked)
		}
	})

	// For every short launcher argv: the override decision, the argv left to the
	// launch and the argv the backend receives all follow from the one rule.
	t.Run("override_decision_and_exec_argv_agree", func(t *testing.T) {
		alphabet := []string{allowCardAnchorFlag, argSeparator, "word"}
		alphabet = append(alphabet, "--append-system-prompt", "--name", "-n", "-cn")
		for name, f := range launcherFormOracle {
			if f.prefix {
				alphabet = append(alphabet, name+"opus", name+argSeparator)
			} else {
				alphabet = append(alphabet, name)
			}
		}
		// The spellings the launcher rewrites into a flag-equal token nobody typed.
		alphabet = append(alphabet, flagBearingWorktreeForms()...)
		// A map range is unordered; sort for a deterministic enumeration.
		sort.Strings(alphabet)

		total, bad := 0, 0
		eachArgv(alphabet, 4, func(g []string) {
			total++
			want := expectWalk(g)
			rest, allowed := consumeAllowCardAnchor(g)
			final := resolveLaunchView(rest).FinalArgv
			var problems []string
			if allowed != want.allowed {
				problems = append(problems, "override decision")
			}
			if !sameArgv(rest, want.rest) {
				problems = append(problems, "argv left to the launch")
			}
			if !sameArgv(final, want.final) {
				problems = append(problems, "argv the backend receives")
			}
			// The same statement without the walk: a flag that was an override is
			// gone from the argv the backend receives. What may remain is only a
			// flag-equal token the -w rewrite made out of an `=` form's value.
			synthesized := 0
			for _, it := range launchTokens(g) {
				if it.at < 0 && it.tok == allowCardAnchorFlag {
					synthesized++
				}
			}
			if countToken(g, allowCardAnchorFlag) == 1 && allowed && countToken(final, allowCardAnchorFlag) > synthesized {
				problems = append(problems, "an override reached the backend")
			}
			// Once the launcher has consumed its override, nothing is left to consume.
			if again, stillAllowed := consumeAllowCardAnchor(rest); stillAllowed || !sameArgv(again, rest) {
				problems = append(problems, "override still standing after consumption")
			}
			if len(problems) == 0 {
				return
			}
			bad++
			if bad <= 8 {
				t.Errorf("argv %q: %s\n  got:  allowed=%v rest=%q backend=%q\n  want: allowed=%v rest=%q backend=%q",
					g, strings.Join(problems, ", "), allowed, rest, final, want.allowed, want.rest, want.final)
			}
		})
		if want := argvSpace(len(alphabet), 4); total != want {
			t.Errorf("enumerated %d argv, want the full space %d (alphabet %d, length 1..4)", total, want, len(alphabet))
		}
		if bad > 0 {
			t.Errorf("%d of %d argv disagree with the rule", bad, total)
		}
	})

	// The launcher's -w and --branch tokens move a flag's neighbour too (a branch
	// pair is removed before the backend reads the argv; a bare -w gains the next
	// name once the flag is gone). The rule's invariants hold over them as well:
	// consumption removes the override flag and nothing else, and leaves nothing
	// further to consume.
	t.Run("worktree_and_branch_tokens", func(t *testing.T) {
		alphabet := []string{"-w", "WT-card", "--branch", "main", allowCardAnchorFlag, "--name", "-b", argSeparator, "--worktree="}
		alphabet = append(alphabet, flagBearingWorktreeForms()...)
		// Metamorphic: a `=` spelling whose value is the flag text and one whose value
		// is inert are the same argv to the rule, because the token the rewrite makes
		// out of the value is never a typed flag. Swapping one for the other must not
		// change which typed flag tokens are removed.
		inert := func(argv []string) []string {
			out := append([]string(nil), argv...)
			for i, tok := range out {
				for _, form := range flagBearingWorktreeForms() {
					if tok == form {
						out[i] = strings.TrimSuffix(form, allowCardAnchorFlag) + "x"
					}
				}
			}
			return out
		}
		total, bad := 0, 0
		eachArgv(alphabet, 4, func(g []string) {
			total++
			rest, allowed := consumeAllowCardAnchor(g)
			var problems []string
			if !sameArgv(withoutToken(rest, allowCardAnchorFlag), withoutToken(g, allowCardAnchorFlag)) {
				problems = append(problems, "consumption changed a token other than the flag")
			}
			if removed := countToken(g, allowCardAnchorFlag) - countToken(rest, allowCardAnchorFlag); allowed != (removed > 0) {
				problems = append(problems, "decision and removal disagree")
			}
			if again, still := consumeAllowCardAnchor(rest); still || !sameArgv(again, rest) {
				problems = append(problems, "override still standing after consumption")
			}
			swappedRest, swappedAllowed := consumeAllowCardAnchor(inert(g))
			if swappedAllowed != allowed || !sameArgv(swappedRest, inert(rest)) {
				problems = append(problems, "a flag-bearing = spelling reads differently from an inert one")
			}
			if len(problems) == 0 {
				return
			}
			bad++
			if bad <= 8 {
				t.Errorf("argv %q: %s (rest=%q allowed=%v)", g, strings.Join(problems, ", "), rest, allowed)
			}
		})
		if want := argvSpace(len(alphabet), 4); total != want {
			t.Errorf("enumerated %d argv, want the full space %d", total, want)
		}
		if bad > 0 {
			t.Errorf("%d of %d argv break an invariant", bad, total)
		}
	})
}

func withoutToken(argv []string, tok string) []string {
	var out []string
	for _, a := range argv {
		if a != tok {
			out = append(out, a)
		}
	}
	return out
}
