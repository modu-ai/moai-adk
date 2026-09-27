// doctor_served_model.go — `moai doctor` check: the model that actually served
// each subagent.
//
// The SubagentStop hook records the served model going forward; this check is
// the read-only sweep over transcripts that already exist. It classifies each
// subagent transcript with the SAME implementation the hook uses
// (hook.ObserveServedModel), so the two surfaces cannot disagree, and reports
// gate-auditor runs whose served model drifted from the expected one or could
// not be determined.
//
// Transcripts sit under the slug of the path a session STARTED in, not under
// the git root, so a sweep of the primary slug alone would miss every session
// launched inside a worktree. The swept slugs are: the primary checkout's
// slug, every slug that begins with that slug plus the worktree-directory
// infix, and the slug of every path `git worktree list` names.
//
// Read-only and advisory: the check writes nothing, and no finding of its own
// can fail the doctor run.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// servedModelCheckName is the doctor check identifier (also the value accepted
// by `moai doctor --check`).
const servedModelCheckName = "Served Model"

// worktreeSlugInfix joins a primary checkout's slug to the name of a worktree
// created under its .claude/worktrees/ directory.
const worktreeSlugInfix = "--claude-worktrees-"

// servedScanInputs is everything the sweep reads from its environment,
// resolved once so a test can supply a fixture instead.
type servedScanInputs struct {
	bases         []string       // Claude configuration base directories
	primary       string         // primary checkout path
	worktreePaths []string       // paths `git worktree list` names
	cfg           *config.Config // project configuration; nil when unreadable
}

// checkServedModel runs the sweep for the project containing cwd.
func checkServedModel(cwd string, verbose bool) DiagnosticCheck {
	return runServedModelScan(servedScanInputsFor(cwd), verbose)
}

func servedScanInputsFor(cwd string) servedScanInputs {
	in := servedScanInputs{bases: escalation.ClaudeConfigBases()}
	tree := auditreceipt.TreeRootFromCWD(cwd)
	if tree == "" {
		tree = cwd
	}
	in.primary = tree
	if primary, entries, err := auditreceipt.IdentifyPrimaryCheckout(tree); err == nil {
		in.primary = primary
		for _, e := range entries {
			in.worktreePaths = append(in.worktreePaths, e.Path)
		}
	}
	if cfg, err := config.NewConfigManager().Load(tree); err == nil {
		in.cfg = cfg
	}
	return in
}

// servedSweepSlugs returns the project directories of one base that the sweep
// covers, sorted.
func servedSweepSlugs(base, primarySlug string, listed map[string]bool) []string {
	entries, err := os.ReadDir(filepath.Join(base, "projects"))
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() {
			continue
		}
		if name == primarySlug || strings.HasPrefix(name, primarySlug+worktreeSlugInfix) || listed[name] {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// dedupeBases drops empty and repeated bases, comparing canonical spellings.
func dedupeBases(bases []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, b := range bases {
		if strings.TrimSpace(b) == "" {
			continue
		}
		key := auditreceipt.Canonical(b)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, b)
	}
	return out
}

// runServedModelScan sweeps the transcripts named by in and renders the check.
func runServedModelScan(in servedScanInputs, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: servedModelCheckName}
	primarySlug := escalation.MemorySlug(in.primary)
	listed := map[string]bool{}
	for _, p := range in.worktreePaths {
		listed[escalation.MemorySlug(p)] = true
	}

	var searched, paths []string
	bases := dedupeBases(in.bases)
	for _, base := range bases {
		for _, slug := range servedSweepSlugs(base, primarySlug, listed) {
			searched = append(searched, filepath.Join(base, "projects", slug))
			found, _ := filepath.Glob(filepath.Join(base, "projects", slug, "*", "subagents", "agent-*.jsonl"))
			sort.Strings(found)
			paths = append(paths, found...)
		}
	}
	observations := observeServedTranscripts(paths, in.cfg)

	counts := map[string]int{}
	var auditorFindings, otherFindings []string
	swept := len(paths)
	for i, obs := range observations {
		counts[obs.Verdict]++
		if obs.Verdict != hook.ServedVerdictDrift && obs.Verdict != hook.ServedVerdictUnknown {
			continue
		}
		line := servedFindingLine(paths[i], obs)
		if auditreceipt.IsAuditorAgent(obs.AgentType) {
			auditorFindings = append(auditorFindings, line)
		} else {
			otherFindings = append(otherFindings, line)
		}
	}

	searchedLine := fmt.Sprintf("bases searched: %s; slug directories swept (%d):%s",
		strings.Join(bases, ", "), len(searched), servedDetailList(searched))
	if len(searched) == 0 {
		searchedLine = fmt.Sprintf("bases searched: %s; slugs looked for: %s, %s%s*, and %d listed worktree slug(s) — none present",
			strings.Join(bases, ", "), primarySlug, primarySlug, worktreeSlugInfix, len(listed))
	}

	if swept == 0 {
		check.Status = uikit.CheckInfo
		check.Message = "swept 0 subagent transcripts — nothing to judge"
		check.Detail = searchedLine
		return check
	}

	check.Message = fmt.Sprintf("swept %d subagent transcripts: ok %d, served_drift %d, unknown %d, unmapped %d; %d gate-auditor run(s) flagged",
		swept, counts[hook.ServedVerdictOK], counts[hook.ServedVerdictDrift], counts[hook.ServedVerdictUnknown],
		counts[hook.ServedVerdictUnmapped], len(auditorFindings))
	detail := searchedLine
	if len(auditorFindings) > 0 {
		detail += "\n  gate-auditor runs served_drift or unknown:" + servedDetailList(auditorFindings)
	}
	if verbose && len(otherFindings) > 0 {
		detail += "\n  other runs served_drift or unknown:" + servedDetailList(otherFindings)
	}
	check.Detail = detail
	check.Status = uikit.CheckOK
	if len(auditorFindings) > 0 {
		check.Status = uikit.CheckWarn
	}
	return check
}

// servedSweepWorkers bounds the concurrent transcript reads of one sweep.
const servedSweepWorkers = 4

// observeServedTranscripts classifies every transcript, returning the
// observations in the order of paths.
//
// @MX:WARN: [AUTO] bounded goroutine fan-out over transcript reads
// @MX:REASON: a sweep reads every subagent transcript on the machine (thousands of files, gigabytes); a fixed pool keeps doctor responsive without unbounded concurrency, and each worker writes only its own result index
func observeServedTranscripts(paths []string, cfg *config.Config) []hook.ServedObservation {
	out := make([]hook.ServedObservation, len(paths))
	next := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < servedSweepWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				out[i] = hook.ObserveServedModel(paths[i], "", cfg)
			}
		}()
	}
	for i := range paths {
		next <- i
	}
	close(next)
	wg.Wait()
	return out
}

// servedDetailList renders items one per indented line.
func servedDetailList(items []string) string {
	if len(items) == 0 {
		return " none"
	}
	return "\n    " + strings.Join(items, "\n    ")
}

// servedFindingLine renders one flagged run: session/agent file, agent type,
// verdict, expected model and served set.
func servedFindingLine(path string, obs hook.ServedObservation) string {
	session := filepath.Base(filepath.Dir(filepath.Dir(path)))
	agentFile := strings.TrimSuffix(filepath.Base(path), ".jsonl")
	expected := obs.ExpectedModel
	if expected == "" {
		expected = "-"
	}
	line := fmt.Sprintf("%s %s/%s %s expected=%s served=[%s]",
		obs.AgentType, session, agentFile, obs.Verdict, expected, strings.Join(obs.ServedModels, ","))
	if obs.Cause != "" {
		line += " (" + obs.Cause + ")"
	}
	return line
}
