package cli

// Card t1966 — `moai init` wizard/flag regression tests (items 1, 2, 7).

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

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
	if err := clearProjectUserName(root); err != nil {
		t.Fatalf("duplicate-key user.yaml: unexpected error %v", err)
	}
	// Iteration-2 behavior: the patch blanks the first (effective) name row.
	if got, _ := os.ReadFile(path); !strings.HasPrefix(string(got), "user:\n  name: \"\"\n") {
		t.Errorf("duplicate-key user.yaml was not patched (name row not blanked):\n%s", got)
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
	if err == nil || !strings.Contains(err.Error(), "uses YAML anchors/aliases; edit user.name by hand") {
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

// YAML indirection matrix: every way user.name can reach user.yaml and still be
// non-empty for the struct decoder. REFUSE rows must return the anchors/aliases
// error with the bytes unchanged; CLEAR rows must blank the name, keep every
// other decoded key, and append no new top-level key.
func TestClearProjectUserName_IndirectionMatrix(t *testing.T) {
	const (
		refuse = "REFUSE"
		clear  = "CLEAR"
	)
	rows := []struct {
		name, content, want string
	}{
		{"plain direct path", "user:\n  name: Alice\n  github_username: keep\n", clear},
		{"flow-style user", "user: {name: Alice, github_username: keep}\n", clear},
		{"anchor on user node", "user: &u\n  name: Alice\nother: *u\n", refuse},
		{"anchor on name node", "user:\n  name: &n Alice\n  github_username: keep\n", refuse},
		{"anchor on name key node", "user:\n  &nk name: Alice\n  github_username: keep\n", clear},
		{"name anchor shared with sibling", "user:\n  name: &identity Alice\n  github_username: *identity\n", refuse},
		{"unreferenced sibling anchor inside user", "user:\n  name: Alice\n  github_username: &g keep\n  alias: *g\n", clear},
		{"alias as user value", "base: &identity\n  name: Alice\nuser: *identity\n", refuse},
		{"alias as name value", "base: &identity Alice\nuser:\n  name: *identity\n", refuse},
		{"alias as top-level key", "k: &k user\n*k :\n  name: Alice\n  github_username: keep\n", refuse},
		{"top-level merge literal", "<<:\n  user:\n    name: Alice\n    github_username: keep\n", refuse},
		{"top-level merge alias", "b: &b\n  user:\n    name: Alice\n    github_username: keep\nz: 1\n<<: *b\n", refuse},
		{"user merge key, name inside base", "base: &base\n  name: Alice\n  github_username: keep\nuser:\n  <<: *base\n", clear},
		{"user merge key, explicit name wins", "base: &base\n  github_username: keep\nuser:\n  <<: *base\n  name: Alice\n", clear},
		{"unrelated top-level anchor", "other: &x 1\nuser:\n  name: Alice\n  github_username: keep\n", clear},
		{"merge key under non-user nested key", "b: &b\n  x: 1\nuser:\n  name: Alice\n  github_username: keep\nnested:\n  <<: *b\n", clear},
		{"second document after ---", "user:\n  name: Alice\n  github_username: keep\n---\nuser:\n  name: Bob\n", clear},
	}
	decode := func(t *testing.T, b []byte) map[string]any {
		t.Helper()
		var m map[string]any
		if err := yaml.Unmarshal(b, &m); err != nil {
			t.Fatalf("decode: %v\n%s", err, b)
		}
		return m
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			path := clearUserNameFixture(t, row.content)
			root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path))))
			err := clearProjectUserName(root)
			after, _ := os.ReadFile(path)
			switch row.want {
			case refuse:
				if err == nil || !strings.Contains(err.Error(), "uses YAML anchors/aliases; edit user.name by hand") {
					t.Errorf("err = %v, want the anchors/aliases refusal", err)
				}
				if string(after) != row.content {
					t.Errorf("file modified:\n got %q\nwant %q", after, row.content)
				}
			case clear:
				if err != nil {
					t.Fatalf("unexpected error %v", err)
				}
				before, now := decode(t, []byte(row.content)), decode(t, after)
				if len(before) != len(now) {
					t.Errorf("top-level keys changed:\n before %v\n after  %v\n%s", before, now, after)
				}
				bu, _ := before["user"].(map[string]any)
				au, _ := now["user"].(map[string]any)
				// An unquoted scalar is patched to an empty row (`name:`), which a
				// map decode reads as nil; the typed config loader reads "".
				if au["name"] != nil && au["name"] != "" {
					t.Errorf("user.name = %v after clear, want empty\n%s", au["name"], after)
				}
				for k, v := range bu {
					if k != "name" && !reflect.DeepEqual(au[k], v) {
						t.Errorf("sibling %q = %v, want %v\n%s", k, au[k], v, after)
					}
				}
			}
		})
	}
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
