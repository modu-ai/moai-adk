package cli

// SPEC-GLM-JEV-KEY-001 run-phase tests: the `moai glm --key <api-key>` save
// flag and the new top-level `moai jev --key <credential>` command. Both
// surfaces store through the owning credential packages (internal/glmcred /
// internal/jevcred) and never echo the full key.
//
// Isolation: every test redirects the credential location through the
// package's existing seams only — userHomeDirFn (aliased to glmcred.HomeDirFn
// in glm.go's init) and jevcred.HomeDirFn — never the developer's real HOME
// (plan.md §D.6). The tests swap package variables, so none of them may run
// in parallel.

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/jevcred"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// redirectCredentialHomes points every credential-location seam at a fresh
// temp home and drops the TestMain MOAI_HOME sandbox so the lookups derive
// from the redirected home (glm_test.go precedent, card t1229). Returns the
// home path.
func redirectCredentialHomes(t *testing.T) string {
	t.Helper()
	t.Setenv(config.EnvHome, "")
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	origHomeFn := userHomeDirFn
	userHomeDirFn = func() (string, error) { return tmpHome, nil }
	t.Cleanup(func() { userHomeDirFn = origHomeFn })

	origJevHomeFn := jevcred.HomeDirFn
	jevcred.HomeDirFn = func() (string, error) { return tmpHome, nil }
	t.Cleanup(func() { jevcred.HomeDirFn = origJevHomeFn })

	return tmpHome
}

// resetCommandFlags clears the values and Changed marks a command tree's
// flags carry over from earlier executions: pflag keeps both across parses,
// so a prior `--help` would flip every later run of the same command into
// cobra's help path before RunE (observed: TestJevHelpDocumentsKeyFlag left
// jev's help flag set and TestJevKeyFlagSavesCredential then stored nothing).
func resetCommandFlags(c *cobra.Command) {
	c.Flags().VisitAll(func(f *pflag.Flag) {
		f.Changed = false
		if f.Value.String() != f.DefValue {
			_ = f.Value.Set(f.DefValue)
		}
	})
	for _, sub := range c.Commands() {
		resetCommandFlags(sub)
	}
}

// resetCommandOutputs clears the output writers a command tree carries over
// from earlier executions: the pre-existing glm tests call glmCmd.SetOut and
// never reset it, which shadows execRoot's rootCmd.SetOut for every later run
// (observed as empty captured output in the mixed glm-family subset run).
func resetCommandOutputs(c *cobra.Command) {
	c.SetOut(nil)
	c.SetErr(nil)
	for _, sub := range c.Commands() {
		resetCommandOutputs(sub)
	}
}

// execRoot runs the root command against args with stdout and stderr captured
// into one buffer, resetting the root command state afterwards.
func execRoot(t *testing.T, args ...string) (string, error) {
	t.Helper()
	resetCommandFlags(rootCmd)
	resetCommandOutputs(rootCmd)
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})
	err := rootCmd.Execute()
	return buf.String(), err
}

// nothingStoredAt asserts the credential file does not exist.
func nothingStoredAt(t *testing.T, envPath string) {
	t.Helper()
	if _, statErr := os.Stat(envPath); !os.IsNotExist(statErr) {
		t.Errorf("nothing must be stored at %s, stat err: %v", envPath, statErr)
	}
}

// overrideLaunch captures the child arguments the launch path would pass to
// claude. Without it, a test whose input reaches the real launch path replaces
// the test process with the claude binary via syscall.Exec
// (launcher.go launchClaudeDefault) — the RED run observed exactly that.
func overrideLaunch(t *testing.T) *[]string {
	t.Helper()
	captured := new([]string)
	origLaunch := launchClaudeFunc
	launchClaudeFunc = func(profile string, args []string) error {
		*captured = args
		return nil
	}
	t.Cleanup(func() { launchClaudeFunc = origLaunch })
	return captured
}

// AC-GJK-001 — root help lists the new command.
func TestRootHelpListsJevCommand(t *testing.T) {
	redirectCredentialHomes(t)
	t.Setenv("NO_COLOR", "1")
	reorderRootHelpCommands(rootCmd)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	if err := runFang(context.Background(), rootCmd); err != nil {
		t.Fatalf("runFang --help: %v", err)
	}
	if out := buf.String(); !strings.Contains(out, "jev") {
		t.Errorf("root help should list the jev command, got:\n%s", out)
	}
}

// AC-GJK-002 — glm help documents the flag, keeping the setup sentence.
func TestGlmHelpDocumentsKeyFlag(t *testing.T) {
	redirectCredentialHomes(t)
	out, err := execRoot(t, "glm", "--help")
	if err != nil {
		t.Fatalf("glm --help should not error: %v", err)
	}
	if !strings.Contains(out, "--key") {
		t.Errorf("glm help should document --key, got:\n%s", out)
	}
	if !strings.Contains(out, "setup") {
		t.Errorf("glm help should keep the setup sentence, got:\n%s", out)
	}
}

// AC-GJK-003 — jev help documents the flag.
func TestJevHelpDocumentsKeyFlag(t *testing.T) {
	redirectCredentialHomes(t)
	out, err := execRoot(t, "jev", "--help")
	if err != nil {
		t.Fatalf("jev --help should not error, got: %v", err)
	}
	if !strings.Contains(out, "--key") {
		t.Errorf("jev help should document --key, got:\n%s", out)
	}
}

// AC-GJK-004 — glm --key save lands on disk: same file, same dotenv form, mode
// 0600, the setup path's masked confirmation, no full-key echo (REQ-GJK-001/002).
func TestGlmKeyFlagSavesKeySameStorage(t *testing.T) {
	home := redirectCredentialHomes(t)
	const key = "test-key-1234567890"
	out, err := execRoot(t, "glm", "--key", key)
	if err != nil {
		t.Fatalf("glm --key should store and exit cleanly, got: %v", err)
	}
	envPath := filepath.Join(home, ".moai", ".env.glm")
	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("expected .env.glm at %s: %v", envPath, err)
	}
	if !strings.Contains(string(data), `GLM_API_KEY="`+key+`"`) {
		t.Errorf("expected stored dotenv form, got:\n%s", data)
	}
	info, err := os.Stat(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("credential file mode = %v, want 0600", info.Mode().Perm())
	}
	if !strings.Contains(out, "GLM API key stored (") {
		t.Errorf("masked confirmation missing, got: %q", out)
	}
	if strings.Contains(out, key) {
		t.Errorf("full key must never be echoed, got: %q", out)
	}
}

// AC-GJK-005 — jev --key save lands on disk: mode 0600, dotenv form,
// confirmation disclosing at most the final four characters, no full-credential
// echo (REQ-GJK-007/008).
func TestJevKeyFlagSavesCredential(t *testing.T) {
	home := redirectCredentialHomes(t)
	const cred = "tsk-cred-1234567890"
	out, err := execRoot(t, "jev", "--key", cred)
	if err != nil {
		t.Fatalf("jev --key should store and exit cleanly, got: %v", err)
	}
	envPath := filepath.Join(home, ".moai", ".env.typesafe")
	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("expected .env.typesafe at %s: %v", envPath, err)
	}
	if !strings.Contains(string(data), `TYPESAFE_API_KEY="`+cred+`"`) {
		t.Errorf("expected stored dotenv form, got:\n%s", data)
	}
	info, err := os.Stat(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("credential file mode = %v, want 0600", info.Mode().Perm())
	}
	if !strings.Contains(out, "Jev credential stored") {
		t.Errorf("confirmation missing, got: %q", out)
	}
	if strings.Contains(out, cred) {
		t.Errorf("full credential must never be echoed, got: %q", out)
	}
	if !strings.Contains(out, cred[len(cred)-4:]) {
		t.Errorf("confirmation should disclose the final four characters, got: %q", out)
	}
}

// AC-GJK-009 — extra-argument refusal: a usage error naming the conflict, and
// nothing stored (REQ-GJK-003).
func TestGlmKeyFlagRefusesExtraArgs(t *testing.T) {
	home := redirectCredentialHomes(t)
	overrideLaunch(t)
	for _, args := range [][]string{
		{"glm", "--key", "K", "-f"},
		{"glm", "--key", "K", "status"},
	} {
		out, err := execRoot(t, args...)
		if err == nil {
			t.Errorf("args %v should refuse with a usage error, got success (%q)", args, out)
			continue
		}
		if !strings.Contains(err.Error(), "--key") || !strings.Contains(err.Error(), "combined") {
			t.Errorf("usage error must name the --key conflict, got: %v", err)
		}
		nothingStoredAt(t, filepath.Join(home, ".moai", ".env.glm"))
	}
}

// AC-GJK-011 — glm --key with no following value fails with the missing-value
// usage error and stores nothing (REQ-GJK-004).
func TestGlmKeyFlagMissingValueErrors(t *testing.T) {
	home := redirectCredentialHomes(t)
	overrideLaunch(t)
	_, err := execRoot(t, "glm", "--key")
	if err == nil {
		t.Fatal("glm --key with no value should fail")
	}
	if !strings.Contains(err.Error(), "--key requires a value") {
		t.Errorf("missing-value usage error expected, got: %v", err)
	}
	nothingStoredAt(t, filepath.Join(home, ".moai", ".env.glm"))
}

// AC-GJK-011 — glm --key with an empty-after-trim value fails with the setup
// path's empty-key error and stores nothing (REQ-GJK-005).
func TestGlmKeyEmptyValueErrors(t *testing.T) {
	home := redirectCredentialHomes(t)
	overrideLaunch(t)
	for _, v := range []string{"", "   "} {
		_, err := execRoot(t, "glm", "--key", v)
		if err == nil {
			t.Errorf("glm --key %q should fail", v)
			continue
		}
		if !strings.Contains(err.Error(), "empty API key") {
			t.Errorf("error should be the setup path's empty-key error, got: %v", err)
		}
		nothingStoredAt(t, filepath.Join(home, ".moai", ".env.glm"))
	}
}

// AC-GJK-010 — bare jev invocation prints help, exits 0, storage untouched
// (REQ-GJK-009).
func TestJevBareInvocationPrintsHelpExitZero(t *testing.T) {
	home := redirectCredentialHomes(t)
	out, err := execRoot(t, "jev")
	if err != nil {
		t.Fatalf("bare jev should print help and exit 0, got error: %v", err)
	}
	if !strings.Contains(out, "--key") {
		t.Errorf("bare jev should print its help text, got: %q", out)
	}
	nothingStoredAt(t, filepath.Join(home, ".moai", ".env.typesafe"))
}

// AC-GJK-011 — jev --key with an empty-after-trim value fails with jev's
// empty-credential error and stores nothing.
func TestJevKeyEmptyValueErrors(t *testing.T) {
	home := redirectCredentialHomes(t)
	for _, v := range []string{"", "   "} {
		_, err := execRoot(t, "jev", "--key", v)
		if err == nil {
			t.Errorf("jev --key %q should fail", v)
			continue
		}
		if !strings.Contains(err.Error(), "empty Jev credential") {
			t.Errorf("jev empty-credential error expected, got: %v", err)
		}
		nothingStoredAt(t, filepath.Join(home, ".moai", ".env.typesafe"))
	}
}

// AC-GJK-015 — glm --key newline value refused before the writer; the stored
// key preserved byte-for-byte (REQ-GJK-012).
func TestGlmKeyFlagNewlineValueRefusesAndPreserves(t *testing.T) {
	home := redirectCredentialHomes(t)
	overrideLaunch(t) // the seeded key would otherwise carry the run into syscall.Exec
	if err := saveGLMKey("stored-key-1234"); err != nil {
		t.Fatalf("seed save failed: %v", err)
	}
	envPath := filepath.Join(home, ".moai", ".env.glm")
	before, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	_, execErr := execRoot(t, "glm", "--key", "first\nsecond")
	if execErr == nil {
		t.Fatal("a newline-bearing key must be refused")
	}
	if !strings.Contains(execErr.Error(), "line breaks") {
		t.Errorf("validation error expected, got: %v", execErr)
	}
	after, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("credential file must stay byte-for-byte identical\nbefore: %q\nafter:  %q", before, after)
	}
}

// REQ-GJK-004 — a separated value that is flag-shaped after trimming is never
// consumed as the key. A bare "-" was exempt from the shape check and a padded
// " -f" was checked before the trim, so both were stored (card t1626). Each
// refusal must leave the credential file byte-for-byte unchanged and print no
// confirmation (AC-GJK-011). The "-f" row is the control: that spelling was
// already refused.
func TestGlmKeyRefusedDashValueRows(t *testing.T) {
	home := redirectCredentialHomes(t)
	overrideLaunch(t)
	envPath := filepath.Join(home, ".moai", ".env.glm")
	for _, value := range []string{"-", " -f", "-f"} {
		t.Run("value="+value, func(t *testing.T) {
			if err := saveGLMKey("stored-key-1234"); err != nil {
				t.Fatalf("seed save failed: %v", err)
			}
			before, err := os.ReadFile(envPath)
			if err != nil {
				t.Fatal(err)
			}
			out, execErr := execRoot(t, "glm", "--key", value)
			if execErr == nil {
				t.Errorf("glm --key %q must be refused, got success (output %q)", value, out)
			}
			if strings.Contains(out, "GLM API key stored") {
				t.Errorf("no confirmation may print for a refused value, got: %q", out)
			}
			after, err := os.ReadFile(envPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Errorf("credential file must stay byte-for-byte identical\nbefore: %q\nafter:  %q", before, after)
			}
		})
	}
}

// AC-GJK-014 — jev newline value refused before the writer; the stored
// credential preserved byte-for-byte (REQ-GJK-012).
func TestJevKeyNewlineValueRefusesAndPreserves(t *testing.T) {
	home := redirectCredentialHomes(t)
	if err := jevcred.Save("stored-cred-1234"); err != nil {
		t.Fatalf("seed save failed: %v", err)
	}
	envPath := filepath.Join(home, ".moai", ".env.typesafe")
	before, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	_, execErr := execRoot(t, "jev", "--key", "first\nsecond")
	if execErr == nil {
		t.Fatal("a newline-bearing credential must be refused")
	}
	if !strings.Contains(execErr.Error(), "line breaks") {
		t.Errorf("validation error expected, got: %v", execErr)
	}
	after, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("credential file must stay byte-for-byte identical\nbefore: %q\nafter:  %q", before, after)
	}
}

// AC-GJK-016 — post-`--` tokens are child passthrough, never scanned. RED by
// MUTANT (plan.md §F): against a whole-args scan (the mutant) a save
// confirmation appears where passthrough is required and this test fails;
// against a no-scan tree it is vacuously green by design.
func TestGlmKeyAfterDashDashPassthrough(t *testing.T) {
	home := redirectCredentialHomes(t)
	const key = "test-key-1234567890"
	launchedArgs := overrideLaunch(t)

	out, err := execRoot(t, "glm", "--", "--key", key)
	nothingStoredAt(t, filepath.Join(home, ".moai", ".env.glm"))
	if strings.Contains(out, "GLM API key stored") {
		t.Errorf("no save confirmation may appear for a post--- token, got: %q", out)
	}
	if err != nil && *launchedArgs == nil {
		// The launch path's own refusal is an acceptable observation — the
		// tokens reached the launch path, never a save.
		t.Logf("launch path refused with: %v", err)
	}
	if *launchedArgs != nil && !slices.Contains(*launchedArgs, "--key") {
		t.Errorf("post--- tokens must reach the child unmodified, got: %v", *launchedArgs)
	}
}

// AC-GJK-011 family edge — the `--key=<value>` single-token spelling is
// accepted by the scan and stores identically (acceptance.md §B edge case).
func TestGlmKeyEqualsFormSaves(t *testing.T) {
	home := redirectCredentialHomes(t)
	const key = "test-key-eq-12345678"
	out, err := execRoot(t, "glm", "--key="+key)
	if err != nil {
		t.Fatalf("glm --key=<value> should store, got: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(home, ".moai", ".env.glm"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `GLM_API_KEY="`+key+`"`) {
		t.Errorf("expected stored dotenv form, got:\n%s", data)
	}
	if !strings.Contains(out, "GLM API key stored (") {
		t.Errorf("masked confirmation missing, got: %q", out)
	}

	// The '=' spelling is unambiguous, so a flag-shaped value stays a value.
	if _, err := execRoot(t, "glm", "--key=-f"); err != nil {
		t.Fatalf("glm --key=-f should store the value, got: %v", err)
	}
	data, err = os.ReadFile(filepath.Join(home, ".moai", ".env.glm"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `GLM_API_KEY="-f"`) {
		t.Errorf("expected the flag-shaped value stored verbatim, got:\n%s", data)
	}
}

// Gate repair P1 — REQ-GJK-010: a conflicting second --key spelling must
// never surface its value in the refusal (observed leak: the whole second
// key appeared verbatim in the usage error).
func TestGlmKeyConflictErrorMasksValue(t *testing.T) {
	home := redirectCredentialHomes(t)
	_, err := execRoot(t, "glm", "--key", "K1", "--key=sk-secret-9999")
	if err == nil {
		t.Fatal("a mixed double-key invocation must be refused")
	}
	if strings.Contains(err.Error(), "sk-secret-9999") {
		t.Errorf("refusal must not disclose the second key value, got: %v", err)
	}
	if !strings.Contains(err.Error(), "****") {
		t.Errorf("refusal should show the masked form, got: %v", err)
	}
	nothingStoredAt(t, filepath.Join(home, ".moai", ".env.glm"))
}

// Gate repair P2 — REQ-GJK-003: launch flags before --key are also tokens in
// the scanned region; the mixed invocation must refuse and store nothing
// (observed defect: `-p work --key K` stored K).
func TestGlmKeyLeadingArgsRefused(t *testing.T) {
	home := redirectCredentialHomes(t)
	overrideLaunch(t)
	_, err := execRoot(t, "glm", "-p", "work", "--key", "K")
	if err == nil {
		t.Fatal("a launch-flag mixed invocation must be refused")
	}
	if !strings.Contains(err.Error(), "--key") {
		t.Errorf("usage error should name the conflict, got: %v", err)
	}
	nothingStoredAt(t, filepath.Join(home, ".moai", ".env.glm"))
}

// Gate repair P2 — REQ-GJK-004: a flag-shaped token after bare --key is a
// missing value, never the value itself (observed defect: `-f` was stored).
func TestGlmKeyExecFlagAsValueRefused(t *testing.T) {
	home := redirectCredentialHomes(t)
	overrideLaunch(t)
	_, err := execRoot(t, "glm", "--key", "-f")
	if err == nil {
		t.Fatal("a flag-shaped token must not be stored as the key")
	}
	if !strings.Contains(err.Error(), "--key requires a value") {
		t.Errorf("missing-value usage error expected, got: %v", err)
	}
	nothingStoredAt(t, filepath.Join(home, ".moai", ".env.glm"))
}

// Card-review repair P1 — REQ-GJK-010: a positional duplicate of the key
// value flows into the conflict error too (observed leak: redactArg passed
// '='-less tokens through verbatim, and a key value IS a positional token
// here).
func TestGlmKeyPositionalDuplicateMasked(t *testing.T) {
	home := redirectCredentialHomes(t)
	_, err := execRoot(t, "glm", "--key", "sk-secret-9999", "sk-secret-9999")
	if err == nil {
		t.Fatal("a positional duplicate beside --key must be refused")
	}
	if strings.Contains(err.Error(), "sk-secret-9999") {
		t.Errorf("refusal must not disclose the key value, got: %v", err)
	}
	if !strings.Contains(err.Error(), "****") {
		t.Errorf("refusal should show the masked form, got: %v", err)
	}
	nothingStoredAt(t, filepath.Join(home, ".moai", ".env.glm"))
}

// Card-review repair P2 — pflag consumes an option token as the string value,
// so `moai jev --key --help` reaches runJev with the literal "--help" and
// would overwrite the stored credential with it. A '-'-prefixed value is
// refused before the writer. glm keeps accepting the explicit `--key=-f`
// spelling because its manual scan can tell the two spellings apart; pflag
// cannot, so both are refused on the jev side (credentials do not start
// with '-').
func TestJevKeyOptionTokenValueRefused(t *testing.T) {
	home := redirectCredentialHomes(t)
	if err := jevcred.Save("stored-cred-1234"); err != nil {
		t.Fatalf("seed save failed: %v", err)
	}
	envPath := filepath.Join(home, ".moai", ".env.typesafe")
	before, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	_, execErr := execRoot(t, "jev", "--key", "--help")
	if execErr == nil {
		t.Fatal("an option-shaped value must be refused, not stored")
	}
	after, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("credential file must stay byte-for-byte identical\nbefore: %q\nafter:  %q", before, after)
	}
}

// Card-review r2 repair — REQ-GJK-010: redactArg's shape heuristics leak
// adversarial key values — a '='-bearing value kept its prefix, a
// dash-leading value passed as a flag name. Only confirmed flag names may
// pass into the usage error; every other token is '****'.
func TestGlmKeyEqualsInsideValueMasked(t *testing.T) {
	home := redirectCredentialHomes(t)
	_, err := execRoot(t, "glm", "--key", "sk-secret=", "sk-secret=")
	if err == nil {
		t.Fatal("a positional duplicate beside --key must be refused")
	}
	if strings.Contains(err.Error(), "sk-secret=") {
		t.Errorf("refusal must not disclose the key value, got: %v", err)
	}
	if !strings.Contains(err.Error(), "****") {
		t.Errorf("refusal should show the masked form, got: %v", err)
	}
	nothingStoredAt(t, filepath.Join(home, ".moai", ".env.glm"))
}

// Card-review r2 repair — REQ-GJK-010: a dash-leading key value must not pass
// the redaction as if it were a flag name.
func TestGlmKeyDashLeadingDuplicateMasked(t *testing.T) {
	home := redirectCredentialHomes(t)
	_, err := execRoot(t, "glm", "--key=-foo", "-foo")
	if err == nil {
		t.Fatal("a dash-leading positional duplicate must be refused")
	}
	if strings.Contains(err.Error(), "-foo") {
		t.Errorf("refusal must not disclose the key value, got: %v", err)
	}
	nothingStoredAt(t, filepath.Join(home, ".moai", ".env.glm"))
}

// Sync-review repair P1 — REQ-GJK-010: pflag echoes the offending token
// verbatim in parse errors, so `moai jev --key <credential> --help=<same
// value>` surfaces the credential before runJev's validation ever runs
// (observed leak: the value appeared in both the error and stderr).
func TestJevFlagParseErrorDoesNotEchoKey(t *testing.T) {
	home := redirectCredentialHomes(t)
	const key = "tsk-secret-1234"
	out, err := execRoot(t, "jev", "--key", key, "--help="+key)
	if err == nil {
		t.Fatal("an invalid flag spelling must be refused")
	}
	if strings.Contains(err.Error(), key) || strings.Contains(out, key) {
		t.Errorf("flag-parse refusal must not echo the key value\nerr: %v\nout: %q", err, out)
	}
	nothingStoredAt(t, filepath.Join(home, ".moai", ".env.typesafe"))
}

// Sync-review repair P2 — jev never touches the deps global (runJev reads
// only flags and the jevcred writer), so it belongs on the lazy-init fast
// path like the launcher commands; otherwise even `moai jev --key x` would
// run the full InitDependencies.
func TestJevIsTrivialCommand(t *testing.T) {
	if !isTrivialCommand([]string{"jev"}) {
		t.Error("jev must be in the trivialCommands fast path (deps-free)")
	}
	if !isTrivialCommand([]string{"jev", "--key", "x"}) {
		t.Error("jev with --key must also skip InitDependencies")
	}
}

// AC-GJK-008 — routing precedence: the scan never intercepts a routed
// subcommand (M2 characterization, REQ-GJK-006).
func TestGlmSetupRoutingUnchanged(t *testing.T) {
	home := redirectCredentialHomes(t)
	out, err := execRoot(t, "glm", "setup", "K2")
	if err != nil {
		t.Fatalf("glm setup should keep working, got: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(home, ".moai", ".env.glm"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `GLM_API_KEY="K2"`) {
		t.Errorf("setup path must store the value, got:\n%s", data)
	}
	if !strings.Contains(out, "GLM API key stored") {
		t.Errorf("setup confirmation missing, got: %q", out)
	}
}

// AC-GJK-007 — both forms write the SAME storage, last writer wins (M2
// characterization).
func TestKeyFormsShareStorageLastWriterWins(t *testing.T) {
	home := redirectCredentialHomes(t)
	envPath := filepath.Join(home, ".moai", ".env.glm")
	readStored := func() string {
		t.Helper()
		data, err := os.ReadFile(envPath)
		if err != nil {
			t.Fatalf("read stored file: %v", err)
		}
		return string(data)
	}

	if _, err := execRoot(t, "glm", "--key", "AAA"); err != nil {
		t.Fatalf("glm --key AAA: %v", err)
	}
	if _, err := execRoot(t, "glm", "setup", "BBB"); err != nil {
		t.Fatalf("glm setup BBB: %v", err)
	}
	if got := readStored(); !strings.Contains(got, `GLM_API_KEY="BBB"`) {
		t.Errorf("last writer (setup BBB) must win, got:\n%s", got)
	}

	if _, err := execRoot(t, "glm", "setup", "AAA"); err != nil {
		t.Fatalf("glm setup AAA: %v", err)
	}
	if _, err := execRoot(t, "glm", "--key", "BBB"); err != nil {
		t.Fatalf("glm --key BBB: %v", err)
	}
	if got := readStored(); !strings.Contains(got, `GLM_API_KEY="BBB"`) {
		t.Errorf("last writer (--key BBB) must win, got:\n%s", got)
	}
}
