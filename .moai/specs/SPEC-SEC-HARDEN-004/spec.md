---
id: SPEC-SEC-HARDEN-004
title: "SEC-HARDEN §F.3 fast-follow — symlink parent-chain write escape + root symlink read amplification containment"
version: "0.1.0"
status: in-progress
created: 2026-06-14
updated: 2026-06-14
author: GOOS행님
priority: P1
tier: S
phase: "v0.2.0"
module: "internal/cli, internal/hook"
lifecycle: spec-anchored
era: V3R6
tags: "security, symlink, cwe-22, hardening, sec-harden"
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-14 | GOOS행님 | 초안 — SEC-HARDEN-003 sync-auditor가 발견한 2건의 HIGH-confidence SHOULD-FIX(인접 심볼릭 링크 공격 클래스)를 봉쇄. F1 symlinked 중간-디렉터리 쓰기 탈출(CWE-22), F2 root 내 symlink 읽기 증폭. 사용자 Option A로 본 전용 SPEC에 이연. Tier S, tdd reproduction-first. |

## A. 개요 (Overview)

본 SPEC은 SEC-HARDEN 보안 강화 라인(선행 SPEC-SEC-HARDEN-001/002/003, 모두 `completed`)의 **§F.3 fast-follow**다. 직전 SPEC-SEC-HARDEN-003의 독립 sync-auditor(verdict PASS-WITH-DEBT, Security MUST-PASS 80/100 PASS)가 **적대적 bypass 구성으로 실증 재현**한 **정확히 2건의 HIGH-confidence SHOULD-FIX**를 외과적 봉쇄로 해소한다. 두 결함 모두 SEC-HARDEN-003이 닫은 leaf-level 가드의 **인접(adjacent) 심볼릭 링크 공격 클래스**이며, SEC-HARDEN-003의 위협 모델이 닫겠다고 *주장하지 않은* 표면이다(따라서 SEC-HARDEN-003의 containment 주장을 반증하지 않으며, fast-follow debt 항목이었다).

권위 출처: `.moai/reports/sync-audit/SPEC-SEC-HARDEN-003-2026-06-14.md` §Findings(라인 62-93) + §Recommendations(라인 123-132). 두 앵커는 본 SPEC plan-phase에서 ground-truth로 재검증했다(라인 번호 정확, 드리프트 없음).

- **F1** — `internal/cli/update.go` `restoreTargetContained`(L2141): symlinked **중간 디렉터리** 쓰기 탈출(CWE-22). 현 가드는 leaf target에만 `isSymlinkEntry(absTarget)`(L2162)를 적용 → configDir 안에 이미 존재하는 symlinked subdir(`configDir/linkdir → /outside`)이 있으면 leaf `pwned.yaml`은 아직 없어(`isSymlinkEntry`=false) `filepath.Rel`이 lexically-contained(no `..`)로 판정 → `os.MkdirAll(filepath.Dir(...))` + `os.WriteFile`가 symlinked dir을 **통과**해 `/outside/evil.yaml`에 쓴다. auditor 재현됨(쓰기가 configDir 밖에 안착).
- **F2** — `internal/hook/file_changed.go` `runMXScan`(L160): symlink-in-root 읽기 증폭. `pathContainedIn(root, input.FilePath)`는 **lexical** 검사 → root 안에 lexically 존재하지만 root 밖 secret을 가리키는 symlink(`root/innocent.go → /secret/secret.go`)가 가드를 통과(contained=true) → `scanner.ScanFile`(via `os.ReadFile`, L170)가 링크를 따라감. auditor 재현됨(ReadFile이 root 밖 내용 반환).

본 SPEC은 새 추상화·새 패키지·새 플래그 표면을 추가하지 않는다. 기존 함수에 봉쇄 가드를 추가하고 SEC-HARDEN-003이 확립한 사적 헬퍼 posture(`isSymlinkEntry` os.Lstat 기반 + `restoreTargetContained` filepath.Rel 기반 in update.go; `pathContainedIn` + `resolveProjectRootFromInputOrEnv` in internal/hook)를 재사용한다(anti-over-engineering, Karpathy Simplicity First).

### A.1 설계 핵심 통찰 — F1은 공유 헬퍼 1곳 수정으로 양 walk 동시 봉쇄

ground-truth 검증 결과(plan-phase): F1의 봉쇄 대상 `restoreTargetContained`(L2141)는 **레거시 walk(`restoreMoaiConfigLegacy` L2085)와 모던 walk(`restoreMoaiConfig` 익명 콜백 L1993) 두 경로가 공유**하는 단일 헬퍼다. 따라서 parent-chain 봉쇄를 이 공유 헬퍼에 추가하면 **양 walk가 자동으로 동시 봉쇄**된다(SEC-HARDEN-003의 leaf 가드 parity가 이미 공유 헬퍼에 안착해 있음). 별도의 "modern-only vs both-walks" 분기 결정이 불필요하다 — 공유 헬퍼 수정이 곧 both-walks 봉쇄다.

## B. 위협 모델 (Threat Model)

본 §B는 sync-auditor Recommendation #2(라인 128-130)를 이행한다: 두 인접 공격 클래스를 명시적으로 열거하고 **본 SPEC 이후 CLOSED 상태로 선언**해, 다음 sync-auditor가 silent gap이 아니라 acknowledged boundary를 보도록 한다.

### B.1 — F1: symlinked 중간 디렉터리 쓰기 탈출 (CWE-22) — 본 SPEC으로 CLOSED

`restoreMoaiConfig`(모던, L1947) / `restoreMoaiConfigLegacy`(L2054)의 복원 쓰기 경로는 각 백업 엔트리에 대해 공유 헬퍼 `restoreTargetContained(configDir, targetPath)`(L2141)로 봉쇄를 검사한 뒤 `os.MkdirAll(filepath.Dir(targetPath))` + `os.WriteFile(targetPath, ...)`로 쓴다.

현 가드의 한계(SEC-HARDEN-003이 닫은 표면 = **leaf만**):
- `restoreTargetContained`는 `filepath.Rel(absConfig, absTarget)`로 lexical `..` 탈출을 거부(L2153-2159)하고, **leaf** `absTarget`에 대해서만 `isSymlinkEntry(absTarget)`(L2162)로 symlink를 거부한다.
- 그러나 `targetPath`의 **중간 경로 구성요소**(parent chain)가 symlink인 경우는 검사하지 않는다. configDir 안에 이미 `configDir/linkdir → /outside` 같은 symlinked subdir이 존재하고 백업이 `linkdir/evil.yaml` relPath를 산출하면: leaf `evil.yaml`은 아직 존재하지 않아 `isSymlinkEntry`=false, `filepath.Rel`은 lexically-contained 판정 → 가드 통과 → `os.MkdirAll` + `os.WriteFile`가 symlinked parent를 따라 `/outside/evil.yaml`에 쓴다(configDir 탈출).

**도달성 제약(severity를 낮추는 사실, FACTUAL framing)**: 본 공격은 freshly-deployed `.moai/config` 안에 **사전 존재(pre-existing)하는 symlinked 디렉터리**를 요구한다. 백업은 스스로 symlinked dir을 심을 수 없다 — symlink 엔트리는 SEC-HARDEN-003 가드(`isSymlinkEntry`)로 스킵되고, 디렉터리는 real dir로 `MkdirAll`된다. 따라서 공격자는 (a) 악의적 템플릿 또는 (b) `moai update` deploy와 restore 사이의 local FS race로 symlinked dir을 심어야 한다. 이는 genuine HIGH-confidence SHOULD-FIX이되 material reachability constraint를 동반한다 — 과장하지도 축소하지도 않는다.

봉쇄 후 상태: 본 SPEC REQ-SEC4-001/002로 `restoreTargetContained`가 parent chain을 `filepath.EvalSymlinks(filepath.Dir(targetPath))`로 해소해 resolved parent의 configDir 봉쇄를 재검사한다. **이 공유 헬퍼 수정으로 양 walk(레거시 + 모던)가 동시 봉쇄되어 본 클래스는 CLOSED.**

### B.2 — F2: symlink-in-root 읽기 증폭 (CWE-61) — 본 SPEC으로 CLOSED

`runMXScan`(L130)은 비동기 MX 사이드카 스캔에서 `resolveProjectRootFromInputOrEnv`로 해소한 root에 대해 `pathContainedIn(root, input.FilePath)`(L160)로 봉쇄를 검사한 뒤 `scanner.ScanFile(input.FilePath)`(L170, via `os.ReadFile`)로 읽는다.

현 가드의 한계(SEC-HARDEN-003이 닫은 표면 = **lexical containment만**):
- `pathContainedIn`은 `filepath.Rel` 기반 **lexical** 검사다. root 안에 lexically 존재하지만 root 밖 secret을 가리키는 symlink(`root/innocent.go → /secret/secret.go`)는 가드를 통과(contained=true)하고, `ScanFile`이 `os.ReadFile`로 링크를 따라가 root 밖 내용을 읽는다.

**영향 제약(severity를 낮추는 사실, FACTUAL framing)**: `ScanFile`은 `@MX:` prefix가 붙은 주석 라인만 `[]Tag`로 추출한다 — 임의 내용을 exfiltrate하지 않는다. 누출 표면은 가리킨 파일의 **MX-tag description 텍스트만**이며, in-root `.moai/state` 사이드카(이 자체는 봉쇄 유지 — root 밖 **쓰기 없음**)에 기록된다. 즉 MX-tag 텍스트의 좁은 정보 노출이며, 공격자는 (a) project root에 symlink를 심고 (b) 그 파일에 `file_changed` hook을 트리거해야 한다. genuine HIGH-confidence SHOULD-FIX이되 material impact constraint(MX-tag 텍스트 한정, out-of-root 쓰기 없음)를 동반한다.

봉쇄 후 상태: 본 SPEC REQ-SEC4-003/004로 `ScanFile` 전에 `filepath.EvalSymlinks(input.FilePath)`로 해소한 실경로의 `pathContainedIn(root, resolved)`를 재검사하거나 `os.Lstat`로 symlink를 스킵한다(SEC-HARDEN-003 update.go C-F2 posture와 정합). 본 클래스는 CLOSED.

### B.3 — 인접하지만 본 SPEC 범위 밖(여전히 OPEN, 명시적 이연 — §F 참조)

다음은 동일 라인의 인접 항목이나 본 Tier S 2-fix 범위 밖이며, §F에서 명시적으로 이연한다(silent gap 아님):
- TOCTOU window(check-vs-use race) — auditor가 offline single-process 위협 모델에 대해 수용함. 본 SPEC은 godoc note만(OPTIONAL, 요구사항 아님).
- root-unresolved silent fail-closed — auditor: no action needed.
- SEC-HARDEN §F.1 `${IFS}` shell-aware split + §F.2 env-trust — 별도 future follow-up SPEC.

## C. 봉쇄 방향 (Fix Direction — 외과적 봉쇄만)

### F1 봉쇄 (additive guard, 공유 헬퍼 1곳 수정 → 양 walk 동시)

1. `restoreTargetContained`(L2141)에 **parent chain symlink 해소**를 추가한다: `filepath.EvalSymlinks(filepath.Dir(targetPath))`로 부모 디렉터리 체인의 symlink를 해소한 뒤, 해소된 parent가 configDir(역시 `EvalSymlinks`로 정규화) 내부인지 `filepath.Rel` 봉쇄로 재검사한다. 위반 시 false 반환(쓰기 거부).
2. 기존 leaf 가드(`filepath.Rel` lexical `..` 거부 + leaf `isSymlinkEntry`)는 **보존**한다 — parent-chain 검사는 그 위에 얹는 additive layer다.
3. `EvalSymlinks(filepath.Dir(targetPath))`는 parent가 아직 존재하지 않을 수 있다(첫 복원). 존재하지 않는 parent는 탈출 위험이 없으므로(아직 symlink 없음) **존재하는 가장 깊은 ancestor까지 해소**하는 fail-closed 전략을 쓴다: `EvalSymlinks` error(not-exist)는 lexical 봉쇄만으로 통과시키되, 해소 성공 시 resolved parent의 configDir 봉쇄를 강제한다. 그 외 resolution error는 fail-closed(false).
4. 공유 헬퍼 수정이므로 레거시 walk(L2085)와 모던 walk(L1993) **둘 다** 자동 봉쇄된다 — 별도 분기 불필요(§A.1).

설계 선택(plan-auditor에 surface): EvalSymlinks-parent-chain vs os.Lstat-per-component. **EvalSymlinks-parent-chain을 채택**한다 — (a) symlink chain 전체를 한 번에 해소해 다단 symlink도 봉쇄, (b) cross-platform(filepath), (c) per-component Lstat loop보다 코드 표면이 작음(Simplicity First). per-component Lstat는 not-exist 구성요소 처리가 복잡하고 다단 symlink를 놓칠 수 있어 기각.

### F2 봉쇄 (additive guard, scan-target 1곳)

1. `runMXScan`(L130)에서 `scanner.ScanFile(input.FilePath)`(L170) **전에** `filepath.EvalSymlinks(input.FilePath)`로 실경로를 해소하고, `pathContainedIn(root, resolved)`로 재검사한다. 위반 시 fail-closed(`slog.Warn` 로그 후 early return, 스캔 없음).
2. 기존 lexical 가드(`pathContainedIn(root, input.FilePath)` L160)는 **보존**한다 — EvalSymlinks 재검사는 그 뒤에 얹는 additive layer다.
3. `EvalSymlinks` error 처리: not-exist(파일이 사라짐)는 스캔 대상이 없으므로 fail-closed로 early return. 그 외 error도 fail-closed.
4. `internal/hook`에 `internal/cli/specid` 패키지를 import하지 않는다 — SEC-HARDEN-003에서 확립한 책임 분리(specid = 문자열 sanitizer; 본 가드 = 경로 봉쇄)를 보존한다.

설계 선택(plan-auditor에 surface): EvalSymlinks-resolve-then-recheck vs os.Lstat-skip-symlink. **EvalSymlinks-resolve-then-recheck를 채택**한다 — `runMXScan`의 의미는 "변경된 파일을 스캔"이므로 symlink 자체를 무조건 스킵하면 정상적인 in-root symlink(드물지만 가능)도 스캔에서 누락된다. 실경로를 해소해 in-root이면 스캔하고 out-of-root이면 거부하는 편이 정밀하다. (update.go C-F2는 백업 복원이라 symlink-skip이 맞지만, scan-read는 resolve-recheck가 맞다 — 두 경로의 의미 차이 반영.)

## D. 요구사항 (Requirements — GEARS notation)

### F1 — symlinked 중간 디렉터리 쓰기 봉쇄

- **REQ-SEC4-001** (Event-detected, fail-closed): When 복원 쓰기 대상 `targetPath`의 부모 디렉터리 체인이 symlink로 configDir 밖을 가리키는 것으로 검출되면(`filepath.EvalSymlinks(filepath.Dir(targetPath))` 해소 결과가 configDir 밖), the `restoreTargetContained` 함수 shall `false`를 반환해 `os.WriteFile`를 거부한다.
- **REQ-SEC4-002** (State-driven, parity preservation): While `restoreTargetContained`가 레거시 walk(`restoreMoaiConfigLegacy`)와 모던 walk(`restoreMoaiConfig`)의 공유 헬퍼일 때, the parent-chain 봉쇄 강화 shall 공유 헬퍼 1곳 수정으로 양 walk를 동시 봉쇄하며 별도 분기를 추가하지 않는다.
- **REQ-SEC4-003** (State-driven, no-regression): While 백업 엔트리가 정규 파일이고 `targetPath`의 leaf와 parent chain이 모두 configDir 내부일 때, the `restoreTargetContained` 함수 shall 기존 leaf 가드(`filepath.Rel` lexical 거부 + leaf `isSymlinkEntry`) 및 복원 동작을 변경 없이 통과시킨다(`true` 반환).

### F2 — symlink-in-root 읽기 봉쇄

- **REQ-SEC4-004** (Event-detected, fail-closed): When `runMXScan`의 `input.FilePath`가 `filepath.EvalSymlinks` 해소 결과 해소된 프로젝트 루트를 탈출하는 symlink로 검출되면, the `runMXScan` 함수 shall `scanner.ScanFile` 없이 `slog.Warn` 로그 후 early return 한다.
- **REQ-SEC4-005** (Ubiquitous, contract preservation): The `file_changed` main handler shall F2 봉쇄 위반 여부와 무관하게 고정 빈 payload를 반환한다 — 비동기 side-effect 실패는 hook 응답에 전파되지 않는다(SEC-HARDEN-003 REQ-SEC3-004 / HOOK-ASYNC-EXPAND REQ-HAE-005 보존).
- **REQ-SEC4-006** (State-driven, no-regression): While `input.FilePath`가 symlink가 아니거나 해소 후에도 루트 내부일 때, the `runMXScan` 함수 shall 기존 lexical `pathContainedIn` 가드를 통과한 뒤 스캔·사이드카 동작을 변경 없이 수행한다.

### 공통

- **REQ-SEC4-007** (Ubiquitous, responsibility separation): The F2 봉쇄 가드 shall `internal/hook` 내에 구현되며 `internal/cli/specid` 패키지를 import하지 않는다(SEC-HARDEN-003 책임 분리 보존).
- **REQ-SEC4-008** (Ubiquitous, no new surface): The 봉쇄 가드 shall 기존 함수 내부의 additive guard로만 구현되며, 새 패키지·새 공개 플래그·새 추상화를 도입하지 않는다.

## E. 비기능 제약 (Non-Functional Constraints)

- **NFR-SEC4-001**: F2 봉쇄는 비동기 `runMXScan` goroutine 구조, 5s `asyncDeadline`, 빈 payload 응답 계약을 변경하지 않는다(SEC-HARDEN-003 NFR-SEC3-001 보존).
- **NFR-SEC4-002**: F2는 프로젝트 루트 해소에 `os.Getenv` 인라인을 쓰지 않고 기존 중앙화된 `resolveProjectRootFromInputOrEnv` 헬퍼만 사용한다(B7 canonical, `internal/hook/CLAUDE.md` "$CLAUDE_PROJECT_DIR resolution priority"; SEC-HARDEN-003 NFR-SEC3-002 보존).
- **NFR-SEC4-003**: 변경 패키지(`internal/hook`, `internal/cli`)의 테스트 커버리지는 변경 전 대비 회귀하지 않는다(language go.md: ≥85% 목표).
- **NFR-SEC4-004**: 모든 봉쇄 거부는 fail-closed(거부·스킵 + 구조화 로그)이며, 거부 시 panic 하지 않는다.
- **NFR-SEC4-005**: 크로스 플랫폼 — `filepath.EvalSymlinks`/`filepath.Rel`/`os.Lstat` 기반 봉쇄는 linux/darwin/windows에서 동작한다(`GOOS=windows GOARCH=amd64 go build` 통과 의무).

## F. Exclusions (What NOT to Build)

본 SPEC을 Tier S 2-fix 외과적 범위로 유지하기 위한 명시적 제외 항목.

### F.1 Out of Scope — 이연·비목표 항목 (deferred / non-goals)

- **TOCTOU window** (`restoreTargetContained` / `runMXScan`의 check-vs-use race) — DEFERRED. auditor가 offline single-process(`moai update` 단일 프로세스, 사용자 본인 머신) 위협 모델에 대해 수용함(report 라인 95-100). 본 SPEC은 OPTIONAL godoc note만 추가할 수 있으며, **요구사항으로 만들지 않는다**.
- **root-unresolved silent fail-closed** — auditor: no action needed(report 라인 102-107). 본 SPEC 범위 밖.
- **SEC-HARDEN §F.1 `${IFS}` shell-aware tokenization** — 별도 future follow-up SPEC. 미해결 `mvdan.cc/sh` 의존성 결정에 종속.
- **SEC-HARDEN §F.2 env-trust 강화** — 별도 future follow-up SPEC.
- **새 추상화·새 패키지·새 플래그 표면 도입** — anti-over-engineering. 기존 함수 내부 봉쇄 가드만 추가한다.
- **`internal/cli/specid` 패키지 자체 변경** — F2 봉쇄는 `internal/hook` 내 기존 헬퍼만 쓰며 specid를 import/변경하지 않는다.
- **F1/F2 외 update.go / file_changed.go 코드 경로** — 본 SPEC은 정확히 두 봉쇄 지점(`restoreTargetContained` parent-chain, `runMXScan` scan-target)만 건드린다.

## G. 선행·후속 (Predecessors / Successors)

- 선행: SPEC-SEC-HARDEN-001 (`completed`), SPEC-SEC-HARDEN-002 (`completed`), SPEC-SEC-HARDEN-003 (`completed`, 본 SPEC의 SHOULD-FIX 발견원).
- 후속 후보: §F.1 `${IFS}` shell-aware SPEC, §F.2 env-trust SPEC.

## H. 재사용 봉쇄 seam (Reusable Containment Seam)

- **F1**: `internal/cli/update.go` `restoreTargetContained(configDir, targetPath)`(L2141, 기존 공유 헬퍼) — leaf 가드 위에 `filepath.EvalSymlinks(filepath.Dir(targetPath))` parent-chain 봉쇄를 additive로 추가. 양 walk(레거시 L2085 + 모던 L1993) 공유.
- **F2**: `internal/hook/file_changed.go` `runMXScan`(L130) + 기존 `pathContainedIn`(L237) + `resolveProjectRootFromInputOrEnv`(path_resolve.go) — `ScanFile`(L170) 전에 `filepath.EvalSymlinks(input.FilePath)` 해소 + `pathContainedIn` 재검사를 additive로 추가. specid import 없음(책임 분리).
