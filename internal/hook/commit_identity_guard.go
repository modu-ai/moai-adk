package hook

// commit_identity_guard.go — the commit-time test-identity guard
// (SPEC-COMMIT-IDENTITY-GUARD-001).
//
// Root cause: a test wrote a fixture identity (`t <t@t.t>`) into the git
// config layer every lane session shares, and 5,483 develop commits landed
// under it before the layer was overwritten by accident (measured in
// .moai/reports/t1289/root-cause.md). This guard is the last line of defense
// at the commit moment: a PreToolUse layer that refuses a shell command whose
// RESOLVED identity — the identity git would actually use, configuration
// layers plus command-level overrides — matches a known test-fixture email.
// History rewriting is out of scope; the guard blocks future commits only.
//
// Fail-open norm (spec.md §6): the deny fires ONLY on positive evidence — an
// exact, trimmed, case-insensitive email match against the effective deny
// list. Every uncertainty (probe error, probe timeout, unparseable probe
// output, missing cwd, unresolved repository scope) allows the command and
// appends one advisory line to .moai/logs/commit-identity-guard-audit.log,
// the same advisory convention as branch-guard-audit.log. A command whose
// target repository does not share this project's git common dir (a /tmp
// fixture repository, for instance) is legitimate test activity and is
// allowed without any identity probe (REQ-CIG-012).

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// testIdentityViolationPrefix opens every deny reason this guard emits, so
// the orchestrator can match the source without parsing the rest of the
// reason (the sibling sentinel convention).
const testIdentityViolationPrefix = "TEST_IDENTITY_VIOLATION:"

// commitIdentityAuditRelPath is the guard's advisory audit log, relative to
// the handler's project root.
const commitIdentityAuditRelPath = ".moai/logs/commit-identity-guard-audit.log"

// commitIdentityProbeTimeout bounds each probe subprocess (plan.md §F M2
// design default of 2s; an overrun is a REQ-CIG-007 failure, never a deny).
// A variable rather than a const only so the timeout path can be exercised
// by a test with a shrunken bound.
var commitIdentityProbeTimeout = 2 * time.Second

// commitVerbs are the git verbs that can create a commit. merge (non-ff) and
// pull (its merge path) are included because they may commit; judging them
// costs one probe for a real identity and allows (spec.md §4 D2).
var commitVerbs = map[string]bool{
	"commit":      true,
	"merge":       true,
	"cherry-pick": true,
	"revert":      true,
	"rebase":      true,
	"am":          true,
	"commit-tree": true,
	"pull":        true,
}

// builtinCommitIdentityDenyEmails is the built-in fixture-email list: the
// complete enumeration of the fixture email literals this repository's
// *_test.go files assign through user.email, GIT_AUTHOR_EMAIL, or
// GIT_COMMITTER_EMAIL (plan.md §B.1 measurement, re-run at the run base —
// 51 literals). internal/hook/commit_identity_guard*_test.go files are
// EXCLUDED from the enumeration because they carry out-of-list control
// values; commit_identity_guard_list_test.go re-runs the same predicate and
// fails when a literal is missing from this list (AC-CIG-010).
var builtinCommitIdentityDenyEmails = []string{
	"a@e.invalid",
	"anchor-test@example.com",
	"audit@example.invalid",
	"board-test@example.com",
	"branch-guard-test@example.com",
	"c@e.invalid",
	"disposal-test@example.com",
	"f@e.com",
	"f@example.com",
	"f3@example.com",
	"f4@example.com",
	"fixture@example.com",
	"fixture@example.invalid",
	"fixture@example.test",
	"fixture@t516.invalid",
	"fx@example.com",
	"guard-test@example.invalid",
	"m2-test@example.com",
	"o@e.x",
	"other@example.com",
	"slot-cli-test@example.com",
	"slot-lease-test@example.com",
	"someone@example.invalid",
	"stopchain-test@example.com",
	"sub@e.x",
	"sync-gate-test@example.com",
	"t@example.com",
	"t@example.invalid",
	"t@local",
	"t@t",
	"t@t.local",
	"t@t.t",
	"t@t.test",
	"t1074@example.invalid",
	"t1204@example.invalid",
	"t371@example.com",
	"t461@example.test",
	"t488@example.invalid",
	"t516@test.invalid",
	"t560@test.invalid",
	"t602@example.invalid",
	"t603@example.invalid",
	"t606@example.test",
	"t655@test.local",
	"t688@example.test",
	"t766@example.invalid",
	"t768@example.test",
	"test@example.com",
	"test@example.invalid",
	"test@test.com",
	"tier-guard-test@example.com",
}

// effectiveDenyEmails returns the union of the built-in list and extra
// (workflow.commit_identity_guard.deny_emails). The config list ADDS to the
// built-in list; it never shrinks it (REQ-CIG-008).
func effectiveDenyEmails(extra []string) []string {
	out := make([]string, 0, len(builtinCommitIdentityDenyEmails)+len(extra))
	out = append(out, builtinCommitIdentityDenyEmails...)
	out = append(out, extra...)
	return out
}

// isDenyListed reports whether email exactly matches a list entry after
// trimming surrounding whitespace, case-insensitively. No pattern, no
// heuristic (spec.md §4 D4).
func isDenyListed(email string, list []string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	for _, e := range list {
		if strings.EqualFold(strings.TrimSpace(e), email) {
			return true
		}
	}
	return false
}

// --- seams ---

// commitIdentityScopeProbe returns the target directory's git common dir
// (absolute). The default resolves through revParseDirs, which runs
// `git rev-parse` confined by gitenv-scrubbed environment (a stray GIT_DIR in
// the hook process cannot redirect the read to another repository).
var commitIdentityScopeProbe = func(dir string) (string, error) {
	_, common, err := revParseDirs(dir)
	return common, err
}

// commitIdentityProbeEnv is the environment handed to the identity probe
// child. Repository-scoping variables are scrubbed (gitenv.Scrub) so a stray
// GIT_DIR cannot point the probe at another repository; identity variables
// (GIT_AUTHOR_*/GIT_COMMITTER_*/EMAIL) are deliberately PRESERVED — they are
// part of the identity git would use (spec.md §6).
var commitIdentityProbeEnv = func() []string {
	return gitenv.Scrub(os.Environ())
}

// commitIdentityVarProbe resolves the default author and committer idents
// git would use in dir, as `Name <email> <epoch> <tz>` lines. env is the
// probe child's environment, passed as a parameter so tests can inject an
// identity into the probe child ONLY (plan.md §D).
var commitIdentityVarProbe = func(dir string, env []string) (author, committer string, err error) {
	if author, err = gitVarIdent(dir, env, "GIT_AUTHOR_IDENT"); err != nil {
		return "", "", err
	}
	if committer, err = gitVarIdent(dir, env, "GIT_COMMITTER_IDENT"); err != nil {
		return "", "", err
	}
	return author, committer, nil
}

// gitVarIdent runs `git var <name>` in dir with a time bound.
func gitVarIdent(dir string, env []string, name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commitIdentityProbeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "var", name)
	cmd.Env = env
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("git var %s: probe timed out after %s", name, commitIdentityProbeTimeout)
		}
		return "", fmt.Errorf("git var %s: %w", name, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// emailFromIdent extracts the email from a `Name <email> epoch tz` ident.
// Output without angle brackets is a resolution failure (REQ-CIG-007).
func emailFromIdent(ident string) (string, error) {
	i := strings.Index(ident, "<")
	j := strings.LastIndex(ident, ">")
	if i < 0 || j <= i {
		return "", fmt.Errorf("unparseable git ident output %q", ident)
	}
	return strings.TrimSpace(ident[i+1 : j]), nil
}

// --- command segmentation ---

// shellSegment is one top-level command segment of a shell call, with the
// connector that preceded it ("" for the first segment). Heredoc bodies and
// shell comments are excluded: both are text the shell never executes as a
// command, so they can neither trigger the guard nor supply overrides.
type shellSegment struct {
	text string
	conn string // connector preceding this segment: "", "&&", "||", ";", "|", "\n"
}

// splitShellSegments splits a shell command into top-level segments,
// respecting single/double quotes, heredoc bodies, and shell comments. The
// heredoc rules mirror substituteHeredocBodies (openers outside quoted
// spans register a delimiter; a body line matching the oldest pending
// delimiter terminates it); the comment rule mirrors shellCommentStart
// (word-start `#` opens a comment that ends the line's command text).
func splitShellSegments(command string) []shellSegment {
	var segs []shellSegment
	var cur strings.Builder
	conn := "" // connector preceding the segment being built
	flush := func() {
		if s := strings.TrimSpace(cur.String()); s != "" {
			segs = append(segs, shellSegment{text: s, conn: conn})
		}
		cur.Reset()
	}
	var pending []string // heredoc delimiters opened, in order
	for _, line := range strings.Split(command, "\n") {
		if len(pending) > 0 {
			// Inside a heredoc body: the terminator is the oldest pending
			// delimiter alone on its line. Neither body nor terminator is
			// command text.
			if strings.TrimSpace(line) == pending[0] {
				pending = pending[1:]
			}
			continue
		}
		quoted := quotedArgumentPattern.FindAllStringIndex(line, -1)
		for _, m := range heredocOpenerPattern.FindAllStringSubmatchIndex(line, -1) {
			if insideAnySpan(m[0], quoted) {
				continue // a `<<EOF` inside a quoted span is data, not a redirect
			}
			for g := 1; g <= 3; g++ {
				if m[2*g] >= 0 {
					pending = append(pending, line[m[2*g]:m[2*g+1]])
					break
				}
			}
		}
		// Walk the line cutting segments at top-level separators. Every byte
		// is written to the current segment verbatim — the segment text is
		// ORIGINAL text (quote characters included) because override and
		// path extraction read it.
		inSingle, inDouble := false, false
		for i := 0; i < len(line); {
			c := line[i]
			switch {
			case inSingle:
				cur.WriteByte(c)
				if c == '\'' {
					inSingle = false
				}
				i++
			case inDouble:
				if c == '\\' && i+1 < len(line) {
					cur.WriteByte(c)
					cur.WriteByte(line[i+1])
					i += 2
					continue
				}
				cur.WriteByte(c)
				if c == '"' {
					inDouble = false
				}
				i++
			case c == '\'':
				inSingle = true
				cur.WriteByte(c)
				i++
			case c == '"':
				inDouble = true
				cur.WriteByte(c)
				i++
			case c == '\\':
				cur.WriteByte(c)
				if i+1 < len(line) {
					cur.WriteByte(line[i+1])
				}
				i += 2 // a trailing backslash line continuation is an accepted residual
			case c == '#':
				if i == 0 || strings.ContainsRune(" \t;&|(", rune(line[i-1])) {
					i = len(line) // comment: the rest of the line is not command text
				} else {
					cur.WriteByte(c)
					i++
				}
			case c == '&' && i+1 < len(line) && line[i+1] == '&':
				flush()
				conn = "&&"
				i += 2
			case c == '&':
				cur.WriteByte(c)
				i++
			case c == '|' && i+1 < len(line) && line[i+1] == '|':
				flush()
				conn = "||"
				i += 2
			case c == '|':
				flush()
				conn = "|"
				i++
			case c == ';':
				flush()
				conn = ";"
				i++
			default:
				cur.WriteByte(c)
				i++
			}
		}
		flush() // end of line: a newline separates commands
		conn = "\n"
	}
	flush()
	return segs
}

// --- trigger detection (spec.md §4 D2) ---

// normalizeSegmentForTrigger runs the branch guard's collapse pipeline over
// one segment: quoted spans collapse to a placeholder (so text carried as
// data never reads as a command), PowerShell call-operator targets and
// command-position backtick escapes de-escape, and the .exe-suffixed git
// spelling normalizes onto the plain one. Heredoc bodies and comments are
// already excluded by splitShellSegments.
func normalizeSegmentForTrigger(text string) string {
	return normalizeGitExeSuffix(substituteCommandBackticks(
		substituteQuotedArguments(substituteCallOperatorTargets(text))))
}

// maskedTokenPlaceholder is the token a quoted span collapses to.
const maskedTokenPlaceholder = "X"

// findCommitTrigger returns the commit-creating verb a segment invokes, or ""
// when it invokes none. A git token counts only in command position: every
// preceding token must be an env assignment (NAME=VALUE) or a transparent
// wrapper (export/env/sudo/command/nice/nohup/time) — `echo git commit` is
// prose, not a git invocation. Git global options (-C, -c, --git-dir, …) are
// skipped to reach the verb; under-matching an unclassifiable form is the
// accepted fail-open direction.
func findCommitTrigger(text string) string {
	scanned := normalizeSegmentForTrigger(text)
	tokens := strings.Fields(scanned)
	for i := range tokens {
		if !isGitCommandToken(tokens, i) {
			continue
		}
		if verb, ok := gitVerbAfterGlobals(tokens[i+1:]); ok && commitVerbs[verb] {
			return verb
		}
	}
	return ""
}

// isGitCommandToken reports whether tokens[i] names the git executable in
// command position.
func isGitCommandToken(tokens []string, i int) bool {
	tok := tokens[i]
	base := tok[strings.LastIndexAny(tok, `/\`)+1:]
	if base != "git" {
		return false
	}
	for _, p := range tokens[:i] {
		switch p {
		case "export", "env", "sudo", "command", "nice", "nohup", "time":
			continue
		}
		if !envAssignPattern.MatchString(p) {
			return false
		}
	}
	return true
}

// envAssignPattern matches a NAME=VALUE environment assignment token.
var envAssignPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// gitVerbAfterGlobals walks git global options from the token list after the
// git token and returns the first non-global token (the verb candidate).
// A masked placeholder stops the walk: nothing reliable lives beyond a
// collapsed quoted span.
func gitVerbAfterGlobals(rest []string) (string, bool) {
	j := 0
	for j < len(rest) {
		t := rest[j]
		if t == maskedTokenPlaceholder {
			return "", false
		}
		span, ok := gitGlobalOptionSpan(t, rest[j+1:])
		if !ok {
			return t, true
		}
		j += span
	}
	return "", false
}

// gitGlobalOptionSpan reports how many tokens the git global option starting
// at t spans (including t itself); ok=false when t is not a global option.
func gitGlobalOptionSpan(t string, next []string) (int, bool) {
	switch t {
	case "-C", "-c":
		if len(next) > 0 {
			return 2, true
		}
		return 0, false
	case "--git-dir", "--work-tree", "--namespace", "--exec-path",
		"--super-prefix", "--shallow-file":
		if strings.Contains(t, "=") || len(next) == 0 {
			return 1, true
		}
		return 2, true
	case "--no-pager", "--paginate", "--no-paginate", "--bare",
		"--no-replace-objects", "--literal-pathspecs", "--glob-pathspecs",
		"--noglob-pathspecs", "--icase-pathspecs", "--no-optional-locks",
		"--end-of-options":
		return 1, true
	}
	// Attached-value forms: -C<path>, -c<key>=<value>, --git-dir=<path>, …
	for _, o := range []string{"-C", "-c", "--git-dir", "--work-tree",
		"--namespace", "--exec-path", "--super-prefix", "--shallow-file"} {
		if strings.HasPrefix(t, o) && len(t) > len(o) {
			return 1, true
		}
	}
	return 0, false
}

// --- target directory and override extraction (spec.md §4 D3/D8) ---

// identityFacts carries what the guard extracted from the triggered git
// segment's ORIGINAL text and its directly chained `cd … &&` / `export … &&
// &&` prefix. Extraction runs on original text because the quote-collapsed
// form no longer carries `--author="t <t@t.t>"` values (spec.md §4 D2).
type identityFacts struct {
	authorEmail    string // --author=<value> — decisive for the author role
	gitAuthorEmail string // GIT_AUTHOR_EMAIL= inline or in a chained export — decisive for the author role
	committerEmail string // GIT_COMMITTER_EMAIL= inline or in a chained export — decisive for the committer role
	configEmail    string // -c user.email=<value> — replaces a role lacking a higher-precedence value
	email          string // EMAIL= — last resort; never decisive on its own (see D3)
	targetDir      string // -C <path> — wins over the chained cd
	chainDir       string // the nearest directly chained `cd <path>`
}

// extractIdentityFacts reads the triggered segment's original text plus the
// directly chained leading cd/export prefix (segments connected by `&&` only;
// a `;` or `|` breaks the chain). A preceding segment that is neither cd nor
// export ends the walk.
func extractIdentityFacts(segs []shellSegment, trigIdx int) identityFacts {
	var facts identityFacts
	applySegmentFacts(segs[trigIdx].text, &facts)
	// segs[k+1].conn is the connector BETWEEN segment k and the segment after
	// it — the chain to the triggered segment holds only while every such
	// connector is `&&`.
	for k := trigIdx - 1; k >= 0 && segs[k+1].conn == "&&"; k-- {
		tokens := shellFields(segs[k].text)
		if len(tokens) == 0 {
			break
		}
		switch tokens[0] {
		case "cd":
			if facts.chainDir == "" {
				for _, p := range tokens[1:] {
					if !strings.HasPrefix(p, "-") {
						facts.chainDir = unquoteOperand(p)
						break
					}
				}
			}
		case "export":
			applyAssignments(tokens[1:], &facts)
		default:
			return facts
		}
	}
	return facts
}

// applySegmentFacts extracts env-assignment prefixes, -C, -c user.email and
// --author from one segment's original text. Tokenization is quote-aware
// (shellFields) so `--author="t <t@t.t>"` and `git -C "my path"` survive.
func applySegmentFacts(segText string, facts *identityFacts) {
	tokens := shellFields(segText)
	for i := range tokens {
		if !isGitCommandToken(tokens, i) {
			continue
		}
		applyAssignments(tokens[:i], facts)
		for j := i + 1; j < len(tokens); j++ {
			t := tokens[j]
			switch {
			case t == "-C" && j+1 < len(tokens):
				facts.targetDir = unquoteOperand(tokens[j+1])
				j++
			case strings.HasPrefix(t, "-C") && len(t) > 2:
				facts.targetDir = unquoteOperand(t[2:])
			case t == "-c" && j+1 < len(tokens):
				applyConfigOption(unquoteOperand(tokens[j+1]), facts)
				j++
			case strings.HasPrefix(t, "-c") && len(t) > 2:
				applyConfigOption(unquoteOperand(t[2:]), facts)
			case t == "--author" && j+1 < len(tokens):
				facts.authorEmail = emailFromAuthorValue(tokens[j+1])
				j++
			case strings.HasPrefix(t, "--author="):
				facts.authorEmail = emailFromAuthorValue(t[len("--author="):])
			}
		}
		return
	}
}

// applyAssignments records GIT_AUTHOR_EMAIL / GIT_COMMITTER_EMAIL / EMAIL
// values from a token list of env assignments.
func applyAssignments(tokens []string, facts *identityFacts) {
	for _, tok := range tokens {
		name, value, ok := strings.Cut(unquoteOperand(tok), "=")
		if !ok {
			continue
		}
		switch name {
		case "GIT_AUTHOR_EMAIL":
			facts.gitAuthorEmail = strings.TrimSpace(value)
		case "GIT_COMMITTER_EMAIL":
			facts.committerEmail = strings.TrimSpace(value)
		case "EMAIL":
			facts.email = strings.TrimSpace(value)
		}
	}
}

// applyConfigOption records a `git -c <key>=<value>` option value; only the
// user.email key bears on identity.
func applyConfigOption(kv string, facts *identityFacts) {
	key, value, ok := strings.Cut(kv, "=")
	if !ok {
		return
	}
	if key == "user.email" {
		facts.configEmail = strings.TrimSpace(value)
	}
}

// emailFromAuthorValue extracts the email an --author value would give:
// the bracketed email of `Name <email>`, or the bare value when it carries
// an '@'. A value without any '@' cannot match an email list and yields "".
func emailFromAuthorValue(value string) string {
	value = unquoteOperand(value)
	if _, after, ok := strings.Cut(value, "<"); ok {
		if email, _, ok2 := strings.Cut(after, ">"); ok2 {
			return strings.TrimSpace(email)
		}
	}
	if strings.Contains(value, "@") {
		return strings.TrimSpace(value)
	}
	return ""
}

// shellFields splits on whitespace outside single/double quotes, keeping
// each quoted group attached to its token (quotes retained). A backslash
// escapes the next byte inside double quotes and outside any quote.
func shellFields(s string) []string {
	var out []string
	var cur strings.Builder
	inSingle, inDouble, started := false, false, false
	flush := func() {
		if started {
			out = append(out, cur.String())
			cur.Reset()
			started = false
		}
	}
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case inSingle:
			if c == '\'' {
				inSingle = false
			}
			cur.WriteByte(c)
			i++
		case inDouble:
			if c == '\\' && i+1 < len(s) {
				cur.WriteByte(c)
				cur.WriteByte(s[i+1])
				i += 2
				continue
			}
			if c == '"' {
				inDouble = false
			}
			cur.WriteByte(c)
			i++
		case c == '\'':
			inSingle = true
			started = true
			cur.WriteByte(c)
			i++
		case c == '"':
			inDouble = true
			started = true
			cur.WriteByte(c)
			i++
		case c == '\\':
			cur.WriteByte(c)
			if i+1 < len(s) {
				cur.WriteByte(s[i+1])
				i += 2
			} else {
				i++
			}
		case c == ' ' || c == '\t':
			flush()
			i++
		default:
			started = true
			cur.WriteByte(c)
			i++
		}
	}
	flush()
	return out
}

// --- repository scope (spec.md §4 D8) ---

// sameCommonDir compares two git common dirs as absolute cleaned paths with
// symlinks resolved (macOS /var ↔ /private/var alias). A path that cannot be
// normalized compares as unequal — the fail-open direction.
func sameCommonDir(a, b string) bool {
	na, nb := normalizeRepoPath(a), normalizeRepoPath(b)
	if na == "" || nb == "" {
		return false
	}
	return strings.EqualFold(na, nb)
}

// normalizeRepoPath resolves symlinks, makes the path absolute against the
// process cwd, and cleans it. "" on failure.
func normalizeRepoPath(p string) string {
	if p == "" {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		p = resolved
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return ""
	}
	return filepath.Clean(abs)
}

// --- the guard ---

// checkCommitIdentity returns DecisionDeny plus a TEST_IDENTITY_VIOLATION:
// reason when the command's resolved identity matches the effective deny
// list; otherwise ("", "") (allow fall-through). denyEmails is the effective
// list (built-in ∪ config), resolved by the caller's wiring.
//
// Decision order: unclassifiable PowerShell indirection (allow + one audit
// line) → trigger classification (no trigger: allow, zero probes) →
// repository scope (different common dir: allow, no identity probe) →
// decisive command-level overrides (a deny-listed decisive value denies
// before any probe, REQ-CIG-007) → identity probes → exact-match comparison.
func checkCommitIdentity(input *HookInput, projectDir string, denyEmails []string) (decision, reason string) {
	if input == nil || len(input.ToolInput) == 0 {
		return "", ""
	}
	command := shellCommandText(input.ToolInput)
	if command == "" {
		return "", ""
	}
	cwd := resolveProjectRootFromInputOrEnv(input, "commit_identity_guard")

	// REQ-CIG-010: a PowerShell indirection construct cannot be classified as
	// a trigger — allow without probes and record exactly one line.
	if isPowerShellTool(input.ToolName) {
		if construct := powerShellIndirection(command); construct != "" && gitWordPattern.MatchString(command) {
			appendUnclassifiedAudit(projectDir, commitIdentityAuditRelPath, input, construct, command, cwd)
			return "", ""
		}
	}

	segs := splitShellSegments(command)
	trigIdx := -1
	for i, s := range segs {
		if findCommitTrigger(s.text) != "" {
			trigIdx = i
			break
		}
	}
	if trigIdx < 0 {
		return "", "" // REQ-CIG-001: no commit-creating verb, no probes
	}

	facts := extractIdentityFacts(segs, trigIdx)
	targetDir := cwd
	switch {
	case facts.targetDir != "":
		targetDir = resolveAgainst(cwd, facts.targetDir)
	case facts.chainDir != "":
		targetDir = resolveAgainst(cwd, facts.chainDir)
	}
	if targetDir == "" {
		appendCommitIdentityFailOpen(input, projectDir, command, cwd,
			"command target directory unresolvable")
		return "", ""
	}

	// Repository scope (D8): both sides resolve their git common dir; any
	// resolution failure allows with one advisory line (REQ-CIG-007 e/f).
	targetCommon, errT := commitIdentityScopeProbe(targetDir)
	if errT != nil {
		appendCommitIdentityFailOpen(input, projectDir, command, targetDir,
			fmt.Sprintf("target repository scope unresolved: %v", errT))
		return "", ""
	}
	if projectDir == "" {
		appendCommitIdentityFailOpen(input, projectDir, command, targetDir,
			"project directory unresolvable; repository scope cannot be compared")
		return "", ""
	}
	projectCommon, errP := commitIdentityScopeProbe(projectDir)
	if errP != nil {
		appendCommitIdentityFailOpen(input, projectDir, command, targetDir,
			fmt.Sprintf("project repository scope unresolved: %v", errP))
		return "", ""
	}
	if !sameCommonDir(targetCommon, projectCommon) {
		return "", "" // REQ-CIG-012: another repository — allow, no identity probe
	}

	// Resolve effective identity per git precedence (D3).
	authorEff := firstNonEmptyStr(facts.authorEmail, facts.gitAuthorEmail, facts.configEmail)
	committerEff := firstNonEmptyStr(facts.committerEmail, facts.configEmail)
	// EMAIL is the last resort: it only applies when the role resolves no
	// user.email configuration or higher override. A successful probe already
	// reflects that resolution, so a probe-succeeded email shadows EMAIL; a
	// failed probe leaves the question undecidable and the command allows
	// (REQ-CIG-007) — EMAIL alone is never positive evidence.

	// A decisive deny-listed value is positive evidence: deny before probing.
	if role, email := matchDecisive(authorEff, committerEff, denyEmails); role != "" {
		return DecisionDeny, testIdentityDenyReason(role, email)
	}

	if authorEff == "" || committerEff == "" {
		authorIdent, committerIdent, err := commitIdentityVarProbe(targetDir, commitIdentityProbeEnv())
		if err != nil {
			appendCommitIdentityFailOpen(input, projectDir, command, targetDir,
				fmt.Sprintf("identity probe failed: %v", err))
			return "", ""
		}
		if authorEff == "" {
			email, perr := emailFromIdent(authorIdent)
			if perr != nil {
				appendCommitIdentityFailOpen(input, projectDir, command, targetDir, perr.Error())
				return "", ""
			}
			authorEff = email
		}
		if committerEff == "" {
			email, perr := emailFromIdent(committerIdent)
			if perr != nil {
				appendCommitIdentityFailOpen(input, projectDir, command, targetDir, perr.Error())
				return "", ""
			}
			committerEff = email
		}
	}

	if isDenyListed(authorEff, denyEmails) {
		return DecisionDeny, testIdentityDenyReason("author", authorEff)
	}
	if isDenyListed(committerEff, denyEmails) {
		return DecisionDeny, testIdentityDenyReason("committer", committerEff)
	}
	return "", "" // REQ-CIG-005: both emails outside the list — allow
}

// matchDecisive judges only values that are DECISIVE (present): it returns
// the first deny-listed role among author and committer, preferring author.
func matchDecisive(authorEff, committerEff string, list []string) (role, email string) {
	if authorEff != "" && isDenyListed(authorEff, list) {
		return "author", authorEff
	}
	if committerEff != "" && isDenyListed(committerEff, list) {
		return "committer", committerEff
	}
	return "", ""
}

// firstNonEmpty returns the first non-empty argument.
func firstNonEmptyStr(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// resolveAgainst makes p absolute against base (a relative -C or cd path is
// relative to the command's working directory).
func resolveAgainst(base, p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	if base == "" {
		return ""
	}
	return filepath.Clean(filepath.Join(base, p))
}

// testIdentityDenyReason builds the deny reason: the sentinel, the matched
// email, the identity role, and a remedy directing the caller at the role's
// actual configuration or command override (spec.md §4 D3).
func testIdentityDenyReason(role, email string) string {
	return fmt.Sprintf("%s %s identity %s matches a known test-fixture email; the commit was refused. "+
		"Fix the identity git would use (the %s override such as GIT_%s_EMAIL/--author/-c user.email, or the repository's user.email), then retry.",
		testIdentityViolationPrefix, role, email, role, strings.ToUpper(role))
}

// appendCommitIdentityFailOpen writes one advisory line naming the cause to
// <projectDir>/.moai/logs/commit-identity-guard-audit.log and a stderr note.
// Logging failures are debug-level only — fail-open must never block the
// allow decision.
func appendCommitIdentityFailOpen(input *HookInput, projectDir, command, dir, cause string) {
	sessionID := ""
	if input != nil {
		sessionID = input.SessionID
	}
	fmt.Fprintf(os.Stderr, "commit_identity_guard: fail-open for command %q at dir %q (audit-log dir %q): %s\n",
		command, dir, projectDir, cause)
	if projectDir == "" {
		return
	}
	entry := fmt.Sprintf("[%s] event=commit-identity-fail-open session=%s command=%q dir=%q cause=%q\n",
		time.Now().UTC().Format(time.RFC3339), sessionID, command, dir, cause)
	logPath := filepath.Join(projectDir, commitIdentityAuditRelPath)
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "commit_identity_guard: could not create audit log dir %q: %v\n", logPath, err)
		return
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "commit_identity_guard: could not open audit log %q: %v\n", logPath, err)
		return
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(entry); err != nil {
		fmt.Fprintf(os.Stderr, "commit_identity_guard: could not write audit log entry: %v\n", err)
	}
}
