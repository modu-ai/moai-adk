package hook

// contract_sign_guard_units_test.go — parser unit tables for the
// contract-sign and contract-decide guard (SPEC-AUTONOMY-PRECONDITION-001
// M2). The AC tests (contract_sign_guard_test.go) exercise the AC fixture
// commands end to end; these tables cover the wrapper-option skippers'
// remaining limbs (split and glued option forms, the option that stops the
// skip, the operand rows of design.md §C.3), the -c payload extraction, the
// --signer value forms, and the tokenizer's quote-removal edges — the
// limbs the AC fixture set does not reach.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestContractSignParserVerdicts walks classifyContractCall over parser
// limbs, asserting the verdict fields rather than the deny decision — the
// decision is the AC tests' subject.
func TestContractSignParserVerdicts(t *testing.T) {
	cases := []struct {
		name         string
		command      string
		found        bool
		verb         string
		signer       string
		classified   bool
		unclassified bool
	}{
		// env: -u NAME split and glued, assignment + -i mix.
		{name: "env -u split", command: "env -u FOO moai contract sign", found: true, verb: "sign", classified: true},
		{name: "env -u glued", command: "env -uFOO moai contract sign", found: true, verb: "sign", classified: true},
		{name: "env assignment + -i decide", command: "env FOO=1 -i moai contract decide", found: true, verb: "decide", classified: true},

		// command: each own option, and a non-invoking tail.
		{name: "command -p", command: "command -p moai contract sign", found: true, verb: "sign", classified: true},
		{name: "command -v lookup", command: "command -v moai", found: false},

		// exec: -a NAME, -c, -l, glued pure form.
		{name: "exec -a NAME", command: "exec -a /usr/bin/moai moai contract sign", found: true, verb: "sign", classified: true},
		{name: "exec -c", command: "exec -c moai contract sign", found: true, verb: "sign", classified: true},
		{name: "exec -l", command: "exec -l moai contract sign", found: true, verb: "sign", classified: true},
		{name: "exec -lc glued", command: "exec -lc moai contract sign", found: true, verb: "sign", classified: true},

		// timeout: split flags with values, glued form, duration operand.
		{name: "timeout full flags", command: "timeout -k 5m -s KILL --preserve-status 30 moai contract sign", found: true, verb: "sign", classified: true},
		{name: "timeout -k glued", command: "timeout -k5m 30 moai contract sign", found: true, verb: "sign", classified: true},
		{name: "timeout unknown flag stops", command: "timeout -v 30 moai contract sign", found: true, verb: "sign", unclassified: true},

		// sudo: -u USER, valueless flags, --, glued pure form.
		{name: "sudo full flags", command: "sudo -u root -E -H -- moai contract sign", found: true, verb: "sign", classified: true},
		{name: "sudo -nE glued", command: "sudo -nE moai contract sign", found: true, verb: "sign", classified: true},
		{name: "sudo -u glued", command: "sudo -uroot moai contract sign", found: true, verb: "sign", classified: true},
		{name: "sudo unknown flag stops", command: "sudo -C 3 moai contract sign", found: true, verb: "sign", unclassified: true},

		// stdbuf: split MODE values and glued MODE.
		{name: "stdbuf split modes", command: "stdbuf -o 1M -e 0 moai contract sign", found: true, verb: "sign", classified: true},
		{name: "stdbuf -i glued", command: "stdbuf -i0 moai contract sign", found: true, verb: "sign", classified: true},

		// nice: split and glued.
		{name: "nice -n split", command: "nice -n 5 moai contract sign", found: true, verb: "sign", classified: true},
		{name: "nice -n glued", command: "nice -n5 moai contract sign", found: true, verb: "sign", classified: true},

		// xargs: -0, -I STR, -P N, glued -n.
		{name: "xargs full flags", command: "xargs -0 -I {} -P 2 moai contract sign", found: true, verb: "sign", classified: true},
		{name: "xargs -n glued", command: "xargs -n2 moai contract sign", found: true, verb: "sign", classified: true},

		// -c payload extraction: options before -c, -c without payload,
		// shells without -c.
		{name: "sh options before -c", command: "sh --norc -c 'moai contract sign'", found: true, verb: "sign", classified: true},
		{name: "sh -c no payload", command: "sh -c", found: false},
		{name: "bash without -c", command: "bash moai contract sign", found: false},
		{name: "bare sh", command: "sh", found: false},

		// Chained wrappers.
		{name: "sudo + env chain", command: "sudo -n env FOO=1 moai contract sign", found: true, verb: "sign", classified: true},

		// Substitution without a carried call is allowed; a substitution
		// word whose CONTENT carries the call is unclassified.
		{name: "substitution no call", command: "echo $(date) ok", found: false},
		{name: "backtick carries call", command: "`echo moai contract sign`", found: true, verb: "sign", unclassified: true},

		// --signer value forms, read from the same word list.
		{name: "signer glued", command: "moai contract sign --signer=llm SPEC-X-001", found: true, verb: "sign", signer: "llm", classified: true},
		{name: "signer human split", command: "moai contract sign --signer human --receipt /tmp/r.json", found: true, verb: "sign", signer: "human", classified: true},

		// Quoting edges: mixed quote styles still match; an escaped space
		// joins the program word, so no call follows.
		{name: "mixed quotes", command: `"moai" 'contract' sign`, found: true, verb: "sign", classified: true},
		{name: "escaped space joins word", command: `moai\ contract sign`, found: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := classifyContractCall(tc.command)
			if v.found != tc.found {
				t.Fatalf("found = %v, want %v (verdict %+v)", v.found, tc.found, v)
			}
			if !tc.found {
				return
			}
			if v.verb != tc.verb {
				t.Errorf("verb = %q, want %q", v.verb, tc.verb)
			}
			if v.signer != tc.signer {
				t.Errorf("signer = %q, want %q", v.signer, tc.signer)
			}
			if v.classified != tc.classified {
				t.Errorf("classified = %v, want %v", v.classified, tc.classified)
			}
			if tc.unclassified && v.classified {
				t.Errorf("verdict %+v is classified, want unclassified", v)
			}
		})
	}
}

// TestContractSignScanCallPriority pins the scoping priority of the
// unclassified verb scan: when a command carries both verbs, the sign rule —
// always in force — wins over the decide rule.
func TestContractSignScanCallPriority(t *testing.T) {
	got := scanCallWords([]string{"moai", "contract", "decide", "then", "moai", "contract", "sign"})
	if got != "sign" {
		t.Fatalf("scanCallWords = %q, want %q — the human-path rule is always in force", got, "sign")
	}
}

// TestContractSignGuardInputEdges covers checkContractSign's input edges and
// the fail-closed unknown-signer limb: nil input, an input with no command
// payload, an empty command, and a --signer value outside A1's closed decider
// set — which is denied even in an unmarked session, because it cannot be
// shown to be off the agent paths (design.md §C.6).
func TestContractSignGuardInputEdges(t *testing.T) {
	if decision, reason := checkContractSign(nil); decision != "" || reason != "" {
		t.Fatalf("nil input: decision = %q reason = %q, want empty", decision, reason)
	}
	if decision, reason := checkContractSign(&HookInput{HookEventName: "PreToolUse", ToolName: "Bash"}); decision != "" || reason != "" {
		t.Fatalf("no tool input: decision = %q reason = %q, want empty", decision, reason)
	}
	if decision, reason := checkContractSign(signGuardInput(t, "")); decision != "" || reason != "" {
		t.Fatalf("empty command: decision = %q reason = %q, want empty", decision, reason)
	}
	t.Setenv(config.EnvFactoryRole, "")
	decision, reason := checkContractSign(signGuardInput(t, "moai contract sign --signer machine SPEC-X-001"))
	if decision != DecisionDeny || !strings.HasPrefix(reason, contractSignViolationPrefix) || !strings.Contains(reason, "unclassified") {
		t.Fatalf("unknown --signer value: decision = %q reason = %q, want the fail-closed deny with the unclassified mark", decision, reason)
	}
}
