---
title: /moai brain
weight: 25
draft: false
---

모호한 아이디어를 검증된 제안으로 변환하는 7단계 아이디에이션 워크플로우입니다.

{{< callout type="info" >}}
**슬래시 커맨드**: Claude Code에서 `/moai:brain`을 입력하면 이 명령어를 바로 실행할 수 있습니다.
{{< /callout >}}

## 개요

`/moai brain`은 `/moai plan` **이전**에 실행하는 사전 명세 워크플로우입니다. 아직 구체화되지
않은 아이디어를 7단계 파이프라인을 통해 검증된 제안으로 변환합니다. 최종 산출물은 Claude Design
핸드오프 패키지이며, 이후 `/moai design` 또는 `/moai project` → `/moai plan`으로 이어집니다.

### 워크플로우 위치

```
brain (1회) → [claude.com Design] → design Path A → project (1회) → plan (SPEC별) → run → sync
```

## 7단계 파이프라인

| 단계 | 이름 | 설명 | 산출물 |
|------|------|------|--------|
| 1 | **Discovery** | 소크라테스식 인터뷰로 아이디어 구체화 (최대 5라운드) | interview notes |
| 2 | **Diverge** | 가능한 접근 방식을 확장하여 탐색 | diverge notes |
| 3 | **Research** | WebSearch + Context7 병렬 조사 | research.md |
| 4 | **Converge** | 탐색 결과를 최적 접근법으로 수렴 | converge notes |
| 5 | **Critical Evaluation** | First Principles + 비판적 평가 | evaluation notes |
| 6 | **Proposal** | SPEC 분해 후보가 포함된 제안서 조립 | proposal.md |
| 7 | **Handoff** | Claude Design 5파일 핸드오프 패키지 생성 | 5-file bundle |

## 입력

`/moai brain <아이디어>` — 아이디어는 어떤 언어, 어떤 형태, 어떤 수준의 구체성이든 가능합니다.

```bash
# 예시
/moai brain "AI 코드 리뷰어 만들고 싶어"
/moai brain "E-commerce 플랫폼의 결제 시스템을 개선하고 싶다"
/moai brain "모바일 앱 푸시 알림 시스템"
```

## IDEA-NNN 자동 증가

워크플로우 시작 시 `.moai/brain/IDEA-NNN/` 디렉터리가 자동 생성됩니다.

```
.moai/brain/
├── IDEA-001/     # 첫 번째 아이디어
│   ├── interview.md
│   ├── research.md
│   ├── proposal.md
│   └── handoff/
├── IDEA-002/     # 두 번째 아이디어
│   └── ...
```

- 기존 IDEA 번호의 최대값 + 1로 자동 증가
- 3자리 zero-padding (IDEA-001, IDEA-002, ...)
- 중간에 중단된 경우 이어하기 제안

## 브랜드 컨텍스트 연동

`.moai/project/brand/`가 존재하면 브랜드 컨텍스트를 자동으로 로드하여 제안서에 반영합니다.

## Claude Design 핸드오프 (Phase 7)

Phase 7에서 생성되는 5파일 핸드오프 패키지:

| 파일 | 내용 |
|------|------|
| `prompt.md` | Claude Design 세션 프롬프트 |
| `context.md` | 프로젝트 컨텍스트 |
| `references.md` | 조사 결과 참고 자료 |
| `acceptance.md` | 수락 기준 |
| `checklist.md` | 디자인 체크리스트 |

이 패키지는 claude.com Design 세션에 직접 붙여넣어 사용할 수 있습니다.

## 이어하기 (Resume)

이전에 중단된 IDEA가 있으면 자동으로 감지하여 이어하기를 제안합니다.

## 관련 문서

- [/moai design](/ko/workflow-commands/moai-design) — Claude Design 핸드오프 수행
- [/moai project](/ko/workflow-commands/moai-project) — 프로젝트 초기 설정
- [/moai plan](/ko/workflow-commands/moai-plan) — SPEC 문서 생성
- [디자인 시스템](/ko/design/getting-started) — 디자인 파이프라인 개요
