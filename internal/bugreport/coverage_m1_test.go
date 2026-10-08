package bugreport

import (
	"errors"
	"runtime"
	"strings"
	"testing"
)

// TestMarkerWrapsAndUnwraps covers the internal marker (design.md section 3,
// row M2): errors.As finds it through wrapping, a non-marker chain does not
// match, and nil is a no-op.
func TestMarkerWrapsAndUnwraps(t *testing.T) {
	if MarkInternal(nil) != nil {
		t.Fatal("MarkInternal(nil) is not nil")
	}

	cause := errors.New("store is not *fileStore")
	marked := MarkInternal(cause)
	if !IsInternalMarker(marked) {
		t.Fatal("IsInternalMarker(marked) = false")
	}
	if !IsInternalMarker(fmtWrap(fmtWrap(marked))) {
		t.Fatal("marker not found through a wrapped chain")
	}
	if IsInternalMarker(cause) {
		t.Fatal("IsInternalMarker(plain error) = true")
	}
	if got := marked.Error(); got == "" || !errors.Is(marked, cause) {
		t.Fatalf("marker loses its cause: %q / errors.Is=%v", got, errors.Is(marked, cause))
	}
}

type wrapErr struct{ inner error }

func (w wrapErr) Error() string { return "wrapped: " + w.inner.Error() }
func (w wrapErr) Unwrap() error { return w.inner }

func fmtWrap(e error) error { return wrapErr{inner: e} }

// TestVerdictEnumClosed pins the four-member verdict set.
func TestVerdictEnumClosed(t *testing.T) {
	for _, v := range []Verdict{VerdictMoai, VerdictUser, VerdictEnvironment, VerdictAmbiguous} {
		if !v.Valid() {
			t.Fatalf("verdict %q does not validate", v)
		}
	}
	for _, v := range []Verdict{Verdict(""), Verdict("m o a i"), Verdict("unknown")} {
		if v.Valid() {
			t.Fatalf("verdict %q validated", v)
		}
	}
}

// TestFramesFromPCCapturesThisStack covers the runtime.Callers entry: the
// frames it returns for this test's stack are bare symbol paths that pass the
// validator, and the bugreport test helpers themselves are dropped.
func TestFramesFromPCCapturesThisStack(t *testing.T) {
	var pc [32]uintptr
	n := runtime.Callers(1, pc[:])
	if n == 0 {
		t.Skip("runtime.Callers returned no program counters")
	}
	got := FramesFromPC(pc[:n])
	if len(got) == 0 {
		t.Log("no moai frames in this stack (test binary naming); nothing to assert")
		return
	}
	for _, name := range got {
		if err := ValidateFrameName(name); err != nil {
			t.Fatalf("frame %q from live stack fails the validator: %v", name, err)
		}
		if len(name) >= len(ModulePrefix) && name[:len(ModulePrefix)] == ModulePrefix {
			t.Fatalf("frame %q still carries the module prefix", name)
		}
	}
}

// TestBuildRejectsInvalidInputs walks Build's refusal branches so a payload
// can only exist with a fully valid input set.
func TestBuildRejectsInvalidInputs(t *testing.T) {
	goodFrames := []string{"internal/cli.Execute"}
	build := validBuild()

	if _, err := Build(Kind("nope"), goodFrames, nil, build); !errors.Is(err, ErrBuildRejected) {
		t.Errorf("unknown kind: err = %v", err)
	}
	if _, err := Build(KindHookHandlerFailure, nil, nil, build); !errors.Is(err, ErrBuildRejected) {
		t.Errorf("no frames: err = %v", err)
	}
	many := make([]string, 0, FrameLimit()+1)
	for i := 0; i <= FrameLimit(); i++ {
		many = append(many, "internal/deep/pkg.Func")
	}
	if _, err := Build(KindPanic, many, nil, build); !errors.Is(err, ErrBuildRejected) {
		t.Errorf("over-limit frames: err = %v", err)
	}
	if _, err := Build(KindPanic, []string{"/Users/leak/x"}, nil, build); !errors.Is(err, ErrBuildRejected) {
		t.Errorf("path-shaped frame: err = %v", err)
	}
	// A detail carrier the kind's row does not admit is refused...
	if _, err := Build(KindPanic, goodFrames, TokenMissingKey, build); !errors.Is(err, ErrBuildRejected) {
		t.Errorf("token on panic kind: err = %v", err)
	}
	// ...and so is a HookDetail whose names were never registered.
	if _, err := Build(KindHookHandlerFailure, goodFrames, HookDetail{EventID: "UnregisteredEvt", HandlerID: "unregisteredHandler"}, build); !errors.Is(err, ErrBuildRejected) {
		t.Errorf("unregistered hook detail: err = %v", err)
	}
	// A dev-build identity ("none" commit) fails the commit allowlist: a dev
	// build must not be able to produce a publishable payload.
	if _, err := Build(KindPanic, goodFrames, nil, BuildIdentity{Version: "v3.2.0", Commit: "none"}); !errors.Is(err, ErrBuildRejected) {
		t.Errorf("dev commit: err = %v", err)
	}
}

// TestBuildSerialisesDetail covers the MarshalJSON detail branch: a payload
// built with a registered hook detail carries its token on the wire and
// ParsePayload restores the same carrier.
func TestBuildSerialisesDetail(t *testing.T) {
	const evt = "TestEvtSerialise"
	const hnd = "testHandlerSerialise"
	if err := RegisterEvent(evt); err != nil {
		t.Fatalf("RegisterEvent: %v", err)
	}
	if err := RegisterHandlerName(hnd); err != nil {
		t.Fatalf("RegisterHandlerName: %v", err)
	}
	hd, err := RegisterHookIdentity(evt, hnd)
	if err != nil {
		t.Fatalf("RegisterHookIdentity: %v", err)
	}

	p, err := Build(KindHookHandlerFailure, []string{"internal/cli.Execute"}, hd, validBuild())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	raw, err := p.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if !contains(string(raw), hd.Token()) {
		t.Fatalf("wire form %s does not carry the hook token %q", raw, hd.Token())
	}
	back, err := ParsePayload(raw)
	if err != nil {
		t.Fatalf("ParsePayload: %v", err)
	}
	got, ok := back.Detail.(HookDetail)
	if !ok || got != hd {
		t.Fatalf("round-tripped detail = %#v, want %#v", back.Detail, hd)
	}

	// The registration accessors report the tables.
	if !EventRegistered(evt) || !HandlerRegistered(hnd) {
		t.Fatal("registered names not reported")
	}
	if EventRegistered("NotRegisteredEvt") || HandlerRegistered("notRegisteredHandler") {
		t.Fatal("unregistered names reported as registered")
	}
	if RegisteredEventCount() == 0 || RegisteredHandlerCount() == 0 {
		t.Fatal("registry counters read zero with entries present")
	}
}

// TestParsePayloadRejectsGarbage covers the decode-failure branch.
func TestParsePayloadRejectsGarbage(t *testing.T) {
	if _, err := ParsePayload([]byte("not json")); !errors.Is(err, ErrPayloadRejected) {
		t.Fatalf("garbage: err = %v", err)
	}
	p := buildValidPayload(t)
	raw, err := p.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// A valid payload whose detail does not belong to the kind is rejected.
	doctored := replaceOnce(string(raw), `"kind":"panic"`, `"kind":"panic","detail":"missing_key"`)
	if doctored == string(raw) {
		t.Fatal("doctoring the payload did not change it")
	}
	if _, err := ParsePayload([]byte(doctored)); !errors.Is(err, ErrDetailRejected) {
		t.Fatalf("token detail on panic kind: err = %v", err)
	}
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }

func replaceOnce(s, old, new string) string { return strings.Replace(s, old, new, 1) }
