package codexwiring

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

const envMoaiCodexBin = "MOAI_CODEX_BIN"

func codexLoaderHasDiagnostic(output, required string) bool {
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "Error: ") && strings.Contains(line, required) {
			return true
		}
		if line == "Error loading config.toml:" && i+1 < len(lines) &&
			strings.Contains(lines[i+1], required) {
			return true
		}
	}
	return false
}

// TestGeneratedConfigKeysLoadInCodex checks the product writer with Codex's
// strict loader before the non-git fixture can start a model turn.
func TestGeneratedConfigKeysLoadInCodex(t *testing.T) {
	bin := os.Getenv(envMoaiCodexBin)
	if bin == "" {
		bin = "codex"
	}
	resolved, err := exec.LookPath(bin)
	if err != nil {
		t.Skipf("CODEX_NOT_INSTALLED: %v", err)
	}
	root := t.TempDir() // deliberately not a git repository
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".codex"), 0o700); err != nil {
		t.Fatal(err)
	}
	trust := "[projects." + strconv.Quote(root) + "]\ntrust_level = \"trusted\"\n"
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(trust), 0o600); err != nil {
		t.Fatal(err)
	}
	generated := string(EnsureMCPTable(nil))
	if !strings.Contains(generated, "[mcp_servers.moai]\n") ||
		!strings.Contains(generated, "default_tools_approval_mode = \"writes\"") {
		t.Fatalf("writer did not emit the existing canonical MCP table: %q", generated)
	}
	check := func(t *testing.T, contents string) string {
		t.Helper()
		path := filepath.Join(root, ".codex", "config.toml")
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, resolved, "exec", "--strict-config",
			"-C", root, "--json",
			"-c", `model_providers.dead={name="dead",base_url="http://127.0.0.1:9/v1",wire_api="responses",stream_max_retries=0,request_max_retries=0}`,
			"-c", `model_provider="dead"`,
			"loader check")
		cmd.Dir = root
		cmd.Env = []string{
			"PATH=" + os.Getenv("PATH"),
			"HOME=" + t.TempDir(),
			"CODEX_HOME=" + home,
			"LANG=C",
		}
		out, runErr := cmd.CombinedOutput()
		if ctx.Err() != nil {
			t.Fatalf("strict loader check exceeded its non-model bound: %v", ctx.Err())
		}
		exitErr, ok := runErr.(*exec.ExitError)
		if !ok || exitErr.ExitCode() != 1 {
			t.Fatalf("strict loader exit = %v, want exit 1; output: %s", runErr, out)
		}
		t.Logf("strict loader exit=1 output=%q", strings.TrimSpace(string(out)))
		return string(out)
	}
	t.Run("generated_config", func(t *testing.T) {
		out := check(t, generated)
		if codexLoaderHasDiagnostic(out, "") ||
			strings.Contains(out, "unknown configuration field") ||
			!strings.Contains(out, "Not inside a trusted directory") {
			t.Fatalf("generated config did not reach the non-git gate: %s", out)
		}
	})
	t.Run("positive_control_misspelled_key", func(t *testing.T) {
		typo := generated + "\n[mcp_servers.moai.tools.codex_role_audit]\napproval_modex = \"approve\"\n"
		out := check(t, typo)
		if !codexLoaderHasDiagnostic(out,
			"unknown configuration field `mcp_servers.moai.tools.codex_role_audit.approval_modex`") {
			t.Fatalf("misspelled key escaped strict loader: %s", out)
		}
	})
	t.Run("positive_control_bad_enum", func(t *testing.T) {
		bad := strings.Replace(generated, `default_tools_approval_mode = "writes"`,
			`default_tools_approval_mode = "bogus"`, 1)
		out := check(t, bad)
		if !codexLoaderHasDiagnostic(out, "unknown variant `bogus`") {
			t.Fatalf("invalid enum escaped strict loader: %s", out)
		}
	})
	t.Run("diagnostic_matcher_controls", func(t *testing.T) {
		required := "unknown configuration field `mcp_servers.moai.tools.codex_role_audit.approval_modex`"
		for _, output := range []string{
			"Error: /tmp/config.toml: " + required,
			"Error loading config.toml:\n/tmp/config.toml: " + required,
		} {
			if !codexLoaderHasDiagnostic(output, required) {
				t.Fatalf("valid loader diagnostic rejected: %s", output)
			}
		}
		for _, output := range []string{
			"Not inside a trusted directory",
			"warning: " + required,
			"Error loading config.toml:",
			"Error loading config.toml:\nother line\n" + required,
			"Error: unknown configuration field `mcp_servers.moai.tools.codex_role_audit.approval_mode`",
			"prefix Error: " + required,
		} {
			if codexLoaderHasDiagnostic(output, required) {
				t.Fatalf("invalid loader diagnostic accepted: %s", output)
			}
		}
		variant := "unknown variant `bogus`"
		for _, output := range []string{
			"Error: /tmp/config.toml: " + variant,
			"Error loading config.toml:\n/tmp/config.toml: " + variant,
		} {
			if !codexLoaderHasDiagnostic(output, variant) {
				t.Fatalf("valid enum loader diagnostic rejected: %s", output)
			}
		}
		for _, output := range []string{
			"Error: unknown variant `other`",
			"Error loading config.toml:\nunknown variant `other`",
		} {
			if codexLoaderHasDiagnostic(output, variant) {
				t.Fatalf("wrong enum diagnostic accepted: %s", output)
			}
		}
	})
}
