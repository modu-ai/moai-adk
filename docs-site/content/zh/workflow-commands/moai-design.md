---
title: /moai design
weight: 26
draft: false
---

하이브리드 디자인 워크플로우로 브랜드 중심의 웹 경험을 생성합니다.

{{< callout type="info" >}}
**슬래시 커맨드**: Claude Code에서 `/moai:design`을 입력하면 이 명령어를 바로 실행할 수 있습니다.
{{< /callout >}}

## 개요

`/moai design`은 자연어 브리프를 브랜드 일관성이 있는 웹 경험으로 변환하는 하이브리드
디자인 워크플로우입니다. 3가지 경로를 지원합니다:

| 경로 | 설명 | 적합한 경우 |
|------|------|------------|
| **Path A** | Claude Design 핸드오프 | claude.com Design 세션 사용 |
| **Path B1** | Figma → 디자인 토큰 | 기존 Figma 디자인 있는 경우 |
| **Path B2** | Pencil MCP → 디자인 토큰 | Pencil 파일로 작업하는 경우 |
| **Path B-Common** | 코드 기반 브랜드 디자인 | 처음부터 시작하는 경우 |

## --mode 플래그

SPEC-V3R2-WF-003의 Multi-Mode Router를 지원합니다.

| 모드 | 설명 |
|------|------|
| `autopilot` (기본) | Path B — 코드 기반 브랜드 디자인 자동 실행 |
| `import` | Path A — Claude Design 핸드오프 번들 가져오기 |
| `team` | Path B를 병렬 팀으로 실행 (Agent Teams 필요) |

```bash
/moai design                          # 경로 선택 AskUserQuestion
/moai design --mode autopilot         # Path B 즉시 실행
/moai design --mode import            # Path A 핸드오프 가져오기
/moai design --mode team              # Path B 병렬 실행
```

## 실행 전 조건 확인

### 브랜드 컨텍스트

`.moai/project/brand/`에 3개 파일이 필요합니다:

```
.moai/project/brand/
├── brand-voice.md       # 브랜드 톤, 용어, 메시지 가이드
├── visual-identity.md   # 색상, 타이포그래피, 시각 언어
└── target-audience.md   # 타겟 고객 프로필
```

파일이 없거나 `_TBD_` 마커가 있으면 브랜드 인터뷰가 먼저 진행됩니다.

### Agency 데이터 감지

기존 `.agency/` 디렉터리가 있으면 마이그레이션을 제안합니다:

```bash
moai migrate agency --dry-run   # 미리보기
moai migrate agency             # 실행
```

## 파이프라인 아키텍처

```
manager-spec → [copywriting, brand-design] (병렬) → expert-frontend → evaluator-active
                                                                       ↑              |
                                                                       └──────────────┘
                                                                  GAN Loop (최대 5회)
```

1. **manager-spec** — 브리프 문서 생성 (Goal/Audience/Brand 섹션)
2. **copywriting + brand-design** — 카피 JSON + 디자인 토큰 (병렬 실행)
3. **expert-frontend** — 코드 구현
4. **evaluator-active** → **GAN Loop** — 품질 검증 반복

## GAN Loop

Builder-Evaluator 반복 루프가 품질을 보장합니다.

| 파라미터 | 기본값 | 설명 |
|-----------|--------|------|
| `pass_threshold` | 0.75 | 통과 임계값 (최소 0.60) |
| `max_iterations` | 5 | 최대 반복 횟수 |
| `escalation_after` | 3 | 사용자 개입 요청 시점 |
| `improvement_threshold` | 0.05 | 정체 감지 임계값 |

## Sprint Contract

`--mode team` 또는 `thorough` 하네스 레벨에서는 Sprint Contract가 협상됩니다:

1. evaluator-active가 수락 체크리스트 생성
2. expert-frontend가 체크리스트 리뷰
3. 구현 → 평가 반복
4. 통과한 기준은 다음 반복에서도 유지 (회귀 불가)

## 관련 문서

- [디자인 시작하기](/ko/design/getting-started) — 디자인 시스템 개요
- [Claude Design 핸드오프](/ko/design/claude-design-handoff) — Path A 상세 가이드
- [코드 기반 경로](/ko/design/code-based-path) — Path B 상세 가이드
- [GAN Loop](/ko/design/gan-loop) — 품질 검증 상세
- [/moai brain](/ko/workflow-commands/moai-brain) — 아이디에이션 워크플로우
