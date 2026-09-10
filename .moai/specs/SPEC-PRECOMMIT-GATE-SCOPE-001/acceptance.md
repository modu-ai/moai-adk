# acceptance.md — SPEC-PRECOMMIT-GATE-SCOPE-001

id: SPEC-PRECOMMIT-GATE-SCOPE-001
created: 2026-09-03

## D. AC Matrix

> 축 (b) + 메커니즘 1(`MOAI_PRECOMMIT=1` 마커 + `gate.pre_commit.enabled`)은 Implementation
> Kickoff Approval에서 운영자가 확정했다(plan.md §C/§F). axis-conditional 표기는 해제됐다.

### AC-001 — 실패 메시지가 remedy를 안내한다 (축 (c), 무조건) — REQ-003

**Given** `moai gate`가 실패하는 프로젝트(예: 기존 실패 테스트 1건)이고 사용자가
`gate.pre_commit.enabled`를 true로 opt-in했다
**When** pre-commit 훅이 실행되어 heavy gate가 실패한다
**Then** 훅의 stderr 출력에 다음 5개 문자열이 모두 존재한다 (grep 기계판정):
`.moai/config/sections/gate.yaml`, `gate.pre_commit.enabled`, `gate.enabled`, `gate.skip_tests`, `gate.disabled_steps`.
또한 기존 안내 `SKIP_MOAI_PRECOMMIT=1`이 유지된다.

### AC-002 — 무관한 기존 실패가 커밋을 막지 않는다 (무조건 — 축 (b) 확정) — REQ-002

**Given** 기존(project-wide) 실패 1건이 있고, staged 변경은 그 실패와 무관하며, 사용자는
`gate.pre_commit.enabled`를 설정하지 않았다(기본 OFF)
**When** 기본 설정으로 `git commit`한다
**Then** 커밋이 성공한다 (heavy gate가 pre-commit 맥락에서 실행되지 않음으로써).

### AC-003 — twin byte identity (무조건) — REQ-005

**Given** `preCommitHookContent`(`internal/cli/hook_install_precommit.go`)와
`internal/template/templates/.git_hooks/pre-commit`이 편집되어 있다
**When** `go test ./internal/cli/ -run TestPreCommitTemplateMatchesConstant -count=1`를 실행한다
**Then** 테스트가 통과한다. (같은 목적의 전 package 실행 `go test ./internal/cli/...`에서도 통과)

### AC-004 — 사용자 gate.yaml이 `moai update`를 생존한다 (무조건) — REQ-006

**Given** 기존 설치 프로젝트의 `.moai/config/sections/gate.yaml`에 사용자 커스텀 값이 있다
(예: `skip_tests: true`, `disabled_steps` 항목, 또는 `moai web` 화면에서 저장한
`pre_commit.enabled` 값)
**When** `moai update`를 실행한다
**Then** update 후에도 해당 커스텀 값이 유지된다 (shipped 템플릿 기본값으로 되돌아가지 않는다).
손편집값과 web 작성값은 **같은 REQ-006/AC-004로 검증한다** — 별도의 중복 AC를 만들지 않는다.
테스트는 fixture 프로젝트(`t.TempDir()`)에서 update 흐름을 실행해 파일 내용을 직접 비교한다.

### AC-005 — t237 소관 단계 무변경 (무조건) — REQ-008

**Given** 이 SPEC의 구현이 완료되어 있다
**When** 훅 본문의 fast-subset 구간을 검사한다
**Then** staged 파일 수집(`git diff --cached --name-only --diff-filter=ACM | grep '\.go$'`)과
`go vet $BT_TAGS $PKGS` 호출 및 `.moai/config/build-tags` 읽기 로직이 구현 전과 의미적으로 동일하다.
기계판정: 해당 구간에 대한 diff가 주석 수준이거나 0이며, `go vet` 관련 줄이 변경되지 않았음을
`git diff <base> -- internal/template/templates/.git_hooks/pre-commit`으로 확인한다.

### AC-006 — opt-in 시 heavy gate가 실행·차단한다 (무조건 — 메커니즘 1 확정) — REQ-004

**Given** 사용자가 `gate.pre_commit.enabled: true`로 opt-in했다
**When** pre-commit 훅(`MOAI_PRECOMMIT=1` 마커 하의 `moai gate`)이 실행되고 gate가 실패한다
**Then** 커밋이 차단되고(종료코드 1) AC-001의 실패 메시지가 출력된다.
반대 케이스: opt-in 상태에서 gate가 통과하면 커밋이 성공한다.
단독 불변 절: `gate.pre_commit.enabled: true`여도 마커 없는 단독 `moai gate` 실행은
기존 `gate.enabled` 계약 그대로다 — 본 키는 단독 실행 판정을 바꾸지 않는다.

### AC-007 — non-moai 프로젝트 무음 통과 (무조건) — REQ-007

**Given** PATH에 `moai`가 없는 환경이다
**When** pre-commit 훅이 실행된다
**Then** heavy gate 없이 종료코드 0으로 통과한다 (기존 동작 회귀 없음).

### AC-008 — 기존 설치에 수리가 반영된다 (무조건) — REQ-002 전제

**Given** 이전 버전의 MoAI 훅(marker 일치)이 설치된 프로젝트다
**When** 새 바이너리로 `moai update`를 실행한다
**Then** `.git/hooks/pre-commit`이 새 내용으로 교체된다 (기존 marker-based overwrite 경로,
`InstallPreCommitHook`의 marker 분기 — 별도 마이그레이션 없이 자동 반영됨을 테스트로 확인).

### AC-009 — pre-commit 맥락 기본 OFF 러너 계약 (무조건) — REQ-001

**Given** `gate.pre_commit.enabled`가 없거나 false이고, 프로젝트에 실패하는 project-wide
test/lint 단계가 존재한다
**When** `MOAI_PRECOMMIT=1 moai gate`를 실행한다
**Then** 종료코드 0으로 통과하고 project-wide heavy 단계(vet/lint/test/typecheck)를 실행하지
않는다 (gate 출력에 해당 단계의 실패가 없음). 같은 명령을 마커 없이(단독) 실행하면 기존 계약대로
실패한다 — 두 결과의 대비가 REQ-001 커밋 단위 계약의 기계적 증명이다.

### AC-010 — `moai web`에서 `gate.pre_commit.enabled` 편집 (무조건) — REQ-009

**Given** `moai web` 설정 화면이 실행 중이다
**When** 설정 UI에서 `gate.pre_commit.enabled`를 토글해 저장한다
**Then** (1) 폼이 이 필드의 컨트롤을 렌더한다 — 네이밍 규약
`name="gate.pre_commit.enabled"`(+ bool hidden companion `gate.pre_commit.enabled__present`);
(2) 저장 후 `.moai/config/sections/gate.yaml`에 `pre_commit.enabled`가 반영된다
(`WriteSectionViaSeam` 경로 — gate.yaml 외 파일에 기록 없음, 주석·미모델링 키 보존);
(3) UI 라벨이 i18n 4-locale(en/ko/ja/zh)로 존재한다. web 렌더 노출의 소관 파일은
`internal/settings/schema_sections.go` FieldDef 등록 + `internal/web/schemaform.go`
`schemaSectionMetas()` 패널 배선이다.

## D.1 심각도

| AC | 심각도 | 근거 |
|----|--------|------|
| AC-001 | Must | 사용자가 remedy에 도달하지 못하면 어느 축으로도 잔여 결함 |
| AC-002 | Must | 사고의 직접 재현 형태 — 이 SPEC의 존재 이유 |
| AC-003 | Must | twin 계약 위반은 배포판과 개발본의 갈라짐을 만든다 |
| AC-004 | Must | remedy 표면이 update마다 소멸하면 REQ-003 안내가 허위가 된다 |
| AC-005 | Must | t237 경합 범위 침범 방지 |
| AC-006 | Must | opt-in 계약의 양방향 검증 + 단독 불변 절 |
| AC-007 | Should | 기존 동작 회귀 방지 |
| AC-008 | Should | 기존 설치 반영은 `moai update` overwrite로 자동이므로 확인용 |
| AC-009 | Must | REQ-001(커밋 단위 계약)의 직접 기계 판정 |
| AC-010 | Must | 운영자 지시(moai web 편집 표면) — REQ-009의 유일 판정 |

## D.2 간접 검증 (Indirect Verification)

- AC-002는 end-to-end(실제 git 커밋 + fixture 프로젝트)로 검증하며, 단위 테스트만으로 대체하지 않는다.
- AC-004는 `CleanMoaiManagedPaths` 경로를 실제로 통과하는 update 흐름 테스트로 검증한다
  (fixture는 `t.TempDir()` — CLAUDE.local.md §6 격리 규칙).
- AC-010은 `go test ./internal/settings/... ./internal/web/...`로 schema 등록·폼 파싱·저장 경로를
  검증하고, UI 라벨은 i18n.js 4-locale 키 존재 grep으로 보조 검증한다.

## D.3 품질 게이트 (Quality Gate Criteria)

- `go test ./internal/cli/... -count=1` 통과
- `go test ./internal/config/... -count=1` 통과
- `go test ./internal/hook/... -count=1` 통과
- `go test ./internal/settings/... ./internal/web/... -count=1` 통과 (web 축)
- `make build` 성공 (템플릿 재생성 + fieldsets_templ.go 재생성 + agents-emit-check + catalog hash 체인)
- 템플릿 중립성: 훅 템플릿에 SPEC ID / 내부 날짜 / 커밋 SHA 없음 (§25.1 forbidden classes)
- go vet 영향 패키지 경고 0

## D.4 Definition of Done

1. AC-001 ~ AC-010 전량 PASS. 각 PASS는 실행한 명령과 관측 출력을 인용한다.
2. 확정 설계(축 (b) + 메커니즘 1 + moai web 편집)가 progress.md에 기록돼 있다.
3. t237과의 병합 순서/충돌 메모가 progress.md에 기록돼 있다.
