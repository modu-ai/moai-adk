package publish

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/feedback"
	"github.com/modu-ai/moai-adk/internal/feedback/outbox"
)

// fixedPayload is a platform-independent payload literal: the golden and
// the round-trip tests pin BYTE-STABLE renderings, so the payload is not
// built from the running platform's runtime.GOOS/GOARCH (Build() would
// make the bytes differ per CI runner).
func fixedPayload() bugreport.Payload {
	return bugreport.Payload{
		Schema:      "v1",
		Kind:        bugreport.KindPanic,
		Fingerprint: "0123456789abcdef",
		Version:     "v3.2.0",
		Commit:      "abcdef1234567",
		OS:          "linux",
		Arch:        "amd64",
		Frames:      []string{"internal/cli.Execute", "internal/cli/root.Execute"},
	}
}

func fixedItem(t *testing.T) feedback.QueueItem {
	t.Helper()
	p := fixedPayload()
	title, body := outbox.RenderReport(p)
	return feedback.QueueItem{ID: "f1", Title: title, Body: body, Fingerprint: p.Fingerprint, Kind: string(p.Kind)}
}

// OccurrenceMarkerForTest renders one occurrence comment body for a
// payload — the sender's own render path, invoked from the tests.
func OccurrenceMarkerForTest(p bugreport.Payload) string {
	title, body := outbox.RenderReport(p)
	return OccurrenceComment(feedback.QueueItem{Title: title, Body: body})
}

// ---- AC-020: the issue contract ----

func TestIssueContractRoundTrip(t *testing.T) {
	p := fixedPayload()
	title, body := outbox.RenderReport(p)
	item := fixedItem(t)
	occurrence := OccurrenceComment(item)

	// The title key parses back to kind + fingerprint.
	kind, fp, ok := ParseTitleKey(title)
	if !ok || kind != string(p.Kind) || fp != p.Fingerprint {
		t.Fatalf("ParseTitleKey(%q) = %q/%q/%v, want %s/%s/true", title, kind, fp, ok, p.Kind, p.Fingerprint)
	}

	// The issue body's marker block recovers every field.
	fields, ok := ParseMarker(body)
	if !ok {
		t.Fatalf("the issue body carries no marker block: %q", body)
	}
	want := map[string]string{
		"schema":      "v1",
		"fingerprint": p.Fingerprint,
		"kind":        string(p.Kind),
		"version":     p.Version,
		"commit":      p.Commit,
		"os_arch":     p.OS + "/" + p.Arch,
		"frames":      strings.Join(p.Frames, ","),
	}
	for k, v := range want {
		if fields[k] != v {
			t.Fatalf("marker field %s = %q, want %q", k, fields[k], v)
		}
	}

	// The occurrence comment's marker recovers the same fields.
	occFields, ok := ParseMarker(occurrence)
	if !ok {
		t.Fatalf("the occurrence comment carries no marker block: %q", occurrence)
	}
	for k, v := range want {
		if occFields[k] != v {
			t.Fatalf("occurrence marker field %s = %q, want %q", k, occFields[k], v)
		}
	}

	// The rendering equals the golden bytes (E23): the body the create path
	// hands to gh.
	goldenPath := filepath.Join("..", "testdata", "bugreport_issue_v1.golden")
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if string(golden) != body {
		t.Fatalf("rendering drifted from the golden:\n--- golden ---\n%s\n--- rendered ---\n%s", golden, body)
	}
}

func TestSameFingerprintSameTitleKey(t *testing.T) {
	p1 := fixedPayload()
	p2 := p1 // same fingerprint, rendered twice
	title1, _ := outbox.RenderReport(p1)
	title2, _ := outbox.RenderReport(p2)
	if title1 != title2 {
		t.Fatalf("the same fingerprint rendered two title keys:\n%s\n%s", title1, title2)
	}
	// And a different fingerprint renders a different key.
	p3 := p1
	p3.Fingerprint = "fedcba9876543210"
	title3, _ := outbox.RenderReport(p3)
	if title3 == title1 {
		t.Fatal("a different fingerprint rendered the same title key")
	}
}

func TestNoLabelsNoBodyEdit(t *testing.T) {
	// Structural: the gh runner's argv carries no --label and no edit
	// subcommand — the seam offers create/search/comment only, and the
	// source of the production runner must never grow either shape. The
	// source's comments are stripped first: the guard pins ARGV SHAPES,
	// not prose about them.
	raw, err := os.ReadFile("ghrunner.go")
	if err != nil {
		t.Fatalf("read ghrunner.go: %v", err)
	}
	var code strings.Builder
	for _, line := range strings.Split(string(raw), "\n") {
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		code.WriteString(line)
		code.WriteString("\n")
	}
	src := code.String()
	if strings.Contains(src, "--label") {
		t.Fatal("ghrunner.go passes --label: labels depend on repository permission and are dropped or rejected for non-collaborators")
	}
	if strings.Contains(src, `"edit"`) {
		t.Fatal("ghrunner.go invokes an issue edit: the issue body is immutable by contract (REQ-ANON-019)")
	}
}

func TestOccurrenceMarkersAreUntrustedInput(t *testing.T) {
	p := fixedPayload()
	title, body := outbox.RenderReport(p)
	item := fixedItem(t)

	// A forged comment whose marker fields disagree with the issue's title
	// key and body block.
	forged := "<!-- moai-bugreport:v1 occurrence schema=v1 fingerprint=deadbeefdeadbeef kind=hook_timeout version=9.9.9 commit=fffffffffffffff os_arch=windows/arm64 frames=evil.Frame -->"
	if !IsOccurrenceComment(forged) {
		t.Fatal("a forged marker-shaped comment was not recognized — the advisory count must include it")
	}
	count := OccurrenceCount([]RemoteComment{{Body: "a plain comment"}, {Body: forged}, {Body: OccurrenceComment(item)}})
	if count != 2 {
		t.Fatalf("OccurrenceCount = %d, want 2 (the forged one counts toward the advisory total)", count)
	}

	// The contract fields come from the title key and the issue body block
	// only — every disagreeing marker field is ignored.
	fields, ok := ContractFields(title, body)
	if !ok {
		t.Fatal("ContractFields could not derive the issue's fields")
	}
	if fields["fingerprint"] != p.Fingerprint {
		t.Fatalf("contract fingerprint = %q, want the title key's %s", fields["fingerprint"], p.Fingerprint)
	}
	if fields["version"] != p.Version {
		t.Fatalf("contract version = %q, want the body block's %s", fields["version"], p.Version)
	}
	if fields["os_arch"] != p.OS+"/"+p.Arch {
		t.Fatalf("contract os_arch = %q, want the body block's %s/%s", fields["os_arch"], p.OS, p.Arch)
	}
	if fields["kind"] != string(p.Kind) {
		t.Fatalf("contract kind = %q, want %s", fields["kind"], p.Kind)
	}
}

func TestPreviewMatchesCreateBytes(t *testing.T) {
	consentOn(t)
	_, item := payloadFixture(t)
	seedQueue(t, item)

	stub := newStubRunner(true) // no existing issue: the create path runs
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	_, creates, _ := stub.recorded()
	if creates != 1 {
		t.Fatalf("creates = %d, want the one create", creates)
	}

	// The bytes the create path hands to gh equal the body the preview
	// prints for the same queued item (REQ-ANON-014's identity, closed at
	// AC-020): the preview prints "<title>\n<body>\n\n" per item and the
	// create hands "<body>" to --body-file.
	previewBody := item.Body
	if stub.creates[0].Body != previewBody {
		t.Fatalf("create body drifted from the preview bytes:\n--- create ---\n%s\n--- preview ---\n%s", stub.creates[0].Body, previewBody)
	}
	if stub.creates[0].Title != item.Title {
		t.Fatalf("create title %q != queued title %q", stub.creates[0].Title, item.Title)
	}
}
