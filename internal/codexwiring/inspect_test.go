package codexwiring

import (
	"testing"
)

// TestInspectMCPTableAbsent verifies the doctor-facing inspector on a config
// with no [mcp_servers.moai] table.
func TestInspectMCPTableAbsent(t *testing.T) {
	for _, in := range []string{
		"",
		"model = \"gpt-5\"\n",
		"[mcp_servers.other]\ncommand = \"moai\"\nargs = [\"mcp-server\"]\ndefault_tools_approval_mode = \"writes\"\n",
	} {
		if got := InspectMCPTable([]byte(in)); got.Present {
			t.Errorf("InspectMCPTable(%q).Present = true, want false (no moai table)", in)
		}
	}
}

// TestInspectMCPTableCanonical verifies the inspector recognizes the exact
// table EnsureMCPTable writes.
func TestInspectMCPTableCanonical(t *testing.T) {
	in := EnsureMCPTable([]byte("model = \"gpt-5\"\n"))
	got := InspectMCPTable(in)
	if !got.Present || !got.Canonical {
		t.Errorf("InspectMCPTable(canonical) = %+v, want Present+Canonical", got)
	}
}

// TestInspectMCPTableDrifted verifies the inspector flags a user-modified
// table (drift is the doctor's to report; the writer never repairs it).
func TestInspectMCPTableDrifted(t *testing.T) {
	drifted := []byte("[mcp_servers.moai]\ncommand = \"my-custom-moai\"\nargs = [\"other\"]\n")
	got := InspectMCPTable(drifted)
	if !got.Present {
		t.Fatal("drifted table not detected as present")
	}
	if got.Canonical {
		t.Errorf("user-modified table reported canonical — drift would go unreported")
	}

	// Partial drift: canonical header + one wrong assignment.
	partial := []byte("[mcp_servers.moai]\ncommand = \"moai\"\nargs = [\"other\"]\ndefault_tools_approval_mode = \"writes\"\n")
	if got := InspectMCPTable(partial); got.Canonical {
		t.Errorf("partially-drifted table (args) reported canonical")
	}
}

// TestInspectMCPTableSectionBoundary verifies the inspector stops at the next
// table header — canonical-looking lines under a LATER table must not count.
func TestInspectMCPTableSectionBoundary(t *testing.T) {
	in := []byte("[mcp_servers.moai]\ncommand = \"moai\"\n\n[other]\nargs = [\"mcp-server\"]\ndefault_tools_approval_mode = \"writes\"\n")
	got := InspectMCPTable(in)
	if got.Canonical {
		t.Errorf("assignments under [other] counted into the moai table verdict")
	}
}
