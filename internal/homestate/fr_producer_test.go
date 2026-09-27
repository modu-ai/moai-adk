package homestate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// frRepoRoot is the repository root relative to this package directory.
const frRepoRoot = "../.."

var frAuditorFiles = []string{
	".claude/agents/moai/plan-auditor.md",
	".claude/agents/moai/sync-auditor.md",
	"internal/template/templates/.claude/agents/moai/plan-auditor.md",
	"internal/template/templates/.claude/agents/moai/sync-auditor.md",
}

var frConventionFiles = []string{
	".moai/docs/audit-artifact-convention.md",
	"internal/template/templates/.moai/docs/audit-artifact-convention.md",
}

const frCiteHeading = "### [HARD] Cite your audit receipt"

func frReadRepoFile(t *testing.T, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(frRepoRoot, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(raw)
}

// frExampleLine returns the first line of body that starts with prefix — the
// template line an auditor is told to write.
func frExampleLine(body, prefix string) string {
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(strings.TrimLeft(line, " \t"), prefix) {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

// AC-020 — both auditors and the audit-artifact convention (local and
// template) instruct the two verdict-file lines; the emitted Codex
// definitions carry them; a verdict file written from the instruction's
// example passes the E-VERDICT reader; and the AUDIT-VERDICT chat-message
// line stays the last instruction so ParseVerdictLine keeps working.
func TestFR_AC020_VerdictLineProducer(t *testing.T) {
	const formatLine = "verdict: <PASS|PASS-WITH-DEBT|FAIL>"
	for _, rel := range append(append([]string{}, frAuditorFiles...), frConventionFiles...) {
		body := frReadRepoFile(t, rel)
		if n := strings.Count(body, "audited_sha:"); n < 1 {
			t.Errorf("%s: audited_sha: count = %d, want ≥ 1", rel, n)
		}
		if n := strings.Count(body, formatLine); n < 1 {
			t.Errorf("%s: %q count = %d, want ≥ 1", rel, formatLine, n)
		}
	}
	for _, rel := range []string{
		"internal/template/templates/.codex/agents/moai/plan-auditor.toml",
		"internal/template/templates/.codex/agents/moai/sync-auditor.toml",
	} {
		if !strings.Contains(frReadRepoFile(t, rel), "audited_sha") {
			t.Errorf("%s does not carry audited_sha (run make agents-emit)", rel)
		}
	}

	// Guardrail (i): nothing after the receipt-citation block instructs the
	// verdict-file lines, and the block still carries AUDIT-VERDICT.
	for _, rel := range frAuditorFiles {
		body := frReadRepoFile(t, rel)
		at := strings.Index(body, frCiteHeading)
		if at < 0 {
			t.Fatalf("%s: %q block is missing", rel, frCiteHeading)
		}
		tail := body[at:]
		next := strings.Index(tail[len(frCiteHeading):], "\n### ")
		block := tail
		if next >= 0 {
			block = tail[:len(frCiteHeading)+next]
		}
		if !strings.Contains(block, "AUDIT-VERDICT:") {
			t.Errorf("%s: the receipt-citation block lost its AUDIT-VERDICT: literal", rel)
		}
		if strings.Contains(tail, "audited_sha") || strings.Contains(tail, formatLine) || strings.Contains(tail, "\nverdict: ") {
			t.Errorf("%s: a verdict-file line instruction appears after the receipt-citation block", rel)
		}
	}

	// A verdict file written exactly as the convention's example is accepted.
	db := frOpen(t)
	repo := frNewRepo(t, true)
	convention := frReadRepoFile(t, frConventionFiles[0])
	verdictTmpl := frExampleLine(convention, "verdict: <")
	shaTmpl := frExampleLine(convention, "audited_sha: <")
	if verdictTmpl == "" || shaTmpl == "" {
		t.Fatalf("convention example lines missing: verdict=%q audited_sha=%q", verdictTmpl, shaTmpl)
	}
	body := "# Plan audit\n\nProse verdict: the plan holds.\n\n" +
		strings.Replace(verdictTmpl, formatLine[len("verdict: "):], "PASS", 1) + "\n" +
		shaTmpl[:strings.Index(shaTmpl, "<")] + repo.Commit + "\n"
	frWrite(t, filepath.Join(repo.Dir, ".moai", "reports", "producer", "plan-audit.md"), body)
	if v, sha, err := ParseAuditVerdictFile("plan-audit.md", []byte(body), repo.Commit); err != nil || v != "PASS" || sha != repo.Commit {
		t.Fatalf("E-VERDICT on the example = (%q, %q, %v), want PASS / %s", v, sha, err, repo.Commit)
	}
	c := frLeasedCard(repo, "producer", CardPlanAudit)
	c.EvidenceSHA = repo.Commit
	frPlace(t, db, c)
	if got, err := db.Transition(context.Background(), frHolderRequest(c, CardKickoff)); err != nil || got.State != CardKickoff {
		t.Fatalf("plan-audit → kickoff on the example file: state=%s err=%v", got.State, err)
	}

	// Guardrail (ii): the chat message ends with AUDIT-VERDICT; appending a
	// verdict-file line after it breaks ParseVerdictLine.
	message := "Report body.\n\nverdict details in the exported file.\n\nAUDIT-VERDICT: PASS spec=SPEC-FIXTURE-001 receipts=none\n"
	if _, ok := auditreceipt.ParseVerdictLine(message); !ok {
		t.Fatalf("ParseVerdictLine rejected the instructed final message")
	}
	if _, ok := auditreceipt.ParseVerdictLine(message + "verdict: PASS\n"); ok {
		t.Fatalf("ParseVerdictLine accepted a message whose last line is a verdict-file line")
	}
}
