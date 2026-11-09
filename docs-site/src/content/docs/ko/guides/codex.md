---
title: Codex
description: 설치나 모델 실행 없이 Codex 지침 발견 상태를 진단하고 명시적인 프로젝트 대상에 moai 스킬 projection을 미리 봅니다.
---

Rust `moai-cli`는 읽기 전용 Codex doctor와 공통 `moai` 스킬 복사를 위한 plan/apply 인터페이스를 제공합니다. 파일 검증, Codex의 발견, 실제 실행은 각각 별개의 결과입니다.

## 준비 사항과 지원 범위

[시작하기](/ko/guides/getting-started/)를 마치고 `MOAI_CLI`·`MOAI_SOURCE`를 절대 경로로 유지하세요. 실제 세션을 열려면 Codex를 별도로 준비해야 합니다. 아래 미리보기는 Codex를 설치하거나 인증하지 않습니다.

확인할 프로젝트 루트에서 실행합니다.

```sh
export MOAI_PROJECT="$PWD"
"$MOAI_CLI" codex-doctor --help
"$MOAI_CLI" codex-projection-plan --help
"$MOAI_CLI" codex --help
```

RC 대상은 Darwin/arm64입니다. 저장소에 `codex-cli 0.163.0` 관측이 기록되어 있지만 버전 확인은 인증이 아닙니다. Desktop·CLI·IDE의 `$moai` 발견·호출은 **`not_run`**입니다. Linux, Windows, 다른 아키텍처, cloud/worktree 호스트 E2E도 로컬 fixture로 인증되지 않습니다.

## 지침 로딩 진단

```sh
"$MOAI_CLI" codex-doctor "$MOAI_PROJECT" --cwd "$MOAI_PROJECT"
```

Codex 실행, 모델 호출, 설정 쓰기 없이 파일을 읽습니다. `result` 아래에서 다음 항목을 확인하세요.

| 필드 | 확인할 내용 |
| --- | --- |
| `agents.global` | Codex 홈에서 발견한 전역 AGENTS 파일. |
| `agents.project_chain`, `agents.included` | 루트에서 작업 디렉터리까지 선택한 지침. |
| `agents.over_budget` | 예측한 지침 크기가 지정·기본 예산을 넘는지 여부. |
| `skills.moai` | 발견한 프로젝트·사용자·관리자 복사본과 중복 경고. |
| `trust` | `not_evaluated`는 `trusted`가 아닙니다. |
| `codex_layer` | 프로젝트 config·hook의 존재이며 실제 로딩 증명이 아닙니다. |

Codex 문서는 프로젝트 디렉터리에서 `AGENTS.override.md` → `AGENTS.md` → 설정한 fallback 이름 순으로 확인하고, 더 가까운 지침을 뒤에 둔다고 설명합니다. 합산 프로젝트 문서 예산의 기본값은 32 KiB입니다. doctor의 `--fallback NAME`(반복 가능)·`--max-bytes N`으로 알고 있는 설정을 반영할 수 있지만 **doctor는 `config.toml`을 읽지 않습니다.** 인자를 생략하면 실제 설정과 가정이 다를 수 있습니다.

untrusted 프로젝트에서는 Codex가 프로젝트 `.codex/`의 설정·hook·rules를 건너뛸 수 있습니다. doctor는 trust를 판정하거나 바꾸지 않습니다. 진단을 해소할 목적으로만 신뢰를 부여하지 마세요. 지침 파일의 존재는 샌드박스도 권한 부여도 아닙니다.

## 프로젝트 스킬 projection 미리보기

소스 패키지 밖에 있으며 마지막 이름이 `moai`인 절대 경로를 명시합니다. 아래는 **프로젝트** 경로를 제안할 뿐 생성하지 않습니다.

```sh
export MOAI_PROJECTION="$MOAI_PROJECT/.agents/skills/moai"
"$MOAI_CLI" codex-projection-plan "$MOAI_SOURCE/skills/moai" \
  --target "$MOAI_PROJECTION" \
  --allow-implicit-invocation false
```

`--allow-implicit-invocation false`는 명시적 스킬 호출이 필요한 정책을 계획합니다. 모든 Codex 자동화를 끄거나 호스트 권한을 바꾸는 옵션은 아닙니다. 대상과 이 옵션 모두 계획에 영향을 주므로 나중에 승인된 apply를 요청한다면 그대로 유지해야 합니다.

`package_digest`, `plan_digest`, `source_manifest`, `files`, `preserved`를 확인하세요.

- 각 계획 파일에는 digest와 대상 판단 `create`·`unchanged`·`conflict`가 있습니다.
- `SKILL.md`는 공통 패키지와 바이트가 같습니다. adapter는 `agents/openai.yaml` 메타데이터와 명시적으로 요청한 호출 정책도 계획합니다.
- 소스 manifest 비교는 `verified`·`mismatch`·`missing`·`unreadable`을 보고할 수 있습니다. 문제가 보고된 계획은 설치 승인이 아닙니다.
- 소유하지 않은 대상 파일은 보존합니다. 충돌은 기존 파일을 덮어쓰거나 삭제하라는 뜻이 아닙니다.
- 계획은 파일이나 hook을 생성하지 않습니다. 계획 성공으로 `host_loaded: false`·`execution_verified: false`가 참이 되지 않습니다.

CLI에는 의도적으로 기본 `--target`이 없습니다. 홈·프로젝트 배치를 몰래 결정하지 않습니다. 이 예제는 `~/.agents/skills`를 바꾸거나 저장소의 미결 배치 정책을 확정하지 않습니다. 홈과 프로젝트에 같은 이름의 스킬이 있으면 선택기에 둘 다 나타날 수 있으며 doctor는 병합·선택하지 않고 경고합니다.

`codex-projection-apply`는 별도 작업입니다. 같은 패키지·대상·메타데이터/정책 옵션과 검토한 `--plan-digest`, 승인된 `--permission-reference`가 필요합니다. 재계획한 뒤 검증되지 않은 소스 manifest나 대상 충돌을 거부하고, 없는 파일만 배타적으로 생성합니다. 일부 생성 후 실패하면 partial이 될 수 있으므로 전체 원자적 설치가 아닙니다. 결과의 `verified`는 설치 바이트 비교이며 **실제 호스트 로딩이 아닙니다.** 이 안내서는 apply나 기존 복사본 제거를 실행하지 않습니다. 적용을 고려하기 전에 [안전과 권한](/ko/reference/safety/)을 읽으세요.

## 론처 미리보기

```sh
"$MOAI_CLI" codex --dry-run
```

`argv`, `notes`, `executed`, `host_loaded`를 읽으세요. 론처는 Codex의 `workspace-write` 샌드박스 모드와 `on-request` 승인을 사용하는 일반 세션을 조합합니다. `--dry-run`에서는 프로세스를 시작하거나 프롬프트를 보내지 않습니다. 명령 조합은 인증·모델 사용 가능 여부·샌드박스 효력을 실측한 결과가 아닙니다.

`--dry-run`을 빼면 현재 디렉터리에서 실제 Codex CLI가 시작됩니다. 호스트는 허용된 작업 공간에 쓰고 기존 설정·hook을 사용하며 맥락을 외부로 보내고 비용을 발생시킬 수 있습니다. 대상, 기존 설정, 승인을 검토한 뒤에만 시작하세요. 론처는 스킬을 설치하거나 프로젝트를 trusted로 만들지 않습니다. Claude 전용 advisor·effort 제어는 Codex에 전달하지 않습니다.

새 세션에는 deprecated된 MoAI 론처 역할 `-l`·`-a`를 쓰지 마세요. 조합된 **Codex** argv의 `-a on-request`는 Codex의 승인 정책이지 deprecated된 MoAI 역할 플래그가 아닙니다.

## 발견된 스킬 사용

검토된 스킬이 실제로 발견되면 Codex CLI에서 `$`를 입력하거나 `/skills`로 선택합니다. `$moai`가 셸 변수로 확장되는 터미널이 아니라 **Codex 안에서** 읽기 전용 요청을 입력하세요.

```text
$moai 이 저장소의 현재 상태만 요약해 줘. 파일 변경, Issue claim, 외부 게시는 하지 마.
```

이 요청은 지침이지 강제 격리가 아닙니다. 승인된 워크플로 참여 시 현행 역할 문법은 `$moai --mode agent`입니다. agent 모드는 self-claim을 시작하며 GitHub claim을 게시할 수 있으므로 읽기 전용 발견 시험으로 쓰지 마세요. 이 RC의 leader 모드는 Claude Code 전용이므로 Codex로 진입하면 안 됩니다. Codex 세션에는 RC 세션 간 `ListAgents`·`SendMessage` 채널이 없습니다. 영속적인 Issue 댓글과 Markdown 인계는 별도 승인 대상입니다.

## 문제 해결과 다음 단계

- **스킬 중복:** doctor가 보고한 경로를 확인하세요. 소유권·범위 결정 없이 기존 복사본을 삭제하지 마세요.
- **manifest·대상 충돌:** 검토한 소스·대상 불일치를 해결하고 새 계획을 만드세요. 검사를 약화하거나 승인 메타데이터를 지어내지 마세요.
- **예상과 다른 지침:** 실제 cwd, override 파일, fallback 이름, 바이트 예산을 doctor 입력과 비교하세요.
- **세션에 설정 미반영:** 프로젝트 trust와 관련될 수 있지만 doctor는 판단할 수 없습니다. 이 CLI에는 Codex config plan/apply 하위 명령이 없습니다. 프로젝트의 `codex_config` 구현은 라이브러리 표면이므로 CLI 명령을 지어내면 안 됩니다.
- **파일은 있지만 `$moai` 없음:** 실제 호스트의 발견 상태와 재시작 안내를 확인하세요. projection 성공은 실제 호출 receipt가 아닙니다.

[CLI 참조](/ko/reference/cli/), [안전 참조](/ko/reference/safety/), [Claude Code 안내서](/ko/guides/claude-code/)로 이어집니다.

## 출처

이 안내서는 고정 revision의 [CLI 선언](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/src/main.rs), [론처](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/src/session.rs), [Codex adapter 계약](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/docs/codex-adapter-contract.md), [모드 워크플로](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/skills/moai/workflows/mode-agent.md), [지원표](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/docs/support-matrix.md)를 사용합니다.

계약에는 2026-11-08/09에 읽은 공식 [스킬](https://learn.chatgpt.com/docs/build-skills)·[AGENTS](https://learn.chatgpt.com/docs/agent-configuration/agents-md)·[설정](https://learn.chatgpt.com/docs/config-file/config-reference) 문서가 기록되어 있습니다. 문서화된 동작을 이 사이트에서 새로 실측한 것으로 해석하면 안 됩니다.
