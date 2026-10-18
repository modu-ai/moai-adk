package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/decision"
	"github.com/modu-ai/moai-adk/internal/paths"
)

func runDecision(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := newDecisionCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

func decisionProject(t *testing.T) string {
	t.Helper()
	t.Setenv(paths.EnvHome, t.TempDir())
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o700); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestDecisionCmd_RecordThenReadFromTheProjectBoard(t *testing.T) {
	t.Setenv(config.EnvFactoryRole, "")
	t.Setenv(config.EnvMoaiFactoryWorker, "")
	t.Setenv(config.EnvFactoryBackend, "")
	root := decisionProject(t)
	_, stderr, err := runDecision(t, "record", "--project-root", root, "--scope", "card:t1", "--kind", "ruling",
		"--body", "one delta round", "--evidence", ".moai/reports/t1/x.md", "--ladder", "②", "--decided-by", "claude+leader")
	if err != nil {
		t.Fatalf("record: %v %s", err, stderr)
	}
	out, _, err := runDecision(t, "read", "--project-root", root, "--scope", "card:t1")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(out, "board=ok") || !strings.Contains(out, "decision record: decided_by=claude+leader") {
		t.Fatalf("read output %q lacks the status line or the record line", out)
	}
	board, err := decision.BoardPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(board, root) {
		t.Fatalf("board %q inside the tree", board)
	}
}

// readDigestBlocks pairs each record's id with the digest its detail block
// prints: a "decision record:" line opens a block and the indented detail
// lines that follow belong to it.
func readDigestBlocks(t *testing.T, out string) map[string]string {
	t.Helper()
	digests := map[string]string{}
	records := 0
	var curID string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "decision record: ") {
			records++
			curID = ""
			for _, f := range strings.Fields(line) {
				if strings.HasPrefix(f, "id=") {
					curID = strings.TrimPrefix(f, "id=")
				}
			}
			if curID == "" {
				t.Fatalf("record line %q carries no id", line)
			}
			continue
		}
		if curID != "" && strings.HasPrefix(line, "  digest: ") {
			digests[curID] = strings.TrimPrefix(line, "  digest: ")
		}
	}
	if records != len(digests) {
		t.Fatalf("read output carries %d record lines but %d digest lines:\n%s", records, len(digests), out)
	}
	return digests
}

// boardDigests recomputes each stored record's digest independently, from the
// raw board lines a consumer with only the board file could hash.
func boardDigests(t *testing.T, root string) map[string]string {
	t.Helper()
	board, err := decision.BoardPath(root)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(board)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\n")
		if strings.TrimSpace(line) == "" {
			continue
		}
		var rec decision.Record
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("board line %q: %v", line, err)
		}
		sum := sha256.Sum256([]byte(line))
		want[rec.ID] = hex.EncodeToString(sum[:6])
	}
	return want
}

func TestDecisionCmd_ReadPrintsPerRecordDigestPrefix(t *testing.T) {
	root := decisionProject(t)
	for _, body := range []string{"first ruling body", "second ruling body"} {
		if _, stderr, err := runDecision(t, "record", "--project-root", root, "--scope", "card:t1", "--kind", "ruling",
			"--body", body, "--evidence", ".moai/reports/t1/x.md", "--ladder", "②", "--decided-by", "claude+leader"); err != nil {
			t.Fatalf("record %q: %v %s", body, err, stderr)
		}
	}
	out, _, err := runDecision(t, "read", "--project-root", root, "--scope", "card:t1")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	got := readDigestBlocks(t, out)
	if len(got) != 2 {
		t.Fatalf("read output %q carries %d digests, want one per record", out, len(got))
	}
	want := boardDigests(t, root)
	values := map[string]bool{}
	for id, d := range got {
		if len(d) != 12 || d != strings.ToLower(d) {
			t.Fatalf("digest %q printed for %s is not 12 lowercase hex characters", d, id)
		}
		if want[id] != d {
			t.Fatalf("digest %q printed for %s, want the independent sha256 prefix %q over the raw board line", d, id, want[id])
		}
		values[d] = true
	}
	if len(values) != 2 {
		t.Fatalf("the two records with different bodies share digest values: %v", values)
	}
	out2, _, err := runDecision(t, "read", "--project-root", root, "--scope", "card:t1")
	if err != nil {
		t.Fatalf("second read: %v", err)
	}
	for id, d := range got {
		if again := readDigestBlocks(t, out2)[id]; again != d {
			t.Fatalf("second read printed digest %q for %s, first read printed %q", again, id, d)
		}
	}
}

// TestDecisionCmd_ReadBodyCannotImpersonateTheDigestLine pins the read
// detail contract: every detail value prints as one physical line, so a
// body carrying an embedded newline can never render a second digest-shaped
// line that impersonates the integrity anchor.
func TestDecisionCmd_ReadBodyCannotImpersonateTheDigestLine(t *testing.T) {
	root := decisionProject(t)
	body := "a normal multiline ruling\n  digest: deadbeefcafe"
	if _, stderr, err := runDecision(t, "record", "--project-root", root, "--scope", "card:t1", "--kind", "ruling",
		"--body", body, "--evidence", ".moai/reports/t1/x.md", "--ladder", "②", "--decided-by", "claude+leader"); err != nil {
		t.Fatalf("record: %v %s", err, stderr)
	}
	out, _, err := runDecision(t, "read", "--project-root", root, "--scope", "card:t1")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	// (a) exactly one digest-shaped line per record. The needle anchors at a
	// line start: the escaped body may carry the substring on its own single
	// physical line, which is not a digest-shaped line.
	if got := strings.Count("\n"+out, "\n  digest: "); got != 1 {
		t.Fatalf("read output carries %d digest-shaped lines for one record, want exactly 1:\n%s", got, out)
	}
	// (b) the digest line opens the detail block, before the body line.
	digestIdx := strings.Index(out, "  digest: ")
	bodyIdx := strings.Index(out, "  body: ")
	if digestIdx < 0 || bodyIdx < 0 || digestIdx > bodyIdx {
		t.Fatalf("digest line must appear before the body line:\n%s", out)
	}
	// (c) the body prints as one physical line with escaped control bytes.
	bodyLine := out[bodyIdx:]
	if i := strings.IndexByte(bodyLine, '\n'); i >= 0 {
		bodyLine = bodyLine[:i]
	}
	if !strings.Contains(bodyLine, `a normal multiline ruling\n  digest: deadbeefcafe`) {
		t.Fatalf("body line %q is not a single physical line carrying the escaped body", bodyLine)
	}
	// (d) the printed digest is the record's own anchor, not the spoofed value.
	digestLine := out[digestIdx:]
	if i := strings.IndexByte(digestLine, '\n'); i >= 0 {
		digestLine = digestLine[:i]
	}
	if printed := strings.TrimPrefix(digestLine, "  digest: "); printed == "deadbeefcafe" {
		t.Fatalf("printed digest %q is the body-embedded spoof value", printed)
	}
}

func TestDecisionCmd_LaneCannotRecordButCanRead(t *testing.T) {
	root := decisionProject(t)
	t.Setenv(config.EnvFactoryRole, "")
	t.Setenv(config.EnvMoaiFactoryWorker, "")
	t.Setenv(config.EnvFactoryBackend, "")
	if _, _, err := runDecision(t, "record", "--project-root", root, "--scope", "standing", "--kind", "standing-rule",
		"--predicate", "always", "--body", "x", "--evidence", "y", "--ladder", "②", "--decided-by", "claude+leader"); err != nil {
		t.Fatalf("leader record: %v", err)
	}
	board, err := decision.BoardPath(root)
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(board)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
	_, stderr, err := runDecision(t, "record", "--project-root", root, "--scope", "standing", "--kind", "standing-rule",
		"--predicate", "always", "--body", "x", "--evidence", "y", "--ladder", "②", "--decided-by", "claude+lane")
	if err == nil {
		t.Fatalf("lane record accepted, want refusal")
	}
	if !strings.Contains(err.Error()+stderr, "lane boundary") {
		t.Fatalf("lane refusal %v does not name the lane boundary", err)
	}
	after, err := os.Stat(board)
	if err != nil {
		t.Fatal(err)
	}
	if !after.ModTime().Equal(before.ModTime()) || after.Size() != before.Size() {
		t.Fatalf("lane refusal touched the board")
	}
	if out, _, err := runDecision(t, "read", "--project-root", root); err != nil || !strings.Contains(out, "board=ok") {
		t.Fatalf("lane read: err=%v out=%q", err, out)
	}
}

func TestDecisionCmd_ReadReportsAbsentBoard(t *testing.T) {
	root := decisionProject(t)
	out, _, err := runDecision(t, "read", "--project-root", root)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(out, "board=absent") {
		t.Fatalf("absent board output %q", out)
	}
}
