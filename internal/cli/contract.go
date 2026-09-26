package cli

// contract.go — `moai contract sign|show|verify`, the CLI over the SPEC
// autonomy contract (SPEC-AUTONOMY-CONTRACT-001 REQ-CONTRACT-008..014, 019,
// 021, 022, 024, 025).
//
// The command is a thin adapter: it resolves the project root, reads the
// effective workflow.autonomy configuration and the constitution registry,
// reads the SPEC's spec.md status, and hands plain values to the pure
// verification core (internal/contract) and the signer
// (internal/contract/sign). Every decision — reason codes, refusal codes, the
// signature itself — belongs to those packages; this file only maps their
// outcomes onto output and exit codes (design.md § Exit codes):
//
//	verify  0 valid     1 invalid   2 usage / I/O error
//	show    0 printed   —           2 usage / I/O error
//	sign    0 signed    1 refused   2 usage / I/O error

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/constitution"
	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// Exit codes of the contract command (design.md § Exit codes).
const (
	contractExitInvalid = 1 // verify: invalid; sign: refused
	contractExitUsage   = 2 // usage or I/O error
)

// contractGetenvFn is the environment the contract command reads (the
// agent-environment markers and the registry override). Tests replace it so
// the runtime's own environment never leaks into a fixture run.
var contractGetenvFn = os.Getenv

// newContractLineReader returns the confirmation reader over the command's
// input stream: one line per call, without the terminator. A final line
// without a newline is returned with a nil error; an empty stream is io.EOF.
var newContractLineReader = func(r io.Reader) func() (string, error) {
	br := bufio.NewReader(r)
	return func() (string, error) {
		line, err := br.ReadString('\n')
		if err != nil && (!errors.Is(err, io.EOF) || line == "") {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}
}

// contractAgentMarkers is the closed agent-environment marker set checked on
// the human signing path (design.md § Agent-Environment Markers).
func contractAgentMarkers() []string {
	return []string{config.EnvClaudeCode, config.EnvClaudeCodeSessionID}
}

// contractUsageError prints msg to stderr and returns the exit-2 error. The
// root error handler stays silent for an exit-coded error, so the message is
// printed here.
func contractUsageError(cmd *cobra.Command, format string, a ...any) error {
	msg := fmt.Sprintf(format, a...)
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "moai contract: "+msg)
	return &exitCodeError{code: contractExitUsage, msg: msg}
}

// contractEnv is everything the three subcommands read before acting.
type contractEnv struct {
	root          string
	autonomy      config.AutonomySettings
	ruleIDs       []string // constitution registry rule IDs
	frozenFiles   []string // distinct file: values of the registry's Frozen entries
	budgetDefault contract.Budget
}

// policy maps the effective configuration onto the verify policy.
func (e contractEnv) policy() contract.Policy {
	return contract.Policy{
		SecondReview:  e.autonomy.SecondReview,
		PushDevelop:   e.autonomy.PushDevelop,
		Mode:          e.autonomy.Mode,
		BudgetDefault: e.budgetDefault,
	}
}

// loadContractEnv resolves the project root, the effective workflow.autonomy
// settings, and the constitution registry values. Configuration warnings and
// the decider configuration error go to stderr; neither fails the command.
//
// @MX:ANCHOR: [AUTO] Single configuration read point of `moai contract`.
// @MX:REASON: sign, show, and verify all read their policy, registry values,
// and project root here, so the three subcommands cannot disagree about the
// effective workflow.autonomy values a contract is judged against.
func loadContractEnv(cmd *cobra.Command) (contractEnv, error) {
	root, err := findProjectRootFn()
	if err != nil {
		return contractEnv{}, contractUsageError(cmd, "%v", err)
	}
	errOut := cmd.ErrOrStderr()

	wf := config.NewDefaultWorkflowConfig()
	if cfg, lerr := config.NewLoader().Load(filepath.Join(root, ".moai")); lerr != nil {
		_, _ = fmt.Fprintf(errOut, "warning: configuration not loaded (%v); using defaults\n", lerr)
	} else if cfg != nil {
		wf = cfg.Workflow
	}
	s := config.ResolveAutonomy(wf)
	for _, w := range s.Warnings {
		_, _ = fmt.Fprintln(errOut, "warning: "+w)
	}
	if s.DeciderError != nil {
		_, _ = fmt.Fprintf(errOut, "configuration error: %v\n", s.DeciderError)
	}

	env := contractEnv{
		root:     root,
		autonomy: s,
		budgetDefault: contract.Budget{
			Turns:        s.BudgetDefault.Turns,
			Operations:   s.BudgetDefault.Operations,
			AuditRetries: s.BudgetDefault.AuditRetries,
		},
	}
	env.ruleIDs, env.frozenFiles = loadContractRegistry(errOut, root)
	return env, nil
}

// loadContractRegistry returns the constitution registry's rule IDs and the
// distinct Frozen-zone file: values. The registry is read from the project the
// contract belongs to (root) — or from MOAI_CONSTITUTION_REGISTRY when set —
// and not from CLAUDE_PROJECT_DIR, which names the primary checkout even for a
// command run inside a worktree. An absent or unreadable registry yields empty
// lists: `constitution:` invariants then report invariant_unresolved and the
// frozen-files set carries only its other two sources.
func loadContractRegistry(errOut io.Writer, root string) (ruleIDs, frozenFiles []string) {
	path := contractGetenvFn(constitution.RegistryPathEnv)
	if path == "" {
		path = filepath.Join(root, filepath.FromSlash(constitution.RegistryRelPath))
	}
	reg, err := constitution.LoadRegistry(path, root)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			_, _ = fmt.Fprintf(errOut, "warning: constitution registry not loaded (%v)\n", err)
		}
		return []string{}, []string{}
	}
	ruleIDs = make([]string, 0, len(reg.Entries))
	for _, r := range reg.Entries {
		ruleIDs = append(ruleIDs, r.ID)
	}
	frozenFiles = []string{}
	for _, r := range reg.FilterByZone(constitution.ZoneFrozen) {
		if !slices.Contains(frozenFiles, r.File) {
			frozenFiles = append(frozenFiles, r.File)
		}
	}
	slices.Sort(frozenFiles)
	return ruleIDs, frozenFiles
}

// verifyContract loads one SPEC's contract inputs and verifies them. The
// returned error is already printed and carries exit code 2.
func verifyContract(cmd *cobra.Command, env contractEnv, id string) (contract.Report, error) {
	dir, err := contract.ResolveSpecDir(env.root, id)
	if err != nil {
		return contract.Report{}, contractUsageError(cmd, "%v", err)
	}
	in, err := contract.LoadDir(dir)
	if err != nil {
		return contract.Report{}, contractUsageError(cmd, "%v", err)
	}
	in.Policy = env.policy()
	in.RegistryRuleIDs = env.ruleIDs
	in.RegistryFrozenFiles = env.frozenFiles
	if status, serr := spec.ParseStatus(dir); serr == nil {
		in.SpecStatus = status
	}
	return contract.Verify(in), nil
}

// writeContractJSON prints the report as one JSON object.
func writeContractJSON(w io.Writer, rep contract.Report) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(rep)
}

// reasonsLine renders the reason list for plain output.
func reasonsLine(reasons []string) string {
	if len(reasons) == 0 {
		return "reasons: none"
	}
	return "reasons: " + strings.Join(reasons, ", ")
}

func runContractVerify(cmd *cobra.Command, id string, asJSON bool) error {
	env, err := loadContractEnv(cmd)
	if err != nil {
		return err
	}
	rep, err := verifyContract(cmd, env, id)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if asJSON {
		if err := writeContractJSON(out, rep); err != nil {
			return contractUsageError(cmd, "write output: %v", err)
		}
	} else {
		_, _ = fmt.Fprintf(out, "%s: %s\n%s\n", rep.SpecID, rep.State, reasonsLine(rep.Reasons))
	}
	if !rep.Valid {
		return &exitCodeError{code: contractExitInvalid, msg: "contract " + rep.State}
	}
	return nil
}

func runContractShow(cmd *cobra.Command, id string, asJSON bool) error {
	env, err := loadContractEnv(cmd)
	if err != nil {
		return err
	}
	rep, err := verifyContract(cmd, env, id)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if asJSON {
		if err := writeContractJSON(out, rep); err != nil {
			return contractUsageError(cmd, "write output: %v", err)
		}
		return nil
	}
	printContractShow(out, rep)
	return nil
}

// contractSection is one top-level schema section of plain `show`.
type contractSection struct {
	name    string
	present bool
	value   any
}

// contractSections lists the schema sections of c in schema order.
func contractSections(c *contract.Contract) []contractSection {
	has := c.HasSection
	return []contractSection{
		{"schema_version", true, c.SchemaVersion},
		{"spec_id", c.SpecID != "", c.SpecID},
		{"card", c.Card != "", c.Card},
		{contract.SectionAcceptance, has(contract.SectionAcceptance), c.Acceptance},
		{contract.SectionInvariants, has(contract.SectionInvariants), c.Invariants},
		{contract.SectionOwnership, has(contract.SectionOwnership), c.Ownership},
		{contract.SectionApproach, has(contract.SectionApproach), c.Approach},
		{contract.SectionActions, has(contract.SectionActions), c.Actions},
		{contract.SectionReobserve, has(contract.SectionReobserve), c.Reobserve},
		{contract.SectionReview, has(contract.SectionReview), c.Review},
		{contract.SectionBudget, has(contract.SectionBudget), c.Budget},
		{contract.SectionEscalateOn, has(contract.SectionEscalateOn), c.EscalateOn},
		{contract.SectionPlanAudit, has(contract.SectionPlanAudit), c.PlanAudit},
		{contract.SectionSignature, c.Signature != nil, c.Signature},
	}
}

// contractSectionNames is the heading order when the contract did not decode.
var contractSectionNames = []string{
	"schema_version", "spec_id", "card", contract.SectionAcceptance, contract.SectionInvariants,
	contract.SectionOwnership, contract.SectionApproach, contract.SectionActions, contract.SectionReobserve,
	contract.SectionReview, contract.SectionBudget, contract.SectionEscalateOn, contract.SectionPlanAudit,
	contract.SectionSignature,
}

// printContractShow renders plain `show`: the signature state and reasons,
// every schema section under its heading, and the derived fields.
func printContractShow(out io.Writer, rep contract.Report) {
	p := func(format string, a ...any) { _, _ = fmt.Fprintf(out, format, a...) }
	p("%s (card %s)\n", rep.SpecID, orNone(rep.Card))
	p("state: %s\n%s\n", rep.State, reasonsLine(rep.Reasons))

	if rep.Contract == nil {
		for _, name := range contractSectionNames {
			p("\n== %s ==\n  (the contract does not decode)\n", name)
		}
	} else {
		for _, s := range contractSections(rep.Contract) {
			p("\n== %s ==\n", s.name)
			if !s.present {
				p("  (absent)\n")
				continue
			}
			p("%s", indentYAML(s.value))
		}
	}

	p("\n== derived ==\n")
	p("contract_sha256: %s\n", orNone(rep.ContractSHA256))
	p("recorded_contract_sha256: %s\n", orNone(rep.RecordedContractSHA256))
	p("signable_contract_sha256: %s\n", orNone(rep.SignableContractSHA256))
	p("acceptance: recorded %s (%s ACs), measured %s (%d ACs)\n",
		orNone(derefString(rep.Acceptance.SHA256)), derefCount(rep.Acceptance.ACCount),
		orNone(rep.Acceptance.MeasuredSHA256), rep.Acceptance.MeasuredACCount)
	p("push_requires_lease: %t\n", rep.PushRequiresLease)
	p("terminal: %t\n", rep.Terminal)
	p("effective_never: %s\n", listOrNone(rep.EffectiveNever))
	p("scratch: %s\n", listOrNone(rep.Scratch))
	p("frozen_files: %s\n", listOrNone(rep.FrozenFiles))
	p("second_review: %s\nmode: %s\n", rep.SecondReview, rep.Mode)
}

// indentYAML renders v as YAML indented by two spaces.
func indentYAML(v any) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		return "  (unrenderable: " + err.Error() + ")\n"
	}
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		b.WriteString("  " + line + "\n")
	}
	return b.String()
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}

func listOrNone(l []string) string {
	if len(l) == 0 {
		return "none"
	}
	return strings.Join(l, ", ")
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefCount(n *int) string {
	if n == nil {
		return "none"
	}
	return strconv.Itoa(*n)
}

// contractSignFlags are the `sign` flags.
type contractSignFlags struct {
	signer  string
	receipt string
	resign  bool
}

func runContractSign(cmd *cobra.Command, ids []string, f contractSignFlags) error {
	env, err := loadContractEnv(cmd)
	if err != nil {
		return err
	}
	s := env.autonomy
	opts := sign.Options{
		ProjectRoot:         env.root,
		SpecIDs:             ids,
		Signer:              f.signer,
		ReceiptPath:         f.receipt,
		Resign:              f.resign,
		Mode:                s.Mode,
		BatchSign:           s.BatchSign,
		SecondReview:        s.SecondReview,
		PushDevelop:         s.PushDevelop,
		Decider:             s.Decider,
		DeciderJevSole:      errors.Is(s.DeciderError, config.ErrKickoffDeciderJevSole),
		JevEnabled:          s.JevEnabled,
		JevMinConfidence:    s.JevMinConfidence,
		BudgetDefault:       env.budgetDefault,
		RegistryRuleIDs:     env.ruleIDs,
		RegistryFrozenFiles: env.frozenFiles,
		AgentMarkers:        contractAgentMarkers(),
	}
	seams := sign.Seams{
		IsTTY:    func() bool { return stdinIsTerminalFn() },
		Getenv:   contractGetenvFn,
		ReadLine: newContractLineReader(cmd.InOrStdin()),
		Out:      cmd.OutOrStdout(),
	}
	res, err := sign.Sign(opts, seams)
	if err != nil {
		if len(res.Written) > 0 || len(res.Unwritten) > 0 {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "written: %s\nunwritten: %s\n",
				listOrNone(res.Written), listOrNone(res.Unwritten))
		}
		return contractUsageError(cmd, "%v", err)
	}
	if res.Refusal != "" {
		return &exitCodeError{code: contractExitInvalid, msg: "refused " + res.Refusal}
	}
	return nil
}

// contractArgs validates the positional SPEC IDs as a usage error (exit 2).
func contractArgs(min, max int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		switch {
		case len(args) < min:
			return contractUsageError(cmd, "%s: missing SPEC ID", cmd.CommandPath())
		case max > 0 && len(args) > max:
			return contractUsageError(cmd, "%s: takes exactly one SPEC ID (%d given)", cmd.CommandPath(), len(args))
		}
		return nil
	}
}

// newContractCmd builds the `moai contract` command tree.
//
// @MX:NOTE: [AUTO] Exit codes are design.md § Exit codes: verify 0/1/2, show
// 0/2, sign 0/1/2. Usage and I/O errors are printed here and carried as an
// exit-coded error, which the root error handler does not print again.
func newContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract",
		Short: "Sign, show, and verify SPEC autonomy contracts",
		Long: `Sign, show, and verify a SPEC's autonomy contract
(.moai/specs/<SPEC-ID>/contract.yaml).

  verify <SPEC-ID> [--json]   check the contract; read-only, safe for hooks
  show <SPEC-ID> [--json]     print the sections, signature state, and derived sets
  sign <SPEC-ID>... [flags]   sign on the human path (interactive terminal and a
                              typed confirmation) or, with --signer llm|llm+jev
                              and --receipt, from a kickoff receipt

A signature does not yet replace Implementation Kickoff Approval.

Exit codes:
  verify  0 valid, 1 invalid, 2 usage or I/O error
  show    0 printed, 2 usage or I/O error
  sign    0 signed, 1 refused, 2 usage or I/O error`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return contractUsageError(c, "%v", err)
	})

	var verifyJSON bool
	verifyCmd := &cobra.Command{
		Use:           "verify <SPEC-ID>",
		Short:         "Verify a SPEC contract (exit 0 valid, 1 invalid, 2 usage or I/O error)",
		Args:          contractArgs(1, 1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			return runContractVerify(c, args[0], verifyJSON)
		},
	}
	verifyCmd.Flags().BoolVar(&verifyJSON, "json", false, "Print the verify object as JSON")

	var showJSON bool
	showCmd := &cobra.Command{
		Use:           "show <SPEC-ID>",
		Short:         "Show a SPEC contract, its signature state, and its derived sets",
		Args:          contractArgs(1, 1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			return runContractShow(c, args[0], showJSON)
		},
	}
	showCmd.Flags().BoolVar(&showJSON, "json", false, "Print the show object as JSON")

	var sf contractSignFlags
	signCmd := &cobra.Command{
		Use:           "sign <SPEC-ID>...",
		Short:         "Sign SPEC contracts (exit 0 signed, 1 refused, 2 usage or I/O error)",
		Args:          contractArgs(1, 0),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			return runContractSign(c, args, sf)
		},
	}
	signCmd.Flags().StringVar(&sf.signer, "signer", "",
		"Signer: human (default, interactive) or llm | llm+jev (the receipt's effective decider; requires --receipt)")
	signCmd.Flags().StringVar(&sf.receipt, "receipt", "",
		"Kickoff receipt, relative to the project root (.moai/specs/<SPEC-ID>/kickoff-receipt.json)")
	signCmd.Flags().BoolVar(&sf.resign, "resign", false,
		"Re-sign a signed contract after its acceptance.md changed (records supersedes)")

	cmd.AddCommand(verifyCmd, showCmd, signCmd)
	return cmd
}

func init() {
	rootCmd.AddCommand(newContractCmd())
}
