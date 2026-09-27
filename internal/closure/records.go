package closure

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// ErrUnknownRecordSchema reports a record line whose schema_version is not
// SchemaVersion. The selection and the report skip such lines and list them
// (design.md §D, acceptance.md §D edge cases); they never count as evidence.
var ErrUnknownRecordSchema = errors.New("closure: unknown record schema_version")

// SecondReviewScope is the reviewed scope of one second review (design.md
// §A.1): the base the review backends resolved and the change between it and
// the audited commit.
type SecondReviewScope struct {
	BaseBranch   string `json:"base_branch"`
	BaseSHA      string `json:"base_sha"`
	HeadSHA      string `json:"head_sha"`
	ChangedFiles int    `json:"changed_files"`
	DiffSHA256   string `json:"diff_sha256"`
}

// SecondReviewBackend is one backend's verdict inside a second-review record.
type SecondReviewBackend struct {
	Backend string `json:"backend"`
	Gate    string `json:"gate"`
	Verdict string `json:"verdict"`
}

// SecondReviewRecord is one line of the card evidence directory's
// `second-review.jsonl` (design.md §A.1). The MCP server writes it; the
// report generator and the push readiness evaluator read it. Empty string
// fields mean what design.md §A.1 says: spec_id "" is an unbound record,
// contract_sha256 "" an absent contract, empty scope fields a git failure.
type SecondReviewRecord struct {
	SchemaVersion    int                   `json:"schema_version"`
	Card             string                `json:"card"`
	ContractCard     string                `json:"contract_card"`
	SpecID           string                `json:"spec_id"`
	ContractSHA256   string                `json:"contract_sha256"`
	HeadSHA          string                `json:"head_sha"`
	Target           string                `json:"target"`
	Scope            SecondReviewScope     `json:"scope"`
	Backends         []SecondReviewBackend `json:"backends"`
	ParticipantCount int                   `json:"participant_count"`
	DisagreementFlag *bool                 `json:"disagreement_flag"`
	AuditReceipt     string                `json:"audit_receipt"`
	BuildCommit      string                `json:"build_commit"`
	RecordedAt       string                `json:"recorded_at"`
}

// VerdictOperator is the git identity of the human who recorded a verdict.
type VerdictOperator struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// VerdictRecord is one line of the card evidence directory's
// `closure-verdict.jsonl` (design.md §A.2). Only `moai contract verdict` on a
// human path writes it; the report generator and the readiness evaluator read
// it. The latest line wins; earlier lines stay as history.
type VerdictRecord struct {
	SchemaVersion int             `json:"schema_version"`
	Card          string          `json:"card"`
	SpecID        string          `json:"spec_id"`
	Verdict       string          `json:"verdict"` // accept | reject | amend-contract
	Note          string          `json:"note"`
	Operator      VerdictOperator `json:"operator"`
	RecordedAt    string          `json:"recorded_at"`
	ReportSHA256  string          `json:"report_sha256"`
	Method        string          `json:"method"`
}

// DecodeSecondReviewLine decodes one second-review record line. A line whose
// schema_version is unknown fails with ErrUnknownRecordSchema; it is never
// silently accepted.
func DecodeSecondReviewLine(data []byte) (SecondReviewRecord, error) {
	var r SecondReviewRecord
	if err := json.Unmarshal(data, &r); err != nil {
		return SecondReviewRecord{}, fmt.Errorf("closure: decode second-review line: %w", err)
	}
	if r.SchemaVersion != SchemaVersion {
		return SecondReviewRecord{}, fmt.Errorf("%w: %d", ErrUnknownRecordSchema, r.SchemaVersion)
	}
	return r, nil
}

// DecodeVerdictRecord decodes one human verdict record line. A line whose
// schema_version is unknown fails with ErrUnknownRecordSchema.
func DecodeVerdictRecord(data []byte) (VerdictRecord, error) {
	var r VerdictRecord
	if err := json.Unmarshal(data, &r); err != nil {
		return VerdictRecord{}, fmt.Errorf("closure: decode verdict line: %w", err)
	}
	if r.SchemaVersion != SchemaVersion {
		return VerdictRecord{}, fmt.Errorf("%w: %d", ErrUnknownRecordSchema, r.SchemaVersion)
	}
	return r, nil
}

// LoadSecondReviews reads the card evidence directory's second-review.jsonl.
// A missing file is "no records": empty slices and a nil error. Valid lines
// decode in file order; lines with an unknown schema_version are skipped and
// listed as "schema_version <n>"; malformed lines are skipped and listed as
// "malformed line <k>" (1-based) so nothing disappears silently.
func LoadSecondReviews(path string) (records []SecondReviewRecord, skipped []string, err error) {
	return loadRecords(path, func(line []byte) (SecondReviewRecord, error) {
		return DecodeSecondReviewLine(line)
	})
}

// LoadVerdictRecords reads the card evidence directory's closure-verdict.jsonl
// with the same skip rules as LoadSecondReviews.
func LoadVerdictRecords(path string) (records []VerdictRecord, skipped []string, err error) {
	return loadRecords(path, func(line []byte) (VerdictRecord, error) {
		return DecodeVerdictRecord(line)
	})
}

// loadRecords is the shared line reader behind both loaders: missing file →
// empty, unknown schema → listed skip, malformed line → listed skip.
func loadRecords[T any](path string, decode func([]byte) (T, error)) ([]T, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []T{}, nil, nil
		}
		return nil, nil, fmt.Errorf("closure: read %s: %w", path, err)
	}
	out := []T{}
	var skipped []string
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		r, derr := decode([]byte(line))
		if derr == nil {
			out = append(out, r)
			continue
		}
		if errors.Is(derr, ErrUnknownRecordSchema) {
			skipped = append(skipped, schemaVersionLabel(line))
			continue
		}
		skipped = append(skipped, fmt.Sprintf("malformed line %d", i+1))
	}
	return out, skipped, nil
}

// LatestVerdict returns the latest verdict record — the file's last valid
// line, appends being chronological — or nil when there is none.
func LatestVerdict(records []VerdictRecord) *VerdictRecord {
	if len(records) == 0 {
		return nil
	}
	return &records[len(records)-1]
}

// schemaVersionLabel extracts the schema_version of an unknown-schema line
// for the skipped list, falling back to "unknown" when the shape surprises us.
func schemaVersionLabel(line string) string {
	var probe struct {
		SchemaVersion int `json:"schema_version"`
	}
	if json.Unmarshal([]byte(line), &probe) == nil && probe.SchemaVersion != 0 {
		return fmt.Sprintf("schema_version %d", probe.SchemaVersion)
	}
	return "schema_version unknown"
}
