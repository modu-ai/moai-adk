package cli

import (
	"context"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/jevcred"
)

// jevProject writes a project root whose workflow.yaml carries the supplied
// jev block, and redirects the credential package's home seam at a temp dir so
// no test reads or writes the developer's real ~/.moai.
func jevProject(t *testing.T, enabled bool) string {
	t.Helper()
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "workflow:\n    jev:\n        enabled: false\n"
	if enabled {
		body = "workflow:\n    jev:\n        enabled: true\n"
	}
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	orig := jevcred.HomeDirFn
	home := t.TempDir()
	jevcred.HomeDirFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { jevcred.HomeDirFn = orig })
	t.Setenv("MOAI_HOME", "")
	t.Setenv(jevcred.EnvTestTypeSafeKey, "")

	return root
}

// stubProbe replaces the reachability seam and counts its calls, so the "sends
// no judgment request" assertion reads a number rather than an absence.
func stubProbe(t *testing.T, err error) *int {
	t.Helper()
	calls := 0
	orig := jevEndpointProbe
	jevEndpointProbe = func(context.Context) error {
		calls++
		return err
	}
	t.Cleanup(func() { jevEndpointProbe = orig })
	return &calls
}

// AC-JEVC-016 / REQ-JEVC-022 — the check reports enabled-state, credential
// presence, and endpoint reachability, and sends no judgment request.
func TestCheckJev_EnabledWithCredentialAndReachableEndpoint(t *testing.T) {
	root := jevProject(t, true)
	calls := stubProbe(t, nil)
	if err := jevcred.Save("NOT-A-REAL-KEY-0123456789wxyz"); err != nil {
		t.Fatal(err)
	}

	got := checkJev(root, false)
	if got.Name != jevCheckName {
		t.Errorf("Name = %q, want %q", got.Name, jevCheckName)
	}
	if got.Status != uikit.CheckOK {
		t.Errorf("Status = %q, want OK", got.Status)
	}
	text := got.Message + " " + got.Detail
	for _, want := range []string{"enabled", "credential", "reachable"} {
		if !strings.Contains(strings.ToLower(text), want) {
			t.Errorf("report does not mention %q: %q", want, text)
		}
	}
	if *calls != 1 {
		t.Errorf("reachability probe calls = %d, want 1", *calls)
	}
	// The bounded disclosure holds here too: the report shows the final four
	// characters and never the credential.
	if strings.Contains(text, "NOT-A-REAL-KEY-0123456789wxyz") {
		t.Error("the doctor report disclosed the whole credential")
	}
	if !strings.Contains(text, "wxyz") {
		t.Errorf("the report does not carry the four-character hint: %q", text)
	}
}

func TestCheckJev_DisabledProbesNothing(t *testing.T) {
	root := jevProject(t, false)
	calls := stubProbe(t, errors.New("probe must not run"))

	got := checkJev(root, false)
	if got.Status != uikit.CheckOK {
		t.Errorf("Status = %q, want OK — a disabled capability is a normal state, not a warning", got.Status)
	}
	if !strings.Contains(strings.ToLower(got.Message), "disabled") {
		t.Errorf("Message = %q, want it to report the disabled state", got.Message)
	}
	if *calls != 0 {
		t.Errorf("reachability probe calls = %d, want 0 — REQ-JEVC-017 forbids a network call while disabled", *calls)
	}
}

// A disabled capability with a credential present must still probe nothing:
// the gate is read before anything else.
func TestCheckJev_DisabledWithCredentialStillProbesNothing(t *testing.T) {
	root := jevProject(t, false)
	calls := stubProbe(t, nil)
	if err := jevcred.Save("NOT-A-REAL-KEY-present"); err != nil {
		t.Fatal(err)
	}
	got := checkJev(root, false)
	if *calls != 0 {
		t.Errorf("reachability probe calls = %d, want 0", *calls)
	}
	if got.Status != uikit.CheckOK {
		t.Errorf("Status = %q, want OK", got.Status)
	}
}

func TestCheckJev_EnabledWithoutCredentialIsAdvisoryNotAFailure(t *testing.T) {
	root := jevProject(t, true)
	stubProbe(t, nil)

	got := checkJev(root, false)
	if got.Status == uikit.CheckFail {
		t.Error("Status = Fail — an absent credential is graceful degradation, never a failure (REQ-JEVC-007)")
	}
	if !strings.Contains(strings.ToLower(got.Message+got.Detail), "credential") {
		t.Errorf("report does not name the missing credential: %q / %q", got.Message, got.Detail)
	}
}

func TestCheckJev_UnreachableEndpointIsAdvisoryNotAFailure(t *testing.T) {
	root := jevProject(t, true)
	stubProbe(t, errors.New("dial tcp: no route to host"))
	if err := jevcred.Save("NOT-A-REAL-KEY-present-here"); err != nil {
		t.Fatal(err)
	}

	got := checkJev(root, false)
	if got.Status == uikit.CheckFail {
		t.Error("Status = Fail — an unreachable endpoint is graceful degradation, never a failure")
	}
	if !strings.Contains(strings.ToLower(got.Message+got.Detail), "unreachable") {
		t.Errorf("report does not name the unreachable endpoint: %q / %q", got.Message, got.Detail)
	}
}

// An unreadable project config must not fail the check: doctor reports, it does
// not gate.
func TestCheckJev_UnreadableConfigDegradesToDisabled(t *testing.T) {
	calls := stubProbe(t, nil)
	got := checkJev(filepath.Join(t.TempDir(), "does-not-exist"), false)
	if got.Status == uikit.CheckFail {
		t.Errorf("Status = Fail on an unreadable project root; want a non-failing report")
	}
	if *calls != 0 {
		t.Errorf("reachability probe calls = %d, want 0 when the gate could not be read", *calls)
	}
}

// ---------------------------------------------------------------------------
// AC-JEVC-003 — the call path is unreachable from any decision of consequence.
// ---------------------------------------------------------------------------

// jevImporters returns the files under dir whose import set names internal/jev.
func jevImporters(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", dir, err)
	}
	fset := token.NewFileSet()
	var hits []string
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		scanned++
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, imp := range f.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if path == "github.com/modu-ai/moai-adk/internal/jev" {
				hits = append(hits, name)
			}
		}
	}
	if scanned == 0 {
		t.Fatalf("scanned 0 non-test Go files under %s — the scan establishes nothing", dir)
	}
	return hits
}

// The capability ships with ZERO consumers, and the guard states that as an
// exact set rather than as an absence: the doctor check is the one file allowed
// to import the client, and it doubles as the positive control proving the
// scanner fires.
func TestJevCallPath_HasExactlyTheDeclaredConsumers(t *testing.T) {
	const allowed = "doctor_jev.go"

	hits := jevImporters(t, ".")
	var unexpected []string
	sawAllowed := false
	for _, h := range hits {
		if h == allowed {
			sawAllowed = true
			continue
		}
		unexpected = append(unexpected, h)
	}
	if !sawAllowed {
		t.Fatalf("positive control failed: %s does not import internal/jev, so the zero-result below is unattributable", allowed)
	}
	if len(unexpected) > 0 {
		t.Errorf("internal/cli files outside the declared consumer set import internal/jev: %v — a completion verdict, a merge approval, a queue mutation, an operator gate, or a slot-wait adjudication MUST NOT reach the call path, not even as an input (REQ-JEVC-012)", unexpected)
	}
}

// The queue-mutation, verdict, and integration-window surfaces are named
// explicitly, so the guard still binds if the allow-list above is ever widened.
func TestJevCallPath_UnreachableFromDecisionSurfaces(t *testing.T) {
	surfaces := []string{
		"gtd.go", "todo_analysis.go", "todo_autodone.go",
		"integration.go", "integration_settings_drift.go",
	}
	fset := token.NewFileSet()
	checked := 0
	for _, name := range surfaces {
		if _, err := os.Stat(name); err != nil {
			continue // the file was renamed; the package-wide guard above still binds
		}
		checked++
		f, err := parser.ParseFile(fset, name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, imp := range f.Imports {
			if strings.Trim(imp.Path.Value, `"`) == "github.com/modu-ai/moai-adk/internal/jev" {
				t.Errorf("%s imports internal/jev — a queue mutation, verdict, or integration-window surface MUST NOT reach the call path (REQ-JEVC-012)", name)
			}
		}
	}
	if checked == 0 {
		t.Fatal("none of the named decision surfaces was found — the scan establishes nothing; update the list")
	}
}
