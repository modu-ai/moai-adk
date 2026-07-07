---
id: SPEC-INTERNAL-SECURITY-001
title: "internal/ 보안 하드닝 — 구현 계획"
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P0
phase: "v3.0.0"
module: "internal/ (web, cli, hook, profile, template)"
lifecycle: spec-anchored
tags: "security, hardening, plan, milestones"
tier: M
---

## §A. 컨텍스트 (Context)

본 계획은 read-only 보안 감사에서 도출된 8건의 verified finding(spec.md §B)을 3개 논리 그룹(web / template-update / hook)으로 묶어 봉쇄하는 접근을 정의한다. 각 remediation은 **방향(DIRECTION)**이며, acceptance.md의 AC가 "무엇이 성립해야 하는가"의 최종 계약이다. 본 계획은 "어떻게"를 제안하되 AC를 좁히거나 넓히지 않는다.

기존 자산 재사용 원칙:
- REQ-SEC-003의 Lstat 가드는 동일 패키지의 `isSymlinkEntry`(update.go:2186)·`update_cleanup.go`(278/304) 패턴을 재사용한다.
- REQ-SEC-007의 EvalSymlinks-resolve-recheck는 `internal/hook/file_changed.go:176-203`의 검증된 패턴을 `pre_tool.go`로 전파한다(새 패턴 발명 금지).
- REQ-SEC-001의 검증 함수는 이미 존재하는 `profile.IsValidProfileName`(profile.go:140)을 사용하고, 중앙화는 `GetPreferencesPath`(preferences.go:82)/`GetProfileDir`(profile.go:114)에 가드를 이식한다.

## §B. 알려진 이슈 / 리스크 (Known Issues & Risks)

- **R1 (false-positive deny)**: REQ-SEC-001/007에서 traversal/symlink 가드가 지나치게 공격적이면 정상 profile 이름이나 new-file Write를 막을 수 있다. → NFR-SEC-003 behavior-preservation AC로 방어(AC-SEC-001 정상경로, AC-SEC-007c new-file fallback).
- **R2 (중앙화 range 파급)**: `GetPreferencesPath`/`GetProfileDir`에 가드를 넣으면 모든 호출자(web GET/POST, CLI profile 명령)에 영향. → 전 호출자 grep 후 정상 이름 회귀 테스트로 방어.
- **R3 (REQ-SEC-006 설계 분기)**: 배선(a) vs 제거(b)는 상호 배타적 설계 결정이다. run-phase 진입 전 오케스트레이터가 사용자에게 어느 방향인지 확인해야 할 수 있다(blocker 후보, §E 참조). AC-SEC-006은 두 분기 모두 검증 가능하도록 작성됨.
- **R4 (REQ-SEC-005 provenance 부재)**: 현재 코드에는 파일의 "사용자 저작" 여부를 판별할 provenance marker가 없다. 보수적 백업 확대(모호 이름도 백업)가 가장 낮은 리스크 경로이나 백업 용량이 늘 수 있다.
- **R5 (Template-First 순서 위반)**: REQ-SEC-004에서 local `.gitignore`를 먼저 고치는 실수. → AC-SEC-004a가 template source 경로를 명시 검증.
- **R6 (web 커버리지 69.8%)**: 회귀 테스트 없이 수정 시 은폐된 회귀 위험. → NFR-SEC-002.

## §C. Pre-flight (착수 전 확인)

- [ ] run-phase 진입 시 `internal/web`, `internal/cli`, `internal/hook`, `internal/profile`, `internal/template` 5개 패키지의 현재 테스트/빌드 baseline 캡처.
- [ ] REQ-SEC-006 설계 분기(배선 vs 제거)를 오케스트레이터가 사용자에게 확인 (§E blocker).
- [ ] `profile.GetPreferencesPath`/`GetProfileDir` 전 호출자 grep(중앙화 파급 범위 확정).
- [ ] `SensitiveContentPatterns` 정의 위치와 Write 스캔 코드 블록(pre_tool.go:713-723) 재확인.

## §D. 제약 (Constraints)

- Template-First Rule 준수(NFR-SEC-001): `.gitignore`는 template source 우선 편집 후 `make build`.
- 모든 심링크/secret 테스트는 `t.TempDir()` 내 fixture만 사용(NFR-SEC-004).
- verification-claim integrity(NFR-SEC-005): AC PASS는 실제 명령 출력에 귀속.
- 시간 추정 금지 — 우선순위 라벨과 milestone 순서만 사용.

## §E. 자체 검증 (Self-Verification)

이 계획을 run-phase로 넘기기 전 오케스트레이터가 사용자에게 확인해야 할 결정:

1. **REQ-SEC-006 방향**: (a) `assertNoUserOwnedNamespaceTouch`를 deploy 경로에 배선 vs (b) 제거 + doctrine/@MX 정정. → 이는 사용자 결정 사항이며, manager-spec은 blocker report로 반환한다. 두 분기 모두 AC-SEC-006에서 검증 가능하나, 안전 관점에서 배선(a)이 방어를 강화하고 제거(b)가 dead code를 없앤다.
2. **REQ-SEC-002 메커니즘**: Sec-Fetch-Site 헤더 검사(간단, 최신 브라우저) vs per-process CSRF 토큰(더 견고). AC-SEC-002는 둘 다 수용.
3. **REQ-SEC-005 전략**: provenance marker 요구 vs 보수적 백업 확대. R4상 보수적 확대가 저리스크.

> **Blocker 후보**: 위 3개 설계 분기는 run-phase 진입 전 사용자 확인이 권장된다. manager-develop 착수 시 결정이 주어지지 않으면 blocker report를 반환한다.

## §F. 마일스톤 (Milestones — 우선순위 기반, 시간 추정 없음)

### M1 — Web 그룹 (우선순위: 최상 / P0+P1)

- **REQ-SEC-001 (P0)**: `handleIndex` 읽기 경로에 `IsValidProfileName` 4xx 가드 추가 + `GetPreferencesPath`/`GetProfileDir` 중앙 traversal 가드 이식.
- **REQ-SEC-002 (P1)**: 상태 변경 route에 same-origin 강제(Sec-Fetch-Site 또는 CSRF 토큰) + 오해 유발 주석 정정.
- 대상 파일: `internal/web/handlers.go`, `internal/web/app.go`, `internal/profile/preferences.go`, `internal/profile/profile.go`.
- 회귀 테스트 필수(NFR-SEC-002): 순회 payload 4xx, cross-site POST 4xx, 정상경로 200.
- **선행 이유**: 인증 없는 임의 경로 읽기(REQ-SEC-001)가 8건 중 가장 즉시 악용 가능한 primitive.

### M2 — Template-update 그룹 (우선순위: 상 / P0+P1)

- **REQ-SEC-006 (P0)**: dead abort sentinel 배선 또는 제거 (§E 결정 후) + `@MX:REASON`(update.go:1274) fan_in 주장 정정.
- **REQ-SEC-003 (P1)**: `collectUserOwnedFiles` walk 콜백에 Lstat/`isSymlinkEntry` 가드 추가 → 심링크 skip/flag.
- **REQ-SEC-004 (P1)**: `internal/template/templates/.gitignore`에 `.moai/backups/` 추가 → `make build`(run-phase).
- **REQ-SEC-005 (P1)**: `isUserOwnedNamespace` prefix-collision 처리(보수적 백업 확대 또는 provenance marker).
- 대상 파일: `internal/cli/update.go`, `internal/cli/update_namespace_protect.go`, `internal/cli/update_archive.go`, `internal/template/templates/.gitignore`.
- **그룹핑 이유**: 4건 모두 `moai update` 파이프라인의 사용자 데이터 보호 경로에 집중되어 파일 편집이 응집적.

### M3 — Hook 그룹 (우선순위: 중 / P1+P2)

- **REQ-SEC-007 (P1)**: `checkFileAccess`에 `file_changed.go` EvalSymlinks-resolve-recheck 패턴 전파(new-file not-exist fallback 포함).
- **REQ-SEC-008 (P2)**: 민감 콘텐츠 스캔의 `if toolName == "Write"` 게이트를 Edit `new_string`도 포함하도록 확장.
- 대상 파일: `internal/hook/pre_tool.go`.
- **후행 이유**: 두 결함 모두 이미 다른 파일에 존재하는 패턴의 전파이며 파급 범위가 단일 파일로 국한.

## §F.1 파일 터치 인벤토리 (File-Touch Inventory)

| 파일 | REQ | 편집 성격 |
|------|-----|-----------|
| internal/web/handlers.go | REQ-SEC-001, REQ-SEC-002 | GET 가드 추가, route same-origin 검사 |
| internal/web/app.go | REQ-SEC-002 | 미들웨어/주석 정정 |
| internal/profile/preferences.go | REQ-SEC-001 | GetPreferencesPath 중앙 가드 |
| internal/profile/profile.go | REQ-SEC-001 | GetProfileDir 중앙 가드 |
| internal/cli/update.go | REQ-SEC-005, REQ-SEC-006 | isUserOwnedNamespace, @MX/deploy 경로 |
| internal/cli/update_namespace_protect.go | REQ-SEC-003, REQ-SEC-006 | collectUserOwnedFiles Lstat, sentinel 배선/제거 |
| internal/cli/update_archive.go | REQ-SEC-003 | copyFile 심링크 방어(보조) |
| internal/template/templates/.gitignore | REQ-SEC-004 | `.moai/backups/` 추가(Template-First) |
| internal/hook/pre_tool.go | REQ-SEC-007, REQ-SEC-008 | checkFileAccess EvalSymlinks, Edit 스캔 |
| (신규) *_test.go | 전체 | 회귀 테스트, 특히 internal/web(69.8%) |

## §G. 안티패턴 (Anti-Patterns)

- local `.gitignore`를 먼저 고치고 template을 잊음(Template-First 위반).
- traversal 가드를 개별 호출자에만 추가하고 `GetPreferencesPath`/`GetProfileDir` 중앙화를 생략(다음 호출자가 다시 우회 — REQ-SEC-001 핵심).
- REQ-SEC-006에서 배선/제거를 결정하지 않고 애매하게 남겨 dead code + 거짓 doctrine 상태 지속.
- 심링크 가드가 new-file Write까지 막아 정상 기능 회귀(AC-SEC-007c 위반).
- web 수정 후 회귀 테스트 없이 완료 선언(NFR-SEC-002 위반).

## §H. 교차 참조 (Cross-References)

- spec.md §B (GEARS 요구사항), acceptance.md (AC 열거).
- CLAUDE.local.md §2 (Template-First Rule), §6/§13 (테스트 격리).
- `internal/hook/file_changed.go:176-203` (REQ-SEC-007 원천 패턴, SPEC-SEC-HARDEN-004 인용).
- `.claude/rules/moai/core/verification-claim-integrity.md` (AC PASS 귀속).
