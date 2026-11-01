package cli

// Round 3 of N1 (decision D1 of the repair after the audit of 22a0afd247): the output of a gate or of a git step
// never enters the ledger or the verb's output. The reason carries the failure kind, the gate, the exit status or
// timeout, the output byte count and the path of a log file; the output goes to that file (mode 0600, in a 0700
// directory beside the ledger, size-capped, one file per card+candidate+kind).
//
// Real ledger, real git, the real `moai af land` verb. The secret value is assembled at run time by the gate from $A
// and $B, so a hit on the whole value can only come from the gate's output, never from the gate text. Words that a
// write hook refuses are spelled as @..@ tokens and expanded at run time (the shapes are the audit's).

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
)

const afD1Value = "hunter2" + "super"

// afD1LogCapUpper is an upper bound on the log file size the tests accept: the code's cap on the output plus room for
// the header that holds the gate text (2 KiB at most) and the note that says how much of the start was dropped. It was
// twice the cap, which let a log that the masking had made 2.5 times longer pass.
const afD1LogCapUpper = mergeq.DefaultFailureLogCap + 4<<10

var afD1Words = strings.NewReplacer(
	"@PW@", "pass"+"word", "@PWU@", "PASS"+"WORD", "@TK@", "to"+"ken", "@TKU@", "TO"+"KEN",
	"@BG@", "-----BE"+"GIN", "@EN@", "-----E"+"ND", "@PK@", "PRIV"+"ATE KEY", "@AZ@", "Author"+"ization",
	"@AK@", "api"+"-key", "@UC@", "://app:", "@GL@", "gl"+"pat-", "@XO@", "xo"+"xb-", "@NP@", "np"+"m_",
	"@CK@", "Coo"+"kie",
)

type afD1Row struct{ name, printf string }

// afRunBoth runs one `moai af` verb and returns stdout, stderr and the error, whatever the exit.
//
// With AF_TEST_BIN set it runs that binary's `af` verb instead (afRunBinary), so one test group can be run against two builds.
func afRunBoth(args ...string) (stdout, stderr string, err error) {
	if bin := os.Getenv("AF_TEST_BIN"); bin != "" {
		return afRunBinary(bin, args...)
	}
	cmd := newAFCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), errOut.String(), err
}

// afLedgerBytes is every byte of the ledger file and of its write-ahead log: a value in any row, notes included.
func afLedgerBytes(t *testing.T, db string) []byte {
	t.Helper()
	var all []byte
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if b, err := os.ReadFile(db + suffix); err == nil {
			all = append(all, b...)
		}
	}
	if len(all) == 0 {
		t.Fatalf("no ledger bytes under %s", db)
	}
	return all
}

var afD1LogPath = regexp.MustCompile(`log: (\S+\.log)`)

// afD1Land runs `af land` with the gate and checks that the value is in none of: stdout, stderr, the error, the whole
// ledger. It returns the reason and the log path the reason names ("" when it names none).
func afD1Land(t *testing.T, gate string, extra ...string) (outcome, reason, logPath string) {
	t.Helper()
	db, repo, _, _, _ := afQueuedCard(t)
	args := append([]string{"land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", gate}, extra...)
	out, errOut, err := afRunBoth(args...)
	ledgerBytes := afLedgerBytes(t, db)
	for where, text := range map[string]string{"stdout": out, "stderr": errOut, "ledger file": string(ledgerBytes)} {
		if strings.Contains(text, afD1Value) {
			t.Errorf("the value is in the %s (%d bytes)", where, len(text))
		}
	}
	if err != nil && strings.Contains(err.Error(), afD1Value) {
		t.Errorf("the value is in the error: %.120v", err)
	}
	var res struct{ Outcome, Reason string }
	if jerr := json.Unmarshal([]byte(out), &res); jerr != nil {
		t.Fatalf("stdout %.200q is not the verb's JSON line (err %v)", out, err)
	}
	if m := afD1LogPath.FindStringSubmatch(res.Reason); m != nil {
		logPath = m[1]
	}
	return res.Outcome, res.Reason, logPath
}

// afD1Shapes runs every row as a failing gate that prints the row.
func afD1Shapes(t *testing.T, rows []afD1Row, prelude string) {
	t.Helper()
	afSkipWithoutSh(t)
	for _, c := range rows {
		t.Run(c.name, func(t *testing.T) {
			gate := `A=hunter2; B=super; ` + prelude + `printf "` + afD1Words.Replace(c.printf) + `\n"; printf 'the last line\n'; exit 1`
			outcome, reason, logPath := afD1Land(t, gate)
			if outcome != "failed" {
				t.Fatalf("outcome %q, want failed (%.200q)", outcome, reason)
			}
			if logPath == "" {
				t.Fatalf("the reason names no log file: %.300q", reason)
			}
			log, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("the named log %s is unreadable: %v", logPath, err)
			}
			if !strings.Contains(string(log), "the last line") {
				t.Errorf("the log lacks the gate's last line (%d bytes)", len(log))
			}
		})
	}
}

// Ordinary output, in the kept window, no controls and no split: a failing tool that prints a credential.
func TestAFLand_D1PlainShapesNeverReachTheLedgerOrStdout(t *testing.T) {
	afD1Shapes(t, []afD1Row{
		{"control_assign", `@PW@=${A}${B}`},
		{"pass_assign", `pass=${A}${B}`},
		{"pwd_assign", `pwd=${A}${B}`},
		{"MYSQL_PWD_assign", `MYSQL_PWD=${A}${B}`},
		{"auth_assign", `auth=${A}${B}`},
		{"flag_pw_space", `mysqldump --@PW@ ${A}${B} db`},
		{"flag_tk_space", `gh --@TK@ ${A}${B} api`},
		{"flag_short_p", `mysql -uroot -p${A}${B} db`},
		{"curl_u", `curl -u user:${A}${B} https://example.test/x`},
		{"postgres_url", `connect failed: postgres@UC@${A}${B}@db.example.test:5432/x`},
		{"DATABASE_URL_postgres", `DATABASE_URL=postgres@UC@${A}${B}@db.example.test/x`},
		{"mysql_url", `mysql@UC@${A}${B}@db.example.test/x`},
		{"mongodb_srv_url", `mongodb+srv@UC@${A}${B}@cluster.example.test/x`},
		{"redis_empty_user_url", `redis://:${A}${B}@cache.example.test:6379/0`},
		{"amqp_url", `amqp@UC@${A}${B}@mq.example.test/`},
		{"ssh_url", `ssh@UC@${A}${B}@git.example.test/x`},
		{"https_url", `https@UC@${A}${B}@example.test/x`},
		{"https_url_value_with_slash", `https@UC@ab/${A}${B}@example.test/x`},
		{"assign_comma_in_value", `@PW@=ab,${A}${B}`},
		{"assign_brace_in_value", `@PW@=ab}${A}${B}`},
		{"assign_second_word", `@PW@=ab ${A}${B}`},
		{"assign_unclosed_quote", `@PW@=\"ab${A}${B}`},
		{"json_escaped_quote", `{\"@PW@\":\"ab\\\"${A}${B}\"}`},
		{"json_double_encoded", `{\\\"@PW@\\\":\\\"${A}${B}\\\"}`},
		{"python_dict", `{'@PW@': '${A}${B}'}`},
		{"xml_element", `<@PW@>${A}${B}</@PW@>`},
		{"yaml_block", `@PW@: |\n  ${A}${B}`},
		{"yaml_folded", `@TK@: >-\n  ${A}${B}`},
		{"value_two_lines_after", `@PW@=\nfirst\nsecond ${A}${B}`},
		{"prose_is", `the @PW@ is ${A}${B}`},
		{"x_api_key", `x-@AK@: ${A}${B}`},
		{"x_auth_token", `X-Auth-@TK@: ${A}${B}`},
		{"private_token", `PRIVATE-@TKU@: ${A}${B}`},
		{"authz_basic", `@AZ@: Basic ${A}${B}`},
		{"authz_token", `@AZ@: @TK@ ${A}${B}`},
		{"authz_custom", `@AZ@: Custom ${A}${B}`},
		{"set_cookie", `Set-@CK@: sid=${A}${B}; Path=/`},
		{"export_github_token", `export GITHUB_@TKU@=${A}${B}`},
		{"gitlab_pat", `@GL@${A}${B}`},
		{"slack_token", `@XO@${A}${B}`},
		{"npm_token", `@NP@${A}${B}`},
		{"pgp_private_block", `@BG@ PGP @PK@ BLOCK-----\n${A}${B}\n@EN@ PGP @PK@ BLOCK-----`},
		{"ssh_openssh_key", `@BG@ OPENSSH @PK@-----\n${A}${B}\n@EN@ OPENSSH @PK@-----`},
		{"invalid_utf8_in_marker", `pass\377word=${A}${B}`},
		{"zero_width_in_marker", "pass" + string(rune(0x200b)) + "word=${A}${B}"},
		{"bidi_in_marker", "pass" + string(rune(0x202e)) + "word=${A}${B}"},
		{"unknown_name", `foo=${A}${B}`},
		{"name_and_sep_on_two_lines", `@PW@\n=${A}${B}`},
		{"marker_then_blank_then_value", `@PW@=\n\n${A}${B}`},
	}, "")
}

// A URL credential whose host has no dot: no e-mail rule to catch it by accident.
func TestAFLand_D1UrlHostWithoutDotNeverReachesTheLedgerOrStdout(t *testing.T) {
	afD1Shapes(t, []afD1Row{
		{"postgres_compose_host", `connect failed: postgres@UC@${A}${B}@db:5432/x`},
		{"mysql_localhost", `mysql@UC@${A}${B}@localhost/x`},
		{"redis_compose_host", `redis://:${A}${B}@cache:6379/0`},
		{"amqp_compose_host", `amqp@UC@${A}${B}@mq:5672/`},
		{"https_localhost", `https@UC@${A}${B}@localhost:8080/x`},
	}, "")
}

// Markers inside an escape sequence, whose value is outside it (the audit's N1 regression), and an unterminated one.
func TestAFLand_D1MarkerInsideAnEscapeSequenceNeverReachesTheLedgerOrStdout(t *testing.T) {
	afD1Shapes(t, []afD1Row{
		{"osc_title_bel", `\033]0;@PW@=\007${A}${B}`},
		{"osc_title_st", `\033]0;@PW@=\033\\\\${A}${B}`},
		{"osc_hyperlink_url", `\033]8;;https@UC@${A}${B}@h/x\007link\033]8;;\007`},
		{"unterminated_osc_then_lines", `\033]0;title\nlast line says why it failed`},
	}, "")
}

// Shapes whose marker is off the 64 KiB window. $P is 70000 bytes of filler.
func TestAFLand_D1MarkerOffTheWindowNeverReachesTheLedgerOrStdout(t *testing.T) {
	afD1Shapes(t, []afD1Row{
		{"control_long_value", `@PW@=${P}${A}${B}`},
		{"foo_long_value", `foo=${P}${A}${B}`},
		{"pass_long_value", `pass=${P}${A}${B}`},
		{"pwd_long_value", `pwd=${P}${A}${B}`},
		{"postgres_url_long_user", `postgres://${P}:${A}${B}@db/x`},
		{"long_line_then_marker_inside_window", `${P} @PW@=${A}${B}`},
		{"flag_pw_space_long", `--@PW@ ${P}${A}${B}`},
		{"marker_then_newline_long", `@PW@=\n${P}${A}${B}`},
		{"marker_newline_two_lines_after", `@PW@=\nfirst\n${P}${A}${B}`},
		{"name_and_sep_on_two_lines", `@PW@\n=${P}${A}${B}`},
		{"colour_on_marker_long", `\033[1m@PW@\033[0m=${P}${A}${B}`},
		{"bel_inside_marker_long", `pass\aword=${P}${A}${B}`},
		{"spaces_marker_600", `@PW@$(printf '%600s' '')=${P}${A}${B}`},
	}, `P=$(head -c 70000 /dev/zero | tr '\0' 'x'); `)
}

// A repository hook prints a credential while `git merge` builds the candidate: the path that never went through
// the failure tail.
func TestAFLand_D1MergeHookStderrNeverReachesTheLedgerOrStdout(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, st, _ := afQueuedCard(t)
	hook := filepath.Join(repo, ".git", "hooks", "pre-merge-commit")
	body := afD1Words.Replace("#!/bin/sh\nA=hunter2; B=super\nprintf 'deploy check: @PW@=%s%s\\033[31m red\\033[0m\\n' \"$A\" \"$B\" >&2\nexit 1\n")
	if err := os.WriteFile(hook, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	out, errOut, err := afRunBoth("land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", "true")
	for where, text := range map[string]string{"stdout": out, "stderr": errOut, "ledger file": string(afLedgerBytes(t, db))} {
		if strings.Contains(text, afD1Value) {
			t.Errorf("the hook's value is in the %s", where)
		}
	}
	// an escape character in the text a reader sees (the ledger file is binary: a 0x1b byte there proves nothing)
	rows, qerr := st.Queue(context.Background())
	if qerr != nil || len(rows) != 1 {
		t.Fatalf("queue %+v %v", rows, qerr)
	}
	for where, text := range map[string]string{"stdout": out, "stderr": errOut, "fail reason": rows[0].FailReason} {
		if strings.Contains(text, "\x1b") {
			t.Errorf("an escape character is in the %s", where)
		}
	}
	var res struct{ Outcome, Reason string }
	if jerr := json.Unmarshal([]byte(out), &res); jerr != nil || res.Outcome != "failed" {
		t.Fatalf("stdout %.200q (err %v), want outcome failed", out, err)
	}
	m := afD1LogPath.FindStringSubmatch(res.Reason)
	if m == nil {
		t.Fatalf("the reason names no log file: %.300q", res.Reason)
	}
	if log, rerr := os.ReadFile(m[1]); rerr != nil || !strings.Contains(string(log), "deploy check") {
		t.Errorf("the merge log %s lacks the hook's line: %v", m[1], rerr)
	}
}

// Volume: a hook that prints 5 MB must not become a 5 MB reason, nor a 5 MB log.
func TestAFLand_D1MergeHookVolumeIsBounded(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, st, _ := afQueuedCard(t)
	hook := filepath.Join(repo, ".git", "hooks", "pre-merge-commit")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\nhead -c 5000000 /dev/zero | tr '\\0' 'x' >&2\necho\necho 'LAST HOOK LINE' >&2\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	out, _, _ := afRunBoth("land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", "true")
	if len(out) > 4096 {
		t.Errorf("stdout is %d bytes", len(out))
	}
	rows, err := st.Queue(context.Background())
	if err != nil || len(rows) != 1 || len(rows[0].FailReason) > 4096 {
		t.Fatalf("queue %+v %v: the fail reason must stay small", rows, err)
	}
	var res struct{ Reason string }
	_ = json.Unmarshal([]byte(out), &res)
	m := afD1LogPath.FindStringSubmatch(res.Reason)
	if m == nil {
		t.Fatalf("the reason names no log file: %.300q", res.Reason)
	}
	log, rerr := os.ReadFile(m[1])
	if rerr != nil || len(log) > afD1LogCapUpper || !strings.Contains(string(log), "LAST HOOK LINE") {
		t.Errorf("log: %d bytes (cap upper bound %d), last line kept=%v, err %v", len(log), afD1LogCapUpper, strings.Contains(string(log), "LAST HOOK LINE"), rerr)
	}
}

// A gate that prints 5 MB and then explains: the explanation is the last line of the log, the log is capped, and
// stdout points at the file instead of holding the line.
func TestAFLand_D1LoudFailingGateKeepsItsLastLineInTheLog(t *testing.T) {
	afSkipWithoutSh(t)
	_, reason, logPath := afD1Land(t, "head -c 5000000 /dev/zero | tr '\\0' x; echo; echo REAL-\"FAILURE\"-LINE; exit 1")
	if logPath == "" {
		t.Fatalf("the reason names no log file: %.300q", reason)
	}
	if strings.Contains(reason, "REAL-FAILURE-LINE") {
		t.Errorf("the output is in the reason: %.300q", reason)
	}
	if len(reason) > 4096 {
		t.Errorf("the reason is %d bytes", len(reason))
	}
	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(log) > afD1LogCapUpper || !strings.Contains(string(log), "REAL-FAILURE-LINE") {
		t.Errorf("log: %d bytes (cap upper bound %d), last line kept=%v", len(log), afD1LogCapUpper, strings.Contains(string(log), "REAL-FAILURE-LINE"))
	}
	if runtime.GOOS != "windows" {
		fi, serr := os.Stat(logPath)
		di, derr := os.Stat(filepath.Dir(logPath))
		if serr != nil || derr != nil || fi.Mode().Perm() != 0o600 || di.Mode().Perm() != 0o700 {
			t.Errorf("log file mode %v dir mode %v (%v %v), want 0600 and 0700", fi.Mode().Perm(), di.Mode().Perm(), serr, derr)
		}
	}
	if !strings.Contains(reason, "bytes") {
		t.Errorf("the reason does not state the output byte count: %.300q", reason)
	}
}

// A log that cannot be written must not stop the failure from being recorded.
func TestAFLand_D1ALogThatCannotBeWrittenDoesNotBlockTheFailure(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, st, _ := afQueuedCard(t)
	// the directory the logs go to (beside the ledger) is a file
	if err := os.WriteFile(filepath.Join(filepath.Dir(db), "gate-logs"), []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := afRunBoth("land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", "echo some output; exit 1")
	if err != nil {
		t.Fatalf("the verb failed because the log could not be written: %v", err)
	}
	var res struct{ Outcome, Reason string }
	if jerr := json.Unmarshal([]byte(out), &res); jerr != nil || res.Outcome != "failed" {
		t.Fatalf("stdout %.200q: want outcome failed", out)
	}
	if !strings.Contains(strings.ToLower(res.Reason), "logging failed") {
		t.Errorf("the reason does not say that logging failed: %.300q", res.Reason)
	}
	rows, qerr := st.Queue(context.Background())
	if qerr != nil || len(rows) != 1 || rows[0].FailReason == "" {
		t.Errorf("the queue row did not take the failure: %+v %v", rows, qerr)
	}
}

// Candidate and subject SHA are full lowercase hex names of commits that exist (D4). The refusal happens at the
// verb: abbreviated, HEAD, a branch name, an option, upper case, a missing commit.
func TestAFVerbs_D4OnlyACompleteExistingCommitSHAIsAccepted(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, cand, _, epoch := afQueuedCard(t)
	e := strconv.FormatInt(epoch, 10)
	bad := []string{"HEAD", "af/trunk", "c1", cand[:7], cand[:12], cand[:39], cand + "0", strings.ToUpper(cand), "--no-verify", "-Xours",
		strings.Repeat("0", 40), strings.Repeat("a", 64), "", " " + cand}
	for _, verb := range []string{"evidence", "enqueue"} {
		for _, v := range bad {
			t.Run(verb+"/"+v, func(t *testing.T) {
				args := []string{verb, "t1", "--db", db, "--epoch", e}
				if verb == "evidence" {
					args = append(args, "--kind", "green", "--subject-sha", v, "--result", "pass")
				} else {
					args = append(args, "--candidate", v)
				}
				out, _, err := afRunBoth(args...)
				if err == nil {
					t.Errorf("%s accepted %q: %.120s", verb, v, out)
					return
				}
				// a text that is no SHA is refused for that, before git is asked; a well-formed one that names no
				// commit is refused for that
				want := "complete lowercase hex"
				if mergeq.IsCompleteSHA(v) {
					want = "names no commit"
				}
				if !strings.Contains(err.Error(), want) {
					t.Errorf("%s %q: refused with %q, want a refusal that says %q", verb, v, err, want)
				}
			})
		}
	}
	// a complete SHA of a commit that exists passes the check (the repository is found through the card's lease)
	if out, errOut, err := afRunBoth("evidence", "t1", "--db", db, "--epoch", e, "--kind", "green", "--subject-sha", cand, "--result", "pass"); err != nil {
		t.Errorf("evidence refused the real commit %s: %v %s %s", cand, err, out, errOut)
	}
	_ = repo
}

// The audit's N13 sequence with the real verbs: the candidate is a branch name. It ends in a refusal at the first
// verb that names the branch, so nothing is measured, queued, landed or finished on it.
func TestAFVerbs_D4BranchNameReworkEndsInARefusal(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, st, epoch := afQueuedCard(t)
	e := strconv.FormatInt(epoch, 10)
	if _, _, err := afRunBoth("evidence", "t1", "--db", db, "--epoch", e, "--kind", "green", "--subject-sha", "c1", "--result", "pass"); err == nil {
		t.Error("evidence accepted the branch name c1 as a subject SHA")
	}
	if _, _, err := afRunBoth("enqueue", "t1", "--db", db, "--epoch", e, "--candidate", "c1"); err == nil {
		t.Error("enqueue accepted the branch name c1 as a candidate")
	}
	_ = repo
	cards, err := st.Cards(context.Background())
	if err != nil || len(cards) != 1 || string(cards[0].Position.Stage) == "DONE" {
		t.Errorf("cards %+v %v: the card must not be finished on a branch name", cards, err)
	}
}

// afUnqueuedCard is afQueuedCard without the queue entry: a claimed card whose commit has green and gate evidence.
func afUnqueuedCard(t *testing.T) (db, repo, cand string, st *ledger.Store, epoch int64) {
	t.Helper()
	return afUnqueuedCardAs(t, ledger.ClaimRequest{SessionID: "s-t1", PID: os.Getpid(), PIDStart: "start-t1"})
}

// afUnqueuedCardAs is afUnqueuedCard with the claim the caller describes (its holder: a process that is gone, a lease that
// has already run out); an empty Worktree is the fixture repository.
func afUnqueuedCardAs(t *testing.T, claim ledger.ClaimRequest) (db, repo, cand string, st *ledger.Store, epoch int64) {
	t.Helper()
	ctx := context.Background()
	db = filepath.Join(t.TempDir(), "af", "ledger.db")
	var err error
	st, err = ledger.Open(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	repo = t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", repo) // the verbs take --repo only for the project's own repository
	afGit(t, repo, "init", "-q", "-b", "work")
	if err := os.WriteFile(filepath.Join(repo, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	afGit(t, repo, "add", "-A")
	afGit(t, repo, "commit", "-q", "-m", "base")
	afGit(t, repo, "branch", "af/trunk")
	afGit(t, repo, "checkout", "-q", "-b", "c1", "af/trunk")
	if err := os.WriteFile(filepath.Join(repo, "a.txt"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	afGit(t, repo, "add", "-A")
	afGit(t, repo, "commit", "-q", "-m", "card t1")
	cand = afGit(t, repo, "rev-parse", "HEAD")
	afGit(t, repo, "checkout", "-q", "work")
	if err := st.AddCard(ctx, ledger.Card{ID: "t1", Title: "t1", Position: core.Position{Stage: core.StageReady}, Priority: 1,
		Spec: ledger.Spec{Class: "A", Paths: []string{"a.txt"}}}); err != nil {
		t.Fatal(err)
	}
	if claim.Worktree == "" {
		claim.Worktree = repo
	}
	c, err := st.Claim(ctx, claim)
	if err != nil {
		t.Fatal(err)
	}
	return db, repo, cand, st, c.Epoch
}

// The queue refuses a name that is not a commit's even when the ledger holds passing evidence under that name (the
// ledger takes any non-empty string; the verb is the trust boundary), and accepts the real commit.
func TestAFVerbs_D4EnqueueRefusesAnEvidencedNameThatIsNotACommit(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, cand, st, epoch := afUnqueuedCard(t)
	afBindFixture(t, st, repo)
	ctx := context.Background()
	e := strconv.FormatInt(epoch, 10)
	afGit(t, repo, "tag", "-a", "v1", "-m", "release", cand)
	tagObject := afGit(t, repo, "rev-parse", "refs/tags/v1")
	for _, name := range []string{"HEAD", "c1", "af/trunk", cand[:12], strings.ToUpper(cand), "--no-verify", strings.Repeat("0", 40), tagObject} {
		if err := st.AddEvidence(ctx, "t1", epoch, ledger.Evidence{Kind: ledger.KindGate, TreeSHA: "t", SubjectSHA: name, Result: ledger.ResultPass}); err != nil {
			t.Fatalf("the ledger refused evidence for %q: %v", name, err)
		}
		if _, _, err := afRunBoth("enqueue", "t1", "--db", db, "--epoch", e, "--candidate", name); err == nil {
			t.Errorf("enqueue accepted %q although the ledger holds a passing gate for it", name)
		}
	}
	if rows, err := st.Queue(ctx); err != nil || len(rows) != 0 {
		t.Fatalf("queue %+v %v, want it empty after the refusals", rows, err)
	}
	if err := st.AddEvidence(ctx, "t1", epoch, ledger.Evidence{Kind: ledger.KindGate, TreeSHA: "t", SubjectSHA: cand, Result: ledger.ResultPass}); err != nil {
		t.Fatal(err)
	}
	if out, errOut, err := afRunBoth("enqueue", "t1", "--db", db, "--epoch", e, "--candidate", cand); err != nil {
		t.Errorf("enqueue refused the real commit: %v %s %s", err, out, errOut)
	}
}

// --repo names the repository the commit is looked up in; the object name of an annotated tag is not a commit's.
func TestAFVerbs_D4TheRepositoryFlagAndTheObjectNameOfATag(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, cand, _, epoch := afQueuedCard(t)
	e := strconv.FormatInt(epoch, 10)
	afGit(t, repo, "tag", "-a", "v1", "-m", "release", cand)
	tagObject := afGit(t, repo, "rev-parse", "refs/tags/v1")
	evidence := func(sha string, extra ...string) error {
		args := append([]string{"evidence", "t1", "--db", db, "--epoch", e, "--kind", "green", "--subject-sha", sha, "--result", "pass"}, extra...)
		_, _, err := afRunBoth(args...)
		return err
	}
	if err := evidence(cand, "--repo", repo); err != nil {
		t.Errorf("--repo with the real commit: %v", err)
	}
	if err := evidence(cand, "--repo", t.TempDir()); err == nil {
		t.Error("a commit looked up in a repository that does not hold it was accepted")
	}
	if err := evidence(tagObject, "--repo", repo); err == nil {
		t.Errorf("the object name %s of an annotated tag was accepted as a commit", tagObject)
	}
	if _, _, err := afRunBoth("enqueue", "t1", "--db", db, "--epoch", e, "--candidate", tagObject, "--repo", repo); err == nil {
		t.Error("enqueue accepted the object name of an annotated tag")
	}
}
