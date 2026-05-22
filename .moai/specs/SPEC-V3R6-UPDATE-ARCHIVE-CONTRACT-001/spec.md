---
id: SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001
title: "moai update --force 플래그 archive 전파 + skip-sync 분기 archive 단락"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.6.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "v3r6, update, archive, force, skip-sync, contract"
tier: M
---

# SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 — moai update Archive 컨트랙트 정합화

## A. Context + Motivation

### A.1 감사 배경

2026-05-23 세션 진행 중 `~/moai/mo.ai.kr` 사용자 프로젝트에서 `moai update --force` 를 반복 실행해도 `ARCHIVE_DRIFT` 경고가 영구 잔존하는 현상이 관찰되어 `moai update` 플로우 전수 감사를 수행하였다. 감사 결과 `internal/cli/update.go` 및 `internal/cli/update_archive.go` 두 파일에서 2건의 상호 독립적 결함이 확인되었고, 두 결함 모두 archive 서브시스템의 컨트랙트 일관성을 깨뜨리고 있어 단일 SPEC으로 묶어 일괄 처리한다.

### A.2 결함 #1 — `--force` 플래그가 archive 호출까지 전파되지 않음 (Critical)

코드 증거:

- `internal/cli/update.go:238` — `archiveLegacySkills(cwd, out)` 호출이 force 인자 없이 발생.
- `internal/cli/update_archive.go:233` — 시그니처 `func archiveLegacySkills(projectRoot string, out io.Writer) (int, error)` — force 파라미터 자체가 정의되지 않음.
- `internal/cli/update_archive.go:138-145, 151-158` — 에러 메시지가 `"Use --force to overwrite."` 라고 안내하지만 실제 force 처리 경로가 존재하지 않음. 메시지가 거짓말을 하고 있음.

실측 영향:

- `~/moai/mo.ai.kr/.claude/skills/moai-domain-backend/SKILL.md` 와 `~/moai/mo.ai.kr/.moai/archive/skills/v2.16/moai-domain-backend/SKILL.md` 가 `diff -q` 기준 `files differ` 로 일치하지 않음.
- 사용자가 `moai update --force` 를 몇 번을 반복 실행하든 `ARCHIVE_DRIFT` 경고는 사라지지 않음. archive 는 May 4 스냅샷이고 `.claude/skills/` 는 그 이후 수정되었으나, `--force` 가 overwrite 경로를 점화시키지 못한다.
- `update.go:85` 의 help 문구 `"Force update even if version matches (still performs backup and merge)"` 와 `update.go:108` 의 godoc 주석 `"--force: Skip backup and force the update"` 는 archive 컨트랙트와 의미 정렬이 안 되어 있다.

### A.3 결함 #2 — skip-sync 분기가 archive 검사를 계속 트리거 (High)

코드 증거:

- `internal/cli/update.go:227-241` — `runTemplateSyncWithProgress(cmd)` 가 nil 을 반환해도 (skip-sync 케이스) 제어 흐름이 무조건 archive 블록으로 진입한다.
- `internal/cli/update.go:765-772` — 패키지 버전과 프로젝트 버전이 동일하고 `--force` 가 false 일 때 `tui.Pill("Template version up-to-date · Skipping sync")` 출력 후 nil 반환. 호출자는 sync 자체를 건너뛰었는지 알 수 없다.
- 결과: 로그에 `"Skipping sync"` 직후 `"Legacy skill archive failed"` 가 연달아 출력되는 모순적 UX 가 발생.

### A.4 두 결함을 단일 SPEC 으로 묶는 이유

두 결함 모두 `archiveLegacySkills` 호출부의 컨트랙트 불일치라는 동일 도메인 위반이다. `--force` 전파 누락은 "archive 가 신호를 받지 못함" 의 한 형태이고, skip-sync 분기 누락은 "archive 가 잘못된 신호를 받음" 의 다른 형태이다. 두 결함 모두 동일한 호출부 라인 (`update.go:233-241`) 수정으로 해소되며, 동일 회귀 테스트 인프라 (`update_archive_test.go`) 를 공유한다. 분할 SPEC 보다 단일 SPEC 으로 묶어 호출부 시그니처 변경 + 호출자 분기 정합을 한 번에 처리하는 것이 컨트랙트 일관성 보장에 유리하다.

### A.5 참조

- `BC-V3R3-007` — Legacy skill archive 컨트랙트 (v2.16 stamp, idempotency 보장)
- `SPEC-V3R6-GEARS-MIGRATION-001` (머지 commit `134a43fac`, 2026-05-22) — 요구사항 표기는 EARS 가 아닌 GEARS notation 적용 (`IF/THEN` 회피, `Where/When/While` 복합 modality 선호)
- `.claude/rules/moai/workflow/spec-workflow.md` — Tier M SPEC 표준 3-file 구조

---

## B. Goals & Non-goals

### B.1 Goals

- `archiveLegacySkills` 시그니처를 확장하여 `force bool` 파라미터를 받도록 한다.
- `runUpdate` 의 archive 호출 시점에 `--force` 플래그 값을 정확히 전파한다.
- skip-sync 분기가 활성화되었을 때 archive 검사 또한 단락(short-circuit)하거나, archive 검사 자체를 sync 흐름과 무관한 별도 책임으로 명시화한다.
- `--force` 활성 시 ARCHIVE_DRIFT 발생하면 기존 archive 내용을 `archive/skills/v2.16-drift-<ISO8601_UTC>/` 로 백업한 뒤 신규 내용으로 덮어쓴다.
- 에러 메시지가 실제 코드 경로와 일치하도록 정합화한다. 존재하지 않는 옵션을 안내하지 않는다.
- 기존 BC-V3R3-007 archive 컨트랙트 (idempotency, content-hash drift detection) 는 force 비활성 경로에서 그대로 보존한다.
- 회귀 방어 테스트 (`TestArchiveForce`, `TestSkipSyncNoArchive`) 를 추가한다.
- `--force` 플래그 help 텍스트와 godoc 주석을 archive 의미까지 포함하도록 갱신한다.

### B.2 Non-goals

- v2.16 → v2.17 으로의 archive version 마이그레이션은 본 SPEC 의 범위가 아니다.
- BC-V3R3-007 archive 디렉터리 레이아웃 (`archive/skills/<version>/<id>/`) 변경은 다루지 않는다.
- `legacySkillIDs` 목록의 변경은 다루지 않는다.
- archive 내용물의 의미적 정합성 (예: 어떤 SKILL.md 가 올바른지) 검증은 다루지 않는다. 본 SPEC 은 "force 가 켜졌을 때 사용자가 의도한 overwrite 가 일어남" 만 보장한다.
- 다른 archive 서브시스템 (예: design folder, evolution dir) 의 컨트랙트는 별도 SPEC 대상이다.

### B.3 Out of Scope (다른 SPEC 으로 흡수)

- **SPEC-V3R6-UPDATE-NOISE-001** (예약명) — `moai update` 출력 중 reserved filename 알림 + 3-way merge 노이즈 정리.
- **SPEC-V3R6-UPDATE-PROGRESS-001** (예약명) — spinner / progress bar 출력 손상 (`runTemplateSyncWithProgress`) 정정.
- V3R3 archive 콘텐츠 시맨틱스 또는 v2.16 → v2.17 마이그레이션은 별도 SPEC.

---

## C. Functional Requirements

본 SPEC 은 SPEC-V3R6-GEARS-MIGRATION-001 (머지 2026-05-22) 도입의 GEARS notation 을 따른다. `IF X THEN Y` 패턴 회피, `Where/When/While` 복합 modality 우선 사용.

### REQ-UAC-001 — archiveLegacySkills 시그니처 확장

**When** `moai update` 가 archive 단계로 진입하고, **where** 호출부가 사용자의 `--force` 의도를 전달해야 할 책임이 있을 때, the system **shall** `archiveLegacySkills` 함수 시그니처를 `func archiveLegacySkills(projectRoot string, out io.Writer, force bool) (int, error)` 로 확장한다.

수용 근거: 현재 시그니처 (`update_archive.go:233`) 는 force 파라미터가 부재하여, 호출부가 `--force` 의도를 표현할 어휘 자체가 없다. 시그니처 확장은 컨트랙트 회복의 전제이다.

### REQ-UAC-002 — runUpdate 호출부의 --force 전파

**Where** `runUpdate` 가 `archiveLegacySkills` 를 호출하고, **when** `cmd.Flags().Lookup("force").Value.String() == "true"` 인 사용자 의도가 명시되어 있을 때, the system **shall** force 인자를 `true` 로 전달한다. **Where** `--force` 가 지정되지 않은 경우 force 인자는 `false` 로 전달된다.

수용 근거: 결함 #1 의 직접 해소. `update.go:238` 호출부 1-line 패치 + 시그니처 변경 호출자 정합.

### REQ-UAC-003 — Force 활성 시 ARCHIVE_DRIFT 해소

**While** `force == true` 인 호출 컨텍스트가 active 이고, **when** `checkArchiveDrift` 가 비어있지 않은 결과를 반환할 때, the system **shall** `ARCHIVE_DRIFT` 에러를 발생시키지 않고 기존 archive 디렉터리를 `archive/skills/v2.16-drift-<ISO8601_UTC>/<id>/` 로 백업한 후 신규 내용으로 overwrite 한다. **Where** 백업 디렉터리 생성 자체가 실패하면 the system **shall** 원본 archive 를 보존하고 에러를 반환한다.

수용 근거: R1 / R4 위험 완화. 사용자가 `--force` 로 overwrite 의도를 명시했더라도 기존 archive 내용을 무손실 보존하여 의도치 않은 customization 손실을 방지한다.

### REQ-UAC-004 — Skip-sync 분기에서 archive 단락

**When** `runTemplateSyncWithProgress` 의 skip-sync 분기 (version match + `!forceUpdate`) 가 활성화되어 sync 가 스킵되었을 때, **where** archive 검사를 동일 호출 사이클에서 트리거하지 않아야 할 책임이 있을 때, the system **shall** `archiveLegacySkills` 호출을 단락(short-circuit)한다. **Where** 사용자가 `--force` 를 명시한 경우 skip-sync 분기 자체가 비활성화되므로 archive 호출은 정상 진행된다.

수용 근거: 결함 #2 의 직접 해소. `runTemplateSyncWithProgress` 의 반환 시그니처를 확장하거나, runUpdate 가 skip-sync 여부를 직접 판별하는 두 접근 중 plan-phase 에서 채택한 방식으로 구현한다.

### REQ-UAC-005 — 에러 메시지 정합화

**While** ARCHIVE_DRIFT 에러 메시지를 생성하는 코드 경로 (`update_archive.go:138-158`) 가 active 이고, **where** force overwrite 경로가 실제로 구현되어 있을 때, the system **shall** 에러 메시지가 실제 코드 경로와 일치하도록 `"Use --force to overwrite."` 안내를 유지한다. **Where** force 경로가 구현되지 않은 시점에는 the system **shall** 사용자에게 거짓 안내를 제공하지 않는다.

수용 근거: 메시지의 신뢰성 회복. 본 SPEC 구현 후에는 `--force` 가 실제로 동작하므로 메시지는 유지된다. 다만 메시지와 코드의 정합 상태를 회귀 테스트로 잠근다.

### REQ-UAC-006 — BC-V3R3-007 Idempotency 보존

**Where** `force == false` 인 호출 경로가 사용될 때, the system **shall** BC-V3R3-007 의 기존 컨트랙트 (`source absent → return nil`, `archive matches → return nil`, `archive differs → ARCHIVE_DRIFT`) 를 정확히 보존한다.

수용 근거: 본 SPEC 은 force 경로 추가일 뿐 기존 idempotency 컨트랙트를 손상시키지 않는다. force 비활성 호출 회귀 테스트로 보증한다.

### REQ-UAC-007 — 회귀 방어 테스트

**When** 본 SPEC 의 구현 코드가 main 에 머지되기 전, the system **shall** 두 신규 회귀 테스트 `TestArchiveForce` (force 경로 검증) 및 `TestSkipSyncNoArchive` (skip-sync 단락 검증) 가 `go test -run` 으로 PASS 됨을 증명한다. **Where** 테스트는 `t.TempDir()` 으로 격리되며 사용자 프로젝트 파일을 변경하지 않는다.

수용 근거: CLAUDE.local.md §6 테스트 격리 원칙 준수. 회귀 방어가 없는 컨트랙트 수정은 다음 세션에서 다시 깨질 수 있다.

---

## D. Acceptance Criteria

각 AC 는 binary (PASS/FAIL) 이며 검증 커맨드를 동반한다. AC 와 REQ 간 traceability 는 `acceptance.md` 의 트레이스 표 참조.

- **AC-UAC-001** — `archiveLegacySkills` 시그니처가 `(projectRoot string, out io.Writer, force bool) (int, error)` 로 확장되었는지 grep 검증.
- **AC-UAC-002** — `runUpdate` 가 `--force` 플래그 값을 archive 호출에 전파하는지 코드 검증 + 단위 테스트 검증.
- **AC-UAC-003** — `TestArchiveForce` 가 force=true 경로에서 drift 가 해소되고 archive 가 overwrite 됨을 검증.
- **AC-UAC-004** — `TestArchiveForce` 가 force=true overwrite 시 `archive/skills/v2.16-drift-<timestamp>/` 백업 디렉터리 생성을 검증.
- **AC-UAC-005** — `TestSkipSyncNoArchive` 가 skip-sync 분기에서 archive 호출이 발생하지 않음을 검증.
- **AC-UAC-006** — force=false 경로에서 BC-V3R3-007 idempotency (source absent / matches / differs) 가 그대로 동작함을 회귀 테스트로 검증.
- **AC-UAC-007** — `git grep -n 'Use --force to overwrite' internal/` 결과의 메시지가 실제 force 경로 구현과 정합함을 코드 리뷰로 검증.
- **AC-UAC-008** — `~/moai/mo.ai.kr` 사용자 프로젝트에서 `moai update --force` 실행 시 ARCHIVE_DRIFT 경고가 더 이상 출력되지 않음을 수동 검증.
- **AC-UAC-009** — `--force` help 텍스트와 godoc 주석이 archive overwrite 의미를 포함하도록 갱신되었는지 검증.
- **AC-UAC-010** — 신규 테스트가 `t.TempDir()` 으로 격리되며, dev project 의 `.claude/skills/` 나 `.moai/archive/` 를 수정하지 않음을 코드 리뷰로 검증.

---

## E. Risks + Plan

### E.1 Risks

- **R1 — BC-V3R3-007 idempotency 컨트랙트 침해 위험**: force 경로 추가가 기존 force=false 경로의 의미를 의도치 않게 변경할 수 있다. **완화**: REQ-UAC-006 + force=false 회귀 테스트 + drift 백업 디렉터리로 무손실 overwrite 보장.
- **R2 — `--force` 의미 분절화**: 현재 `--force` 는 (a) 버전 일치 무시 (`update.go:760-772`), (b) backup/merge 강제 (help 텍스트), (c) backup 스킵 (godoc 주석) 세 가지가 모두 명세에 등장한다. archive overwrite 의미 추가로 의미가 더 복잡해진다. **완화**: plan.md 에 명시적 의미 표 작성 + godoc 주석을 단일 진실 공급원(SSOT)으로 정렬.
- **R3 — 테스트 flakiness**: archive 디렉터리 타임스탬프 백업 (`v2.16-drift-<ISO8601_UTC>`) 의 시간 의존성. **완화**: 테스트에서 timestamp 생성기를 injectable seam 으로 노출하거나, glob 패턴으로 백업 디렉터리 존재 검증.
- **R4 — 사용자의 drift archive 가 의도된 customization**: `~/moai/mo.ai.kr` 의 archive 변경이 의도된 것이라면 force overwrite 가 작업물을 파괴한다. **완화**: REQ-UAC-003 의 자동 백업으로 무손실 복구 경로 보장. 백업 디렉터리는 향후 복원 명령 (`moai migrate restore-archive-drift`) 의 대상이 될 수 있다.
- **R5 — skip-sync 단락 방식의 두 접근 (return value 확장 vs runUpdate 직접 판별) 선택 trade-off**: plan.md 에서 두 방식의 비교 후 채택. 구현 단순성과 호출자 응집도의 trade-off.

### E.2 High-level Plan

상세는 `plan.md` 참조. 본 SPEC 의 plan-phase 산출물은 markdown only 이며, run-phase 진입 후 Go 코드 변경은 ~3-5 파일 (`update.go`, `update_archive.go`, `update_archive_test.go`, 신규 `update_skip_sync_test.go`) + 기존 테스트 시그니처 호출 정합 업데이트로 예상된다.

---

## F. References

- `internal/cli/update.go:85` — `--force` 플래그 정의.
- `internal/cli/update.go:227-241` — archive 호출 블록.
- `internal/cli/update.go:760-772` — skip-sync 분기.
- `internal/cli/update_archive.go:54-94` — `archiveSkill` 본체.
- `internal/cli/update_archive.go:126-162` — `checkArchiveDrift` 본체.
- `internal/cli/update_archive.go:233-266` — `archiveLegacySkills` 시그니처.
- `internal/cli/update_archive_test.go` — 기존 archive 회귀 테스트 인프라.
- `BC-V3R3-007` — Legacy skill archive 컨트랙트.
- `SPEC-V3R6-GEARS-MIGRATION-001` — GEARS notation 표준.
- `.claude/rules/moai/workflow/spec-workflow.md` — Tier M SPEC 표준 구조.
