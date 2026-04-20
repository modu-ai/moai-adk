---
title: CG 모드 (Claude + GLM)
weight: 20
draft: false
---
# CG 모드 (Claude + GLM)

## CG 모드란?

CG(Claude + GLM) 모드는 리더가 **Claude API**를 사용하고 워커가 **GLM API**를 사용하는 하이브리드 모드입니다. tmux 세션 수준의 환경 변수 격리를 통해 구현됩니다.

## 아키텍처

```
moai cg 실행
    │
    ├── 1. GLM 설정을 tmux 세션 환경변수에 주입
    │      (ANTHROPIC_AUTH_TOKEN, BASE_URL, MODEL_* 변수)
    │
    ├── 2. settings.local.json에서 GLM 환경변수 제거
    │      → 리더 pane은 Claude API 사용
    │
    ├── 3. CLAUDE_CODE_TEAMMATE_DISPLAY=tmux 설정
    │      → 워커들은 새 pane에서 GLM 환경변수 상속
    │
    └── 4. Claude Code 실행 (현재 프로세스 대체)
```

```
┌─────────────────────────────────────────────────────────────┐
│  리더 (현재 tmux pane, Claude API)                           │
│  - /moai --team 실행 시 워크플로우 오케스트레이션             │
│  - plan, quality, sync 단계 처리                             │
│  - GLM 환경변수 없음 → Claude API 사용                       │
└──────────────────────┬──────────────────────────────────────┘
                       │ Agent Teams (새 tmux pane)
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  팀원 (새 tmux pane, GLM API)                                │
│  - tmux 세션 환경변수 상속 → GLM API 사용                     │
│  - run 단계에서 구현 작업 실행                                │
│  - SendMessage로 리더와 통신                                  │
└─────────────────────────────────────────────────────────────┘
```

## 사용 방법

### 1단계: GLM API 키 저장 (최초 1회)

```bash
moai glm sk-your-glm-api-key
```

키는 `~/.moai/.env.glm`에 안전하게 저장됩니다.

### 2단계: tmux 환경 확인

이미 tmux를 사용 중이라면 새 세션을 만들 필요가 없습니다.

```bash
# tmux를 사용 중이 아니라면:
tmux new -s moai
```

> **팁**: VS Code 터미널 기본값을 tmux로 설정하면 이 단계를 완전히 건너뛸 수 있습니다.

### 3단계: CG 모드 실행

```bash
moai cg
```

`moai cg`는 현재 pane에서 자동으로 Claude Code를 실행합니다. 별도로 `claude`를 실행할 필요가 없습니다.

### 4단계: 팀 워크플로우 실행

```bash
/moai --team "사용자 인증 기능 구현"
```

## 중요 사항

| 항목 | 설명 |
|------|------|
| **tmux 환경** | 이미 tmux를 사용 중이면 새 세션 불필요. VS Code 터미널 기본값을 tmux로 설정하면 편리 |
| **자동 실행** | `moai cg`가 현재 pane에서 Claude Code를 자동 실행. 별도 `claude` 명령 불필요 |
| **세션 종료** | session_end 훅이 자동으로 tmux 세션 환경변수 정리 → 다음 세션은 Claude 사용 |
| **팀 통신** | SendMessage 도구로 리더↔워커 간 통신 |
| **모드 전환** | `moai glm`에서 전환 시 `moai cg`가 GLM 설정을 자동 초기화 — 중간에 `moai cc` 불필요 |

## 디스플레이 모드

Agent Teams는 두 가지 디스플레이 모드를 지원합니다:

| 모드 | 설명 | 통신 | 리더/워커 분리 |
|------|------|------|--------------|
| `in-process` | 기본 모드, 모든 터미널 | ✅ SendMessage | ❌ 같은 환경 |
| `tmux` | 분할 화면 표시 | ✅ SendMessage | ✅ 세션 환경변수 격리 |

> **CG 모드는 `tmux` 디스플레이 모드에서만 리더/워커 API 분리가 가능합니다.**

## 모드 비교

| 명령어 | 리더 | 워커 | tmux 필요 | 비용 절감 | 용도 |
|--------|------|------|----------|----------|------|
| `moai cc` | Claude | Claude | 아니오 | - | 복잡한 작업, 최고 품질 |
| `moai glm` | GLM | GLM | 권장 | ~70% | 비용 최적화 |
| `moai cg` | Claude | GLM | **필수** | **~60%** | 품질 + 비용 균형 |

### 언제 CG 모드를 사용해야 하나요?

**CG 모드 적합:**
- 구현 중심의 SPEC 실행 (run 단계)
- 코드 생성 작업
- 테스트 작성
- 문서 생성

**Claude 전용(cc) 적합:**
- 아키텍처 설계/계획 (Opus 추론 필요)
- 보안 리뷰 (Claude의 보안 트레이닝 필요)
- 복잡한 디버깅 (고급 추론 필요)

## 문제 해결

| 문제 | 원인 | 해결 |
|------|------|------|
| 워커가 Claude API 사용 | tmux 세션 환경변수 미설정 | tmux 내에서 `moai cg` 재실행 |
| `moai cg` 후 Claude Code 미실행 | tmux 외부에서 실행 | `tmux new -s moai` 후 재실행 |
| 세션 종료 후 GLM 환경변수 잔류 | session_end 훅 실패 | `moai cc`로 수동 정리 |

## 다음 단계

- [모델 정책](/ko/multi-llm/model-policy) — 에이전트별 모델 배정
- [이중 실행 모드](/ko/getting-started/faq) — Sub-Agent vs Agent Teams
- [CLI 레퍼런스](/ko/getting-started/cli) — moai cc, moai glm, moai cg 상세
