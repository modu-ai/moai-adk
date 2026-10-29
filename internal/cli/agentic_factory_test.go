package cli

import (
	"bytes"
	"testing"
)

// Issue #1948: the canonical command name is provisional (lane-17 contract
// SPEC fixes it), so the skeleton is hidden and changes no state.
func TestAgenticFactory_RegisteredHiddenOnRoot(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "agentic-factory" {
			if !c.Hidden {
				t.Error("agentic-factory skeleton must be hidden")
			}
			return
		}
	}
	t.Fatal("agentic-factory not registered on root")
}

func TestAgenticFactory_RunReportsProvisional(t *testing.T) {
	cmd := newAgenticFactoryCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("provisional")) {
		t.Errorf("output lacks provisional notice: %q", out.String())
	}
}

func TestAgenticFactory_LeavesLegacyFactoryFlagsAlone(t *testing.T) {
	// Legacy entry points (`moai cc -f/-l`, `moai factory`) must stay registered.
	want := map[string]bool{"cc": false, "factory": false, "integration": false}
	for _, c := range rootCmd.Commands() {
		if _, ok := want[c.Name()]; ok {
			want[c.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("legacy command %q no longer registered", name)
		}
	}
}
