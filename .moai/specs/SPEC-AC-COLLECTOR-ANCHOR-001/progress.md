---
id: SPEC-AC-COLLECTOR-ANCHOR-001
title: "인라인 AC 수집기 항목 문법 앵커 — 진행 기록"
status: in-progress
created: 2026-09-08
updated: 2026-09-08
author: manager-develop
card: t528
---

# SPEC-AC-COLLECTOR-ANCHOR-001 — 진행 기록

카드 `t528`. 워크트리 `.claude/worktrees/t528`, 브랜치 `WT-ac-collector-anchor`.
착수 트리 `52f863f36`.

모든 수치는 REQ-ACA-001-016의 세 요소 — **트리 SHA · 판별식 · 분모 파일 목록** — 를 함께 인용한다.

---

## §E.1 마일스톤 진행

| M | 내용 | 상태 |
|---|---|---|
| M0 | 대조군·측정 도구 고정 (파서 편집 이전) | 완료 |
| M1 | 네 축 앵커 구현 | 진행 |
| M2 | 코퍼스 재측정 | 대기 |
| M3 | before-image 대조 | 대기 |
| M4 | 뮤턴트 경계 | 대기 |
| M5 | 섹션 스코핑 불변 픽스처 | 대기 |
| M6 | 과수용 실측 | 대기 |
| M7 | CLI 렌더 + 하드 에러 게이트 | 대기 |
| M8 | CoverageRule delta + 바이트 동일성 | 대기 |

---

## §E.2 Run-phase Evidence

### M0 — 대조군과 측정 도구를 파서 편집 **이전에** 고정했다

**(a) 프로브 승격 (REQ-ACA-001-014).**
`.moai/reports/t528/probe/ac_anchor_probe_test.go` → `internal/spec/zz_t528_anchor_probe_test.go`
(커밋된 테스트). 판별식 `declRe`(판별식 B)는 산출물 사본과 **byte-identical**로 옮겼다.

승격판이 원본과 다른 점은 둘이고, 둘 다 의도적이다.

1. **출력 디렉터리가 `T528_PROBE_OUT`로 분기한다** (기본 `probe/out`). 원본은
   `probe/positive-needle.txt` · `probe/filelist.txt`를 **덮어쓴다** — 넓힘 이후에 한 번만
   돌아도 핀된 before-image가 사라지고, 그러면 무회귀 판정이 넓힌 파서와 자기 자신의 비교가
   된다(AC-ACA-001-002가 금지한 바로 그 모양).
2. **수용 열이 둘이다.** `baselineRe`는 넓힘 **이전** 앵커의 동결 사본이고, live 열은 실제
   `parseSingleACLine`을 호출한다. 원본은 정규식 사본 하나만 태워서, 파서를 넓혀도 같은 216을
   다시 내는 **공허한 재측정**이 된다. 두 열을 한 프로세스에서 함께 내면 기준선과 재측정이
   같은 실행·같은 트리·같은 분모 위에 서고, 비교가 두 실행을 가로지르지 않는다.

**(b) 승격판이 플랜 단계 기준선을 그대로 재유도한다.**

명령 · verbatim 출력: `.moai/reports/t528/probe/m0-before-run.txt` (`EXIT=0`)

```
$ T528_PROBE_OUT=../../.moai/reports/t528/probe/before \
  go test ./internal/spec/ -run TestT528Anchor -v -count=1 -timeout 600s

    DENOMINATOR spec.md read = 807 (list: .../before/filelist.txt)
    IN-SECTION declarations = 1167
      accepted by FROZEN baseline anchor = 216
      accepted by LIVE parseSingleACLine = 216
      rejected by LIVE parseSingleACLine = 951
      NEWLY accepted (live yes, baseline no) = 0
    POSITIVE-NEEDLE files (live) = 18
    BASELINE-NEEDLE files (frozen anchor) = 18
    FULLY-BLIND files (>=1 decl, 0 accepted by live) = 101
    B.1 numeric-tail candidate: covered = 1128  UNCOVERED = 39
      SEP :  800 · SEP (  232 · SEP —  93 · SEP «none/other»  42
--- PASS: TestT528Anchor (0.18s)
EXIT=0
```

- **트리 SHA**: `52f863f36` (`internal/spec/parser.go` 미편집)
- **판별식**: B — `zz_t528_anchor_probe_test.go`의 `declRe`
- **분모 목록**: `probe/before/filelist.txt` (807)

플랜 단계 산출물과 대조: `positive-needle.txt` · `filelist.txt` · `blind-files.txt`
세 파일 모두 `diff -q` **무출력**(IDENTICAL). 승격이 판별식을 옮기지 않았다는 증거다.

**live == frozen == 216, newly-accepted == 0**이 승격판 live 열의 등가성 증명이다. 이 등가가
없으면 M2의 재측정 값은 216과 비교할 수 없다.

**비공허성 (AC-ACA-001-016)**: `[no tests to run]` / `[no test files]` 토큰 **부재**,
`--- PASS: TestT528Anchor` 존재, `EXIT=0`. 셋 다 위 출력에 있다.

**(c) before-image 산출물** — 전부 `.moai/reports/t528/probe/before/`, 파서 편집 이전 커밋.

| 산출물 | 내용 |
|---|---|
| `filelist.txt` | 분모 807 |
| `positive-needle.txt` | 양성 대조군 18 (live) |
| `baseline-needle.txt` | 동결 앵커 기준 18 |
| `blind-files.txt` | 전맹 101 |
| `rootac.txt` | **파일별 루트 AC 트리 모양** 807행 — M3 대조 기준. `ID(child,child)` 형식이라 루트 집합과 트리 모양을 함께 나른다 |
| `reqmap.txt` | 파일별 AC → REQ 매핑 216행 — AC-ACA-001-008의 매핑 불변 대조 기준 |
| `still-rejected.txt` | 거절 951행 (경로:줄번호:원문) |
| `newly-accepted.txt` | 0행 (정의상 비어 있음) |
| `specview-errors.txt` | CLI 스윕 원출력 448행 |
| `specview-before-20260908.md` | CLI 하드 에러 before-image + 정정 2건 |
| `specview-sweep.sh` | 스윕 스크립트 (재유도용) |

### [HARD] M0에서 나온 정정 2건 — 문서의 전제가 실측과 어긋났다

정본은 `.moai/reports/t528/probe/before/specview-before-20260908.md`. 요지만 옮긴다.

**정정 1 — `--acceptance` 플래그는 존재하지 않는다.**
`spec.md` §1.3 · `plan.md` §E M7 · `acceptance.md` §D.12/§D.15가 전부
`moai spec view <ID> --acceptance`를 명령으로 적는다. `internal/cli/spec_view.go:37`이
등록하는 플래그는 `--shape-trace` 하나뿐이고, acceptance 뷰가 곧 `moai spec view <ID>`다.
적힌 명령으로 코퍼스를 훑으면 `SWEPT=807 NONZERO_EXIT=807` — 전부 `Unknown flag`이며
수집기에 대해 아무 말도 하지 않는다. **이 카드가 다섯 번 당한 「오형성 판별식이 낸 자신
있는 수」와 같은 모양이라 기록으로 남긴다.** 이후 모든 측정은 실제 명령을 쓴다.

**정정 2 — CLI 하드 에러 before-image는 0이 아니라 442다.**

```
SWEPT=807  NONZERO_EXIT=448
  parse error:            442   전부 "acceptance criteria section not found"
  spec.md not found for:    6   _archive/<ID>/ 경로 — 파서에 도달조차 하지 않음
  addressable SPECs:      801
```

`plan.md` §E M7과 `acceptance.md` §D.15는 before-image를 **0**으로 핀하며
`probe/duplicate-20260908.txt`를 인용한다. 그 산출물이 재는 것은 **중복 ID 재료**(진짜로 0)
이지 CLI 하드 에러가 아니다 — 서로 다른 두 양이고 0이 한쪽에서 다른 쪽으로 옮겨졌다.
다만 `acceptance.md` §D.15의 **Gap** 절이 「실제 `moai spec view`를 코퍼스 전체에 돌린 적은
아직 없다」고 스스로 적어 두었으므로, 이 측정은 그 간극을 **닫는** 것이지 문서를 뒤집는
것이 아니다.

442 전부가 `acceptance criteria section not found` — `findACSectionStart`가 -1을 돌려줄 때
`internal/cli/spec_view.go:85`의 `default:`가 치명으로 만드는 **헤딩 축**이며, 이 카드의
범위 밖(`spec.md` §4)이고 항목 문법 넓힘이 닿지 않는다.
**`DuplicateAcceptanceID` / 깊이 초과 부류는 0건**이다 — 넓힘이 새로 만들 수 있는 유일한
부류가 이것이고, 그래서 이 0이 실제 게이트의 기준선이다.

따라서 판정 게이트는 REQ-ACA-001-013이 실제로 적은 것 — **기준선에 없던 하드 에러 == 0** —
이며 442에 대한 delta로 잰다. 절대값 0은 이 코퍼스에서 달성 가능한 수가 아니었다.

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: null
run_commit_sha: null
run_status: in-progress
ac_pass_count: null
ac_fail_count: null
```
