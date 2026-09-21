package merge

// mcp_retained_advisory_test.go pins the retained-key advisory the .mcp.json
// merge prints (card t1040).
//
// The retention itself is already correct and is characterized elsewhere
// (mcp_snapshot_base_test.go, K3): a key the new template no longer carries
// survives in the user's file. What the merged OUTPUT cannot say is WHY it
// survived — a key absent from the base because the user added it and a key
// the template retired look identical afterwards. Only the base side separates
// them, so the advisory is the merge's own report of that distinction.
//
// The case is reachable only under a canonical snapshot base: a DERIVED base is
// pruned to keys the new template still carries, so no key can be in the base
// and absent from the template at once.

import (
	"strings"
	"testing"
)

// advisoryFor renders the advisory line the merge is expected to print for a
// retained key path. The wording is spelled out literally rather than taken
// from a production constant so the test pins the text, matching the sibling
// YAML path's already-shipped line.
func advisoryFor(keyPath string) string {
	return `advisory: retained key "` + keyPath + `" absent from new template (preserved from user config)`
}

// A key the new render retired, retained from the user's config under a
// canonical snapshot base, is named in the merge log.
func TestMergeUserFiles_MCPRetainedKey_AdvisoryNamesTheKey(t *testing.T) {
	root := newSnapshotProject(t)
	writeRel(t, root, mcpCanonicalSnapshotRel, mcpPriorRender)

	doc, log := mergeMCPJSON(t, root, mcpUserCopy, mcpNewRender)

	// Premise assertion: the advisory is about a key that actually survived.
	// Without this the test could pass over an empty merge.
	if server(doc, "gone") == nil {
		t.Fatalf("mcpServers.gone was not retained — the advisory has no subject: %v", doc)
	}

	want := advisoryFor("mcpServers.gone")
	if !strings.Contains(log, want) {
		t.Errorf("merge log missing retained-key advisory\nwant substring: %s\ngot log:\n%s", want, log)
	}
}

// Control for the assertion above: when the new render retires nothing, no
// advisory is printed. The user's own key (mcpServers.mine) is absent from the
// base too, so this case is precisely the one the merged output alone cannot
// tell apart from a retirement — if the advisory fired here it would be
// reporting a user addition as a template retirement.
//
// The positive control for the absence assertion is the test above: the same
// substring check does fire on a genuine retirement, so a zero here is a
// measured zero rather than a check that can never match.
func TestMergeUserFiles_MCPRetainedKey_NoAdvisoryWhenNothingRetired(t *testing.T) {
	root := newSnapshotProject(t)

	// A base and a user copy that differ only by the user's own server, and a
	// render that only ADDS a key — nothing is retired anywhere.
	const base = `{"mcpServers":{"context7":{"command":"npx"}}}`
	const user = `{"mcpServers":{"context7":{"command":"npx"},"mine":{"command":"my-own-server"}}}`
	const render = `{"mcpServers":{"context7":{"command":"npx"},"moai":{"command":"moai"}}}`

	writeRel(t, root, mcpCanonicalSnapshotRel, base)
	doc, log := mergeMCPJSON(t, root, user, render)

	if server(doc, "mine") == nil {
		t.Fatalf("mcpServers.mine lost — the control does not model a surviving user key: %v", doc)
	}
	if strings.Contains(log, "advisory: retained key") {
		t.Errorf("advisory printed although nothing was retired\nlog:\n%s", log)
	}
}

// The advisory is an OUTPUT-only addition: the merged file must not gain, lose,
// or reorder a byte. Both goldens below were captured from this tree BEFORE the
// advisory existed (.moai/reports/t1040/pre-change-golden.log), so this test is
// a genuine before/after comparison rather than a restatement of current
// behaviour.
func TestMergeUserFiles_MCPRetainedKey_MergedBytesUnchanged(t *testing.T) {
	const retiredCaseGolden = `{
  "mcpServers": {
    "context7": {
      "args": [
        "-y",
        "@upstash/context7-mcp@latest"
      ],
      "command": "npx"
    },
    "gone": {
      "command": "retired"
    },
    "mine": {
      "command": "my-own-server"
    },
    "moai": {
      "args": [
        "mcp-server"
      ],
      "command": "moai"
    }
  }
}`

	const plainCaseGolden = `{
  "mcpServers": {
    "context7": {
      "command": "npx"
    },
    "mine": {
      "command": "my-own-server"
    },
    "moai": {
      "command": "moai"
    }
  }
}`

	t.Run("retired key case", func(t *testing.T) {
		root := newSnapshotProject(t)
		writeRel(t, root, mcpCanonicalSnapshotRel, mcpPriorRender)
		mergeMCPJSON(t, root, mcpUserCopy, mcpNewRender)

		if got := string(readRel(t, root, mcpRel)); got != retiredCaseGolden {
			t.Errorf("merged .mcp.json bytes changed\ngot:\n%s\nwant:\n%s", got, retiredCaseGolden)
		}
	})

	t.Run("plain case", func(t *testing.T) {
		root := newSnapshotProject(t)
		const base = `{"mcpServers":{"context7":{"command":"npx"}}}`
		const user = `{"mcpServers":{"context7":{"command":"npx"},"mine":{"command":"my-own-server"}}}`
		const render = `{"mcpServers":{"context7":{"command":"npx"},"moai":{"command":"moai"}}}`

		writeRel(t, root, mcpCanonicalSnapshotRel, base)
		mergeMCPJSON(t, root, user, render)

		if got := string(readRel(t, root, mcpRel)); got != plainCaseGolden {
			t.Errorf("merged .mcp.json bytes changed\ngot:\n%s\nwant:\n%s", got, plainCaseGolden)
		}
	})
}

// The advisory is scoped to .mcp.json. The .claude/settings.json path shares
// the same engine and the same snapshot-base mechanism, and its output must
// stay byte-identical — the operator has not authorized changing that surface.
func TestMergeUserFiles_SettingsRetainedKey_StaysSilent(t *testing.T) {
	root := newSnapshotProject(t)

	const base = `{"model":"opus","retired":{"a":1}}`
	const user = `{"model":"opus","retired":{"a":1},"mine":true}`
	const render = `{"model":"opus","added":true}`

	writeRel(t, root, canonicalSnapshotRel, base)
	doc, log := mergeSettings(t, root, user, render)

	// Premise: settings.json really did retain a retired key, so silence here
	// is a scoping decision rather than an absent case.
	if leaf(doc, "retired", "a") == nil {
		t.Fatalf("settings.json did not retain the retired key — silence proves nothing: %v", doc)
	}
	if strings.Contains(log, "advisory: retained key") {
		t.Errorf("settings.json emitted a retained-key advisory\nlog:\n%s", log)
	}
}
