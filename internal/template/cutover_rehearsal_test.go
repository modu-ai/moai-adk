// cutover_rehearsal_test.go: AC-GFD-018 (SPEC-GITHUB-FLOW-DEFAULT-001 M6, design
// D-8) for scripts/cutover-rehearsal.sh, which rehearses the develop/main
// convergence and its rollback in a scratch clone with a scratch bare "origin".
//
// The fixture is a synthetic source repository with the divergence shape the real
// repository has: a develop with many commits and a main with one commit develop
// lacks. The source's own `origin` points at an unresolvable https URL on
// purpose: if the script ever contacted it, the run would fail. Independent
// oracles (git merge-tree in the source) judge the printed identities, so the
// test does not merely read the script's own claims back.
package template_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// cvoSource builds the divergent source repository and returns its path plus the
// develop and main tips (also published as refs/remotes/origin/{develop,main}).
func cvoSource(t *testing.T) (dir, develop, main string) {
	t.Helper()
	dir = filepath.Join(t.TempDir(), "source")
	rlsGit(t, filepath.Dir(dir), "init", "-q", "-b", "develop", dir)
	rlsGit(t, dir, "remote", "add", "origin", "https://invalid.invalid/never/contacted.git")
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	commit := func(msg string) {
		rlsGit(t, dir, "add", "-A")
		rlsGit(t, dir, "commit", "-q", "-m", msg)
	}
	write("common.txt", "common\n")
	commit("base")
	write("f1.txt", "1\n")
	commit("c1")
	fork := rlsGit(t, dir, "rev-parse", "HEAD")
	for i := 2; i <= 25; i++ {
		write(fmt.Sprintf("f%d.txt", i), fmt.Sprintf("%d\n", i))
		commit(fmt.Sprintf("develop commit %d", i))
	}
	develop = rlsGit(t, dir, "rev-parse", "HEAD")
	rlsGit(t, dir, "checkout", "-q", "-b", "main-side", fork)
	write("main-only.txt", "only on main\n")
	commit("main-only commit")
	main = rlsGit(t, dir, "rev-parse", "HEAD")
	rlsGit(t, dir, "checkout", "-q", "develop")
	rlsGit(t, dir, "branch", "-q", "-D", "main-side")
	rlsGit(t, dir, "update-ref", "refs/remotes/origin/develop", develop)
	rlsGit(t, dir, "update-ref", "refs/remotes/origin/main", main)
	return dir, develop, main
}

// cvoRehearse runs the rehearsal and asserts the poison stayed untouched.
func cvoRehearse(t *testing.T, args ...string) rlsResult {
	t.Helper()
	poison := cvoNewPoison(t)
	res := rlsRun(t, t.TempDir(), poison.env(), "bash", append([]string{cvoScript(t, cvoRehearsalRel)}, args...)...)
	poison.assertUntouched(t)
	return res
}

// cvoEvidence extracts `key=value` pairs from the evidence line that starts with
// the given label.
func cvoEvidence(t *testing.T, out, label string) map[string]string {
	t.Helper()
	for _, line := range strings.Split(rlsANSIRe.ReplaceAllString(out, ""), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, label+" ") {
			continue
		}
		kv := map[string]string{}
		for _, f := range strings.Fields(strings.TrimPrefix(line, label+" ")) {
			if k, v, ok := strings.Cut(f, "="); ok {
				kv[k] = v
			}
		}
		return kv
	}
	t.Errorf("evidence line %q is missing\n--- output ---\n%s", label, out)
	return map[string]string{}
}

func cvoWant(t *testing.T, kv map[string]string, key, want string) {
	t.Helper()
	if got := kv[key]; got != want {
		t.Errorf("%s = %q, want %q", key, got, want)
	}
}

var cvoLocalRemoteRe = regexp.MustCompile(`^\S+\s+(/\S+)\s+\((fetch|push)\)$`)

// TestCutoverRehearsal is AC-GFD-018: absorb, converge with a merge commit,
// tree identity, ancestry, the revert's tree, and the ref-restoring rollback.
func TestCutoverRehearsal(t *testing.T) {
	rlsRequireTools(t)

	t.Run("convergence_tree_identity_ancestry_and_rollback", func(t *testing.T) {
		src, develop, main := cvoSource(t)
		// Independent oracles, computed in the source and never in the script.
		absorbTree := rlsGit(t, src, "merge-tree", "--write-tree", develop, main)
		mainTree := rlsGit(t, src, "rev-parse", main+"^{tree}")
		developTree := rlsGit(t, src, "rev-parse", develop+"^{tree}")
		if absorbTree == developTree || mainTree == developTree {
			t.Fatalf("the fixture must diverge: absorb=%s develop=%s main=%s", absorbTree, developTree, mainTree)
		}

		work := filepath.Join(t.TempDir(), "m6")
		res := cvoRehearse(t, "--source", src, "--workdir", work)
		out := rlsNorm(res.out)
		t.Logf("exit=%d\n%s", res.exit, res.out)
		if res.exit != 0 {
			t.Fatalf("exit code = %d, want 0\n%s", res.exit, out)
		}
		rlsMustContain(t, out, "REHEARSAL-RESULT PASS")

		start := cvoEvidence(t, res.out, "start")
		cvoWant(t, start, "develop", develop)
		cvoWant(t, start, "main", main)
		cvoWant(t, start, "develop-tree", developTree)
		cvoWant(t, start, "main-tree", mainTree)

		absorb := cvoEvidence(t, res.out, "absorb")
		cvoWant(t, absorb, "parents", develop+","+main) // a merge commit: develop first, main second
		cvoWant(t, absorb, "tree", absorbTree)

		conv := cvoEvidence(t, res.out, "convergence")
		absorbSHA := absorb["develop"]
		cvoWant(t, conv, "parents", main+","+absorbSHA) // a merge commit, never a squash
		cvoWant(t, conv, "tree", absorbTree)

		ident := cvoEvidence(t, res.out, "tree-identity")
		cvoWant(t, ident, "main-tree", absorbTree)
		cvoWant(t, ident, "develop-tree", absorbTree)
		cvoWant(t, ident, "equal", "yes")

		anc := cvoEvidence(t, res.out, "ancestry")
		cvoWant(t, anc, "develop-tip-is-ancestor-of-main", "yes")

		rev := cvoEvidence(t, res.out, "revert")
		cvoWant(t, rev, "tree", mainTree)
		cvoWant(t, rev, "start-main-tree", mainTree)
		cvoWant(t, rev, "equal", "yes")

		rb := cvoEvidence(t, res.out, "rollback")
		cvoWant(t, rb, "main", main)
		cvoWant(t, rb, "develop", develop)
		cvoWant(t, rb, "origin-main", main)
		cvoWant(t, rb, "origin-develop", develop)
		cvoWant(t, rb, "identical-to-start", "yes")

		// The scratch repositories, read after the run: refs are back at the
		// starting values, locally and in the scratch bare origin.
		w := filepath.Join(work, "work")
		o := filepath.Join(work, "origin.git")
		cvoWant(t, map[string]string{"main": rlsGit(t, w, "rev-parse", "main")}, "main", main)
		cvoWant(t, map[string]string{"develop": rlsGit(t, w, "rev-parse", "develop")}, "develop", develop)
		cvoWant(t, map[string]string{"main": rlsGit(t, o, "rev-parse", "refs/heads/main")}, "main", main)
		cvoWant(t, map[string]string{"develop": rlsGit(t, o, "rev-parse", "refs/heads/develop")}, "develop", develop)

		// The rollback was read from a recorded file, not from shell memory.
		rec, err := os.ReadFile(filepath.Join(work, "pre-merge-refs.txt"))
		if err != nil {
			t.Fatalf("the pre-merge refs record is missing: %v", err)
		}
		rlsMustContain(t, string(rec), "main "+main, "develop "+develop)
	})

	t.Run("remotes_are_local_paths_and_the_source_is_untouched", func(t *testing.T) {
		src, _, _ := cvoSource(t)
		before := rlsGit(t, src, "for-each-ref") + "|" + rlsGit(t, src, "config", "--get-regexp", "^remote\\.")
		work := filepath.Join(t.TempDir(), "m6")
		res := cvoRehearse(t, "--source", src, "--workdir", work)
		if res.exit != 0 {
			t.Fatalf("exit code = %d\n%s", res.exit, rlsNorm(res.out))
		}
		after := rlsGit(t, src, "for-each-ref") + "|" + rlsGit(t, src, "config", "--get-regexp", "^remote\\.")
		if before != after {
			t.Errorf("the rehearsal changed the source repository:\nbefore=%s\nafter=%s", before, after)
		}
		remotes := rlsGit(t, filepath.Join(work, "work"), "remote", "-v")
		if strings.Contains(remotes, "://") || strings.Contains(remotes, "invalid.invalid") {
			t.Errorf("the scratch clone has a non-local remote:\n%s", remotes)
		}
		for _, line := range strings.Split(strings.TrimSpace(remotes), "\n") {
			if m := cvoLocalRemoteRe.FindStringSubmatch(strings.TrimSpace(line)); m == nil {
				t.Errorf("remote line is not `<name> <absolute local path> (fetch|push)`: %q", line)
			}
		}
		// The same `git remote -v` is part of the printed evidence.
		rlsMustContain(t, rlsNorm(res.out), "remotes:")
	})
}

// TestCutoverRehearsalRefusals pins what the script must refuse.
func TestCutoverRehearsalRefusals(t *testing.T) {
	rlsRequireTools(t)

	t.Run("a_non_local_remote_is_refused", func(t *testing.T) {
		repo := filepath.Join(t.TempDir(), "r")
		rlsGit(t, filepath.Dir(repo), "init", "-q", "-b", "develop", repo)
		rlsGit(t, repo, "remote", "add", "origin", "https://invalid.invalid/x.git")
		res := cvoRehearse(t, "--check-remotes-only", repo)
		if res.exit == 0 {
			t.Errorf("a repository with an https remote must be refused, got exit 0")
		}
		rlsMustContain(t, rlsNorm(res.out), "non-local remote", "origin")

		ssh := filepath.Join(t.TempDir(), "s")
		rlsGit(t, filepath.Dir(ssh), "init", "-q", "-b", "develop", ssh)
		rlsGit(t, ssh, "remote", "add", "origin", "git@github.com:x/y.git")
		if res := cvoRehearse(t, "--check-remotes-only", ssh); res.exit == 0 {
			t.Errorf("an scp-style ssh remote must be refused, got exit 0")
		}
	})

	t.Run("local_path_remotes_are_accepted", func(t *testing.T) {
		base := t.TempDir()
		bare := filepath.Join(base, "bare.git")
		rlsGit(t, base, "init", "-q", "--bare", bare)
		repo := filepath.Join(base, "r")
		rlsGit(t, base, "init", "-q", "-b", "develop", repo)
		rlsGit(t, repo, "remote", "add", "origin", bare)
		res := cvoRehearse(t, "--check-remotes-only", repo)
		if res.exit != 0 {
			t.Errorf("local-only remotes must be accepted, got exit %d\n%s", res.exit, rlsNorm(res.out))
		}
	})

	t.Run("a_non_empty_workdir_is_refused", func(t *testing.T) {
		src, _, _ := cvoSource(t)
		work := t.TempDir()
		if err := os.WriteFile(filepath.Join(work, "keep.txt"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		res := cvoRehearse(t, "--source", src, "--workdir", work)
		if res.exit != 2 {
			t.Errorf("a non-empty workdir must be a usage error (exit 2), got %d\n%s", res.exit, rlsNorm(res.out))
		}
	})

	t.Run("a_workdir_inside_the_source_is_refused", func(t *testing.T) {
		src, _, _ := cvoSource(t)
		res := cvoRehearse(t, "--source", src, "--workdir", filepath.Join(src, "m6"))
		if res.exit != 2 {
			t.Errorf("a workdir inside the source must be refused (exit 2), got %d\n%s", res.exit, rlsNorm(res.out))
		}
	})

	t.Run("a_source_without_the_remote_tracking_refs_is_refused", func(t *testing.T) {
		src, _, _ := cvoSource(t)
		rlsGit(t, src, "update-ref", "-d", "refs/remotes/origin/main")
		res := cvoRehearse(t, "--source", src, "--workdir", filepath.Join(t.TempDir(), "m6"))
		if res.exit != 2 {
			t.Errorf("a source lacking origin/main must be a usage error (exit 2), got %d\n%s", res.exit, rlsNorm(res.out))
		}
		rlsMustContain(t, rlsNorm(res.out), "main")
	})

	t.Run("usage_errors_are_exit_2", func(t *testing.T) {
		if res := cvoRehearse(t, "--no-such-flag"); res.exit != 2 {
			t.Errorf("unknown flag exit = %d, want 2", res.exit)
		}
		if res := cvoRehearse(t); res.exit != 2 {
			t.Errorf("no arguments exit = %d, want 2", res.exit)
		}
	})
}

// TestCutoverRehearsalRealHistory repeats the convergence on this repository's
// own history (AC-GFD-018's pinned tips when present). It needs the
// remote-tracking refs and a full-size clone, so it runs only when asked:
// CUTOVER_REHEARSAL_REAL=1. A skip here is a visible Gap, never a pass.
func TestCutoverRehearsalRealHistory(t *testing.T) {
	rlsRequireTools(t)
	if os.Getenv("CUTOVER_REHEARSAL_REAL") != "1" {
		t.Skip("set CUTOVER_REHEARSAL_REAL=1 to rehearse on this repository's history (large scratch clone)")
	}
	root := findProjectRootForMirrorTest(t)
	work := filepath.Join(t.TempDir(), "m6")
	res := cvoRehearse(t, "--source", root, "--workdir", work)
	t.Logf("exit=%d\n%s", res.exit, res.out)
	if res.exit != 0 {
		t.Fatalf("exit code = %d", res.exit)
	}
	ident := cvoEvidence(t, res.out, "tree-identity")
	cvoWant(t, ident, "equal", "yes")
}
