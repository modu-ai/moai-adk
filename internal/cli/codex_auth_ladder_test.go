package cli

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// ─── AC-CL-008 / AC-CL-009 named expectation constants (판정 어휘: 정확 일치) ───
//
// The two-stage auth ladder replaces the substring classifier that read an
// error line as a successful login (spec §A.2). Every expectation below is
// compared with `==` against a named constant, never `strings.Contains`.

// sentinelCodexAuthToken is planted in fixtures to prove no credential value
// survives into an error, a stream, or a log sink.
const sentinelCodexAuthToken = "SENTINEL-TOKEN-9x9"

// codexAuthLadderSkipNames pins the ONLY tests in this SPEC that skip on
// GOOS=windows (AC-CL-014: the count is 3 and the names are constants).
var codexAuthLadderSkipNames = [3]string{
	"TestCodexLoginStatusRunner_StderrOnlyFixture",
	"TestCodexLoginStatusRunner_StdoutOnlyFixture",
	"TestCodexLoginStatusRunner_BothStreamsFixture",
}

// withCodexLoginStatusRunner swaps the low-level login-status seam and
// restores it on cleanup. The seam returns the two streams SEPARATELY so a
// test can express "stdout empty, stderr carries the answer" — the exact
// state the production defect erased (plan §C.2).
func withCodexLoginStatusRunner(t *testing.T, r codexLoginStatusRunnerFunc) {
	t.Helper()
	prev := codexLoginStatusRunner
	codexLoginStatusRunner = r
	t.Cleanup(func() { codexLoginStatusRunner = prev })
}

// countingLoginStatusRunner records how many times the command probe ran, so
// a test can assert stage 1 short-circuited the ladder (calls == 0).
type countingLoginStatusRunner struct {
	stdout, stderr []byte
	exitCode       int
	err            error
	calls          int
}

func (c *countingLoginStatusRunner) run(_ context.Context, _ string) ([]byte, []byte, int, error) {
	c.calls++
	return c.stdout, c.stderr, c.exitCode, c.err
}

// ─── AC-CL-008 stage 1: classifyCodexAuthFile is a PURE byte judgement ───

func TestClassifyCodexAuthFile_Table(t *testing.T) {
	cases := []struct {
		name         string
		raw          string
		wantProvider string
		wantOK       bool
		wantErr      bool
	}{
		{"chatgpt+access_token", `{"auth_mode":"chatgpt","tokens":{"access_token":"x"}}`, codexAuthChatGPT, true, false},
		{"chatgpt+empty_object", `{"auth_mode":"chatgpt","tokens":{}}`, "", false, false},
		{"chatgpt+null_token_values", `{"auth_mode":"chatgpt","tokens":{"access_token":null,"id_token":null}}`, "", false, false},
		{"chatgpt+empty_string_values", `{"auth_mode":"chatgpt","tokens":{"access_token":"","id_token":""}}`, "", false, false},
		{"chatgpt+no_tokens_key", `{"auth_mode":"chatgpt"}`, "", false, false},
		{"chatgpt+tokens_null", `{"auth_mode":"chatgpt","tokens":null}`, "", false, false},
		{"apikey+value", `{"auth_mode":"apikey","OPENAI_API_KEY":"x"}`, codexAuthAPIKey, true, false},
		{"apikey+null", `{"auth_mode":"apikey","OPENAI_API_KEY":null}`, "", false, false},
		{"apikey+empty", `{"auth_mode":"apikey","OPENAI_API_KEY":""}`, "", false, false},
		{"apikey+whitespace_only", `{"auth_mode":"apikey","OPENAI_API_KEY":"   "}`, "", false, false},
		{"unknown_mode", `{"auth_mode":"totally-new-mode","tokens":{"access_token":"x"}}`, "", false, false},
		{"mode_is_number", `{"auth_mode":123,"tokens":{"access_token":"x"}}`, "", false, true},
		{"mode_is_array", `{"auth_mode":["chatgpt"],"tokens":{"access_token":"x"}}`, "", false, true},
		{"mode_wrong_case", `{"auth_mode":"CHATGPT","tokens":{"access_token":"x"}}`, "", false, false},
		{"parse_failure", `{`, "", false, true},
		{"empty_bytes", ``, "", false, true},
		{"chatgpt+object_with_space", `{"auth_mode":"chatgpt","tokens":{ }}`, "", false, false},
		{"chatgpt+tokens_array", `{"auth_mode":"chatgpt","tokens":[]}`, "", false, false},
		{"chatgpt+tokens_string", `{"auth_mode":"chatgpt","tokens":"x"}`, "", false, false},
		{"chatgpt+token_false", `{"auth_mode":"chatgpt","tokens":{"access_token":false}}`, "", false, false},
		{"chatgpt+token_zero", `{"auth_mode":"chatgpt","tokens":{"access_token":0}}`, "", false, false},
		{"chatgpt+token_whitespace", `{"auth_mode":"chatgpt","tokens":{"access_token":"   "}}`, "", false, false},
		{"chatgpt+irrelevant_key", `{"auth_mode":"chatgpt","tokens":{"irrelevant":"x"}}`, "", false, false},
		{"chatgpt+account_metadata_only", `{"auth_mode":"chatgpt","tokens":{"account_id":"x"}}`, "", false, false},
		{"chatgpt+tokens_false", `{"auth_mode":"chatgpt","tokens":false}`, "", false, false},
		{"chatgpt+tokens_zero", `{"auth_mode":"chatgpt","tokens":0}`, "", false, false},
		{"apikey+false", `{"auth_mode":"apikey","OPENAI_API_KEY":false}`, "", false, false},
		{"apikey+zero", `{"auth_mode":"apikey","OPENAI_API_KEY":0}`, "", false, false},
		{"apikey+array", `{"auth_mode":"apikey","OPENAI_API_KEY":[]}`, "", false, false},
		{"apikey+object", `{"auth_mode":"apikey","OPENAI_API_KEY":{}}`, "", false, false},
		{"chatgpt+id_token", `{"auth_mode":"chatgpt","tokens":{"id_token":"x"}}`, codexAuthChatGPT, true, false},
		{"chatgpt+refresh_token", `{"auth_mode":"chatgpt","tokens":{"refresh_token":"x"}}`, codexAuthChatGPT, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotProvider, gotOK, gotErr := classifyCodexAuthFile([]byte(tc.raw))
			if gotProvider != tc.wantProvider {
				t.Errorf("provider = %q, want %q", gotProvider, tc.wantProvider)
			}
			if gotOK != tc.wantOK {
				t.Errorf("ok = %v, want %v", gotOK, tc.wantOK)
			}
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("err = %v, want non-nil=%v", gotErr, tc.wantErr)
			}
		})
	}
}

// TestCodexAuthFileTypes_NoSecretRetention is the CLOSED-SET reflection
// assertion of AC-CL-008: enumerate every field of codexAuthFile and its
// nested types recursively; the kind set must fall inside
// {string, bool, int, struct}, and exactly one string field may exist — the
// `auth_mode` enum. A map/slice/array/interface/pointer field would be able to
// retain a credential VALUE, which is what this forbids. Enumerating forbidden
// kinds instead would miss a retention shape nobody listed.
func TestCodexAuthFileTypes_NoSecretRetention(t *testing.T) {
	allowed := map[reflect.Kind]bool{
		reflect.String: true,
		reflect.Bool:   true,
		reflect.Int:    true,
		reflect.Struct: true,
	}
	var stringTags []string
	seen := map[reflect.Type]bool{}
	var walk func(reflect.Type, string)
	walk = func(rt reflect.Type, path string) {
		if seen[rt] {
			return
		}
		seen[rt] = true
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			k := f.Type.Kind()
			if !allowed[k] {
				t.Errorf("field %s.%s has kind %v, outside the allowed closed set {string,bool,int,struct}", path, f.Name, k)
				continue
			}
			if k == reflect.String {
				stringTags = append(stringTags, f.Tag.Get("json"))
			}
			if k == reflect.Struct {
				walk(f.Type, path+"."+f.Name)
			}
		}
	}
	walk(reflect.TypeOf(codexAuthFile{}), "codexAuthFile")

	if len(stringTags) != 1 {
		t.Fatalf("string fields = %v, want exactly 1 (auth_mode)", stringTags)
	}
	if stringTags[0] != "auth_mode" {
		t.Errorf("the single string field carries json tag %q, want %q", stringTags[0], "auth_mode")
	}
}

// TestReadCodexAuthFile_NoSentinelLeak plants a credential sentinel in a file
// that FAILS to parse and captures all FOUR channels REQ-CL-008 names
// (retained / logged / wrapped): the returned error text, stdout, stderr, and
// the log sink.
func TestReadCodexAuthFile_NoSentinelLeak(t *testing.T) {
	home := t.TempDir()
	// Unterminated JSON carrying the sentinel ⇒ parse failure with the secret present.
	fixture := `{"auth_mode":"chatgpt","tokens":{"access_token":"` + sentinelCodexAuthToken + `"`
	writeCodexAuthFixture(t, home, fixture)

	outR, outW, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("pipe: %v", pipeErr)
	}
	errR, errW, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("pipe: %v", pipeErr)
	}
	prevOut, prevErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outW, errW
	var logSink strings.Builder
	prevLogOut := log.Writer()
	log.SetOutput(&logSink)

	provider, ok, err := readCodexAuthFile(home)

	os.Stdout, os.Stderr = prevOut, prevErr
	log.SetOutput(prevLogOut)
	_ = outW.Close()
	_ = errW.Close()
	capturedOut := readAllString(t, outR)
	capturedErr := readAllString(t, errR)

	if ok || provider != "" {
		t.Errorf("unparseable file must not classify: provider=%q ok=%v", provider, ok)
	}
	if err == nil {
		t.Fatalf("unparseable file must return an error")
	}
	channels := map[string]string{
		"error":  err.Error(),
		"stdout": capturedOut,
		"stderr": capturedErr,
		"log":    logSink.String(),
	}
	for name, body := range channels {
		if strings.Contains(body, sentinelCodexAuthToken) {
			t.Errorf("channel %s leaked the credential sentinel: %q", name, body)
		}
	}
	// The error must still name the path and a reason (REQ-CL-008).
	if !strings.Contains(err.Error(), filepath.Join(home, "auth.json")) {
		t.Errorf("error must name the path, got %q", err.Error())
	}
}

func readAllString(t *testing.T, r *os.File) string {
	t.Helper()
	var sb strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		sb.Write(buf[:n])
		if err != nil {
			break
		}
	}
	_ = r.Close()
	return sb.String()
}

// ─── AC-CL-009 stage 2: parseCodexAuthLine is a PURE parser ───

// codexAuthReferenceGrammar is the test's OWN copy of the whole-line grammar.
// It is deliberately NOT imported from the implementation — sharing the
// constant would verify nothing.
var codexAuthReferenceGrammar = regexp.MustCompile(`(?i)^[ \t]*logged in using (chatgpt|api key)[ \t]*\r?$`)

// referenceClassify is the independent oracle the property assertion compares against.
func referenceClassify(combined string) string {
	set := map[string]bool{}
	for _, ln := range strings.Split(combined, "\n") {
		if m := codexAuthReferenceGrammar.FindStringSubmatch(ln); m != nil {
			set[strings.ToLower(m[1])] = true
		}
	}
	if len(set) != 1 {
		return codexAuthUnknown
	}
	for k := range set {
		if k == "chatgpt" {
			return codexAuthChatGPT
		}
		return codexAuthAPIKey
	}
	return codexAuthUnknown
}

func TestParseCodexAuthLine_Table(t *testing.T) {
	cases := []struct {
		in       string
		exitCode int
		want     string
	}{
		{"Logged in using ChatGPT", 0, codexAuthChatGPT},
		{"Logged in using API key", 0, codexAuthAPIKey},
		{"Logged in using ChatGPT", 1, codexAuthChatGPT},
		{"error: API key missing", 1, codexAuthUnknown},
		{"provider configuration unreadable", 1, codexAuthUnknown},
		{"failed to reach ChatGPT backend", 1, codexAuthUnknown},
		{"Logged in state unavailable: API key missing", 1, codexAuthUnknown},
		{"Logged in using ChatGPT (session expired)", 1, codexAuthUnknown},
		{"Logged in using Acme SSO", 0, codexAuthUnknown},
		{"warning: cache stale\nLogged in using ChatGPT", 0, codexAuthChatGPT},
		{"", 0, codexAuthUnknown},
		{"Logged in using ChatGPT\nLogged in using API key", 0, codexAuthUnknown},
		{"Logged in using ChatGPT\nLogged in using ChatGPT", 0, codexAuthChatGPT},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%q/rc%d", tc.in, tc.exitCode), func(t *testing.T) {
			if got := parseCodexAuthLine([]byte(tc.in), tc.exitCode); got != tc.want {
				t.Errorf("parseCodexAuthLine(%q, %d) = %q, want %q", tc.in, tc.exitCode, got, tc.want)
			}
		})
	}
}

// TestParseCodexAuthLine_NormalizationAxis mechanically derives every
// combination of field-observed noise from the two positive baselines. A
// lookup table memorising the fixed rows dies here.
func TestParseCodexAuthLine_NormalizationAxis(t *testing.T) {
	baselines := []struct {
		base, mixed, want string
	}{
		{"Logged in using ChatGPT", "Logged in using chatGPT", codexAuthChatGPT},
		{"Logged in using API key", "Logged in using api KEY", codexAuthAPIKey},
	}
	for _, b := range baselines {
		for _, core := range []string{b.base, b.mixed} {
			for _, lead := range []string{"", " "} {
				for _, trailSpace := range []string{"", " "} {
					for _, trailTab := range []string{"", "\t"} {
						for _, cr := range []string{"", "\r"} {
							in := lead + core + trailSpace + trailTab + cr
							if got := parseCodexAuthLine([]byte(in), 0); got != b.want {
								t.Errorf("derived case %q = %q, want %q", in, got, b.want)
							}
						}
					}
				}
			}
		}
	}
}

// TestParseCodexAuthLine_PropertyEquivalence generates 1,000 inputs from a
// fixed seed and asserts two-way equivalence with the independently-written
// reference grammar.
func TestParseCodexAuthLine_PropertyEquivalence(t *testing.T) {
	fragments := []string{
		"Logged in using ChatGPT", "Logged in using API key", "logged in using chatgpt",
		"Logged in state unavailable: API key missing", "error: API key missing",
		"provider configuration unreadable", "failed to reach ChatGPT backend",
		"Logged in using Acme SSO", "Logged in using ChatGPT (session expired)",
		"warning: cache stale", "", "   ", "\t", "chatgpt", "api key",
		"Logged in using ChatGPT\r", "  Logged in using API key  ",
	}
	rng := rand.New(rand.NewSource(0x5EED)) //nolint:gosec // deterministic corpus, not cryptography
	for i := 0; i < 1000; i++ {
		n := rng.Intn(4) + 1
		parts := make([]string, 0, n)
		for j := 0; j < n; j++ {
			parts = append(parts, fragments[rng.Intn(len(fragments))])
		}
		in := strings.Join(parts, "\n")
		rc := rng.Intn(3)
		got := parseCodexAuthLine([]byte(in), rc)
		want := referenceClassify(in)
		if got != want {
			t.Fatalf("case %d: parseCodexAuthLine(%q, %d) = %q, reference = %q", i, in, rc, got, want)
		}
	}
}

// ─── AC-CL-008 combine rule ───

func TestCombineCodexStreams_Cases(t *testing.T) {
	cases := []struct {
		name           string
		stdout, stderr string
		wantLines      []string
	}{
		{"stderr only", "", "Logged in using ChatGPT\n", []string{"Logged in using ChatGPT"}},
		{"stdout only", "Logged in using ChatGPT\n", "", []string{"Logged in using ChatGPT"}},
		{"both", "noise from stdout\n", "Logged in using ChatGPT\n", []string{"noise from stdout", "Logged in using ChatGPT"}},
		{"neither", "", "", nil},
		{"blank lines dropped", "\n  \n", "\t\n", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := combineCodexStreams([]byte(tc.stdout), []byte(tc.stderr))
			var gotLines []string
			if len(got) > 0 {
				gotLines = strings.Split(string(got), "\n")
			}
			if !reflect.DeepEqual(gotLines, tc.wantLines) {
				t.Errorf("combineCodexStreams lines = %q, want %q", gotLines, tc.wantLines)
			}
		})
	}
}

// ─── AC-CL-008 production combine path (fixture executables) ───
//
// These THREE tests are the ONLY GOOS-gated skips this SPEC adds (AC-CL-014);
// their names are pinned in codexAuthLadderSkipNames.

func codexFixtureRun(t *testing.T, script string) (string, string, int) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture executables are POSIX-only; the pure-function tests cover every platform")
	}
	path, err := filepath.Abs(filepath.Join("testdata", script))
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	stdout, stderr, code, runErr := defaultLoginStatusRunner(context.Background(), path)
	if runErr != nil {
		t.Fatalf("fixture %s: %v", script, runErr)
	}
	return string(stdout), string(stderr), code
}

func TestCodexLoginStatusRunner_StderrOnlyFixture(t *testing.T) {
	stdout, stderr, _ := codexFixtureRun(t, "codex-login-status-stderr-only.sh")
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if stderr != "Logged in using ChatGPT\n" {
		t.Errorf("stderr = %q, want the status line", stderr)
	}
	combined := combineCodexStreams([]byte(stdout), []byte(stderr))
	if got := strings.Split(string(combined), "\n"); len(got) != 1 || got[0] != "Logged in using ChatGPT" {
		t.Errorf("combined = %q, want the single status line", got)
	}
	if got := parseCodexAuthLine(combined, 0); got != codexAuthChatGPT {
		t.Errorf("parse(combined) = %q, want %q", got, codexAuthChatGPT)
	}
}

func TestCodexLoginStatusRunner_StdoutOnlyFixture(t *testing.T) {
	stdout, stderr, _ := codexFixtureRun(t, "codex-login-status-stdout-only.sh")
	if stdout != "Logged in using ChatGPT\n" {
		t.Errorf("stdout = %q, want the status line", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	combined := combineCodexStreams([]byte(stdout), []byte(stderr))
	if got := strings.Split(string(combined), "\n"); len(got) != 1 || got[0] != "Logged in using ChatGPT" {
		t.Errorf("combined = %q, want the single status line", got)
	}
}

func TestCodexLoginStatusRunner_BothStreamsFixture(t *testing.T) {
	stdout, stderr, _ := codexFixtureRun(t, "codex-login-status-both.sh")
	if stdout != "noise from stdout\n" {
		t.Errorf("stdout = %q", stdout)
	}
	if stderr != "Logged in using ChatGPT\n" {
		t.Errorf("stderr = %q", stderr)
	}
	combined := combineCodexStreams([]byte(stdout), []byte(stderr))
	lines := strings.Split(string(combined), "\n")
	if len(lines) != 2 {
		t.Fatalf("combined line count = %d, want 2 (a discarded stream dies here)", len(lines))
	}
	if lines[0] != "noise from stdout" || lines[1] != "Logged in using ChatGPT" {
		t.Errorf("combined = %q", lines)
	}
}

// ─── AC-CL-008 ladder integration (low-level runner stub, not a final-value stub) ───

func TestClassifyCodexAuth_LadderIntegration(t *testing.T) {
	t.Run("valid auth.json short-circuits the probe", func(t *testing.T) {
		home := t.TempDir()
		writeCodexAuthFixture(t, home, `{"auth_mode":"chatgpt","tokens":{"access_token":"x"}}`)
		t.Setenv("CODEX_HOME", home)
		stub := &countingLoginStatusRunner{stderr: []byte("Logged in using API key\n")}
		withCodexLoginStatusRunner(t, stub.run)

		if got := classifyCodexAuth(context.Background(), "/fake/codex"); got != codexAuthChatGPT {
			t.Errorf("AuthProvider = %q, want %q", got, codexAuthChatGPT)
		}
		if stub.calls != 0 {
			t.Errorf("runner calls = %d, want 0 — stage 1 must be wired into the assembly", stub.calls)
		}
	})

	t.Run("no auth.json + stderr-only probe", func(t *testing.T) {
		home := t.TempDir() // no auth.json inside
		t.Setenv("CODEX_HOME", home)
		stub := &countingLoginStatusRunner{stderr: []byte("Logged in using ChatGPT\n")}
		withCodexLoginStatusRunner(t, stub.run)

		if got := classifyCodexAuth(context.Background(), "/fake/codex"); got != codexAuthChatGPT {
			t.Errorf("AuthProvider = %q, want %q (the §A.2 defect: stderr was discarded)", got, codexAuthChatGPT)
		}
		if stub.calls != 1 {
			t.Errorf("runner calls = %d, want 1", stub.calls)
		}
	})

	t.Run("no auth.json + stdout-only probe", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("CODEX_HOME", home)
		stub := &countingLoginStatusRunner{stdout: []byte("Logged in using ChatGPT\n")}
		withCodexLoginStatusRunner(t, stub.run)

		if got := classifyCodexAuth(context.Background(), "/fake/codex"); got != codexAuthChatGPT {
			t.Errorf("AuthProvider = %q, want %q", got, codexAuthChatGPT)
		}
	})
}

// TestClassifyCodexAuth_RejectedAuthFileFallsBackToProbe closes carried debt
// D2: REQ-CL-008 requires a PRESENT-but-rejected auth.json to fall through to
// the command probe, and no acceptance criterion asserted it.
func TestClassifyCodexAuth_RejectedAuthFileFallsBackToProbe(t *testing.T) {
	rejected := map[string]string{
		"empty token object": `{"auth_mode":"chatgpt","tokens":{}}`,
		"unknown mode":       `{"auth_mode":"totally-new-mode","tokens":{"access_token":"x"}}`,
		"unparseable":        `{`,
	}
	for name, body := range rejected {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			writeCodexAuthFixture(t, home, body)
			t.Setenv("CODEX_HOME", home)
			stub := &countingLoginStatusRunner{stderr: []byte("Logged in using ChatGPT\n")}
			withCodexLoginStatusRunner(t, stub.run)

			if got := classifyCodexAuth(context.Background(), "/fake/codex"); got != codexAuthChatGPT {
				t.Errorf("AuthProvider = %q, want %q", got, codexAuthChatGPT)
			}
			if stub.calls != 1 {
				t.Errorf("runner calls = %d, want 1 — a rejected file must descend to the probe", stub.calls)
			}
		})
	}
}

// TestClassifyCodexAuth_UnreadableProbeIsAGap covers the four
// judgement-impossible axes: both streams empty, runner error, non-zero exit
// with a non-matching line, and an unparseable auth.json with a silent probe.
func TestClassifyCodexAuth_UnreadableProbeIsAGap(t *testing.T) {
	axes := map[string]struct {
		authJSON string
		stub     *countingLoginStatusRunner
	}{
		"both streams empty":     {"", &countingLoginStatusRunner{}},
		"runner error":           {"", &countingLoginStatusRunner{err: errors.New("probe failed")}},
		"non-zero + no grammar":  {"", &countingLoginStatusRunner{stderr: []byte("error: API key missing\n"), exitCode: 1}},
		"auth.json parse failed": {`{`, &countingLoginStatusRunner{}},
	}
	for name, ax := range axes {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			if ax.authJSON != "" {
				writeCodexAuthFixture(t, home, ax.authJSON)
			}
			t.Setenv("CODEX_HOME", home)
			withCodexLoginStatusRunner(t, ax.stub.run)
			if got := classifyCodexAuth(context.Background(), "/fake/codex"); got != codexAuthUnknown {
				t.Errorf("AuthProvider = %q, want %q", got, codexAuthUnknown)
			}
		})
	}
}

func writeCodexAuthFixture(t *testing.T, home, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(home, "auth.json"), []byte(body), 0o600); err != nil {
		t.Fatalf("write auth.json: %v", err)
	}
}
