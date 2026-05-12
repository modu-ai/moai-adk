// Package harness provides validation utilities for the GAN loop harness.
// REQ-HRN-002-017: prior-judgment leak detection for evaluator spawn prompts.
package harness

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrPriorJudgmentLeak는 evaluator spawn 프롬프트에서
// 이전 iteration 판단 흔적이 탐지된 경우 반환되는 sentinel 오류입니다.
// design-constitution §11.4.1 및 REQ-HRN-002-017에 의해 정의됩니다.
var ErrPriorJudgmentLeak = errors.New("HRN_EVAL_PRIOR_JUDGMENT_LEAK: prior iteration judgment detected in evaluator spawn prompt (violates design-constitution §11.4.1)")

// forbiddenSubstrings는 evaluator spawn 프롬프트에서
// 이전 iteration 판단 흔적을 나타내는 금지 서브스트링 목록입니다.
var forbiddenSubstrings = []string{
	"Score:",
	"Feedback:",
	"Verdict:",
}

// iterationPattern은 숫자 iteration 참조 패턴을 탐지하는 정규식입니다.
// "Iteration 3", "iteration 2" 등을 탐지합니다.
var iterationPattern = regexp.MustCompile(`(?i)\bIteration\s+\d+`)

// priorEvaluatorPattern은 이전 evaluator 언급 패턴을 탐지하는 정규식입니다.
// "previous evaluator", "prior evaluator" 등을 탐지합니다.
var priorEvaluatorPattern = regexp.MustCompile(`(?i)\b(previous|prior)\s+evaluator`)

// DetectPriorJudgmentLeak는 evaluator spawn 프롬프트에서
// 이전 iteration 판단 흔적(score, feedback, verdict, iteration 번호 참조 등)을 탐지합니다.
// 탐지된 경우 ErrPriorJudgmentLeak을 반환합니다.
//
// @MX:WARN reason="prior-judgment leak detection per REQ-HRN-002-017"
// @MX:REASON: design-constitution §11.4.1 — prior iteration judgment rationales, scoring internals,
// or reflection traces MUST NOT appear in the evaluator's context window.
func DetectPriorJudgmentLeak(spawnPrompt string) error {
	// 금지 서브스트링 탐지
	for _, sub := range forbiddenSubstrings {
		if strings.Contains(spawnPrompt, sub) {
			return fmt.Errorf("%w: forbidden substring %q found in spawn prompt", ErrPriorJudgmentLeak, sub)
		}
	}

	// 숫자 iteration 참조 패턴 탐지
	if loc := iterationPattern.FindStringIndex(spawnPrompt); loc != nil {
		matched := spawnPrompt[loc[0]:loc[1]]
		return fmt.Errorf("%w: forbidden iteration reference %q found in spawn prompt", ErrPriorJudgmentLeak, matched)
	}

	// 이전 evaluator 언급 패턴 탐지
	if loc := priorEvaluatorPattern.FindStringIndex(spawnPrompt); loc != nil {
		matched := spawnPrompt[loc[0]:loc[1]]
		return fmt.Errorf("%w: forbidden prior-evaluator reference %q found in spawn prompt", ErrPriorJudgmentLeak, matched)
	}

	return nil
}
