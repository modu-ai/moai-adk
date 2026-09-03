# t452 — AC-CSL-012 수정 조문의 뮤턴트 검증 (오케스트레이터 독립 측정)

조문을 고친 주체(manager-spec)와 판정하는 주체를 분리하기 위해, 아래는 오케스트레이터가
**직접 실행해 관측한** 값이다. manager-spec 의 자체 보고와 독립이다.

트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t452`
HEAD: `76afb6dea` · baseline: `c529b2e4aaf5148aee7e6c67649bf392837bbb06`

## 판정 명령 (수정된 조문 문면 그대로)

```
$ git diff --name-only "$BASELINE_SHA" -- internal/template/templates/ | wc -l
$ git diff "$BASELINE_SHA" -- internal/template/templates/ | grep '^+' | grep -v '^+++' \
    | grep -cE 'SPEC-|[0-9]{4}-[0-9]{2}-[0-9]{2}|[0-9a-f]{40}|/Users/'
```

## 방향 A — 현 트리는 초록이어야 한다

```
스윕한 파일 수 : 1     (조문의 "1 이상" 충족 — 빈 스윕이 아니다)
매치           : 0     (통과)
```

자리표시자 `SPEC: {SPEC-ID}` 는 변경되지 않은 줄에 있으므로 이제 스윕 대상 밖이다.

## 방향 B — 진짜 금지 토큰을 diff 로 넣으면 적색이어야 한다

manager-spec 의 뮤턴트는 `SPEC-` **한 종류만** 시험했다(그쪽 Gap 에 스스로 적었다).
그 구멍을 메우려고 금지 토큰 **4 종을 각각 한 줄씩** 넣었다 — 각 셀렉터가 살아 있으면
줄 단위 매치가 정확히 4 가 된다.

투입한 4 줄:

```
# SPEC-MUTANT-PROBE-001
# 2026-09-03
# 0123456789abcdef0123456789abcdef01234567
# /Users/goos/somewhere
```

```
매치 : 4     (적색 — 4 종 셀렉터가 모두 살아 있다)
```

한 종류라도 죽어 있었다면 4 미만이 나왔을 것이다. 좁히기가 과했다면 0 이 나왔을 것이다.
둘 다 아니므로 조문은 공허하지 않다.

## 되돌림 — 바이트 수준으로 증명

되돌리는 과정에서 **실패 1 건을 관측했고 여기 적는다.** BSD `sed -i '' -e '$d' -e '$d' -e '$d' -e '$d'`
는 네 번 주어도 마지막 줄을 **한 줄만** 지운다(같은 스크립트의 `$` 가 한 번의 패스에서
같은 줄을 가리킨다). 뮤턴트 3 줄이 남았다.

`git status --porcelain` 만 봤다면 잡지 못했을 자리다 — 그 파일은 어느 쪽이든 `M` 으로 보인다.
바이트 대조가 잡았다. 최종 제거는 `git checkout HEAD -- <path>` 로 했다(그 파일의 미커밋
변경은 내가 방금 넣은 뮤턴트뿐임을 직전 `git diff` 로 확인한 뒤 실행).

```
$ git diff -- internal/template/templates/.codex/agents/moai/sync-auditor.toml
(무출력 — 커밋본과 바이트 동일)

$ git status --porcelain
 M .claude/settings.json                                    ← 이 카드 범위 밖 (리드 확인, 커밋 금지)
 M .moai/specs/SPEC-CODEX-SKILL-LOADER-001/acceptance.md    ← AC-CSL-012 수정
?? .moai/reports/t452/merge-order-constraint.md
```

되돌림 후 방향 A 를 다시 돌려 `0` 을 재확인했다.

## 조문 수정의 폭발 반경

```
$ git diff --stat -- .moai/specs/.../acceptance.md
 1 file changed, 12 insertions(+), 1 deletion(-)

$ git diff -U0 -- .moai/specs/.../acceptance.md | grep '^@@'
@@ -94 +94 @@        ← When 줄 1 개 치환
@@ -97,0 +98,11 @@   ← 정정 기록 11 줄 추가
```

두 헌크 모두 `AC-CSL-012` 블록 안이다. 다른 12 개 AC · 처분표 · `Then` · `[HARD]` 조항 ·
금지 토큰 집합 · "1 이상" 임계 · 대조군 의무는 그대로다.

## AC-CSL-012 재판정 — **PASS**

스윕 1 (≥1) · 매치 0 · 대조군이 4 종 셀렉터의 생존을 보였다.

## Gaps — 관측하지 않은 것

- **삭제로만 이뤄진 중립성 위반은 이 조문이 잡지 못한다.** 좁힌 대상이 `^+` 줄이기 때문이다.
  원래 조문은 파일 전체를 봤으므로 이 축을 우연히 덮고 있었고, 이번 축소로 사라졌다.
  판정의 의도("새로 들여놓지 않았다")와는 무관한 축이므로 의도적 축소로 남기지만,
  덮이던 것이 사라졌다는 사실 자체는 숨기지 않는다.
- 정정 기록이 인용한 `sync-auditor.toml` **60 번째 줄**이라는 좌표는 baseline 시점 기준이며
  파일이 바뀌면 어긋난다. 토큰 자체는 `grep 'SPEC: {SPEC-ID}'` 로 재측정 가능하다.
- 정정 기록이 인용한 `go test ./internal/template/ -count=1` 종료코드 0 은
  **오케스트레이터가 HEAD `76afb6dea` 에서 직접 관측한 값**이다(manager-spec 은 이 회차에
  돌리지 않았고 그 사실을 자기 Gap 에 적었다). 귀속처를 여기 명시해 둔다.
