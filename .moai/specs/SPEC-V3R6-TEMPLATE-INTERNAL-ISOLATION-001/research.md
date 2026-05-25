---
id: SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001
title: "Research — Template Internal-Content Isolation"
version: "0.1.2"
status: implemented
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template, isolation, research, predecessor-analysis"
tier: M
---

# Research — Template Internal-Content Isolation

## §A. 연구 목적

본 research 문서는 plan-phase 시점에 다음 3가지 질문에 답한다:

- **Q1**: Predecessor partial cleanup commits (`20a66df85` pass 1 + `40dc43f5b` pass 2)가 사용한 변환 패턴은 무엇이며, 본 SPEC이 이를 어떻게 확장해야 하는가?
- **Q2**: `internal/template/` 디렉토리에 본 SPEC의 lint test와 협업하거나 충돌할 수 있는 기존 audit test가 있는가?
- **Q3**: CLAUDE.local.md의 기존 isolation 정책 (§2 §15 §21 §24)이 본 SPEC §25와 어떻게 cross-reference 되어야 하는가?

## §B. Predecessor Cleanup Pattern 분석

### B.1 Pass 1 (`20a66df85`) — 9 files, -80 lines

Pass 1 commit message 분석에서 다음 substitution 패턴을 추출:

| File | Transformation pattern |
|------|----------------------|
| `.claude/rules/moai/NOTICE.md` | **전체 섹션 통째 삭제** — "Anthropic 2026 Alignment" 45-line 인용 section 제거 (Audit narrative 통째 삭제 패턴) |
| `CLAUDE.md` (template-managed) | SPEC ID 인용 제거 + Archive Cross-Reference generalization |
| `.claude/rules/moai/development/spec-frontmatter-schema.md` | 4-spot SPEC ID 제거 (Status Transition intro, Forward-looking enforcement, OwnershipTransitionRule default, Implementation files) |
| `.claude/rules/moai/core/agent-common-protocol.md` | Hook Invocation Surface SPEC ID + REQ-ATR-009/014 제거 + AC-ATR-022 grep example generalized |
| `.claude/rules/moai/workflow/archived-agent-rejection.md` | frontmatter tag + intro SPEC ID + archive date 제거 |
| `.claude/rules/moai/workflow/orchestration-mode-selection.md` | frontmatter tag + REQ-ATR-008/013/017 인용 제거 + design.md §B.4 cross-ref 제거 |
| `.claude/hooks/moai/sync-phase-quality-gate.sh` | header comment SPEC ID + REQ-ATR-009/014 제거 |
| `.claude/hooks/moai/status-transition-ownership.sh` | header comment SPEC ID + REQ-ATR-009 제거 |
| `.claude/hooks/moai/team-ac-verify.sh` | header comment SPEC ID + REQ-ATR-013/009 제거 |

**관찰**:

- Hook script header comment는 SPEC ID + REQ 인용을 단순 삭제 (학습 가치 0)
- Rule files의 inline SPEC ID 인용은 generic prose ("선행 SPEC", "the architectural decision")로 재구성
- NOTICE.md "Anthropic 2026 Alignment" 통째 섹션 제거는 가장 aggressive한 변환 — Audit narrative 자체가 메인테이너용이므로 사용자에게 의미 없음
- Frontmatter `tags:` 필드의 SPEC ID 인용은 단순 제거 (다른 generic tag로 대체하지 않음)

### B.2 Pass 2 (`40dc43f5b`) — 2 files

| File | Transformation pattern |
|------|----------------------|
| `agent-authoring.md` | 4-spot generalization (Finding A3 citation, decision tree retention ceiling, per-spawn pattern, archive cross-reference) — Finding A1-A6 citations were generalized to "an established best practice" form |
| `spec-frontmatter-schema.md` | 2 additional SPEC ID references 제거 (FrontmatterSchemaRule REQ-SPC-003-006 + enforcement header) |

**관찰**:

- Pass 2는 pass 1의 spec-frontmatter-schema 작업 cleanup을 보강 (incremental discovery 패턴 — 1회 audit으로 모든 leak 식별 어려움)
- Pass 2의 agent-authoring.md cleanup은 **"Finding A1-A6 citation generalization"** 패턴을 정립 — 본 SPEC §C 변환 dictionary S3 (Audit citation) rule의 직접 근거

### B.3 본 SPEC이 확장해야 할 영역

Predecessor가 다루지 않은 leak class:

- **C5 (Commit sha)**: predecessor commit message 분석에서 commit sha 인용 cleanup 사례 부재. 본 SPEC W1 M1 classification에서 발견 시 design.md §B.S5 rule 적용
- **C6 (Compound)**: predecessor가 NOTICE.md에서 section 통째 삭제로 implicit하게 처리했으나, 본 SPEC은 명시적 priority rule (S1 > S2 > S3 > S4 > S5) 도입
- **35 file 중 skill SKILL.md 6 files**: predecessor는 SKILL.md 직접 cleanup 없음. 본 SPEC W2 sub-batch 4.3 (.claude/skills/) 첫 시도가 됨

### B.4 Substitution vocabulary (predecessor에서 학습)

Pass 1 + Pass 2 diff에서 자주 등장한 generic substitute keyword (본 SPEC 일관성 vocabulary):

- "predecessor SPEC" / "선행 SPEC" — SPEC ID 대체 (영어 본문 / Korean 본문 각각)
- "established best practice" — Finding/Audit citation 대체
- "an architectural decision" — 구체적 SPEC-derived 결정 대체
- "earlier rebuild" — agent-team-rebuild 류 historical narrative 대체
- "the archive directory" — `agent-archive-2026-05-25/` 경로 대체

본 SPEC W2 cleanup은 이 vocabulary와 일관되게 적용 (AC-TII-002 검증 대상).

## §C. 기존 Template Test 조사

### C.1 `internal/template/` 디렉토리의 audit test inventory

다음 test files가 본 SPEC의 lint test와 협업 또는 충돌 가능:

| Test file | 목적 | 본 SPEC과의 관계 |
|----------|------|---------------|
| `lang_boundary_audit_test.go` | 16-language neutrality 강제 (§15 정책) | 협업 — 같은 `internal/template/templates/` 디렉토리 순회 패턴; `filepath.Walk` + regex 매칭 idiom 동일 |
| `embedded_namespace_test.go` | namespace separation 강제 (§24 정책) | 협업 — namespace 차원 격리; 본 SPEC은 content 차원 격리. 상호 보완 |
| `commands_audit_test.go` + `commands_root_audit_test.go` | command 파일 audit | 협업 — 본 SPEC W6 M6 maintainer-only audit (97/98/99 검사)와 부분 중복 가능; M6 시 inspection 후 중복 영역 식별 |
| `dev_only_skill_test.go` | dev-only skill 격리 | 협업 — §21 dev-only file 격리와 같은 의도; 본 SPEC §25는 보다 일반화된 isolation doctrine |
| `namespace_protection_audit_test.go` | namespace protection contract | 협업 — `moai update` namespace 보호 검사. 본 SPEC scope과 직교 |
| `hardcoded_path_audit_test.go` | hardcoded path detection | 협업 — 본 SPEC §14 hardcoding-prevention 정책과 같은 family; lint test pattern 재사용 가능 |
| `seq_thinking_retire_audit_test.go` | 특정 retired feature audit | 비교 사례 — 단일 feature audit pattern. 본 SPEC은 보다 광범위 (cross-feature) |
| `agent_frontmatter_audit_test.go` | agent frontmatter schema 검증 | 협업 — agent 파일 audit pattern 재사용 가능 |

### C.2 본 SPEC의 lint test 구조 (research-based)

`lang_boundary_audit_test.go`의 `TestLanguageNeutrality` 패턴을 참고하여 다음 구조 채택:

```go
// internal/template/internal_content_leak_test.go (proposed)
package template

import (
    "io/fs"
    "regexp"
    "strings"
    "testing"
)

// TestTemplateNoInternalContentLeak walks internal/template/templates/ and asserts
// no template file contains moai-adk dev-internal-content token.
//
// Sentinel: TEMPLATE_INTERNAL_CONTENT_LEAK_FORBIDDEN
// Policy: CLAUDE.local.md §25 (Template Internal-Content Isolation)
func TestTemplateNoInternalContentLeak(t *testing.T) {
    t.Parallel()

    fsys, err := EmbeddedTemplates()
    if err != nil {
        t.Fatalf("EmbeddedTemplates() error: %v", err)
    }

    leakPatterns := []*regexp.Regexp{
        regexp.MustCompile(`SPEC-V3R6-[A-Z0-9-]+-\d{3}`),
        regexp.MustCompile(`REQ-ATR-\d{3}`),
        regexp.MustCompile(`Audit \d+ Findings?`),
        regexp.MustCompile(`Finding A[1-6]`),
        regexp.MustCompile(`archive-2026-05-25`),
    }

    type allowlistEntry struct {
        path      string
        rationale string
    }
    allowlist := []allowlistEntry{
        // Empty by default — entries added per design.md §C.3 justification rules
    }

    walkErr := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
        if err != nil || d.IsDir() {
            return nil
        }
        // Read embedded file
        body, readErr := fs.ReadFile(fsys, path)
        if readErr != nil {
            return nil
        }
        text := string(body)
        for _, re := range leakPatterns {
            if loc := re.FindStringIndex(text); loc != nil {
                // Check allowlist
                exempted := false
                for _, entry := range allowlist {
                    if strings.HasSuffix(path, entry.path) {
                        exempted = true
                        break
                    }
                }
                if !exempted {
                    t.Errorf("TEMPLATE_INTERNAL_CONTENT_LEAK_FORBIDDEN: %s contains leak token %q at offset %d; see CLAUDE.local.md §25",
                        path, text[loc[0]:loc[1]], loc[0])
                }
            }
        }
        return nil
    })
    if walkErr != nil {
        t.Fatalf("WalkDir error: %v", walkErr)
    }
}
```

위 구조는 `lang_boundary_audit_test.go`의 검증된 idiom과 일관. 본 implementation은 M3 W4 단계의 baseline.

### C.3 충돌 가능성 평가

- **`commands_audit_test.go`** vs 본 SPEC AC-TII-009: 일부 path 검사 중복. M6 W6에서 `commands_audit_test.go` 가 검사하는 영역과 본 SPEC §21 dev-only check 영역의 boundary 확인 후 중복 시 본 SPEC test의 검사 범위 축소
- **`dev_only_skill_test.go`** vs 본 SPEC §25: skill 차원 dev-only 격리는 기존 테스트 담당; 본 SPEC §25는 content 차원 broader isolation. 상호 보완

## §D. CLAUDE.local.md 기존 Isolation 정책과의 cross-reference 설계

§25 작성 시 다음 cross-reference 필수:

| 기존 § | 관계 | §25 cross-reference 방식 |
|-------|------|----------------------|
| §2 Template-First Rule | "어떻게 update할지" 다룸 | §25 안에서 "§2 Template-First Rule에 의해 templates가 SSOT" 명시 |
| §15 Language Neutrality | "어떤 언어 허용" 다룸 (16-language 균등) | §25 안에서 "§15와 같은 isolation family, content 차원 추가" 명시 |
| §21 Dev-Only Commands Isolation | "어떤 파일 격리 (97/98/99 prefix)" 다룸 | §25 안에서 "§21이 file-class 차원, §25는 content-token 차원" 명시 |
| §22 Dev Settings Intent | "settings.local.json 의도 분리" 다룸 | §25는 §22의 운영 원칙 (3가지 [HARD] 항목) 패턴을 따라 작성 |
| §24 Harness Namespace 분리 | "namespace 차원" 다룸 | §25 안에서 "§24가 namespace 차원, §25는 content 차원" 명시 |

§25 자체는 다른 SPEC ID 인용 금지 (AC-TII-011) — 위 cross-reference는 generic § 번호 인용만 허용.

## §E. 사용자 학습 가치 평가 (외부화 필요 후보 식별)

본 plan-phase 시점 (W1 M1 classification 전) 외부화 후보 estimate:

| 후보 leak (개략) | 학습 가치 | 외부화 권장? |
|--------------|---------|-----------|
| `.claude/skills/moai-foundation-core/SKILL.md` line 248 — 12 archived agent enumeration | 중간 (사용자가 "왜 agent가 8개인가" 이해 도움) | Option A: generic prose ("Some agents were retired"); Option B 미적용 (정보 손실 minor) |
| `.claude/rules/moai/workflow/spec-workflow.md` line 17 — 8 retained / 12 archived 인용 | 낮음 (구체적 8 agent enum이 doctrine 본문에 필요 없음) | Option A: 단순 generic prose |
| `.claude/skills/moai-foundation-thinking/SKILL.md` — Audit citation 가능성 | 사실 확인 후 결정 | M1 W1 classification 단계에서 결정 |

이 평가는 W1 M1 단계에서 35 files 전수 분류 시 finalize. 본 plan-phase 시점에서는 estimate만 제시.

## §F. 결론 + plan-phase 권고

1. **Substitution dictionary 6-rule**은 predecessor 2-pass의 ad-hoc 변환을 systematic하게 격상; 일관성 보장
2. **Lint test는 `lang_boundary_audit_test.go` 패턴 재사용**으로 작성 — 검증된 idiom, 추가 학습 비용 0
3. **Allowlist는 default empty**로 시작; W1 M1 classification에서 발견 시에만 design.md §C 정당화 룰 적용하여 추가
4. **CLAUDE.local.md §25는 maintainer doctrine 영역**이므로 manager-develop 대신 orchestrator-direct 작성 권장 (plan.md §F M2 owner)
5. **CI integration은 기존 workflow 통합 우선**; 신규 workflow 신설은 over-engineering

## §G. 미해결 질문 (defer to run-phase)

- **Q-defer-1**: `.moai/docs/generic-patterns-guide.md` (template 내) 가 LNCO-001 산출물인데, 사용자 deploy 의도였는지 maintainer-internal 의도였는지 — W6 M6 maintainer audit pass에서 결정. 결정 시까지 default Option A (사용자 deploy 의도 가정 + leak token cleanup)
- **Q-defer-2**: `commands_audit_test.go` 와 본 SPEC M6 audit 영역 중복 정도 — M6 시작 시 inspection 후 결정. 중복 시 본 SPEC test의 dev-only file 검사 범위 축소 또는 commands_audit_test extending
- **Q-defer-3**: 35 files 중 일부가 사용자 학습 가치 매우 높을 시 외부화 (design.md §B.4 Option B) 가 필요할 수 있음 — W1 M1 classification에서 final 결정
