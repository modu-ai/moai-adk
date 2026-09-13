---
title: CG 폐기와 설정 이전
weight: 20
draft: false
description: CG 폐기와 설정 이전
---

`moai cg`는 폐기되었습니다. Claude나 GLM을 실행하지 않고 설정 이전 안내와 함께 종료합니다. `moai cc`의 별칭이 아닙니다. `llm.team_mode: cg`가 남은 프로젝트는 세션을 실행하기 전에 이전할 구성을 명시적으로 선택해야 합니다.

## 변경 전 미리보기

프로젝트 루트에서 선택지를 먼저 확인합니다. 미리보기는 설정을 바꾸거나 백업을 만들지 않습니다.

```bash
moai migrate cg
moai migrate cg --target claude-only
```

## Claude 전용 역할로 이전

GLM 팀원 자동 배정을 없애는 역할 변화에 동의할 때만 다음 명령으로 적용합니다.

```bash
moai migrate cg --target claude-only --apply --accept-role-change
```

`llm.team_mode: claude`, `llm.gateway.teammate_mode: in-process`, `llm.gateway.teammate_provider: inherit`을 저장합니다. 기존 혼합 역할 배정을 없애는 변경이며, Claude 리더와 GLM 팀원 창의 분업을 보존하는 이전이 아닙니다.

## 혼합 구성의 검증 조건

`claude-glm` 대상은 Claude 리더와 tmux의 GLM 팀원 구성을 뜻합니다. 현재 TEAMMATE 통합 검증을 통과하지 않아 적용과 실행은 사용할 수 없으며 미리보기만 가능합니다. tmux를 설치하거나 `verified: true`를 적어도 이 제한은 해제되지 않습니다.

```bash
moai migrate cg --target claude-glm
```

## 설정 보존과 오류 처리

알 수 없는 설정 값, 주석, GLM 모델 설정, 자격증명 참조는 보존합니다. 적용 전 원본 바이트를 `.moai/backups/cg-migration/<source-sha256>.yaml`에 그대로 백업합니다. YAML의 서식은 달라질 수 있습니다. 같은 대상으로 다시 실행하면 변경 없음으로 끝나며, 반대 대상으로 재이전하면 오류가 납니다.

충돌하는 gateway 값, 비어 있지 않은 `llm.mode`, 중복 키, 지원하지 않는 YAML 별칭은 명시적으로 해결해야 합니다. 미리보기와 사전 검사 실패는 원본과 백업 디렉터리를 바꾸지 않습니다. 쓰기 도중 실패하면 백업이 남을 수 있으므로 오류와 보존된 원본을 확인한 뒤 다시 진행합니다.

## 자격증명 취급 {#tmux-env-security}

과거 CG 환경 변수 주입 안내는 현재 사용할 수 있는 런처의 동작이 아닙니다. 설정 이전은 provider를 실행하거나 자격증명을 옮기지 않습니다. 백업에는 원래 설정이 담기므로 접근 권한을 제한해야 합니다.

## 다음 단계

역할 변화에 동의하여 `claude-only`로 이전한 뒤에는 지원 런처를 명시적으로 선택합니다. `moai cc`와 `moai glm`은 기존 혼합 역할을 재현하지 않습니다. GPT gateway 실행은 별도 통합 검증 대상이며 이 이전으로 활성화되지 않습니다.

- [CLI](/ko/cli-reference/launchers/)
