// todo_decider_select_test.go — SPEC-TCD-LLM-DECIDER-001 M1: the standing
// decider selection surface (AC-TLD-002 selection matrix, AC-TLD-007
// default-off regression guard). The selector reads MOAI_TODO_DECIDER once
// per add invocation; these tests drive it through t.Setenv and therefore
// stay NON-parallel (process-env mutation).
package cli

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// unsetTodoDeciderEnvForTest registers the env restore with t.Setenv and then actually
// removes the variable, so a test can observe the UNSET state without leaking
// a value into sibling tests.
func unsetTodoDeciderEnvForTest(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "") // registers the post-test restore
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
}

// TestTodoDeciderSelection_UnsetEnvKeepsDeterministicDefault — AC-TLD-002(a)
// and AC-TLD-007: with the selector variable UNSET the add path keeps the
// deterministic default decider — the dormant-default framing (REQ-TLD-007).
func TestTodoDeciderSelection_UnsetEnvKeepsDeterministicDefault(t *testing.T) {
	unsetTodoDeciderEnvForTest(t, config.EnvTodoDecider)

	dec, err := todoDeciderFromEnv()
	if err != nil {
		t.Fatalf("unset selection errored: %v", err)
	}
	if _, ok := dec.(factory.DefaultCardDecider); !ok {
		t.Errorf("unset selection returned %T, want the deterministic factory.DefaultCardDecider", dec)
	}
}

// TestTodoDeciderSelection_EmptyAndDefaultValuesKeepDefaultDecider —
// AC-TLD-002(b): an empty value and the explicit "default" both resolve to
// the deterministic default, never to a refusal.
func TestTodoDeciderSelection_EmptyAndDefaultValuesKeepDefaultDecider(t *testing.T) {
	for _, value := range []string{"", factory.DeciderIdentityDefault} {
		t.Setenv(config.EnvTodoDecider, value)
		dec, err := todoDeciderFromEnv()
		if err != nil {
			t.Fatalf("selection for %q errored: %v", value, err)
		}
		if _, ok := dec.(factory.DefaultCardDecider); !ok {
			t.Errorf("selection for %q returned %T, want the deterministic default", value, dec)
		}
	}
}

// TestTodoDeciderSelection_LLMValueSelectsLLMDecider — AC-TLD-002(c): the
// "llm" value selects the LLM-backed decider.
func TestTodoDeciderSelection_LLMValueSelectsLLMDecider(t *testing.T) {
	t.Setenv(config.EnvTodoDecider, factory.DeciderIdentityLLM)

	dec, err := todoDeciderFromEnv()
	if err != nil {
		t.Fatalf("llm selection errored: %v", err)
	}
	if !todoDeciderIsLLM(dec) {
		t.Errorf("llm selection returned %T, want the LLM-backed decider", dec)
	}
}

// TestTodoDeciderSelection_UnknownValueIsUsageRefusal — AC-TLD-002(d): an
// operator-authored value outside the accepted set is refused as a usage
// error (exit 2) naming the accepted set, with nothing written — an
// operator-authored misconfiguration fails loud, never silently reverts to
// the default (the EnvClaudeBin precedent).
func TestTodoDeciderSelection_UnknownValueIsUsageRefusal(t *testing.T) {
	t.Setenv(config.EnvTodoDecider, "banana")

	_, err := todoDeciderFromEnv()
	var refusal *exitCodeError
	if !errors.As(err, &refusal) || refusal.ExitCode() != 2 {
		t.Fatalf("value %q produced %v, want a usage exit-2 refusal", "banana", err)
	}
	for _, want := range []string{"banana", factory.DeciderIdentityDefault, factory.DeciderIdentityLLM} {
		if !strings.Contains(refusal.Error(), want) {
			t.Errorf("refusal message %q does not name %q", refusal.Error(), want)
		}
	}
}

// TestTodoDeciderSelection_JevValuesAreNamedRefusals — AC-TLD-002(e): the
// jev identities are refused with the parent SPEC's named refusal wording,
// never accepted and never silently defaulted.
func TestTodoDeciderSelection_JevValuesAreNamedRefusals(t *testing.T) {
	for _, value := range []string{factory.DeciderIdentityJev, factory.DeciderIdentityLLMJev} {
		t.Setenv(config.EnvTodoDecider, value)
		_, err := todoDeciderFromEnv()
		var refusal *exitCodeError
		if !errors.As(err, &refusal) || refusal.ExitCode() != 2 {
			t.Fatalf("value %q produced %v, want a usage exit-2 refusal", value, err)
		}
		if !strings.Contains(refusal.Error(), "jev is never a product classification decider") {
			t.Errorf("refusal for %q does not carry the named jev refusal wording: %q", value, refusal.Error())
		}
	}
}

// TestTodoAdd_SuppliedFileBeatsLLMSelection — AC-TLD-002(f): where
// --classification-file is supplied together with the llm selection, the
// supplied judgment file wins and its own decider identity is recorded.
func TestTodoAdd_SuppliedFileBeatsLLMSelection(t *testing.T) {
	_, store := todoFixture(t)
	t.Setenv(config.EnvTodoDecider, factory.DeciderIdentityLLM)
	path := writeClassificationFile(t, `{"priority":"low","blocked":true,"mode":"parallelizable","decider":"human","reason":"operator said so"}`)

	out, _, err := runTodo(t, "add", "file beats env card", "--classification-file", path)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	id := strings.TrimSpace(strings.SplitN(out, " ", 2)[0])

	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range rec.Items {
		if it.ID != id {
			continue
		}
		c := it.Classification
		if c == nil || c.Decider != factory.DeciderIdentityHuman || c.Priority != factory.ClassPriorityLow {
			t.Errorf("classification = %+v, want the SUPPLIED file judgment (the file beat the llm selection)", c)
		}
	}
}

// TestTodoAdd_UnsetEnvBehavesPreSPEC — AC-TLD-007: with the selector unset
// the add records the deterministic default classification and stderr carries
// no fallback notice — the pre-SPEC behavior is unchanged (a healthy default
// judgment is not a failure, so it prints none).
func TestTodoAdd_UnsetEnvBehavesPreSPEC(t *testing.T) {
	_, store := todoFixture(t)
	unsetTodoDeciderEnvForTest(t, config.EnvTodoDecider)

	out, errOut, err := runTodo(t, "add", "dormant default card")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if n := strings.Count(errOut, "classification decider unavailable"); n != 0 {
		t.Errorf("unset-env add printed %d fallback notices, want 0 (stderr: %q)", n, errOut)
	}
	id := strings.TrimSpace(strings.SplitN(out, " ", 2)[0])
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range rec.Items {
		if it.ID != id {
			continue
		}
		c := it.Classification
		if c == nil || c.Decider != factory.DeciderIdentityDefault {
			t.Errorf("unset-env classification = %+v, want the deterministic default", c)
		}
	}
}
