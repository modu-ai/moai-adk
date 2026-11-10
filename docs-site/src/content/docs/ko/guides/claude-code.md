---
title: Claude Code
description: 실제 세션을 시작하지 않고 Claude Code의 지침 로딩을 진단하고 프로젝트 설정과 론처 명령을 미리 봅니다.
---

Rust `moai-cli`가 Claude Code를 위해 준비할 수 있는 작업을 확인합니다. 파일을 계획하거나 지침 로딩을 예측한 결과는 실제 세션이 `moai` 스킬을 읽었다는 증거가 아닙니다.

## 준비 사항과 지원 범위

[시작하기](/ko/guides/getting-started/)를 마치고 절대 경로 변수 `MOAI_CLI`·`MOAI_SOURCE`를 유지하세요. 진단에 사용할 Claude Code는 해당 제품의 공식 절차로 별도 설치합니다. `claude-doctor`는 버전을 확인할 수 있지만 설치나 인증을 대신하지 않습니다.

확인할 프로젝트에서 대상을 명시합니다.

```sh
export MOAI_PROJECT="$PWD"
"$MOAI_CLI" claude-doctor --help
"$MOAI_CLI" cc --help
```

RC 대상은 Darwin/arm64입니다. 저장소에는 Claude Code **2.1.319** 관측과 스킬 정적 검증이 기록되어 있지만 실제 `/moai` 발견·호출·권한 거부 E2E는 **`not_run`**이며 인증된 지원이 아닙니다. Linux, Windows, 다른 아키텍처, cloud 세션도 이 근거로 검증되지 않습니다.

## 지침 로딩 진단

```sh
"$MOAI_CLI" claude-doctor "$MOAI_PROJECT"
```

doctor는 지침 파일과 관련 설정을 읽으며 `claude --version`을 실행할 수 있습니다. 모델 세션을 시작하거나 import를 만들거나 `CLAUDE.md`를 삭제하거나 설정을 바꾸지는 않습니다. `result.agents_md_status`, `reason`, `load_path`, `basis`, `suggestions`를 확인하세요.

- AGENTS 직접 로딩에는 Claude Code **2.1.277 이상**이 필요합니다. 버전을 알 수 없으면 불확실한 상태이지 지원 증명이 아닙니다.
- 기본 지침 파일 정책에서는 확인 범위의 `CLAUDE.md`, `.claude/CLAUDE.md`, `CLAUDE.local.md`가 `AGENTS.md` 직접 로딩을 억제할 수 있습니다.
- 기존 `@AGENTS.md` import가 있다면 import 경로로 읽을 수 있습니다. doctor는 대안을 제안할 뿐 적용하지 않습니다.
- `basis: predicted`에 주의하세요. managed 정책과 별도의 `--settings` 파일은 확인하지 않습니다. 상태 이름이 `loaded`여도 실제 세션 receipt가 아닌 예측입니다.

기존 지침을 보존하세요. 진단 상태를 바꾸려고 충돌하는 파일을 삭제하지 말고 프로젝트 소유자와 import·정책 변경안을 검토하세요.

## 프로젝트 설정 미리보기

Rust 설정 인터페이스는 **프로젝트** `.claude/settings.json`을 대상으로 계획합니다. 예를 들어 Claude 내장 Default 출력 스타일 선택을 미리 볼 수 있습니다.

```sh
printf '%s\n' '{"outputStyle":"Default"}' |
  "$MOAI_CLI" claude-settings-plan "$MOAI_PROJECT"
```

이 명령은 아무것도 쓰지 않습니다. 계획의 action, 충돌, 예상·현재 digest, `plan_digest`를 읽으세요. 다른 사용자 소유 값이 있으면 충돌할 수 있습니다. 소유권 기록을 삭제해 우회하지 마세요. HOME 자체나 `.claude`가 `CLAUDE_CONFIG_DIR`인 대상은 계획 단계에서 거부합니다.

`claude-settings-apply`는 별도의 변경 작업입니다. 같은 desired JSON과 `--plan-digest`, `--permission-reference`를 받아 재계획한 뒤 프로젝트 설정과 `.moai-cli/manifests/claude-settings.json`을 씁니다. 이 안내서는 실행하지 않습니다. reference는 권한 승인이 아니므로 정확한 diff를 먼저 승인해야 합니다. 다른 사용자 소유 키는 보존해야 합니다.

설정 병합은 **자산 설치가 아닙니다.** Rust 대체 경로는 임시 Python adapter의 출력 스타일 파일, `statusline.py`, 라이선스를 복사하지 않습니다. 없는 스크립트를 `statusLine.command`에 지정하거나 설정에 이름이 있다고 스타일 파일도 있다고 가정하지 마세요. Rust 바이너리는 그 Python adapter를 호출하지 않습니다. 기존 hook·권한도 별도 검토 대상이며, 이 안내서들이 기준으로 삼는 고정 revision은 MoAI hook 세트를 생성하지 않습니다.

## 론처 미리보기

론처는 저장소 개발용 편의 도구이지 스킬 설치 프로그램이 아닙니다. MoAI 역할과 실험적 advisor 없이 일반 세션 명령을 확인합니다.

```sh
"$MOAI_CLI" cc --no-advisor --permissions --dry-run
```

`argv`, `notes`, `executed`, `host_loaded`를 확인하세요. `--dry-run`은 JSON만 조합하므로 Claude Code 시작, 프롬프트 전송, 자격증명 검증을 하지 않습니다. 역할 없는 `cc`는 권한 확인 생략 플래그를 추가하지 않습니다. `--permissions`도 론처가 해당 플래그를 추가하지 않도록 명시하지만 Claude Code에 이미 설정된 느슨한 권한을 덮어쓰지는 **않습니다**.

Claude 론처 기본값은 모델 `opus`, effort `medium`, advisor `fable`입니다. `--no-advisor`는 소스가 실험적·Anthropic API 전용으로 표시한 advisor 옵션을 제거합니다. 검토한 명령에서 `--dry-run`만 빼면 현재 디렉터리에서 실제 호스트가 시작됩니다. 호스트 인증, hook, 외부 전송, 비용을 의도적으로 수락할 때만 실행하세요. 기존 project·local·user·managed 권한 설정을 먼저 검토해야 합니다.

새 세션에는 deprecated된 MoAI 론처 역할 `-l`·`-a`를 쓰지 마세요. 이전 동작이 남아 있으므로 일반 세션을 열고 그 안에서 스킬 모드를 선택합니다.

## 발견된 스킬 사용

`moai-cli` 빌드나 `cc` 실행만으로 `/moai`가 나타나지 않습니다. Claude Code가 스킬 경로에서 검토된 `moai`를 먼저 발견해야 합니다. 이 안내서는 `~/.claude/skills`나 프로젝트 `.claude/skills`에 복사하지 않습니다. 홈·프로젝트 설치와 실제 로딩은 별도 승인과 확인이 필요한 단계입니다.

검토된 스킬이 이미 보인다면 셸이 아니라 **Claude Code 안에서** 읽기 전용 요청을 입력합니다.

```text
/moai 이 저장소의 현재 상태만 요약해 줘. 파일 변경, 구현 시작, 외부 게시는 하지 마.
```

이 요청은 지침이지 샌드박스가 아닙니다. 승인된 워크플로에 참여할 때의 현행 문법은 `/moai --mode leader`(기본 autonomy `semi`) 또는 `/moai --mode agent`입니다. agent 모드는 self-claim을 시작하며 GitHub claim을 게시할 수 있으므로 읽기 전용 시험이 아닙니다. 이 RC의 leader 모드는 Claude 전용이고 독립 감사나 사용자 결정을 생략하지 않습니다. 어느 모드도 저장소 텍스트에서 권한을 얻지 않습니다.

## 문제 해결과 다음 단계

- **AGENTS 억제:** doctor의 이유를 읽고 기존 CLAUDE 지침을 보존하세요. 제안된 해결책은 별도 승인합니다.
- **스킬 없음:** 호스트의 실제 스킬 발견 상태를 확인하고 해당 제품의 reload·재시작 안내를 따르세요. 바이너리 존재나 설정 성공은 스킬 로딩 증거가 아닙니다.
- **설정 충돌·오래된 계획:** 달라진 바이트를 확인하고 재계획하세요. permission reference를 지어내거나 이전 digest를 재사용하지 마세요.
- **호스트 미실행:** `--dry-run` 성공은 명령 조합만 확인하며 호스트 설치, 인증, 모델 사용 가능 여부, E2E를 검증하지 않습니다.

다른 adapter는 [Codex 안내서](/ko/guides/codex/)를, 작업 경계는 [CLI 참조](/ko/reference/cli/)와 [안전과 권한](/ko/reference/safety/)을 확인하세요.

## 출처

명령과 동작의 기준은 고정 revision의 [CLI 선언](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/src/main.rs), [론처 구현](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/src/session.rs), [Claude adapter 계약](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/docs/claude-adapter-contract.md), [지원표](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/docs/support-matrix.md)입니다. 현행 모드 진입은 [공통 스킬](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/skills/moai/SKILL.md)에 정의되어 있습니다.

저장소 계약은 2026-11-08에 읽은 공식 [memory](https://code.claude.com/docs/en/memory)·[skills](https://code.claude.com/docs/en/skills) 문서를 기록합니다. 문서화된 기능을 이 사이트에서 새로 실측한 것으로 해석하면 안 됩니다.
