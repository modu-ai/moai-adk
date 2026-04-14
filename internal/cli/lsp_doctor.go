package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
)

// ErrLSPMissingServers is returned by `moai lsp doctor` when one or more
// language servers required by the project's detected languages are not
// installed. Cobra surfaces this as a non-zero exit code (REQ-LM-007 AC4).
var ErrLSPMissingServers = errors.New("lsp doctor: missing language servers for detected project languages")

// LSPServerStatus represents the install status of a single language server.
type LSPServerStatus struct {
	// Language is the canonical language identifier.
	Language string
	// Command is the primary binary name from lsp.yaml.
	Command string
	// Installed reports whether the binary (or a fallback) is available in PATH.
	Installed bool
	// ResolvedBinary is the binary that was found (primary or fallback).
	// Empty when Installed is false.
	ResolvedBinary string
	// InstallHint is the user-facing install command from lsp.yaml.
	InstallHint string
}

// LSPDoctorReport is the result of runLSPDoctorReport (REQ-LM-007).
type LSPDoctorReport struct {
	// ProjectLanguages is the list of languages detected via project_markers.
	ProjectLanguages []string
	// InstalledServers is the list of language server statuses where Installed=true.
	InstalledServers []LSPServerStatus
	// MissingServers is the list of language server statuses where Installed=false.
	MissingServers []LSPServerStatus
	// ReadinessStatus is the aggregate status: "ready", "partial", or "missing".
	ReadinessStatus string
}

// lspCmd is the parent `moai lsp` subcommand.
//
// @MX:NOTE: [AUTO] lspCmd — parent command for moai lsp subcommands (doctor, etc.)
var lspCmd = &cobra.Command{
	Use:     "lsp",
	Short:   "LSP server management utilities",
	GroupID: "tools",
	Long:    "Commands for managing and diagnosing LSP (Language Server Protocol) servers.",
}

// lspDoctorCmd is the `moai lsp doctor` subcommand (REQ-LM-007).
var lspDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Report LSP server readiness for the current project",
	Long: `Scans the current project for language markers, then reports:
  - Which languages the project uses
  - Which language servers are installed
  - Which servers are missing + install hints
  - Aggregate readiness status

Exits with a non-zero status when any language server required by the
detected project languages is missing (REQ-LM-007 AC4).`,
	RunE: runLSPDoctor,
}

// lspDoctorJSON, when true, serialises the report as JSON instead of the
// human-readable text format (REQ-LM-007 AC4).
var lspDoctorJSON bool

// runLSPDoctor executes `moai lsp doctor`, honouring the --json flag and
// returning ErrLSPMissingServers when required language servers are missing.
func runLSPDoctor(cmd *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("lsp doctor: get working directory: %w", err)
	}
	lspYAML := filepath.Join(cwd, ".moai", "config", "sections", "lsp.yaml")
	report, err := runLSPDoctorReport(cwd, lspYAML)
	if err != nil {
		return fmt.Errorf("lsp doctor: %w", err)
	}

	out := cmd.OutOrStdout()
	if lspDoctorJSON {
		if err := renderLSPDoctorReportJSON(out, report); err != nil {
			return fmt.Errorf("lsp doctor: json render: %w", err)
		}
	} else {
		renderLSPDoctorReport(out, report)
	}

	// Non-zero exit when any detected project language lacks its server.
	// ReadinessStatus is "missing" when no required server is installed and
	// "partial" when some are missing.
	if report.ReadinessStatus == "missing" || report.ReadinessStatus == "partial" {
		return ErrLSPMissingServers
	}
	return nil
}

// renderLSPDoctorReportJSON serialises the report as pretty-printed JSON
// (REQ-LM-007 AC4). Machine-readable output for CI and tooling pipelines.
func renderLSPDoctorReportJSON(w io.Writer, r *LSPDoctorReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

func init() {
	lspDoctorCmd.Flags().BoolVar(&lspDoctorJSON, "json", false, "Emit report as machine-readable JSON")
	lspCmd.AddCommand(lspDoctorCmd)
	rootCmd.AddCommand(lspCmd)
}

// runLSPDoctorReport builds an LSPDoctorReport for the given project directory (REQ-LM-007).
//
// @MX:ANCHOR: [AUTO] runLSPDoctorReport — core logic for moai lsp doctor
// @MX:REASON: fan_in >= 3 — lspDoctorCmd.RunE, unit tests, and integration paths all call this
func runLSPDoctorReport(projectDir, lspYAMLPath string) (*LSPDoctorReport, error) {
	cfg, err := config.Load(lspYAMLPath)
	if err != nil {
		return nil, fmt.Errorf("load lsp config %q: %w", lspYAMLPath, err)
	}

	// Detect project languages via project_markers (REQ-LM-001, REQ-LM-002).
	projectLangs := detectProjectLanguages(projectDir, cfg)

	// Check binary availability for each server.
	var installed, missing []LSPServerStatus
	// Sorted for deterministic output.
	langs := sortedLanguages(cfg)
	for _, lang := range langs {
		sc := cfg.Servers[lang]
		resolved, found := resolveAnyBinary(sc.Command, sc.FallbackBinaries)
		status := LSPServerStatus{
			Language:       lang,
			Command:        sc.Command,
			Installed:      found,
			ResolvedBinary: resolved,
			InstallHint:    sc.InstallHint,
		}
		if found {
			installed = append(installed, status)
		} else {
			missing = append(missing, status)
		}
	}

	// Compute aggregate readiness based on project languages.
	readiness := computeReadiness(projectLangs, installed, missing)

	return &LSPDoctorReport{
		ProjectLanguages: projectLangs,
		InstalledServers: installed,
		MissingServers:   missing,
		ReadinessStatus:  readiness,
	}, nil
}

// detectProjectLanguages returns the canonical language identifiers for which at
// least one project_marker exists in projectDir (REQ-LM-001).
func detectProjectLanguages(projectDir string, cfg *config.ServersConfig) []string {
	var detected []string
	langs := sortedLanguages(cfg)
	for _, lang := range langs {
		sc := cfg.Servers[lang]
		if hasAnyMarker(projectDir, sc.ProjectMarkers) {
			detected = append(detected, lang)
		}
	}
	return detected
}

// hasAnyMarker reports whether any of the marker filenames exist directly in dir.
// Glob patterns (e.g. "*.csproj") are matched via filepath.Glob.
func hasAnyMarker(dir string, markers []string) bool {
	for _, marker := range markers {
		pattern := filepath.Join(dir, marker)
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return true
		}
	}
	return false
}

// resolveAnyBinary tries the primary command and then each fallback in order.
// Returns the resolved binary path and true on first success, or ("", false) if all fail.
func resolveAnyBinary(primary string, fallbacks []string) (string, bool) {
	candidates := append([]string{primary}, fallbacks...)
	for _, cmd := range candidates {
		if cmd == "" {
			continue
		}
		if p, err := exec.LookPath(cmd); err == nil {
			return p, true
		}
	}
	return "", false
}

// computeReadiness returns "ready", "partial", or "missing" based on whether the
// servers for detected project languages are installed.
func computeReadiness(projectLangs []string, installed, missing []LSPServerStatus) string {
	if len(projectLangs) == 0 {
		return "ready"
	}

	langSet := make(map[string]bool, len(projectLangs))
	for _, l := range projectLangs {
		langSet[l] = true
	}

	missingNeeded := 0
	for _, m := range missing {
		if langSet[m.Language] {
			missingNeeded++
		}
	}

	installedNeeded := 0
	for _, s := range installed {
		if langSet[s.Language] {
			installedNeeded++
		}
	}

	switch {
	case missingNeeded == 0:
		return "ready"
	case installedNeeded > 0:
		return "partial"
	default:
		return "missing"
	}
}

// sortedLanguages returns the language keys from cfg sorted alphabetically.
func sortedLanguages(cfg *config.ServersConfig) []string {
	langs := make([]string, 0, len(cfg.Servers))
	for lang := range cfg.Servers {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
}

// renderLSPDoctorReport writes a human-readable report to w (REQ-LM-007).
func renderLSPDoctorReport(w io.Writer, r *LSPDoctorReport) {
	fmt.Fprintf(w, "=== moai lsp doctor ===\n\n")

	fmt.Fprintf(w, "Project languages detected: ")
	if len(r.ProjectLanguages) == 0 {
		fmt.Fprintf(w, "(none — no project markers found)\n")
	} else {
		fmt.Fprintf(w, "%v\n", r.ProjectLanguages)
	}

	fmt.Fprintf(w, "\nInstalled servers (%d):\n", len(r.InstalledServers))
	for _, s := range r.InstalledServers {
		fmt.Fprintf(w, "  [ok]  %-20s -> %s\n", s.Language, s.ResolvedBinary)
	}

	if len(r.MissingServers) > 0 {
		fmt.Fprintf(w, "\nMissing servers (%d):\n", len(r.MissingServers))
		for _, s := range r.MissingServers {
			fmt.Fprintf(w, "  [--]  %-20s (binary: %s)\n", s.Language, s.Command)
			if s.InstallHint != "" {
				fmt.Fprintf(w, "        install: %s\n", s.InstallHint)
			}
		}
	}

	fmt.Fprintf(w, "\nReadiness: %s\n", r.ReadinessStatus)
}
