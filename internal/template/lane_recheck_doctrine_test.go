package template

import (
	"io/fs"
	"strings"
	"testing"
)

// Card t1451 guard: the shipped lane-stall doctrine must keep naming WHO wakes a
// stopped lane and WHAT a woken lane reads. The t1370 watchdog was present in
// both stalled card trees and still never ran, because the doctrine left the
// awaken carrier out of scope and the cause table had no row for a lane waiting
// on a delegate that had already stopped.
//
// Each check is a lexical conjunction inside ONE named section of the shipped
// (embedded) copy. It proves the terms of an invariant occur together there; it
// cannot prove a lane obeys them — that limit is read by the sync audit.

type doctrineNeed struct {
	file    string
	heading string // section start; "" = whole file
	need    []string
	forbid  []string
}

// sectionOf returns the text from heading up to the next heading of the same or
// a shallower level, or the whole document when heading is empty.
func sectionOf(doc, heading string) (string, bool) {
	if heading == "" {
		return doc, true
	}
	i := strings.Index(doc, "\n"+heading+"\n")
	if i < 0 {
		return "", false
	}
	rest := doc[i+1+len(heading):]
	level := len(heading) - len(strings.TrimLeft(heading, "#"))
	end := len(rest)
	for _, prefix := range []string{"\n# ", "\n## ", "\n### "}[:level] {
		if j := strings.Index(rest, prefix); j >= 0 && j < end {
			end = j
		}
	}
	return rest[:end], true
}

func TestLaneRecheckDoctrine(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates(): %v", err)
	}
	const (
		skill    = ".claude/skills/moai-lane-watchdog/SKILL.md"
		auto     = ".claude/rules/moai/workflow/auto-semantics.md"
		dispatch = ".claude/rules/moai/workflow/kanban-dispatch.md"
	)
	needs := []doctrineNeed{
		{
			file:    auto,
			heading: "## 5. The awaken rule",
			need:    []string{"standing", "CronCreate", "recurring", "before the first stage", "CronList"},
		},
		{
			file:    auto,
			heading: "## 8. Harness neutrality",
			forbid:  []string{"out of scope"},
		},
		{
			file:    skill,
			heading: "## 2. Classify the cause",
			need:    []string{"awaited-delegate", "available", "API error"},
		},
		{
			file:    skill,
			heading: "### 3.5 awaited-delegate",
			need:    []string{"deliverable", "never wait for the report", "SendMessage"},
		},
		{
			file:    skill,
			heading: "## Boundaries (hard)",
			forbid:  []string{"no scheduler change"},
		},
		{
			file:    dispatch,
			heading: "### Lane waits are explicit, and stalls are watched",
			need:    []string{"CronCreate", "recurring", "disk evidence"},
		},
	}
	for _, n := range needs {
		raw, err := fs.ReadFile(fsys, n.file)
		if err != nil {
			t.Fatalf("read %s: %v", n.file, err)
		}
		body, ok := sectionOf(string(raw), n.heading)
		if !ok {
			t.Errorf("%s: section %q not found", n.file, n.heading)
			continue
		}
		for _, want := range n.need {
			if !strings.Contains(body, want) {
				t.Errorf("%s § %q lacks %q", n.file, n.heading, want)
			}
		}
		for _, bad := range n.forbid {
			if strings.Contains(body, bad) {
				t.Errorf("%s § %q still carries the retired phrase %q", n.file, n.heading, bad)
			}
		}
	}
}
