# SPEC-GRAPH-GATE-RESTAMP-001 — 인수 기준

## §A. 검증 방법 제약 (모든 AC에 공통으로 구속)

- **AC-6 (종료코드 직접 판독).** `moai graph check`를 프로세스로 검증하는 모든 자리에서 종료코드는
  **파이프 없이** 기록한다. `moai graph check | tail`은 `tail`의 rc 0을 잡아 붉은 게이트를 초록으로
  읽는다 — 다른 레인에서 실측된 오독이다.
  - 금지: `./bin/moai graph check | tail -5`
  - 허용: `./bin/moai graph check > out.txt 2>&1; echo "RC=$?"; tail -5 out.txt`
  - Go 테스트 안에서는 `exec.Command(...).Run()`의 `*exec.ExitError`를 직접 판독한다.
- **기대 신호 선고정.** 각 AC는 값을 재기 **전에** 기대 verdict와 임계값 대비 위치를 테스트 코드에
  못박는다. 측정 후에 기대치를 맞추는 것은 검증이 아니다.
- **픽스처 격리.** 모든 픽스처 git 저장소는 `t.TempDir()` 안에 만든다.

## §B. 인수 기준

### AC-1 — 맨손 재스탬프는 RED여야 한다 (뮤턴트 방향, 이 SPEC의 본질)

**Given** codemaps 본문(`.moai/project/codemaps/*.md`)이 오래되어 described roots와 크게 어긋난 픽스처
저장소가 있고, 그 본문이 마지막으로 변경된 이후 described roots 아래 파일이 임계값 이상 변경되어 커밋되어
있으며, codemaps 본문은 **한 글자도 고치지 않은** 상태에서
**When** `graph stamp codemaps`가 HEAD에서 다시 실행되어 `provenance.json`의 `commit_sha`가 HEAD로
갱신되고, 이어서 codemaps 층을 판정하면
**Then** 그 층의 `verdict`는 `stale`이고 `value`는 `threshold` 이상이다.

- 기대 신호는 측정 전에 고정한다: `verdict == "stale"` **그리고** `value >= threshold`.
- 수리 전 코드에서 동일 픽스처가 `verdict == "fresh"`를 내는 것을 먼저 관측해 RED를 확립한다. RED 미관측
  상태의 초록은 이 AC를 만족시키지 못한다.
- 보고된 `content_anchor_source`는 `last-body-change`다.

### AC-2 — 진짜 재생성은 GREEN이어야 한다

**Given** 동일한 픽스처 저장소에서 codemaps 본문이 실제로 다시 쓰인 상태이고
**When** codemaps 층을 판정하면
**Then** 그 층의 `verdict`는 `fresh`다.

두 하위 경우를 모두 덮는다.

- **AC-2a (미커밋 재생성 — 앵커 규칙 A).** 본문이 작업 트리에서만 다시 쓰이고 커밋되지 않은 상태.
  기대: `verdict == "fresh"`, `content_anchor`는 스탬프된 sha와 동일.
- **AC-2b (커밋된 재생성 — 앵커 규칙 B).** 본문이 다시 쓰이고 커밋된 뒤 그 지점에서 스탬프된 상태.
  기대: `verdict == "fresh"`, `content_anchor`는 그 재생성 커밋, `content_anchor_source`는
  `last-body-change`.

### AC-3 — 규칙 B의 앵커는 스탬프 sha의 조상이다

**Given** 앵커가 규칙 B로 해석된 임의의 픽스처
**When** 해석된 `content_anchor`와 스탬프된 sha를 비교하면
**Then** `git merge-base --is-ancestor <content_anchor> <stamped_sha>`가 rc 0으로 성립한다.

- 이 성질이 §B.2의 보수성 보장(값이 더 초록인 쪽으로 움직이지 않음)의 기계적 근거다.
- 실측 대조군(이 워크트리, 2026-09-04): 앵커 `2f28bc394`, 스탬프 `ad272be20`, 조상성 rc 0,
  described-root diff 405 → 570.

### AC-4 — 앵커 해석 결과는 결코 `fresh`가 아니다 (C1이 도달 가능한 증인)

**Given** codemaps 디렉터리에 `provenance.json`만 있고 본문 문서가 하나도 없는 픽스처 — 기존
`internal/graph/check_test.go` `writeCodemapsProvenanceBlock`(안쪽 헬퍼)이 직접 만드는 바로 그 모양
(`spec.md` §B.4)
**When** codemaps 층을 판정하면
**Then** `verdict == "absent"` 이고 반환된 시스템 오류가 **nil**이며, verdict는 특히 `fresh`가 **아니다**.

- 기대 신호는 측정 전에 고정한다: `verdict == "absent"` **그리고** `err == nil` **그리고**
  `verdict != "fresh"`. 이것이 C1이며 exit 1이다.
- reason은 본문 부재를 명시해야 한다.
- 선례: 형제 층 `checkCitations`의 `docs == 0` 분기가 absent + 오류 없음으로 처분한다
  (`internal/graph/check_citations.go`). C1은 그 처분과의 정합이지 새 발명이 아니다.

**이 AC가 잠그는 성질은 C1 하나가 아니라 불변식이다: 어떤 앵커 해석 결과도 결코 `fresh`를 낳지 않는다.**
C1은 그 불변식의 도달 가능한 증인이고, C2는 같은 불변식의 도달 불가능한 갈래다.

**C2 분기는 픽스처로 덮지 않는다 — 의도된 미검증이다.** manager-develop이 측정한 바에 따르면 얕은 경계
커밋과 루트 커밋 모두 모든 파일을 ADDED로 보고하므로, 본문이 존재하는 한
`git log -1 <S> -- <본문 pathspec>`은 항상 비어 있지 않다. 즉 C2에 도달하는 git 상태가 없다. 그럼에도
C2 분기는 **fail-closed로 의도적으로 남긴다**: 도달했을 때 absent + 시스템 오류여야 하며 결코 fresh가
아니어야 한다(REQ-GGR-006a, REQ-GGR-007).

- 도달 불가 분기를 **선언하고 근거를 남긴 미검증**은 정직하다. 언급 없이 남긴 미검증은 정직하지 않다.
- **C2에 도달하는 척하는 픽스처를 만들지 않는다.** 도달하지 못하는 픽스처의 초록은 공허한 초록이다.

### AC-5 — 회귀 잠금 (AC-7이 이 항목의 구조적 전제를 잠근다)

**Given** 기존 `internal/graph` 테스트 스위트
**When** `go test ./internal/graph/...`를 실행하면
**Then** 모두 통과한다.

**허가된 픽스처 편집 — 정확히 1건.**

- `TestGraphCheckCmd_AbsentExitsOne`은 C1 하에서 **무수정으로 통과한다**. codemaps가 mx-index / edges와
  함께 absent가 되고 exit는 1로 유지되며, 이 테스트가 grep하는 reason 문자열은 그 두 층에서 계속 나온다.
- `TestCheckFreshness_DescribedRootsScopeFidelity`는 여전히 실패한다. 본문 없는 안쪽 헬퍼를 쓰면서
  codemaps `fresh`/value 0을 단언하기 때문이다. 이 테스트의 주제는 **described-roots 범위 충실성**이고
  본문과 무관하다 — 본문 없는 모양은 부수적이다. 따라서 이 픽스처에 codemaps 본문 문서 하나를 준다.
  모양은 `writeCodemapsProvenance`가 이미 쓰는 것과 동일하다(미추적 `modules.md`로 충분하며, 규칙 A를
  타고 S에 앵커돼 기존 단언이 그대로 유지된다).
- **그 테스트의 단언, 이름, 뮤턴트 판별식은 바꾸지 않는다.** 다른 어떤 기존 테스트도 수정하지 않는다 —
  이 1건 외에 AC-5의 회귀 잠금은 원문대로 유효하다.

추가로 다음이 변경되지 않았음을 확인한다.

- `pv.Dirty` fingerprint 경로의 동작(지표 토큰 `generation-fingerprint-mismatch`, threshold 1).
- mx-index / edges / citations 층의 동작.
- 지표 토큰 `described-source-diff` 문자열 자체.
- 임계값 기본값 40.

### AC-6 — 종료코드 판독 방식 (§A에 정의, 검증-방법 제약)

모든 `moai graph check` 검증 기록이 파이프 없는 종료코드 판독을 사용한다. 파이프를 거친 rc를 근거로 삼은
기록이 하나라도 있으면 실패다.

### AC-7 — 미추적 본문 픽스처는 규칙 A로 해석되어 판정 가능해야 한다 (회귀 잠금)

**Given** codemaps 본문이 작업 트리에 존재하지만 **한 번도 커밋된 적 없는** 픽스처 — 기존
`internal/graph/check_test.go` `writeCodemapsProvenance`가 만드는 바로 그 모양(base 커밋 이후
`modules.md`를 디스크에 쓰고 커밋하지 않음, `spec.md` §B.4)
**When** codemaps 층을 판정하면
**Then** 앵커는 규칙 A로 해석되어 `content_anchor`가 스탬프된 sha와 같고,
`content_anchor_source == "working-tree-differs-from-stamp"`이며, verdict는 **판정 가능한 값**
(`fresh` 또는 `stale`)이다 — `absent`가 **아니다**.

- 기대 신호는 측정 전에 고정한다: `verdict != "absent"` **그리고** 반환된 시스템 오류가 nil.
- 이 AC는 `spec.md` §B.4가 기록한 구멍의 회귀 잠금이다. 추적 차이만 보는 프로브(합집합이 아닌 diff 단독)로
  구현하면 이 픽스처는 규칙 A·B를 모두 통과하지 못하고 규칙 C(absent)로 떨어진다. **이 AC가 없으면 AC-5는
  테스트가 아니라 약속에 그친다** — 기존 픽스처가 전부 absent로 깨지는 상황을 잡아 줄 것이 없다.
- 뮤턴트 방향: 규칙 A의 프로브에서 `ls-files --others` 항을 제거하면 이 AC는 RED가 되어야 한다.

## §C. 엣지 케이스

- `provenance.json`이 없거나 파싱 불가 → 기존 absent 처분 유지(이 SPEC이 바꾸지 않는다).
- `pv.Dirty`가 참 → 앵커 해석을 수행하지 않는다. `content_anchor` / `content_anchor_source`는 비어 있고
  `omitempty`로 JSON에서 빠진다.
- clean 스탬프인데 `pv.CommitSHA`가 빈 문자열 → 기존 absent 처분 유지.
- described root가 유효성 검사에 실패 → 기존 absent 처분 유지.

## §D. 품질 게이트

- `go vet ./internal/graph/...` 무경고.
- `golangci-lint run` 신규 지적 0.
- `go test ./internal/graph/...` 통과, 신규 코드 경로가 테스트로 덮임.
- 변경 파일이 `internal/graph/check.go` + `internal/graph/` 신규 테스트로 한정됨
  (`git diff --name-only`로 확인).

## §E. Definition of Done

- [ ] AC-1 ~ AC-7 전부 통과, 각 항목의 명령과 출력이 증거로 남음
- [ ] AC-1의 수리 전 RED 관측이 기록됨
- [ ] `.moai/reports/t478/`에 검증 산출물 반출
- [ ] 커밋 메시지가 `t478`을 명시
- [ ] `provenance.json` 미변경 (`git status --porcelain .moai/project/codemaps/`가 깨끗)
