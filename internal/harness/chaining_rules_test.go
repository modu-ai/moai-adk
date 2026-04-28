package harness

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestWriteChainingRules_BasicRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chaining-rules.yaml")
	original := ChainingRules{
		Version: 1,
		Chains: []ChainEntry{
			{
				Phase:        "run",
				When:         map[string]string{"agent": "manager-tdd"},
				InsertBefore: []string{"my-harness/ios-architect"},
				InsertAfter:  []string{"my-harness/swiftui-engineer"},
			},
		},
	}
	if err := WriteChainingRules(path, original); err != nil {
		t.Fatalf("write: %v", err)
	}
	round, err := ReadChainingRules(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !reflect.DeepEqual(original, round) {
		t.Errorf("round-trip mismatch:\nwant: %+v\ngot:  %+v", original, round)
	}
}

func TestWriteChainingRules_EmptyArraysMarshalAsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chaining-rules.yaml")
	rules := ChainingRules{
		Version: 1,
		Chains: []ChainEntry{
			{Phase: "plan", When: map[string]string{"agent": "manager-spec"}},
		},
	}
	if err := WriteChainingRules(path, rules); err != nil {
		t.Fatalf("write: %v", err)
	}
	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "insert_before: []") {
		t.Errorf("expected `insert_before: []`, got:\n%s", content)
	}
	if !strings.Contains(content, "insert_after: []") {
		t.Errorf("expected `insert_after: []`, got:\n%s", content)
	}
}

func TestWriteChainingRules_NoEnsureAllowedCall(t *testing.T) {
	// Critical: t.TempDir() absolute paths MUST succeed (no FROZEN guard
	// at file-write level).
	path := filepath.Join(t.TempDir(), "chaining-rules.yaml")
	if err := WriteChainingRules(path, ChainingRules{Version: 1}); err != nil {
		t.Fatalf("layer must not call EnsureAllowed; got: %v", err)
	}
}

func TestReadChainingRules_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chaining-rules.yaml")
	original := ChainingRules{
		Version: 2,
		Chains: []ChainEntry{
			{Phase: "plan", When: map[string]string{"agent": "manager-spec"}, InsertBefore: []string{}, InsertAfter: []string{"my-harness/arch"}},
			{Phase: "run", When: map[string]string{"agent": "manager-tdd"}, InsertBefore: []string{"my-harness/ios"}, InsertAfter: []string{}},
		},
	}
	if err := WriteChainingRules(path, original); err != nil {
		t.Fatal(err)
	}
	read, err := ReadChainingRules(path)
	if err != nil {
		t.Fatal(err)
	}
	if read.Version != 2 {
		t.Errorf("version: %d", read.Version)
	}
	if len(read.Chains) != 2 {
		t.Errorf("chains len: %d", len(read.Chains))
	}
	if read.Chains[0].Phase != "plan" || read.Chains[1].Phase != "run" {
		t.Errorf("phase order broken: %+v", read.Chains)
	}
}

func TestWriteChainingRules_EmptyPath(t *testing.T) {
	if err := WriteChainingRules("", ChainingRules{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestReadChainingRules_EmptyPath(t *testing.T) {
	if _, err := ReadChainingRules(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestReadChainingRules_FileNotFound(t *testing.T) {
	if _, err := ReadChainingRules(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Fatal("expected error")
	}
}

func TestReadChainingRules_InvalidYaml(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.yaml")
	_ = os.WriteFile(path, []byte("not: valid: yaml: [unclosed"), 0o644)
	if _, err := ReadChainingRules(path); err == nil {
		t.Fatal("expected yaml parse error")
	}
}
