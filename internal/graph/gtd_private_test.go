package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/telemetry"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/worktree"
)

type fakePrivatePermissionChecker struct {
	current string
	owners  map[string]string
	modes   map[string]os.FileMode
	err     error
}

type faultPrivatePublisher struct {
	failAt, calls int
	removed       []string
}

func (f *faultPrivatePublisher) fail() error {
	f.calls++
	if f.calls == f.failAt {
		return errors.New("injected")
	}
	return nil
}
func (f *faultPrivatePublisher) PrepareDir(string) error { return f.fail() }
func (f *faultPrivatePublisher) WriteTemp(_ string, pattern string, _ []byte) (string, error) {
	if err := f.fail(); err != nil {
		return "", err
	}
	return pattern, nil
}
func (f *faultPrivatePublisher) Install(string, string) error { return f.fail() }
func (f *faultPrivatePublisher) Remove(path string)           { f.removed = append(f.removed, path) }

func (f fakePrivatePermissionChecker) CurrentUID() (string, error) { return f.current, f.err }
func (f fakePrivatePermissionChecker) OwnerUID(path string) (string, error) {
	return f.owners[path], f.err
}
func (f fakePrivatePermissionChecker) Mode(path string) (os.FileMode, error) {
	return f.modes[path], f.err
}

func TestBuildPrivateGTDProjectionPrivacy(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "todo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(dir, "backlog.db")
	edges := []PrivateGTDEdge{
		{From: "g2", To: "g3", Kind: "related_to", Sensitivity: "private"},
		{From: "g1", To: "g2", Kind: "depends_on", Sensitivity: "private"},
		{From: "g1", To: "g2", Kind: "depends_on", Sensitivity: "private"},
	}
	result, err := BuildPrivateGTDProjection(context.Background(), dbPath, 7, edges)
	if err != nil {
		t.Fatal(err)
	}
	if result.EdgeCount != 2 || !PrivateGTDProjectionFresh(result.MetaPath, 7) {
		t.Fatalf("projection = %+v", result)
	}
	for _, check := range []struct {
		path string
		perm os.FileMode
	}{{dir, 0o700}, {result.DataPath, 0o600}, {result.MetaPath, 0o600}} {
		info, err := os.Stat(check.path)
		if err != nil || info.Mode().Perm() != check.perm {
			t.Fatalf("%s mode=%v err=%v, want %v", check.path, info.Mode().Perm(), err, check.perm)
		}
	}
	if files := PrivateGTDBackupFiles(result, false); len(files) != 0 {
		t.Fatalf("default backup exposed projection: %v", files)
	}
	if files := PrivateGTDBackupFiles(result, true); len(files) != 2 {
		t.Fatalf("opt-in backup files = %v", files)
	}

	deniedDir := filepath.Join(t.TempDir(), "todo")
	_, err = BuildPrivateGTDProjection(context.Background(), filepath.Join(deniedDir, "backlog.db"), 1, []PrivateGTDEdge{{From: "x", To: "y", Kind: "related_to", Sensitivity: "unknown"}})
	if err == nil {
		t.Fatal("ambiguous privacy classification allowed")
	}
	if _, statErr := os.Stat(filepath.Join(deniedDir, "gtd-edges.jsonl")); !os.IsNotExist(statErr) {
		t.Fatalf("partial projection exists after denial: %v", statErr)
	}
	if err := RevokePrivateGTDProjection(result); err != nil {
		t.Fatal(err)
	}
	if PrivateGTDProjectionFresh(result.MetaPath, 7) {
		t.Fatal("revoked projection remains readable")
	}
}

func TestPrivateGTDProjectionFailClosedBranches(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := BuildPrivateGTDProjection(cancelled, filepath.Join(t.TempDir(), "backlog.db"), 1, nil); err == nil {
		t.Fatal("cancelled build allowed")
	}
	for _, invalid := range []struct {
		path string
		rev  int64
	}{{"relative/backlog.db", 1}, {filepath.Join(t.TempDir(), "other.db"), 1}, {filepath.Join(t.TempDir(), "backlog.db"), -1}} {
		if _, err := BuildPrivateGTDProjection(context.Background(), invalid.path, invalid.rev, nil); err == nil {
			t.Fatalf("invalid source allowed: %+v", invalid)
		}
	}
	dir := t.TempDir()
	projection, err := BuildPrivateGTDProjection(context.Background(), filepath.Join(dir, "backlog.db"), 2, []PrivateGTDEdge{{From: "a", To: "b", Kind: "depends_on", Sensitivity: "secret"}})
	if err != nil {
		t.Fatal(err)
	}
	if PrivateGTDProjectionFresh(projection.MetaPath, 3) {
		t.Fatal("stale revision accepted")
	}
	if err := os.WriteFile(projection.MetaPath, []byte("not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if PrivateGTDProjectionFresh(projection.MetaPath, 2) {
		t.Fatal("malformed metadata accepted")
	}
	projection, err = BuildPrivateGTDProjection(context.Background(), filepath.Join(dir, "backlog.db"), 2, []PrivateGTDEdge{{From: "a", To: "b", Kind: "depends_on", Sensitivity: "secret"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(projection.DataPath); err != nil {
		t.Fatal(err)
	}
	if PrivateGTDProjectionFresh(projection.MetaPath, 2) {
		t.Fatal("missing data artifact accepted")
	}
	projection, err = BuildPrivateGTDProjection(context.Background(), filepath.Join(dir, "backlog.db"), 2, []PrivateGTDEdge{{From: "a", To: "b", Kind: "depends_on", Sensitivity: "secret"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projection.DataPath, []byte("tampered\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if PrivateGTDProjectionFresh(projection.MetaPath, 2) {
		t.Fatal("tampered generation accepted")
	}
	if err := RevokePrivateGTDProjection(projection); err != nil {
		t.Fatal(err)
	}
	if err := RevokePrivateGTDProjection(projection); err != nil {
		t.Fatalf("repeat revoke: %v", err)
	}
	blocked := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blocked, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := writePrivateTemp(blocked, ".tmp-*", nil); err == nil {
		t.Fatal("temp file created below regular file")
	}
	if _, err := BuildPrivateGTDProjection(context.Background(), filepath.Join(blocked, "backlog.db"), 1, nil); err == nil {
		t.Fatal("projection created below regular file")
	}
	nonEmptyDir := filepath.Join(t.TempDir(), "non-empty")
	if err := os.MkdirAll(nonEmptyDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nonEmptyDir, "child"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := RevokePrivateGTDProjection(PrivateGTDProjection{DataPath: nonEmptyDir}); err == nil {
		t.Fatal("revoke removal error hidden")
	}
	renameBlocked := t.TempDir()
	if err := os.Mkdir(filepath.Join(renameBlocked, privateGTDDataName), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildPrivateGTDProjection(context.Background(), filepath.Join(renameBlocked, "backlog.db"), 1, nil); err == nil {
		t.Fatal("blocked data publication reported success")
	}
	metaBlocked := t.TempDir()
	if err := os.Mkdir(filepath.Join(metaBlocked, privateGTDMetaName), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildPrivateGTDProjection(context.Background(), filepath.Join(metaBlocked, "backlog.db"), 1, nil); err == nil {
		t.Fatal("blocked metadata publication reported success")
	}
	if _, err := os.Stat(filepath.Join(metaBlocked, privateGTDDataName)); !os.IsNotExist(err) {
		t.Fatalf("unpaired data survived metadata publication failure: %v", err)
	}
	target := t.TempDir()
	linked := filepath.Join(t.TempDir(), "todo")
	if err := os.Symlink(target, linked); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildPrivateGTDProjection(context.Background(), filepath.Join(linked, "backlog.db"), 1, nil); err == nil {
		t.Fatal("symlink projection directory accepted")
	}
	if _, err := os.Stat(filepath.Join(target, privateGTDDataName)); !os.IsNotExist(err) {
		t.Fatalf("projection escaped through symlink: %v", err)
	}
}

func TestPrivateGTDProjectionAccountAndCanonicalCollectorIsolation(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "db", "project", "todo")
	projection, err := BuildPrivateGTDProjection(context.Background(), filepath.Join(dir, "backlog.db"), 1, []PrivateGTDEdge{{From: "PRIVATE-FROM", To: "PRIVATE-TO", Kind: "related_to", Sensitivity: "secret"}})
	if err != nil {
		t.Fatal(err)
	}
	system := SystemPrivateGTDPermissionChecker()
	uid, err := system.CurrentUID()
	if err != nil {
		t.Fatal(err)
	}
	if err := AuthorizePrivateGTDProjection(projection, uid, system); err != nil {
		t.Fatalf("kernel owner/mode authorization: %v", err)
	}
	paths := []string{filepath.Dir(projection.DataPath), projection.DataPath, projection.MetaPath}
	fake := fakePrivatePermissionChecker{current: "501", owners: map[string]string{}, modes: map[string]os.FileMode{}}
	for i, path := range paths {
		fake.owners[path] = "501"
		if i == 0 {
			fake.modes[path] = 0o700
		} else {
			fake.modes[path] = 0o600
		}
	}
	if err := AuthorizePrivateGTDProjection(projection, "501", fake); err != nil {
		t.Fatal(err)
	}
	if err := AuthorizePrivateGTDProjection(projection, "502", fake); err == nil {
		t.Fatal("different account authorized")
	}
	fake.owners[projection.DataPath] = "502"
	if err := AuthorizePrivateGTDProjection(projection, "501", fake); err == nil {
		t.Fatal("foreign owner authorized")
	}
	fake.owners[projection.DataPath] = "501"
	fake.modes[projection.MetaPath] = 0o644
	if err := AuthorizePrivateGTDProjection(projection, "501", fake); err == nil {
		t.Fatal("weak mode authorized")
	}

	privateBytes, err := os.ReadFile(projection.DataPath)
	if err != nil {
		t.Fatal(err)
	}
	privateMeta, err := os.ReadFile(projection.MetaPath)
	if err != nil {
		t.Fatal(err)
	}

	store := kanban.NewBacklogStore(filepath.Join(dir, "backlog.json"))
	if _, err := kanban.CaptureGTDItem(context.Background(), store, kanban.CaptureInput{
		Content: "ordinary export item", Source: "test", SourceAllowed: true,
		Sensitivity: kanban.SensitivityPublic, EventID: "collector-export",
	}); err != nil {
		t.Fatal(err)
	}

	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(repo, 0o700); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	if err := os.WriteFile(filepath.Join(repo, "ordinary.txt"), []byte("ordinary repository data"), 0o600); err != nil {
		t.Fatal(err)
	}

	outputs := make(map[string][]byte, 7)
	snapshot, err := worktree.Capture(context.Background(), worktree.CaptureOptions{RepoDir: repo, SnapshotID: "gtd-privacy"})
	if err != nil {
		t.Fatal(err)
	}
	outputs["repository-snapshot"] = mustJSON(t, snapshot)

	templateFS, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	var templateOutput bytes.Buffer
	if err := fs.WalkDir(templateFS, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		templateOutput.WriteString(path)
		templateOutput.WriteByte('\n')
		raw, readErr := fs.ReadFile(templateFS, path)
		if readErr != nil {
			return readErr
		}
		templateOutput.Write(raw)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	outputs["embedded-template-manifest"] = templateOutput.Bytes()

	gitCmd := exec.Command("git", "ls-files", "--cached", "--others", "--exclude-standard", "-z")
	gitCmd.Dir = repo
	gitOutput, err := gitCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git collector: %v: %s", err, gitOutput)
	}
	outputs["git-index-manifest"] = gitOutput

	if err := kanban.AppendSlotLeaseAudit(repo, kanban.SlotLeaseAuditEntry{Event: "refuse", Reason: "ordinary", Resource: "lane-10", SessionID: "collector-log"}); err != nil {
		t.Fatal(err)
	}
	outputs["application-log"] = mustRead(t, filepath.Join(repo, ".moai", "logs", kanban.SlotLeaseAuditFileName))

	now := time.Now().UTC()
	if err := telemetry.RecordSkillUsage(repo, telemetry.UsageRecord{Timestamp: now, SessionID: "collector-telemetry", SkillID: "moai-gtd", Trigger: telemetry.TriggerAuto, ContextHash: telemetry.HashContext("ordinary"), AgentType: "mission-governor", Phase: "run", Outcome: telemetry.OutcomeSuccess}); err != nil {
		t.Fatal(err)
	}
	report, err := telemetry.GenerateReport(repo, 1)
	if err != nil {
		t.Fatal(err)
	}
	telemetryPath := filepath.Join(repo, ".moai", "evolution", "telemetry", "usage-"+now.Format("2006-01-02")+".jsonl")
	outputs["telemetry-report"] = append(mustRead(t, telemetryPath), []byte(report.String())...)

	exported, err := kanban.ExportGTD(context.Background(), store, true)
	if err != nil {
		t.Fatal(err)
	}
	outputs["gtd-export"] = exported

	backupPath := filepath.Join(root, "backup", "backlog.db")
	backupErr := kanban.BackupGTDStore(context.Background(), store, backupPath, false)
	if backupErr == nil {
		t.Fatal("default backup unexpectedly succeeded")
	}
	outputs["default-backup"] = []byte(backupErr.Error())
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Fatalf("default backup artifact exists: %v", err)
	}

	if len(outputs) != 7 {
		t.Fatalf("production collector count=%d, want 7", len(outputs))
	}
	names := make([]string, 0, len(outputs))
	for name := range outputs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		raw := outputs[name]
		if len(raw) == 0 {
			t.Errorf("%s produced no invocation evidence", name)
			continue
		}
		for marker, needle := range map[string][]byte{
			"projection path":  []byte(projection.DataPath),
			"sidecar path":     []byte(projection.MetaPath),
			"projection bytes": privateBytes,
			"sidecar bytes":    privateMeta,
			"private node":     []byte("PRIVATE-FROM"),
		} {
			if bytes.Contains(raw, needle) {
				t.Errorf("%s leaked %s", name, marker)
			}
		}
		t.Logf("collector=%s output_bytes=%d leak_hits=0", name, len(raw))
	}
	if got := PrivateGTDBackupFiles(projection, false); len(got) != 0 {
		t.Fatalf("default backup leak: %v", got)
	}
	if got := PrivateGTDBackupFiles(projection, true); len(got) != 2 {
		t.Fatalf("explicit opt-in backup=%v", got)
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return raw
}

func TestPrivateGTDProjectionPublicationFaultsCleanUp(t *testing.T) {
	for stage := 1; stage <= 5; stage++ {
		publisher := &faultPrivatePublisher{failAt: stage}
		if _, _, err := publishPrivateGTDProjectionWith(publisher, "/private", []byte("data"), []byte("meta")); err == nil {
			t.Fatalf("stage %d failure ignored", stage)
		}
		if stage >= 3 && len(publisher.removed) == 0 {
			t.Fatalf("stage %d did not clean temporary state", stage)
		}
	}
	publisher := &faultPrivatePublisher{failAt: 99}
	data, meta, err := publishPrivateGTDProjectionWith(publisher, "/private", nil, nil)
	if err != nil || filepath.Base(data) != privateGTDDataName || filepath.Base(meta) != privateGTDMetaName {
		t.Fatalf("data=%q meta=%q err=%v", data, meta, err)
	}
}
