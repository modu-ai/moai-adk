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

### AC-4 — 측정 불가는 absent로 남는다 (fresh 아님, stale 아님)

**Given** 스탬프된 sha의 이력 안에 codemaps 본문을 건드린 커밋이 하나도 없는 픽스처(얕은 이력 또는 본문이
초기 커밋 이후 한 번도 변경되지 않은 저장소)
**When** codemaps 층을 판정하면
**Then** `verdict == "absent"` 이고 반환된 시스템 오류가 non-nil이다.

- `fresh`도 `stale`도 아니어야 한다. 두 값 중 어느 쪽이라도 이 AC는 실패다.
- 규칙 A의 합집합 정정 이후 이 경로는 사실상 도달하기 어렵다(`spec.md` §D.1). 픽스처는 본문이 S 시점
  본문과 추적·미추적 양쪽에서 동일하면서 S의 이력에 본문 커밋이 없는 상태를 만들어야 하며, 얕은 이력으로
  본문 커밋을 잘라내는 형태가 그 예다. 도달이 어렵다는 사실이 이 AC를 면제하지는 않는다 — 도달했을 때
  절대 fresh가 아니어야 한다는 것이 요지다.

### AC-5 — 회귀 잠금 (AC-7이 이 항목의 구조적 전제를 잠근다)

**Given** 기존 `internal/graph` 테스트 스위트
**When** `go test ./internal/graph/...`를 실행하면
**Then** 모두 통과한다.

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
