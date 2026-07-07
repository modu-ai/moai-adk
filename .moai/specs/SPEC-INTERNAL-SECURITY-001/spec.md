---
id: SPEC-INTERNAL-SECURITY-001
title: "internal/ 보안 하드닝 — audit-origin security gap 봉쇄 (web / template-update / hook)"
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P0
phase: "v3.0.0"
module: "internal/ (web, cli, hook, profile, template)"
lifecycle: spec-anchored
tags: "security, hardening, path-traversal, csrf, symlink, dead-code, hook, template-first"
tier: M
---

## HISTORY

- **2026-07-08** — read-only internal/ 보안 감사 스윕에서 도출된 8건의 verified finding을 GEARS 요구사항으로 변환하여 초안 작성. 본 SPEC은 신규 기능이 아니라 **기존 코드의 보안 결함 봉쇄**를 정의한다(감사 기원). 각 finding은 독립 감사에서 file:line 앵커와 함께 재검증되었으며, plan-phase 착수 직전 anchor spot-check(handlers.go:153/282, update.go:1274, pre_tool.go:714, 양 `.gitignore`, `assertNoUserOwnedNamespaceTouch` 비-test 호출자 0건, `file_changed.go` EvalSymlinks 존재)로 재확인되었다.

> **감사 기원 (audit-origin) 명시**: 본 SPEC은 코드를 실행하거나 수정하지 않는 read-only 감사에서 도출되었다. 각 요구사항은 "무엇이 성립해야 하는가"(acceptance.md의 AC가 정의)를 규정하며, "어떻게 고치는가"는 plan.md가 접근을 제안한다. remediation DIRECTION은 처방이 아니라 방향이다.

---

## §A. 개요 (Overview)

`internal/` 하위 3개 서브시스템(web 로컬 설정 콘솔, template-update 파이프라인, pre-tool hook 게이트)에서 서로 독립적이지만 공통 성격(경로/심링크 미검증, 방어 메커니즘 미배선, Template-First 위반)을 갖는 8건의 보안 갭이 확인되었다. 본 SPEC은 이를 3개 논리 그룹으로 묶어 mechanically-verifiable acceptance criteria로 봉쇄한다.

- **Web 그룹** (REQ-SEC-001, REQ-SEC-002): 인증 없는 GET 경로 순회 읽기 primitive + drive-by CSRF 변조.
- **Template-update 그룹** (REQ-SEC-003 ~ REQ-SEC-006): 심링크 역참조 백업 유출, Template-First `.gitignore` 누락, namespace 오분류, dead abort sentinel.
- **Hook 그룹** (REQ-SEC-007, REQ-SEC-008): 심링크 쓰기 게이트 우회, Edit 민감 콘텐츠 스캔 누락.

심각도: P0 2건(REQ-SEC-001 web 순회 읽기, REQ-SEC-006 dead sentinel), P1 5건, P2 1건.

---

## §B. GEARS 요구사항 (Requirements)

### Web 그룹

#### REQ-SEC-001 [P0] — GET 읽기 경로 순회 차단 + 중앙화된 traversal guard

- **근거 finding**: `internal/web/handlers.go:153` `handleIndex`가 `?profile=`를 `selectedProfile(r)`(app.go:174)로 받아 `buildIndexView(selected)` → `readPreferences`로 그대로 전달하며 `profile.IsValidProfileName` 가드가 없다. 반면 `handleSave`(handlers.go:282)는 가드한다. `profile.GetPreferencesPath(name)`은 `filepath.Join(GetBaseDir(), name, "preferences.yaml")`를 traversal 검사 없이 수행한다. GET은 `hostCheckMiddleware`(app.go:131-149, POST/PUT/PATCH만 게이트)를 우회하므로 `GET /?profile=../../../../etc`는 **인증 없는 임의 경로 읽기 primitive**(literally `preferences.yaml`로 명명된 파일).

> **When** 인증되지 않은 GET 요청이 `?profile=` 값을 전달하면, the web handler **shall** 파일을 읽기 전에 `profile.IsValidProfileName`(또는 중앙화된 가드)를 통과하지 못하는 값을 4xx 응답으로 거부한다.

> The traversal guard **shall** `profile.GetPreferencesPath` / `profile.GetProfileDir` 안에 중앙화되어 어떤 호출자도 경로 구분자·`..`를 포함한 이름으로 우회할 수 없다.

#### REQ-SEC-002 [P1] — 상태 변경 route의 same-origin 강제 + 오해 유발 주석 정정

- **근거 finding**: `internal/web/app.go:131` `hostCheckMiddleware`는 DNS-rebinding은 막지만 CSRF는 막지 못한다(cross-origin auto-submit POST는 정직한 `Host: 127.0.0.1:3041`을 실어 `isLoopbackHost`를 통과). 고정 기본 포트 3041(internal/cli/web.go:55). CSRF 토큰·Origin/Referer/Sec-Fetch-Site 검사·SameSite 쿠키가 어디에도 없다. `/save`, `/profile/delete`, `/profile/create`, `/__shutdown__`(DoS)의 drive-by 변조가 `moai web` 실행 중 실질적으로 가능하다. 코드 주석은 CSRF 보호를 거짓 주장한다.

> **When** 상태 변경 route(`/save`, `/profile/delete`, `/profile/create`, `/__shutdown__`)에 cross-site 요청이 도달하면, the web handler **shall** `Sec-Fetch-Site` 헤더 검사 또는 per-process CSRF 토큰으로 same-origin을 강제하고, 위반 시 4xx로 거부한다.

> The web app **shall not** hostCheckMiddleware가 CSRF 보호를 제공한다고 주장하는 오해 유발 주석을 유지한다(실제 메커니즘 부재 시 주석 정정).

### Template-update 그룹

#### REQ-SEC-003 [P1] — 사용자 소유 namespace walk 시 심링크 역참조 백업 차단

- **근거 finding**: `internal/cli/update_namespace_protect.go:100-136` `collectUserOwnedFiles`가 `.claude/skills`, `.claude/agents`, `.moai/harness`를 `filepath.WalkDir`로 순회하며 항목을 Lstat하지 않는다. `internal/cli/update_archive.go:331-354` `copyFile`은 `os.Stat`+`os.Open`(둘 다 심링크를 따름)으로 **TARGET의 내용**을 `.moai/backups/update-<stamp>/`에 기록한다. 사용자 소유 namespace 아래의 심링크가 `~/.ssh/id_rsa`를 가리키면 `moai update` 시 그 내용이 백업으로 유출된다. 올바른 Lstat 가드는 동일 패키지(update.go:2186 `isSymlinkEntry`, update_cleanup.go:278/304)에 이미 존재하나 여기에 적용되지 않았다.

> **When** `collectUserOwnedFiles`가 사용자 소유 namespace를 순회하면, the update command **shall** 각 항목을 `Lstat`하고 심링크를 skip(또는 flag)하여 `copyFile`이 심링크의 역참조된 target 내용을 `.moai/backups/`에 결코 기록하지 않도록 한다.

#### REQ-SEC-004 [P1] — Template-First `.gitignore` `.moai/backups/` 항목

- **근거 finding**: `internal/template/templates/.gitignore`에는 `.moai/backups/` 라인이 없고, root local `.gitignore:118`에는 있으나 템플릿으로 역동기화되지 않았다. REQ-SEC-003과 결합 시 `moai update` 후 `git add -A`한 사용자가 백업 내용(심링크 역참조된 secret 포함)을 커밋할 수 있다.

> The template source `.gitignore` **shall** `.moai/backups/` 항목을 포함한다.

> **Where** Template-First Rule(CLAUDE.local.md §2)이 적용되면, the change **shall** `internal/template/templates/.gitignore`를 먼저 편집한 뒤 `make build`로 임베드하며, local 프로젝트 디렉터리에 직접 추가하지 않는다.

#### REQ-SEC-005 [P1] — MoAI-reserved prefix를 가진 사용자 파일의 백업 보호

- **근거 finding**: `internal/cli/update.go:1312-1343` `isUserOwnedNamespace`가 prefix 매칭을 사용한다. `.claude/skills/<name>`·`.claude/agents/<name>`이 `moai`/`moai-`/`manager-`/`expert-`/`builder-`/`evaluator-`로 시작하면 MoAI-managed(사용자 소유 아님)로 취급된다. 사용자가 직접 만든 `moai-my-notes`·`expert-mydomain.md` 같은 파일은 오분류되어 백업에서 제외되고(REQ-SEC-006상 abort-sentinel fallback도 없어) update 시 덮어쓰기에 무방비다.

> **When** update 파이프라인이 MoAI-reserved prefix(`moai-`/`manager-`/`expert-`/`builder-`/`evaluator-`)를 가지나 provenance가 사용자 저작인 파일을 분류하면, the update command **shall** 백업 없이 그 파일을 덮어쓰거나 삭제하지 않는다(명시적 user-owned marker 요구, provenance 검증 강화, 또는 모호한 이름의 보수적 백업 확대 중 plan이 선택한 방식으로).

#### REQ-SEC-006 [P0] — dead abort sentinel 배선 또는 제거 + doctrine/@MX 정합

- **근거 finding**: `internal/cli/update_namespace_protect.go:225` `assertNoUserOwnedNamespaceTouch`(REQ-UNP-006 "pre-modification abort sentinel")는 production 코드 어디서도 호출되지 않는다(전 리포 비-test 호출자 0건 재검증). 게다가 update.go:1274 `@MX:REASON`이 이를 fan_in≥3 호출자로 거짓 나열한다. headline user-namespace 안전 메커니즘이 doctrine상 active로 주장되면서 실제로는 배선되지 않은 dead code다.

> The user-namespace safety mechanism **shall** (a) 파괴적 연산(overwrite/delete) 이전에 실행되는 실제 pre-modification 게이트로 production deploy 경로에 배선되거나, (b) 제거되고 doctrine/@MX 주장이 정정된다 — 둘 중 하나여야 한다.

> The codebase **shall not** 안전 함수를 dead 상태로 보유하면서 doctrine이 그것을 active라고 주장하는 상태를 유지한다.

### Hook 그룹

#### REQ-SEC-007 [P1] — checkFileAccess 심링크 해석(EvalSymlinks) 후 deny 매칭

- **근거 finding**: `internal/hook/pre_tool.go:665-704` `checkFileAccess`는 `file_path`를 `filepath.Abs`로만 해석하고 project-boundary 검사와 DenyPatterns 매칭(`.ssh/.*`, `id_rsa.*`, `.pem$`) 이전에 `filepath.EvalSymlinks`를 결코 호출하지 않는다. project-internal 심링크(`notes.txt -> ~/.ssh/id_rsa`)는 어떤 deny 패턴에도 매칭되지 않고 boundary 검사를 통과하지만, Write 도구가 심링크를 따라 실제 secret을 덮어쓴다(CWE-61). 올바른 EvalSymlinks-resolve-recheck 패턴은 `internal/hook/file_changed.go:176-203`(SPEC-SEC-HARDEN-004 / REQ-SEC4-004 인용)에 이미 존재하나 이 더 위험한 게이트에 전파되지 않았다.

> **When** pre-tool hook이 Write/Edit의 `file_path` 접근을 검사하면, `checkFileAccess` **shall** project-boundary 검사와 DenyPatterns 매칭 이전에 `filepath.EvalSymlinks`로 경로를 해석한다(아직 존재하지 않는 new-file Write에 대해서는 unresolved 경로로 fallback).

#### REQ-SEC-008 [P2] — Edit new_string 민감 콘텐츠 스캔

- **근거 finding**: `internal/hook/pre_tool.go:713-723` 민감 콘텐츠 패턴 스캔(private key, `sk-`/`ghp_`/`AKIA` regex)이 `if toolName == "Write"`로만 게이트된다. Edit 도구 호출(new_string 운반)은 스캔되지 않으므로 Edit로 주입된 secret은 동일 콘텐츠가 Write로 왔다면 트리거될 deny를 우회한다.

> **When** the tool is `Edit`, the pre-tool hook **shall** `new_string`을 파싱하여 `SensitiveContentPatterns`로 통과시키고, Write 콘텐츠에 이미 적용된 deny 동작과 동일하게 처리한다.

---

## §C. 비기능 제약 (Non-functional Constraints)

- **NFR-SEC-001 (Template-First 준수)**: REQ-SEC-004가 건드리는 `.gitignore`는 반드시 `internal/template/templates/.gitignore`를 먼저 편집하고 `make build`로 임베드해야 한다. local `.gitignore`에 직접 추가하는 것은 위반이다(CLAUDE.local.md §2).
- **NFR-SEC-002 (회귀 테스트 의무)**: `internal/web` 패키지 커버리지가 현재 69.8%(클러스터 최저)이므로, web 그룹(REQ-SEC-001, REQ-SEC-002) 수정은 반드시 회귀 테스트를 추가해야 한다(순회 payload 거부, cross-site POST 거부).
- **NFR-SEC-003 (behavior preservation)**: 정상 profile 이름의 GET/POST, 정상 파일 Write, new-file Write, 정상 user namespace 백업은 수정 후에도 동일하게 동작해야 한다(false-positive deny 금지).
- **NFR-SEC-004 (테스트 격리)**: 심링크/secret 관련 테스트는 `t.TempDir()` 내에서만 fixture를 생성하고 실제 `~/.ssh` 등 사용자 파일을 절대 건드리지 않는다.
- **NFR-SEC-005 (verification-claim integrity)**: 각 AC의 PASS 주장은 실제 실행한 `go test`/`grep` 출력에 귀속되어야 한다. dead-sentinel 배선/제거 여부(REQ-SEC-006)는 grep 실측으로 확정한다.

---

## §D. 제외 사항 (Exclusions)

본 SPEC은 "무엇을 만들지 않는가"를 명시적으로 규정한다. out of scope 항목은 다음과 같다.

### Out of Scope — 신규 인증/세션 프레임워크 도입

- `moai web`은 loopback-only 로컬 개발 도구이며, 본 SPEC은 로그인/세션/RBAC 같은 전면 인증 프레임워크를 도입하지 않는다. REQ-SEC-002는 same-origin 강제(Sec-Fetch-Site 또는 per-process 토큰)에 한정한다.
- 원격 노출(non-loopback bind)이나 멀티 유저 접근 제어는 별도 SPEC 소관이다.

### Out of Scope — 기존 SPEC-SEC-HARDEN 시리즈 재작업

- SPEC-SEC-HARDEN-001~005 및 SPEC-GO-TOOLCHAIN-SEC-001이 이미 봉쇄한 결함은 본 SPEC에서 재구현하지 않는다. REQ-SEC-007은 `file_changed.go`의 기존 패턴을 `pre_tool.go`로 **전파**할 뿐 새 패턴을 발명하지 않는다.

### Out of Scope — 코드 구현 자체 (plan-phase only)

- 본 문서는 plan-phase 산출물이다. 실제 `internal/` 코드 수정, `make build`, `go test` 실행, 커밋은 run-phase(manager-develop)에 위임되며 별도 세션에서 수행된다.

### Out of Scope — 16-language template ruleset 배포

- REQ-SEC-004의 `.gitignore` 항목 외에 astgrep-rules·언어별 배포 자산 등 다른 template 중립성 작업은 본 SPEC 범위 밖이다.

---

## §E. 추적성 (Traceability)

| REQ | 그룹 | 우선순위 | AC | 주 대상 파일 |
|-----|------|----------|-----|-------------|
| REQ-SEC-001 | web | P0 | AC-SEC-001a/b/c | internal/web/handlers.go, internal/profile/preferences.go, internal/profile/profile.go |
| REQ-SEC-002 | web | P1 | AC-SEC-002a/b/c | internal/web/app.go, internal/web/handlers.go |
| REQ-SEC-003 | template-update | P1 | AC-SEC-003a/b | internal/cli/update_namespace_protect.go, internal/cli/update_archive.go |
| REQ-SEC-004 | template-update | P1 | AC-SEC-004a/b | internal/template/templates/.gitignore |
| REQ-SEC-005 | template-update | P1 | AC-SEC-005a/b | internal/cli/update.go |
| REQ-SEC-006 | template-update | P0 | AC-SEC-006a/b/c | internal/cli/update.go, internal/cli/update_namespace_protect.go |
| REQ-SEC-007 | hook | P1 | AC-SEC-007a/b/c | internal/hook/pre_tool.go |
| REQ-SEC-008 | hook | P2 | AC-SEC-008a/b | internal/hook/pre_tool.go |

전체 AC 열거와 Given-When-Then 시나리오는 acceptance.md 참조.
