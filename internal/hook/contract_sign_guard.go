package hook

// contract_sign_guard.go — the contract-sign and contract-decide guard
// (SPEC-AUTONOMY-PRECONDITION-001 REQ-AP-003 / REQ-AP-004 / REQ-AP-005 /
// REQ-AP-009 / REQ-AP-011 / REQ-AP-012; design.md §C).
//
// A contract signed by an agent is worthless in any autonomy mode, and
// `moai contract decide` is the lead session's own path when the decider is
// `llm` or `llm+jev` — so the two verbs are not one rule but three
// (design.md §C.2 step 6):
//
//	HUMAN PATH  `sign` with no --signer or --signer human → deny in every
//	            session, with no exemption of any kind (REQ-AP-003). Signing
//	            happens at an operator terminal; the tool-call boundary is
//	            the one signal an agent cannot unset. The deny is
//	            mode-independent — active under `guided` too (design.md
//	            §C.7).
//	ROLE GATE   `sign --signer llm` / `--signer llm+jev`, and `decide` →
//	            deny only where the calling session's MOAI_FACTORY_ROLE
//	            equals the role value constant; allowed otherwise
//	            (REQ-AP-011). The allow direction is a requirement: it is
//	            the lead's own decide path, and denying it would deny the
//	            caller the epic depends on.
//	FAIL CLOSED an invocation carrying `contract` together with `sign` or
//	            `decide` whose structure cannot be classified — command
//	            substitution, a variable in program position, eval, an
//	            unenumerated wrapper, or nesting deeper than one -c level —
//	            is denied under the rule in force (REQ-AP-004): `sign`
//	            everywhere, `decide` only under the marker, because denying
//	            a no-marker session's decide on the strength of a parse
//	            failure would trade a real path for a hypothetical one.
//
// The deny does not depend on the verb existing in the installed binary:
// PreToolUse runs before execution (design.md §C.4). The guard reads no
// receipt: `--receipt` travels through as an opaque argument it never
// inspects (design.md §C.5). It fails CLOSED — a wrongly allowed signature
// voids the contract model, while a wrongly denied sign costs the operator
// one terminal command, which is where signing belongs anyway (design.md
// §C.6). The guard reads one environment variable, the REQ-AP-012 constant,
// and no harness-presence marker beyond it (REQ-AP-009), no project state,
// and no record; an allowed call leaves the hook output byte-identical to
// the no-guard baseline and writes no audit line.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// contractSignViolationPrefix is the deny sentinel the orchestrator matches.
const contractSignViolationPrefix = "CONTRACT_SIGN_AGENT_VIOLATION:"

// contractMaxDashCLevels is the depth of -c indirection the guard resolves.
// Exactly one level; deeper nesting is unclassifiable (REQ-AP-004).
const contractMaxDashCLevels = 1

// contractSignVerdict is the parse outcome for one command line.
type contractSignVerdict struct {
	found      bool   // a contract sign/decide invocation pattern is carried
	verb       string // "sign" or "decide"
	signer     string // --signer value, "" when absent
	classified bool   // false → the invocation is unclassifiable (fail closed)
}

// checkContractSign returns DecisionDeny plus a sentinel-prefixed reason when
// the command invokes `moai contract sign` or `moai contract decide` under a
// deny rule in force; otherwise ("", ""). Independent of
// workflow.autonomy.mode (REQ-AP-003 / REQ-AP-010).
func checkContractSign(input *HookInput) (decision string, reason string) {
	if input == nil || len(input.ToolInput) == 0 {
		return "", ""
	}
	command := extractIntegrationCommand(input.ToolInput)
	if command == "" {
		return "", ""
	}
	v := classifyContractCall(command)
	if !v.found {
		return "", ""
	}

	deny := func() (string, string) {
		if v.classified {
			return DecisionDeny, fmt.Sprintf(
				"%s program \"moai\": an agent invocation of `moai contract %s` is denied at the tool-call boundary. "+
					"Human signing happens at an operator terminal; run the command there.",
				contractSignViolationPrefix, v.verb)
		}
		return DecisionDeny, fmt.Sprintf(
			"%s unclassified invocation carrying `moai contract %s` — the structure cannot be classified, so it is denied "+
				"fail-closed under the rule in force (REQ-AP-004). Run the command from an operator terminal.",
			contractSignViolationPrefix, v.verb)
	}

	switch v.verb {
	case "sign":
		switch v.signer {
		case "", "human":
			// The human signing path: denied in every session, no exemption
			// (REQ-AP-003). --receipt is never read (design.md §C.5).
			return deny()
		case "llm", "llm+jev":
			// The non-interactive path: gated on the role marker (REQ-AP-011).
			if !contractRoleMarker() {
				return "", ""
			}
			return deny()
		default:
			// A signer value outside A1's closed decider set (human | llm |
			// llm+jev) cannot be shown to be off the agent paths — fail
			// closed (design.md §C.6).
			v.classified = false
			return deny()
		}
	default: // decide
		if !contractRoleMarker() {
			return "", ""
		}
		return deny()
	}
}

// contractRoleMarker reports whether the calling session claims the factory
// worker role (REQ-AP-011). The marker's name and value come from the
// internal/config constants REQ-AP-012 defines, so the retired alias `agent`
// is not accepted here by construction. A session that sets no marker makes
// no role claim.
func contractRoleMarker() bool {
	return os.Getenv(config.EnvFactoryRole) == config.FactoryRoleWorker
}

// classifyContractCall parses one command line per design.md §C.2: shell-word
// splitting with quote removal (not quoted-span scrubbing, which would erase
// 'moai'), leading NAME=value stripping, the closed prefix-wrapper set of
// REQ-AP-005, a basename program match, global flags before the verb, and
// exactly one level of sh -c / bash -c / zsh -c indirection.
func classifyContractCall(command string) contractSignVerdict {
	return classifyContractWords(shellSplitWords(command), command, 0)
}

// classifyContractWords classifies one parse level. raw is the text the words
// were split from, needed for the command-substitution test; depth is the -c
// recursion level already consumed.
func classifyContractWords(words []string, raw string, depth int) contractSignVerdict {
	var none contractSignVerdict

	// Command substitution anywhere makes the structure unclassifiable while
	// a deny-eligible call is carried (REQ-AP-004).
	if strings.Contains(raw, "$(") || strings.Contains(raw, "`") {
		if verb := carriedVerb(words); verb != "" {
			return contractSignVerdict{found: true, verb: verb}
		}
		return none
	}

	// Strip leading NAME=value assignments.
	i := 0
	for i < len(words) && isShellAssignment(words[i]) {
		i++
	}

	// Strip the closed prefix-wrapper set (REQ-AP-005), skipping each
	// wrapper's own options so they are never read as the program word.
wrapperLoop:
	for i < len(words) {
		switch words[i] {
		case "eval":
			// eval re-parses its arguments as a command; the structure
			// cannot be classified (REQ-AP-004).
			rest := shellSplitWords(strings.Join(words[i+1:], " "))
			if verb := carriedVerb(rest); verb != "" {
				return contractSignVerdict{found: true, verb: verb}
			}
			return none
		case "env":
			i = skipEnvOptions(words, i)
		case "command":
			i = skipCommandOptions(words, i)
		case "exec":
			i = skipExecOptions(words, i)
		case "nohup":
			i++
		case "script":
			// script -c CMD runs CMD as a command string — parse it once
			// more (one level, the fail-closed direction); without -c, skip
			// -q and the typescript file operand.
			j := i + 1
			for j < len(words) && words[j] == "-q" {
				j++
			}
			if j < len(words) && words[j] == "-c" && j+1 < len(words) {
				if depth >= contractMaxDashCLevels {
					if verb := carriedVerb(words); verb != "" {
						return contractSignVerdict{found: true, verb: verb}
					}
					return none
				}
				return classifyContractWords(shellSplitWords(words[j+1]), words[j+1], depth+1)
			}
			if j < len(words) && !strings.HasPrefix(words[j], "-") {
				j++
			}
			i = j
		case "timeout":
			i = skipTimeoutOptions(words, i)
		case "sudo":
			i = skipSudoOptions(words, i)
		case "stdbuf":
			i = skipStdbufOptions(words, i)
		case "nice":
			i = skipNiceOptions(words, i)
		case "xargs":
			i = skipXargsOptions(words, i)
		default:
			break wrapperLoop
		}
	}

	if i >= len(words) {
		return none
	}
	program := words[i]

	// A variable in program position (REQ-AP-004).
	if strings.HasPrefix(program, "$") {
		if verb := carriedVerb(words); verb != "" {
			return contractSignVerdict{found: true, verb: verb}
		}
		return none
	}

	base := filepath.Base(program)
	rest := words[i+1:]

	if base == "moai" {
		verb, signer, ok := matchMoaiContractCall(rest)
		if !ok {
			return none
		}
		return contractSignVerdict{found: true, verb: verb, signer: signer, classified: true}
	}

	if base == "sh" || base == "bash" || base == "zsh" {
		if payload, ok := shellDashCPayload(rest); ok {
			if depth >= contractMaxDashCLevels {
				// Nesting beyond one -c level (REQ-AP-004).
				if verb := carriedVerb(words); verb != "" {
					return contractSignVerdict{found: true, verb: verb}
				}
				return none
			}
			return classifyContractWords(shellSplitWords(payload), payload, depth+1)
		}
		return none
	}

	// An unenumerated word in program position (REQ-AP-004): the structure is
	// unclassifiable only while the deny-eligible call pattern follows as
	// separate words. A program that carries the call as quoted data —
	// `echo "moai contract sign"`, `git commit -m "sign the contract"` — is
	// classified, and after quote removal the call is not a word sequence.
	if verb := scanCallWords(rest); verb != "" {
		return contractSignVerdict{found: true, verb: verb}
	}
	return none
}

// matchMoaiContractCall matches `contract` followed by `sign`|`decide` in the
// words after the program word, reading the --signer value from the same
// word list (design.md §C.2 step 6 — no separate parser for it). Global
// flags between the program word and the verb are simply other words in the
// list; `--receipt` is opaque and skipped by not being inspected.
func matchMoaiContractCall(words []string) (verb, signer string, ok bool) {
	for i := 0; i+1 < len(words); i++ {
		if words[i] == "--signer" && i+1 < len(words) {
			signer = words[i+1]
		} else if v, found := strings.CutPrefix(words[i], "--signer="); found {
			signer = v
		}
	}
	for i := 0; i+1 < len(words); i++ {
		if words[i] == "contract" && (words[i+1] == "sign" || words[i+1] == "decide") {
			return words[i+1], signer, true
		}
	}
	return "", "", false
}

// scanCallWords reports which deny-eligible verb the words carry as
// `contract` immediately followed by the verb — `sign` in priority, because
// the human-path rule REQ-AP-003 is always in force. "" when neither.
func scanCallWords(words []string) string {
	hasDecide := false
	for i := 0; i+1 < len(words); i++ {
		if words[i] != "contract" {
			continue
		}
		switch words[i+1] {
		case "sign":
			return "sign"
		case "decide":
			hasDecide = true
		}
	}
	if hasDecide {
		return "decide"
	}
	return ""
}

// carriedVerb finds the deny-eligible verb an UNCLASSIFIABLE command carries:
// the parsed words first, then the whitespace fields of each word's content —
// eval's argument and a payload reached past one -c level arrive as single
// words whose contents still carry the call. Used only on the unclassified
// path, where the classification is already decided and only the rule in
// force is being scoped; never on the classified path, where a quoted string
// must not read as a call (AC-AP-006's `echo "moai contract sign"`).
func carriedVerb(words []string) string {
	if verb := scanCallWords(words); verb != "" {
		return verb
	}
	var flat []string
	for _, w := range words {
		// Substitution delimiters attach to adjacent tokens in the word
		// contents ("sign`") and would defeat the adjacency scan; strip them
		// before fielding.
		flat = append(flat, strings.Fields(strings.NewReplacer("`", "", ")", "").Replace(w))...)
	}
	return scanCallWords(flat)
}

// shellDashCPayload extracts the command string of `sh|bash|zsh -c <string>`:
// options are skipped and the first non-flag word is the payload.
func shellDashCPayload(words []string) (string, bool) {
	for i := 0; i < len(words); i++ {
		if words[i] == "-c" {
			if i+1 < len(words) {
				return words[i+1], true
			}
			return "", false
		}
		if !strings.HasPrefix(words[i], "-") {
			return "", false
		}
	}
	return "", false
}

// isShellAssignment reports whether the word is a NAME=value environment
// assignment: leading assignments are stripped, and env's own NAME=value
// arguments are skipped the same way.
func isShellAssignment(word string) bool {
	name, _, found := strings.Cut(word, "=")
	if !found || name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c == '_', c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z':
		case i > 0 && c >= '0' && c <= '9':
		default:
			return false
		}
	}
	return true
}

// The wrapper option skippers below implement the closed prefix-wrapper set
// of design.md §C.3 (REQ-AP-005). Each returns the index of the first word
// after the wrapper's own options — the candidate program word. Skipping an
// option wrongly is how a wrapper bypass survives a matcher that claims to
// handle it, so each row implements only the shapes its §C.3 entry names.

// skipEnvOptions skips env's NAME=value pairs, -i, and -u NAME.
func skipEnvOptions(words []string, i int) int {
	j := i + 1
	for j < len(words) {
		w := words[j]
		switch {
		case isShellAssignment(w), w == "-i":
			j++
		case w == "-u" && j+1 < len(words):
			j += 2
		case len(w) > 2 && strings.HasPrefix(w, "-u"):
			j++
		default:
			return j
		}
	}
	return j
}

// skipCommandOptions skips command's -p, -v, and -V.
func skipCommandOptions(words []string, i int) int {
	j := i + 1
	for j < len(words) && (words[j] == "-p" || words[j] == "-v" || words[j] == "-V") {
		j++
	}
	return j
}

// skipExecOptions skips exec's -a NAME, -c, and -l (and glued pure forms
// such as -lc).
func skipExecOptions(words []string, i int) int {
	j := i + 1
	for j < len(words) {
		w := words[j]
		switch {
		case w == "-a" && j+1 < len(words):
			j += 2
		case w == "-c" || w == "-l":
			j++
		case len(w) > 2 && strings.HasPrefix(w, "-") && strings.Trim(w, "-acl") == "":
			j++
		default:
			return j
		}
	}
	return j
}

// skipTimeoutOptions skips timeout's -k DURATION, -s SIG, --preserve-status,
// and the leading duration operand.
func skipTimeoutOptions(words []string, i int) int {
	j := i + 1
	for j < len(words) {
		w := words[j]
		switch {
		case w == "--preserve-status":
			j++
		case (w == "-k" || w == "-s") && j+1 < len(words):
			j += 2
		case len(w) > 2 && (strings.HasPrefix(w, "-k") || strings.HasPrefix(w, "-s")):
			j++
		case !strings.HasPrefix(w, "-"):
			return j + 1
		default:
			return j
		}
	}
	return j
}

// skipSudoOptions skips sudo's -n, -u USER, -E, -H, and -- (plus glued pure
// forms of the valueless flags, such as -nE).
func skipSudoOptions(words []string, i int) int {
	j := i + 1
	for j < len(words) {
		w := words[j]
		switch {
		case w == "-n", w == "-E", w == "-H", w == "--":
			j++
		case w == "-u" && j+1 < len(words):
			j += 2
		case len(w) > 2 && strings.HasPrefix(w, "-u"):
			j++
		case len(w) > 2 && strings.HasPrefix(w, "-") && strings.Trim(w, "-nEH") == "":
			j++
		default:
			return j
		}
	}
	return j
}

// skipStdbufOptions skips stdbuf's -i/-o/-e MODE, in split and glued forms
// (`-oL`).
func skipStdbufOptions(words []string, i int) int {
	j := i + 1
	for j < len(words) {
		w := words[j]
		switch {
		case (w == "-i" || w == "-o" || w == "-e") && j+1 < len(words):
			j += 2
		case len(w) > 2 && (strings.HasPrefix(w, "-i") || strings.HasPrefix(w, "-o") || strings.HasPrefix(w, "-e")):
			j++
		default:
			return j
		}
	}
	return j
}

// skipNiceOptions skips nice's -n N, in split and glued forms.
func skipNiceOptions(words []string, i int) int {
	j := i + 1
	for j < len(words) {
		w := words[j]
		switch {
		case w == "-n" && j+1 < len(words):
			j += 2
		case len(w) > 2 && strings.HasPrefix(w, "-n"):
			j++
		default:
			return j
		}
	}
	return j
}

// skipXargsOptions skips xargs's -n N, -0, -I STR, and -P N, in split and
// glued forms (`-n1`).
func skipXargsOptions(words []string, i int) int {
	j := i + 1
	for j < len(words) {
		w := words[j]
		switch {
		case w == "-0":
			j++
		case (w == "-n" || w == "-P" || w == "-I") && j+1 < len(words):
			j += 2
		case len(w) > 2 && (strings.HasPrefix(w, "-n") || strings.HasPrefix(w, "-P") || strings.HasPrefix(w, "-I")):
			j++
		default:
			return j
		}
	}
	return j
}

// shellSplitWords splits a command into shell words with QUOTE REMOVAL —
// the design.md §C.2 step-1 requirement. Single-quoted spans are literal;
// double-quoted spans honor backslash escapes; a backslash outside quotes
// quotes the next character; $() and backtick substitution regions are kept
// as single words (the guard tests for their presence separately, per
// REQ-AP-004, rather than trying to classify their contents).
func shellSplitWords(command string) []string {
	var words []string
	var sb strings.Builder
	inWord := false
	flush := func() {
		if inWord {
			words = append(words, sb.String())
			sb.Reset()
			inWord = false
		}
	}
	i := 0
	for i < len(command) {
		c := command[i]
		switch {
		case c == '\'':
			inWord = true
			j := i + 1
			for j < len(command) && command[j] != '\'' {
				j++
			}
			sb.WriteString(command[i+1 : j])
			i = j + 1
		case c == '"':
			inWord = true
			i++
			for i < len(command) && command[i] != '"' {
				if command[i] == '\\' && i+1 < len(command) &&
					(command[i+1] == '"' || command[i+1] == '\\' || command[i+1] == '$') {
					sb.WriteByte(command[i+1])
					i += 2
					continue
				}
				sb.WriteByte(command[i])
				i++
			}
			i++
		case c == '\\':
			inWord = true
			if i+1 < len(command) {
				sb.WriteByte(command[i+1])
				i += 2
			} else {
				i++
			}
		case c == '$' && i+1 < len(command) && command[i+1] == '(':
			inWord = true
			depth := 0
			j := i + 1
			for j < len(command) {
				if command[j] == '(' {
					depth++
				} else if command[j] == ')' {
					depth--
					if depth == 0 {
						break
					}
				}
				j++
			}
			if j < len(command) {
				j++
			}
			sb.WriteString(command[i:j])
			i = j
		case c == '`':
			inWord = true
			j := i + 1
			for j < len(command) && command[j] != '`' {
				j++
			}
			sb.WriteByte('`')
			sb.WriteString(command[i+1 : j])
			if j < len(command) {
				sb.WriteByte('`')
				j++
			}
			i = j
		case c == ' ' || c == '\t' || c == '\n':
			flush()
			i++
		default:
			inWord = true
			sb.WriteByte(c)
			i++
		}
	}
	flush()
	return words
}
