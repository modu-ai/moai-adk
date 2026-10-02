// aside_skill_policy_test.go: policy guard for the optional Aside browser
// reference skill (moai-ref-aside-browser).
//
// The skill carries the safety rules for the optional Aside browser: orchestrator-only
// operation, no unrestricted permission mode, Guard by omission, a read-only
// discipline for the REPL surface, operator confirmation for state changes, and
// advice-only installation. Each rule is pinned as a literal anchor so that two
// authors cannot disagree about what the text must say.
//
// The checker is a pure function over the skill text (checkAsideSkill). The real
// skill is judged by it, and then the same text is mutated (one anchor removed, or
// one bad line added) and every mutant must produce its named violation: a control
// that passes the checker would mean the checker cannot see that defect.
//
// The e2e wiring is judged the same way: checkAsideE2E reads the e2e workflow,
// checkAsideTester reads the e2e-tester definition, and the e2e_negative_controls
// group mutates the real text (a literal removed, a carve-out stripped from one
// site, a forbidden line added, the CI clause moved) and requires each mutant to
// produce its named violation.
//
// Verified in isolation via:
//
//	go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1
package template_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// asideSkillRelPath is the template location of the skill, relative to the
// templates root.
const asideSkillRelPath = ".claude/skills/moai-ref-aside-browser/SKILL.md"

// asideListingCap is the combined description + when_to_use character cap of one
// skill-listing entry (maxSkillDescriptionChars).
const asideListingCap = 1536

// Rule names. They double as the subtest names of TestAsideSkillPolicyAnchors.
const (
	asideRuleListingCap        = "description_within_listing_cap"
	asideRuleFullAccess        = "full_access_prohibited"
	asideRuleGuardByOmission   = "guard_by_omission"
	asideRuleReplReadOnly      = "repl_read_only_limit"
	asideRuleWriteConfirmation = "write_confirmation_via_orchestrator"
	asideRuleOrchestratorOnly  = "orchestrator_only_operation"
	asideRuleSubagentSkill     = "subagent_never_invokes_skill"
	asideRuleNoAutoInstall     = "no_auto_install_advise_only"
	asideRuleNoCredentials     = "no_credentials_in_outputs"
)

// asideSkillRules lists every skill-side rule in a fixed order.
var asideSkillRules = []string{
	asideRuleListingCap,
	asideRuleFullAccess,
	asideRuleGuardByOmission,
	asideRuleReplReadOnly,
	asideRuleWriteConfirmation,
	asideRuleOrchestratorOnly,
	asideRuleSubagentSkill,
	asideRuleNoAutoInstall,
	asideRuleNoCredentials,
}

// Literal anchors. Backticks are part of several literals, so these are
// interpreted strings.
const (
	asideLitNeverPassFullAccess  = "Never pass `--permission full-access`"
	asideLitNeverRequestFullAcc  = "never request full-access"
	asideLitOmitPermission       = "omit `--permission`"
	asideLitGuard                = "Guard"
	asideLitAsideRepl            = "aside repl"
	asideLitNoPermissionFlag     = "no permission flag"
	asideLitReplOperations       = "navigation, reading, and screenshot"
	asideLitExplicitConfirmation = "explicit operator confirmation"
	asideLitQuestionChannel      = "question channel"
	asideLitOperatedAlone        = "operated by the orchestrator alone"
	asideLitNeverInvokes         = "never invokes Aside"
	asideLitAdviseInstall        = "advise the operator to run `aside skills install`"
	asideLitInstallCommand       = "aside skills install"
	asideLitOperator             = "operator"
	asideLitNoCredentials        = "never place credentials, cookies, tokens, or session data"
)

// asideRequiredLiterals maps each rule to the literals that must appear in the
// skill text. full_access_prohibited and description_within_listing_cap carry
// their own logic below (a required literal plus a forbidden-line scan, and a
// character measurement respectively).
var asideRequiredLiterals = map[string][]string{
	asideRuleFullAccess:        {asideLitNeverPassFullAccess},
	asideRuleGuardByOmission:   {asideLitOmitPermission, asideLitGuard},
	asideRuleReplReadOnly:      {asideLitAsideRepl, asideLitNoPermissionFlag, asideLitReplOperations},
	asideRuleWriteConfirmation: {asideLitExplicitConfirmation, asideLitQuestionChannel},
	asideRuleOrchestratorOnly:  {asideLitOperatedAlone},
	asideRuleSubagentSkill:     {asideLitNeverInvokes},
	asideRuleNoAutoInstall:     {asideLitAdviseInstall},
	asideRuleNoCredentials:     {asideLitNoCredentials},
}

// asideInstallVerbRe matches the install tooling that must never appear on a
// line that mentions Aside. Word boundaries keep "pip" from matching inside
// longer words.
var asideInstallVerbRe = regexp.MustCompile(`(?i)\b(npm|npx|curl|brew|cargo|pip|go install)\b`)

// asideMentionRe matches a mention of the product, case-insensitively.
var asideMentionRe = regexp.MustCompile(`(?i)aside`)

// asideViolation is one failed anchor, tagged with the rule (subtest) it belongs
// to.
type asideViolation struct {
	Rule string
	Msg  string
}

// asideFrontmatter holds the two fields the listing cap is measured over.
type asideFrontmatter struct {
	Description string `yaml:"description"`
	WhenToUse   string `yaml:"when_to_use"`
}

// asideSplitFrontmatter returns the YAML block between the leading "---" fence
// and the next "---" fence.
func asideSplitFrontmatter(text string) (string, bool) {
	if !strings.HasPrefix(text, "---\n") {
		return "", false
	}
	rest := text[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", false
	}
	return rest[:end], true
}

// checkAsideSkill judges the skill text against every skill-side anchor and
// returns one violation per failed anchor. A nil result means the text complies.
func checkAsideSkill(text string) []asideViolation {
	var out []asideViolation
	add := func(rule, format string, args ...any) {
		out = append(out, asideViolation{Rule: rule, Msg: fmt.Sprintf(format, args...)})
	}

	// description_within_listing_cap: description plus when_to_use, in characters.
	if block, ok := asideSplitFrontmatter(text); !ok {
		add(asideRuleListingCap, "no YAML frontmatter block found")
	} else {
		var fm asideFrontmatter
		if err := yaml.Unmarshal([]byte(block), &fm); err != nil {
			add(asideRuleListingCap, "frontmatter is not valid YAML: %v", err)
		} else {
			n := utf8.RuneCountInString(strings.TrimSpace(fm.Description)) +
				utf8.RuneCountInString(strings.TrimSpace(fm.WhenToUse))
			if n > asideListingCap {
				add(asideRuleListingCap, "description plus when_to_use is %d characters, cap is %d", n, asideListingCap)
			}
		}
	}

	// Required literals, rule by rule.
	for _, rule := range asideSkillRules {
		for _, lit := range asideRequiredLiterals[rule] {
			if !strings.Contains(text, lit) {
				add(rule, "required literal %q is missing", lit)
			}
		}
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// full_access_prohibited: a closed two-literal allow-list. Any other line
		// that contains full-access is a violation.
		if strings.Contains(line, "full-access") &&
			!strings.Contains(line, asideLitNeverPassFullAccess) &&
			!strings.Contains(line, asideLitNeverRequestFullAcc) {
			add(asideRuleFullAccess, "line %d mentions full-access outside the allow-list: %q", i+1, line)
		}

		// no_auto_install_advise_only: every line naming the install command must
		// also carry the operator framing, and no line may pair Aside with install
		// tooling.
		if strings.Contains(line, asideLitInstallCommand) && !strings.Contains(line, asideLitOperator) {
			add(asideRuleNoAutoInstall, "line %d names %q without %q: %q", i+1, asideLitInstallCommand, asideLitOperator, line)
		}
		if asideMentionRe.MatchString(line) && asideInstallVerbRe.MatchString(line) {
			add(asideRuleNoAutoInstall, "line %d pairs Aside with install tooling: %q", i+1, line)
		}
	}
	return out
}

// asideViolationsFor filters violations down to one rule.
func asideViolationsFor(vs []asideViolation, rule string) []asideViolation {
	var out []asideViolation
	for _, v := range vs {
		if v.Rule == rule {
			out = append(out, v)
		}
	}
	return out
}

// readAsideTemplate reads one template file by its path relative to the templates
// root. A read failure is returned, not fatal, so the caller can fail every named
// subtest.
func readAsideTemplate(t *testing.T, rel string) (string, error) {
	t.Helper()
	root := findNeutralityRoot(t) // .../internal/template/templates
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// readAsideSkill reads the template skill file.
func readAsideSkill(t *testing.T) (string, error) {
	t.Helper()
	return readAsideTemplate(t, asideSkillRelPath)
}

// TestAsideSkillPolicyAnchors pins the Aside skill's safety rules.
//
// Every subtest runs the pure checker on the real skill and requires no
// violation for its rule. The negative_controls group then mutates the same text
// and requires each mutant to produce its named violation, so the checker is
// proven able to see every defect it claims to guard.
func TestAsideSkillPolicyAnchors(t *testing.T) {
	text, readErr := readAsideSkill(t)
	var violations []asideViolation
	if readErr == nil {
		violations = checkAsideSkill(text)
	}

	for _, rule := range asideSkillRules {
		t.Run(rule, func(t *testing.T) {
			if readErr != nil {
				t.Fatalf("skill file unreadable: %v", readErr)
			}
			for _, v := range asideViolationsFor(violations, rule) {
				t.Errorf("%s", v.Msg)
			}
		})
	}

	t.Run("negative_controls", func(t *testing.T) {
		if readErr != nil {
			t.Fatalf("skill file unreadable: %v", readErr)
		}

		// requireViolation asserts the checker reports the named rule on mutated
		// text.
		requireViolation := func(t *testing.T, mutated, rule string) {
			t.Helper()
			if len(asideViolationsFor(checkAsideSkill(mutated), rule)) == 0 {
				t.Errorf("the checker did not report %q on the mutated text; the control passed the checker, so the checker cannot see this defect", rule)
			}
		}

		// Each required literal, removed in turn, must trigger its rule.
		for _, rule := range asideSkillRules {
			for i, lit := range asideRequiredLiterals[rule] {
				name := fmt.Sprintf("remove/%s/%d", rule, i)
				t.Run(name, func(t *testing.T) {
					if !strings.Contains(text, lit) {
						t.Fatalf("literal %q is not in the real skill; the control would be vacuous", lit)
					}
					requireViolation(t, strings.ReplaceAll(text, lit, ""), rule)
				})
			}
		}

		// Bad lines, added in turn, must trigger their rule.
		bad := []struct {
			name, line, rule string
		}{
			{"add/unqualified_full_access", "aside --permission full-access", asideRuleFullAccess},
			{"add/never_forget_full_access", "never forget to run aside --permission full-access", asideRuleFullAccess},
			{"add/npm_install", "npm i -g aside", asideRuleNoAutoInstall},
			{"add/install_command_without_operator", "run `aside skills install` now", asideRuleNoAutoInstall},
		}
		for _, b := range bad {
			t.Run(b.name, func(t *testing.T) {
				requireViolation(t, text+"\n"+b.line+"\n", b.rule)
			})
		}

		// An oversize description must trip the listing cap.
		t.Run("add/oversize_description", func(t *testing.T) {
			mutated := strings.Replace(text, "description: >", "description: >\n  "+strings.Repeat("x", asideListingCap+64), 1)
			if mutated == text {
				t.Fatalf("no 'description: >' fold marker in the real skill; the control would be vacuous")
			}
			requireViolation(t, mutated, asideRuleListingCap)
		})

		// Text with no frontmatter at all cannot satisfy the cap rule either.
		t.Run("remove/frontmatter", func(t *testing.T) {
			requireViolation(t, strings.TrimPrefix(text, "---\n"), asideRuleListingCap)
		})
	})

	runAsideE2EAnchors(t)
}

// Template locations of the two e2e-side files, relative to the templates root.
const (
	asideE2ERelPath    = ".claude/skills/moai/workflows/e2e.md"
	asideTesterRelPath = ".claude/agents/moai/e2e-tester.md"
)

// asideE2EMinSites is the number of lines the role-based enumeration finds in the
// workflow: the 15 baseline sites, the 6 autofix-delegation sites that hand a
// repair to a subagent and could be read as handing it the re-run, and the one
// precedence sentence. An enumeration that finds fewer is treated as emptied.
const asideE2EMinSites = 22

// e2e-side rule names. They double as subtest names of TestAsideSkillPolicyAnchors.
const (
	asideRuleSubagentTester = "subagent_never_invokes_tester"
	asideRuleE2EExplicit    = "e2e_explicit_only"
	asideRuleE2ECI          = "e2e_ci_excluded"
	asideRuleE2EOrchestrate = "e2e_orchestrator_executes_aside"
	asideRuleE2EOwner       = "e2e_execution_owner_carveout"
	asideRuleE2EBounded     = "e2e_aside_output_bounded"
	asideRuleE2ESilent      = "e2e_silent_fallback_no_aside_message"
	asideRuleE2EHeader      = "e2e_missing_toolchain_carveout"
	asideRuleE2EBypass      = "e2e_tool_bypass_line_carveout"
	asideRuleE2ESites       = "e2e_every_site_carved_out"
	asideRuleE2EAutofix     = "e2e_autofix_orchestrator_reverifies"
	asideRuleE2ENoInstall   = "e2e_no_install_command"
	asideRuleE2EScreenshot  = "e2e_repl_screenshot_evidence"
)

// asideE2ERules lists every e2e-workflow rule in a fixed order.
var asideE2ERules = []string{
	asideRuleE2EExplicit,
	asideRuleE2ECI,
	asideRuleE2EOrchestrate,
	asideRuleE2EOwner,
	asideRuleE2EBounded,
	asideRuleE2ESilent,
	asideRuleE2EHeader,
	asideRuleE2EBypass,
	asideRuleE2ESites,
	asideRuleE2EAutofix,
	asideRuleE2ENoInstall,
	asideRuleE2EScreenshot,
}

// e2e-side literal anchors.
const (
	asideLitOnlyExplicit      = "only when `--tool aside` is passed explicitly"
	asideLitNeverAutoDetected = "never auto-detected"
	asideLitNeverOffered      = "never offered"
	asideLitNeverRecommended  = "never recommended"
	asideLitOrchestratorRuns  = "the ORCHESTRATOR runs every Aside step"
	asideLitSkillLoad         = `Skill("moai-ref-aside-browser")`
	asideLitSilently          = "silently"
	asideLitNoAsideMessage    = "no Aside-specific message"
	asideLitExceptAside       = "except Aside"
	asideLitOrchestratorCaps  = "ORCHESTRATOR"
	asideLitExecutionOwner    = "Execution owner"
	asideLitBoundedOutput     = "Bounded output"
	asideLitRunsDir           = "e2e/.runs/"
	asideLitBypassSentence    = "missing-toolchain sequence if absent"
	asideLitHeaderPrefix      = "Missing-toolchain sequence"
	asideLitScreenshot        = "screenshot captured through `aside repl`"
	asideLitE2EDir            = "e2e/"
	asideLitCI                = "CI=true"
	asideLitUnavailable       = "unavailable"

	// asideLitPrecedence is the one sentence on which `silently` counts. A site
	// line carries `except Aside` instead; the sentence is exempt only as a line of
	// its own (optionally a list item), never appended to another line.
	asideLitPrecedence = "An absent or excluded Aside is not a missing toolchain: continue silently on the platform default, with no Aside-specific message, prompt, install attempt, or failure; this rule takes precedence over the Surface and Install steps and over the `--tool` bypass."
)

// asideE2ERequiredLiterals maps each e2e rule to the literals that must appear
// somewhere in the workflow. The carve-out, CI, site, install, and screenshot
// rules carry their own line-level logic in checkAsideE2E.
var asideE2ERequiredLiterals = map[string][]string{
	asideRuleE2EExplicit:    {asideLitOnlyExplicit, asideLitNeverAutoDetected, asideLitNeverOffered, asideLitNeverRecommended},
	asideRuleE2EOrchestrate: {asideLitOrchestratorRuns, asideLitSkillLoad},
	asideRuleE2ESilent:      {asideLitSilently, asideLitNoAsideMessage},
	asideRuleE2EScreenshot:  {asideLitScreenshot},
}

// asideSiteRe is the role-based enumeration of the workflow lines that delegate
// script creation, execution, or recording to the e2e-tester, run the
// missing-toolchain sequence, state who owns execution or output, or hand an e2e
// failure to the autofix subagent (the Input, Cycle, and Validate bullets of the
// delegation contract, and every line that names the manager-develop autofix
// delegation). Every such line needs an Aside carve-out.
var asideSiteRe = regexp.MustCompile(`[Dd]elegate .*(script creation|test execution|execution|recording)|Phase [234]: e2e-tester|[Mm]issing[- ]toolchain|Execution owner|Bounded output|toolchain probe/install|probes the DEFAULT toolchain|^- \*\*(Input|Cycle|Validate)\*\*:|manager-develop autofix|manager-develop \(autofix\)`)

// asideAutofixValidateLine reports whether the line is the Validate bullet of the
// autofix delegation contract.
func asideAutofixValidateLine(l string) bool {
	return strings.HasPrefix(strings.TrimSpace(l), "- **Validate**:")
}

// asideCarveOutRe matches the appended parenthetical of a carved-out site. Removing
// it restores the site's original text.
var asideCarveOutRe = regexp.MustCompile(` \(except Aside: .*\)$`)

// asideNoteWords are phrases that, on a line that mentions Aside, would instruct an
// Aside-specific note, warning, or message to the operator. They are compared
// lowercase.
var asideNoteWords = []string{
	"warn",
	"fallback note",
	"notify",
	"report that aside",
	"tell the operator",
	"inform the operator",
	"let the operator know",
	"tell the user",
	"inform the user",
	"let the user know",
}

// asideStripCarveOut removes the appended Aside carve-out from one line.
func asideStripCarveOut(line string) string {
	return asideCarveOutRe.ReplaceAllString(line, "")
}

// asideIsPrecedenceLine reports whether the line is exactly the precedence
// sentence, optionally as a list item.
func asideIsPrecedenceLine(line string) bool {
	s := strings.TrimPrefix(strings.TrimSpace(line), "- ")
	return s == asideLitPrecedence
}

// asideIsCILine reports whether the line states the CI exclusion of Aside.
func asideIsCILine(line string) bool {
	return strings.Contains(line, asideLitCI) &&
		strings.Contains(line, "Aside") &&
		strings.Contains(line, asideLitUnavailable)
}

// asideBlockAround returns the first and last index of the run of non-blank lines
// that contains line i.
func asideBlockAround(lines []string, i int) (int, int) {
	s, e := i, i
	for s > 0 && strings.TrimSpace(lines[s-1]) != "" {
		s--
	}
	for e+1 < len(lines) && strings.TrimSpace(lines[e+1]) != "" {
		e++
	}
	return s, e
}

// asideFindLine returns the index of the first line satisfying pred, or -1.
func asideFindLine(lines []string, pred func(string) bool) int {
	for i, l := range lines {
		if pred(l) {
			return i
		}
	}
	return -1
}

// asideSiteLineIndexes returns the indexes of every line the site enumeration
// matches.
func asideSiteLineIndexes(lines []string) []int {
	var out []int
	for i, l := range lines {
		if asideSiteRe.MatchString(l) {
			out = append(out, i)
		}
	}
	return out
}

// asideReplaceLine returns the text with line i replaced.
func asideReplaceLine(lines []string, i int, repl string) string {
	cp := append([]string(nil), lines...)
	cp[i] = repl
	return strings.Join(cp, "\n")
}

// asideDropLines returns the text without the lines satisfying pred.
func asideDropLines(lines []string, pred func(string) bool) string {
	var kept []string
	for _, l := range lines {
		if !pred(l) {
			kept = append(kept, l)
		}
	}
	return strings.Join(kept, "\n")
}

// checkAsideTester judges the e2e-tester definition. A nil result means the text
// carries the subagent prohibition.
func checkAsideTester(text string) []asideViolation {
	if strings.Contains(text, asideLitNeverInvokes) {
		return nil
	}
	return []asideViolation{{Rule: asideRuleSubagentTester, Msg: fmt.Sprintf("required literal %q is missing", asideLitNeverInvokes)}}
}

// checkAsideE2E judges the e2e workflow text against every e2e-side anchor and
// returns one violation per failed anchor. A nil result means the text complies.
func checkAsideE2E(text string) []asideViolation {
	var out []asideViolation
	add := func(rule, format string, args ...any) {
		out = append(out, asideViolation{Rule: rule, Msg: fmt.Sprintf(format, args...)})
	}

	// Required literals, rule by rule.
	for _, rule := range asideE2ERules {
		for _, lit := range asideE2ERequiredLiterals[rule] {
			if !strings.Contains(text, lit) {
				add(rule, "required literal %q is missing", lit)
			}
		}
	}

	lines := strings.Split(text, "\n")

	// Single-site carve-out rules: every line the matcher selects must carry the
	// carve-out, and at least one such line must exist. tail, when set, must sit
	// inside the carve-out itself, not in the original text before it.
	requireCarveOut := func(rule, what string, match func(string) bool, tail string) {
		found := false
		for i, line := range lines {
			if !match(line) {
				continue
			}
			found = true
			idx := strings.Index(line, asideLitExceptAside)
			if idx < 0 {
				add(rule, "line %d (%s) lacks %q: %q", i+1, what, asideLitExceptAside, line)
				continue
			}
			if tail != "" && !strings.Contains(line[idx:], tail) {
				add(rule, "line %d (%s): the carve-out lacks %q: %q", i+1, what, tail, line)
			}
		}
		if !found {
			add(rule, "no line carries %s", what)
		}
	}
	requireCarveOut(asideRuleE2EOwner, "the execution-owner statement",
		func(l string) bool { return strings.Contains(l, asideLitExecutionOwner) }, "")
	requireCarveOut(asideRuleE2EBounded, "the bounded-output rule",
		func(l string) bool { return strings.Contains(l, asideLitBoundedOutput) }, asideLitRunsDir)
	requireCarveOut(asideRuleE2EHeader, "the missing-toolchain sequence header",
		func(l string) bool { return strings.HasPrefix(strings.TrimSpace(l), asideLitHeaderPrefix) }, "")
	requireCarveOut(asideRuleE2EBypass, "the --tool bypass sentence",
		func(l string) bool { return strings.Contains(l, asideLitBypassSentence) }, "")

	// e2e_autofix_orchestrator_reverifies: the Validate bullet of the autofix
	// delegation contract tells the subagent to re-run the e2e spec, so its carve-out
	// must hand the re-verification to the orchestrator and keep Aside away from the
	// subagent.
	requireCarveOut(asideRuleE2EAutofix, "the autofix Validate bullet", asideAutofixValidateLine, asideLitOrchestratorCaps)
	requireCarveOut(asideRuleE2EAutofix, "the autofix Validate bullet", asideAutofixValidateLine, asideLitNeverInvokes)

	// e2e_ci_excluded: the CI exclusion must sit where the --tool path reads it,
	// that is inside the Aside paragraph or on the --tool bypass line, and not only
	// in the no-flag branch.
	var ciLines []int
	for i, l := range lines {
		if asideIsCILine(l) {
			ciLines = append(ciLines, i)
		}
	}
	if len(ciLines) == 0 {
		add(asideRuleE2ECI, "no line carries %q, Aside, and %q", asideLitCI, asideLitUnavailable)
	} else {
		inScope := func(i int) bool {
			if strings.Contains(lines[i], asideLitBypassSentence) {
				return true
			}
			for j, l := range lines {
				if !strings.Contains(l, asideLitSkillLoad) {
					continue
				}
				if s, e := asideBlockAround(lines, j); i >= s && i <= e {
					return true
				}
			}
			return false
		}
		ok := false
		for _, i := range ciLines {
			if inScope(i) {
				ok = true
			}
		}
		if !ok {
			add(asideRuleE2ECI, "the %q exclusion line is neither in the Aside paragraph nor on the --tool bypass line", asideLitCI)
		}
	}

	// e2e_silent_fallback_no_aside_message: the precedence sentence stands as its
	// own line, and no line that mentions Aside instructs a note or a message.
	if asideFindLine(lines, asideIsPrecedenceLine) < 0 {
		add(asideRuleE2ESilent, "the precedence sentence is not present as a line of its own")
	}
	for i, l := range lines {
		if !asideMentionRe.MatchString(l) {
			continue
		}
		low := strings.ToLower(l)
		for _, w := range asideNoteWords {
			if strings.Contains(low, w) {
				add(asideRuleE2ESilent, "line %d mentions Aside and %q: %q", i+1, w, l)
			}
		}
	}

	// e2e_every_site_carved_out: every enumerated line carries `except Aside`;
	// `silently` counts only on the exact precedence-sentence line.
	sites := asideSiteLineIndexes(lines)
	if len(sites) < asideE2EMinSites {
		add(asideRuleE2ESites, "the site enumeration found %d lines, at least %d are expected", len(sites), asideE2EMinSites)
	}
	for _, i := range sites {
		if asideIsPrecedenceLine(lines[i]) {
			continue
		}
		if !strings.Contains(lines[i], asideLitExceptAside) {
			add(asideRuleE2ESites, "line %d matches the site pattern without %q: %q", i+1, asideLitExceptAside, lines[i])
		}
	}

	// e2e_no_install_command: no line that mentions Aside names install tooling.
	for i, l := range lines {
		if asideMentionRe.MatchString(l) && asideInstallVerbRe.MatchString(l) {
			add(asideRuleE2ENoInstall, "line %d pairs Aside with install tooling: %q", i+1, l)
		}
	}

	// e2e_repl_screenshot_evidence: one line names the REPL screenshot and the e2e/
	// directory together.
	if asideFindLine(lines, func(l string) bool {
		return strings.Contains(l, asideLitScreenshot) && strings.Contains(l, asideLitE2EDir)
	}) < 0 {
		add(asideRuleE2EScreenshot, "no line carries both %q and %q", asideLitScreenshot, asideLitE2EDir)
	}
	return out
}

// runAsideE2EAnchors adds the e2e-side subtests of TestAsideSkillPolicyAnchors:
// the e2e-tester half of the subagent anchor, one subtest per e2e rule, and the
// e2e_negative_controls group.
func runAsideE2EAnchors(t *testing.T) {
	e2eText, e2eErr := readAsideTemplate(t, asideE2ERelPath)
	testerText, testerErr := readAsideTemplate(t, asideTesterRelPath)

	var e2eViolations, testerViolations []asideViolation
	if e2eErr == nil {
		e2eViolations = checkAsideE2E(e2eText)
	}
	if testerErr == nil {
		testerViolations = checkAsideTester(testerText)
	}

	t.Run(asideRuleSubagentTester, func(t *testing.T) {
		if testerErr != nil {
			t.Fatalf("e2e-tester definition unreadable: %v", testerErr)
		}
		for _, v := range asideViolationsFor(testerViolations, asideRuleSubagentTester) {
			t.Errorf("%s", v.Msg)
		}
	})
	for _, rule := range asideE2ERules {
		t.Run(rule, func(t *testing.T) {
			if e2eErr != nil {
				t.Fatalf("e2e workflow unreadable: %v", e2eErr)
			}
			for _, v := range asideViolationsFor(e2eViolations, rule) {
				t.Errorf("%s", v.Msg)
			}
		})
	}

	t.Run("e2e_negative_controls", func(t *testing.T) {
		if e2eErr != nil {
			t.Fatalf("e2e workflow unreadable: %v", e2eErr)
		}
		if testerErr != nil {
			t.Fatalf("e2e-tester definition unreadable: %v", testerErr)
		}
		runAsideE2EControls(t, e2eText, testerText)
	})
}

// runAsideE2EControls mutates the real e2e workflow and e2e-tester text; every
// mutant must make the matching checker report its named rule.
func runAsideE2EControls(t *testing.T, e2eText, testerText string) {
	t.Helper()
	lines := strings.Split(e2eText, "\n")

	requireE2E := func(t *testing.T, mutated, rule string) {
		t.Helper()
		if len(asideViolationsFor(checkAsideE2E(mutated), rule)) == 0 {
			t.Errorf("the checker did not report %q on the mutated text; the control passed the checker, so the checker cannot see this defect", rule)
		}
	}
	// requireSiteAt asserts the sites rule names the 1-based line number.
	requireSiteAt := func(t *testing.T, mutated string, lineNo int) {
		t.Helper()
		want := fmt.Sprintf("line %d ", lineNo)
		for _, v := range asideViolationsFor(checkAsideE2E(mutated), asideRuleE2ESites) {
			if strings.Contains(v.Msg, want) {
				return
			}
		}
		t.Errorf("the sites rule did not name %q on the mutated text; the strip control passed the checker", want)
	}

	// Each required literal, removed in turn, must trigger its rule.
	for _, rule := range asideE2ERules {
		for i, lit := range asideE2ERequiredLiterals[rule] {
			t.Run(fmt.Sprintf("remove/%s/%d", rule, i), func(t *testing.T) {
				if !strings.Contains(e2eText, lit) {
					t.Fatalf("literal %q is not in the real e2e.md; the control would be vacuous", lit)
				}
				requireE2E(t, strings.ReplaceAll(e2eText, lit, ""), rule)
			})
		}
	}

	// The e2e-tester prohibition, removed, must trigger its rule.
	t.Run("remove/tester_sentence", func(t *testing.T) {
		if !strings.Contains(testerText, asideLitNeverInvokes) {
			t.Fatalf("literal %q is not in the real e2e-tester definition; the control would be vacuous", asideLitNeverInvokes)
		}
		mutated := strings.ReplaceAll(testerText, asideLitNeverInvokes, "")
		if len(asideViolationsFor(checkAsideTester(mutated), asideRuleSubagentTester)) == 0 {
			t.Errorf("the checker did not report %q on the mutated text", asideRuleSubagentTester)
		}
	})

	// Single-site carve-outs, stripped in turn (strip restores the original text).
	stripCases := []struct {
		name  string
		pred  func(string) bool
		rules []string
	}{
		{"strip/execution_owner", func(l string) bool { return strings.Contains(l, asideLitExecutionOwner) },
			[]string{asideRuleE2EOwner, asideRuleE2ESites}},
		{"strip/bounded_output", func(l string) bool { return strings.Contains(l, asideLitBoundedOutput) },
			[]string{asideRuleE2EBounded, asideRuleE2ESites}},
		{"strip/missing_toolchain_header", func(l string) bool { return strings.HasPrefix(strings.TrimSpace(l), asideLitHeaderPrefix) },
			[]string{asideRuleE2EHeader, asideRuleE2ESites}},
		{"strip/bypass_sentence_line", func(l string) bool { return strings.Contains(l, asideLitBypassSentence) },
			[]string{asideRuleE2EBypass, asideRuleE2ESites}},
		{"strip/autofix_validate_line", asideAutofixValidateLine,
			[]string{asideRuleE2EAutofix, asideRuleE2ESites}},
	}
	for _, c := range stripCases {
		t.Run(c.name, func(t *testing.T) {
			i := asideFindLine(lines, c.pred)
			if i < 0 {
				t.Fatalf("no such line in the real e2e.md; the control would be vacuous")
			}
			stripped := asideStripCarveOut(lines[i])
			if stripped == lines[i] {
				t.Fatalf("line %d carries no appended carve-out to strip; the control would be vacuous: %q", i+1, lines[i])
			}
			mutated := asideReplaceLine(lines, i, stripped)
			for _, rule := range c.rules {
				requireE2E(t, mutated, rule)
			}
		})
	}

	// The bounded-output carve-out must itself carry the runs directory.
	t.Run("mutate/bounded_tail_without_runs_dir", func(t *testing.T) {
		i := asideFindLine(lines, func(l string) bool { return strings.Contains(l, asideLitBoundedOutput) })
		if i < 0 {
			t.Fatalf("no bounded-output line in the real e2e.md; the control would be vacuous")
		}
		idx := strings.Index(lines[i], asideLitExceptAside)
		if idx < 0 {
			t.Fatalf("the bounded-output line carries no carve-out; the control would be vacuous: %q", lines[i])
		}
		mutated := lines[i][:idx] + strings.ReplaceAll(lines[i][idx:], asideLitRunsDir, "")
		requireE2E(t, asideReplaceLine(lines, i, mutated), asideRuleE2EBounded)
	})

	// The autofix carve-out must hand the re-verification to the orchestrator and
	// keep Aside away from the subagent, each part on its own.
	for _, c := range []struct{ name, removed string }{
		{"mutate/autofix_validate_without_orchestrator", asideLitOrchestratorCaps},
		{"mutate/autofix_validate_without_subagent_prohibition", asideLitNeverInvokes},
	} {
		t.Run(c.name, func(t *testing.T) {
			i := asideFindLine(lines, asideAutofixValidateLine)
			if i < 0 {
				t.Fatalf("no autofix Validate bullet in the real e2e.md; the control would be vacuous")
			}
			idx := strings.Index(lines[i], asideLitExceptAside)
			if idx < 0 {
				t.Fatalf("the autofix Validate bullet carries no carve-out; the control would be vacuous: %q", lines[i])
			}
			tail := strings.ReplaceAll(lines[i][idx:], c.removed, "")
			if tail == lines[i][idx:] {
				t.Fatalf("the carve-out of line %d does not carry %q; the control would be vacuous: %q", i+1, c.removed, lines[i])
			}
			requireE2E(t, asideReplaceLine(lines, i, lines[i][:idx]+tail), asideRuleE2EAutofix)
		})
	}

	// The named chain and summary sites, stripped in turn: the Phase 2 and Phase 3
	// chain lines, Execution Summary step 4, and the autofix sites (the Input and
	// Cycle bullets, the loop line, the Phase 3.5 chain line, and step 7.5).
	namedSites := []struct{ name, marker string }{
		{"strip/chain_phase2_line", "Phase 2: e2e-tester (script creation)"},
		{"strip/chain_phase3_line", "Phase 3: e2e-tester (CLI-first execution)"},
		{"strip/summary_step4_line", "Missing toolchain: probe"},
		{"strip/autofix_input_line", "- **Input**: failing journey"},
		{"strip/autofix_cycle_line", "- **Cycle**: localize"},
		{"strip/autofix_loop_line", "delegate grouped fixes"},
		{"strip/chain_phase35_line", "Phase 3.5 (--autofix only)"},
		{"strip/summary_step75_line", "7.5. (if --autofix"},
	}
	for _, s := range namedSites {
		t.Run(s.name, func(t *testing.T) {
			i := asideFindLine(lines, func(l string) bool { return strings.Contains(l, s.marker) })
			if i < 0 {
				t.Fatalf("no line carries %q in the real e2e.md; the control would be vacuous", s.marker)
			}
			stripped := asideStripCarveOut(lines[i])
			if stripped == lines[i] {
				t.Fatalf("line %d carries no appended carve-out to strip; the control would be vacuous: %q", i+1, lines[i])
			}
			requireSiteAt(t, asideReplaceLine(lines, i, stripped), i+1)
		})
	}

	// Every enumerated site, stripped in turn.
	t.Run("strip/every_enumerated_site", func(t *testing.T) {
		sites := asideSiteLineIndexes(lines)
		if len(sites) < asideE2EMinSites {
			t.Fatalf("the enumeration finds %d lines in the real e2e.md, at least %d are expected; the control would be vacuous", len(sites), asideE2EMinSites)
		}
		stripped := 0
		for _, i := range sites {
			if asideIsPrecedenceLine(lines[i]) {
				continue
			}
			s := asideStripCarveOut(lines[i])
			if s == lines[i] {
				t.Errorf("line %d matches the site pattern and carries no appended carve-out: %q", i+1, lines[i])
				continue
			}
			stripped++
			requireSiteAt(t, asideReplaceLine(lines, i, s), i+1)
		}
		if stripped < asideE2EMinSites-1 {
			t.Errorf("only %d sites were stripped; the control would be too narrow", stripped)
		}
	})

	// A site list below the minimum count must fail.
	t.Run("delete/sites_below_minimum", func(t *testing.T) {
		mutated := asideDropLines(lines, func(l string) bool { return asideSiteRe.MatchString(l) })
		if mutated == e2eText {
			t.Fatalf("no site line to delete; the control would be vacuous")
		}
		requireE2E(t, mutated, asideRuleE2ESites)
	})

	// `silently` is no substitute for the carve-out on a site line.
	t.Run("add/silently_site_line", func(t *testing.T) {
		requireE2E(t, e2eText+"\n- Delegate test execution to the e2e-tester silently.\n", asideRuleE2ESites)
	})
	t.Run("swap/silently_for_carve_out", func(t *testing.T) {
		i := asideFindLine(lines, func(l string) bool { return strings.Contains(l, "Phase 2: e2e-tester (script creation)") })
		if i < 0 {
			t.Fatalf("no Phase 2 chain line in the real e2e.md; the control would be vacuous")
		}
		stripped := asideStripCarveOut(lines[i])
		if stripped == lines[i] {
			t.Fatalf("line %d carries no appended carve-out to swap; the control would be vacuous: %q", i+1, lines[i])
		}
		requireSiteAt(t, asideReplaceLine(lines, i, stripped+" (silently)"), i+1)
	})
	t.Run("append/precedence_sentence_to_site_line", func(t *testing.T) {
		i := asideFindLine(lines, func(l string) bool { return strings.Contains(l, "Phase 3: e2e-tester (CLI-first execution)") })
		if i < 0 {
			t.Fatalf("no Phase 3 chain line in the real e2e.md; the control would be vacuous")
		}
		requireSiteAt(t, asideReplaceLine(lines, i, asideStripCarveOut(lines[i])+" "+asideLitPrecedence), i+1)
	})

	// Lines that would instruct a note, warning, or message about Aside.
	noteLines := []string{
		"- When Aside is absent, warn the operator.",
		"- Add a fallback note to the report when Aside is skipped.",
		"- Notify the operator when Aside is unavailable.",
		"- Report that Aside was skipped.",
		"- Tell the operator that Aside is not installed.",
		"- Inform the operator that Aside is unavailable.",
		"- Let the operator know when Aside fails its probe.",
	}
	for i, nl := range noteLines {
		t.Run(fmt.Sprintf("add/note_line_%d", i), func(t *testing.T) {
			requireE2E(t, e2eText+"\n"+nl+"\n", asideRuleE2ESilent)
		})
	}

	// Lines that would run or advise install tooling for Aside.
	installLines := []string{
		"- If Aside is absent, run npm i -g aside.",
		"- Fetch Aside with curl and run the installer.",
		"- Install Aside with brew when the probe fails.",
	}
	for i, il := range installLines {
		t.Run(fmt.Sprintf("add/install_line_%d", i), func(t *testing.T) {
			requireE2E(t, e2eText+"\n"+il+"\n", asideRuleE2ENoInstall)
		})
	}

	// The CI clause, removed, or moved into the no-flag branch only, must fail.
	t.Run("remove/ci_line", func(t *testing.T) {
		mutated := asideDropLines(lines, asideIsCILine)
		if mutated == e2eText {
			t.Fatalf("no CI exclusion line in the real e2e.md; the control would be vacuous")
		}
		requireE2E(t, mutated, asideRuleE2ECI)
	})
	t.Run("move/ci_line_to_no_flag_branch", func(t *testing.T) {
		ci := asideFindLine(lines, asideIsCILine)
		anchor := asideFindLine(lines, func(l string) bool { return strings.Contains(l, "environment detected") })
		if ci < 0 || anchor < 0 {
			t.Fatalf("CI line (%d) or no-flag CI anchor line (%d) not found in the real e2e.md; the control would be vacuous", ci, anchor)
		}
		var moved []string
		for i, l := range lines {
			if i == ci {
				continue
			}
			moved = append(moved, l)
			if i == anchor {
				moved = append(moved, lines[ci])
			}
		}
		requireE2E(t, strings.Join(moved, "\n"), asideRuleE2ECI)
	})

	// The screenshot evidence line, removed, must fail.
	t.Run("remove/screenshot_evidence_lines", func(t *testing.T) {
		mutated := asideDropLines(lines, func(l string) bool { return strings.Contains(l, asideLitScreenshot) })
		if mutated == e2eText {
			t.Fatalf("no screenshot evidence line in the real e2e.md; the control would be vacuous")
		}
		requireE2E(t, mutated, asideRuleE2EScreenshot)
	})
}
