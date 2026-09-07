---
title: moai doctor 진단
weight: 60
draft: false
---

`moai doctor` 는 시스템 전반을 한 번에 진단합니다. Claude Code 설정, 의존성, 프로젝트 구조, 언어별 개발 도구, 환경을 차례로 검사하고, 문제를 찾으면 고칠 방법까지 함께 알려 줍니다.

에이전트(스스로 일하는 AI)가 세팅을 무심히 건드리면서 설정 드리프트가 생기는 일이 잦기 때문에, 한 번에 전체 상태를 보여 주는 단일 진단 커맨드가 필요합니다. 하네스(harness) 게이트와 SPEC 라이프사이클이 모두 `moai` CLI 를 거쳐 동작하므로, 어느 한 축이 깨져 있으면 다른 커맨드의 메시지만으로는 원인을 짚기 어렵습니다. 따라서 `moai doctor` 는 개별 커맨드가 실패할 때 가장 먼저 실행해 볼 1차 진단 도구입니다.

## 개요

```bash
moai doctor [OPTIONS]
```

## 플래그

| 플래그 | 설명 |
|--------|------|
| `-v, --verbose` | 상세 진단 정보 (도구 버전, 언어 감지 결과) 표시 |
| `--fix` | 감지된 문제의 수정 방법 제안 |
| `--export` | 진단 결과를 JSON 파일로 내보내기 |
| `--check <tool>` | 특정 검사만 실행 (예: git, go, config) |

## 하위 명령어

특정 영역만 깊이 들여다볼 때 쓰는 하위 명령어도 있습니다.

| 명령어 | 설명 |
|--------|------|
| `moai doctor config` | 설정 진단 — 병합된 설정을 provenance 와 함께 검사 |
| `moai doctor hook` | 27개 훅 이벤트 커버리지 표시 |
| `moai doctor permission` | 권한 해석 진단 |
| `moai doctor sandbox` | 샌드박스 백엔드 가용성 진단 |

`moai doctor config` 는 다시 `dump`(병합 설정 덤프)와 `diff <tier-a> <tier-b>`(두 설정 티어 비교) 를 제공합니다.

## Home Disk Usage 진단 {{< new-badge v3.1.1 >}}

`moai doctor` 전체 진단에는 **Home Disk Usage** 항목이 함께 나옵니다. `~/.moai` 홈 디렉터리가 얼마나 찼는지를 보고하는 **권고(advisory)** 성격의 검사라, 임계값을 넘어도 다른 명령을 막지 않습니다.

| 보고 항목 | 내용 |
|-----------|------|
| 전체 크기 | `~/.moai` 총 용량과 상위 3개 항목 |
| 프로필별 내역 | `claude-profiles/<프로필>` 각각의 크기와 범주 분해 |
| 릴리즈 개수 | `releases/`에 남아 있는 바이너리 수와 현재 버전 |
| 정리 가능량 | `moai clean --home`이 실제로 지울 수 있는 추정 바이트 |
| `~/.claude` | 크기만 보고 — 어떤 경로로도 정리 대상이 아님 |

정리 가능량이 임계값(컴파일된 기본값 500 MB)을 넘으면 상태가 WARN으로 바뀌고 `moai clean --home`(기본 dry-run)을 권합니다. 그 아래면 OK로 남습니다. `~/.moai`가 아예 없으면 "보고할 것 없음"으로 OK 처리됩니다.

이 추정치는 `moai clean --home`이 쓰는 것과 **같은 스캐너**를 호출하므로, doctor가 말하는 숫자와 clean이 실제로 지우는 목록이 어긋나지 않습니다. 자세한 내용은 [홈 디렉터리 위생](/ko/advanced/home-hygiene)에 있습니다.

## Hook Delivery 진단 {{< new-badge v3.1.4 >}}

`moai update` 는 템플릿이 새로 추가한 훅 항목을 기존 프로젝트의 `.claude/settings.json` 에 넣지 못한 채 조용히 넘어갈 수 있습니다. **Hook Delivery** 검사는 이런 누락을 찾아냅니다. 프로젝트가 이미 가지고 있는 훅 이벤트 키 안에서, 배포 템플릿에는 있는데 프로젝트 설정에는 없는 항목을 비교해 보고합니다.

| 보고 내용 | 설명 |
|-----------|------|
| 누락된 항목 | `hooks.PreToolUse missing handle-pre-tool.sh (matcher AskUserQuestion)` 형태로 이벤트 키와 매처까지 알려 줍니다 |
| 수정 방법 | 이름이 지정된 이벤트 키 아래에, 사용 중인 moai 버전의 템플릿 설정에서 해당 블록을 복사해 넣도록 안내합니다 |
| 업데이트 후 점검 | `moai update` 뒤 관리 파일이 삭제되지 않았는지 확인하는 명령(`git status --porcelain \| grep '^ D'`)을 제시합니다 — 걸리면 `git restore -- <경로>` 로 되돌린 뒤 항목을 다시 넣습니다 |

이 검사는 읽기 전용입니다 — `.claude/settings.json` 을 절대 쓰지 않습니다. 템플릿이 처음 제공하는 이벤트 키와 사용자가 직접 만든 항목은 누락으로 세지 않고, opt-out 된 항목도 요구하지 않습니다(템플릿은 프로젝트의 `hook.opt_in.enabled` 설정을 반영해 렌더링됩니다). 모든 항목이 일치하면 `ok` 를 보고합니다.

## Codex Wiring 진단 {{< new-badge v3.1.4 >}}

`moai doctor` 전체 진단에는 **Codex Wiring** 항목이 함께 나옵니다. 프로젝트의 Codex 배선(생성된 훅, MCP 등록, 스킬 미러)이 온전한지 살피는 조언(advisory) 성격의 검사로, fail-open 이며 읽기 전용입니다. 발견한 문제는 보고할 뿐, 절대 스스로 고치지 않습니다.

| 검사 항목 | 내용 |
|-----------|------|
| `.codex/hooks.json` 존재 · 키 화이트리스트 | 배선 파일이 있는지, 키가 화이트리스트에 맞는지 봅니다. 키 하나라도 어긋나면 codex 가 파일 전체를 조용히 무시하므로, 이 검사가 그 침묵을 대신 관측합니다 |
| 사이드카 해시 | 배포 때 기록해 둔 훅 내용의 해시와 현재 파일을 비교해, 훅을 직접 고친 뒤에 남는 발산을 잡습니다 |
| `moai` 바이너리 PATH | 생성된 훅 명령은 `moai hook ...` 형태라 PATH 에서 `moai` 를 찾지 못하면 하나도 발화하지 않습니다 |
| `.codex/config.toml` 의 `[mcp_servers.moai]` | MCP 등록 테이블의 존재와, 표준 등록 형태와의 일치 여부를 봅니다. 이 테이블은 사용자 소유라 doctor 는 보고만 하고 수리하지 않습니다 |
| `.agents/skills` 스킬 미러 | Codex CLI 는 `.claude/skills` 를 훑지 않으므로, 미러가 없거나 링크가 끊겨 있으면 codex 는 이 프로젝트의 MoAI 스킬을 볼 수 없습니다 |
| 사용자측 `[[skills.config]]` 등록 | `~/.codex/config.toml`(`$CODEX_HOME` 설정 시 그 파일)의 스킬 등록에서 경로와 `enabled` 키 형태를 봅니다. 이 항목은 codex 가 관여하는 경우, 즉 프로젝트가 배선돼 있거나 codex 가 설치된 경우에만 검사합니다 |

검사가 문제를 발견하면 코드에 박힌 수정 지시문이 함께 표시됩니다.

| 발견 | 지시문 |
|------|--------|
| 배선이 아예 없는 프로젝트 | `moai init --agent codex` |
| 훅 변경 뒤 사이드카 발산 | `codex /hooks` 로 변경된 훅을 재신뢰 |
| 스킬 미러 부재 · 끊긴 링크 | `moai update --templates-only --force --yes` |
| 가리키던 스킬 파일이 사라진 등록 | 항목을 제거하거나 스킬 파일을 복원 |

이 검사가 내는 발견 가운데 fatal 로 분류되는 것은 단 하나입니다. 사용자측 설정의 `[[skills.config]]` 항목에 `enabled` 키가 없거나 bare TOML 불리언이 아닌 값이 들어 있으면 codex 는 모든 호출에서 exit 1 로 끝납니다. 이 유일한 fatal 발견에 걸리면 doctor 결과가 Fail 이 되어 종료 코드 1 로 이어집니다. 다만 이 동작은 codex-cli 0.153.4 에서 관측된 것으로, 해당 릴리스에서 확인됐을 뿐 다른 릴리스까지 일반화되지 않습니다. 올바른 형태는 모든 항목에 `enabled = true` 또는 `enabled = false` 를 명시하는 것입니다. 이 하나를 제외한 모든 발견은 조언형이라 doctor 의 종료 코드를 바꾸지 않습니다.

codex 가 설치되지 않은 머신에서 배선이 없는 claude-only 프로젝트라면, 이 검사는 조용히 건너뜁니다 — 정보성 스킵이며 경고 행을 만들지 않습니다. 참고로, 더 이상 존재하지 않는 스킬 파일을 가리키는 유령(ghost) 등록을 한 번에 수거하는 기능은 이 검사의 지시문 밖에 있으며, 별도 동사인 `moai clean --codex-skills` 이 맡습니다.

훅 신뢰 모델과 스킬 미러를 포함한 Codex 배선 전반은 [Codex 듀얼 하네스](/ko/advanced/codex-dual-harness) 문서에 자세히 있습니다.

## 종료 코드

스크립트나 CI 래퍼에서 `moai doctor` 를 부를 때는 요약 줄이 아니라 종료 코드를 읽습니다.

| 종료 코드 | 의미 |
|-----------|------|
| `0` | Fail 항목 없음. Warn 은 권고라 종료 코드를 바꾸지 않습니다 |
| `1` | 한 건 이상이 Fail. 요약의 `Fail N` 이 그대로 반영됩니다 |

Constitution Registry 항목은 레지스트리가 파싱되는지만 보지 않고 `moai constitution validate` 와 **같은 드리프트 검사**를 돌립니다. 따라서 같은 체크아웃에서 doctor 가 ok 라고 하는데 validate 가 실패하는 일은 없습니다. `MOAI_CONSTITUTION_SKIP_VALIDATE=1` 로 우회하면 doctor 도 구조 검사 판정으로 돌아갑니다.

## 예시

```bash
# 전체 진단
moai doctor

# 상세 진단
moai doctor --verbose

# 진단 결과 내보내기
moai doctor --export diagnostics.json

# 특정 영역 진단
moai doctor hook          # 훅 커버리지 표
moai doctor permission    # 권한 해석
moai doctor sandbox       # 샌드박스 백엔드
```

---

관련: [프로젝트 상태](/ko/cli-reference/status) · [CLI 개요](/ko/getting-started/cli)
