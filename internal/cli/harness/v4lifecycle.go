// Package harness — v4 harness lifecycle handlers (SPEC-V3R6-HARNESS-V4-001 M4).
//
// ListHarnesses / EditHarness / RemoveHarness implement design §B.3 lifecycle:
//   - list:   enumerate harnesses by scanning .claude/commands/harness/*.md and
//             joining each with its manifest.json (REQ-HV4-011 / AC-HV4-011a)
//   - edit:   locate the manifest + specialist files for editing (manifest is
//             the SSOT; editing it propagates to Runner behavior on next run)
//   - remove: atomic removal of command + workflow + specialists + skills +
//             manifest, fail-closed if any referenced artifact is missing
//             (orphan prevention, REQ-HV4-011 / AC-HV4-011b/c)
//
// The functions are pure filesystem operations against a projectRoot. The cobra
// command wrappers (newHarnessV4ListCmd etc.) live in package cli and delegate
// here; this file holds the testable logic and shares the C-HRA-008 boundary
// guard (TestPropose_NoAskUserQuestion scans this directory).
//
// @MX:ANCHOR: [AUTO] v4 harness lifecycle handlers (list/edit/remove)
// @MX:REASON: [AUTO] fan_in >= 3 candidate: cobra wrappers, lifecycle tests, moai SKILL.md Branch A dispatcher
package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/harness/v4manifest"
)

// v4CommandsDir is the user-owned directory holding thin-wrapper command files
// (.claude/commands/harness/<name>.md) and their co-located manifest
// subdirectories (.claude/commands/harness/<name>/manifest.json). Matches the
// RunnerTemplate read path (M3 runner_template.go readManifest).
const v4CommandsDir = ".claude/commands/harness"

// v4WorkflowsDir holds the Runner Workflow scripts (harness-<name>-run.js).
const v4WorkflowsDir = ".claude/workflows"

// v4AgentsDir holds the specialist agent definitions (user-owned namespace).
const v4AgentsDir = ".claude/agents/harness"

// v4SkillsDir is the parent of the per-harness companion skill directories
// (.claude/skills/harness-<name>-*/).
const v4SkillsDir = ".claude/skills"

// HarnessEntry is a single harness enumerated by ListHarnesses.
type HarnessEntry struct {
	// Name is the harness name (derived from the command filename stem).
	Name string `json:"name"`

	// Domain is the human-readable domain from manifest.json (empty if the
	// manifest is missing or unreadable).
	Domain string `json:"domain"`

	// EntryCommand is the /harness:<name> string from the manifest.
	EntryCommand string `json:"entry_command"`

	// RunnerWorkflow is the harness-<name>-run.js filename from the manifest.
	RunnerWorkflow string `json:"runner_workflow"`

	// ManifestMissing is true when a command file exists but its co-located
	// manifest.json does not (partial state). List surfaces this so the user
	// can decide whether to repair or remove.
	ManifestMissing bool `json:"manifest_missing"`

	// CommandPath is the absolute path to the thin-wrapper command file.
	CommandPath string `json:"command_path"`

	// ManifestPath is the absolute path to the manifest.json (may not exist
	// when ManifestMissing is true).
	ManifestPath string `json:"manifest_path"`
}

// ListHarnesses enumerates every harness under projectRoot by scanning
// .claude/commands/harness/*.md and joining each with its manifest.json
// (AC-HV4-011a). A command whose manifest is missing is still listed with
// ManifestMissing=true — list never crashes on partial state; remove handles
// atomicity. Returns entries sorted by Name for deterministic output.
func ListHarnesses(projectRoot string) ([]HarnessEntry, error) {
	commandsDir := filepath.Join(projectRoot, v4CommandsDir)
	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// No harness directory → zero harnesses. Not an error.
			return nil, nil
		}
		return nil, fmt.Errorf("v4lifecycle: list: read commands dir: %w", err)
	}

	var harnesses []HarnessEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		cmdPath := filepath.Join(commandsDir, e.Name())
		manifestPath := filepath.Join(commandsDir, name, "manifest.json")

		entry := HarnessEntry{
			Name:        name,
			CommandPath: cmdPath,
			ManifestPath: manifestPath,
		}

		if data, mErr := os.ReadFile(manifestPath); mErr == nil {
			var m v4manifest.Manifest
			if jErr := json.Unmarshal(data, &m); jErr == nil {
				entry.Domain = m.Domain
				entry.EntryCommand = m.EntryCommand
				entry.RunnerWorkflow = m.RunnerWorkflow
			}
			// If JSON unmarshal fails, Domain stays empty but the harness is
			// still listed (ManifestMissing=false because the file exists).
		} else if os.IsNotExist(mErr) {
			entry.ManifestMissing = true
		}
		harnesses = append(harnesses, entry)
	}

	sort.Slice(harnesses, func(i, j int) bool {
		return harnesses[i].Name < harnesses[j].Name
	})
	return harnesses, nil
}

// HarnessEditPaths is the set of files EditHarness surfaces for editing. The
// manifest is the SSOT; editing it propagates to Runner behavior on the next
// invocation. The specialist files are listed so the user can revise role
// definitions alongside the manifest.
type HarnessEditPaths struct {
	// Name is the harness name.
	Name string `json:"name"`

	// ManifestPath is the manifest.json absolute path (SSOT).
	ManifestPath string `json:"manifest_path"`

	// SpecialistPaths are the absolute paths to the harness's specialist agent
	// definition files under .claude/agents/harness/.
	SpecialistPaths []string `json:"specialist_paths"`

	// SkillPaths are the companion skill directory paths (if any).
	SkillPaths []string `json:"skill_paths"`
}

// EditHarness locates the manifest + specialist + skill files for the named
// harness so the user (or orchestrator) can open them for editing. The manifest
// MUST exist (it is the SSOT); a missing manifest returns an error so the user
// is not directed to edit a harness whose SSOT is gone.
func EditHarness(projectRoot, name string) (HarnessEditPaths, error) {
	commandsDir := filepath.Join(projectRoot, v4CommandsDir)
	manifestPath := filepath.Join(commandsDir, name, "manifest.json")
	if _, err := os.Stat(manifestPath); err != nil {
		if os.IsNotExist(err) {
			return HarnessEditPaths{}, fmt.Errorf("v4lifecycle: edit: manifest not found for harness %q at %s", name, manifestPath)
		}
		return HarnessEditPaths{}, fmt.Errorf("v4lifecycle: edit: stat manifest: %w", err)
	}

	paths := HarnessEditPaths{
		Name:         name,
		ManifestPath: manifestPath,
	}

	// Specialist agent files: .claude/agents/harness/harness-<name>*-specialist.md
	agentsDir := filepath.Join(projectRoot, v4AgentsDir)
	if entries, err := os.ReadDir(agentsDir); err == nil {
		prefix := "harness-" + name
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if strings.HasPrefix(e.Name(), prefix) && strings.HasSuffix(e.Name(), "-specialist.md") {
				paths.SpecialistPaths = append(paths.SpecialistPaths, filepath.Join(agentsDir, e.Name()))
			}
		}
		sort.Strings(paths.SpecialistPaths)
	}

	// Companion skill directory: .claude/skills/harness-<name>*/
	skillsDir := filepath.Join(projectRoot, v4SkillsDir)
	if entries, err := os.ReadDir(skillsDir); err == nil {
		prefix := "harness-" + name
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			if strings.HasPrefix(e.Name(), prefix) {
				paths.SkillPaths = append(paths.SkillPaths, filepath.Join(skillsDir, e.Name()))
			}
		}
		sort.Strings(paths.SkillPaths)
	}

	return paths, nil
}

// RemoveHarness atomically removes the named harness's command file, manifest,
// Runner workflow, specialist agent files, and companion skill directories
// (AC-HV4-011b). It FAILS CLOSED (AC-HV4-011c): if any referenced artifact is
// missing, it refuses to remove anything and returns an error naming the
// missing artifact — no partial state is left behind.
//
// The atomicity contract: the precheck phase collects every referenced path
// from the manifest and verifies existence BEFORE deleting anything. Only if
// all artifacts are present does the delete phase run.
func RemoveHarness(projectRoot, name string) error {
	commandsDir := filepath.Join(projectRoot, v4CommandsDir)
	cmdPath := filepath.Join(commandsDir, name+".md")
	manifestPath := filepath.Join(commandsDir, name, "manifest.json")

	// --- precheck phase: collect referenced paths from the manifest ---

	// The command file itself is the entry artifact; its absence means the
	// harness does not exist at all.
	if _, err := os.Stat(cmdPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("v4lifecycle: remove: harness %q command file not found at %s", name, cmdPath)
		}
		return fmt.Errorf("v4lifecycle: remove: stat command: %w", err)
	}

	// The manifest is the SSOT — its absence is an orphan-state signal that
	// MUST fail closed (AC-HV4-011c).
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("v4lifecycle: remove: manifest not found for harness %q at %s (orphan state — refusing to leave command behind)", name, manifestPath)
		}
		return fmt.Errorf("v4lifecycle: remove: read manifest: %w", err)
	}
	var m v4manifest.Manifest
	if err := json.Unmarshal(manifestData, &m); err != nil {
		return fmt.Errorf("v4lifecycle: remove: parse manifest: %w", err)
	}

	// Collect every referenced path to verify before deleting.
	runnerPath := filepath.Join(projectRoot, v4WorkflowsDir, m.RunnerWorkflow)
	if _, err := os.Stat(runnerPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("v4lifecycle: remove: Runner workflow not found for harness %q at %s (orphan state)", name, runnerPath)
		}
		return fmt.Errorf("v4lifecycle: remove: stat runner: %w", err)
	}

	// Specialist agent files: glob .claude/agents/harness/harness-<name>*-specialist.md
	// (matches both harness-<name>-auditor-specialist.md and any future
	// harness-<name>-<role>-specialist.md variants).
	agentsDir := filepath.Join(projectRoot, v4AgentsDir)
	var specialistPaths []string
	if entries, dErr := os.ReadDir(agentsDir); dErr == nil {
		prefix := "harness-" + name
		for _, e := range entries {
			if !e.IsDir() && strings.HasPrefix(e.Name(), prefix) && strings.HasSuffix(e.Name(), "-specialist.md") {
				specialistPaths = append(specialistPaths, filepath.Join(agentsDir, e.Name()))
			}
		}
	}
	// Specialists are not strictly required to exist (a harness may have zero
	// generated specialist files yet), so we do NOT fail closed on an empty
	// specialist set — only the manifest + Runner + command are the load-bearing
	// artifacts. But IF specialist files are referenced and some are missing,
	// that partial state also fails closed.
	for _, p := range specialistPaths {
		if _, err := os.Stat(p); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("v4lifecycle: remove: specialist file not found for harness %q at %s (orphan state)", name, p)
			}
			return fmt.Errorf("v4lifecycle: remove: stat specialist: %w", err)
		}
	}

	// Companion skill directories: .claude/skills/harness-<name>*/
	// (matches both harness-<name>/ single-companion and harness-<name>-<role>/
	// multi-specialist layouts).
	skillsDir := filepath.Join(projectRoot, v4SkillsDir)
	var skillPaths []string
	if entries, dErr := os.ReadDir(skillsDir); dErr == nil {
		prefix := "harness-" + name
		for _, e := range entries {
			if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
				skillPaths = append(skillPaths, filepath.Join(skillsDir, e.Name()))
			}
		}
	}
	for _, p := range skillPaths {
		if _, err := os.Stat(p); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("v4lifecycle: remove: skill directory not found for harness %q at %s (orphan state)", name, p)
			}
			return fmt.Errorf("v4lifecycle: remove: stat skill: %w", err)
		}
	}

	// --- delete phase: all prechecks passed; remove everything ---

	// Remove skill directories first (deepest leaf).
	for _, p := range skillPaths {
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("v4lifecycle: remove: delete skill %s: %w", p, err)
		}
	}
	// Remove specialist agent files.
	for _, p := range specialistPaths {
		if err := os.Remove(p); err != nil {
			return fmt.Errorf("v4lifecycle: remove: delete specialist %s: %w", p, err)
		}
	}
	// Remove the Runner workflow.
	if err := os.Remove(runnerPath); err != nil {
		return fmt.Errorf("v4lifecycle: remove: delete runner %s: %w", runnerPath, err)
	}
	// Remove the manifest subdirectory (manifest.json + any co-located files).
	manifestDir := filepath.Join(commandsDir, name)
	if err := os.RemoveAll(manifestDir); err != nil {
		return fmt.Errorf("v4lifecycle: remove: delete manifest dir %s: %w", manifestDir, err)
	}
	// Remove the thin-wrapper command file last.
	if err := os.Remove(cmdPath); err != nil {
		return fmt.Errorf("v4lifecycle: remove: delete command %s: %w", cmdPath, err)
	}

	return nil
}
