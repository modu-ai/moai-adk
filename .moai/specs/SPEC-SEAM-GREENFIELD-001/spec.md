---
id: SPEC-SEAM-GREENFIELD-001
title: "seam greenfield 첫 저장 500 결함 — absent 섹션 파일 원자적 기록의 stat 부재-불내성 (t544)"
version: "0.1.0"
status: completed
created: 2026-09-08
updated: 2026-09-08
author: GOOS
priority: P1
phase: "v3.2.0"
module: "internal/settings/yamlpatch"
lifecycle: spec-anchored
tier: S
tags: "yamlpatch, atomic-write, greenfield, absent-file, seam, web-console, defect, red-first"
related_specs: [SPEC-WEB-WRITE-SAFETY-001, SPEC-MCP-CONSOLE-001, SPEC-PRECOMMIT-GATE-SCOPE-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-08 | GOOS | 최초 draft. 카드 t544 (Class B — 결함, 원인 특정 완료). 결함 체인 6곳 plan-phase 직접 확인(트리 `52f863f36`, 2026-09-08). 카드 전제 대비 정정 1건: 기존 서브테스트 `TestYAMLPatchAtomicWriteErrors/"stat missing target"`(`yamlpatch_test.go:383-389`)가 결함 동작을 기대값으로 인코딩 중 — 수리 시 기대를 뒤집어 재작성하는 것이 통제 유지가 아니라 통계의 일부다. 디스패치의 `TestPatchFileValueInvariantPreservesBytes` 인용은 유효했다 — 테스트는 settings 패키지 외부 테스트 `internal/settings/write_safety_test.go:29`에 실재한다(plan-phase 초안의 부재 판정은 grep 범위가 yamlpatch 패키지에 한정돼 settings 패키지 테스트를 놓친 오판정이었고, 후속 커밋에서 수리됐다). |

---

## §1 Context & Motivation

### §1.1 관측된 결함

`moai web` 콘솔에서 섹션 파일이 **존재하지 않는** 탭(MCP 등)에서 첫 Save를 누르면 HTTP 500이 뜨고 아무 파일도 만들어지지 않는다. 반면 같은 시점에 읽기 계층은 그 부재를 정상(greenfield)으로 취급한다 — 설정 값을 읽는 쪽은 부재를 기본값으로 해석하며 오류로 치지 않는다. 저장만 실패한다.

### §1.2 측정된 결함 체인 (본 SPEC 저작 세션 직접 확인 — 트리 `52f863f36`, 2026-09-08)

| # | 사실 | 위치 | 확인 방법 |
|---|------|------|----------|
| C1 | `PatchFile`은 absent 파일을 greenfield 문서로 시작한다 — `os.ReadFile`의 `IsNotExist` → `data = []byte("{}\n")`. 의도는 코드 주석이 명시한다: "섹션 로더들이 absent를 기본값으로 해석하므로(greenfield tolerance) seam 쓰기도 그것을 거울처럼 따른다 — 첫 편집이 파일을 만든다" | `internal/settings/yamlpatch/yamlpatch.go:65-75` (주석 70-73) | 직접 판독 |
| C2 | greenfield 문서에서 `lineSplice`는 ok=false(키 미해소) → 재직렬화 폴백 → `atomicWrite(path, out)` | `internal/settings/yamlpatch/yamlpatch.go:79-83, 103-107` | 직접 판독 |
| C3 | `atomicWrite`는 `os.Stat(path)`를 **무조건** 호출하고 모든 stat 오류 — `IsNotExist` 포함 — 를 `yamlpatch: stat %s: %w`로 반환한다. stat의 유일한 목적은 원본 파일의 권한 모드 보존(`info.Mode().Perm()`, `:381`의 `os.Chmod`에 사용; `os.CreateTemp`(`:367`)은 그렇지 않으면 0600을 남긴다)이다 | `internal/settings/yamlpatch/yamlpatch.go:361-365, 367, 381` | 직접 판독 |
| C4 | 사용자 가시 체인: Save 핸들러 → `a.applySchemaEdits` → 오류 시 `renderErrorPage` → `a.render(w, http.StatusInternalServerError, view)` — Save 500 | `internal/web/handlers.go:513-517` → `internal/settings/sectionapply.go:105`(`WriteSectionViaSeam` 호출점) → `internal/settings/sectionwrite.go:75-76` → C3 → `internal/web/handlers.go:600-604` | 직접 판독 |
| C5 | seam 대상 섹션 키 등록부는 성장 중 — `mcp`(SPEC-MCP-CONSOLE-001), `gate`(SPEC-PRECOMMIT-GATE-SCOPE-001 M2), `crosssession`, `report`가 순서대로 추가됐다. 구형 템플릿으로 초기화된 프로젝트는 새 섹션 파일을 정당하게 결여한다 — **absent 섹션 파일은 부패가 아니라 템플릿 드리프트의 정상 상태**다 | `internal/settings/sectionwrite.go:31-54` (mcp `:45`, gate `:53`) | 직접 판독 |
| C6 | seam 게이트는 무변경 제출을 조용히 건너뛴다 — 제출값이 보존 스칼라와 같으면(`:68-70`) 스킵, 키가 absent이고 bool이 제출값이 absent 기본 해석과 같으면(`:72-80`) 스킵. 이 게이트를 통과한 실변경만 `WriteSectionViaSeam`에 도달한다 | `internal/settings/sectionapply.go:56-80` | 직접 판독 |
| C7 | 템플릿 섹션 파일의 측정된 모드는 전부 `-rw-r--r--`(0644) — `archive.yaml`, `cache.yaml`, `constitution.yaml`, `context.yaml`, `crosssession.yaml` 등 | `internal/template/templates/.moai/config/sections/` | `ls -l` 실측 |
| C8 | 기존 서브테스트가 **결함 동작을 기대값으로 인코딩 중**이다 — `atomicWrite`를 absent 대상으로 호출하면 오류를 기대한다. 수리는 이 기대를 뒤집는다 | `internal/settings/yamlpatch/yamlpatch_test.go:383-389` (`"stat missing target"` 서브테스트) | 직접 판독 |

모든 file:line 앵커는 트리 `52f863f36`(2026-09-08) 실측값이다. run-phase 착수 시 content-token 기준 재검증 의무가 있다(plan.md §C). 참고: 디스패치가 인용한 `yamlpatch.go:190-193`은 드리프트된 인용이다 — 그 줄은 오늘 lineSplice 재파스 가드다. 살아있는 위치는 C3(`361-365`)이다.

### §1.3 판정 방향 — 계층 대 계층 부재-계약 불일치

부재 파일의 생성은 **의도된 계약**이다 — 읽기 계층의 주석(C1)이 그렇게 선언하고 섹션 로더들의 greenfield 관용을 거울처럼 따른다. 따라서 **읽기 계층은 옳고, 쓰기 계층(`atomicWrite`의 무조건 `os.Stat`)이 결함 쪽이다.** 수리 방향: `atomicWrite`에서 stat 오류가 `os.IsNotExist`를 만족하면 문서화된 기본 모드로 진행하고, 그 외의 stat 오류(예: 권한)는 오류로 유지한다. 기본 모드 권고는 0644다 — §4에서 0600 대안을 고려하고 0644를 채택한 근거를 기록했다.

---

## §2 Requirements (GEARS)

**사용자 스토리**: "나(moai-adk 개발자)가 구형 템플릿으로 초기화된 프로젝트에서 `moai web`을 열어 MCP 탭에서 설정을 고치고 Save를 누르면, 첫 저장이 500으로 죽지 않는다. 파일이 없던 탭의 첫 저장은 파일을 만들고 내 값을 기록한다."

### REQ-1 (Ubiquitous)

The `atomicWrite` function of the `internal/settings/yamlpatch` package shall tolerate an absent target file: when its pre-write stat fails with an error satisfying `os.IsNotExist`, it shall complete the write instead of returning the stat error.

### REQ-2 (Ubiquitous)

A section file created by `atomicWrite` on an absent target shall carry permission mode 0644 — the measured convention of every file in `internal/template/templates/.moai/config/sections/` (C7).

### REQ-3 (Event-driven)

When the pre-write stat error does NOT satisfy `os.IsNotExist` (e.g. a non-directory component, a permission failure), the `atomicWrite` function shall keep returning the wrapped stat error (`yamlpatch: stat %s: %w`) — absent-tolerance shall not widen into all-stat-error tolerance.

### REQ-4 (Ubiquitous)

The temp-file + rename atomic-write convention (same-directory temp file, rename commit) shall be preserved — the absent-tolerant path shall not switch to a direct non-atomic write.

### REQ-5 (Event-driven)

When the web console Save submits a real change (a submitted value differing from the key's absent-default interpretation, per the C6 gate) for a seam section whose file is absent, the Save handler shall return HTTP 200 and the section file shall be created on disk carrying the submitted edit and mode 0644.

### REQ-6 (Unwanted)

The web-level guard test shall not use a value-invariant no-op submission as its exercise shape — the C6 gate silently skips such submissions, which would make the guard vacuously green without any fix.

### REQ-7 (Unwanted)

The line-splice byte-preservation invariants (blank lines, comments, key order, unknown keys, quoted style, typed scalars) shall not change: writes to a present section file keep their current behavior.

### REQ-8 (Event-driven)

When a mutant probe reverts the fix and a guard fails to catch the reversion, the actor shall record the uncaught mutant in progress.md as a statement of that guard's boundary — an uncaught mutant is recorded evidence, never a silently dropped run.

---

## §3 Acceptance Criteria (Given-When-Then — Tier S inline)

### AC-001 — yamlpatch 단위 greenfield 생성 (two-cell)

- **Given** 섹션 파일이 존재하지 않는 임시 트리, **When** 실변경 edit(예: absent bool 키에 명시 `false`)으로 `PatchFile`을 호출, **Then** 오류 없이 반환하고 파일이 생성되며 내용에 edit이 반영되고 모드가 0644다.
- **RED-now cell** (트리 `52f863f36`에서 채득 예정): 수리 전 트리에서는 `yamlpatch: stat …: no such file or directory`로 FAIL — RED인 이유는 결함 그 자체다(C3).
- **Green path cell**: M2 수리가 이 AC를 PASS로 뒤집는다. 뮤턴트 채득은 AC-006과 동일 절차를 단위 테스트에 적용한다.

### AC-002 — absent 외의 stat 오류는 오류로 유지

- **Given** 대상 경로의 부모가 파일(ENOTDIR — `IsNotExist`가 아닌 stat 오류)인 경우, **When** `atomicWrite`를 호출, **Then** `yamlpatch: stat …` 래핑 오류로 실패한다 — absent 관용이 전체 stat 오류로 넓어지지 않았음을 증명한다(REQ-3).

### AC-003 — 기존 서브테스트 기대 전환 (결함 인코딩 해소)

- **Given** `TestYAMLPatchAtomicWriteErrors`의 `"stat missing target"` 서브테스트(`yamlpatch_test.go:383-389`), **When** 수리가 착지, **Then** 해당 서브테스트는 absent 생성 성공 + 0644를 기대하도록 **재작성**되며 같은 테스트 함수의 `"read-only directory"` 서브테스트(`:391-413`, temp 생성 실패 계열 — stat이 아님)는 무수정으로 GREEN을 유지한다.

### AC-004 — 웹 레벨 greenfield 첫 저장 (two-cell + vacuous-green 방지)

- **Given** `.moai/config/sections/mcp.yaml`이 없는 프로젝트 트리, **When** MCP 탭에서 C6 게이트를 통과하는 실변경 제출(absent-default와 다른 명시 값)로 POST /save, **Then** 응답은 HTTP 200이고 `mcp.yaml`이 디스크에 생성되며 제출 edit을 담고 모드는 0644다.
- **RED-now cell**: 수리 전 트리(`52f863f36`)에서는 HTTP 500 + 파일 부재. **Green path cell**: M2 수리가 PASS로 뒤집는다.
- **Vacuous-green 방지 (REQ-6)**: 제출은 실변경 형태여야 한다 — 값-불변 no-op 제출은 C6 게이트가 조용히 건너뛰어 수리 없이도 통과한다. 검증 방법: 수리 전 RED 채득에서 500 배너 메시지가 `section config write failed` + stat 오류를 직접 가리켜야 한다 — 이것이 제출이 `PatchFile`에 도달했음의 증거다.

### AC-005 — 통제군: 기존 불변 무손상

- **Given** 존재하는 섹션 파일들, **When** 현행 스칼라 치환·upsert·다중 편집 흐름이 실행, **Then** 실존 통제 테스트군이 GREEN을 유지한다 — PatchFile 직접 통제군 `TestPatchFileValueInvariantPreservesBytes`(`write_safety_test.go:29`), `TestPatchFileScalarChangePreservesPresentation`(`:52`), `TestPatchFileSpliceFallsBackForUpsert`(`:320`), `TestPatchFileSpliceQuotedScalarChange`(`:341`) + yamlpatch 패키지 `TestYAMLPatchScalarReplace_WorkflowFixture`, `TestYAMLPatchPreservesQuotedStyle`, `TestYAMLPatchPreservesTypedScalars`, `TestYAMLPatchMultiEditSingleWrite`(`yamlpatch_test.go`) + settings 스키마 패밀리 `TestApplySchemaEditsSeamRoundTrip`, `TestApplySchemaEditsGateSeamRoundTrip`, `TestApplySchemaEditsAllFieldsRoundTrip`(`internal/settings`).

### AC-006 — 뮤턴트 채득 의무 (RED 진실의 판별 증거)

- **Given** 커밋된 수리 트리, **When** 수리 되돌림 오버레이를 커밋 트리 위에 적용(t517 F1-D 방법 — `SPEC-WEB-WRITE-SAFETY-001/progress.md:188-196`), **Then** (1) 오버레이 상태에서 AC-001·AC-004 가드가 FAIL로 떨어지는 것을 verbatim 채득한다(커맨드 + exit code + 트리 SHA 명기), (2) 오버레이를 복원하고 같은 커맨드 PASS를 확인한다, (3) `git diff --stat` 무출력으로 복원 완전성을 증명한다, (4) 가드가 잡지 못한 뮤턴트가 하나라도 나오면 그 사실과 뮤턴트 내용을 progress.md에 기록한다(REQ-8 — 부재-가드 AC는 RED-now만으로 채택할 수 없고 뮤턴트가 유일한 판별 증거며, 못 잡은 뮤턴트도 가드의 경계를 그리는 기록이다).

---

## §4 Design Decision — absent-허용 기본 모드 0644

**결정**: `atomicWrite`의 absent 경로는 파일을 0644로 생성한다. 이는 행위 수준의 수용 기준이며 구현 상수는 yamlpatch 패키지 안에 단일 정의점으로 둔다(구현 세부는 plan.md §F M2).

**고려한 대안 — 0600 (`os.CreateTemp` 기본값)**: absent 경로에서 stat 보존이 불가능하므로 어느 기본값이든 선택은 명시적이다. 0600을 기각한 근거: (1) 템플릿 관례는 측정상 0644다(C7) — 0600으로 만들어진 첫 파일은 형제 섹션 파일 전부와 모드가 어긋나고, 이후 편집에서 stat 보존이 그 어긋남을 영구 보존한다; (2) 섹션 파일은 git으로 동기화되고 컨테이너·비소유자 읽기가 예상되는 표면이다 — 소유자 전용 모드는 이 표면의 사용 방식과 반대로 간다.

---

## §5 Out of Scope

### Out of Scope — seam 이외 쓰기 경로

- typed 섹션 경로(`internal/config/manager.go` `Save()`), 프로필 스토어, glmcred 등 `PatchFile`/`atomicWrite` 밖의 쓰기 경로의 부재 처리는 본 SPEC이 다루지 않는다.

### Out of Scope — renderer (t545 인접 카드)

- 카드 t545(렌더 이름-유일성)는 `internal/web` 표면을 공유할 뿐 다른 결함 축이다. 본 SPEC은 renderer에 접촉하지 않으며 범위 중복이 없다.

### Out of Scope — absent 쌍 전수 스윕 (§8 후속)

- seam 표면의 모든 absent-file 읽기/쓰기 쌍에 대한 제3 사례 존재 여부 스윕은 후속 카드 소관이다(§8).

---

## §6 Verification & Quality Gates

- 검증 범위는 건드린 패키지로 한정한다: `go test ./internal/settings/... ./internal/web/...` — 전체 스위트(`go test ./...`)는 로컬에서 금지(레인 부하 규율), 전 패키지 판정은 CI 몫이다.
- RED-first 필수: AC-001·AC-004는 수리 전 트리에서 RED 채득을 먼저 하고, 채득 로그에 커맨드 + exit code + 트리 SHA를 명기한다.
- 부재 주장의 증거 절차는 `/usr/bin/grep`이다(이 트리의 셸 `grep`은 조용히 건너뛰는 ugrep 래퍼다).
- LSP 게이트: run 단계 임계(zero errors/type-errors/lint-errors) 적용.

---

## §7 Related SPECs & Adjacency

- **SPEC-WEB-WRITE-SAFETY-001**: 같은 seam 게이트(C6)와 뮤턴트-오버레이 방법론의 원천. 이 SPEC의 REQ-WWS-003(absent-기본 bool 제출은 실변경으로 기록)이 본 결함의 제출 형태(AC-004)를 정의했다. 계열 관측은 §8.
- **SPEC-MCP-CONSOLE-001 / SPEC-PRECOMMIT-GATE-SCOPE-001**: `sectionRootKeys` 성장의 원천(C5) — absent 섹션 파일을 정상 상태로 만든 존재들.

---

## §8 Series Observation — 계열 두 번째 사례 (계층 대 계층 부재-계약 불일치)

이 결함은 같은 코드베이스에서 **두 번째** 사례다:

- **사례 1 — t517 F1-D**: 게이트는 absent를 ON으로 읽고 렌더는 absent를 OFF로 읽었다(`SPEC-WEB-WRITE-SAFETY-001/progress.md:188-196`; 수리 착지, 가드 `TestHandleSaveUntouchedRenderedBodyLeavesTrackedConfigByteIdentical`).
- **사례 2 — 본 SPEC (t544)**: 읽기 계층은 absent를 greenfield-생성으로 취급하는데 쓰기 계층은 absent를 stat-실패로 취급한다.

두 사례의 공통 클래스는 **계층 간 absent 의미론의 발산**이다. 같은 클래스가 두 번 재발했으므로, 세 번째가 존재하는지 묻는 후속 카드가 정당하다 — seam 표면의 모든 absent-file 읽기/쓰기 쌍을 대상으로 한 스윕이 후속 작업이며, 본 SPEC의 범위가 아니다(§5).
