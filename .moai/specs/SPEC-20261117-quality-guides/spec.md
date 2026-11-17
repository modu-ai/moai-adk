---
id: SPEC-20261117-quality-guides
title: "품질 모델과 언어 가이드 — 1단계(지침과 데이터)"
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-legacy-verdicts
---

## Intent

Rust 제품에는 품질 모델이 없었어요. `verify.md`는 "저장소가 실제로 쓰는 검사 명령을 확인"이라고만 하고, 사용자
프로젝트에서 어떤 명령을 돌릴지는 정하지 않아요. 2026-11-17 사용자 승인(조사 제안 G: 감지한 도구에서 품질 모델,
AC 추적, 의존성 존재 확인)과 Go 세대 계승 결정의 새 설계 2번("품질 모델과 함께 언어·도메인 가이드")에 따라, 이
SPEC은 1단계로 지침과 데이터만 더해요: `internal/quality`와 감지기가 아는 언어마다 짧은 가이드. 새 명령은 없어요.

## Requirements

요구 문장은 영어 GEARS예요(`src/spec.rs` `parse_gears`). `When` 요구가 없어 실패 경로 요구도 없어요.

- REQ-1: The skill package shall carry `references/internal/quality.md` that takes each check command from the project's own declarations or from a guide default whose evidence is present, and records any other check as `not_run`.
- REQ-2: The quality reference shall state the AC-to-test tracing rule and the dependency existence check.
- REQ-3: The quality reference shall map every language `project-detect` can report to one guide under `references/languages/`.
- REQ-4: Each language guide shall stay within 4,096 bytes and be named by the quality reference.
- REQ-5: The VERIFY and RUN workflows shall each point to the quality reference in one added line.
- REQ-6: The chain rules shall name `internal/quality` so that a route reads it.
- REQ-7: The source manifest shall carry an entry with a current sha256 for every new package file.
- REQ-8: The change shall carry its own SPEC at `.moai/specs/SPEC-20261117-quality-guides/spec.md`.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-6): `internal/quality`가 내부 참고 목록에 있고, 머리말 형식을 지키고, chain 규칙이 이름을 불러요.
  verify: `cargo test --locked --offline --test agent_defs_package`
- AC-2 (REQ-3, REQ-4): 감지기 언어마다 가이드 행이 있고, 행이 부르는 가이드가 모두 있고, 이름 없는 가이드가 없고, 가이드는
  4,096바이트 이하예요.
  verify: `python3 -m unittest tests/skill_references_test.py`
- AC-3 (REQ-5): 두 작업 흐름이 품질 참고를 불러요.
  verify: `grep -c 'internal/quality' skills/moai/workflows/verify.md skills/moai/workflows/run.md` → 각 `1`
- AC-4 (REQ-7): manifest가 최신이에요.
  verify: `python3 -I tools/source_manifest.py check` → exit 0; `cargo test --locked --offline --test product_cli`
- AC-5 (REQ-1…REQ-8): 바뀐 곳에 닿는 검사가 통과해요.
  verify: `moai spec check --all --project .`(이 가지에서 빌드한 바이너리); `cargo test --locked --offline --test mode_workflow_steps`;
  `cargo test --locked --offline --bin moai`; `cargo fmt --check`; `moai package-inspect skills/moai`;
  `python3 -m unittest tests/source_manifest_test.py tests/repo_scans_test.py tests/ci_scope_test.py`

## Scope

- write: `skills/moai/references/internal/quality.md`, `skills/moai/references/languages/`, `skills/moai/workflows/verify.md`,
  `skills/moai/workflows/run.md`, `skills/moai/references/chains/index.md`, `skills/moai/references/source-manifest.json`,
  `tests/agent_defs_package.rs`, `tests/skill_references_test.py`, `.moai/specs/SPEC-20261117-quality-guides/spec.md`
- read-only: `src/project_detect.rs`, `src/profile.rs`, `src/spec_doc.rs`, `src/spec_links.rs`, `skills/moai/workflows/project.md`,
  `.moai-cli/manifests/legacy-asset-map.json`(개념만), Go 원본 `4fa2aa9db2`(git 객체, 개념만)
- forbidden: `src/`, `tools/`, `AGENTS.md`, `skills/moai/SKILL.md`, 모드 파일, `agents/`, 다른 internal 참고
  (`sdd`, `code-atlas`, `build-folder`), Go 파일 복사, 실제 HOME, 네트워크 쓰기

## Plan

1. `internal/quality.md`: 명령의 출처 순서(프로젝트 선언 → 근거 있는 가이드 기본값 → `not_run`), 설치 금지, 검사 종류와
   기록 형식, 감지 언어 → 가이드 표, AC↔시험 추적, 의존성 존재 확인.
2. `references/languages/` 가이드 10개(감지기 13개 언어; js·ts, java·kotlin, c·cpp는 한 파일).
3. `verify.md`·`run.md`에 한 줄씩, `chains/index.md`에 규칙 한 줄.
4. 시험: `agent_defs_package.rs` 내부 참고 목록에 `quality`, `skill_references_test.py`에 가이드 시험 3개.
5. manifest `write`·`check`, 검사, 커밋, mutation, `## Result`.

## Decisions

- D1 명령은 추측하지 않아요: 프로젝트 선언이 먼저, 가이드 기본값은 그 도구의 설정·lockfile·의존성 근거가 있을 때만, 둘 다
  없으면 `not_run`. 도구 설치는 하지 않아요(설정된 도구가 없으면 `blocked`).
- D2 프로필에는 명령이 없어요(`detected`는 언어·manifest·LSP·MCP·DB만, `tools`는 인터뷰 이름). 그래서 명령은 프로젝트
  파일에서 읽고 출처 `file:line`을 남겨요. 프로필에 명령을 기록하는 일은 후속(새 CLI 필요)이에요.
- D3 추적: AC는 `verify:` 명령을 하나 가져요(`moai spec check`가 이미 강제). 시험은 이름이나 바로 위 주석에
  `<SPEC-ID> AC-<n>`을 적어요. `src/spec_links.rs`의 `req_<segs>_<digits>` 이름 규칙은 옛 `REQ-C5-01` 꼴만 읽고 새
  `REQ-n`은 읽지 못해서 쓰지 않았어요. ACC `@cc:proves` 표식이 생기면 그 주석을 대신해요.
- D4 의존성 존재 확인은 레지스트리 명령이나 404를 주는 공개 주소로 해요. 네트워크가 없으면 `not_run`이고 추가하지
  않아요. 근거 수치는 Spracklen et al.(USENIX Security 2025, Python·JavaScript 코드 대상)의 평균 최소
  5.2%(상용 모델)·21.7%(공개 모델)예요(감사 N1로 문장 고침).
- D5 가이드는 감지기의 13개 언어를 모두 덮어요(리더 지시의 "감지기 언어마다 가이드" 시험). 같은 도구를 쓰는 언어는 한
  파일로 묶었어요: javascript·typescript, java·kotlin, c·cpp.
- D6 `internal/quality`는 `agents/*.md`의 `reads:`를 바꿀 수 없어(범위 밖) chain 규칙 한 줄로 읽히게 했어요. 시험은 그
  둘 중 하나를 요구해요.
- D7 Go 세대 언어 규칙(`design:language-domain-guides`로 보류한 행)은 개념만 읽었고 복사하지 않았어요. 그 파일들은 프레임워크
  예시 위주라, 새 가이드는 검사 명령·근거·함정만 담았어요. 도메인 가이드(백엔드·DB 등)는 후속 단계예요.

## Result

구현 에이전트가 2026-11-17에 채웠어요. SYNC와 AUDIT가 다시 확인해요. 기준 `daf8c68c69`(origin/v2026.11.0), 변경 커밋
`74ef7bf706`. 빌드 폴더는 이 worktree의 `target/`, `CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, macOS arm64, cargo는
한 번에 하나. `moai heavy-slot-acquire`가 75(스왑 81%)라 리더 지시대로 슬롯 밖에서 돌렸어요.

가이드 크기(바이트): rust 1,900 · python 2,328 · javascript-typescript 2,357 · go 1,510 · java-kotlin 1,881 · ruby 1,450 ·
php 1,575 · c-cpp 1,851 · csharp 1,706 · swift 1,567. `internal/quality.md` 4,148.

- AC-1: met — `agent_defs_package` exit 0(17 passed). mutation: chain 규칙 줄 삭제 → exit 101 "internal/quality is
  read by no agent and no chain"; 되돌린 뒤 exit 0.
- AC-2: met — `skill_references_test.py` OK(5). 가이드 시험은 파일을 쓴 뒤에 썼어요(RED 없음); 대신 mutation 2개:
  swift 행 삭제 → 2개 FAIL(행 목록, 이름·파일 일치); go 행을 `languages/golang.md`로 → 1개 FAIL. 되돌린 뒤 OK.
- AC-3: met — `verify.md:1`, `run.md:1`. 처음에 `references/internal/quality.md` 경로로 썼다가 `product_cli`
  `routes_reference_existing_workflows`(작업 흐름은 패키지 경로를 싣지 않음)가 exit 101로 막아 `internal/quality`로 바꿨어요.
- AC-4: met — `tools/source_manifest.py check` exit 0(새 항목 11개는 origin을 손으로 쓰고 `write`가 sha256을 계산);
  `product_cli` exit 0(75 passed).
- AC-5: met — `moai spec check --all --project .` exit 0(SPEC 4개, errors 0, warnings 0); `mode_workflow_steps` 0(42);
  `--bin moai` 0(27); `cargo fmt --check` 0; `moai package-inspect skills/moai` exit 0(digest
  `sha256:c3854cf03e960688434430d84a487388dce85f68a993f4e6a6f789c025c13099`); Python 4개 파일 88개 OK.

실행 안 함: 전체 `cargo test`, clippy(Rust는 시험 목록 한 줄만 바뀜), 가이드의 명령을 각 언어 도구로 실제로 돌려 보는
일(이 기기에 도구가 다 있지 않고, 1단계는 지침만), Linux·Windows.

감사(head `9e6f2a8665`) FAIL B1·N1 수리: tip `2489ad7c39` 병합(`b4d928c98f`) — `tests/agent_defs_package.rs`
내부 참고 목록에 tip의 `build-folder`와 `quality`를 둘 다 두고, `source-manifest.json`은 `tools/source_manifest.py write`가
두 쪽 항목을 합쳐 다시 계산(check exit 0). N1: `internal/quality`의 수치 문장을 "Python·JavaScript 코드, 평균 최소
5.2%·21.7%"로 고침(4,206 B). 수리 뒤: `agent_defs_package` 0(18), `product_cli` 0(75), `mode_workflow_steps` 0(42),
`moai package-inspect skills/moai` 0(digest `sha256:58de8088726d2f20c58fffaff6a38b8f544a9e5fad2cc0fe3f2cc8ed0b7020d2`),
`moai spec check --all` 0(SPEC 9개, errors 0, warnings 0), Python `skill_references`·`source_manifest` OK, `cargo fmt --check` 0.
