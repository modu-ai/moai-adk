package cli

// doctor_sandbox.go implements the `moai doctor sandbox` subcommand.
// It reports sandbox backend availability and per-agent resolved backend
// for the current host. Maps to SPEC-V3R2-RT-003 REQ-005/032.
//
// @MX:NOTE: [AUTO] doctor_sandbox mirrors the doctor_config.go pattern:
//           sandboxCmd is registered as a subcommand of doctorCmd in init().
// @MX:SPEC: SPEC-V3R2-RT-003 REQ-005/032

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/sandbox"
)

// SandboxAvailability holds the detection result for a single backend.
type SandboxAvailability struct {
	Backend   sandbox.Sandbox
	Available bool
	Note      string
}

// sandboxCmd is the `moai doctor sandbox` subcommand.
var sandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "Sandbox backend availability diagnostics",
	Long: `Report sandbox backend availability and per-role resolved backend for the current host.

Maps to SPEC-V3R2-RT-003 REQ-005/032 and AC-V3R2-RT-003-04.

Examples:
  moai doctor sandbox              # Full availability report
  moai doctor sandbox --profile implementer  # Dump resolved profile for role`,
	RunE: runDoctorSandbox,
}

func init() {
	doctorCmd.AddCommand(sandboxCmd)
	sandboxCmd.Flags().String("profile", "", "Dump the resolved sandbox profile for the given agent role (e.g., implementer)")
}

// runDoctorSandbox executes the sandbox diagnostics.
func runDoctorSandbox(cmd *cobra.Command, _ []string) error {
	w := cmd.OutOrStdout()
	profileRole := getStringFlag(cmd, "profile")

	if profileRole != "" {
		return runSandboxProfileDump(w, profileRole)
	}

	return runSandboxAvailabilityReport(w)
}

// runSandboxAvailabilityReport prints backend availability + per-role resolved backend.
func runSandboxAvailabilityReport(w io.Writer) error {
	_, _ = fmt.Fprintln(w, "Sandbox Backend Availability")
	_, _ = fmt.Fprintln(w, "============================")
	_, _ = fmt.Fprintf(w, "OS: %s\n\n", runtime.GOOS)

	// 백엔드별 가용성 확인
	backends := []struct {
		s     sandbox.Sandbox
		check func() bool
	}{
		{sandbox.SandboxBubblewrap, func() bool { return sandbox.NewBubblewrapBackend().Available() }},
		{sandbox.SandboxSeatbelt, func() bool { return sandbox.NewSeatbeltBackend().Available() }},
		{sandbox.SandboxDocker, func() bool { return sandbox.NewDockerBackend().Available() }},
	}

	for _, b := range backends {
		avail := b.check()
		icon := "✓"
		note := "available"
		if !avail {
			icon = "✗"
			switch b.s {
			case sandbox.SandboxBubblewrap:
				note = "unavailable — install via: sudo apt-get install bubblewrap OR flatpak"
			case sandbox.SandboxSeatbelt:
				note = "unavailable — requires macOS 10.5+ with sandbox-exec"
			case sandbox.SandboxDocker:
				note = "unavailable — install Docker or run in CI with MOAI_TEST_DOCKER=1"
			}
		}
		_, _ = fmt.Fprintf(w, "  %s %-12s %s\n", icon, string(b.s)+":", note)
	}

	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Per-Role Resolved Backend")
	_, _ = fmt.Fprintln(w, "-------------------------")

	l := sandbox.NewLauncher()
	ciActive := os.Getenv("CI") == "1"

	roles := []string{"implementer", "tester", "designer", "researcher", "analyst", "reviewer", "architect"}
	for _, role := range roles {
		resolved := l.ResolveForRole(role)
		_, _ = fmt.Fprintf(w, "  %-12s → %s", role, string(resolved))
		if ciActive && resolved == sandbox.SandboxDocker {
			_, _ = fmt.Fprint(w, " (CI=1 override)")
		}
		_, _ = fmt.Fprintln(w)
	}

	if ciActive {
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Note: CI=1 detected — implementer/tester/designer roles use docker backend.")
	}

	return nil
}

// runSandboxProfileDump prints the resolved sandbox profile for the given role.
func runSandboxProfileDump(w io.Writer, role string) error {
	l := sandbox.NewLauncher()
	resolved := l.ResolveForRole(role)

	_, _ = fmt.Fprintf(w, "Resolved sandbox for role %q: %s\n\n", role, string(resolved))

	opts := sandbox.SandboxOptions{
		WritableScope:    []string{"/tmp/agent-worktree", ".moai/state"},
		ReadOnlyScope:    []string{"/usr", "/lib"},
		NetworkAllowlist: sandbox.DefaultNetworkAllowlist,
		MaxOutputBytes:   sandbox.DefaultMaxOutputBytes,
	}

	switch resolved {
	case sandbox.SandboxSeatbelt:
		profile, err := sandbox.GenerateSBPL(opts)
		if err != nil {
			return fmt.Errorf("generate SBPL: %w", err)
		}
		_, _ = fmt.Fprintln(w, "SBPL Profile (sandbox-exec -p <profile>):")
		_, _ = fmt.Fprintln(w, "---")
		_, _ = fmt.Fprintln(w, profile)

	case sandbox.SandboxBubblewrap:
		args, err := sandbox.GenerateBwrapArgs(opts)
		if err != nil {
			return fmt.Errorf("generate bwrap args: %w", err)
		}
		_, _ = fmt.Fprintln(w, "Bubblewrap Args:")
		_, _ = fmt.Fprintln(w, "---")
		_, _ = fmt.Fprintf(w, "bwrap")
		for _, a := range args {
			_, _ = fmt.Fprintf(w, " \\\n  %s", a)
		}
		_, _ = fmt.Fprintln(w, " \\\n  -- <cmd>")

	case sandbox.SandboxDocker:
		snippet, err := sandbox.GenerateDockerSnippet(opts)
		if err != nil {
			return fmt.Errorf("generate docker snippet: %w", err)
		}
		_, _ = fmt.Fprintln(w, "Dockerfile Snippet (docker run):")
		_, _ = fmt.Fprintln(w, "---")
		_, _ = fmt.Fprintln(w, snippet)

	case sandbox.SandboxNone:
		_, _ = fmt.Fprintln(w, "No sandbox profile (sandbox: none — no isolation applied).")

	default:
		return fmt.Errorf("unsupported sandbox backend: %q", resolved)
	}

	return nil
}
