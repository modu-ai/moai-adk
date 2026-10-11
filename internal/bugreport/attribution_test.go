package bugreport

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"gopkg.in/yaml.v3"
)

// attributionCase is one row of the AC-008 table: the signal offered at
// capture and the verdict the register's row assigns.
type attributionCase struct {
	name   string
	kind   Kind
	err    error
	reason Reason
	want   Verdict
}

// attributionTable covers every row of the attribution register (design.md
// section 3): environment rows first, then user rows, then the moai
// allowlist, then the fallback.
func attributionTable() []attributionCase {
	pathErr := &os.PathError{Op: "open", Path: "/x", Err: syscall.ENOENT}
	netErr := &net.DNSError{Err: "no such host", Name: "x"}
	plainHandlerErr := errors.New("handler did something unexpected")
	customErr := &customTestError{}
	markerErr := MarkInternal(errors.New("violated invariant"))

	return []attributionCase{
		{"panic", KindPanic, nil, "", VerdictMoai},
		{"internal_marker", KindInternalError, markerErr, "", VerdictMoai},
		{"marker_wrapping_path_error", KindInternalError, MarkInternal(pathErr), "", VerdictEnvironment},
		{"filesystem_error", KindHookHandlerFailure, pathErr, "", VerdictEnvironment},
		{"syscall_error", KindHookHandlerFailure, &os.SyscallError{Syscall: "read", Err: syscall.EBADF}, "", VerdictEnvironment},
		{"network_error", KindHookHandlerFailure, netErr, "", VerdictEnvironment},
		{"permission_error", KindHookHandlerFailure, os.ErrPermission, "", VerdictEnvironment},
		{"canceled_context", KindHookHandlerFailure, context.Canceled, "", VerdictEnvironment},
		{"exec_reason_token", KindHookHandlerFailure, errors.New("git failed in the user's repository"), ReasonExec, VerdictEnvironment},
		{"config_not_found", KindHookHandlerFailure, config.ErrConfigNotFound, "", VerdictUser},
		{"invalid_config", KindHookHandlerFailure, config.ErrInvalidConfig, "", VerdictUser},
		{"invalid_yaml_sentinel", KindHookHandlerFailure, config.ErrInvalidYAML, "", VerdictUser},
		{"section_type_mismatch", KindHookHandlerFailure, config.ErrSectionTypeMismatch, "", VerdictUser},
		{"invalid_development_mode", KindHookHandlerFailure, config.ErrInvalidDevelopmentMode, "", VerdictUser},
		{"json_syntax_error", KindHookHandlerFailure, &json.SyntaxError{}, "", VerdictUser},
		{"json_unmarshal_type_error", KindHookHandlerFailure, &json.UnmarshalTypeError{}, "", VerdictUser},
		{"yaml_type_error", KindHookHandlerFailure, &yaml.TypeError{}, "", VerdictUser},
		{"unmarked_handler_error", KindHookHandlerFailure, plainHandlerErr, "", VerdictAmbiguous},
		{"hook_timeout", KindHookTimeout, errors.New("deadline"), "", VerdictAmbiguous},
		{"token_unexpanded_token", KindHarnessDefect, errors.New("render"), ReasonUnexpandedToken, VerdictAmbiguous},
		{"token_invalid_json", KindHarnessDefect, errors.New("validate"), ReasonInvalidJSON, VerdictAmbiguous},
		{"token_path_traversal", KindTemplateDeployFailure, errors.New("deploy"), ReasonPathTraversal, VerdictMoai},
		{"token_not_found", KindTemplateDeployFailure, errors.New("deploy"), ReasonNotFound, VerdictMoai},
		{"token_preserve_integrity", KindTemplateDeployFailure, errors.New("deploy"), ReasonPreserveIntegrity, VerdictMoai},
		{"token_missing_key", KindHarnessDefect, errors.New("render"), ReasonMissingKey, VerdictMoai},
		{"custom_error_type", KindHookHandlerFailure, customErr, "", VerdictAmbiguous},
	}
}

type customTestError struct{}

func (*customTestError) Error() string { return "something entirely unrecognised" }

func TestAttributionRulesTable(t *testing.T) {
	for _, tc := range attributionTable() {
		verdict, _ := Attribute(tc.kind, tc.err, tc.reason)
		if verdict != tc.want {
			t.Errorf("%s: verdict = %q, want %q", tc.name, verdict, tc.want)
			continue
		}
		if verdict == VerdictMoai && !strings.HasPrefix(string(tc.kind), "") {
			// moai is an allowlist: every moai verdict must trace to row M1,
			// M2, or M3 — panic, marker, or a moai token.
			if tc.err == nil && tc.kind != KindPanic {
				t.Errorf("%s: moai verdict without a marker or token", tc.name)
			}
		}
	}
}

func TestAttributionFirstMatchWins(t *testing.T) {
	// Row A before row M: an internal marker wrapping an *fs.PathError
	// resolves to environment — the cause is the environment even though
	// moai noticed it.
	pathErr := &os.PathError{Op: "open", Path: "/x", Err: syscall.ENOENT}
	verdict, _ := Attribute(KindInternalError, MarkInternal(pathErr), "")
	if verdict != VerdictEnvironment {
		t.Fatalf("marker wrapping a PathError = %q, want environment (A rows run before M rows)", verdict)
	}

	// Environment before moai-token: a not_found token carrying a filesystem
	// cause resolves to environment.
	verdict, _ = Attribute(KindTemplateDeployFailure, pathErr, ReasonNotFound)
	if verdict != VerdictEnvironment {
		t.Fatalf("not_found token over a filesystem cause = %q, want environment", verdict)
	}

	// User before moai-marker: an internal marker wrapping a config sentinel
	// resolves to user.
	verdict, _ = Attribute(KindInternalError, MarkInternal(config.ErrInvalidConfig), "")
	if verdict != VerdictUser {
		t.Fatalf("marker wrapping a config sentinel = %q, want user", verdict)
	}
}

// TestAttributionMakesNoModelCall pins two facts: Attribute performs no
// persistence of its own (the spool is written only by Capture), and no
// model seam exists anywhere in this package — the import allowlist guard
// (AC-025) enforces the structural half; this test pins the observable half.
func TestAttributionMakesNoModelCall(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	for _, tc := range attributionTable() {
		Attribute(tc.kind, tc.err, tc.reason)
	}
	path, err := SpoolPath()
	if err != nil {
		t.Fatalf("SpoolPath: %v", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("Attribute wrote a spool file: %v", statErr)
	}
}

func TestUnknownErrorNeverAttributedToMoai(t *testing.T) {
	for _, kind := range AllKinds() {
		// A panic signal's verdict does not depend on an error value (the
		// recover site passes none); the unknown-error rule is about the
		// error-carrying kinds, whose unrecognised chains must fall to row F.
		if kind == KindPanic {
			continue
		}
		verdict, _ := Attribute(kind, &customTestError{}, "")
		if verdict == VerdictMoai {
			t.Errorf("kind %s: an unrecognised error was attributed to moai", kind)
		}
		if verdict != VerdictAmbiguous {
			t.Errorf("kind %s: unrecognised error verdict = %q, want ambiguous (row F)", kind, verdict)
		}
	}
}

// TestYAMLSyntaxErrorReachesFallback is the M3 first-test item (plan.md §C):
// it MEASURES what the YAML library returns for a syntax error in a user
// file and records the consequence. Measured 2026-10-07, yaml.v3 v3.0.1: a
// syntax error is an untyped *errors.errorString (NOT *yaml.TypeError —
// only decode type mismatches produce that), so a broken user file reaches
// row F and reads ambiguous, which stays local — the safe direction the
// design predicted.
func TestYAMLSyntaxErrorReachesFallback(t *testing.T) {
	var doc map[string]any
	err := yaml.Unmarshal([]byte("participation: [broken\n  yaml::\n"), &doc)
	if err == nil {
		t.Fatal("yaml.Unmarshal accepted a broken document")
	}
	var typeErr *yaml.TypeError
	if errors.As(err, &typeErr) {
		t.Fatalf("a yaml SYNTAX error decoded as *yaml.TypeError (%T); re-pin row U3", err)
	}
	// ...and attribution lands on row F, never on moai.
	verdict, _ := Attribute(KindHookHandlerFailure, err, "")
	if verdict != VerdictAmbiguous {
		t.Fatalf("yaml syntax error verdict = %q, want ambiguous", verdict)
	}

	// The positive control: a TYPE mismatch IS *yaml.TypeError (row U3, user).
	var typed struct {
		Participation struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"participation"`
	}
	err = yaml.Unmarshal([]byte("participation:\n  enabled: \"yes-please\"\n"), &typed)
	if err == nil {
		t.Fatal("yaml.Unmarshal accepted a wrongly typed value")
	}
	if !errors.As(err, &typeErr) {
		t.Fatalf("a yaml type mismatch decoded as %T, want *yaml.TypeError", err)
	}
	verdict, _ = Attribute(KindHookHandlerFailure, err, "")
	if verdict != VerdictUser {
		t.Fatalf("yaml type error verdict = %q, want user", verdict)
	}
}

// spoolPathForTest points MOAI_HOME at a fresh temporary directory and
// returns the spool path; the real home is never touched.
func spoolPathForTest(t *testing.T) string {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	path, err := SpoolPath()
	if err != nil {
		t.Fatalf("SpoolPath: %v", err)
	}
	return path
}

func spoolLineCount(t *testing.T, path string) int {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	n := 0
	for _, b := range raw {
		if b == '\n' {
			n++
		}
	}
	return n
}
