package cli

// Card t1966 — `moai init` wizard/flag regression tests (items 1, 2, 7).

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// Item 1: the already-initialized probe must run BEFORE the wizard, so an
// initialized project gets the `moai update` guidance before any question.
func TestRunInit_AlreadyInitializedProbeRunsBeforeWizard(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	var calls int
	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		calls++
		return &wizard.WizardResult{}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "config"), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := newInitTestCmd()
	err := runInit(cmd, []string{projectDir})
	if err == nil {
		t.Fatal("runInit on an initialized project: want the already-initialized error, got nil")
	}
	if !strings.Contains(err.Error(), "moai update") {
		t.Errorf("error lacks the `moai update` guidance: %v", err)
	}
	if calls != 0 {
		t.Errorf("wizard ran %d time(s) before the already-initialized probe, want 0", calls)
	}
}

// Item 2: --bundles catalog validation lives in validateInitFlags (PreRunE).
func TestValidateInitFlags_UnknownBundleFailsEarly(t *testing.T) {
	cmd := newInitTestCmd()
	if err := cmd.Flags().Set("bundles", "no-such-bundle-t1966"); err != nil {
		t.Fatal(err)
	}
	err := validateInitFlags(cmd, nil)
	if err == nil {
		t.Fatal("validateInitFlags accepted an unknown bundle name")
	}
	if !strings.Contains(err.Error(), "no-such-bundle-t1966") || !strings.Contains(err.Error(), "--bundles") {
		t.Errorf("error should name the flag and the unknown bundle: %v", err)
	}
}

func TestValidateInitFlags_KnownBundleAccepted(t *testing.T) {
	cmd := newInitTestCmd()
	if err := cmd.Flags().Set("bundles", "backend"); err != nil {
		t.Fatal(err)
	}
	if err := validateInitFlags(cmd, nil); err != nil {
		t.Errorf("known bundle rejected: %v", err)
	}
}

// Item 5: development_mode is not a wizard question on the init path (TDD is
// the silent default; --mode is the only override), so a wizard result field
// must not reach the persisted config. The former dead mapping let it through.
func TestRunInit_WizardDevelopmentModeDoesNotReachConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{ProjectName: "devmode-proj", DevelopmentMode: "ddd", GitMode: "manual"}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	projectDir := filepath.Join(t.TempDir(), "devmode-proj")
	if err := runInit(newInitTestCmd(), []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "quality.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var mode string
	for _, line := range strings.Split(string(data), "\n") {
		if k, v, ok := strings.Cut(strings.TrimSpace(line), ":"); ok && k == "development_mode" {
			mode = strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	if mode != "tdd" {
		t.Errorf("quality.yaml development_mode = %q, want %q (silent default; wizard result must not set it)", mode, "tdd")
	}
}

// Item 7: the --enable-lsp registration default must match its help text
// ("default: true"), like --enforce-quality.
func TestInitCmd_EnableLSPRegisteredDefaultMatchesHelp(t *testing.T) {
	f := initCmd.Flags().Lookup("enable-lsp")
	if f == nil {
		t.Fatal("--enable-lsp not registered")
	}
	if f.DefValue != "true" {
		t.Errorf("--enable-lsp DefValue = %q, want \"true\" (help says default: true)", f.DefValue)
	}
}

// Item 3 follow-up (coordinator P2): a profile already holds a user name; the
// wizard answer is an explicit clear. The cleared name must reach user.yaml —
// runInit and syncInitProfile must not re-save the profile's name.
func TestRunInit_ClearedUserNameDoesNotResaveProfileName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := profile.WritePreferences(profile.GetCurrentName(), profile.ProfilePreferences{UserName: "Profile Alice"}); err != nil {
		t.Fatalf("seed profile: %v", err)
	}

	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{UserNameCleared: true, GitMode: "manual"}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	projectDir := filepath.Join(t.TempDir(), "cleared-name-proj")
	if err := runInit(newInitTestCmd(), []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "user.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "Profile Alice") {
		t.Errorf("user.yaml re-saved the profile name after the wizard cleared it:\n%s", data)
	}
}

// F1 (sync audit of d22e7d554d): on the RESUME path no template is redeployed,
// so an explicit user_name clear must be patched into the user.yaml the
// interrupted run left behind.
func TestInitResume_ClearedUserNameReachesUserYAML(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "resume-clear-proj")
	resumeSeedRun(t, projectDir, "claude")

	userYAML := filepath.Join(projectDir, ".moai", "config", "sections", "user.yaml")
	if err := os.WriteFile(userYAML, []byte("user:\n  name: \"Alice\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	prev := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{UserNameCleared: true}, nil
	}
	t.Cleanup(func() { runWizardFn = prev })

	if err := resumeRun(t, projectDir, ""); err != nil {
		t.Fatalf("resume failed: %v", err)
	}
	data, err := os.ReadFile(userYAML)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "Alice") {
		t.Errorf("successful resume retained user.name after an explicit clear:\n%s", data)
	}
}

// N1/N2 (iteration-2 sync audit): clearProjectUserName acts only when a
// non-empty user.name exists.
func clearUserNameFixture(t *testing.T, content string) (path string) {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path = filepath.Join(dir, "user.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestClearProjectUserName_ZeroByteFileIsNoError(t *testing.T) {
	path := clearUserNameFixture(t, "")
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path))))
	if err := clearProjectUserName(root); err != nil {
		t.Errorf("zero-byte user.yaml: clearProjectUserName returned %v, want nil", err)
	}
}

func TestClearProjectUserName_NoNameKeyStaysByteIdentical(t *testing.T) {
	const content = "user:\n  other: x\n"
	path := clearUserNameFixture(t, content)
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path))))
	if err := clearProjectUserName(root); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(path); string(got) != content {
		t.Errorf("user mapping without a name key was rewritten:\n got %q\nwant %q", got, content)
	}
}

// N3: a user.yaml that fails to parse must NOT be skipped. Pinned against the
// iteration-2 code (unconditional PatchFile): a syntax error was reported as an
// error, and a duplicate key was patched so the name was actually cleared.
func TestClearProjectUserName_DuplicateKeyStillClearsName(t *testing.T) {
	path := clearUserNameFixture(t, "user:\n  name: \"Alice\"\n  name: \"Alice\"\n")
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path))))
	const content = "user:\n  name: \"Alice\"\n  name: \"Alice\"\n"
	err := clearProjectUserName(root)
	// Write-then-verify (iteration 9): a duplicate-key file cannot be decoded
	// to prove equivalence, so the clear is refused and the file kept as is.
	if err == nil || !strings.Contains(err.Error(), refuseText) {
		t.Errorf("duplicate-key user.yaml: err = %v, want the refusal", err)
	}
	if got, _ := os.ReadFile(path); string(got) != content {
		t.Errorf("duplicate-key user.yaml was modified:\n%s", got)
	}
}

func TestClearProjectUserName_SyntaxErrorReturnsError(t *testing.T) {
	path := clearUserNameFixture(t, "user:\n  name: \"Alice\"\n  bad: [unclosed\n")
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path))))
	if err := clearProjectUserName(root); err == nil {
		t.Error("syntax-error user.yaml holding a name: clearProjectUserName returned nil, want the patch error")
	}
}

// F8: yamlpatch does not resolve aliases and edits the anchor node itself, so
// a user.yaml where user or user.name takes part in an anchor/alias pair is
// refused untouched instead of patched.
func assertClearRefusedUntouched(t *testing.T, content string) {
	t.Helper()
	path := clearUserNameFixture(t, content)
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path))))
	err := clearProjectUserName(root)
	if err == nil || !strings.Contains(err.Error(), refuseText) {
		t.Errorf("err = %v, want the anchors/aliases refusal", err)
	}
	if got, _ := os.ReadFile(path); string(got) != content {
		t.Errorf("file was modified:\n got %q\nwant %q", got, content)
	}
}

func TestClearProjectUserName_SharedAnchorRefusedUntouched(t *testing.T) {
	assertClearRefusedUntouched(t, "user:\n  name: &identity Alice\n  github_username: *identity\n")
}

func TestClearProjectUserName_NameAliasRefusedUntouched(t *testing.T) {
	assertClearRefusedUntouched(t, "base: &identity Alice\nuser:\n  name: *identity\n")
}

func TestClearProjectUserName_UserAliasRefusedUntouched(t *testing.T) {
	assertClearRefusedUntouched(t, "base: &identity\n  name: Alice\nuser: *identity\n")
}

// F9: a name supplied through YAML indirection (merge key or alias key at the
// top level) has no direct user.name path; patching would append a shadowing
// top-level `user:` mapping and drop the inherited siblings.
func TestClearProjectUserName_TopLevelMergeLiteralRefusedUntouched(t *testing.T) {
	assertClearRefusedUntouched(t, "<<:\n  user:\n    name: Alice\n    github_username: keep\n")
}

func TestClearProjectUserName_TopLevelMergeAliasRefusedUntouched(t *testing.T) {
	assertClearRefusedUntouched(t, "b: &b\n  user:\n    name: Alice\n    github_username: keep\n<<: *b\n")
}

func TestClearProjectUserName_AliasKeyRefusedUntouched(t *testing.T) {
	assertClearRefusedUntouched(t, "k: &k user\n*k :\n  name: Alice\n  github_username: keep\n")
}

// N8: an anchor on the user node itself is refused.
func TestClearProjectUserName_UserAnchorRefusedUntouched(t *testing.T) {
	assertClearRefusedUntouched(t, "user: &u\n  name: Alice\nother: *u\n")
}

// A merge key INSIDE user is not a defect: only the name row is blanked.
func TestClearProjectUserName_MergeKeyInsideUserStillClears(t *testing.T) {
	path := clearUserNameFixture(t, "base: &base\n  github_username: keep\nuser:\n  <<: *base\n  name: \"Alice\"\n")
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path))))
	if err := clearProjectUserName(root); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	got, _ := os.ReadFile(path)
	if strings.Contains(string(got), "Alice") || !strings.Contains(string(got), "github_username: keep") {
		t.Errorf("name not cleared or sibling disturbed:\n%s", got)
	}
}

// YAML indirection matrix (iteration-7 audit rows A01..E19, reconstructed from
// the verdict table; P01..P03 are files the product itself writes). Expected
// outcomes under the allow-list guard:
//
//	refuse      error with the generic text, bytes unchanged
//	clear       name blanked; effect oracle holds (see assertClearEffect)
//	noop        nothing to clear: no error, bytes unchanged
//	clearBroken like clear, but the loader cannot read the file before or after
//	            (duplicate keys, carried debt N6): only structure is compared
//	patchError  yamlpatch reports an error, bytes unchanged
const refuseText = "uses YAML structure this command cannot edit safely (anchors, aliases, merge keys, explicit tags, non-string keys, document markers, or several documents); edit user.name by hand"

type matrixRow struct {
	id, content, want string
}

// exactAfter, when a row id is present, is the whole file expected after a CLEAR
// (every other byte preserved). pinAfter maps user siblings whose decoded value
// is expected to change in the re-encode (null -> empty string), pinned as
// observed.
var (
	exactAfter = map[string]string{}
	pinAfter   = map[string]map[string]any{}
)

// commentTexts collects every comment of the YAML stream (head, line, foot) so
// a lost comment, including one after the document end, fails the oracle.
func commentTexts(b []byte) []string {
	var out []string
	var walk func(n *yaml.Node)
	walk = func(n *yaml.Node) {
		for _, c := range []string{n.HeadComment, n.LineComment, n.FootComment} {
			for _, l := range strings.Split(c, "\n") {
				if l = strings.TrimSpace(l); l != "" {
					out = append(out, l)
				}
			}
		}
		for _, c := range n.Content {
			walk(c)
		}
	}
	dec := yaml.NewDecoder(strings.NewReader(string(b)))
	for {
		var n yaml.Node
		if dec.Decode(&n) != nil {
			break
		}
		walk(&n)
	}
	sort.Strings(out)
	return out
}

func yamlDocCount(t *testing.T, b []byte) int {
	t.Helper()
	dec := yaml.NewDecoder(strings.NewReader(string(b)))
	n := 0
	for {
		var node yaml.Node
		if err := dec.Decode(&node); err != nil {
			break
		}
		n++
	}
	return n
}

// loaderName reads user.name the way the config loader does (merge-resolving
// struct decode); ok is false when the loader cannot read the file.
func loaderName(b []byte) (name string, ok bool) {
	var doc struct {
		User struct {
			Name string `yaml:"name"`
		} `yaml:"user"`
	}
	if yaml.Unmarshal(b, &doc) != nil {
		return "", false
	}
	return doc.User.Name, true
}

// assertClearEffect is the effect oracle for every CLEAR row: loader name is
// empty, every other decoded user key is unchanged (null and "" are DIFFERENT
// values; a changed notation must be pinned per row), nothing outside user
// changed, top-level key count and document count are unchanged, no comment of
// the stream was lost, and the optional exact bytes match.
func assertClearEffect(t *testing.T, row matrixRow, before, after []byte) {
	pin, exact := pinAfter[row.id], exactAfter[row.id]
	t.Helper()
	if n, ok := loaderName(after); !ok || n != "" {
		t.Errorf("loader name after clear = %q (readable=%v), want empty\n%s", n, ok, after)
	}
	if b, a := yamlDocCount(t, before), yamlDocCount(t, after); b != a {
		t.Errorf("document count %d -> %d\n%s", b, a, after)
	}
	var bm, am map[string]any
	if yaml.Unmarshal(before, &bm) != nil || yaml.Unmarshal(after, &am) != nil {
		t.Fatalf("decode failed\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if len(bm) != len(am) {
		t.Errorf("top-level keys %d -> %d\n%s", len(bm), len(am), after)
	}
	bu, _ := bm["user"].(map[string]any)
	au, _ := am["user"].(map[string]any)
	for k, v := range bu {
		if k == "name" {
			continue
		}
		want := v
		if pinned, ok := pin[k]; ok {
			want = pinned
		}
		got, present := au[k]
		if !present || !reflect.DeepEqual(got, want) {
			t.Errorf("user key %q = %#v (present=%v), want %#v\n%s", k, got, present, want, after)
		}
	}
	for k := range au {
		if _, had := bu[k]; !had && k != "name" {
			t.Errorf("user key %q appeared\n%s", k, after)
		}
	}
	for k, v := range bm {
		if got, present := am[k]; k != "user" && (!present || !reflect.DeepEqual(got, v)) {
			t.Errorf("top-level key %q = %#v, want %#v\n%s", k, got, v, after)
		}
	}
	if bc, ac := commentTexts(before), commentTexts(after); !reflect.DeepEqual(bc, ac) {
		t.Errorf("comments changed:\n before %q\n after  %q\n%s", bc, ac, after)
	}
	if exact != "" && string(after) != exact {
		t.Errorf("bytes differ from the expected whole file:\n got %q\nwant %q", after, exact)
	}
}

func TestClearProjectUserName_IndirectionMatrix(t *testing.T) {
	rows := []matrixRow{
		{"A01", "user:\n  name: Alice\n  github_username: keep\n", "clear"},
		{"A02", "user: {name: Alice, github_username: keep}\n", "clear"},
		{"A03", "{user: {name: Alice, github_username: keep}}\n", "clear"},
		{"A04", "b: &b\n  name: Alice\n  github_username: keep\nuser: *b\n", "refuse"},
		{"A05", "k: &k user\n*k :\n  name: Alice\n  github_username: keep\n", "refuse"},
		{"A06", "k: &user user\n*user :\n  name: Alice\n  github_username: keep\n", "refuse"},
		{"A07", "k: &user other\n*user :\n  name: Alice\nuser:\n  name: Carol\n  github_username: keep\n", "refuse"},
		{"A08", "<<:\n  user:\n    name: Alice\n    github_username: keep\n", "refuse"},
		{"A09", "b: &b\n  user:\n    name: Alice\n    github_username: keep\n<<: *b\n", "refuse"},
		{"A10", "a: &a\n  x: 1\nb: &b\n  user:\n    name: Alice\n    github_username: keep\n<<: [*a, *b]\n", "refuse"},
		{"A11", "b: &b\n  user:\n    name: Bob\n<<: *b\nuser:\n  name: Alice\n  github_username: keep\n", "refuse"},
		{"A12", "b: &b\n  user:\n    name: Alice\n    github_username: keep\n<<: *b\nuser:\n  other: x\n", "noop"},
		{"A13", "b: &b\n  user:\n    name: Alice\n    github_username: keep\n!!merge <<: *b\n", "refuse"},
		{"A14", "\"<<\": x\nuser:\n  other: y\n", "noop"},
		{"A15", "other: &x 1\nuser:\n  name: Alice\n  github_username: keep\nref: *x\n", "clear"},
		{"A16", "b: &b\n  x: 1\nuser:\n  name: Alice\n  github_username: keep\nnested:\n  <<: *b\n", "clear"},
		{"A17", "--- &r\nuser:\n  name: Alice\n  github_username: keep\n", "clear"},
		{"A18", "- user\n- name\n", "refuse"},
		{"B01", "user: &u\n  name: Alice\nother: *u\n", "refuse"},
		{"B02", "user: &u\n  name: Alice\n", "refuse"},
		{"B03", "user: &u {name: Alice, github_username: keep}\n", "refuse"},
		{"B04", "user:\n  <<: {name: Alice, github_username: keep}\n", "clear"},
		{"B05", "base: &base\n  name: Alice\n  github_username: keep\nuser:\n  <<: *base\n", "clear"},
		{"B06", "a: &a\n  name: Alice\nb: &b\n  github_username: keep\nuser:\n  <<: [*a, *b]\n", "clear"},
		{"B07", "base: &base\n  github_username: keep\nuser:\n  <<: *base\n  name: Alice\n", "clear"},
		{"B08", "base: &base\n  name: &n Alice\n  github_username: keep\nuser:\n  <<: *base\nref: *n\n", "clear"},
		{"C01", "base: &identity Alice\nuser:\n  name: *identity\n", "refuse"},
		{"C02", "user:\n  name: &identity Alice\n  github_username: *identity\n", "refuse"},
		{"C03", "user:\n  name: &n Alice\n  github_username: keep\n", "refuse"},
		{"C04", "user:\n  &nk name: Alice\n  github_username: keep\n", "refuse"},
		{"C05", "k: &name name\nuser:\n  *name : Alice\n  github_username: keep\n", "refuse"},
		{"C06", "k: &name other\nuser:\n  *name : Alice\n  name: Carol\n  github_username: keep\n", "refuse"},
		// C07/C08: an explicit tag on the name value is refused (iteration-8 rule).
		{"C07", "user:\n  name: !!str Alice\n  github_username: keep\n", "refuse"},
		{"C08", "user:\n  name: !!binary QWxpY2U=\n  github_username: keep\n", "refuse"},
		{"C09", "user:\n  name: |\n    Alice\n  github_username: keep\n", "clear"},
		{"C10", "user:\n  name: >\n    Alice\n    Smith\n  github_username: keep\n", "clear"},
		{"C11", "user:\n  name: Alice\n    Smith\n  github_username: keep\n", "clear"},
		{"C12", "user:\n  name: \"Alice\"\n  github_username: keep\n", "clear"},
		{"C13", "user:\n  name: 'Alice'\n  github_username: keep\n", "clear"},
		{"C14", "user:\n  name: Alice # who\n  github_username: keep\n", "clear"},
		{"C15", "user:\n  name:\n    Alice\n  github_username: keep\n", "clear"},
		{"C16", "user:\n  name: 123\n  github_username: keep\n", "clear"},
		{"C17", "user:\n  name: true\n  github_username: keep\n", "clear"},
		{"C18", "user: {name: Alice, github_username: Alice}\n", "clear"},
		// C19: the re-encode turns the flow null sibling into '' (audit N10/F15);
		// the equivalence check now refuses it (was CLEAR with a pinned change).
		{"C19", "user: {nameAlice: , name: Alice}\n", "refuse"},
		{"D01", "\"user\":\n  name: Alice\n  github_username: keep\n", "clear"},
		{"D02", "user:\n  \"name\": Alice\n  github_username: keep\n", "clear"},
		{"D03", "!!str user:\n  name: Alice\n  github_username: keep\n", "clear"},
		{"D04", "? user\n:\n  name: Alice\n  github_username: keep\n", "clear"},
		{"D05", "user:\n  \"n\\u0061me\": Alice\n  github_username: keep\n", "clear"},
		{"E01", "user:\n  name: Alice\nuser:\n  name: Bob\n", "refuse"},
		{"E02", "user:\n  name: Alice\n  name: Bob\n", "refuse"},
		{"E03", "x: 1\nx: 2\n", "refuse"},
		{"E04", "x: 1\nx: 2\nuser:\n  name: Alice\n", "refuse"},
		{"E05", "user:\r\n  name: Alice\r\n  github_username: keep\r\n", "clear"},
		{"E06", "\xef\xbb\xbfuser:\n  name: Alice\n  github_username: keep\n", "clear"},
		{"E07", "user:\n  name: Alice\n  github_username: keep\n---\nuser:\n  name: Bob\n", "refuse"},
		{"E08", "---\n---\nuser:\n  name: Alice\n", "noop"},
		{"E09", "---\nuser:\n  name: Alice\n  github_username: keep\n", "clear"},
		{"E10", "user:\n  name:\n", "noop"},
		{"E11", "user:\n  name: \"\"\n", "noop"},
		{"E12", "user:\n  name: ~\n", "noop"},
		{"E13", "user:\n  github_username: keep\n", "noop"},
		{"E14", "foo: bar\n", "noop"},
		{"E15", "", "noop"},
		{"E16", "# only a comment\n", "noop"},
		{"E17", "user:\n", "noop"},
		{"E18", "user:\n  name: Alice\n  bad: [unclosed\n", "patchError"},
		{"E19", "user:\n  name:\n    - a\n    - b\n", "refuse"},
		// F01: the earlier 17-row table's row 7 (anchored sibling VALUE inside user).
		{"F01", "user:\n  name: Alice\n  github_username: &g keep\n  alias: *g\n", "clear"},
		// G01/G02: a top-level merge key next to a direct plain user (separate
		// failure point from A11 for the top-level-merge rule, audit N12).
		{"G01", "b: &b\n  x: 1\n<<: *b\nuser:\n  name: Alice\n  github_username: keep\n", "refuse"},
		{"G02", "<<: {x: 1}\nuser:\n  name: Alice\n  github_username: keep\n", "refuse"},
		// Iteration-8 auditor shapes (inputs copied from its probe file).
		{"N01", "!!merge user:\n  name: Alice\n  github_username: keep\n", "refuse"},
		{"N02", "user:\n  !!merge foo: {name: Bob}\n  name: Alice\n", "refuse"},
		{"N03", "! <<: {user: {name: Alice}}\n", "refuse"},
		{"N04", "!foo <<: {x: 1}\nuser:\n  name: Alice\n  github_username: keep\n", "refuse"},
		{"N05", "!!binary user:\n  name: X\nuser:\n  name: Alice\n", "refuse"},
		{"N06", "!!binary dXNlcg==:\n  name: Alice\n", "refuse"},
		{"N07", "user:\n  !!binary bmFtZQ==: Alice\n  github_username: keep\n", "refuse"},
		{"N08", "user:\n  !!binary name: QWxpY2U=\n  name: Alice\n", "refuse"},
		{"N09", "user:\n  ? name\n  : Alice\n  github_username: keep\n", "clear"},
		{"N10", "user: # me\n  name: Alice # mine\n  github_username: keep\n", "clear"},
		{"N11", "user:\n  name: Alice\n  github_username: keep\n---\n", "refuse"},
		{"N12", "user:\n  name: Alice\n  github_username: keep\n...\n", "refuse"},
		{"N13", "# head\n\n# second\nuser:\n\n  # about name\n  name: Alice\n\n  github_username: keep # gh\n\n# foot\n", "clear"},
		{"N14", "Alice\n", "refuse"},
		{"N15", "user:\nname: Alice\n", "noop"},
		{"N16", "user:\n  user:\n    name: Bob\n  name: Alice\n", "clear"},
		{"N17", "user:\n  name: |\n    Alice\n    Smith\n  github_username: keep\nother: x\n", "clear"},
		{"N18", "team: Alice\nuser:\n  name: Alice\n  github_username: Alice\n", "clear"},
		{"N19", "user:\n  name: " + strings.Repeat("A", 5000) + "\n  github_username: keep\n", "clear"},
		{"N20", "user:\n  name:\tAlice\n  github_username: keep\n", "clear"},
		{"N21", "# c\r\nuser:\r\n  # n\r\n  name: \"Alice\"\r\n", "clear"},
		{"N22", "user:\n  name: Alice\n  github_username: keep", "clear"},
		{"N23", "user:\n  name: \"Alice\"", "clear"},
		{"N24", "user:\n  name: \"A: b # c 'd' \\\"e\\\"\"\n  github_username: keep\n", "clear"},
		{"N25", "user:\n  name: Alice\n  <<: {name: Bob}\n", "clear"},
		{"N26", "b: &b {name: Bob}\nuser:\n  name: Alice\n  <<: *b\n", "clear"},
		{"N27", "b: &b {name: Bob}\nuser:\n  <<: *b\n  gh: &g keep\nref: *g\n", "clear"},
		{"N28", "%YAML 1.1\n---\nuser:\n  name: Alice\n  github_username: keep\n", "clear"},
		{"N29", "user:\n  name: Alice\n---\n# trailing comment\n", "refuse"},
		{"N30", "user:\n  <<: {name: Bob}\n  <<: {gh: keep}\n", "refuse"},
		{"N31", "user: !!map &u\n  name: Alice\n", "refuse"},
		{"N32", "user:\n  name: !!null Alice\n", "refuse"},
		{"N33", "!!str user: !!map\n  name: Alice\n", "refuse"},
		{"N34", "user:\n    # who\n    name: Alice   # trailing\n    github_username: keep\n", "clear"},
		{"N35", "user: {name: Alice, github_username: keep} # flow\n", "clear"},
		{"N36", "user:\n  name: 김구스\n  github_username: keep\n", "clear"},
		{"N37", "user:\n  name: ' '\n", "clear"},
		{"N38", "k: &kk other\n*kk : 1\nuser:\n  name: Alice\n", "refuse"},
		{"N39", "&ka other: 1\nuser:\n  name: Alice\n", "refuse"},
		{"N40", "user:\n  github_username: keep\n---\nuser:\n  name: Bob\n", "noop"},
		{"X01", "user:\n  <<: {name: Alice, github_username: keep}\n  !!binary name: ignored\n", "refuse"},
		{"X02", "user:\n  name: Alice\n... # keep end comment\n# keep after end\n", "refuse"},
		{"X03", "user:\n  name: Alice\n  github_username: keep # keep end comment\n# keep after end\n", "clear"},
		{"X06", "user:\n  name: Alice\n\n# --- section two ---\n# explains other\nother:\n  x: 1 # why\n", "clear"},
		{"X07", "user:\n  name: Alice\n  # trailing note inside user\nother: 1\n", "clear"},
		{"X08", "user:\n  name: Alice\n# eof note", "clear"},
	}
	// Iteration-9 rows (inputs copied from the auditor's probe files).
	rows = append(rows,
		matrixRow{"Y01", "user: {nameAlice: , name: Alice}\nother: {null_value: }\n", "refuse"},
		matrixRow{"Y02", "user:\n  name: Alice\nother: {null_value: }\n", "clear"},
		matrixRow{"Y03", "user:\n  name: |\n    Alice\nother: {null_value: }\n", "refuse"},
		matrixRow{"Y04", "user:\n  name: |\n    Alice\nother:\n  null_value:\n", "clear"},
		matrixRow{"Y05", "user: {name: \"\\x41lice\", github_username: keep}\u0085... # gone\n# after\n", "refuse"},
		matrixRow{"Y06", "user: {name: \"\\x41lice\", github_username: keep}\n... # gone\n# after\n", "refuse"},
		matrixRow{"Y07", "user:\n  name: Alice\u0085... # c\n", "refuse"},
		matrixRow{"Y08", "user: {name: \"\\x41lice\", github_username: keep}\u2028... # gone\n# after\n", "refuse"},
	)
	// F14: whole-file bytes for the plain-name rows that go through the
	// in-place edit, so dropping Quoted:true (re-encode) is caught.
	exactAfter["A01"] = "user:\n  name: \"\"\n  github_username: keep\n"
	exactAfter["N13"] = "# head\n\n# second\nuser:\n\n  # about name\n  name: \"\"\n\n  github_username: keep # gh\n\n# foot\n"

	// P rows: the files the product itself produces must stay inside the
	// allow-list (CLEAR, effect oracle holding).
	tmpl, err := os.ReadFile(filepath.Join("..", "template", "templates", ".moai", "config", "sections", "user.yaml.tmpl"))
	if err != nil {
		t.Fatalf("read the user.yaml template: %v", err)
	}
	p01 := strings.Replace(string(tmpl), `{{yamlEscape .UserName}}`, "Alice", 1)
	if strings.Contains(p01, "{{") {
		t.Fatalf("template has directives beyond .UserName; render it properly:\n%s", p01)
	}
	// P01/X04: the product's own rendered template is edited in place - every
	// other byte, the template's comment lines included, survives.
	exactAfter["P01"] = strings.Replace(p01, `name: "Alice"`, `name: ""`, 1)
	exactAfter["X04"] = exactAfter["P01"]
	exactAfter["X05"] = strings.Replace(strings.Replace(p01, `"Alice"`, "Alice", 1), "name: Alice", `name: ""`, 1)
	rows = append(rows,
		matrixRow{"X04", p01, "clear"},
		matrixRow{"X05", strings.Replace(p01, `"Alice"`, "Alice", 1), "clear"},
		matrixRow{"P01", p01, "clear"},
		matrixRow{"P01b", "user:\n  name: \"Alice\"\n", "clear"}, // generateConfigsFallback shape
		matrixRow{"P02", initWrittenUserYAML(t), "clear"},
		matrixRow{"P03", resumeLeftoverUserYAML(t), "clear"},
	)

	for _, row := range rows {
		t.Run(row.id, func(t *testing.T) {
			path := clearUserNameFixture(t, row.content)
			root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path))))
			err := clearProjectUserName(root)
			after, _ := os.ReadFile(path)
			switch row.want {
			case "refuse":
				if err == nil || !strings.Contains(err.Error(), refuseText) {
					t.Errorf("err = %v, want the refusal", err)
				}
				if string(after) != row.content {
					t.Errorf("file modified:\n got %q\nwant %q", after, row.content)
				}
			case "noop":
				if err != nil || string(after) != row.content {
					t.Errorf("noop row: err=%v modified=%v", err, string(after) != row.content)
				}
			case "patchError":
				if err == nil || string(after) != row.content {
					t.Errorf("patchError row: err=%v modified=%v", err, string(after) != row.content)
				}
			case "clear":
				if err != nil {
					t.Fatalf("unexpected error %v", err)
				}
				assertClearEffect(t, row, []byte(row.content), after)
			case "clearBroken":
				if err != nil {
					t.Fatalf("unexpected error %v", err)
				}
				if string(after) == row.content || yamlDocCount(t, after) != yamlDocCount(t, []byte(row.content)) {
					t.Errorf("clearBroken row: file unchanged or document count changed\n%s", after)
				}
			}
		})
	}
}

// The runtime verification refuses a patch whose result is not equivalent to
// the original (here: Y03, the block-scalar name forces a re-encode that turns
// a flow null into ”), leaves the original byte-identical and removes its
// temporary copy.
func TestClearProjectUserName_VerificationRefusalLeavesNoTempFile(t *testing.T) {
	const content = "user:\n  name: |\n    Alice\nother: {null_value: }\n"
	path := clearUserNameFixture(t, content)
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path))))
	err := clearProjectUserName(root)
	if err == nil || !strings.Contains(err.Error(), refuseText) {
		t.Errorf("err = %v, want the refusal", err)
	}
	if got, _ := os.ReadFile(path); string(got) != content {
		t.Errorf("original modified:\n%s", got)
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 || entries[0].Name() != "user.yaml" {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("directory listing = %v, want only user.yaml", names)
	}
}

// The equivalence check itself, case by case (the re-encoder happens to keep
// comments, so only a direct unit test can pin each clause).
func TestClearIsEquivalent(t *testing.T) {
	const before = "# head\nuser:\n  name: Alice # who\n  gh: keep\nother: 1\n"
	cases := []struct {
		name, after string
		want        bool
	}{
		{"only name blanked", "# head\nuser:\n  name: \"\" # who\n  gh: keep\nother: 1\n", true},
		{"name null, not empty string", "# head\nuser:\n  name: # who\n  gh: keep\nother: 1\n", false},
		{"name kept", before, false},
		{"sibling changed", "# head\nuser:\n  name: \"\" # who\n  gh: lost\nother: 1\n", false},
		{"top-level value changed", "# head\nuser:\n  name: \"\" # who\n  gh: keep\nother: 2\n", false},
		{"comment lost", "user:\n  name: \"\"\n  gh: keep\nother: 1\n", false},
		{"second document appended", "# head\nuser:\n  name: \"\" # who\n  gh: keep\nother: 1\n---\nx: 1\n", false},
		{"undecodable", "user: [\n", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := clearIsEquivalent([]byte(before), []byte(c.after)); got != c.want {
				t.Errorf("clearIsEquivalent = %v, want %v", got, c.want)
			}
		})
	}
	const two = "user:\n  name: Alice\n---\nx: 1\n"
	if clearIsEquivalent([]byte(two), []byte("user:\n  name: \"\"\n")) {
		t.Error("a dropped second document must not be equivalent")
	}
}

// R01-R03 (audit N13): after clearing a user that only inherits its name
// through a merge key, the name is set again and cleared a second time. Pinned
// behavior: the second clear succeeds (the re-encoder's explicit `!!merge <<`
// spelling is accepted inside user) and the loader reads "" again.
func TestClearProjectUserName_RoundTrip(t *testing.T) {
	rows := []struct{ id, content string }{
		{"R01", "user:\n  <<: {name: Alice, github_username: keep}\n"},
		{"R02", "base: &base\n  name: Alice\n  github_username: keep\nuser:\n  <<: *base\n"},
		{"R03", "user:\n  <<: {github_username: keep}\n  name: |\n    Alice\n"},
	}
	for _, r := range rows {
		t.Run(r.id, func(t *testing.T) {
			path := clearUserNameFixture(t, r.content)
			root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path))))
			err1 := clearProjectUserName(root)
			// R03's block-scalar name forces a re-encode; whatever the first
			// clear decided, the pinned property is that a refusal never
			// alters the file and a success blanks the name.
			after1, _ := os.ReadFile(path)
			if err1 != nil {
				if string(after1) != r.content {
					t.Fatalf("refused clear altered the file:\n%s", after1)
				}
				return
			}
			if n, ok := loaderName(after1); !ok || n != "" {
				t.Fatalf("first clear: loader name %q ok=%v\n%s", n, ok, after1)
			}
			if err := yamlpatch.PatchFile(path, []yamlpatch.KeyEdit{{Path: []string{"user", "name"}, Value: "Bob"}}); err != nil {
				t.Fatalf("set name: %v", err)
			}
			set, _ := os.ReadFile(path)
			err2 := clearProjectUserName(root)
			after2, _ := os.ReadFile(path)
			if err2 != nil {
				if string(after2) != string(set) {
					t.Fatalf("refused second clear altered the file:\n%s", after2)
				}
				t.Fatalf("second clear refused (pinned: must succeed): %v\n%s", err2, set)
			}
			if n, ok := loaderName(after2); !ok || n != "" {
				t.Errorf("second clear: loader name %q ok=%v\n%s", n, ok, after2)
			}
		})
	}
}

// initWrittenUserYAML returns the user.yaml a fresh `moai init` writes when the
// wizard supplies the name "Alice".
func initWrittenUserYAML(t *testing.T) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{UserName: "Alice", GitMode: "manual"}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })
	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })
	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })
	projectDir := filepath.Join(t.TempDir(), "p02-proj")
	if err := runInit(newInitTestCmd(), []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "user.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if n, ok := loaderName(data); !ok || n != "Alice" {
		t.Fatalf("precondition: init-written user.yaml name = %q (readable=%v)\n%s", n, ok, data)
	}
	return string(data)
}

// resumeLeftoverUserYAML returns the user.yaml an interrupted init leaves behind.
func resumeLeftoverUserYAML(t *testing.T) string {
	t.Helper()
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "p03-proj")
	resumeSeedRun(t, projectDir, "claude")
	data, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "user.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if n, ok := loaderName(data); !ok || n == "" {
		t.Fatalf("precondition: the interrupted run left no non-empty name (name=%q readable=%v):\n%s", n, ok, data)
	}
	return string(data)
}
func TestClearProjectUserName_UnrelatedAnchorStillClears(t *testing.T) {
	path := clearUserNameFixture(t, "other: &x 1\nuser:\n  name: \"Alice\"\n  note: keep\n")
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path))))
	if err := clearProjectUserName(root); err != nil {
		t.Fatalf("unrelated anchor: unexpected error %v", err)
	}
	got, _ := os.ReadFile(path)
	if strings.Contains(string(got), "Alice") || !strings.Contains(string(got), "other: &x 1") {
		t.Errorf("name not cleared or sibling disturbed:\n%s", got)
	}
}

// N4: a whitespace-only name is non-empty and is cleared.
func TestClearProjectUserName_WhitespaceNameIsCleared(t *testing.T) {
	path := clearUserNameFixture(t, "user:\n  name: \"  \"\n")
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path))))
	if err := clearProjectUserName(root); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(path); strings.Contains(string(got), `"  "`) {
		t.Errorf("whitespace-only name survived the clear:\n%s", got)
	}
}

func TestClearProjectUserName_NoUserMappingStaysByteIdentical(t *testing.T) {
	const content = "foo: bar\n"
	path := clearUserNameFixture(t, content)
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path))))
	if err := clearProjectUserName(root); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(path); string(got) != content {
		t.Errorf("file without a user mapping was rewritten:\n got %q\nwant %q", got, content)
	}
}
