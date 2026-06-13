---
id: SPEC-SEC-HARDEN-003
title: "SEC-HARDEN §F.3 fast-follow — 비격리 경로 봉쇄 (MX 사이드카 + 레거시 백업 복원)"
version: "0.1.0"
status: draft
created: 2026-06-14
updated: 2026-06-14
author: GOOS
priority: P1
tier: S
phase: "v0.2.0"
module: "internal/hook, internal/cli"
lifecycle: spec-anchored
tags: "security, path-traversal, symlink, containment, hardening, sec-harden"
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-14 | GOOS | 초안 — SEC-HARDEN-001/002의 §F.3 fast-follow. MEDIUM 2건(C-F1 MX 사이드카 비격리, C-F2 레거시 백업 심볼릭 링크) 봉쇄. Tier S, tdd reproduction-first. |

## A. 개요 (Overview)

본 SPEC은 SEC-HARDEN 보안 강화 라인(선행 SPEC-SEC-HARDEN-001, SPEC-SEC-HARDEN-002, 모두 `completed`)의 **§F.3 fast-follow**다. 선행 두 SPEC이 닫은 HIGH 등급 결함 이후, hook/update 재스윕에서 확정된 **정확히 2건의 MEDIUM 결함**을 외과적 봉쇄(surgical containment)로 해소한다.

두 앵커는 모두 2026-06-14 오케스트레이터가 ground-truth로 검증했다(라인 번호 정확, 드리프트 없음).

- **C-F1** — `internal/hook/file_changed.go` `runMXScan`: 비동기 MX 사이드카 스캔·기록이 hook stdin JSON에서 온 **비격리·공격자 영향 가능 경로**에서 동작.
- **C-F2** — `internal/cli/update.go` `restoreMoaiConfigLegacy`: 레거시 백업 복원이 **심볼릭 링크를 따라가** `backupDir`/`configDir` 밖으로 임의 읽기·쓰기를 허용.

본 SPEC은 새 추상화·새 패키지·새 플래그 표면을 추가하지 않는다. 기존 함수 2곳에 봉쇄 가드 2개를 추가하고 기존 헬퍼를 재사용한다(anti-over-engineering, Karpathy Simplicity First).

## B. 위협 모델 (Threat Model)

### C-F1 — MX 사이드카 비격리 경로 (CWE-22)

`PostToolUse` `file_changed` hook은 파일 편집 시 트리거되며 hook stdin JSON으로 `input.FilePath` / `input.CWD`를 받는다. 비동기 side-effect(`runMXScan`, REQ-HAE-001)는 현재 두 값을 **검증 없이 그대로 신뢰**한다.

- L143 `tags, err := scanner.ScanFile(input.FilePath)` — `input.FilePath`가 해소된 프로젝트 루트 내부에 있는지 확인 없이 스캔(비격리 읽기).
- L154 `projectDir := input.CWD` → L156 `stateDir := filepath.Join(projectDir, ".moai", "state")` → L157-158 `mx.NewManager(stateDir).UpdateFile(input.FilePath, tags)` — 사이드카 **쓰기** 대상이 **미검증 hook CWD**에서 유도된다. 조작된 `input.CWD`는 `.moai/state` 사이드카를 프로젝트 밖에 쓴다.

악의적 파일 편집이 `file_changed` hook을 트리거하며 `input.FilePath` / `input.CWD`를 공급하는 시나리오가 위협 표면이다.

### C-F2 — 레거시 백업 복원 심볼릭 링크 (CWE-61 / CWE-22)

`restoreMoaiConfigLegacy`는 레거시(3-way merge 이전) 백업 포맷을 처리한다.

- L2041 `filepath.Walk(backupDir, ...)` — 백업 트리 순회.
- L2062 `backupData, err := os.ReadFile(backupPath)` — 읽기 시 심볼릭 링크를 따라감.
- L2060 `targetPath := filepath.Join(configDir, relPath)` + L2072 / L2085 / L2088 `os.WriteFile(targetPath, ...)` — 쓰기 시 심볼릭 링크를 따라감(L2082는 `mergeYAMLDeep`로 쓰기 아님).

조작된 백업 디렉터리에 심볼릭 링크가 있거나 `configDir` 안 `targetPath`에 기존 심볼릭 링크가 있으면, 복원이 링크를 따라 임의 파일 읽기 증폭(CWE-61) 또는 `configDir`를 탈출하는 `relPath`/target 심볼릭 링크 경유 임의 쓰기(CWE-22)를 유발한다.

### C-F2 동형(sibling) — 모던 복원 경로 (in-scope, 본 SPEC 검증으로 확정)

`restoreMoaiConfig`(update.go L1947 — 모던 sections 복원 경로; 봉쇄 대상은 그 안의 **익명 `filepath.Walk(sectionsBackupDir, ...)` 콜백** L1964-2035이며 별도 named 함수가 아님. `configDir`는 L1948 지역변수로 콜백 스코프 내)은 **동일 심볼릭 링크 클래스**를 갖는다: L1985 `os.ReadFile(backupPath)`(링크 따라감) + L1982 `targetPath := filepath.Join(configDir, "sections", relPath)` → L1998/L2019/L2031/L2034 `os.WriteFile(targetPath, ...)`(링크 따라감). 모던 경로가 주 코드 경로이므로 레거시 폴백만 봉쇄하면 동일 취약점이 주 경로에 열려 있는 셈이다. 따라서 동일 클래스가 **provably present**임을 근거로 **in-scope sibling**으로 포함한다(아래 REQ-SEC3-006).

## C. 봉쇄 방향 (Fix Direction — 외과적 봉쇄만)

### C-F1 봉쇄 (additive guard, 비동기 구조 무변경)

1. 프로젝트 루트를 **기존** `internal/hook/path_resolve.go` `resolveProjectRootFromInputOrEnv(input, caller)` 헬퍼로 해소(B7 canonical priority: `input.CWD` → `CLAUDE_PROJECT_DIR` env → `os.Getwd()` fallback). `os.Getenv` 인라인 금지 — 중앙화된 헬퍼 사용.
2. `input.FilePath`와 계산된 사이드카 `stateDir` **둘 다** 해소된 루트 내부인지 스캔·쓰기 **전에** 봉쇄 검증.
3. 봉쇄 위반 시 fail-closed: `slog.Warn` 로그 후 early return. 비동기 side-effect는 **절대** hook 응답에 실패를 전파하지 않는다(REQ-HAE-005 design intent 보존 — main handler는 고정 빈 payload 반환).

제약: 비동기 goroutine 구조, 5s `asyncDeadline`, 빈 payload 응답 계약을 변경하지 않는다. 봉쇄는 additive guard다.

### C-F2 봉쇄 (additive guard, 레거시 + 모던 sibling)

1. 순회한 각 엔트리를 `os.Lstat`으로 검사; 심볼릭 링크 엔트리는 거부/스킵(링크를 통한 `os.ReadFile` 금지).
2. 모든 `os.WriteFile` 전에 `filepath.Rel(configDir, targetPath)`가 `configDir`를 탈출하지 않는지 검증(`..` prefix 아님, configDir 밖 절대경로 아님). target에 기존 심볼릭 링크가 있으면 거부.
3. SEC-HARDEN-001/002가 확립한 봉쇄 idiom(`internal/cli/specid` `ValidateSpecID` / `ValidateNoTraversal`) 재사용 — 단, 기존 헬퍼는 문자열 `..` 거부만 하고 root-relative `filepath.Rel` 봉쇄나 심볼릭 링크(lstat) 처리는 하지 않으므로, root-relative 봉쇄 + lstat 가드를 추가한다(중복 회피: 새 가드는 `internal/cli` 내 작은 사적 헬퍼로 두고 specid 패키지를 import하지 않는다 — specid는 문자열 sanitizer, 본 가드는 경로 봉쇄로 책임이 다르다).

제약: `restoreMoaiConfigLegacy`는 레거시 포맷을 처리한다. 모던 경로(`restoreMoaiConfig`)는 동일 심볼릭 링크 클래스가 provably present이므로 in-scope sibling으로 포함한다(§B sibling 근거).

## D. 요구사항 (Requirements — GEARS notation)

### C-F1 — MX 사이드카 봉쇄

- **REQ-SEC3-001** (Event-driven, containment): When `file_changed` hook의 비동기 `runMXScan`이 트리거되면, the `runMXScan` 함수 shall `resolveProjectRootFromInputOrEnv(input, "runMXScan")`로 프로젝트 루트를 해소한 뒤 스캔·쓰기를 수행한다.
- **REQ-SEC3-002** (Event-detected, fail-closed): When `input.FilePath`가 해소된 프로젝트 루트를 탈출하는 경로로 검출되면, the `runMXScan` 함수 shall `slog.Warn`로 로그 후 스캔 없이 early return 한다.
- **REQ-SEC3-003** (Event-detected, fail-closed): When 계산된 사이드카 `stateDir`(또는 그로부터 유도된 쓰기 대상)이 해소된 프로젝트 루트를 탈출하는 것으로 검출되면, the `runMXScan` 함수 shall 사이드카 쓰기 없이 `slog.Warn` 로그 후 early return 한다.
- **REQ-SEC3-004** (Ubiquitous, contract preservation): The `file_changed` main handler shall 봉쇄 위반 여부와 무관하게 고정 빈 payload를 반환한다 — 비동기 side-effect 실패는 hook 응답에 전파되지 않는다(REQ-HAE-005 보존).

### C-F2 — 레거시·모던 백업 복원 봉쇄

- **REQ-SEC3-005** (Event-detected, symlink reject): When `restoreMoaiConfigLegacy`의 `filepath.Walk` 순회 엔트리가 심볼릭 링크로 검출되면(`os.Lstat`), the 복원 함수 shall 해당 엔트리를 `os.ReadFile` 없이 스킵한다.
- **REQ-SEC3-006** (Event-detected, symlink reject, sibling): When `restoreMoaiConfig`(모던 경로)의 순회 엔트리가 심볼릭 링크로 검출되면, the 복원 함수 shall 해당 엔트리를 `os.ReadFile` 없이 스킵한다 — 모던 경로는 동일 심볼릭 링크 클래스를 갖는 in-scope sibling이다.
- **REQ-SEC3-007** (Event-detected, traversal reject): When 복원 쓰기 대상 `targetPath`가 `configDir`를 탈출하는 것으로 검출되면(`filepath.Rel`이 `..` prefix 또는 configDir 밖 절대경로 산출), the 복원 함수 shall `os.WriteFile` 없이 해당 엔트리를 거부한다.
- **REQ-SEC3-008** (State-driven, no-regression): While 백업 엔트리가 정규 파일이고 `targetPath`가 `configDir` 내부일 때, the 복원 함수 shall 기존 머지·복원 동작을 변경 없이 수행한다.

### 공통

- **REQ-SEC3-009** (Ubiquitous, no new surface): The 봉쇄 가드 shall 기존 함수 내부의 additive guard로만 구현되며, 새 패키지·새 공개 플래그·새 추상화를 도입하지 않는다.

## E. 비기능 제약 (Non-Functional Constraints)

- **NFR-SEC3-001**: C-F1 봉쇄는 비동기 `runMXScan` goroutine 구조, 5s `asyncDeadline`, 빈 payload 응답 계약을 변경하지 않는다.
- **NFR-SEC3-002**: C-F1은 프로젝트 루트 해소에 `os.Getenv` 인라인을 쓰지 않고 중앙화된 `resolveProjectRootFromInputOrEnv` 헬퍼만 사용한다(B7 canonical, `internal/hook/CLAUDE.md` "$CLAUDE_PROJECT_DIR resolution priority").
- **NFR-SEC3-003**: 변경 패키지(`internal/hook`, `internal/cli`)의 테스트 커버리지는 변경 전 대비 회귀하지 않는다(language go.md: ≥85% 목표).
- **NFR-SEC3-004**: 모든 봉쇄 거부는 fail-closed(거부·스킵 + 구조화 로그)이며, 거부 시 panic 하지 않는다.
- **NFR-SEC3-005**: 크로스 플랫폼 — `filepath.Rel`/`os.Lstat` 기반 봉쇄는 linux/darwin/windows에서 동작한다.

## F. Exclusions (What NOT to Build)

본 SPEC을 Tier S 2-fix 외과적 범위로 유지하기 위한 명시적 제외 항목.

### Out of Scope — 이연·비목표 항목 (deferred / non-goals)

- §F.1 `${IFS}` shell-aware tokenization 강화 — DEFERRED. 미해결 `mvdan.cc/sh` 의존성 결정에 종속됨. 별도 follow-up SPEC.
- §F.2 env-trust 강화 — 별도 follow-up SPEC.
- 새 추상화·새 패키지·새 플래그 표면 도입 — anti-over-engineering. 기존 함수 내부 봉쇄 가드만 추가한다.
- `internal/cli/specid` 패키지 자체 변경 — C-F2 봉쇄는 별도 사적 경로-봉쇄 헬퍼를 쓰며 specid(문자열 sanitizer)를 변경하지 않는다.

> 참고: 모던 3-way merge 복원 경로(`restoreMoaiConfig`)는 **동일 심볼릭 링크 클래스가 provably present**임이 본 SPEC plan-phase에서 ground-truth 검증되었으므로 제외가 아니라 **in-scope sibling**으로 포함된다(REQ-SEC3-006). 본 SPEC은 레거시·모던 두 복원 경로의 심볼릭 링크 클래스만 봉쇄하며, 그 외 update.go 코드 경로는 건드리지 않는다.

## G. 선행·후속 (Predecessors / Successors)

- 선행: SPEC-SEC-HARDEN-001 (`completed`), SPEC-SEC-HARDEN-002 (`completed`).
- 후속 후보: §F.1 `${IFS}` shell-aware SPEC, §F.2 env-trust SPEC.

## H. 재사용 봉쇄 seam (Reusable Containment Seam)

- **C-F1**: `internal/hook/path_resolve.go` `resolveProjectRootFromInputOrEnv(input, caller)` — B7 canonical 프로젝트 루트 resolver(기존). root-relative 봉쇄 검사는 본 함수 내 additive guard로 추가.
- **C-F2**: SEC-HARDEN-002 M1 `internal/cli/specid` 봉쇄 idiom을 모델로 하되, root-relative `filepath.Rel` 봉쇄 + `os.Lstat` 심볼릭 링크 가드를 `internal/cli` 내 작은 사적 헬퍼로 추가(specid import 없음 — 책임 분리).
