---
id: SPEC-REMOVAL-GUARD-EXTRAS-001
title: "Remove superseded rm removal regex from deployed security extras and test the deployed policy"
version: "0.1.0"
status: draft
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/template/templates, internal/hook, internal/settings/testdata"
lifecycle: spec-anchored
tier: M
tags: "pretool, security-guard, false-positive, extras, template, t511"
---

## HISTORY

| Version | Date       | Author       | Change |
|---------|------------|--------------|--------|
| 0.1.0   | 2026-09-07 | manager-spec | Initial draft — plan-phase artifacts (Tier M, card t511). GitHub #1658 + #1686: 잔여 extras 정규식 제거 + DEPLOYED policy(기본 + extras 병합) 회귀 표면 신설. 본체 구조 체크(t286, 034dad48e)는 PRESERVE. |

## A. Context (Why)

### A.1 본체는 이미 수리됐다 — 잔여 결함은 extras 사본이다

카드 t286이 커밋 `034dad48e`로 PreToolUse 위험 제거 가드의 MAIN 본체를 수리했다. `internal/hook/dangerous_removal.go`는 여섯 개의 `rm -rf <target>` 내장 정규식을 **tokenize-and-resolve-target 구조 체크**로 교체했다. 플래그 순서·클러스터·따옴표와 무관하게 **해석된 대상 경로**로 판정하며, `internal/hook/dangerous_removal_test.go`의 4개 테스트가 오늘도 통과한다 (본 러인 실측: `go test ./internal/hook/ -run 'TestDangerousRemoval' -count=1` → `ok ... 0.669s`, tree `0b1e27877`).

남은 것은 **배포되는 `security.yaml` extras 목록에 그대로 남아 있는 결함 정규식의 사본**이다. 이 SPEC의 전체 범위는 그 잔여 사본과, 그것이 t286을 살아남게 만든 테스트 공백이다.

### A.2 잔여 결함의 위치와 제거 불가능성

결함 라인은 트리 안에 **3곳**에서 확인됐다 (본 러인 직접 판독):

| # | 경로 | 행 | 내용 |
|---|------|----|------|
| C1 | `internal/template/templates/.moai/config/sections/security.yaml` | 12 | `- 'rm\s+-rf\s+/[^.]'  # rm -rf targeting root paths` |
| C2 | `.moai/config/sections/security.yaml` (로컬 추적 dogfood 사본) | 7 | `- rm\s+-rf\s+/[^.]` |
| C3 | `internal/settings/testdata/sections/security.yaml` (settings 스키마 테스트 fixture) | 7 | `- rm\s+-rf\s+/[^.]` — 디스패치에 없던 세 번째 사본. `internal/settings/sectionwrite_test.go`의 `seedSectionFixture`와 `schema_sections_test.go`가 소비 |

제거가 사용자 쪽에서 불가능한 이유는 두 겹이다:

1. **extras는 확장 전용이다.** 템플릿 헤더(`security.yaml:2`)와 `MergeExtraPatterns`(`internal/hook/pre_tool.go:298-314`)가 모두 "extend (never replace)"를 명시한다 — append-only 설계상 extras로 내장 패턴을 좁힐 방법이 없다.
2. **`.moai/config`는 `moai update`마다 통째로 삭제·재배포된다** (`CleanMoaiManagedPaths`, CLAUDE.local.md §2.3). 사용자가 라인을 지워도 다음 update가 되살린다.

### A.3 결함은 양방향이었다 — 그리고 extras 사본은 "넓은 쪽"만 남아 있다

`t286`이 수리한 본체 정규식은 좁음(플래그 순서 우회)과 넓음(모든 절대경로 오탐)을 동시에 틀렸다. extras 사본 `rm\s+-rf\s+/[^.]`은 넓은 쪽 결함을 그대로 물려받았다. 본 러인의 패턴 수준 기계 확인 (Python `re`, 구성된 subject, tree `0b1e27877`, 2026-09-07):

| subject | `(?i)rm\s+-rf\s+/[^.]` 매치 |
|---------|------------------------------|
| `rm -rf /tmp/moai-spec-repro-123` (깊은 경로 — 일상 스크래치 정리) | **True (오탐)** |
| `rm -rf /` (bare root, 문자열 끝) | **False (자기 주석 "root paths"조차 못 잡음)** |
| `rm -rf /usr` (최상위 디렉터리) | True |
| `rm -rf /.` | False |
| `echo example: rm -rf /tmp/x` (unquoted 데이터 언급) | **True (오탐)** |
| `rm -rf $HOME` (변수 간접) | False |

`[^.]` 접미사는 `/` 뒤에 비-점 문자 1개를 요구하므로 bare root 형태를 못 잡고, 깊은 절대경로는 전부 잡는다 — 자기 주석("rm -rf targeting root paths") 대비 커버리지가 뒤집혀 있다. (제거 후에는 무의해진다.)

### A.4 라이브 RED — 배포 경로에서 오늘도 매일 발화한다

이 러인 세션(tree `0b1e27877`, 워크트리 `.claude/worktrees/t511`)의 live PreToolUse 가드는 `deps.go:240-243`이 조합한 **배포 policy**(기본 + extras 병합)를 그대로 실행한다. 세 가지 클래스 모두 동일 메시지로 거부됐고, 명령은 실행조차 되지 않았다:

- **R1 (heredoc 데이터 언급)**: `cat > /tmp/t511-spec-repro.md` heredoc 본문에 패턴 리터럴과 `rm -rf /tmp/moai-spec-repro-123` 예시 형태를 담은 쓰기 → 거부. verbatim: `Dangerous command blocked: (?i)rm\s+-rf\s+/[^.]`
- **R2 (unquoted 데이터 언급)**: `echo example: rm -rf /tmp/moai-spec-repro-456` → 동일 메시지로 거부.
- **R3 (실행 클래스 — #1658 원보고 형태)**: `rm -rf /tmp/t511-scratch-nonexistent-dir` → 동일 메시지로 거부.

거부 메시지가 패턴 리터럴 자체(`(?i)rm\s+-rf\s+/[^.]`)라는 점이 발화 주체를 특정한다. 구조 체크가 발화했으면 메시지는 `removal of protected path %q` 형태일 것이다(`pre_tool.go:948`), 그리고 세 사례 모두 대상이 비보호 깊은 경로라 구조 체크는 통과했다. 즉 **extras 정규식이 유일한 발화 성분**이다. R1은 운영자 라이브 재현(2026-09-07)과 바이트 동일하며, 본 카드 앞으로 5번째 독립 관측이다.

### A.5 테스트 공백이 t286을 살아남게 한 이유

기존 4개 테스트(`internal/hook/dangerous_removal_test.go`)는 전부 `decideBash` 헬퍼가 `preToolHandler{policy: DefaultSecurityPolicy()}`로 **기본 policy만** 구성한다(파일 12행). 템플릿 extras는 로드되지 않는다. 배포 정책(기본 + extras)을 exerc하는 테스트가 한 개도 없으므로, extras 경로의 오탐은 t286의 GREEN 아래에 그대로 남았다. 이 공백 자체가 이 SPEC의 수리 대상이다 (REQ-RGE-004).

### A.6 #1686 제안 2(override 키)를 배제하는 근거

#1686은 `disable_builtin_bash_patterns` 류의 override 키를 제안했다. 구조 체크가 이미 두 방향을 모두 바로잡았고(플래그 순서 개념 자체가 소멸), extras 사본 제거가 오탐 원천을 없애므로 override 키의 필요가 소멸한다. 보고자 자신의 프레이밍("앵커 하나로 오탐 대부분이 사라진다")조차 구조 체크가 초과 달성한다. 신규 config 표면은 추가하지 않는다 (§F Out of Scope).

## B. Approach

### B.1 수리 형태 — 제거 한 줄, 회귀 표면 신설

수리는 두 부분이다:

1. **M1 — 잔여 라인 제거**: C1(템플릿) → `make build` → C2(dogfood)·C3(testdata) 순. Template-First 주기(CLAUDE.local.md §2 [HARD]). 템플릿 파일에 이슈 번호·SPEC ID·날짜 주석을 **추가하지 않는다**(중립성, §C.6).
2. **M2 — DEPLOYED policy 회귀 표면**: `DefaultSecurityPolicy()` + `MergeExtraPatterns(...)` 로 배포 policy를 구성하는 헬퍼를 새로 두고, 허용 방향(데이터 언급·일상 정리)과 차단 방향(보호 대상 전 플래그 순서)을 모두 배포 policy로 exerc한다. 템플릿 본체에서 임베디드 FS로 `security.yaml`을 읽어 구성하는 것이 원칙(패키지 import 사이클 없음 — 본 러인 확인: `internal/template`은 `internal/hook`을 import하지 않음). disk 경로 폴백과 inline 동등 구성이 차수 폴백이다.
3. **M3 — 문서화된 한계 기록**: 코드 변경 없이 spec.md에 scope-out으로 기록 (§F).

### B.2 양방향 쌍 규율

카드의 HARD 지시("한쪽만 넣으면 반대 방향이 되살아난다")에 따라 모든 방향 완화 AC는 방향 강화 AC와 쌍을 이룬다:

- **완화(허용)**: heredoc 깊은 경로 예시, unquoted 데이터 언급, 실행형 스크래치 정리 (RED-now: R1·R2·R3)
- **강화(차단)**: 보호 대상 전 플래그 순서/클러스터/따옴표/복합문 (현재 GREEN, 뮤턴트 B로 RED 확인)

뮤턴트 프로브 두 개를 SPEC이 명명하고 런 페이즈가 실행한다:

- **뮤턴트 A**: 제거한 라인을 템플릿에 재추가 → AC-RGE-001/002/003 테스트가 RED로 복귀해야 한다. 재추가에도 GREEN이면 테스트가 extras를 로드하지 않는다는 뜻(공허한 통과) — REQ-RGE-004의 판별 증거.
- **뮤턴트 B**: `checkBashCommand`의 구조 체크 호출(`pre_tool.go:947-949`)을 되돌림 → AC-RGE-005/006 테스트가 RED로 복귀해야 한다. t286 수리의 회귀를 잡는 축.

### B.3 기존 테스트와의 관계

기존 4개 `TestDangerousRemoval_*` 테스트는 건드리지 않는다(builtin-only 구성도 유지 — 그 자체가 기본 policy의 회귀 표면이다). 배포 policy 표면은 별도 테스트 함수군으로 추가한다.

## C. Requirements (GEARS)

### C.1 결함 제거 (오탐 방향)

- **REQ-RGE-001** (Event-detected): **When** PreToolUse Bash 가드가 배포 policy(기본 + `security.yaml` extras)로 명령을 스캔할 때, heredoc 본문 또는 unquoted 텍스트가 깊은 경로 제거의 **예시 형태를 데이터로 언급**하는 것만으로 차단되어서는 안 된다. (R1·R2 RED-now)

- **REQ-RGE-002** (Event-detected): **When** 배포 policy로 스캔할 때, 비보호 깊은 경로를 대상으로하는 **실제 제거 명령**(예: `/tmp` 스크래치 정리)은 구조 체크의 판정을 따라야 하며 extras 정규식으로 차단되어서는 안 된다. (R3 RED-now — #1658 원보고)

- **REQ-RGE-003** (Ubiquitous): 템플릿 `security.yaml`의 `extra_dangerous_bash_patterns` 목록은 구조 체크가 이미 판정하는 형태를 포섭하는 잔여 정규식(`rm\s+-rf\s+/[^.]`)을 **carry해서는 안 된다** — 템플릿(C1), 로컬 추적 dogfood 사본(C2), settings testdata fixture(C3) 3곳 모두에서.

### C.2 차단 방향 보존 (t286 수리의 회귀 방어)

- **REQ-RGE-004** (Ubiquitous): 배포 policy에서 파일시스템 루트·최상위 디렉터리·홈 별칭(`~`, `$HOME`, `${HOME}`)·`.git`·`node_modules`를 대상으로하는 제거는 플래그 순서·클러스터·따옴표·복합문(`echo x && rm -fr /`) 형태와 무관하게 거부되어야 한다 — 구조 체크 경유.

- **REQ-RGE-005** (Ubiquitous): heredoc 본문이 보호 대상(`rm -rf /` 등)을 담을 때도 배포 policy는 거부해야 한다 — 개행 분할이 본문 행을 명령 세그먼트로 심사하기 때문 (현재 동작의 특성화; §F M3 한계 기록과 표면을 공유).

- **REQ-RGE-006** (Ubiquitous): 제거 가드 동작을 주장하는 회귀 테스트는 **배포 policy**(기본 + extras 병합)를 구성해야 한다 — builtin-only policy 구성만으로는 REQ-RGE-001~005를 커버했다고 주장할 수 없다.

- **REQ-RGE-007** (Ubiquitous): REQ-RGE-001~002의 허용 AC는 **뮤턴트 A**(잔여 라인 재추가)에서 RED로 복귀해야 하고, REQ-RGE-004~005의 차단 AC는 **뮤턴트 B**(구조 체크 호출 되돌림)에서 RED로 복귀해야 한다 — 뮤턴트는 SPEC이 명명하고 런 페이즈가 실행한다.

### C.3 배포 메커니즘

- **REQ-RGE-008** (State): 템플릿 편집 후 `make build`가 성공해야 하고, 임베디드 FS가 편집된 `security.yaml`을 반영해야 한다 (`//go:embed all:templates` 재컴파일).

- **REQ-RGE-009** (Ubiquitous): 템플릿 파일은 이슈 번호·SPEC ID·내부 날짜를 참조하는 주석을 **획득해서는 안 된다** (C1-C8 중립성, CI 가드 `template-neutrality-check.yaml`).

### C.4 확장 계약 불변

- **REQ-RGE-010** (Ubiquitous): 이 수리는 extras의 "extend, never replace" 계약과 사용자 facing override 표면 부재를 변경해서는 안 된다 — 다른 extras 라인(curl/wget/chmod/mkfs/dd)은 건드리지 않는다.

## D. Constraints (HARD)

- `internal/hook/branch_guard.go`의 로직은 변경 금지 (`substituteQuotedArguments` 포함 — t512/#1659 소관).
- `rm\s+-rf\s+/[^.]` 외의 어떤 extras 패턴 라인도 변경 금지.
- 기존 `TestDangerousRemoval_*` 4개 테스트와 `decideBash` 헬퍼는 보존 (builtin-only 구성 포함).
- 코드 주석·godoc 영어; 커밋 메시지 영어(Conventional Commits, 카드 id t511 포함).
- SPEC 본문 한국어 산문 + 영어 기술 식별자 (documentation: ko).
- plan 페이즈에서 코드·템플릿 파일 수정 금지 — 본 페이즈는 SPEC 아티팩트만 저작.

## E. Cross-References

- GitHub #1658 (양방향 정규식 결함 원보고), #1686 (잔여 extras 오탐 + override 제안)
- t286 / 커밋 `034dad48e` — 본체 구조 체크 수리 (PRESERVE)
- t512 / GitHub #1659 — worktree-guard heredoc brace folding (다른 가드, 다른 카드)
- `.claude/rules/moai/development/verification-completeness.md` — two-cell 채택, 뮤턴트 프로브, empty-sweep 규율
- `.claude/rules/moai/core/verification-claim-integrity.md` — RED 증거의 baseline 귀속
- `internal/hook/dangerous_removal.go` / `dangerous_removal_test.go` / `pre_tool.go` / `internal/cli/deps.go` / `internal/hook/security/config.go`

## F. Out of Scope

### Out of Scope — #1686 제안 2: `disable_builtin_bash_patterns` override 키

구조 체크 + extras 사본 제거가 필요를 소멸시킨다 (§A.6). 신규 config 표면·스키마 키·로더 변경 없음. 재검토 트리거: 구조 체크가 판정하지 못하는 새 오탐 클래스가 extras 경로로 유입될 때 — 그때도 정규식 override가 아니라 판정기 확장이 정답 방향이다.

### Out of Scope — 소비자 인지 heredoc folding (M3 문서화된 한계)

heredoc 본문이 **bare root 형태**(`rm -rf /`)를 담으면 구조 체크가 여전히 거부한다 (개행 분할이 본문 행을 명령 세그먼트로 심사). 실행 소비자(`bash <<EOF`)에게는 올바르고 데이터 소비자(`cat <<EOF`)에게는 과잉이다. 소비자 인지 folding은 실행-데이터 분류를 요구하고 실제 under-blocking 위험을 수반하므로, 운영자 결정으로 유예한다. **재검토 트리거(@MX:DEBT 스타일)**: `@MX:CEILING` — 데이터 소비 heredoc 본문의 보호 대상 언급은 거부됨(과잉, 안전 방향); `@MX:UPGRADE` — 데이터 소비 heredoc의 보호 대상 **예시 언급**까지 허용하는 소비자 인지 판정이 요구되고, 실행-데이터 분류의 under-blocking 위험이 별도 설계로 검토될 때. 카드의 데이터 언급 요구는 보고된 구체 형태(Write 도구 — 가드 대상 아님, heredoc 깊은 경로 예시 — M1으로 수리)에 대해서는 충족된다.

### Out of Scope — 셸 변수 간접 (`rm -rf $TARGET`)

텍스트 수준 가드는 "쓰인 대로의 명령"만 본다. `$TARGET`의 해석은 셸이 하므로 본 설계로 해결 불가능하다 — 기록하는 것이 정답이지 "수정"이 아니다 (M3와 함께 문서화).

### Out of Scope — extras 목록의 다른 내장 패턴 중복 (관측만)

`curl|sh`, `wget|sh` 등 extras 라인 중 일부는 내장 기본값과 중복으로 보인다. 본 SPEC은 관측으로만 기록하고 만지지 않는다 — 각 라인의 실제 중복 여부는 전용 측정의 소관이다.

### Out of Scope — t512 / GitHub #1659

worktree-guard heredoc brace folding은 다른 가드(`branch_guard.go` 인용부호 패턴)의 다른 결함이다. 본 SPEC은 `substituteQuotedArguments`의 동작을 변경하지 않는다.
