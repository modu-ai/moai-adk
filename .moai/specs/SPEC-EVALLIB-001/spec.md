---
spec_id: SPEC-EVALLIB-001
title: "Evaluator Prompt Library"
created: "2026-04-01"
status: approved
priority: high
module: template
version: "1.0.0"
lifecycle: spec-anchored
---

# Evaluator Prompt Library

## Overview

evaluator-active 에이전트가 프로젝트 유형과 도메인에 맞는 평가 기준을 동적으로 로드할 수 있도록 Evaluator Profile 라이브러리를 구축한다. 기본 프로필(default), 엄격 프로필(strict), 관대 프로필(lenient), 프론트엔드 전용 프로필(frontend) 4종을 제공하며, 특히 프론트엔드 프로필에는 AI 생성 코드의 일반적 패턴(AI-slop)을 명시적으로 감점하는 기준을 포함한다. 평가 결과는 JSONL 형식으로 누적 기록되어 장기적 프로필 튜닝에 활용된다.

## Requirements (EARS Format)

### REQ-001: Evaluator Profile Directory

**When** a MoAI-ADK project is deployed, the system **shall** create a `.moai/config/evaluator-profiles/` directory containing 4 default profile files:

- `default.md`: Balanced evaluation (Functionality 40%, Security 25%, Craft 20%, Consistency 15%)
- `strict.md`: Elevated thresholds (Security 35%, all dimensions require 80%+ to pass)
- `lenient.md`: Reduced thresholds for prototyping (Functionality 60%, Security 20%, Craft 10%, Consistency 10%)
- `frontend.md`: Frontend-specific criteria with anti-AI-slop detection

**When** a SPEC has an `evaluator_profile` field in its frontmatter, the system **shall** load the matching profile.
**When** no `evaluator_profile` is specified, the system **shall** use the default profile from harness.yaml configuration.

**Acceptance**:
- **Given** a deployed MoAI-ADK project
- **When** the evaluator-profiles directory is inspected
- **Then** all 4 profile files exist with complete evaluation criteria
- **Given** a SPEC with `evaluator_profile: strict` in frontmatter
- **When** evaluator-active runs
- **Then** the strict.md profile criteria are applied
- **Given** a SPEC with no `evaluator_profile` field
- **When** evaluator-active runs
- **Then** the default.md profile is used

### REQ-002: Frontend Anti-AI-Slop Criteria

**When** the frontend.md profile is active, the system **shall** evaluate implementations against 3 dimensions:
- Originality (40%): Penalizes generic AI patterns
- Design Quality (30%): Evaluates visual design coherence and intentionality
- Craft & Functionality (30%): Evaluates code quality and interactive behavior

The profile **shall** explicitly penalize the following AI-slop patterns:
- Stock card UI layouts without design justification
- Default Bootstrap/Tailwind utility-class-only styling without custom design tokens
- Purple/blue gradient backgrounds without brand alignment
- Generic placeholder text ("Lorem ipsum", "Welcome to our platform")
- Identical component structure across unrelated sections
- Missing hover/focus/active states on interactive elements

**Acceptance**:
- **Given** a frontend implementation using stock card layout with purple gradient and generic text
- **When** evaluator-active runs with frontend.md profile
- **Then** the Originality dimension score is below passing threshold
- **And** specific AI-slop patterns are cited in the evaluation report
- **Given** a frontend implementation with custom design tokens, intentional layout, and brand-aligned colors
- **When** evaluator-active runs with frontend.md profile
- **Then** the Originality dimension score is above passing threshold

### REQ-003: Iterative Improvement Workflow

**When** evaluator-active completes an evaluation, the system **shall** log the result to `.moai/metrics/evaluation-log.jsonl` with the following fields:
- timestamp
- spec_id
- profile_used
- dimension_scores (per-dimension breakdown)
- overall_verdict (PASS/FAIL)
- flags (list of triggered rules)

**When** the evaluation log reaches 50 SPEC entries, the system **shall** generate a summary analysis for False Positive/False Negative tuning recommendations.

**Acceptance**:
- **Given** evaluator-active completes an evaluation
- **When** the evaluation finishes
- **Then** a JSONL entry is appended to `.moai/metrics/evaluation-log.jsonl`
- **And** the entry contains all required fields
- **Given** the evaluation log contains 50 entries
- **When** the periodic analysis trigger fires
- **Then** a summary is generated identifying potential False Positive and False Negative patterns
- **And** profile tuning recommendations are produced (but not auto-applied)

## Exclusions (What NOT to Build)

- Shall NOT auto-modify profiles based on analysis results (human review required for profile changes)
- Shall NOT create custom profiles automatically (users create custom profiles manually)
- Shall NOT evaluate non-code artifacts (documentation, images, design mockups)
- Shall NOT override SPEC-level evaluator_profile with harness-level defaults

## Acceptance Criteria

**Scenario 1: Profile Selection via SPEC Frontmatter**
- Given a SPEC with `evaluator_profile: frontend` in frontmatter
- When evaluator-active is invoked
- Then frontend.md profile criteria are loaded
- And the 3 frontend-specific dimensions (Originality, Design Quality, Craft & Functionality) are used

**Scenario 2: Default Profile Fallback**
- Given a SPEC with no evaluator_profile field
- And harness.yaml specifies default_profile: "default"
- When evaluator-active is invoked
- Then default.md profile criteria are loaded
- And the 4 standard dimensions are used

**Scenario 3: Anti-AI-Slop Detection**
- Given a React component with stock card layout, purple gradient, and "Welcome to our platform" text
- When evaluator-active evaluates with frontend.md profile
- Then at least 3 AI-slop flags are raised
- And Originality score is below 50%
- And the evaluation report lists each detected pattern

**Scenario 4: Evaluation Logging**
- Given 3 consecutive SPEC evaluations
- When all evaluations complete
- Then evaluation-log.jsonl contains exactly 3 new entries
- And each entry has valid JSON with all required fields

## Technical Approach

1. Create evaluator-profiles directory with 4 Markdown profile files in template
2. Define profile schema: dimensions, weights, thresholds, penalty_patterns
3. Add profile loading logic to evaluator-active.md agent definition
4. Implement JSONL logging for evaluation results in run.md Phase 2.8a
5. Design periodic analysis trigger (every 50 entries) as advisory output

## Files to Modify

- `internal/template/templates/.moai/config/evaluator-profiles/default.md` (NEW)
- `internal/template/templates/.moai/config/evaluator-profiles/strict.md` (NEW)
- `internal/template/templates/.moai/config/evaluator-profiles/lenient.md` (NEW)
- `internal/template/templates/.moai/config/evaluator-profiles/frontend.md` (NEW)
- `internal/template/templates/.claude/agents/moai/evaluator-active.md` (profile loading logic)
