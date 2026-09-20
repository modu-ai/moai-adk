package statusline

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// twoTreeFixture builds a real repository and a linked worktree of it, each
// carrying its own .moai/config/sections/statusline.yaml. The divergent case
// t957 repairs needs BOTH trees to be real git trees: the config root used to
// come from a cwd walk-up (which stops at the worktree) while the board root
// came from the state anchor (which resolves back to the primary), so one
// render read two roots. A fixture with a single tree cannot show that.
//
// Returns the primary checkout path and the linked worktree path, both
// canonical (EvalSymlinks) — on macOS t.TempDir() hands back a /var path that
// git reports as /private/var, and an uncanonicalised expectation compares two
// spellings of the same directory.
func twoTreeFixture(t *testing.T) (primary, worktree string) {
	t.Helper()

	base := t.TempDir()
	primary = filepath.Join(base, "primary")
	if err := os.MkdirAll(primary, 0o755); err != nil {
		t.Fatalf("mkdir primary: %v", err)
	}

	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
		}
	}

	run(primary, "init", "--initial-branch=main")
	if err := os.WriteFile(filepath.Join(primary, "seed.txt"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	run(primary, "add", "seed.txt")
	run(primary, "commit", "-m", "seed")

	worktree = filepath.Join(base, "wt")
	run(primary, "worktree", "add", "-b", "wt-branch", worktree)

	writeStatuslineYAML(t, primary, "github")
	writeStatuslineYAML(t, worktree, "none")

	primary = canonical(t, primary)
	worktree = canonical(t, worktree)
	return primary, worktree
}

func canonical(t *testing.T, p string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("EvalSymlinks(%s): %v", p, err)
	}
	return resolved
}

func writeStatuslineYAML(t *testing.T, root, forge string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	body := "statusline:\n  forge: \"" + forge + "\"\n  theme: " + filepath.Base(root) + "\n"
	if err := os.WriteFile(filepath.Join(dir, "statusline.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write statusline.yaml: %v", err)
	}
}

// stdinJSON renders a statusline stdin payload carrying the given workspace
// and worktree fields. Empty values are omitted so a case can exercise the
// anchor chain's later arms.
func stdinJSON(t *testing.T, projectDir, currentDir, originalCwd, cwd string) []byte {
	t.Helper()
	payload := map[string]any{"session_id": "t957"}
	if projectDir != "" || currentDir != "" {
		ws := map[string]any{}
		if projectDir != "" {
			ws["project_dir"] = projectDir
		}
		if currentDir != "" {
			ws["current_dir"] = currentDir
		}
		payload["workspace"] = ws
	}
	if originalCwd != "" {
		payload["worktree"] = map[string]any{"original_cwd": originalCwd}
	}
	if cwd != "" {
		payload["cwd"] = cwd
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal stdin: %v", err)
	}
	return raw
}

// anchorOf is the board-root side of the render: what forgeOverride and every
// other state consumer already resolve to.
func anchorOf(t *testing.T, raw []byte) string {
	t.Helper()
	var input StdinData
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &input); err != nil {
			t.Fatalf("unmarshal stdin: %v", err)
		}
	}
	return resolveStateAnchor(&input)
}

// TestT957_PremiseTwoTreesDisagree pins the premise the repair rests on: with
// the session sitting in the worktree, the cwd walk-up root (the fallback,
// which is what internal/cli's findProjectRoot returns) and the state anchor
// name DIFFERENT trees. If this ever stops holding, the divergence this card
// repairs no longer exists and the acceptance test below is vacuous.
func TestT957_PremiseTwoTreesDisagree(t *testing.T) {
	primary, worktree := twoTreeFixture(t)

	raw := stdinJSON(t, "", "", "", worktree)
	anchor := anchorOf(t, raw)

	if anchor != primary {
		t.Fatalf("anchor = %q, want the primary checkout %q", anchor, primary)
	}
	if anchor == worktree {
		t.Fatalf("anchor and the cwd walk-up root agree (%q) — the divergent case is gone", worktree)
	}
}

// TestT957_ConfigRootFollowsTheAnchor is the acceptance condition: whenever the
// anchor names a tree, the root a render reads statusline.yaml from is THAT
// tree — not the one a cwd walk-up would stop at. One render, one root.
func TestT957_ConfigRootFollowsTheAnchor(t *testing.T) {
	primary, worktree := twoTreeFixture(t)

	cases := []struct {
		name string
		raw  []byte
	}{
		{"project_dir names the primary", stdinJSON(t, primary, worktree, "", worktree)},
		{"original_cwd names the primary", stdinJSON(t, "", worktree, primary, worktree)},
		{"no project fields, cwd in the worktree", stdinJSON(t, "", "", "", worktree)},
		{"current_dir in the worktree", stdinJSON(t, "", worktree, "", "")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			anchor := anchorOf(t, tc.raw)
			if anchor == "" {
				t.Fatalf("fixture produced no anchor — case cannot test alignment")
			}
			got, replay := ResolveConfigRoot(bytes.NewReader(tc.raw), worktree)
			if got != anchor {
				t.Errorf("ResolveConfigRoot = %q, want the anchor %q (fallback was %q)", got, anchor, worktree)
			}
			// The seam sits in front of Build, which needs the same bytes
			// again. A replay that lost or reordered them would render a
			// statusline with no session data at all.
			back, err := io.ReadAll(replay)
			if err != nil {
				t.Fatalf("read replay: %v", err)
			}
			if !bytes.Equal(back, tc.raw) {
				t.Errorf("replay = %s, want the payload verbatim %s", back, tc.raw)
			}
		})
	}
}

// TestT957_FallbackWhenNoAnchor is the counter-case that bounds the repair.
// Outside a git repository the anchor is empty by REQ-SA-003 and every state
// consumer SKIPS rather than reading a different tree — so there is no second
// root to diverge from, and the config read keeps working off the caller's own
// root. Without this the repair would silently withdraw statusline.yaml from
// every non-git .moai project.
func TestT957_FallbackWhenNoAnchor(t *testing.T) {
	plain := canonical(t, t.TempDir())
	writeStatuslineYAML(t, plain, "github")

	raw := stdinJSON(t, "", "", "", plain)
	if anchor := anchorOf(t, raw); anchor != "" {
		t.Fatalf("fixture is inside a git repository (anchor %q) — the no-anchor case is not being exercised", anchor)
	}

	if got, _ := ResolveConfigRoot(bytes.NewReader(raw), plain); got != plain {
		t.Errorf("ResolveConfigRoot = %q, want the fallback %q", got, plain)
	}
}

// TestT957_NoStdinKeepsTodaysRoot measures the path the dispatch flagged as
// the repair's cost: a render that reads no stdin at all (a terminal, where
// readStdinWithTimeout hands Build an empty reader).
//
// It does NOT change. stateanchor.Resolve deliberately refuses to consult the
// process working directory — "no directory context, no anchor" — so a payload
// carrying no location fields resolves to "", and the config read falls back
// to exactly the root it uses today. The behaviour change is therefore
// confined to the divergent case this card repairs: stdin present, anchor
// non-empty, and the anchor naming a different tree than the cwd walk-up.
func TestT957_NoStdinKeepsTodaysRoot(t *testing.T) {
	_, worktree := twoTreeFixture(t)

	if anchor := anchorOf(t, nil); anchor != "" {
		t.Fatalf("anchor = %q for an empty payload, want empty — the no-stdin premise does not hold", anchor)
	}

	// Both shapes the render can present: no reader at all, and a reader that
	// yields nothing (what readStdinWithTimeout hands Build on a terminal).
	if got, _ := ResolveConfigRoot(nil, worktree); got != worktree {
		t.Errorf("ResolveConfigRoot(nil) = %q, want today's root %q unchanged", got, worktree)
	}
	if got, _ := ResolveConfigRoot(io.MultiReader(), worktree); got != worktree {
		t.Errorf("ResolveConfigRoot(empty reader) = %q, want today's root %q unchanged", got, worktree)
	}
}
