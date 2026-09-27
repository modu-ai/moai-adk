package cli

// contract_verdict.go — `moai contract verdict <card-id>
// <accept|reject|amend-contract>`, the human verdict recorder
// (SPEC-AUTONOMY-CLOSURE-001 REQ-CLOSURE-020).
//
// Only a human on an interactive terminal records a verdict: an
// agent-environment marker (A1's closed marker set), a non-terminal standard
// input, a mismatched typed confirmation, a missing closure report, or a
// missing git identity refuses without writing (exit 1). Usage errors exit 2.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/closure"
)

// Valid verdict values (design.md §A.2).
var contractVerdicts = []string{"accept", "reject", "amend-contract"}

// contractGitIdentityFn reads the operator's git identity; tests replace it.
var contractGitIdentityFn = func(root string) (string, string, error) {
	name, err := gitConfigValue(root, "user.name")
	if err != nil || name == "" {
		return "", "", errors.New("git user.name is not set")
	}
	email, err := gitConfigValue(root, "user.email")
	if err != nil || email == "" {
		return "", "", errors.New("git user.email is not set")
	}
	return name, email, nil
}

// gitConfigValue reads one local git configuration value (trimmed).
func gitConfigValue(root, key string) (string, error) {
	cmd := exec.Command("git", "config", key)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// contractVerdictRefusal is one refusal reason of the human path.
type contractVerdictRefusal struct {
	code   string
	detail string
}

func runContractVerdict(cmd *cobra.Command, cardID, verdictValue, note string) error {
	if !isContractVerdict(verdictValue) {
		return contractUsageError(cmd, "verdict must be one of %s", strings.Join(contractVerdicts, "|"))
	}
	refuse := func(r contractVerdictRefusal) error {
		return &exitCodeError{code: contractExitInvalid, msg: "refused " + r.code + ": " + r.detail}
	}

	// Agent-environment markers (A1's closed marker set).
	for _, marker := range contractAgentMarkers() {
		if contractGetenvFn(marker) != "" {
			return refuse(contractVerdictRefusal{code: "agent_marker",
				detail: "an agent environment records a verdict only a human may record"})
		}
	}
	// Interactive terminal.
	if !contractStdinIsTerminalFn() {
		return refuse(contractVerdictRefusal{code: "not_tty",
			detail: "standard input is not an interactive terminal"})
	}
	// The closure report exists.
	specID, cause := cardToSpecID(cardID)
	if specID == "" {
		return contractUsageError(cmd, "%s", cause)
	}
	runDir, err := contractRunDirFn()
	if err != nil {
		return contractUsageError(cmd, "%v", err)
	}
	home, err := closure.ResolveEvidenceHome(runDir, cardID)
	if err != nil {
		return contractUsageError(cmd, "card evidence home: %v", err)
	}
	ev := closure.EvidenceFor(home, cardID)
	reportData, ok, err := readFileIfExists(ev.ReportJSON)
	if err != nil {
		return contractUsageError(cmd, "read report: %v", err)
	}
	if !ok {
		return refuse(contractVerdictRefusal{code: "report_missing",
			detail: "no closure report in " + ev.Dir + "; run `moai contract report " + cardID + "` first"})
	}
	var parsed closure.Report
	if err := json.Unmarshal(reportData, &parsed); err != nil {
		return refuse(contractVerdictRefusal{code: "report_missing",
			detail: "the closure report does not parse: " + err.Error()})
	}

	// Typed confirmation: the token is "<verdict> <card-id>".
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Type %q to record the verdict: ", verdictValue+" "+cardID)
	line, readErr := newContractLineReader(cmd.InOrStdin())()
	if readErr != nil || strings.TrimSpace(line) != verdictValue+" "+cardID {
		return refuse(contractVerdictRefusal{code: "confirmation_mismatch",
			detail: "the typed confirmation did not match " + verdictValue + " " + cardID})
	}

	// Git identity.
	name, email, iderr := contractGitIdentityFn(home)
	if iderr != nil {
		return refuse(contractVerdictRefusal{code: "git_identity_missing", detail: iderr.Error()})
	}

	rec := closure.VerdictRecord{
		SchemaVersion: closure.SchemaVersion,
		Card:          cardID,
		SpecID:        specID,
		Verdict:       verdictValue,
		Note:          note,
		Operator:      closure.VerdictOperator{Name: name, Email: email},
		RecordedAt:    time.Now().UTC().Format(time.RFC3339),
		ReportSHA256:  closure.CanonicalReportHash(&parsed),
		Method:        "interactive-tty",
	}
	line2, err := json.Marshal(rec)
	if err != nil {
		return contractUsageError(cmd, "encode verdict: %v", err)
	}
	if err := appendJSONL(ev.ClosureVerdict, line2); err != nil {
		return contractUsageError(cmd, "append verdict: %v", err)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "recorded %s for card %s\n", verdictValue, cardID)
	return nil
}

func isContractVerdict(v string) bool {
	for _, ok := range contractVerdicts {
		if ok == v {
			return true
		}
	}
	return false
}

// readFileIfExists reads a file, reporting absence distinctly.
func readFileIfExists(path string) ([]byte, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return data, true, nil
}

// appendJSONL appends one line to a .jsonl file, creating it (single write
// per record; spec.md §E).
func appendJSONL(path string, line []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(line, '\n'))
	return err
}

// newContractVerdictCmd builds the `verdict` subcommand.
func newContractVerdictCmd() *cobra.Command {
	var note string
	cmd := &cobra.Command{
		Use:           "verdict <card-id> <accept|reject|amend-contract>",
		Short:         "Record the operator's verdict (human path only; exit 0 recorded, 1 refused, 2 usage)",
		Args:          contractArgs(2, 2),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			return runContractVerdict(c, args[0], args[1], note)
		},
	}
	cmd.Flags().StringVar(&note, "note", "", "Optional note carried on the verdict record")
	return cmd
}
