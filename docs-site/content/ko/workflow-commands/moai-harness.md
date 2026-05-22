---
title: /moai harness
weight: 55
draft: false
---

V3R4 Self-Evolving Harness 학습 시스템을 운영하는 명령어입니다. 4단계 진화 사다리(observer → heuristic → rule → frozen-zone)와 5계층 안전 파이프라인(frozen-guard → canary → contradiction → rate-limit → human oversight)을 안내합니다.

{{< callout type="info" >}}
**슬래시 커맨드**: Claude Code에서 `/moai harness`를 입력하면 이 명령어를 바로 실행할 수 있습니다.
{{< /callout >}}

## 개요

`/moai harness`는 MoAI-ADK의 자가 진화(Self-Evolving) 학습 서브시스템을 안전하게 운영하기 위한 4개의 verb(`status`, `apply`, `rollback`, `disable`)를 제공합니다. PostToolUse 훅이 관찰한 사용 로그(`.moai/harness/usage-log.jsonl`)를 바탕으로 4-tier 진화 사다리를 따라 제안이 분류되고, Tier-4 frozen-zone 변경은 반드시 사용자 승인(AskUserQuestion)을 거쳐서만 적용됩니다.

핵심 개념:

- **Observer**: PostToolUse 훅이 모든 도구 사용을 `.moai/harness/usage-log.jsonl`에 append-only로 기록합니다.
- **4-Tier Evolution Ladder**: 관찰 → 휴리스틱 → 규칙 → frozen-zone 제안의 4단계 분류.
- **5-Layer Safety Pipeline**: 모든 진화 제안은 5계층 안전 검증을 통과해야 적용됩니다.
- **CLI Retirement**: V3R4부터 모든 verb는 workflow body의 파일시스템 작업으로만 수행되며, Go 바이너리 서브커맨드는 호출되지 않습니다.

## 명령어 형식

```bash
/moai harness {status | apply | rollback <YYYY-MM-DD> | disable}
```

- 인자가 비어 있으면 도움말을 출력합니다.
- 모든 verb는 권한 있는 사용자(orchestrator 메인 컨텍스트)에서만 실행됩니다.

## 4 verbs 상세

### status

현재 harness 학습 상태와 보류 중인 Tier-4 제안 목록, 그리고 7일 rate-limit 윈도우 사용량을 출력합니다.

- **읽기 전용**: 파일을 수정하지 않습니다.
- **출력 정보**:
  - `learning.enabled` 설정 값 (`.moai/config/sections/harness.yaml`)
  - `.moai/harness/proposals/`에 보류 중인 Tier-4 제안 수
  - `.moai/harness/learning-history/applied/` 디렉토리 기준 7일 윈도우 내 적용 건수
  - 최근 tier 승격 이벤트(`tier-promotions.jsonl`)
  - Frozen Guard 위반 로그(`frozen-guard-violations.jsonl`)

### apply

가장 오래된 보류 Tier-4 제안 한 건을 5-Layer Safety 파이프라인에 통과시켜 적용합니다. 적용 전 반드시 orchestrator의 `AskUserQuestion` 라운드가 실행되며, 사용자가 명시적으로 동의해야 합니다.

- **사전 조건**:
  - 7일 윈도우 내 적용 건수가 1건 미만 (REQ-HRN-FND-012 rate-limit floor).
  - 제안 페이로드 무결성 검증 통과.
- **사용자 선택지(권장 / Modify / Defer / Reject)**: 첫 옵션은 `(권장)` 표시되며, Apply 시 사전 스냅샷이 `.moai/harness/learning-history/snapshots/<ISO-DATE>/`에 보관됩니다.

### rollback `<YYYY-MM-DD>`

지정된 날짜의 스냅샷을 사용해 직전 적용을 되돌립니다. 이미 다른 진화가 누적되어 있다면 충돌 보고서를 출력하고 사용자 승인을 다시 요청합니다.

- **인자**: ISO-8601 날짜 형식(YYYY-MM-DD). 형식 위반 시 에러.
- **효과**: `.moai/harness/learning-history/applied/<DATE>.json`이 `rolled-back/`로 이동하며, 영향받은 파일이 스냅샷 상태로 복원됩니다.

### disable

harness 학습을 일시 중단합니다(`learning.enabled: false`). PostToolUse 관찰 자체는 계속되지만 4-tier 분류기와 제안 생성기가 비활성화됩니다.

- **사용 시점**: 진화 제안이 의심스럽거나 외부 점검을 진행할 때.
- **재활성화**: `.moai/config/sections/harness.yaml`에서 `learning.enabled: true`로 복귀.

## 4-Tier Evolution Ladder

| Tier | 분류 | 자동 적용 여부 | 비고 |
|------|------|----------------|------|
| Tier-1 | Observation | n/a (수동 검토) | passive 로그 누적만 |
| Tier-2 | Heuristic | 제안만 표시 | orchestrator가 사용자에게 권유 |
| Tier-3 | Rule | non-frozen 영역만 자동 적용 가능 | canary 통과 필수 |
| Tier-4 | Frozen-zone | **사용자 승인 필수** | 5-Layer Safety 완주 후 적용 |

Frozen-zone은 `.claude/rules/moai/design/constitution.md` §2 및 `.claude/rules/moai/core/zone-registry.md`에 의해 정의됩니다.

## 5-Layer Safety Pipeline

1. **L1 Frozen Guard**: Frozen zone 영역에 대한 변경 시도 차단.
2. **L2 Canary**: 격리된 샌드박스에서 변경 영향 시뮬레이션.
3. **L3 Contradiction**: 다른 활성 규칙과의 충돌 탐지.
4. **L4 Rate Limit**: 7일 윈도우 내 최대 1회 적용(REQ-HRN-FND-012).
5. **L5 Human Oversight**: orchestrator 주도의 `AskUserQuestion` 승인 라운드.

5계층 중 어느 하나라도 거부되면 `apply`는 중단되고 제안은 `pending` 상태로 유지됩니다.

## 사용 예시

```bash
# 1) 현재 상태 점검
/moai harness status

# 2) 보류 중인 Tier-4 제안을 검토 후 적용
/moai harness apply

# 3) 직전 적용을 어제 날짜 스냅샷으로 되돌리기
/moai harness rollback 2026-05-21

# 4) 학습 일시 중단
/moai harness disable
```

## 관련 자료

- [`.claude/skills/moai/workflows/harness.md`](https://github.com/modu-ai/moai-adk) — workflow body SSOT
- [`SPEC-V3R4-HARNESS-001`](https://github.com/modu-ai/moai-adk) — V3R4 foundation SPEC (3개 V3R3 harness SPEC 통합)
- [`/moai plan`](/ko/workflow-commands/moai-plan) — SPEC 문서 생성
- [`/moai run`](/ko/workflow-commands/moai-run) — DDD/TDD 구현
- [`/moai sync`](/ko/workflow-commands/moai-sync) — 문서 동기화 + PR 생성
