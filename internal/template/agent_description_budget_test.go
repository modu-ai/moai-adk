// Agent-description budget guard (SPEC-PREFIX-DIET-001 REQ-PFD-010/011, card t1450).
//
// Every agent description rides the session prefix of every MoAI session, so the sum of the
// `description:` blocks of the deployed agents and the largest single block are capped. The caps
// only move down: they are below the anchor values (sum 11,155 and largest 2,182 UTF-16 units,
// anchor 5d5ff1aae).
package template

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

const (
	// agentDescTotalBudget caps the summed UTF-16 length of all agent description blocks.
	agentDescTotalBudget = 10460
	// agentDescPerAgentCap caps a single agent description block.
	agentDescPerAgentCap = 1815

	agentDescAnchorTotal = 11155
	agentDescAnchorMax   = 2182
)

// dietDescriptionBlock returns the text after `description:` up to the next top-level frontmatter
// key, including the trailing newline (the same block the plan.md section C extractor returns).
func dietDescriptionBlock(fm string) (string, bool) {
	const key = "\ndescription:"
	i := strings.Index(fm, key)
	if i < 0 {
		return "", false
	}
	rest := fm[i+len(key):]
	off := 0
	first := true
	for off <= len(rest) {
		nl := strings.IndexByte(rest[off:], '\n')
		var line string
		var next int
		if nl < 0 {
			line, next = rest[off:], len(rest)+1
		} else {
			line, next = rest[off:off+nl], off+nl+1
		}
		if !first && line != "" {
			if r, _ := utf8.DecodeRuneInString(line); !unicode.IsSpace(r) {
				return rest[:off], true
			}
		}
		off, first = next, false
	}
	return rest, true
}

// dietAgentDescriptionSizes maps each agent file name to the UTF-16 length of its description block.
func dietAgentDescriptionSizes(files map[string]string) (map[string]int, []string) {
	sizes := map[string]int{}
	var problems []string
	for name, text := range files {
		head, _, err := dietSplitFrontmatter(text)
		if err != nil {
			problems = append(problems, fmt.Sprintf("AGENT_PARSE %s: %v", name, err))
			continue
		}
		block, ok := dietDescriptionBlock(strings.TrimSuffix(head[3:], "---\n"))
		if !ok {
			problems = append(problems, fmt.Sprintf("AGENT_NO_DESCRIPTION %s", name))
			continue
		}
		sizes[name] = dietUTF16Len(block)
	}
	return sizes, problems
}

// dietCheckAgentDescriptions names the offending file and size for every cap that is exceeded.
func dietCheckAgentDescriptions(files map[string]string, total, perAgent int) []string {
	sizes, v := dietAgentDescriptionSizes(files)
	names := make([]string, 0, len(sizes))
	sum := 0
	for n, s := range sizes {
		names = append(names, n)
		sum += s
	}
	sort.Strings(names)
	for _, n := range names {
		if sizes[n] > perAgent {
			v = append(v, fmt.Sprintf("AGENT_OVER_CAP %s description is %d UTF-16 units, over the per-agent cap %d", n, sizes[n], perAgent))
		}
	}
	if sum > total {
		v = append(v, fmt.Sprintf("AGENT_OVER_TOTAL the description blocks total %d UTF-16 units, over the budget %d", sum, total))
	}
	return v
}

func dietReadAgents(t *testing.T) map[string]string {
	t.Helper()
	dir := filepath.Join("templates", ".claude", "agents", "moai")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read agents dir: %v", err)
	}
	out := map[string]string{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read agent %s: %v", e.Name(), err)
		}
		out[e.Name()] = string(raw)
	}
	if len(out) == 0 {
		t.Fatal("no agent definitions found")
	}
	return out
}

func TestAgentDescriptionBudget(t *testing.T) {
	agents := dietReadAgents(t)
	sizes, problems := dietAgentDescriptionSizes(agents)
	for _, p := range problems {
		t.Error(p)
	}
	names := make([]string, 0, len(sizes))
	sum, largest := 0, 0
	for n, s := range sizes {
		names = append(names, n)
		sum += s
		largest = max(largest, s)
	}
	sort.Strings(names)
	for _, n := range names {
		t.Logf("agent-description %s %d", n, sizes[n])
	}
	t.Logf("agent-description-total=%d largest=%d agents=%d", sum, largest, len(sizes))

	if agentDescTotalBudget >= agentDescAnchorTotal || agentDescPerAgentCap >= agentDescAnchorMax {
		t.Errorf("caps (%d, %d) must stay below the anchor values (%d, %d)", agentDescTotalBudget, agentDescPerAgentCap, agentDescAnchorTotal, agentDescAnchorMax)
	}
	for _, line := range dietCheckAgentDescriptions(agents, agentDescTotalBudget, agentDescPerAgentCap) {
		t.Error(line)
	}

	t.Run("oversized_description_names_file_and_size", func(t *testing.T) {
		mutated := map[string]string{}
		for k, v := range agents {
			mutated[k] = v
		}
		const victim = "manager-develop.md"
		text, ok := mutated[victim]
		if !ok {
			t.Fatalf("%s missing", victim)
		}
		pad := strings.Repeat("padding ", agentDescPerAgentCap/8+1)
		mutated[victim] = strings.Replace(text, "\ntools:", "\n  "+pad+"\ntools:", 1)
		got := dietCheckAgentDescriptions(mutated, agentDescTotalBudget, agentDescPerAgentCap)
		if !slices.ContainsFunc(got, func(l string) bool {
			return strings.HasPrefix(l, "AGENT_OVER_CAP "+victim) && strings.Contains(l, "UTF-16 units")
		}) {
			t.Fatalf("an oversized description was not caught per agent; got %v", got)
		}
		if !slices.ContainsFunc(got, func(l string) bool { return strings.HasPrefix(l, "AGENT_OVER_TOTAL") }) {
			t.Fatalf("an oversized description was not caught in the total; got %v", got)
		}
	})
}
