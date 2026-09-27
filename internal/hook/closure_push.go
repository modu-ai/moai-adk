package hook

// closure_push.go — the A4 hook checks (SPEC-AUTONOMY-CLOSURE-001):
//
//   - checkContractVerdict denies any Bash invocation of
//     `moai contract verdict` in EVERY mode (REQ-CLOSURE-021). It is a
//     string match on the command text with no I/O — a convenience guard
//     whose enforcement is the terminal check of REQ-CLOSURE-020
//     (spec.md §H).
//   - checkClosurePush denies a push of (or possibly of) the integration
//     branch under contract mode when any in-push contract is not ready
//     (REQ-CLOSURE-015..018). The mode check is the first statement: under
//     any other mode it returns before any file read or subprocess
//     (REQ-CLOSURE-023). The classifier fails closed — a destination that
//     cannot be proven different from the integration branch is
//     push_check_undetermined, and undetermined denies.
//
// Both run after checkBashCommand (the @MX:ANCHOR there forbids a
// conditional return above it) and before the Write/Edit block.

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/modu-ai/moai-adk/internal/closure"
	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/constitution"
)

// Sentinels (design.md §C.5). The orchestrator matches the prefixes.
const (
	ClosureVerdictDenyPrefix = "CLOSURE_VERDICT_HUMAN_ONLY:"
	ClosurePushStopPrefix    = "CLOSURE_PUSH_STOP:"
)

// closurePushEvalInvocations counts the push-readiness evaluations the hook
// actually ran (the subprocess seam of AC-CLOSURE-023: under guided the
// count stays at zero because the mode check returns first).
var closurePushEvalInvocations int

// checkContractVerdict denies a Bash call whose command text invokes
// `moai contract verdict`. Every mode, no I/O.
func checkContractVerdict(input *HookInput) (decision, reason string) {
	command := shellCommandText(input.ToolInput)
	if command == "" {
		return "", ""
	}
	if contractVerdictRe.MatchString(command) {
		return DecisionDeny, ClosureVerdictDenyPrefix +
			" `moai contract verdict` is reserved to a human on an interactive terminal (REQ-CLOSURE-020/021); run it at the operator's terminal."
	}
	return "", ""
}

// contractVerdictRe matches the verdict command in a shell command text
// (allowing the documented shell spellings; a wrapper evasion is the named
// accepted residual of spec.md §H).
var contractVerdictRe = regexp.MustCompile(`\bmoai\s+contract\s+verdict\b`)

// closurePushEvalResult is one check's outcome: an empty decision allows.
type closurePushEvalResult struct {
	decision string
	reason   string
}

// checkClosurePush evaluates push readiness for the Bash call under contract
// mode. The mode check is the first statement (REQ-CLOSURE-023): under
// guided — or any mode other than contract — it returns before reading a
// file or starting a subprocess.
func checkClosurePush(cfg ConfigProvider, input *HookInput) closurePushEvalResult {
	if cfg == nil || input == nil {
		return closurePushEvalResult{}
	}
	c := cfg.Get()
	if c == nil {
		return closurePushEvalResult{}
	}
	s := config.ResolveAutonomy(c.Workflow)
	if s.Mode != "contract" {
		return closurePushEvalResult{}
	}

	command := shellCommandText(input.ToolInput)
	if command == "" {
		return closurePushEvalResult{}
	}
	tree := resolveProjectRootFromInputOrEnv(input, "closure_push")
	integration := integrationBranchFor(tree)
	if integration == "" {
		return closurePushEvalResult{} // no integration branch to protect
	}

	pushes := classifyPushCommand(command, integration)
	var sources []classifiedPush
	for _, p := range pushes {
		if p.undetermined {
			return closurePushEvalResult{decision: DecisionDeny, reason: ClosurePushStopPrefix +
				" push_check_undetermined (" + p.cause + ")"}
		}
		if p.evaluate {
			if p.tree == "" {
				p.tree = tree
			}
			sources = append(sources, p)
		}
	}
	if len(sources) == 0 {
		return closurePushEvalResult{} // no push of the integration branch
	}

	registryIDs, registryFrozen := closureHookRegistry(tree)
	var findings []string
	for _, p := range sources {
		source, err := gitio.ResolveRef(p.tree, p.source)
		if err != nil {
			findings = append(findings, "push_check_undetermined (source "+p.source+" does not resolve)")
			continue
		}
		result := closure.EvaluatePush(closure.GitSeam{}, closure.PushEvalInput{
			Tree:                p.tree,
			Remote:              p.remote,
			Integration:         integration,
			Source:              source,
			SecondReview:        s.SecondReview,
			RegistryRuleIDs:     registryIDs,
			RegistryFrozenFiles: registryFrozen,
		})
		closurePushEvalInvocations++
		if result.Undetermined {
			findings = append(findings, "push_check_undetermined ("+result.Cause+")")
			continue
		}
		for _, cr := range result.Results {
			findings = append(findings, cr.SpecID+"="+strings.Join(cr.Codes, ","))
		}
	}
	if len(findings) == 0 {
		return closurePushEvalResult{}
	}
	return closurePushEvalResult{decision: DecisionDeny,
		reason: ClosurePushStopPrefix + " " + strings.Join(findings, "; ")}
}

// integrationBranchFor resolves the tree's integration branch. A config-
// orphaned worktree (no .moai/config of its own) takes the value from its
// primary checkout — the same fallback convention the audit-gate reader
// uses; a repository with no git-flow integration branch yields "" and the
// guard stays inert.
func integrationBranchFor(tree string) string {
	if v := config.LoadGitFlowDevelopBranch(tree); v != "" {
		return v
	}
	common, err := gitio.CommonDir(tree)
	if err != nil {
		return ""
	}
	return config.LoadGitFlowDevelopBranch(filepath.Dir(common))
}

// closureHookRegistry loads the tree's constitution registry values (the
// same registry `moai contract` reads; an absent registry yields empty
// lists, exactly as the CLI path).
func closureHookRegistry(root string) (ruleIDs, frozenFiles []string) {
	reg, err := constitution.LoadRegistry(root+"/"+constitution.RegistryRelPath, root)
	if err != nil {
		return []string{}, []string{}
	}
	ruleIDs = []string{}
	frozenFiles = []string{}
	for _, r := range reg.Entries {
		ruleIDs = append(ruleIDs, r.ID)
	}
	for _, r := range reg.FilterByZone(constitution.ZoneFrozen) {
		if !slices.Contains(frozenFiles, r.File) {
			frozenFiles = append(frozenFiles, r.File)
		}
	}
	slices.Sort(frozenFiles)
	return ruleIDs, frozenFiles
}

// ─── push classification (design.md §C.1, fail-closed) ───

// classifiedPush is one git push subcommand of the command text.
type classifiedPush struct {
	remote   string
	source   string // the source ref to resolve
	tree     string
	evaluate bool // the destination is (or resolves to) the integration branch
	// undetermined denies without evaluation (REQ-CLOSURE-017).
	undetermined bool
	cause        string
}

var (
	pushSegmentSplitRe = regexp.MustCompile(`&&|\|\||;|\n`)
	pushCandidateRe    = regexp.MustCompile(`\bgit\b[^;&|\n]*\bpush\b`)
	unprovableRe       = regexp.MustCompile(`\$\(|` + "`" + `|\beval\s|\b(?:sh|bash|zsh|dash)\s+-c\b|\$[A-Za-z_{]`)
)

// classifyPushCommand splits the command text into shell segments and
// classifies every git push among them against the caller-resolved
// integration branch. A command without a push candidate anywhere in its
// text is not an A4 subject at all (REQ-CLOSURE-015 judges Bash calls that
// push); when a candidate exists, an unprovable context anywhere in the
// command (command substitution, eval, a wrapper shell, a variable operand)
// makes the whole call undetermined.
func classifyPushCommand(command, integration string) []classifiedPush {
	if !pushCandidateRe.MatchString(command) {
		return nil
	}
	if unprovableRe.MatchString(command) {
		return []classifiedPush{{undetermined: true,
			cause: "command substitution, eval, a wrapper shell, or a variable operand"}}
	}
	var out []classifiedPush
	for _, segment := range pushSegmentSplitRe.Split(command, -1) {
		if p := classifyPushSegment(strings.TrimSpace(segment), integration); p != nil {
			out = append(out, *p)
		}
	}
	return out
}

// classifyPushSegment classifies one `git [-C tree] [-c k=v] push …`
// segment against the integration branch, or nil when the segment is not a
// git push.
func classifyPushSegment(segment, integration string) *classifiedPush {
	words := strings.Fields(segment)
	if len(words) < 2 || words[0] != "git" {
		return nil
	}
	i := 1
	segTree := "" // "" → the caller's tree
	for i < len(words) {
		switch words[i] {
		case "-C":
			if i+1 >= len(words) {
				return &classifiedPush{undetermined: true, cause: "-C without a path"}
			}
			segTree = words[i+1]
			i += 2
		case "-c":
			if i+2 > len(words) {
				return &classifiedPush{undetermined: true, cause: "-c without a key=value"}
			}
			i += 2 // -c <key>=<value>
		case "push":
			i++
			return classifyPushOperands(segTree, integration, words[i:])
		default:
			if strings.HasPrefix(words[i], "-") {
				i++
				continue
			}
			return nil // a non-push git subcommand
		}
	}
	return nil
}

// classifyPushOperands classifies the operands after `git push`.
func classifyPushOperands(segTree, integration string, words []string) *classifiedPush {
	p := &classifiedPush{tree: segTree, remote: "origin"}
	var operands []string
	for _, w := range words {
		switch {
		case w == "--all" || w == "--mirror":
			p.undetermined, p.cause = true, "--all/--mirror may include the integration branch"
			return p
		case strings.HasPrefix(w, "-"):
			continue // --tags, --force, -q, --set-upstream, …
		default:
			operands = append(operands, strings.TrimPrefix(w, "+"))
		}
	}

	// No operands: destination = the upstream of the current branch.
	if len(operands) == 0 {
		branch, err := currentBranchOf(p.tree)
		if err != nil || branch == "" {
			p.undetermined, p.cause = true, "bare push: current branch unresolvable"
			return p
		}
		upstream, err := upstreamBranchOf(p.tree, branch)
		if err != nil || upstream == "" {
			p.undetermined, p.cause = true, "bare push: no upstream for "+branch
			return p
		}
		name := upstream
		if j := strings.LastIndex(upstream, "/"); j >= 0 {
			name = upstream[j+1:]
			p.remote = upstream[:j]
		}
		if name == integration {
			p.evaluate = true
			p.source = branch
		}
		return p
	}

	// `<remote> <refspec>...` when the first operand is not refspec-shaped;
	// otherwise the operands are all refspecs.
	rest := operands
	if len(rest) > 1 && !strings.Contains(rest[0], ":") && !strings.HasPrefix(rest[0], "refs/") &&
		rest[0] != "HEAD" {
		p.remote = rest[0]
		rest = rest[1:]
	}
	for _, refspec := range rest {
		dst := refspec
		if j := strings.Index(refspec, ":"); j >= 0 {
			dst = refspec[j+1:]
		}
		dst = strings.TrimPrefix(dst, "refs/heads/")
		if dst == "" {
			p.undetermined, p.cause = true, "refspec "+refspec+" has an unresolvable destination"
			return p
		}
		if dst == integration {
			p.evaluate = true
			p.source = refspec
			if j := strings.Index(refspec, ":"); j >= 0 {
				p.source = refspec[:j]
			}
			return p
		}
	}
	return nil // every destination differs from the integration branch
}

// currentBranchOf and upstreamBranchOf run git in the segment's tree; ""
// tree means the caller's tree (the hook resolved it already).
func currentBranchOf(tree string) (string, error) {
	dir := tree
	if dir == "" {
		dir = "."
	}
	return gitio.CurrentBranch(dir)
}

func upstreamBranchOf(tree, branch string) (string, error) {
	dir := tree
	if dir == "" {
		dir = "."
	}
	return gitio.UpstreamBranch(dir, branch)
}

// shellCommandText reads the shell command from the tool input.
func shellCommandText(toolInput json.RawMessage) string {
	if len(toolInput) == 0 {
		return ""
	}
	var in struct {
		Command string `json:"command"`
	}
	if json.Unmarshal(toolInput, &in) != nil {
		return ""
	}
	return in.Command
}
