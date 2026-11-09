---
title: 시작하기
description: Rust moai-cli를 소스에서 빌드하고 홈 설치나 AI 호스트 실행 없이 프로젝트 계획을 확인합니다.
---

Rust MoAI의 명령은 기존 Go 명령 `moai`가 아니라 `moai-cli`입니다. 이 안내서는 릴리스 후보를 소스에서 빌드하고 읽기 전용 미리보기를 실행합니다. 바이너리를 빌드해도 Claude Code나 Codex에 `moai` 스킬이 설치되지는 않습니다.

## 준비 사항과 지원 범위

- Darwin/arm64 Mac, Git, rustup으로 관리하는 기존 Rust 설치가 필요합니다. 저장소의 `rust-toolchain.toml`은 Rust **1.94.1**을 지정합니다. 처음 사용할 때 rustup이 해당 도구 체인과 구성 요소를 내려받을 수 있습니다.
- 복제에 필요한 네트워크와, 캐시에 없다면 Rust 구성 요소·Cargo 의존성을 내려받을 네트워크가 필요합니다. `--locked`는 의존성 해석을 고정할 뿐 오프라인 실행을 뜻하지 않습니다.
- 터미널과 새 체크아웃 경로를 준비합니다. 기존 Go 설치 위에 빌드하거나 그 자격증명을 복사하지 마세요.

패키지 버전은 `2026.11.0-rc.1`이며 `publish = false`입니다. 이 문서는 공개 Rust 설치 프로그램이나 crates.io 설치 경로를 확인하지 않았으므로 `cargo install moai-cli`나 설치 스크립트를 안내하지 않습니다.

**지원 대상과 인증은 다릅니다.** Darwin/arm64의 Claude Code·Codex가 RC 대상입니다. 소스의 지원표에는 결정적 fixture 결과 5건과 Claude 정적 검증 결과 1건이 있으며, 실제 호스트 로딩을 인증한 결과가 아닙니다. `/moai`·`$moai`의 실제 로딩·호출은 `not_run`입니다. Darwin/x86_64, Linux, Windows, ZCode, cloud/worktree 호스트 E2E도 이 결과로 인증되지 않습니다.

## 분리된 소스 체크아웃 빌드

아래 명령은 새 체크아웃과 빌드 디렉터리를 만듭니다. 홈에 바이너리를 설치하거나 PATH를 바꾸거나 호스트를 시작하지 않습니다. detached revision은 이 안내서의 기준 소스입니다.

```sh
git clone --branch v2026.11.0 https://github.com/modu-ai/moai-adk.git moai-adk-rust
cd moai-adk-rust
git checkout --detach cf33967c42c31f020fd0e35dc07b6b82093119f1
CARGO_TARGET_DIR="$PWD/target-docs" cargo build --locked --release
export MOAI_SOURCE="$PWD"
export MOAI_CLI="$MOAI_SOURCE/target-docs/release/moai-cli"
"$MOAI_CLI" --version
"$MOAI_CLI" --help
```

Cargo/rustup은 체크아웃 밖의 일반 캐시에 쓸 수 있습니다. 소스 빌드이지 완전한 샌드박스 빌드라는 보장은 아닙니다. 지정한 도구 체인이나 의존성을 구할 수 없다면 확인되지 않은 설치 프로그램으로 대체하지 말고 해당 준비 사항을 해결하세요. 오프라인 빌드는 캐시가 이미 채워져 있어야 하며, 같은 대상 디렉터리를 지정하고 `cargo build --offline --locked --release`를 사용합니다.

`MOAI_CLI`는 절대 경로로 유지합니다. 이름을 `moai`로 바꾸거나 Go 바이너리를 덮어쓰거나 Go의 `moai init`·`moai update`를 Rust 설정 단계로 실행하지 마세요. 위 변수는 현재 셸에만 적용되므로 셸을 다시 열었다면 다시 설정해야 합니다.

## 적용하지 않고 확인

먼저 체크아웃과 공통 스킬 패키지를 확인합니다.

```sh
"$MOAI_CLI" identity "$MOAI_SOURCE"
"$MOAI_CLI" package-inspect "$MOAI_SOURCE/skills/moai"
```

`identity`는 Git 작업 공간 식별자와 상태 경로를 보고할 뿐 생성하지 않습니다. `package-inspect`는 스킬 트리 전체를 읽어 콘텐츠 digest를 계산합니다. digest는 관측한 콘텐츠의 동일성을 나타내며 **배포자 신뢰, 호스트 발견, 실행 권한**의 증명이 아닙니다.

실제 프로젝트를 바꾸지 않고 첫 프로필 계획을 확인하려면 임시 디렉터리를 만들고 완전한 v2 요청을 전달합니다.

```sh
export MOAI_PROJECT="$(mktemp -d)"
"$MOAI_CLI" project-plan "$MOAI_PROJECT" <<'JSON'
{
  "schema_version": 2,
  "project_id": "docs-first-project",
  "purpose": "Evaluate a local Rust workflow",
  "tools": ["Rust"],
  "development_methods": ["TDD"],
  "hosts": ["claude", "codex"],
  "interview_revision": "docs-first-plan",
  "decisions": []
}
JSON
```

디렉터리는 `mktemp`가 만듭니다. **`project-plan`은 아무것도 쓰지 않습니다.** `.moai-cli/project.yaml` 생성, 스킬 설치, 인증, 모델 호출, GitHub 게시도 하지 않습니다. JSON envelope의 `ok`와 `result.action`, `result.delta`, `result.plan_digest`를 읽으세요. 계획 출력은 설정 적용이 아닙니다. 프로필 값에는 비공개 경로, 토큰, 자격증명을 넣지 마세요.

## 적용 경계 이해하기

임시 예제를 실제 프로젝트에 그대로 적용하지 마세요. 실제 설정에서는 정확한 대상, 제안 파일, 충돌, 사용자 승인을 먼저 확인합니다. 내부 apply 명령은 plan digest와 호출자가 제공한 permission reference를 사용하지만 **reference 문자열은 메타데이터이지 권한 검증이 아닙니다.** 계획을 apply에 전달할 때 입력은 전체 envelope가 아니라 그 안의 `result`입니다.

프로젝트 설정 순서는 프로필 plan/apply → generation plan/apply → 필요한 경우 직접 agents plan/apply입니다. generation은 자신의 `.moai-cli/workflows/<id>/` 디렉터리를 만들 수 있습니다. 직접 `agents-plan`을 호출할 때는 참조하는 워크플로 디렉터리가 이미 있어야 합니다. 이 조건을 우회하려고 빈 디렉터리를 만들지 마세요.

Rust 프로젝트 상태는 `.moai-cli/`, 홈 루트는 `~/.moai-cli/`입니다. Go의 `.moai/`·`~/.moai/` 및 사용자 소유 호스트 설정과는 별개입니다. 위 미리보기 명령은 Go 상태를 이전하거나 삭제하지 않습니다. 변경 전에는 [안전과 권한](/ko/reference/safety/)을 읽으세요.

## 호스트별 다음 단계

- [Claude Code](/ko/guides/claude-code/): AGENTS 로딩 예측, 프로젝트 설정 계획, 론처 확인.
- [Codex](/ko/guides/codex/): 지침 발견 진단과 명시적 스킬 projection 미리보기.
- [CLI 참조](/ko/reference/cli/): 명령군, JSON 출력, 작업 경계.

실행 파일이 빌드됐다고 스킬이 로딩된 것은 아닙니다. 두 호스트 안내서는 파일 계획과 실제 호스트 실행을 구분하며 실물 검증은 미확인으로 명시합니다.

## 출처

이 안내서는 보존 중인 Go 웹사이트가 아니라 [고정한 Rust 소스](https://github.com/modu-ai/moai-adk/tree/cf33967c42c31f020fd0e35dc07b6b82093119f1)를 설명합니다.

- [패키지 manifest](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/Cargo.toml), [도구 체인 지정](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/rust-toolchain.toml).
- [Rust 실행 계약](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/docs/rust-product-path.md), [CLI 선언](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/src/main.rs), [지원표](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/docs/support-matrix.md).
