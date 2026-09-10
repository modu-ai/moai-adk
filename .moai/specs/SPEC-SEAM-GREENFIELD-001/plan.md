---
id: SPEC-SEAM-GREENFIELD-001
title: "plan — seam greenfield 첫 저장 500 결함 수리 (t544)"
version: "0.1.0"
created: 2026-09-08
updated: 2026-09-08
author: GOOS
module: "internal/settings/yamlpatch"
tier: S
---

> 본 산출물은 `status:` 필드를 두지 않는다(SPEC 디렉터리의 상태축은 `spec.md` 단일 소관 — frontmatter 스키마 SSOT § Artifact Statelessness).

## §A Context

- **카드**: t544 (Class B — 결함, 원인 특정 완료). 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t544`, 브랜치 `WT-save-absent-file`, plan-phase 기준 HEAD `52f863f36`.
- **Tier**: S (artifact set = spec.md + plan.md; AC는 spec.md §3 inline). 결함 수리는 소수 LOC + 테스트, 5파일 미만.
- **산출물**: `.moai/specs/SPEC-SEAM-GREENFIELD-001/{spec.md, plan.md, progress.md}`.
- **결함 요약**: absent 섹션 파일의 첫 seam 저장이 `atomicWrite`의 무조건 `os.Stat`에서 죽는다 — 읽기 계층은 absent를 greenfield로 취급하는데 쓰기 계층만 실패한다. 근거 표와 판정 방향: spec.md §1.2-§1.3.
- **run-phase 방법론**: RED-GREEN-REFACTOR (tdd) — defect-fix with explicit reproduction. RED는 AC-001·AC-004의 two-cell로 강제된다.

## §B Known Issues (본 SPEC 도메인 관련만)

- **B-a (기대 전환 대상)**: `TestYAMLPatchAtomicWriteErrors/"stat missing target"`(`internal/settings/yamlpatch/yamlpatch_test.go:383-389`)가 absent 대상 오류를 기대값으로 인코딩 중 — 수리와 **함께** 재작성해야 한다. 재작성을 잊으면 수리 후 이 테스트가 RED로 남아 수리를 가린다.
- **B-b (부재 판정의 스캔 범위)**: plan-phase 초안이 "`TestPatchFileValueInvariantPreservesBytes`는 실재하지 않는다"고 판정한 것은 오판정이었다 — grep 범위가 yamlpatch 패키지에 한정돼 settings 패키지 외부 테스트를 보지 못했다. 실재 위치: `internal/settings/write_safety_test.go:29`. PatchFile 직접 통제군은 그 파일의 4건이다(`:29`·`:52`·`:320`·`:341` — spec.md AC-005). 부재 주장은 스캔 범위를 명시하고, 그 범위 밖의 패키지까지 확장해 재측정해야 한다.
- **B-c (vacuous-green)**: 웹 가드의 제출이 값-불변 no-op이면 C6 게이트(`internal/settings/sectionapply.go:56-80`)가 조용히 건너뛰어 수리 없이도 통과한다. 실변경 제출 형태만 유효하다(spec.md AC-004).
- **B-d (부재-가드 채택 규율)**: absence-guard AC는 RED-now만으로 채택될 수 없다 — 뮤턴트가 유일한 판별 증거이고, 못 잡은 뮤턴트도 기록해야 가드의 경계가 그려진다(spec.md REQ-8/AC-006; `.claude/rules/moai/development/verification-completeness.md` §2).
- **B-e (셸 grep)**: 이 트리의 셸 `grep`은 ugrep 래퍼로 조용히 건너뛴다 — 부재 주장 증거는 `/usr/bin/grep` 또는 `-a`로 채득한다.
- **B-f (CI 3-tier)**: spec-lint / golangci-lint / per-OS test는 각자 실패할 수 있다 — 기존 baseline과 NEW 결함을 구분해 보고한다.

## §C Pre-flight (run-phase 착수 시)

```bash
# 1. 브랜치 + HEAD 재확인 (디스패치 값과 다르면 정지·보고)
git rev-parse --short HEAD && git branch --show-current

# 2. 줄 앵커 재검증 — content-token 기준 (줄 인용은 낡는다)
/usr/bin/grep -n "stat %s: %w" internal/settings/yamlpatch/yamlpatch.go
/usr/bin/grep -n "stat missing target" internal/settings/yamlpatch/yamlpatch_test.go

# 3. 기존 baseline 측정 (수정 전 상태 — NEW vs 기존 구분용)
go test ./internal/settings/yamlpatch/... ./internal/settings/... ./internal/web/...

# 4. 타임디렉터리 격리 확인 — 테스트가 프로젝트 루트에 파일을 만들지 않는지
#    (신규 테스트는 t.TempDir() 안에서만 absent 트리를 구성한다)
```

## §D Constraints

### PRESERVE (수정 금지)

- `internal/settings/sectionapply.go` — C6 게이트의 no-op 스킵 의미론(REQ-WWS-003)은 그대로다. absent-허용은 쓰기 계층만 고친다.
- `internal/settings/sectionwrite.go`, `internal/settings/yamlpatch/yamlpatch.go`의 `PatchFile` 읽기 경로(C1의 greenfield 주석 포함) — 읽기 계층은 옳다.
- `internal/web/`의 렌더러 전반 — 카드 t545 소관, 본 SPEC 무접촉.
- 기존 테스트 중 `"read-only directory"` 서브테스트(`yamlpatch_test.go:391-413`)와 AC-005 통제군 — 무수정 GREEN.
- 라인-스플라이스 바이트 보존 불변(빈 줄·주석·키 순서·unknown key·인용 스타일·typed 스칼라).
- temp+rename 원자성 — absent 경로에서 직접 쓰기로의 전환 금지(REQ-4).

### 금지 사항

- `go test ./...` 로컬 전체 실행 금지 — 범위는 `./internal/settings/... ./internal/web/...`.
- `--no-verify`, force-push 금지. 커밋은 Conventional Commits + `🗿 MoAI` 트레일러 + 카드 id(t544).
- 새 의존성 추가 금지. 기본 모드 상수는 yamlpatch 패키지 내 단일 정의(하드코딩 규율상 패키지 로컬 const로 충분 — config 배관 불필요).

## §E Self-Verification (run-phase 완료 보고 항목)

E1 AC PASS/FAIL 매트릭스(spec.md §3 AC-001..006, 커맨드+실출력+트리 SHA 귀속) · E2 크로스플랫폼 빌드(`GOOS=windows GOARCH=amd64 go build ./...` — absent 경로 코드는 POSIX 권한 비트를 다루므로 cross-compile 확인 필수) · E3 범위 패키지 커버리지 · E5 lint(NEW vs baseline 구분) · E6 브랜치 HEAD + 미푸시 커밋 수 · E8 RED verbatim 채득(AC-001·AC-004 수리 전 실패 출력 — test-first 반증 가능성의 근거).

## §F Milestones

결정 가변성 순 — 가장 변할 가능성이 큰 결정(수리 형태)을 M2 앞부분에 두고, 기계적 재검증을 뒤로 보낸다.

### M1 — RED 가드 (yamlpatch 단위 + 웹 레벨)

1. yamlpatch 단위 greenfield 테스트 신설: absent 경로에 `PatchFile` 실변경 edit → 생성 + 0644 + 내용 반영 (AC-001). 수리 전 트리에서 RED 채득(커맨드 + exit code + 트리 SHA).
2. 웹 레벨 가드 신설: mcp.yaml이 없는 트리에서 POST /save 실변경 제출 → 200 + 파일 생성 (AC-004). 제출은 C6 게이트를 통과하는 실변경 형태(B-c). 수리 전 RED 채득 — 500 배너가 stat 오류를 가리켜 `PatchFile` 도달을 증명.
- 판정: 두 가드가 결함을 직접 가리키는 FAIL이어야 한다(틀린-이유 RED 금지).

### M2 — 수리 (atomicWrite absent-허용)

1. `atomicWrite`(`internal/settings/yamlpatch/yamlpatch.go:361-365`)의 무조건 stat-fail을 분기한다: `os.IsNotExist` → 문서화된 기본 모드 0644(패키지 단일 const)로 진행, 그 외 stat 오류는 기존 래핑으로 유지 (REQ-1..4).
2. `"stat missing target"` 서브테스트를 absent 생성 성공 + 0644 기대로 재작성 (AC-003, B-a).
3. M1의 두 가드가 PASS로 뒤집히는지 확인.

### M3 — 뮤턴트 + 통제 재확인

1. t517 F1-D 렌더-오버레이 방법(`SPEC-WEB-WRITE-SAFETY-001/progress.md:188-196`)을 커밋 트리 위 오버레이로 적용: 수리 되돌림 → FAIL verbatim 채득 → 복원 → PASS → `git diff --stat` 무출력 (AC-006).
2. 못 잡은 뮤턴트가 있으면 내용과 함께 progress.md에 기록 (REQ-8, B-d).
3. AC-005 통제군 + `"read-only directory"` 무손상 확인.

### M4 — 범위 한정 판정

1. `go test ./internal/settings/... ./internal/web/...` + `GOOS=windows GOARCH=amd64 go build ./...` + lint.
2. §E 자가검증 항목 채움 → 완료 보고(카드 id + 브랜치/HEAD + 미푸시 수 + 증거 경로).

## §G Anti-Patterns

- 수리 없이 기존 `"stat missing target"` 테스트만 고쳐 결함을 "수정된 것"으로 만들기 — 테스트 기대 전환은 수리와 동일 커밋 범위에서 함께 가며, 가드는 결함이 살아있을 때 RED였음이 채득돼야 한다.
- 값-불변 no-op 제출로 웹 가드 통과시키기 (vacuous green, B-c).
- absent 관용을 전체 stat 오류로 넓히기 (REQ-3 위반).
- `/usr/bin/grep` 대신 ugrep 래퍼 `grep`으로 부재 주장 채득 (B-e).

## §H Cross-References

- spec.md: `.moai/specs/SPEC-SEAM-GREENFIELD-001/spec.md` (REQ-1..8, AC-001..006, §4 설계 결정, §8 계열 관측)
- 방법론 원천: `SPEC-WEB-WRITE-SAFETY-001/progress.md:188-196` (렌더-오버레이 뮤턴트 절차), `SPEC-WEB-WRITE-SAFETY-001/spec.md` (REQ-WWS-003 — absent bool 제출 의미론)
- 규율: `.claude/rules/moai/development/verification-completeness.md` §2 (two-cell + 뮤턴트 프로브), `verification-claim-integrity.md` §2 (baseline 귀속)
- 인접 카드: t545 (renderer 이름-유일성 — 다른 축, 범위 중복 없음; spec.md §5/§7)
