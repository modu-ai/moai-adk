package closure

import (
	"errors"
	"strings"
	"testing"
)

// fixtureSecondReviewLine returns a valid second-review record line per
// design.md §A.1.
func fixtureSecondReviewLine() string {
	return `{"schema_version":1,"card":"c1","contract_card":"c1","spec_id":"SPEC-EXAMPLE-001",` +
		`"contract_sha256":"deadbeef","head_sha":"abc1234","target":"baseBranch",` +
		`"scope":{"base_branch":"main","base_sha":"base01","head_sha":"abc1234",` +
		`"changed_files":7,"diff_sha256":"diff01"},` +
		`"backends":[{"backend":"claude","gate":"required","verdict":"pass"},` +
		`{"backend":"codex","gate":"required","verdict":"fail"}],` +
		`"participant_count":2,"disagreement_flag":true,"audit_receipt":"rcpt-1",` +
		`"build_commit":"build01","recorded_at":"2026-09-26T09:00:00Z"}`
}

// TestDecodeSecondReviewLine pins design.md §A.1 decode.
func TestDecodeSecondReviewLine(t *testing.T) {
	r, err := DecodeSecondReviewLine([]byte(fixtureSecondReviewLine()))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if r.Card != "c1" || r.ContractCard != "c1" || r.SpecID != "SPEC-EXAMPLE-001" {
		t.Fatalf("card fields = %q/%q/%q", r.Card, r.ContractCard, r.SpecID)
	}
	if r.Target != "baseBranch" || r.HeadSHA != "abc1234" {
		t.Fatalf("target/head = %q/%q", r.Target, r.HeadSHA)
	}
	if r.Scope.BaseBranch != "main" || r.Scope.BaseSHA != "base01" ||
		r.Scope.HeadSHA != "abc1234" || r.Scope.ChangedFiles != 7 || r.Scope.DiffSHA256 != "diff01" {
		t.Fatalf("scope = %+v", r.Scope)
	}
	if len(r.Backends) != 2 || r.Backends[1].Backend != "codex" || r.Backends[1].Verdict != "fail" {
		t.Fatalf("backends = %+v", r.Backends)
	}
	if r.ParticipantCount != 2 || r.DisagreementFlag == nil || !*r.DisagreementFlag {
		t.Fatalf("participants/disagreement = %d/%v", r.ParticipantCount, r.DisagreementFlag)
	}
	if r.AuditReceipt != "rcpt-1" || r.BuildCommit != "build01" {
		t.Fatalf("receipt/build = %q/%q", r.AuditReceipt, r.BuildCommit)
	}
}

// TestDecodeSecondReviewUnknownSchema pins the strict schema-version check:
// an unknown schema_version never decodes (it is skipped and listed later).
func TestDecodeSecondReviewUnknownSchema(t *testing.T) {
	line := strings.Replace(fixtureSecondReviewLine(), `"schema_version":1`, `"schema_version":99`, 1)
	_, err := DecodeSecondReviewLine([]byte(line))
	if err == nil {
		t.Fatalf("expected error for schema_version 99, got nil")
	}
	if !errors.Is(err, ErrUnknownRecordSchema) {
		t.Fatalf("error = %v, want ErrUnknownRecordSchema", err)
	}
}

// TestDecodeSecondReviewMalformed pins a malformed line errors.
func TestDecodeSecondReviewMalformed(t *testing.T) {
	if _, err := DecodeSecondReviewLine([]byte("not json")); err == nil {
		t.Fatalf("expected error for malformed line, got nil")
	}
}

// TestLoadSecondReviewsSkips pins design.md §D: lines with an unknown
// schema_version are skipped and listed; valid lines decode.
func TestLoadSecondReviewsSkips(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/second-review.jsonl"
	content := fixtureSecondReviewLine() + "\n" +
		strings.Replace(fixtureSecondReviewLine(), `"schema_version":1`, `"schema_version":2`, 1) + "\n" +
		"\n"
	if err := writeFileForTest(path, content); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	recs, skipped, err := LoadSecondReviews(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(recs) != 1 || recs[0].Card != "c1" {
		t.Fatalf("records = %+v, want 1 valid", recs)
	}
	if len(skipped) != 1 || skipped[0] != "schema_version 2" {
		t.Fatalf("skipped = %v, want [schema_version 2]", skipped)
	}
}

// TestDecodeVerdictRecord pins design.md §A.2 decode.
func TestDecodeVerdictRecord(t *testing.T) {
	line := `{"schema_version":1,"card":"c1","spec_id":"SPEC-EXAMPLE-001","verdict":"accept",` +
		`"note":"ok","operator":{"name":"Jane Doe","email":"jane@example.com"},` +
		`"recorded_at":"2026-09-26T09:00:00Z","report_sha256":"rep01","method":"interactive-tty"}`
	r, err := DecodeVerdictRecord([]byte(line))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if r.Verdict != "accept" || r.Card != "c1" || r.Operator.Name != "Jane Doe" ||
		r.Operator.Email != "jane@example.com" || r.ReportSHA256 != "rep01" ||
		r.Method != "interactive-tty" {
		t.Fatalf("record = %+v", r)
	}

	badSchema := strings.Replace(line, `"schema_version":1`, `"schema_version":7`, 1)
	if _, err := DecodeVerdictRecord([]byte(badSchema)); !errors.Is(err, ErrUnknownRecordSchema) {
		t.Fatalf("error = %v, want ErrUnknownRecordSchema", err)
	}
}

// TestLoadVerdictRecords pins latest-line-wins ordering: the file's last
// valid line is the latest verdict, and the count of decoded lines returns.
func TestLoadVerdictRecords(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/closure-verdict.jsonl"
	first := `{"schema_version":1,"card":"c1","spec_id":"SPEC-EXAMPLE-001","verdict":"reject",` +
		`"operator":{"name":"A","email":"a@x"},"recorded_at":"2026-09-26T08:00:00Z",` +
		`"report_sha256":"old","method":"interactive-tty"}`
	second := strings.Replace(first, `"reject"`, `"accept"`, 1)
	second = strings.Replace(second, "2026-09-26T08:00:00Z", "2026-09-26T09:00:00Z", 1)
	if err := writeFileForTest(path, first+"\n"+second+"\n"); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	recs, skipped, err := LoadVerdictRecords(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(recs) != 2 || len(skipped) != 0 {
		t.Fatalf("recs=%d skipped=%d, want 2/0", len(recs), len(skipped))
	}
	latest := LatestVerdict(recs)
	if latest == nil || latest.Verdict != "accept" {
		t.Fatalf("latest = %+v, want accept", latest)
	}
}
