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

// readAsideSkill reads the template skill file. A read failure is returned, not
// fatal, so the caller can fail every named subtest.
func readAsideSkill(t *testing.T) (string, error) {
	t.Helper()
	root := findNeutralityRoot(t) // .../internal/template/templates
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(asideSkillRelPath)))
	if err != nil {
		return "", err
	}
	return string(data), nil
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
}
