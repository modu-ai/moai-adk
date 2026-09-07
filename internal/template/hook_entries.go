package template

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// SettingsTemplateName is the embedded path of the Claude Code settings
// template, relative to the FS returned by EmbeddedTemplates.
const SettingsTemplateName = ".claude/settings.json.tmpl"

// HookEntry is one Claude Code hook registration reduced to the tuple the
// wiring comparison keys on: (event, matcher, script, if, timeout, async).
//
// Every field is a scalar so the struct is comparable and usable as a map key.
// An absent JSON key maps to its zero value, which is what makes "no async
// key" and "async: false" compare equal — they are the same runtime behaviour.
type HookEntry struct {
	Event   string
	Matcher string
	Script  string
	If      string
	Timeout int
	Async   bool
}

// String renders the entry for a diagnostic message.
func (e HookEntry) String() string {
	s := e.Event + "/" + e.Script
	if e.Matcher != "" {
		s += fmt.Sprintf(" matcher=%q", e.Matcher)
	}
	if e.If != "" {
		s += fmt.Sprintf(" if=%q", e.If)
	}
	return s + fmt.Sprintf(" timeout=%d async=%t", e.Timeout, e.Async)
}

// settingsDoc is the hook-bearing subset of a Claude Code settings document.
// The top-level statusLine block is deliberately not modelled: it carries
// "type": "command" but is not a hook, and counting it is the 34-vs-33 error
// the investigation corrected.
type settingsDoc struct {
	Hooks map[string][]struct {
		Matcher string `json:"matcher"`
		Hooks   []struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
			If      string   `json:"if"`
			Timeout int      `json:"timeout"`
			Async   bool     `json:"async"`
		} `json:"hooks"`
	} `json:"hooks"`
}

// ParseHookEntries extracts the hook-entry set from a Claude Code settings
// document — either the rendered template or a project's settings.json.
// Entries are returned in a deterministic order.
func ParseHookEntries(settings []byte) ([]HookEntry, error) {
	var doc settingsDoc
	if err := json.Unmarshal(settings, &doc); err != nil {
		return nil, fmt.Errorf("parse settings JSON: %w", err)
	}

	var entries []HookEntry
	for event, groups := range doc.Hooks {
		for _, g := range groups {
			for _, h := range g.Hooks {
				entries = append(entries, HookEntry{
					Event:   event,
					Matcher: g.Matcher,
					Script:  hookScriptName(h.Command, h.Args),
					If:      h.If,
					Timeout: h.Timeout,
					Async:   h.Async,
				})
			}
		}
	}
	sortHookEntries(entries)
	return entries, nil
}

// RenderHookEntries renders the settings template from fsys with ctx and
// returns its hook-entry set.
//
// The template source is a parameter rather than the embedded FS so a caller
// can inject a fixture template; production passes EmbeddedTemplates(). This
// seam is what makes "did the check actually render the template?" observable
// — a hardcoded expected-entry list cannot name a script it has never seen.
func RenderHookEntries(fsys fs.FS, ctx *TemplateContext) ([]HookEntry, error) {
	if fsys == nil {
		return nil, fmt.Errorf("render settings template: no template source")
	}
	rendered, err := NewRenderer(fsys).Render(SettingsTemplateName, ctx)
	if err != nil {
		return nil, fmt.Errorf("render settings template: %w", err)
	}
	entries, err := ParseHookEntries(rendered)
	if err != nil {
		return nil, fmt.Errorf("render settings template: %w", err)
	}
	return entries, nil
}

// DiffHookEntries compares two hook-entry sets in BOTH directions.
//
// templateOnly holds entries the template carries and the project does not (a
// missing registration); projectOnly holds entries the project carries and the
// template does not (an extra registration). Identical duplicates are compared
// by multiplicity, so a doubled registration is itself a divergence.
func DiffHookEntries(templateEntries, projectEntries []HookEntry) (templateOnly, projectOnly []HookEntry) {
	tc, pc := countEntries(templateEntries), countEntries(projectEntries)
	for e, n := range tc {
		for i := 0; i < n-pc[e]; i++ {
			templateOnly = append(templateOnly, e)
		}
	}
	for e, n := range pc {
		for i := 0; i < n-tc[e]; i++ {
			projectOnly = append(projectOnly, e)
		}
	}
	sortHookEntries(templateOnly)
	sortHookEntries(projectOnly)
	return templateOnly, projectOnly
}

// HookEntryScripts returns the distinct script names of the given entries,
// sorted, each suffixed with " x<N>" when it occurs more than once.
func HookEntryScripts(entries []HookEntry) []string {
	counts := make(map[string]int, len(entries))
	order := make([]string, 0, len(entries))
	for _, e := range entries {
		if counts[e.Script] == 0 {
			order = append(order, e.Script)
		}
		counts[e.Script]++
	}
	sort.Strings(order)
	out := make([]string, 0, len(order))
	for _, s := range order {
		if counts[s] > 1 {
			out = append(out, fmt.Sprintf("%s x%d", s, counts[s]))
			continue
		}
		out = append(out, s)
	}
	return out
}

func countEntries(in []HookEntry) map[HookEntry]int {
	m := make(map[HookEntry]int, len(in))
	for _, e := range in {
		m[e]++
	}
	return m
}

// hookScriptName derives the script name from a hook entry's command/args.
// The wrapper form is `bash -c '<fallback>' <script-path>`, so the script is
// the last argument naming a path under .claude/hooks/ (or any .sh path). An
// unrecognised shape falls back to the command itself so it is still named
// rather than silently dropped.
func hookScriptName(command string, args []string) string {
	for i := len(args) - 1; i >= 0; i-- {
		a := args[i]
		if strings.Contains(a, "/.claude/hooks/") || strings.HasSuffix(a, ".sh") {
			return path.Base(a)
		}
	}
	if strings.HasSuffix(command, ".sh") {
		return path.Base(command)
	}
	return command
}

func sortHookEntries(entries []HookEntry) {
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		switch {
		case a.Event != b.Event:
			return a.Event < b.Event
		case a.Script != b.Script:
			return a.Script < b.Script
		case a.Matcher != b.Matcher:
			return a.Matcher < b.Matcher
		case a.If != b.If:
			return a.If < b.If
		case a.Timeout != b.Timeout:
			return a.Timeout < b.Timeout
		default:
			return !a.Async && b.Async
		}
	})
}
