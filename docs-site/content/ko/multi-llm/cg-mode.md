---
title: CG 모드 (Claude + GLM)
weight: 20
draft: false
---

## CG 모드란?

CG(Claude + GLM) 모드는 리더가 **Claude API**를, 워커가 **GLM API**를 사용하는
하이브리드 모드입니다. tmux 세션 수준의 환경 변수 격리로 구현되며, "계획은
Claude가 깊게, 구현은 GLM이 싸게"라는 토크노믹스 배분을 한 세션 안에서
실행합니다. 구현 중심 작업 기준 약 60-70% 비용이 절감됩니다.

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
    ├── 3. settings.local.json에 teammateMode: "tmux" 설정
    │      → 워커들은 새 pane에서 GLM 환경변수 상속
    │
    └── 4. Claude Code 실행 (현재 프로세스 대체)
```

```
┌─────────────────────────────────────────────────────────────┐
│  리더 (현재 tmux pane, Claude API)                           │
│  - 워크플로우 오케스트레이션                                  │
│  - plan, quality, sync 단계 처리                             │
│  - GLM 환경변수 없음 → Claude API 사용                       │
└──────────────────────┬──────────────────────────────────────┘
                       │ 팀원 spawn (새 tmux pane)
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

### 4단계: 워크플로우 실행

```bash
/moai "사용자 인증 기능 구현"
```

이후는 평소와 같습니다. 오케스트레이터(리더, Claude)가 계획·품질·동기화를
맡고, 구현 물량이 큰 작업은 새 tmux pane의 GLM 팀원에게 위임됩니다.

> **참고**: 과거의 `--team` 플래그(Agent Teams 정적 오케스트레이션 계층)는
> v3.0에서 은퇴했습니다. 강제로 지정해도 sub-agent 모드로 폴백됩니다. CG
> 모드의 리더/워커 분리는 Claude Code 내장 teammate 런타임(tmux pane)으로
> 동작하며, 이 런타임은 그대로 유지됩니다.

## 중요 사항

| 항목 | 설명 |
|------|------|
| **tmux 환경** | 이미 tmux를 사용 중이면 새 세션 불필요. VS Code 터미널 기본값을 tmux로 설정하면 편리 |
| **자동 실행** | `moai cg`가 현재 pane에서 Claude Code를 자동 실행. 별도 `claude` 명령 불필요 |
| **세션 종료** | session_end 훅이 자동으로 tmux 세션 환경변수 정리 → 다음 세션은 Claude 사용 |
| **팀 통신** | SendMessage 도구로 리더↔워커 간 통신 |
| **모드 전환** | `moai glm`에서 전환 시 `moai cg`가 GLM 설정을 자동 초기화 — 중간에 `moai cc` 불필요 |

## tmux 환경 변수 주입 보안 모델 {#tmux-env-security}

v3.0.0 부터 `moai cg` 가 GLM token (`ANTHROPIC_AUTH_TOKEN`) 을 tmux 세션 환경 변수에 주입할 때, **argv 채널** (`tmux set-environment <KEY> <VALUE>`) 대신 **source-file 채널** (`tmux source-file <tmp>`) 을 사용합니다. token 은 더 이상 `ps auxe`, `/proc/<pid>/cmdline`, auditd 로그, sysmon 추적, 크래시 덤프에 평문으로 노출되지 않습니다 (CWE-214).

### 주입 흐름

1. `~/.moai/run/` 아래 임시 파일을 `mkstemp` 로 생성 (mode `0o600` 강제)
2. `set-environment -t <session> <KEY> <VALUE>` 한 줄을 기록
3. `tmux source-file <tmp>` 로 tmux 가 그 파일을 읽어 환경에 주입
4. 주입 직후 `os.Remove` 로 unlink

argv 에는 임시 파일 경로만 노출되며 token 자체는 노출되지 않습니다.

### Non-sensitive 값은 argv 유지

`CLAUDE_CONFIG_DIR`, `ANTHROPIC_BASE_URL`, `ANTHROPIC_DEFAULT_*_MODEL` 등 token 이 아닌 값은 기존 argv 경로를 유지합니다 (보안 위협 없음).

### 사용자 책임

`~/.moai/.env.glm` source 파일은 사용자 환경에서 `0o600` 권한을 유지해야 합니다. 이는 `moai glm` 명령이 자동으로 설정합니다:

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

### 자체 점검

CG 모드 실행 중 token 이 argv 에 노출되는지 확인:

```bash
# moai cg 실행 후 새 tmux 세션 내에서
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 기대값: 0 matches (token 이 argv 에 없음)
```

자세한 위협 모델, 실패 시 동작 (`ErrTmuxSensitiveInjectFailed` sentinel), 추가 점검 절차는 [보안 노트 — CWE-214](/ko/advanced/security-notes/#cwe-214) 를 참조하세요.

## 디스플레이 모드 (teammateMode)

`teammateMode`는 Claude Code 내장 디스플레이 설정으로, `settings.local.json`에
저장됩니다. MoAI의 team-mode(과거 `--team` 플래그, v3.0 은퇴)와는 다른
개념입니다 — teammate 런타임 자체는 Claude Code가 제공하며, `teammateMode`는
그 표시 방식만 제어합니다.

| 값 | 설명 | 리더/워커 분리 | CG 모드 |
|------|------|--------------|---------|
| `in-process` | 기본값, 같은 터미널 인라인 | 불가 | 미사용 |
| `auto` | 환경 자동 감지 | 미지원 | 미사용 |
| `tmux` | tmux 분할 화면 | 세션 환경변수 격리 | {{< icon check ok >}} 사용 |
| `iterm2` | iTerm2 분할 화면 | 미지원 | 미사용 |

`moai cg`와 `moai glm`은 `settings.local.json`의 `teammateMode`를 `"tmux"`로
설정하고, `moai cc`는 빈 값으로 해제합니다. 과거의 `CLAUDE_CODE_TEAMMATE_DISPLAY`
환경변수는 `teammateMode` 설정이 우선합니다.

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
- [자주 묻는 질문](/ko/getting-started/faq) — 실행 모드 관련 FAQ
- [CLI 레퍼런스](/ko/getting-started/cli) — moai cc, moai glm, moai cg 상세
