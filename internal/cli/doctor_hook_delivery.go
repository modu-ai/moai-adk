// Package cli — doctor_hook_delivery.go
//
// Hook Delivery doctor check (SPEC-UPDATE-HOOK-DELIVERY-001, Option B —
// detect + guide). The shipped template's hook entries are compared against
// the project's .claude/settings.json within the hook event keys the project
// already carries; entries the template carries but the project file lacks
// are reported with per-entry placement and remediation guidance.
//
// READ-ONLY by mandate (REQ-UHD-008): this check never writes the user's
// settings.json. `moai update` merge behavior is deliberately unchanged —
// Option A (delivery) was rejected by the operator on 2026-09-03; the merge
// path's silent drop of additions inside carried event keys is documented in
// the SPEC and characterized by TestMergeDropsTemplateAdditionInsideCarriedEventKey.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/template"
)

// hookDeliveryScriptPattern matches the .claude/hooks/moai handler path a
// hook entry references. Forward slashes are canonical in settings.json hook
// args on every platform (the entries run through bash), so no separator
// alternation is needed.
var hookDeliveryScriptPattern = regexp.MustCompile(`\.claude/hooks/moai/[A-Za-z0-9._/-]+\.sh`)

// hookDeliveryCheckCommand is the post-update deletion check the guidance
// points at: the manual verification a user runs after `moai update` to see
// whether the update deleted managed files from their project.
const hookDeliveryCheckCommand = "git status --porcelain | grep '^ D'"

// checkHookDelivery reports hook entries the shipped template's settings.json
// carries but the project's .claude/settings.json is missing, within hook
// event keys the project already carries.
//
// Status contract:
//   - ok   — full parity, or one of the informational skips below
//     (no settings.json, no hooks object, template set unavailable)
//   - warn — missing entries found, and/or a malformed input anomaly
//     (unparseable settings.json, a hook event key holding a non-array)
//
// The check is advisory (doctor exits 0 on warn) and never writes anything.
func checkHookDelivery(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Hook Delivery"}

	settingsPath := filepath.Join(projectRoot, ".claude", "settings.json")
	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			check.Status = uikit.CheckOK
			check.Message = "skipped — no .claude/settings.json in this project"
			return check
		}
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf(".claude/settings.json unreadable — hook-delivery comparison skipped: %v", err)
		return check
	}

	var userRoot map[string]any
	if err := json.Unmarshal(raw, &userRoot); err != nil {
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf(".claude/settings.json is not valid JSON — hook-delivery comparison skipped: %v", err)
		return check
	}

	userHooksValue, present := userRoot["hooks"]
	if !present {
		check.Status = uikit.CheckOK
		check.Message = "skipped — .claude/settings.json carries no hooks"
		return check
	}
	userHooks, ok := userHooksValue.(map[string]any)
	if !ok {
		check.Status = uikit.CheckWarn
		check.Message = ".claude/settings.json hooks value is not an object — hook-delivery comparison skipped"
		return check
	}

	templateHooks, err := renderedTemplateHooks(projectRoot)
	if err != nil {
		// The template side is the binary's own embedded content; failing to
		// render it is an internal anomaly, not a user-facing problem, so the
		// check degrades to an informational skip rather than a nag.
		check.Status = uikit.CheckOK
		check.Message = "skipped — shipped template hook set unavailable"
		if verbose {
			check.Detail = fmt.Sprintf("template render error: %v", err)
		}
		return check
	}

	var missing []string
	var anomalies []string
	for _, eventKey := range sortedTemplateEventKeys(templateHooks) {
		userArr, ok := userHooks[eventKey].([]any)
		if !ok {
			if _, carried := userHooks[eventKey]; carried {
				// REQ-UHD-011: a carried event key holding a non-array is
				// reported and skipped — never modified, never guessed at.
				anomalies = append(anomalies, fmt.Sprintf("hooks.%s is not an array — skipped", eventKey))
			}
			// A key the user does NOT carry is template-introduced and
			// delivered by today's merge (REQ-UHD-001) — not a gap, so it is
			// deliberately not reported here.
			continue
		}
		templateArr, templateIsArr := templateHooks[eventKey].([]any)
		if !templateIsArr {
			continue // malformed template side: skip the key rather than panic
		}
		templateIDs := hookEntryIdentities(templateArr)
		userIDs := hookEntryIdentities(userArr)
		for id, displayName := range templateIDs {
			if _, found := userIDs[id]; !found {
				missing = append(missing, fmt.Sprintf("hooks.%s missing %s", eventKey, displayName))
			}
		}
	}

	if len(missing) == 0 && len(anomalies) == 0 {
		check.Status = uikit.CheckOK
		check.Message = "hook entries match the shipped template"
		if verbose {
			check.Detail = fmt.Sprintf("compared %d carried hook event key(s)", len(userHooks))
		}
		return check
	}

	check.Status = uikit.CheckWarn
	sort.Strings(missing)
	sort.Strings(anomalies)
	lines := append([]string{}, missing...)
	lines = append(lines, anomalies...)
	if len(missing) > 0 {
		lines = append(lines, fmt.Sprintf(
			"re-add each missing entry under the named event key (copy the block from the template settings.json of your moai version); after moai update verify no managed file was deleted: %s",
			hookDeliveryCheckCommand))
	}
	check.Message = strings.Join(lines, "; ")

	if verbose {
		var detail []string
		for _, m := range missing {
			detail = append(detail, m+" — add the entry inside that event key's array in .claude/settings.json")
		}
		detail = append(detail, anomalies...)
		detail = append(detail, fmt.Sprintf(
			"post-update verification: %s — a hit means moai update deleted a managed file; restore it with git restore -- <path> before re-adding hook entries",
			hookDeliveryCheckCommand))
		check.Detail = strings.Join(detail, "\n")
	}
	return check
}

// renderedTemplateHooks renders the embedded settings.json.tmpl for the build
// under test and returns its hooks object. The render honors the project's
// hook.opt_in toggle so a project that opted out of the observability hook
// series is never flagged for hooks its own updates deliberately omitted.
func renderedTemplateHooks(projectRoot string) (map[string]any, error) {
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return nil, fmt.Errorf("load embedded templates: %w", err)
	}
	rendered, err := template.NewRenderer(embedded).Render(
		".claude/settings.json.tmpl",
		template.NewTemplateContext(
			template.WithPlatform(runtime.GOOS),
			template.WithHookOptIn(readHookOptInEnabled(projectRoot)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("render template settings.json: %w", err)
	}
	var root map[string]any
	if err := json.Unmarshal(rendered, &root); err != nil {
		return nil, fmt.Errorf("parse rendered template settings.json: %w", err)
	}
	hooks, ok := root["hooks"].(map[string]any)
	if !ok {
		return nil, errors.New("rendered template settings.json carries no hooks object")
	}
	return hooks, nil
}

// hookEntryIdentities derives one identity string per entry of a hook event
// array (SPEC-UPDATE-HOOK-DELIVERY-001 design.md §G identity rule). An entry
// referencing .claude/hooks/moai handler scripts is identified by those paths
// plus its matcher value — the matcher distinguishes the several blocks that
// wire the SAME handler for different event patterns (the template's
// hooks.PreToolUse carries one block per matcher, and the defect surface this
// SPEC reports on is exactly "a new matcher block inside a carried event key").
// Cosmetic edits elsewhere in the entry (timeout, wrapper text) do not change
// the identity. An entry referencing no script falls back to its canonical
// JSON serialization. The returned map keys are identities, values are
// human-readable names for reports.
func hookEntryIdentities(entries []any) map[string]string {
	identities := make(map[string]string, len(entries))
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			continue // unserializable entry cannot be identified; skip it
		}
		scripts := hookDeliveryScriptPattern.FindAllString(string(data), -1)
		matcher := ""
		if entryMap, ok := entry.(map[string]any); ok {
			if m, ok := entryMap["matcher"].(string); ok {
				matcher = m
			}
		}
		if len(scripts) > 0 {
			// Dedupe repeated references while preserving first-seen order.
			seen := make(map[string]bool, len(scripts))
			unique := scripts[:0]
			for _, s := range scripts {
				if !seen[s] {
					seen[s] = true
					unique = append(unique, s)
				}
			}
			id := strings.Join(unique, ",")
			if matcher != "" {
				id += "|" + matcher
			}
			name := path.Base(unique[0])
			if matcher != "" {
				name = fmt.Sprintf("%s (matcher %s)", name, matcher)
			}
			identities[id] = name
			continue
		}
		identities[string(data)] = string(data)
	}
	return identities
}

// sortedTemplateEventKeys returns the template hooks' event keys in
// deterministic order so report lines are stable across runs.
func sortedTemplateEventKeys(hooks map[string]any) []string {
	keys := make([]string, 0, len(hooks))
	for k := range hooks {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
