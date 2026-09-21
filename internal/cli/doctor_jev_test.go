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

// The consumer set is stated as an EXACT set rather than as an absence: every
// file allowed to import the client is named, and the doctor check doubles as
// the positive control proving the scanner fires.
//
// SPEC-JEV-CORE-001 shipped this set with exactly one member, because that SPEC
// ships the capability with zero consumers and says so in its own scope
// (`spec.md` §D: "near-duplicate marking, lane-question routing, and skill
// suggestion belong to SPEC-JEV-CONSUMERS-001"). Extending the set is therefore
// the declared way a consumer arrives, not a weakening of the guard — what the
// guard forbids is an UNDECLARED importer.
//
// `todo_jev_finding.go` is the first declared consumer (SPEC-JEV-CONSUMERS-001
// M4, REQ-JEVN-001: Consumer C, near-duplicate card marking). Its own file
// comment carries the scope it operates under, and the decision-surface guard
// below still binds it.
func TestJevCallPath_HasExactlyTheDeclaredConsumers(t *testing.T) {
	const control = "doctor_jev.go"
	allowed := map[string]string{
		control:                 "SPEC-JEV-CORE-001 — the doctor check",
		"todo_jev_finding.go":   "SPEC-JEV-CONSUMERS-001 M4 — Consumer C, near-duplicate marking",
		"jev_skill_suggest.go":  "SPEC-JEV-CONSUMERS-001 M6 — Consumer B, skill suggestion (gate-unrun)",
	}

	hits := jevImporters(t, ".")
	var unexpected []string
	sawControl := false
	for _, h := range hits {
		if h == control {
			sawControl = true
		}
		if _, ok := allowed[h]; ok {
			continue
		}
		unexpected = append(unexpected, h)
	}
	if !sawControl {
		t.Fatalf("positive control failed: %s does not import internal/jev, so the zero-result below is unattributable", control)
	}
	if len(unexpected) > 0 {
		t.Errorf("internal/cli files outside the declared consumer set import internal/jev: %v — a completion verdict, a merge approval, a queue mutation, an operator gate, or a slot-wait adjudication MUST NOT reach the call path, not even as an input (REQ-JEVC-012)", unexpected)
	}
}

// The queue-mutation, verdict, and integration-window surfaces are named
// explicitly, so the guard still binds if the allow-list above is ever widened.
//
// The scan measures TWO things, and the second was added by
// SPEC-JEV-CONSUMERS-001 M4 because the first alone had become evadable: these
// surfaces live in the SAME PACKAGE as the consumer, so a surface can reach the
// call path through a plain function call while importing nothing. An
// import-only guard would have gone green on exactly that arrangement and read
// as "unreachable" while the reference was one identifier away. The symbol scan
// closes it.
//
// ONE exception is declared, named, and cited — SPEC-JEV-CONSUMERS-001
// REQ-JEVN-001 requires Consumer C to record its finding at card ADMISSION,
// which is `appendAnalyzedCard` in todo_analysis.go. Nothing else on any named
// surface may reference the consumer, and no surface may import the client.
//
// The exception is a recorded tension, not a resolved one: REQ-JEVC-011 and
// REQ-JEVC-012 of SPEC-JEV-CORE-001 forbid consulting Jev for a `moai todo`
// mutation and forbid a Jev answer mutating the backlog queue, and
// SPEC-JEV-CONSUMERS-001 authorises precisely a finding append during
// `todo add` without reconciling that wording. Consumer C is gated off by
// default and its shipping gate (REQ-JEVO-009) has not been run, so nothing
// currently reaches the call path in a shipped build; the wording
// reconciliation belongs to the SPEC layer.
func TestJevCallPath_UnreachableFromDecisionSurfaces(t *testing.T) {
	surfaces := []string{
		"gtd.go", "todo_analysis.go", "todo_autodone.go",
		"integration.go", "integration_settings_drift.go",
	}
	// Same-package identifiers through which a surface could reach the client
	// without importing it.
	consumerSymbols := []string{
		"jevNearDuplicateProbe",
		"liveJevNearDuplicateProbe",
		"appendJevNearDuplicateFinding",
		"jevFindingSignalFragment",
	}
	// file -> symbol -> the SPEC clause authorising the reference.
	authorised := map[string]map[string]string{
		"todo_analysis.go": {
			"appendJevNearDuplicateFinding": "SPEC-JEV-CONSUMERS-001 REQ-JEVN-001 (admission-path record)",
			"jevFindingSignalFragment":      "SPEC-JEV-CONSUMERS-001 REQ-JEVN-005 (render form)",
		},
	}

	fset := token.NewFileSet()
	checked := 0
	sawAuthorised := false
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
		body, err := os.ReadFile(name) // #nosec G304 -- fixed in-repository source path
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, sym := range consumerSymbols {
			if !strings.Contains(string(body), sym) {
				continue
			}
			if _, ok := authorised[name][sym]; ok {
				sawAuthorised = true
				continue
			}
			t.Errorf("%s references %s — a decision surface reaches the Jev consumer through a "+
				"same-package call, which an import-only scan would have missed (REQ-JEVC-012)", name, sym)
		}
	}
	if checked == 0 {
		t.Fatal("none of the named decision surfaces was found — the scan establishes nothing; update the list")
	}
	if !sawAuthorised {
		t.Fatal("positive control failed: no authorised consumer reference was found on any named surface, " +
			"so the symbol scan above matched nothing and its silence asserts nothing")
	}
}
