---
title: CLI 핵심 참조
description: Rust moai-cli의 조회, 프로젝트 프로필, 패키지, 관리 지침 명령 참조.
---

이 문서는 Rust **moai-cli** RC의 핵심 명령만 다룹니다. Go **moai**와 다른 실행 파일이므로 서로 바꾸어 실행하지 마세요. 내부 JSON 명령의 존재가 공개 설치 프로그램이나 안정된 공개 명령 체계를 뜻하지는 않습니다. 먼저 [시작하기](/ko/guides/getting-started/)와 [안전 및 복구](/ko/reference/safety/)를 읽으세요.

## 변경 없이 확인하기

직접 빌드했거나 출처를 확인한 Rust 실행 파일로 실행합니다.

~~~sh
moai-cli --version
moai-cli --help
moai-cli project-apply --help
moai-cli identity .
moai-cli install-inspect --json
moai-cli baseline-inventory .
~~~

도움말과 버전은 일반 CLI 출력이며, 나머지는 JSON envelope를 반환합니다. identity는 쓰기 없이 clone/workspace 식별자와 상태 경로를 보여 줍니다. 연결 worktree들은 clone ID를 공유하지만 workspace ID는 다릅니다. 현재 ID는 정규화된 경로에서 파생되므로 저장소를 옮기면 달라질 수 있습니다. 대상이 예상과 다르거나 git 조회가 실패하면 중단하세요.

install-inspect는 관측한 설치 상태를 조회할 뿐 설치하지 않습니다. baseline-inventory는 읽기 전용이며, 기록된 Go 기준점을 조사할 때 선택적으로 `--go-baseline-commit SHA`와 `--go-repo REPO`를 받습니다. 어느 명령도 마이그레이션이나 Go 삭제를 승인하지 않습니다.

## 명령 표기

아래는 **문법 참조**이지 승인까지 포함한 실행 예제가 아닙니다. PROJECT는 기존 대상 디렉터리, PACKAGE는 명시적으로 지정한 로컬 스킬 패키지 디렉터리입니다. DIGEST는 검토한 계획의 `plan_digest`입니다. REFERENCE는 해당 작업과 범위에 대해 호출 호스트가 실제로 기록한 승인 참조여야 합니다. 임의로 만들지 마세요.

~~~text
moai-cli project-plan PROJECT
moai-cli project-apply PROJECT --plan-digest DIGEST --permission-reference REFERENCE
moai-cli package-inspect PACKAGE
moai-cli package-resolve PROJECT [--repo-package PACKAGE] [--user-package PACKAGE]
moai-cli agents-plan PROJECT
moai-cli agents-apply PROJECT --plan-digest DIGEST --permission-reference REFERENCE
moai-cli apply-recover PROJECT --receipt ID --permission-reference REFERENCE
~~~

## 프로젝트 프로필

`project-plan`은 stdin으로 타입이 정해진 Profile JSON을 받아 쓰기 없이 계획을 반환합니다. 새 계획 입력은 `schema_version: 2`여야 합니다. 필수 필드는 `schema_version`, `project_id`, `purpose`, `tools`, `development_methods`, `hosts`이며, hosts 값은 `claude` 또는 `codex`입니다. `package_pin`, `interview_revision`, 구조화된 `decisions`는 선택 사항입니다. 알 수 없는 필드와 잘못된 값은 거부됩니다. 자격 증명이나 인터뷰 원문을 넣지 마세요.

반환된 `action`, 키별 `delta`, 대상 바인딩, 예상 해시와 digest를 검토합니다. action은 `create`, `update`, `unchanged`, `conflict`입니다. apply의 stdin에는 **성공 envelope의 result에 담긴 정확한 계획 객체**를 전달합니다. envelope 전체나 수정한 프로필을 전달하는 것이 아닙니다. 적용 대상은 `.moai-cli/project.yaml`과 프로필 기준선·영수증 기록입니다. 프로필 쓰기 성공만으로 호스트 로딩이 입증되지는 않습니다.

## 패키지

`package-inspect`는 SKILL.md와 지원 파일 전체에 대해 `moai-package-tree-v1` digest를 계산합니다. 이는 콘텐츠 식별이지 게시자 신뢰나 실행 안전성 검증이 아닙니다.

`package-resolve`는 읽기 전용입니다. 프로필에 package pin이 있으면 명시적으로 전달한 저장소 패키지가 정확히 일치해야 하며, 없거나 다르면 다른 패키지로 대체하지 않고 실패합니다. pin이 없으면 명시적으로 전달한 사용자 패키지만 선택합니다. 두 플래그는 구문상 선택 사항이지만 해당 분기에서 필요한 패키지는 반드시 제공해야 합니다. 홈 디렉터리를 암묵적으로 검색하거나 호스트를 활성화하지 않습니다.

## 관리 지침

`agents-plan`의 stdin 요청 형태는 다음과 같습니다.

~~~json
{
  "protocol_version": 1,
  "language": "ko",
  "workflow_reference": ".moai-cli/workflows",
  "local_file": "none"
}
~~~

이 예제는 요청 형태를 설명하며 디렉터리를 생성하지 않습니다. 저장소 상대 경로로 지정한 워크플로 디렉터리가 이미 존재해야 하고 프로젝트 프로필도 필요합니다. 워크플로가 아직 생성되지 않았다면 generation 절차를 먼저 사용하세요. 검사를 통과하려고 빈 디렉터리를 만들지 마세요. 영어 출력은 `language`를 `en`으로 설정합니다. `local_file`은 `none`, `ignored`, `tracked` 중 하나이며 개인 지침 파일 생성을 선택합니다.

모든 파일 결정과 경고를 확인하세요. `conflict`가 하나라도 있으면 적용할 수 없습니다. 검토한 성공 계획의 `result`를 stdin으로 전달하고 그 digest와 실제 승인 참조를 지정합니다. 이 명령군은 AGENTS.md의 관리 구역, 선택적인 AGENTS.local.md 최초 생성, 관련 .gitignore 구역과 generation manifest를 다룹니다. 임의의 사용자 파일을 관리하는 명령이 아닙니다. 호스트 설정은 [Claude Code](/ko/guides/claude-code/)와 [Codex](/ko/guides/codex/)를 참고하세요.

## 결과와 실패 처리

내부 operation은 stdout에 JSON envelope 하나를 반환합니다. 성공은 `{protocol_version, operation, ok, result}`, 실패는 `{protocol_version, operation, ok, error}` 형태입니다. `result`를 사용하기 전에 `ok`와 종료 코드를 확인하세요. 종료 코드는 성공 0, operation 실패 1, 명령행 사용 오류 2입니다. 도움말·버전·런처·알 수 없는 명령은 envelope 계약의 예외입니다.

`error.kind`와 제공되는 `next_actions`를 확인하세요. 오래된 계획은 digest를 강제로 바꾸지 말고 다시 계획하고 검토합니다. 부분 적용은 영수증부터 확인합니다. 이 페이지는 forge 쓰기, 릴리스·마이그레이션 작업과 전체 내부 명령 목록을 다루지 않습니다.
