package contract

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func newSpecDir(t *testing.T) (root, dir string) {
	t.Helper()
	root = t.TempDir()
	dir = filepath.Join(root, ".moai", "specs", fixtureSpecID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return root, dir
}

func symlinkOrSkip(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable on this platform: %v", err)
	}
}

func TestResolveSpecDir_RejectsNonCanonicalIDs(t *testing.T) {
	for _, id := range []string{"", "../SPEC-X-001", "SPEC-X-001/../../etc", "spec-x-001", "SPEC-X-1"} {
		if _, err := ResolveSpecDir(t.TempDir(), id); !errors.Is(err, ErrInvalidSpecID) {
			t.Errorf("ResolveSpecDir(%q) error = %v, want ErrInvalidSpecID", id, err)
		}
	}
}

func TestLoadDir_ReadsContractAcceptanceAndReceipt(t *testing.T) {
	_, dir := newSpecDir(t)
	contractBytes := signFixture(renderFixture(fixtureOpts{}))
	writeFile(t, filepath.Join(dir, ContractFile), contractBytes)
	writeFile(t, filepath.Join(dir, AcceptanceFile), []byte(fixtureAcceptance))
	writeFile(t, filepath.Join(dir, ReceiptFile), []byte(`{"receipt_version":1}`))

	in, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if in.SpecID != fixtureSpecID || string(in.Contract) != string(contractBytes) {
		t.Errorf("SpecID=%q contract bytes equal=%v", in.SpecID, string(in.Contract) == string(contractBytes))
	}
	if !in.AcceptancePresent || string(in.Acceptance) != fixtureAcceptance {
		t.Errorf("acceptance present=%v", in.AcceptancePresent)
	}
	if !in.ReceiptPresent || string(in.Receipt) != `{"receipt_version":1}` {
		t.Errorf("receipt present=%v bytes=%q", in.ReceiptPresent, in.Receipt)
	}
}

func TestLoadDir_OptionalFilesAbsent(t *testing.T) {
	_, dir := newSpecDir(t)
	writeFile(t, filepath.Join(dir, ContractFile), []byte(renderFixture(fixtureOpts{})))
	in, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if in.AcceptancePresent || in.Acceptance != nil || in.ReceiptPresent || in.Receipt != nil {
		t.Errorf("absent files reported present: %+v", in)
	}
}

func TestLoadDir_RelativePath(t *testing.T) {
	_, dir := newSpecDir(t)
	writeFile(t, filepath.Join(dir, ContractFile), []byte(renderFixture(fixtureOpts{})))
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(wd, dir)
	if err != nil {
		t.Skipf("no relative path from %s to %s: %v", wd, dir, err)
	}
	in, err := LoadDir(rel)
	if err != nil {
		t.Fatalf("LoadDir(%q): %v", rel, err)
	}
	if in.SpecID != fixtureSpecID {
		t.Errorf("SpecID = %q", in.SpecID)
	}
}

func TestLoadDir_ContractMissing(t *testing.T) {
	_, dir := newSpecDir(t)
	if _, err := LoadDir(dir); !errors.Is(err, ErrContractMissing) {
		t.Errorf("error = %v, want ErrContractMissing", err)
	}
}

func TestLoadDir_SpecDirMissing(t *testing.T) {
	_, err := LoadDir(filepath.Join(t.TempDir(), "SPEC-NOPE-001"))
	if err == nil {
		t.Fatalf("LoadDir on a missing directory returned no error")
	}
}

func TestLoadDir_ContractNotARegularFile(t *testing.T) {
	_, dir := newSpecDir(t)
	if err := os.Mkdir(filepath.Join(dir, ContractFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadDir(dir); err == nil {
		t.Errorf("a directory named contract.yaml was accepted")
	}
}

func TestLoadDir_SymlinkEscapingSpecDir(t *testing.T) {
	root, dir := newSpecDir(t)
	outside := filepath.Join(root, "outside.yaml")
	writeFile(t, outside, []byte(renderFixture(fixtureOpts{})))
	symlinkOrSkip(t, outside, filepath.Join(dir, ContractFile))
	if _, err := LoadDir(dir); !errors.Is(err, ErrPathEscapesSpecDir) {
		t.Errorf("error = %v, want ErrPathEscapesSpecDir", err)
	}
}

func TestLoadDir_AcceptanceSymlinkEscapingSpecDir(t *testing.T) {
	root, dir := newSpecDir(t)
	writeFile(t, filepath.Join(dir, ContractFile), []byte(renderFixture(fixtureOpts{})))
	outside := filepath.Join(root, "acceptance.md")
	writeFile(t, outside, []byte(fixtureAcceptance))
	symlinkOrSkip(t, outside, filepath.Join(dir, AcceptanceFile))
	if _, err := LoadDir(dir); !errors.Is(err, ErrPathEscapesSpecDir) {
		t.Errorf("error = %v, want ErrPathEscapesSpecDir", err)
	}
}

func TestLoadDir_SymlinkInsideSpecDir(t *testing.T) {
	_, dir := newSpecDir(t)
	target := filepath.Join(dir, "contract.real.yaml")
	writeFile(t, target, []byte(renderFixture(fixtureOpts{})))
	symlinkOrSkip(t, "contract.real.yaml", filepath.Join(dir, ContractFile))
	in, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if len(in.Contract) == 0 {
		t.Errorf("contract bytes empty")
	}
}

func TestLoadDir_SymlinkedSpecDir(t *testing.T) {
	// The SPEC directory itself reached through a symlink: containment is
	// judged against the resolved directory, so its own files still load.
	root, dir := newSpecDir(t)
	writeFile(t, filepath.Join(dir, ContractFile), []byte(renderFixture(fixtureOpts{})))
	link := filepath.Join(root, fixtureSpecID)
	symlinkOrSkip(t, dir, link)
	in, err := LoadDir(link)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if in.SpecID != fixtureSpecID {
		t.Errorf("SpecID = %q", in.SpecID)
	}
}

func TestLoadDir_DanglingSymlink(t *testing.T) {
	root, dir := newSpecDir(t)
	symlinkOrSkip(t, filepath.Join(root, "gone.yaml"), filepath.Join(dir, ContractFile))
	if _, err := LoadDir(dir); err == nil {
		t.Errorf("a dangling contract symlink was accepted")
	}
}
