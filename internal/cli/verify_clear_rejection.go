package cli

// verify_clear_rejection.go — `moai verify clear-rejection <agent-type> <spec-id>
// --evidence <path>`, the one operator path that clears a receipt-kind rejection
// mark (SPEC-AUDIT-ENGINE-CONVERGENCE-001 REQ-AEC-011, REQ-AEC-012, REQ-AEC-013).
//
// It is a human-path verb: it reuses the contract signing path's agent-environment
// markers, its interactive-terminal check and its typed confirmation, and refuses
// (exit 1, nothing written, nothing removed) unless every check holds. There is
// no MCP tool, hook path or configuration key that does the same.
//
// What the checks prove is narrow. They show that this process has neither agent
// marker set and has a terminal on standard input. They do not stop an agent that
// unsets the markers and allocates a pseudo-terminal, a Codex session (which sets
// neither marker), or a direct deletion of the mark file. Under the cooperating-
// process threat model this guards against operator mistakes; it is not a
// security boundary (spec.md section E, B4).

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/closure/gitio"
)

// Refusal codes, in the shape `moai contract verdict` uses: exit 1, message
// "refused <code>: <detail>".
const (
	clearRefAgentMarker    = "agent_marker"
	clearRefNotTTY         = "not_tty"
	clearRefMarkNotFound   = "mark_not_found"
	clearRefEvidencePath   = "evidence_path"
	clearRefEvidenceSpec   = "evidence_spec"
	clearRefConfirmation   = "confirmation_mismatch"
	clearRefGitIdentity    = "git_identity_missing"
	clearRefHeadUnreadable = "head_unreadable"

	// clearUnknownSpecArg is the SPEC argument that names a mark whose verdict
	// line carried no SPEC ID (auditreceipt.UnknownSpec); the evidence file must
	// then name the agent type instead.
	clearUnknownSpecArg = "unknown"

	// clearEvidenceMaxBytes bounds the evidence file the verb reads and hashes.
	clearEvidenceMaxBytes = 1 << 20
)

func clearRefusal(code, format string, a ...any) error {
	return &exitCodeError{code: contractExitInvalid, msg: "refused " + code + ": " + fmt.Sprintf(format, a...)}
}

// newVerifyClearRejectionCmd builds the `clear-rejection` verb under `moai verify`.
//
// @MX:NOTE: [AUTO] the only caller of auditreceipt.ClearRejectionOnApproval; keep it that way (REQ-AEC-011, AC-AEC-004c sweeps for it).
func newVerifyClearRejectionCmd(projectRoot *string) *cobra.Command {
	var evidence string
	cmd := &cobra.Command{
		Use:   "clear-rejection <agent-type> <spec-id>",
		Short: "Clear one receipt-kind rejection mark on operator approval (human path only; exit 0 cleared, 1 refused, 2 usage)",
		Long: `Clear exactly one receipt-kind rejection mark of the current tree on operator
approval, and append one approval record to the audit-receipt store.

The verb refuses, writing and removing nothing, unless: neither CLAUDECODE nor
CLAUDE_CODE_SESSION_ID is set; standard input is an interactive terminal; the
named mark exists for this tree; the evidence file is under .moai/reports/ and
names the mark's SPEC ID (for the SPEC argument "unknown": the agent type); and
you type "<agent-type> <spec-id>" when asked.`,
		Args:          cobra.ExactArgs(2),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			return runVerifyClearRejection(c, *projectRoot, args[0], args[1], evidence)
		},
	}
	cmd.Flags().StringVar(&evidence, "evidence", "", "evidence file under .moai/reports/ that names the mark's SPEC ID (required)")
	return cmd
}

func runVerifyClearRejection(cmd *cobra.Command, projectRootFlag, agentType, specArg, evidenceArg string) error {
	errOut := cmd.ErrOrStderr()
	refused := func(err error) error {
		_, _ = fmt.Fprintln(errOut, "moai verify clear-rejection: "+err.Error())
		return err
	}
	usage := func(format string, a ...any) error {
		msg := fmt.Sprintf(format, a...)
		_, _ = fmt.Fprintln(errOut, "moai verify clear-rejection: "+msg)
		return &exitCodeError{code: contractExitUsage, msg: msg}
	}
	agentType, specArg = strings.TrimSpace(agentType), strings.TrimSpace(specArg)
	if agentType == "" || specArg == "" || strings.TrimSpace(evidenceArg) == "" {
		return usage("an agent type, a SPEC ID (or %q) and --evidence are required", clearUnknownSpecArg)
	}

	// Agent-environment markers: the contract signing path's closed set.
	for _, marker := range contractAgentMarkers() {
		if contractGetenvFn(marker) != "" {
			return refused(clearRefusal(clearRefAgentMarker, "%s is set: an agent environment does not clear a rejection", marker))
		}
	}
	// Interactive terminal (refuses `</dev/null`).
	if !contractStdinIsTerminalFn() {
		return refused(clearRefusal(clearRefNotTTY, "standard input is not an interactive terminal"))
	}

	rootAbs, err := verifyResolveRoot(projectRootFlag)
	if err != nil {
		return usage("%v", err)
	}
	tree := auditreceipt.Canonical(rootAbs)
	store, err := verifyStoreRoot(tree)
	if err != nil {
		return usage("the receipt store cannot be resolved: %v", err)
	}

	// The named mark exists for THIS tree.
	markSpec := specArg
	if specArg == clearUnknownSpecArg {
		markSpec = "" // keyed auditreceipt.UnknownSpec
	}
	if _, rerr := auditreceipt.ReadRejectionInKind(store, tree, auditreceipt.KindReceipt, agentType, markSpec); rerr != nil {
		if os.IsNotExist(rerr) {
			return refused(clearRefusal(clearRefMarkNotFound, "no receipt-kind rejection mark %s / %s for this tree", agentType, specArg))
		}
		return usage("the rejection mark cannot be read: %v", rerr)
	}

	// The evidence file: under .moai/reports/ (symlinks resolved), naming the SPEC.
	evPath, evBytes, rerr := readClearEvidence(tree, evidenceArg)
	if rerr != nil {
		return refused(rerr)
	}
	needle, what := specArg, "SPEC ID"
	if specArg == clearUnknownSpecArg {
		needle, what = agentType, "agent type"
	}
	if !namesToken(evBytes, needle) {
		return refused(clearRefusal(clearRefEvidenceSpec, "the evidence file %s does not name the mark's %s %q", evPath, what, needle))
	}

	// Typed confirmation: "<agent-type> <spec-id>".
	want := agentType + " " + specArg
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Type %q to clear this rejection: ", want)
	line, readErr := newContractLineReader(cmd.InOrStdin())()
	if readErr != nil || strings.TrimSpace(line) != want {
		return refused(clearRefusal(clearRefConfirmation, "the typed confirmation did not match %q", want))
	}

	name, email, iderr := contractGitIdentityFn(tree)
	if iderr != nil {
		return refused(clearRefusal(clearRefGitIdentity, "%v", iderr))
	}
	head, herr := gitio.Head(tree)
	if herr != nil {
		return refused(clearRefusal(clearRefHeadUnreadable, "%v", herr))
	}

	sum := sha256.Sum256(evBytes)
	a := auditreceipt.Approval{
		Operator:       name,
		Email:          email,
		Confirmation:   want,
		EvidencePath:   evPath,
		EvidenceSHA256: hex.EncodeToString(sum[:]),
		Head:           head,
	}
	a.Mark.AgentType, a.Mark.SpecID = agentType, markSpec
	if err := auditreceipt.ClearRejectionOnApproval(store, tree, &a); err != nil {
		if errors.Is(err, auditreceipt.ErrMarkNotFound) {
			return refused(clearRefusal(clearRefMarkNotFound, "the mark disappeared before it could be cleared"))
		}
		return refused(fmt.Errorf("clear-rejection failed: %w", err))
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\ncleared %s / %s; approval recorded\n", agentType, specArg)
	return nil
}

// readClearEvidence resolves the evidence argument against the tree, requires the
// resolved file (symlinks followed) to be a regular file under the tree's
// .moai/reports/, and returns its tree-relative slash path and bytes.
func readClearEvidence(tree, arg string) (string, []byte, error) {
	p := arg
	if !filepath.IsAbs(p) {
		p = filepath.Join(tree, p)
	}
	reports, err := filepath.EvalSymlinks(filepath.Join(tree, ".moai", "reports"))
	if err != nil {
		return "", nil, clearRefusal(clearRefEvidencePath, "the tree has no .moai/reports/ directory")
	}
	resolved, err := filepath.EvalSymlinks(filepath.Clean(p))
	if err != nil {
		return "", nil, clearRefusal(clearRefEvidencePath, "the evidence file %q cannot be resolved", arg)
	}
	rel, err := filepath.Rel(reports, resolved)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", nil, clearRefusal(clearRefEvidencePath, "the evidence file %q is outside .moai/reports/", arg)
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() {
		return "", nil, clearRefusal(clearRefEvidencePath, "the evidence path %q is not a regular file", arg)
	}
	f, err := os.Open(resolved)
	if err != nil {
		return "", nil, clearRefusal(clearRefEvidencePath, "the evidence file %q cannot be read", arg)
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(io.LimitReader(f, clearEvidenceMaxBytes+1))
	if err != nil || len(data) > clearEvidenceMaxBytes {
		return "", nil, clearRefusal(clearRefEvidencePath, "the evidence file %q cannot be read in full (limit %d bytes)", arg, clearEvidenceMaxBytes)
	}
	treeRel, err := filepath.Rel(tree, resolved)
	if err != nil {
		treeRel = filepath.Join(".moai", "reports", rel)
	}
	return filepath.ToSlash(treeRel), data, nil
}

// namesToken reports whether data contains token as a whole identifier: not
// preceded by a letter, digit or hyphen, and not followed by a letter or digit,
// so SPEC-X-001 is not found inside SPEC-X-0012.
func namesToken(data []byte, token string) bool {
	re := regexp.MustCompile(`(?:^|[^A-Za-z0-9-])` + regexp.QuoteMeta(token) + `(?:$|[^A-Za-z0-9])`)
	return re.Match(data)
}
