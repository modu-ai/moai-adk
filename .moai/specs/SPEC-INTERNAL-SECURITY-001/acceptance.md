---
id: SPEC-INTERNAL-SECURITY-001
title: "internal/ 보안 하드닝 — 인수 기준"
version: "0.1.0"
status: implemented
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P0
phase: "v3.0.0"
module: "internal/ (web, cli, hook, profile, template)"
lifecycle: spec-anchored
tags: "security, hardening, acceptance, given-when-then"
tier: M
---

## §A. 개요

각 AC는 test 또는 grep으로 mechanically-verifiable하다. AC 하위 ID(예: `AC-SEC-001a`/`AC-SEC-001b`)는 한 논리 AC 그룹의 짝 sub-criteria다. 모든 심링크/secret 테스트는 `t.TempDir()` 내 fixture만 사용한다(NFR-SEC-004).

---

## §B. Web 그룹 인수 기준

### AC-SEC-001 (REQ-SEC-001 — GET 순회 읽기 차단)

**AC-SEC-001a** — 순회 payload 거부
- **Given** `moai web`이 실행 중이고 profile base dir 밖에 임의 파일이 존재한다.
- **When** `GET /?profile=../../../../etc/passwd`(또는 동등한 순회 payload) 요청이 도달한다.
- **Then** 응답이 HTTP 4xx이고, base dir 밖의 어떤 파일도 읽히지 않는다.
- 검증: `httptest` 기반 test — 응답 코드 4xx 확인 + traversal target이 열리지 않음 확인.

**AC-SEC-001b** — read 경로 가드 배선
- **Given** `handleIndex` 코드.
- **Then** GET 읽기 경로가 `IsValidProfileName`(또는 중앙 가드)를 호출한다.
- 검증: `grep -n "IsValidProfileName" internal/web/handlers.go`가 `handleIndex`/`buildIndexView` 경로에서 매칭 (또는 중앙 가드 호출 확인).

**AC-SEC-001c** — 중앙화된 traversal 가드
- **Given** `profile.GetPreferencesPath`/`profile.GetProfileDir`.
- **When** name이 경로 구분자(`/`, `\`) 또는 `..`를 포함한다.
- **Then** 함수가 에러를 반환하거나 base dir 밖 경로를 생성하지 않는다.
- 검증: `internal/profile` 단위 test — 순회 name 입력 시 거부/안전 경로.

### AC-SEC-002 (REQ-SEC-002 — same-origin 강제)

**AC-SEC-002a** — cross-site 변조 거부
- **Given** `moai web` 실행 중.
- **When** `/save`(또는 `/profile/delete`, `/profile/create`, `/__shutdown__`)에 `Sec-Fetch-Site: cross-site`(또는 토큰 부재/불일치) POST가 도달한다.
- **Then** 응답이 4xx이고 상태 변경이 일어나지 않는다.
- 검증: `httptest` test — 각 상태 변경 route에 대해 cross-site 요청 4xx.

**AC-SEC-002b** — same-origin 요청 성공
- **Given** `moai web` 실행 중.
- **When** `Sec-Fetch-Site: same-origin`(또는 유효 CSRF 토큰) POST가 도달한다.
- **Then** 정상 처리된다(behavior preservation, NFR-SEC-003).
- 검증: `httptest` test — same-origin POST 2xx/3xx.

**AC-SEC-002c** — 오해 유발 주석 정정
- **Then** hostCheckMiddleware가 CSRF 보호를 제공한다고 주장하는 주석이, 실제 메커니즘이 부재한 채로는 존재하지 않는다.
- 검증: `grep -in "csrf" internal/web/app.go` 결과가 실제 CSRF 메커니즘 코드와 정합(거짓 주장 부재).

---

## §C. Template-update 그룹 인수 기준

### AC-SEC-003 (REQ-SEC-003 — 심링크 역참조 백업 차단)

**AC-SEC-003a** — 심링크 target 미백업
- **Given** `t.TempDir()` 내 user-owned namespace 아래에 심링크가 namespace 밖의 secret fixture를 가리킨다.
- **When** `moai update` 백업 로직이 실행된다.
- **Then** `.moai/backups/`에 심링크 target의 내용이 복사되지 않는다.
- 검증: test — 백업 산출물에 secret 콘텐츠 부재 확인.

**AC-SEC-003b** — Lstat 가드 배선
- **Then** `collectUserOwnedFiles`(또는 walk 콜백)가 항목을 정규 파일로 취급하기 전에 `Lstat`/`isSymlinkEntry`를 호출한다.
- 검증: `grep -nE "Lstat|isSymlinkEntry" internal/cli/update_namespace_protect.go` 매칭.

### AC-SEC-004 (REQ-SEC-004 — Template-First .gitignore)

**AC-SEC-004a** — template source 항목
- **Then** `internal/template/templates/.gitignore`가 `.moai/backups/` 라인을 포함한다.
- 검증: `grep -n "^\.moai/backups/" internal/template/templates/.gitignore`.

**AC-SEC-004b** — 임베드 반영 (run-phase, make build 후)
- **Given** `make build` 실행 후.
- **Then** 임베드된 template FS가 serve하는 `.gitignore`가 `.moai/backups/`를 포함한다.
- 검증: `internal/template` test — 임베드 FS에서 `.gitignore` 내용에 `.moai/backups/` 존재. (make build는 run-phase 소관.)

### AC-SEC-005 (REQ-SEC-005 — reserved-prefix 사용자 파일 보호)

**AC-SEC-005a** — 모호 이름 백업/보호
- **Given** `t.TempDir()` 내 `.claude/skills/moai-my-notes/SKILL.md`(또는 `expert-mydomain.md`)가 사용자 저작이다.
- **When** `moai update`가 실행된다.
- **Then** 이 파일이 백업 pass에 포함되거나 abort sentinel에 의해 보호되어, 백업 없이 덮어써지거나 삭제되지 않는다.
- 검증: test — update 후 원본 보존 확인 또는 백업 산출물에 포함 확인.

**AC-SEC-005b** — prefix-collision 동작 문서화
- **Then** reserved-prefix collision 케이스의 동작이 test로 명세된다.
- 검증: 해당 케이스를 다루는 test case 존재.

### AC-SEC-006 (REQ-SEC-006 — dead sentinel 배선 또는 제거)

> 이 AC는 배선(a)/제거(b) 두 설계 분기 모두 검증 가능하도록 작성됨. 정확히 한 분기가 성립해야 한다.

**AC-SEC-006a** — (분기 a) 배선된 경우
- **When** 설계 결정이 배선이면.
- **Then** `assertNoUserOwnedNamespaceTouch`(또는 그 대체 게이트)가 파괴적 연산 이전에 도달 가능한 ≥1개의 비-test production 호출자를 가진다.
- 검증: `grep -rn "assertNoUserOwnedNamespaceTouch\|<대체게이트>" internal/cli --include="*.go" | grep -v "_test.go"`가 실제 호출 site를 보여줌.

**AC-SEC-006b** — (분기 b) 제거된 경우
- **When** 설계 결정이 제거면.
- **Then** 함수와 그 `@MX` 거짓 fan_in 주장이 사라진다.
- 검증: `grep -rn "assertNoUserOwnedNamespaceTouch" internal/`가 0건 AND update.go:1274 `@MX:REASON`이 더 이상 이를 나열하지 않음.

**AC-SEC-006c** — @MX fan_in 정합
- **Then** 어떤 `@MX` 주석도 존재하지 않거나 dead인 호출자를 포함한 fan_in 개수를 주장하지 않는다.
- 검증: update.go 인근 `@MX:REASON` fan_in 주장이 실제 grep 호출자 수와 일치.

---

## §D. Hook 그룹 인수 기준

### AC-SEC-007 (REQ-SEC-007 — 심링크 쓰기 게이트)

**AC-SEC-007a** — 심링크 경유 deny target 차단
- **Given** `t.TempDir()` 내 project-internal 심링크(`notes.txt -> <deny-listed target>`, 예: `id_rsa`/`.pem`/`.ssh` 하위 fixture).
- **When** pre-tool hook이 Write `file_path=notes.txt` 접근을 검사한다.
- **Then** 접근이 deny된다.
- 검증: `internal/hook` test — 심링크 경유 deny target에 대해 deny 결정.

**AC-SEC-007b** — EvalSymlinks 배선
- **Then** `checkFileAccess`가 deny 매칭 이전에 `filepath.EvalSymlinks`를 호출한다.
- 검증: `grep -n "EvalSymlinks" internal/hook/pre_tool.go`가 `checkFileAccess` 범위에서 매칭.

**AC-SEC-007c** — new-file Write fallback 보존
- **Given** 아직 존재하지 않는 새 파일 경로(심링크 아님).
- **When** Write 접근을 검사한다.
- **Then** EvalSymlinks not-exist에도 정상 경로로 fallback하여 new-file Write가 성공한다(behavior preservation, NFR-SEC-003).
- 검증: `internal/hook` test — 미존재 정상 경로 Write allow.

### AC-SEC-008 (REQ-SEC-008 — Edit 민감 콘텐츠 스캔)

**AC-SEC-008a** — Edit secret 거부
- **Given** pre-tool hook.
- **When** `new_string`에 private key / `sk-` / `ghp_` / `AKIA` secret을 담은 Edit 호출이 도달한다.
- **Then** deny된다(동일 콘텐츠의 Write와 동일 동작).
- 검증: `internal/hook` test — secret 담은 Edit deny.

**AC-SEC-008b** — Edit 경로 게이트 확장
- **Then** 민감 콘텐츠 스캔이 `if toolName == "Write"`로만 게이트되지 않고 Edit 경로도 커버한다.
- 검증: `grep -n "toolName ==" internal/hook/pre_tool.go` 및 스캔 호출부가 Edit를 포함(또는 Write/Edit 공통 스캔 경로).

---

## §E. 품질 게이트 / Definition of Done

- [ ] 8개 REQ의 모든 AC가 PASS(실제 `go test`/`grep` 출력 귀속, NFR-SEC-005).
- [ ] `internal/web`, `internal/cli`, `internal/hook`, `internal/profile`, `internal/template` 5개 패키지 `go test ./...` 통과.
- [ ] `internal/web` 커버리지가 수정 전(69.8%) 이상 — 회귀 테스트 추가(NFR-SEC-002).
- [ ] Template-First 준수: `.gitignore`는 template source 우선 편집 + `make build`(NFR-SEC-001).
- [ ] false-positive deny 없음: 정상 profile 이름 GET/POST, new-file Write, 정상 namespace 백업 모두 정상(NFR-SEC-003).
- [ ] REQ-SEC-006 설계 분기(배선/제거) 중 정확히 하나가 성립하고 doctrine/@MX가 정합.
- [ ] `go vet ./...`, `golangci-lint run` 클린.

## §F. 엣지 케이스

- profile 이름이 빈 문자열 / 유니코드 정규화 우회(`..%2f`) / 절대 경로(`/etc/...`) — AC-SEC-001a 확장 케이스로 다룰 것.
- 심링크 체인(심링크 → 심링크 → target) — EvalSymlinks가 최종 target까지 해석하는지(REQ-SEC-007).
- 백업 대상이 심링크인 디렉터리(파일뿐 아니라 dir 심링크) — REQ-SEC-003 walk 케이스.
- `Sec-Fetch-Site` 헤더를 보내지 않는 구형 클라이언트 — same-origin 판정 정책 명확화(REQ-SEC-002; 헤더 부재 시 거부 vs 토큰 fallback).
