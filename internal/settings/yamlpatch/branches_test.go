package yamlpatch

// branches_test.go — card t1619 (SPEC-UPDATE-REPAIR-RC30-001): the branches of
// patchFile and its helpers that the card changed and that no existing test
// reached (yamlpatch.go lines 175, 179, 183, 187, 436, and 450-452). Each test
// names the line it exercises.

import (
	"os"
	"strings"
	"testing"
)

// yamlpatchFixture is a small llm section used by the branch tests.
const yamlpatchFixture = "llm:\n  harness: claude\n"

// TestPatchFileRefusesPathThroughScalar covers the applyEdit refusal on the
// re-encode path of patchFile: the path descends through a scalar, which is not
// a mapping, so the edit is refused and the file is unchanged. (The same wrap
// inside appendSection is line 175; see TestPatchFileWithModeSectionEditErrorIsReturned.)
func TestPatchFileRefusesPathThroughScalar(t *testing.T) {
	path := writeTempYAML(t, yamlpatchFixture)
	err := PatchFile(path, []KeyEdit{{Path: []string{"llm", "harness", "deeper"}, Value: "x"}})
	if err == nil || !strings.Contains(err.Error(), "parent is not a mapping") {
		t.Fatalf("PatchFile err = %v, want a refusal naming the non-mapping parent", err)
	}
	if got := readFileForBranchTest(t, path); got != yamlpatchFixture {
		t.Errorf("file changed after a refused edit:\n%s", got)
	}
}

// TestPatchFileRemoveAbsentKeyBesideUpsert covers the removal of an absent key
// on the re-encode path (yamlpatch.go line 436). An upsert in the same batch
// forces the re-encode path; the absent key is a no-op there, and the upsert
// still lands.
func TestPatchFileRemoveAbsentKeyBesideUpsert(t *testing.T) {
	path := writeTempYAML(t, yamlpatchFixture)
	err := PatchFile(path, []KeyEdit{
		{Path: []string{"llm", "absent"}, Remove: true},
		{Path: []string{"llm", "new_key"}, Value: "x"},
	})
	if err != nil {
		t.Fatalf("PatchFile: %v", err)
	}
	got := readFileForBranchTest(t, path)
	if !strings.Contains(got, "harness: claude") || !strings.Contains(got, "new_key: x") {
		t.Errorf("result lost a key or missed the upsert:\n%s", got)
	}
}

// TestPatchFileQuotedScalarOnReencodePath covers the quoted scalar write
// (yamlpatch.go lines 450-452): a scalar edit marked Quoted, taken on the
// re-encode path because a companion upsert in the batch rules out the line
// splice, is written double-quoted.
func TestPatchFileQuotedScalarOnReencodePath(t *testing.T) {
	path := writeTempYAML(t, yamlpatchFixture)
	err := PatchFile(path, []KeyEdit{
		{Path: []string{"llm", "harness"}, Value: "codex", Quoted: true},
		{Path: []string{"llm", "new_key"}, Value: "x"},
	})
	if err != nil {
		t.Fatalf("PatchFile: %v", err)
	}
	if got := readFileForBranchTest(t, path); !strings.Contains(got, `harness: "codex"`) {
		t.Errorf("harness was not written double-quoted:\n%s", got)
	}
}

// TestPatchFileWithModeCommentOnlyFileGetsSection covers the section append for
// a comment-only file (yamlpatch.go line 187): a file whose last byte is not a
// newline gets one before the section, so the section starts on its own line.
func TestPatchFileWithModeCommentOnlyFileGetsSection(t *testing.T) {
	path := writeTempYAML(t, "# header")
	if err := PatchFileWithMode(path, []KeyEdit{{Path: []string{"llm", "harness"}, Value: "codex"}}, 0o644); err != nil {
		t.Fatalf("PatchFileWithMode: %v", err)
	}
	got := readFileForBranchTest(t, path)
	if !strings.HasPrefix(got, "# header\n") || !strings.Contains(got, "harness: codex") {
		t.Errorf("section not appended on its own line after the comment:\n%s", got)
	}
}

// TestPatchFileWithModeEmptySectionLeavesFile covers the no-content return of
// appendSection (yamlpatch.go line 179): an edit that removes an absent key
// builds an empty section, so the comment-only file is left as it was.
func TestPatchFileWithModeEmptySectionLeavesFile(t *testing.T) {
	const body = "# header\n"
	path := writeTempYAML(t, body)
	if err := PatchFileWithMode(path, []KeyEdit{{Path: []string{"llm", "absent"}, Remove: true}}, 0o644); err != nil {
		t.Fatalf("PatchFileWithMode: %v", err)
	}
	if got := readFileForBranchTest(t, path); got != body {
		t.Errorf("a comment-only file changed for an empty section:\n%s", got)
	}
}

// TestPatchFileRejectsInvalidUTF8Value covers the encode refusal on the
// re-encode path of patchFile (the same refusal inside appendSection is line
// 183; see TestPatchFileWithModeSectionEncodeRefusal): a quoted upsert whose value is not valid UTF-8
// pins the !!str tag, and the encoder refuses an explicitly tagged string that
// is not UTF-8 (yaml.v3 encode.go, "cannot marshal invalid UTF-8 data as !!str").
// The file is left unchanged. An untagged value of the same bytes is not
// refused: yaml.v3 writes it as !!binary, which is why Quoted is set here.
func TestPatchFileRejectsInvalidUTF8Value(t *testing.T) {
	path := writeTempYAML(t, yamlpatchFixture)
	err := PatchFile(path, []KeyEdit{
		{Path: []string{"llm", "note"}, Value: "\xff", Quoted: true},
	})
	if err == nil || !strings.Contains(err.Error(), "yamlpatch: encode") {
		t.Fatalf("PatchFile err = %v, want an encode refusal", err)
	}
	if got := readFileForBranchTest(t, path); got != yamlpatchFixture {
		t.Errorf("file changed after an encode refusal:\n%s", got)
	}
}

// TestPatchFileWithModeSectionEditErrorIsReturned covers the edit refusal inside
// appendSection (yamlpatch.go line 175). A comment-only file gets its section
// from the edits alone, so a second edit that descends through the scalar the
// first one created is refused; the file is left unchanged.
func TestPatchFileWithModeSectionEditErrorIsReturned(t *testing.T) {
	const body = "# header\n"
	path := writeTempYAML(t, body)
	err := PatchFileWithMode(path, []KeyEdit{
		{Path: []string{"llm"}, Value: "x"},
		{Path: []string{"llm", "harness"}, Value: "codex"},
	}, 0o644)
	if err == nil || !strings.Contains(err.Error(), "parent is not a mapping") {
		t.Fatalf("PatchFileWithMode err = %v, want a refusal naming the non-mapping parent", err)
	}
	if got := readFileForBranchTest(t, path); got != body {
		t.Errorf("a comment-only file changed after a refused section:\n%s", got)
	}
}

// TestPatchFileWithModeSectionEncodeRefusal covers the encode refusal inside
// appendSection (yamlpatch.go line 183). A comment-only file takes the section
// path, and a quoted value that is not valid UTF-8 is refused by the encoder.
func TestPatchFileWithModeSectionEncodeRefusal(t *testing.T) {
	const body = "# header\n"
	path := writeTempYAML(t, body)
	err := PatchFileWithMode(path, []KeyEdit{
		{Path: []string{"llm", "note"}, Value: "\xff", Quoted: true},
	}, 0o644)
	if err == nil || !strings.Contains(err.Error(), "yamlpatch: encode") {
		t.Fatalf("PatchFileWithMode err = %v, want an encode refusal", err)
	}
	if got := readFileForBranchTest(t, path); got != body {
		t.Errorf("a comment-only file changed after an encode refusal:\n%s", got)
	}
}

// readFileForBranchTest returns the file body or fails the test.
func readFileForBranchTest(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
