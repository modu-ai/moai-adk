// todo_classify_llm.go — SPEC-TCD-LLM-DECIDER-001 M2: the LLM-backed
// classification decider (REQ-TLD-001): one in-process HTTPS call to the
// project's configured GLM endpoint, authenticated through the existing GLM
// credential reader, validated through the single classification validator
// with the decider identity FORCED to llm before validation (REQ-TLD-004 —
// a response claiming human/jev is NOT a failure; the identity is
// overwritten and the judgment accepted).
//
// Every LLM-side failure — absent credentials, connection failure, non-2xx,
// timeout, non-JSON, out-of-set fields — is an error return. The fallback
// branch stays in the seam (todo_classify.go's notice + fail-safe default),
// never synthesized inside the decider (plan OD-C): the entry points print
// the ONE notice and promote the defaults.
//
// @MX:SPEC: SPEC-TCD-LLM-DECIDER-001
package cli

import (
	"errors"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// llmCardDecider implements kanban.CardDecider over the configured GLM
// endpoint (REQ-TLD-001). It is stateless: the endpoint, credential, HTTP
// client, and timeout resolve through the package seams below, so tests
// inject fakes and no real network ever runs in the suite (plan §D.5).
type llmCardDecider struct{}

// newLLMCardDecider returns the LLM-backed decider the selector installs.
func newLLMCardDecider() llmCardDecider { return llmCardDecider{} }

// todoDeciderIsLLM reports whether dec is the LLM-backed decider whose
// judgment the entry points compute BEFORE the queue lock (REQ-TLD-005 /
// plan OD-D). Every other decider — deterministic, static, unavailable —
// resolves inside the locked write as before.
func todoDeciderIsLLM(dec kanban.CardDecider) bool {
	_, ok := dec.(llmCardDecider)
	return ok
}

// Classify asks the configured GLM endpoint to judge the card text and
// returns the validated classification with its decider identity forced to
// llm.
//
// @MX:TODO: [AUTO] transport implementation lands in M2 (request build,
// injectable doer + key-loader seams, response parse, validate); until then
// every judgment degrades to the seam's fail-safe branch.
func (llmCardDecider) Classify(string) (kanban.CardClassification, error) {
	return kanban.CardClassification{}, errors.New("llm decider: transport not implemented")
}
