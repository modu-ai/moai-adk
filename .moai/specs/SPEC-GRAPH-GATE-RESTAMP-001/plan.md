# SPEC-GRAPH-GATE-RESTAMP-001 — 구현 계획

> 절 순서는 **번복 가능성이 높은 결정부터**다. 위쪽 절이 리뷰가 가장 필요한 곳이고,
> 아래쪽은 기계적인 작업이다.

## §A. 맥락

`moai graph check`의 codemaps 층이 스탬프된 커밋을 측정 기점으로 삼는 탓에, 본문을 안 고치고 스탬프만
다시 찍으면 게이트가 초록이 된다. 기점을 "본문이 마지막으로 실제 변경된 지점"으로 옮겨 그 경로를 닫는다.
문제 진술과 실측은 `spec.md` §A / §B.

## §B. 결정 완료 항목 (착수를 막는 미해소 항목 없음)

두 건 모두 레인이 2026-09-04에 결정했다. 운영자 판단 대상이 아니며 재개하지 않는다.

- **규칙 A 토큰 명명 — 종결.** 프로브가 실제로 재는 것은 "미커밋 재생성"이 아니라 "작업 트리가 스탬프와
  다름"이다(`spec.md` §B.3, §D.2). 로직은 배차문 그대로 두고 토큰만 `working-tree-differs-from-stamp`로
  개명한다. 로직 변경 0.
- **규칙 C의 C1/C2 분리 — 종결.** 본문-부재는 선행 `codemaps directory missing` 분기(`os.Stat(dir).IsDir()`)를
  통과하므로 도달 가능하며, 시스템 오류로 처분하면 기존 테스트 2건이 깨진다(`spec.md` §B.5). C1은 absent +
  오류 nil(exit 1), C2는 absent + 오류(exit 2). C1은 형제 층 `checkCitations`와의 정합이다.
- **규칙 A 프로브 = 합집합 — 종결.** 기존 픽스처의 codemaps 본문은 미추적이라(`spec.md` §B.4) `git diff`
  단독 프로브로는 규칙 A가 발화하지 않고 규칙 C로 떨어져 미추적-본문 픽스처가 깨진다. 프로브를
  `git diff` ∪ `git ls-files --others --exclude-standard`로 정정하며, 이는 `gitDiffNameList`가 이미 쓰는
  관용구의 재사용이다. 규칙 B/C의 로직은 불변.

- `spec.md` §G Q1(임계값 40)은 이 SPEC의 구현 범위 밖이며, 착수를 막지 않는다. 기록용이다.

## §C. 기술 접근

### 데이터 구조 변경 (가장 번복 가능성 높음)

`internal/graph/check.go`의 `LayerReport`에 보고 전용 필드 2개를 추가한다.

```go
// ContentAnchor is the sha the described-source diff was measured from.
ContentAnchor string `json:"content_anchor,omitempty"`
// ContentAnchorSource names how ContentAnchor was resolved.
ContentAnchorSource string `json:"content_anchor_source,omitempty"`
```

두 필드는 **게이팅에 쓰이지 않는다**. 기존 `Contribution` / `DrivingPaths`가 그렇듯 보고 전용이며,
JSON 소비자에게 새 키가 추가되는 것 외의 계약 변화는 없다(`omitempty`이므로 dirty 경로 출력은 바이트
동일).

### 앵커 해석 함수 (신규)

`checkCodemaps` 안의 clean 경로에서만 호출되는 헬퍼를 하나 새로 만든다:

```go
func resolveContentAnchor(projectRoot, stampedSHA string) (anchor, source string, err error)
```

- 본문 pathspec은 상수로 뽑는다: `.moai/project/codemaps/` +
  `:(exclude).moai/project/codemaps/provenance.json`.
- 규칙 A: `git diff --name-only <S> -- <본문 pathspec>`와
  `git ls-files --others --exclude-standard -- <본문 pathspec>`의 **합집합** → 비어 있지 않으면
  `(S, "working-tree-differs-from-stamp")`. 합집합 형태는 `gitDiffNameList`의 관용구 재사용이며, 추적
  쪽만 거르면 미추적 본문이 세어지지 않아 미추적-본문 픽스처가 규칙 C로 떨어진다(`spec.md` §B.4).
  본문-부재 픽스처는 합집합으로도 구제되지 않으며 C1이 그 몫이다.
- 규칙 B: `git log -1 --format=%H <S> -- <본문 pathspec>` → 비어 있지 않으면 `(그 sha, "last-body-change")`.
- 규칙 C: 둘 다 비면 **본문 집합의 유무로 갈린다**.
  - **C1 (본문 부재)**: `verdict=absent` + 본문 부재 reason, **오류 nil** (exit 1). 형제 층
    `checkCitations`의 `docs == 0` 처분과 동일(`spec.md` §D.1 선례).
  - **C2 (본문 존재, 앵커 미해석)**: `verdict=absent` + 시스템 오류 (exit 2). 배차 원안 그대로.
  - 본문 집합의 유무는 규칙 A가 이미 계산한 합집합과 S 시점 본문 목록으로 판정한다 — 세 번째 git 호출
    관용구를 새로 만들지 않는다.

기존 `gitOutput` 헬퍼를 재사용한다(신규 프로세스 실행 규약을 만들지 않는다).

### `checkCodemaps` 호출부 변경

clean 경로에서 `pv.CommitSHA` 유효성 확인 직후, `gitDiffNameList` 호출 **전에** 앵커를 해석하고,
`gitDiffNameList`에 넘기는 커밋 인자를 `pv.CommitSHA`에서 해석된 앵커로 바꾼다. 오류 시 기존
not-comparable 처분과 같은 모양으로 `verdict=absent` + 시스템 오류를 반환한다.

`pv.Dirty` 분기, mx-index / edges / citations 층은 손대지 않는다.

## §D. 제약

- 변경 파일은 `internal/graph/check.go` 하나 + `internal/graph/` 아래 신규 테스트 파일.
- `internal/cli/graph_stamp.go`, `internal/mx/provenance.go`, `.github/workflows/graph-freshness.yml`는
  건드리지 않는다.
- 지표 토큰 `described-source-diff` 유지, 임계값 40 유지.
- 코드/주석/식별자는 영문. 커밋 메시지는 영문이며 `t478`을 명시한다.

## §E. 자체 검증

- 픽스처 git 저장소 기반 테이블 테스트로 AC-1 ~ AC-4를 덮는다(`t.TempDir()` 안에서 `git init`).
- AC-1은 **뮤턴트 방향**이다: 수리 전 코드로 픽스처를 돌리면 fresh가 나와야 하고(RED 확립),
  수리 후 stale이 나와야 한다. RED를 먼저 관측하지 않은 초록은 공허하다.
- 기대 신호(verdict + 값의 임계값 대비 위치)를 **재기 전에** 테스트에 못박는다.
- `moai graph check`를 프로세스로 검증할 때 종료코드는 파이프 없이 직접 읽는다(AC-6).
- 검증 범위: `go test ./internal/graph/...`. 전체 스위트는 CI 몫이다.

## §F. 마일스톤

**M1 (Priority High) — 앵커 해석과 보고 필드**
`LayerReport` 필드 2개 추가, `resolveContentAnchor` 구현, `checkCodemaps` clean 경로 배선.
규칙 C는 C1/C2로 갈라 배선한다 — C1은 absent + 오류 nil, C2만 기존 not-comparable 처분에 합류시킨다.

**M2 (Priority High) — 뮤턴트 테스트**
AC-1(맨손 재스탬프 → stale)과 AC-2(진짜 재생성 미커밋/커밋 → fresh)를 픽스처 저장소로 덮는다.
AC-1은 수리 전 RED 관측을 증거로 남긴다.

**M3 (Priority Medium) — 조상성·불변식·회귀 잠금**
AC-3(앵커 조상성), AC-4(C1 → absent + 오류 nil, 그리고 "어떤 앵커 해석 결과도 fresh가 아니다" 불변식;
C2는 도달 불가로 선언하고 픽스처를 만들지 않는다), AC-5(기존 테스트 무회귀 + 허가된 픽스처 편집 1건),
AC-7(미추적 본문 픽스처 → 규칙 A, absent 아님).

**M4 (Priority Low) — 증거 정리**
`.moai/reports/t478/`에 검증 산출물을 반출하고 진행 기록을 닫는다.

순서 근거: M1이 없으면 M2가 걸 대상이 없고, M2의 RED 관측은 M1 착지 **직전** 시점에만 가능하므로
M1 → M2를 뒤집을 수 없다. M3는 M1의 오류 경로에 의존한다. M4는 순수 기계 작업이라 맨 뒤다.

## §G. 안티패턴

- **행복 경로만 증명하기.** AC-2만 통과시키는 수리는 위조-초록을 그대로 둔 채 초록이 된다. AC-1이 본질이다.
- **규칙 A 프로브를 추적 차이만으로 구현하기.** 미추적 본문이 세어지지 않아 미추적-본문 픽스처가 규칙 C로
  떨어진다. AC-7이 이 뮤턴트를 잡는다.
- **본문-부재를 시스템 오류로 처분하기.** 확정된 관측을 실패한 측정으로 보고하는 것이며, exit 1을 단언하는
  기존 테스트를 exit 2로 깨뜨린다(`spec.md` §B.5 실측). AC-4가 이 뮤턴트를 잡는다.
- **C2에 도달하는 척하는 픽스처 만들기.** 도달 불가 분기의 초록은 공허한 초록이다. C2는 선언하고 남긴다.
- **파이프로 종료코드 읽기.** `moai graph check | tail`은 `tail`의 rc 0을 잡아 초록으로 읽힌다 —
  다른 레인에서 실측된 오독이다.
- **앵커 walk를 S 이력 밖으로 넓히기.** 조상성 보장이 깨지고, 값이 더 초록인 쪽으로 움직일 수 있다.
- **지표 토큰 교체.** 인용면(CI 워크플로, 선행 SPEC)이 일제히 스테일해진다.
- **`provenance.json` 편집.** 이 카드는 그 파일을 만지지 않는다.

## §H. 교차 참조

- `spec.md` §D.1 (앵커 해석 절차), §D.2 (규칙 A 라벨 문제), §G (운영자 결정)
- `acceptance.md` (AC-1 ~ AC-6)
- `internal/graph/check.go` — `checkCodemaps`, `gitDiffNameList`, `LayerReport`, `gitOutput`
