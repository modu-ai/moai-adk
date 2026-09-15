package mission

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCompletionReceiptBindsLineageEvidenceAndAncestry(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".moai", "state", "mission", "governance", "completion.json")
	now := time.Now()
	base := CompletionReceipt{Version: 1, MissionID: "m", ContractHash: "c", SnapshotHash: "s", HeadSHA: "0123456789012345678901234567890123456789", Issuer: "completion-auditor", Status: "passed", Evidence: map[string]bool{"tests": true, "landed": true}, LandedAncestry: true, ExpiresAt: now.Add(time.Hour)}
	exp := CompletionExpectation{MissionID: "m", ContractHash: "c", SnapshotHash: "s", HeadSHA: base.HeadSHA, RequiredEvidence: []string{"tests", "landed"}, RequireLandedAncestry: true, Now: now}
	if err := WriteCompletionReceipt(root, path, base); err != nil {
		t.Fatal(err)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Fatalf("mode=%o", info.Mode().Perm())
	}
	if _, err := LoadCompletionReceipt(root, path, exp); err != nil {
		t.Fatal(err)
	}
	mutants := []func(*CompletionReceipt){
		func(r *CompletionReceipt) { r.MissionID = "other" }, func(r *CompletionReceipt) { r.Evidence["tests"] = false },
		func(r *CompletionReceipt) { r.ExpiresAt = now.Add(-time.Second) }, func(r *CompletionReceipt) { r.LandedAncestry = false },
		func(r *CompletionReceipt) { r.Status = "failed" }, func(r *CompletionReceipt) { r.HeadSHA = "wrong" },
	}
	for i, mutate := range mutants {
		r := base
		r.Evidence = map[string]bool{"tests": true, "landed": true}
		mutate(&r)
		if err := WriteCompletionReceipt(root, path, r); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCompletionReceipt(root, path, exp); err == nil {
			t.Fatalf("mutant %d accepted", i)
		}
	}
	if err := WriteCompletionReceipt(root, path, base); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	raw[len(raw)/2] ^= 1
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCompletionReceipt(root, path, exp); err == nil {
		t.Fatal("tamper accepted")
	}
}

func TestCompletionReceiptRefusesUnsafeAndMalformedInputs(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "state", "mission", "governance")
	path := filepath.Join(dir, "completion.json")
	now := time.Now()
	base := CompletionReceipt{Version: 1, MissionID: "m", ContractHash: "c", SnapshotHash: "s", HeadSHA: "h", Issuer: "completion-auditor", Status: "passed", Evidence: map[string]bool{"done": true}, ExpiresAt: now.Add(time.Hour)}
	exp := CompletionExpectation{MissionID: "m", ContractHash: "c", SnapshotHash: "s", HeadSHA: "h", RequiredEvidence: []string{"done"}, Now: now}
	if err := WriteCompletionReceipt(root, "relative.json", base); err == nil {
		t.Fatal("relative receipt accepted")
	}
	if _, err := LoadCompletionReceipt(root, path, exp); err == nil {
		t.Fatal("missing receipt accepted")
	}
	if err := WriteCompletionReceipt(root, path, base); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCompletionReceipt(root, path, exp); err == nil {
		t.Fatal("weak receipt accepted")
	}
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCompletionReceipt(root, path, exp); err == nil {
		t.Fatal("malformed receipt accepted")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "target.json")
	if err := os.WriteFile(target, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if err := WriteCompletionReceipt(root, path, base); err == nil {
		t.Fatal("symlink receipt overwritten")
	}
	if _, err := LoadCompletionReceipt(root, path, exp); err == nil {
		t.Fatal("symlink receipt loaded")
	}
	if _, err := LoadCompletionReceipt(root, filepath.Join(root, "outside.json"), exp); err == nil {
		t.Fatal("outside receipt loaded")
	}
	invalidTime := base
	invalidTime.ExpiresAt = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := WriteCompletionReceipt(root, filepath.Join(dir, "invalid-time.json"), invalidTime); err == nil {
		t.Fatal("unencodable receipt accepted")
	}
	directoryPath := filepath.Join(dir, "directory.json")
	if err := os.Mkdir(directoryPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := WriteCompletionReceipt(root, directoryPath, base); err == nil {
		t.Fatal("directory destination accepted")
	}
}
