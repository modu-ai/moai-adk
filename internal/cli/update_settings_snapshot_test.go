package cli

// update_settings_snapshot_test.go pins where the real flows call the
// .claude/settings.json base-snapshot lifecycle
// (SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001, card t656, plan.md M4):
//
//	AC-USB-005 — the staging copy is the pre-merge render and the canonical
//	             base is untouched until the merge has run (pre-merge hook);
//	AC-USB-007 — staging, promotion, and the leftover judgement run at the
//	             right point of clean-reinstall, template sync (both backup
//	             branches), runUpdate (abort and version-skip), and runInit;
//	AC-USB-008 — a staging or promotion failure never blocks a flow.
//
// Isolation: every test works under t.TempDir(), injects the home through
// homeSeamSpy (never t.Setenv("HOME")), and replaces package-level seams, so
// none of these tests may call t.Parallel(). The runUpdate and runInit cells
// chdir into their project because both flows resolve the project from the
// working directory.

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

const (
	snapLiveRel      = ".claude/settings.json"
	snapCanonicalRel = ".moai/cache/template-snapshot/claude/settings.json"
	snapPendingRel   = ".moai/cache/template-snapshot/claude/settings.json.pending"
	snapRetiredDeny  = "Write(./secrets/**)"
)

// settingsRenderDeployer is a deployer double that behaves like the real one
// for .claude/settings.json only: it writes render there and records the file
// as template-managed with the render hash, which is what the staging helper
// reads to decide the deploy wrote the file.
type settingsRenderDeployer struct {
	render string
	calls  int
}

func (d *settingsRenderDeployer) Deploy(_ context.Context, projectRoot string, mgr manifest.Manager, _ *template.TemplateContext) error {
	d.calls++
	path := filepath.Join(projectRoot, filepath.FromSlash(snapLiveRel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(d.render), 0o644); err != nil {
		return err
	}
	return mgr.Track(snapLiveRel, manifest.TemplateManaged, manifest.HashBytes([]byte(d.render)))
}

func (d *settingsRenderDeployer) ListTemplates() []string { return nil }

func (d *settingsRenderDeployer) ValidateAll(context.Context, *template.TemplateContext) error {
	return nil
}

func (d *settingsRenderDeployer) ExtractTemplate(string) ([]byte, error) { return nil, nil }

func snapWrite(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func snapRead(t *testing.T, root, rel string) (string, bool) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return "", false
	}
	return string(data), true
}

func snapDecode(t *testing.T, data string) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal([]byte(data), &doc); err != nil {
		t.Fatalf("decode %q: %v", data, err)
	}
	return doc
}

func snapAssertLiveKeys(t *testing.T, root string, want map[string]float64) {
	t.Helper()
	live, ok := snapRead(t, root, snapLiveRel)
	if !ok {
		t.Fatalf("live %s is absent", snapLiveRel)
	}
	doc := snapDecode(t, live)
	for key, value := range want {
		if got, ok := doc[key].(float64); !ok || got != value {
			t.Errorf("live %s = %v, want %v (doc %s)", key, doc[key], value, live)
		}
	}
}

func snapAssertCanonical(t *testing.T, root, want string) {
	t.Helper()
	got, ok := snapRead(t, root, snapCanonicalRel)
	if !ok {
		t.Errorf("canonical snapshot absent, want %s", want)
		return
	}
	if got != want {
		t.Errorf("canonical snapshot = %s, want %s", got, want)
	}
}

func snapAssertNoPending(t *testing.T, root string) {
	t.Helper()
	if _, ok := snapRead(t, root, snapPendingRel); ok {
		t.Errorf("staging copy still present after the flow")
	}
}

func snapCountPrefixed(out, prefix string) int {
	n := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			n++
		}
	}
	return n
}

// snapChdir moves into root for the duration of the test.
func snapChdir(t *testing.T, root string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir %s: %v", root, err)
	}
}

// snapUseSyncDeployer routes the template-sync deploy through d.
func snapUseSyncDeployer(t *testing.T, d template.Deployer) {
	t.Helper()
	prev := newTemplateSyncDeployer
	newTemplateSyncDeployer = func(fs.FS) template.Deployer { return d }
	t.Cleanup(func() { newTemplateSyncDeployer = prev })
}

// snapRunCleanReinstall drives the clean-reinstall flow over a v2 scenario
// with d as its deployer.
func snapRunCleanReinstall(t *testing.T, root string, d template.Deployer) (out, errOut string) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	migrate := &stubMigrateRunner{}
	if _, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
		Out:              &outBuf,
		ErrOut:           &errBuf,
		Deployer:         d,
		RunMigrateAgency: migrate.Run,
	}); err != nil {
		t.Fatalf("runCleanReinstall returned %v, want nil\nout:\n%s\nerr:\n%s", err, outBuf.String(), errBuf.String())
	}
	return outBuf.String(), errBuf.String()
}

// snapRunTemplateSync drives the template-sync flow in root.
func snapRunTemplateSync(t *testing.T, root string, d template.Deployer) (out, errOut string) {
	t.Helper()
	snapChdir(t, root)
	snapUseSyncDeployer(t, d)

	var outBuf, errBuf bytes.Buffer
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("config", false, "")
	cmd.Flags().Bool("no-hooks", true, "")
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetContext(context.Background())
	if err := runTemplateSyncWithReporter(cmd, nil, true); err != nil {
		t.Fatalf("runTemplateSyncWithReporter: %v\nout:\n%s\nerr:\n%s", err, outBuf.String(), errBuf.String())
	}
	return outBuf.String(), errBuf.String()
}

// snapNewUpdateProject builds a v3 project for runUpdate: the project marker,
// a template_version, and an empty manifest.
func snapNewUpdateProject(t *testing.T, templateVersion string) string {
	t.Helper()
	root := t.TempDir()
	snapWrite(t, root, ".moai/config/sections/system.yaml",
		"moai:\n  version: v3.0.0\n  template_version: \""+templateVersion+"\"\n")
	snapWrite(t, root, ".moai/manifest.json", `{"version":"1","files":{}}`)
	return root
}

// snapRunUpdate invokes runUpdate in root as a non-dry-run update with --yes.
// version.Version is pinned to "dev" so the binary-update step is skipped.
func snapRunUpdate(t *testing.T, root string) (out, errOut string, err error) {
	t.Helper()

	origDeps := deps
	t.Cleanup(func() { deps = origDeps })
	deps = &Dependencies{}

	origVersion := version.Version
	t.Cleanup(func() { version.Version = origVersion })
	version.Version = "dev"

	snapChdir(t, root)

	var outBuf, errBuf bytes.Buffer
	cmd := &cobra.Command{Use: "update-settings-snapshot-test"}
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetContext(context.Background())
	cmd.Flags().Bool("check", false, "")
	cmd.Flags().Bool("shell-env", false, "")
	cmd.Flags().Bool("config", false, "")
	cmd.Flags().Bool("binary", false, "")
	cmd.Flags().Bool("templates-only", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("no-hooks", true, "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().String("profile", "", "")
	err = runUpdate(cmd, nil)
	return outBuf.String(), errBuf.String(), err
}

// snapAssertSandboxHome guards against a seam that failed to take: the flows
// below must never see the operator's real home.
func snapAssertSandboxHome(t *testing.T, sentinel string) {
	t.Helper()
	got, err := userHomeDirFn()
	if err != nil || got != sentinel {
		t.Fatalf("userHomeDirFn = %q (%v), want the sandbox %q", got, err, sentinel)
	}
}

// AC-USB-005 — inside a clean-reinstall flow, at the pre-merge hook the
// canonical base still holds the prior bytes and the staging copy holds the
// render; after the flow the render is canonical while the live file carries
// the user's key, so it differs from the render.
func TestCleanReinstall_SettingsSnapshotStagedBeforeMerge(t *testing.T) {
	sentinel, _ := homeSeamSpy(t)
	snapAssertSandboxHome(t, sentinel)

	const render = `{"a":1,"permissions":{"deny":["Read(./.env)"]}}`
	root := makeScenarioA(t)
	snapWrite(t, root, snapLiveRel,
		`{"userOnly":true,"permissions":{"deny":["`+snapRetiredDeny+`","Read(./.env)"]}}`)
	snapWrite(t, root, snapCanonicalRel, `{"marker":"prior"}`)

	var seenCanonical, seenPending string
	hookCalls := 0
	prevHook := preMergeSettingsSnapshotHook
	t.Cleanup(func() { preMergeSettingsSnapshotHook = prevHook })
	preMergeSettingsSnapshotHook = func(projectRoot string) {
		hookCalls++
		seenCanonical, _ = snapRead(t, projectRoot, snapCanonicalRel)
		seenPending, _ = snapRead(t, projectRoot, snapPendingRel)
	}

	d := &settingsRenderDeployer{render: render}
	snapRunCleanReinstall(t, root, d)

	if d.calls != 1 || hookCalls != 1 {
		t.Fatalf("deploy calls = %d, hook calls = %d; want 1 and 1 (the observations below would be vacuous)", d.calls, hookCalls)
	}
	if seenCanonical != `{"marker":"prior"}` { // (i)
		t.Errorf("(i) canonical at the pre-merge hook = %q, want the prior bytes", seenCanonical)
	}
	if seenPending != render { // (ii)
		t.Errorf("(ii) staging copy at the pre-merge hook = %q, want the render %q", seenPending, render)
	}
	snapAssertCanonical(t, root, render) // (iii)
	snapAssertNoPending(t, root)
	live, _ := snapRead(t, root, snapLiveRel) // (iv)
	if doc := snapDecode(t, live); doc["userOnly"] != true || live == render {
		t.Errorf("(iv) live settings.json = %s, want the user's userOnly key and bytes different from the render", live)
	}
}

// AC-USB-007 — the lifecycle runs at the right point of every real flow.
func TestSettingsSnapshot_WriteSites(t *testing.T) {
	const (
		r1 = `{"a":1}`
		r2 = `{"a":2,"K":1}`
		r3 = `{"a":3,"K":1,"L":1}`
	)

	t.Run("clean_reinstall", func(t *testing.T) {
		sentinel, _ := homeSeamSpy(t)
		snapAssertSandboxHome(t, sentinel)
		root := makeScenarioA(t)
		snapWrite(t, root, snapCanonicalRel, r1)
		snapWrite(t, root, snapLiveRel, r1)

		snapRunCleanReinstall(t, root, &settingsRenderDeployer{render: r2})

		snapAssertLiveKeys(t, root, map[string]float64{"a": 2, "K": 1})
		snapAssertCanonical(t, root, r2)
	})

	for _, tc := range []struct {
		name       string
		withConfig bool
	}{
		{"template_sync_backup_empty", false},
		{"template_sync_backup_filled", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sentinel, _ := homeSeamSpy(t)
			snapAssertSandboxHome(t, sentinel)
			root := t.TempDir()
			snapWrite(t, root, ".moai/manifest.json", `{"version":"1","files":{}}`)
			if tc.withConfig {
				snapWrite(t, root, ".moai/config/sections/system.yaml", "moai:\n  template_version: \"0.0.0\"\n")
			}
			snapWrite(t, root, snapCanonicalRel, r1)
			snapWrite(t, root, snapLiveRel, r1)

			out, _ := snapRunTemplateSync(t, root, &settingsRenderDeployer{render: r2})

			// The two cells differ only in the Restore Settings branch taken;
			// confirm the fixture really selected the intended branch. Only the
			// configBackupPath != "" branch writes the sections snapshot.
			_, restored := snapRead(t, root, ".moai/cache/template-snapshot/sections/system.yaml")
			if restored != tc.withConfig {
				t.Fatalf("config restore branch ran = %v, want %v (the cell does not exercise its branch)\n%s", restored, tc.withConfig, out)
			}
			snapAssertLiveKeys(t, root, map[string]float64{"a": 2, "K": 1})
			snapAssertCanonical(t, root, r2)
		})
	}

	t.Run("update_leftover_abort", func(t *testing.T) {
		sentinel, _ := homeSeamSpy(t)
		snapAssertSandboxHome(t, sentinel)
		root := snapNewUpdateProject(t, "0.0.0")
		snapWrite(t, root, snapCanonicalRel, r1)
		snapWrite(t, root, snapPendingRel, r2) // an earlier flow stopped after its deploy
		snapWrite(t, root, snapLiveRel, r2)
		d := &settingsRenderDeployer{render: r3}
		snapUseSyncDeployer(t, d)

		out, errOut, err := snapRunUpdate(t, root)
		if err != nil {
			t.Fatalf("runUpdate: %v\nout:\n%s\nerr:\n%s", err, out, errOut)
		}
		if d.calls != 1 {
			t.Fatalf("deploy calls = %d, want 1 (the sync did not run)", d.calls)
		}
		snapAssertLiveKeys(t, root, map[string]float64{"a": 3, "K": 1, "L": 1})
		snapAssertCanonical(t, root, r3)
		snapAssertNoPending(t, root)
	})

	t.Run("update_leftover_version_skip", func(t *testing.T) {
		sentinel, _ := homeSeamSpy(t)
		snapAssertSandboxHome(t, sentinel)
		// template_version equals the pinned package version "dev", so the
		// sync is skipped. The leftover and the live file carry a retired v2
		// deny entry (N3-02): the deny-rule strip rewrites the live file, so a
		// judgement placed after the strip compares against the stripped file,
		// discards, and leaves r1 canonical.
		const leftover = `{"a":2,"K":1,"permissions":{"deny":["` + snapRetiredDeny + `"]}}`
		root := snapNewUpdateProject(t, "dev")
		snapWrite(t, root, snapCanonicalRel, r1)
		snapWrite(t, root, snapPendingRel, leftover)
		snapWrite(t, root, snapLiveRel, leftover)
		d := &settingsRenderDeployer{render: r3}
		snapUseSyncDeployer(t, d)

		out, errOut, err := snapRunUpdate(t, root)
		if err != nil {
			t.Fatalf("runUpdate: %v\nout:\n%s\nerr:\n%s", err, out, errOut)
		}
		if d.calls != 0 {
			t.Fatalf("deploy calls = %d, want 0 (the version-match skip did not fire)", d.calls)
		}
		if live, _ := snapRead(t, root, snapLiveRel); strings.Contains(live, snapRetiredDeny) {
			t.Fatalf("the deny-rule strip did not rewrite the live file; the cell cannot discriminate the :384 boundary\n%s", live)
		}
		snapAssertCanonical(t, root, leftover)
		snapAssertNoPending(t, root)
	})

	t.Run("init", func(t *testing.T) {
		sentinel, _ := homeSeamSpy(t)
		snapAssertSandboxHome(t, sentinel)

		// The bundle stand-in records the live file the moment it runs — the
		// deployed render — then rewrites it deterministically, whatever the
		// tier and gate environment.
		var renderAtBundle string
		bundleCalls := 0
		prevBundle := applyAutonomyTierBundleFn
		t.Cleanup(func() { applyAutonomyTierBundleFn = prevBundle })
		applyAutonomyTierBundleFn = func(_, _, projectSettingsPath, _ string) error {
			bundleCalls++
			data, err := os.ReadFile(projectSettingsPath)
			if err != nil {
				return err
			}
			renderAtBundle = string(data)
			doc := snapDecode(t, renderAtBundle)
			doc["a"] = 9
			rewritten, err := json.Marshal(doc)
			if err != nil {
				return err
			}
			return os.WriteFile(projectSettingsPath, rewritten, 0o644)
		}

		projectDir := filepath.Join(t.TempDir(), "snap-init")
		cmd := newInitTestCmd()
		var outBuf, errBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetErr(&errBuf)
		if err := runInit(cmd, []string{projectDir}); err != nil {
			t.Fatalf("runInit: %v\nstderr:\n%s", err, errBuf.String())
		}

		if bundleCalls != 1 || renderAtBundle == "" {
			t.Fatalf("bundle calls = %d, render seen = %q; the ordering check would be vacuous", bundleCalls, renderAtBundle)
		}
		snapAssertCanonical(t, projectDir, renderAtBundle)
		snapAssertNoPending(t, projectDir)
		live, _ := snapRead(t, projectDir, snapLiveRel)
		if got := snapDecode(t, live)["a"]; got != float64(9) || live == renderAtBundle {
			t.Errorf("live a = %v, want 9 and bytes different from the staged render", got)
		}
	})
}

// AC-USB-007 source-position substitute for the init leftover judgement.
// acceptance.md AC-USB-007 admits a source-position check where a behaviour
// cell cannot be built: driving runInit over an existing project with a
// leftover staging copy needs --force re-initialisation, which rewrites far
// more than settings.json. The check is tightened as the acceptance requires:
// within runInit, the judgement is an unconditional statement (no leading
// `if`/`else`) placed before the first executor.Execute( call.
func TestSettingsSnapshot_InitLeftoverJudgementPrecedesExecute(t *testing.T) {
	src, err := os.ReadFile("init.go")
	if err != nil {
		t.Fatalf("read init.go: %v", err)
	}
	body := string(src)
	start := strings.Index(body, "func runInit(")
	if start < 0 {
		t.Fatal("runInit not found in init.go")
	}
	body = body[start:]
	judge := strings.Index(body, "backup.JudgeLeftoverSettingsSnapshot(")
	execute := strings.Index(body, "executor.Execute(")
	if judge < 0 || execute < 0 {
		t.Fatalf("judgement at %d, executor.Execute at %d; both must be present in runInit", judge, execute)
	}
	if judge > execute {
		t.Errorf("the leftover judgement (offset %d) runs after executor.Execute (offset %d)", judge, execute)
	}
	lineStart := strings.LastIndex(body[:judge], "\n") + 1
	if prefix := strings.TrimSpace(body[lineStart:judge]); prefix != "" {
		t.Errorf("the leftover judgement is not an unconditional statement: %q precedes it on its line", prefix)
	}
}

// AC-USB-008 — a staging or promotion failure warns once and the flow returns
// nil. `helper` and `promote_failure` live in the backup and merge packages;
// these are the flow-level cells.
func TestSettingsSnapshot_WriteFailureDoesNotBlock(t *testing.T) {
	t.Run("clean_reinstall", func(t *testing.T) {
		sentinel, _ := homeSeamSpy(t)
		snapAssertSandboxHome(t, sentinel)
		const render = `{"a":2,"K":1}`
		const user = `{"a":1,"u":1}`

		// Control: the same flow with nowhere broken, whose merge also runs on
		// the derived base (no canonical copy exists before the merge).
		control := makeScenarioA(t)
		snapWrite(t, control, snapLiveRel, user)
		snapRunCleanReinstall(t, control, &settingsRenderDeployer{render: render})
		want, _ := snapRead(t, control, snapLiveRel)

		root := makeScenarioA(t)
		snapWrite(t, root, snapLiveRel, user)
		snapWrite(t, root, ".moai/cache/template-snapshot/claude", "a file where the claude/ directory belongs")
		out, errOut := snapRunCleanReinstall(t, root, &settingsRenderDeployer{render: render})

		if n := snapCountPrefixed(errOut, backup.SettingsSnapshotWriteFailedPrefix); n != 1 {
			t.Errorf("write-failed lines on stderr = %d, want 1:\n%s", n, errOut)
		}
		if n := snapCountPrefixed(out+errOut, "Warning: template snapshot write failed:"); n != 0 {
			t.Errorf("the sections-snapshot warning fired (%d); the planted file must not affect sections", n)
		}
		if got, _ := snapRead(t, root, snapLiveRel); got != want {
			t.Errorf("settings.json = %s, want the fallback merge result %s", got, want)
		}
	})

	// N3-06 — the leftover judgement's promotion fails at the runUpdate call
	// site; the update still returns nil and warns exactly once.
	t.Run("update_leftover_promote_failure", func(t *testing.T) {
		sentinel, _ := homeSeamSpy(t)
		snapAssertSandboxHome(t, sentinel)
		const leftover = `{"a":2,"K":1}`
		root := snapNewUpdateProject(t, "dev")
		snapWrite(t, root, snapCanonicalRel+"/squatter", "x")
		snapWrite(t, root, snapPendingRel, leftover)
		snapWrite(t, root, snapLiveRel, leftover)

		out, errOut, err := snapRunUpdate(t, root)
		if err != nil {
			t.Fatalf("runUpdate returned %v, want nil\nout:\n%s\nerr:\n%s", err, out, errOut)
		}
		if n := snapCountPrefixed(errOut, backup.SettingsSnapshotPromoteFailedPrefix); n != 1 {
			t.Errorf("promote-failed lines on stderr = %d, want 1:\n%s", n, errOut)
		}
		if info, statErr := os.Stat(filepath.Join(root, filepath.FromSlash(snapCanonicalRel))); statErr != nil || !info.IsDir() {
			t.Errorf("the squatting directory at the canonical path was disturbed (%v)", statErr)
		}
	})
}
