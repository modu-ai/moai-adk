package codexwiring

import (
	"strings"
	"testing"
)

// TestEnsureMCPTableCreatesWhenAbsent verifies the create-if-absent contract:
// a config without [mcp_servers.moai] gains exactly the canonical table
// (AC-CW-003, REQ-CW-004).
func TestEnsureMCPTableCreatesWhenAbsent(t *testing.T) {
	in := []byte("# user config\nmodel = \"gpt-5\"\n")
	out := EnsureMCPTable(in)
	body := string(out)
	if !strings.HasPrefix(body, string(in)) {
		t.Errorf("existing content must be byte-preserved before the appended table:\nin:  %q\nout: %q", in, out)
	}
	for _, want := range []string{
		"[mcp_servers.moai]",
		`command = "moai"`,
		`args = ["mcp-server"]`,
		`default_tools_approval_mode = "writes"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("canonical table missing %q:\n%s", want, out)
		}
	}
	if strings.Count(body, "[mcp_servers.moai]") != 1 {
		t.Errorf("exactly one [mcp_servers.moai] table expected, got %d:\n%s",
			strings.Count(body, "[mcp_servers.moai]"), out)
	}
}

// TestEnsureMCPTableNoToolEnumeration verifies the capability-based approval
// stance: no enabled_tools / disabled_tools key is ever written (AC-CW-003,
// REQ-CW-004, spec §A.4).
func TestEnsureMCPTableNoToolEnumeration(t *testing.T) {
	for _, in := range [][]byte{
		nil,
		[]byte("model = \"gpt-5\"\n"),
		[]byte("[mcp_servers.user-server]\ncommand = \"npx\"\n"),
	} {
		out := EnsureMCPTable(in)
		if strings.Contains(string(out), "enabled_tools") || strings.Contains(string(out), "disabled_tools") {
			t.Errorf("tool-name enumeration leaked into config.toml (§A.4 violation):\n%s", out)
		}
	}
}

// TestEnsureMCPTableUserTableUntouched verifies an existing user-owned
// [mcp_servers.moai] table is never overwritten — the doctor reports drift
// instead (AC-CW-007 second clause, REQ-CW-005).
func TestEnsureMCPTableUserTableUntouched(t *testing.T) {
	in := []byte("[mcp_servers.moai]\ncommand = \"my-custom-moai\"\nargs = [\"other\"]\n")
	out := EnsureMCPTable(in)
	if string(out) != string(in) {
		t.Errorf("user-owned [mcp_servers.moai] table must be byte-invariant:\nin:  %q\nout: %q", in, out)
	}
}

// TestEnsureMCPTableOtherServersPreserved verifies user tables under
// mcp_servers survive (AC-CW-007 first clause).
func TestEnsureMCPTableOtherServersPreserved(t *testing.T) {
	in := []byte("[mcp_servers.other]\ncommand = \"npx\"\nargs = [\"-y\", \"some-server\"]\n")
	out := EnsureMCPTable(in)
	body := string(out)
	if !strings.Contains(body, "[mcp_servers.other]") || !strings.Contains(body, "some-server") {
		t.Errorf("user [mcp_servers.other] table lost:\n%s", out)
	}
	if !strings.Contains(body, "[mcp_servers.moai]") {
		t.Errorf("canonical table missing alongside user table:\n%s", out)
	}
}

// TestEnsureMCPTableIdempotent verifies a second pass is a byte no-op
// (REQ-CW-006).
func TestEnsureMCPTableIdempotent(t *testing.T) {
	once := EnsureMCPTable([]byte("model = \"gpt-5\"\n"))
	twice := EnsureMCPTable(once)
	if string(once) != string(twice) {
		t.Errorf("mcp table merge not idempotent:\nonce:  %q\ntwice: %q", once, twice)
	}
}
