# t528 — AC 수집기 앵커 실측 (plan-phase 기준선)

- card: t528
- worktree: `.claude/worktrees/t528` · branch `WT-ac-collector-anchor`
- base: `52f863f36` (= `origin/develop` 측정 시점)
- 측정 일자: 2026-09-08
- 측정 도구: `internal/spec/zz_t528_probe_test.go` (임시 프로브, 커밋하지 않음)

---

## Claim

`spec.ParseAcceptanceCriteria` 는 AC 섹션 안에서 실제로 도달한 AC 선언 줄
1160 개 중 216 개(18.6%)만 수용하고 944 개를 거절한다. in-section 선언을
가진 채 결과가 0 인 `spec.md` 는 100 개다.

## Evidence

프로브를 코퍼스에 직접 태운 출력 (verbatim):

```
=== RUN   TestT528Probe
    zz_t528_probe_test.go:106: spec.md=806  with ##acceptance heading=360
    zz_t528_probe_test.go:107: AC declaration lines: in-section=1160  outside-section=180
    zz_t528_probe_test.go:108: in-section verdict: accepted-by-current-parser=216  rejected=944
    zz_t528_probe_test.go:109: files with >=1 in-section decl and ZERO accepted (fully blind) = 100
    zz_t528_probe_test.go:110: files with >=1 accepted (positive-needle control) = 18
    zz_t528_probe_test.go:126:   ID-SHAPE AC-N                         150
    zz_t528_probe_test.go:126:   ID-SHAPE AC-VNRN-RT-N-N               109
    zz_t528_probe_test.go:126:   ID-SHAPE AC-WFN-N                     85
    zz_t528_probe_test.go:126:   ID-SHAPE AC-SPC-N-N                   63
    zz_t528_probe_test.go:126:   ID-SHAPE AC-ORC-N-N                   56
    zz_t528_probe_test.go:126:   ID-SHAPE AC-EXTN-N                    53
    zz_t528_probe_test.go:126:   ID-SHAPE AC-MIGN-N                    42
    zz_t528_probe_test.go:126:   ID-SHAPE AC-CON-N-N                   40
    zz_t528_probe_test.go:126:   ID-SHAPE AC-HRN-N-N                   33
    zz_t528_probe_test.go:126:   ID-SHAPE AC-UTIL-N-N                  30
    zz_t528_probe_test.go:126:   SEP :                            759
    zz_t528_probe_test.go:126:   SEP (                            228
    zz_t528_probe_test.go:126:   SEP —                            91
    zz_t528_probe_test.go:126:   SEP «none/other»                 82
--- PASS: TestT528Probe (0.28s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/spec	0.830s
```

명령: `go test ./internal/spec/ -run TestT528Probe -v -count=1 -timeout 600s`

### 대조군 (양성 needle) — 「못 읽는다」와 「안 돈다」의 판별

수집기가 실제로 도는 것을 먼저 세웠다. 1차 프로브 출력 (verbatim):

```
    zz_t528_probe_test.go:44: spec.md files=806  parsed>0=18  parsed==0=788  totalRootAC=216
    zz_t528_probe_test.go:46: POSITIVE-NEEDLE: ../../.moai/specs/SPEC-ARTIFACT-STATELESS-001/spec.md -> AC-AST-001-01
    zz_t528_probe_test.go:46: POSITIVE-NEEDLE: ../../.moai/specs/SPEC-CLIFIX-CONCURRENCY-001/spec.md -> AC-CONC-001-001
    zz_t528_probe_test.go:46: POSITIVE-NEEDLE: ../../.moai/specs/SPEC-CLIFIX-CONTRACT-001/spec.md -> AC-CONT-001-001
```

수집기는 돈다. 0 은 미실행이 아니라 거절이다.

### 거절 원인 — 코퍼스에서 유도한 실제 줄 모양

```
- **AC-ASE-001** — **Given** the run-phase commit on `WT-achwd-strip-exempt`,
- AC-AUDIT-SNAPSHOT-001 (A1): sticky cache — past-24h unchanged-hash skip still fires.
- AC-AUTONOMY-TIERS-001 (REQ-001): `moai init` wizard offers 3-tier selection; …
```

현행 앵커 (`internal/spec/parser.go:218`):

```go
acIDPattern := regexp.MustCompile(`^(AC-[A-Z0-9]+-[0-9]+-[0-9]+(?:\.[a-z](?:\.[a-z]+)?)?)\s*:\s*`)
```

두 축이 동시에 막는다.

1. **ID 세그먼트 문법** — `AC-<X>-<숫자>-<숫자>` 만 받는다. 실사용 상위 형태는
   `AC-001`(150), 4 세그먼트(109), 알파 중간 2 세그먼트(85).
2. **구분자** — ID 직후 `:` 만 받는다. 실제로는 `(A1):` 형 한정어 228,
   엠대시 `—` 91, 기타 82.

## Baseline-attribution

- 트리: `.claude/worktrees/t528`, HEAD `52f863f36`, 편집 없음(프로브 파일만 추가, 미커밋)
- 코퍼스: 해당 트리의 `.moai/specs/**/spec.md` 806 개
- 소비자 조사 명령: `grep -rn --include='*.go' '\.Criteria' internal/spec/ internal/cli/`
  → 프로덕션 소비자 2 곳: `internal/spec/lint.go:915`(CoverageRule,
  advisory warning), `internal/cli/spec_view.go:72`(`moai spec view --acceptance`)

## Gaps

- **헤딩 축은 재지 않았다.** 섹션 밖 선언 줄 180 개는 세기만 했고, 그 줄들이
  진짜 AC 인지 산문인지 표본 확인하지 않았다. 이 카드의 범위 밖(운영자 판단
  2026-09-08: 항목 문법 축만).
- **acceptance.md 사이드는 이 카드 범위 밖이다.** `lint_coverage_sibling.go`
  가 이미 `ExtractRequirementMappings` 로 우회 처리했다.
- **수리 후 delta 는 아직 재지 않았다.** 넓힌 문법이 실제로 몇 줄을 회수하는지,
  그리고 CoverageRule finding 수가 어떻게 움직이는지는 run 단계 측정 대상이다.
- **t518 과의 파일 충돌 여부를 직접 재지 못했다.** 다른 워크트리를 가로질러
  git 조회를 할 수 없어, 리드에게 보고하고 판단을 넘겼다.

## Residual-risk

- 넓힌 ID 문법이 **AC 가 아닌 불릿**을 AC 로 읽을 수 있다. 섹션 스코핑
  (`##…acceptance` 헤딩 요구)이 유지되므로 위험은 그 섹션 안으로 한정되지만,
  0 은 아니다. 뮤턴트로 경계를 그려야 한다.
- `doc.Criteria` 가 커지면 `ValidateDepth` / `DuplicateAcceptanceID` 가
  새로 발화할 수 있다. 이는 코퍼스가 원래 갖고 있던 사실이 드러나는 것이지만,
  **착지 시 코퍼스를 붉게 만들 수 있다** — run 단계에서 반드시 계측한다.
- `buildTree` 의 들여쓰기 기반 계층화가 넓힌 줄들에서 의도치 않은 부모-자식
  관계를 만들 수 있다.

---

## 1 차 측정에서의 정정 (기록 보존)

1 차 분류는 「헤딩 없음 + AC 토큰 보유 = 37 파일」을 blind 로 셌다. 표본을
열어보니 대부분 **본문 산문 언급**이었다:

```
248:AC-BH-006's.
180:AC-BLKG-008(1)의 열거가 이 줄을 놓쳤던 것이 최종 감사 T1이며, …
```

needle 이 불릿을 요구하지 않아 산문을 선언으로 셌다. 불릿 필수로 좁혀
재측정한 것이 위 수치다. 이 정정은 간극을 **넓히는** 방향이므로 남긴다.
