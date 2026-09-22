# 배치 끝 정리 판정 — 2026-08-31

- 대상: codemaps 일괄 재생성 + `MEMORY.md` 색인 압축 (카드 없음, 운영자 승인분)
- 측정 트리: `.claude/worktrees/t287`, 브랜치 `WT-guard-substitution`
- 흡수 기준: `origin/develop = fef7a4b9b` → 흡수 병합 `7fc0af324`

## 1. codemaps 재생성

### Claim
배치 7건이 누적시킨 codemaps staleness 를 한 번에 닫았다. 특정 카드 귀속이 아니다.

### Evidence

착수 시점(흡수 후, 재생성 전):

    $ ./bin/moai graph check
    codemaps  metric=described-source-diff value=52 threshold=40 verdict=stale

재생성 + 스탬프 후, 같은 바이너리로 재판독:

    $ ./bin/moai graph check
    codemaps  metric=described-source-diff value=0 threshold=40 verdict=fresh
    mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=absent
    edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=absent

변경 표면 (`git diff --stat origin/develop...HEAD`): 6 파일, +44/-19 — 문서 5 + `provenance.json`.
Go 코드 0.

### Baseline-attribution

판정에 쓴 바이너리는 이 트리 산물이다. `make build` rc=0 이후
`strings bin/moai | grep -c 7fc0af324` → **4** (흡수 병합 SHA 가 BuildID 에 박힘).
**PATH 바이너리는 쓰지 않았다** — 설치본은 이 트리보다 뒤처질 수 있고, 그러면 이 트리의 규칙이
판정에 실리지 않는다.

수치 표본 재확인(오케스트레이터 독립):

    $ find internal/cli -name '*.go' ! -name '*_test.go' | wc -l   → 261
    $ ls -d internal/*/ | wc -l                                     → 64

둘 다 문서에 적힌 값과 일치.

### 스탬프 앵커

`provenance.json` 의 `commit_sha` = `7fc0af324`(흡수 병합). 브랜치 로컬 HEAD 가 아니라
develop 을 이미 담고 있는 병합 커밋이며, 이 저장소는 `--no-ff` 로 합치므로 병합 후에도 도달 가능하다.
`--commit` 플래그는 쓰지 않았다(squash 아님).

### Gaps

- `mx-index` / `edges` 는 이 워크트리에서 `verdict=absent` 다 — 미추적 런타임 산물이라 새 워크트리에
  없을 뿐이며, 재생성 **전후 동일**하다. 이 작업이 만든 상태가 아니다. CI 에서는 둘 다 fresh 로 관측된다.
- `graph check` 전체 종료 코드는 `1` 이다(absent 두 층 때문). codemaps 층만 fresh 로 닫혔다.
- `docs-truth.md` 의 다른 드리프트 3건(에이전트 수 11 vs 12, `/moai` 명령 수 15 vs 16, GLM 티어표가
  `glm-5.2` 기준)은 **고치지 않았다**. `f1d80b305..HEAD` 창 이전부터 있던 것이고 `.claude/` 소관이라
  이 배치의 described roots 밖이다.
- `.moai/project/structure.md:71` 이 같은 패키지 수를 낡은 값으로 들고 있다 — 별건.
- `internal/kanban` 은 이 diff 에 변경이 있으나 여섯 문서 어디에도 원래 커버리지가 없었다. 이 작업이
  깬 것이 아니라 기존 공백이다.
- Go 빌드/테스트/vet 미실행. 변경이 마크다운 5 + JSON 1 이라 닿는 Go 패키지가 없다 — 다만 그것은
  논증이지 측정이 아니다.

## 2. MEMORY.md 색인 압축

### Claim
로더 상한 초과를 해소하면서 색인 항목은 하나도 잃지 않았다.

### Evidence

    before: 31,416 bytes / 121 entries / 160 lines
    after:  25,571 bytes / 121 entries / 160 lines      (상한 25,600)

    $ grep -c '^- \[' MEMORY.md   → before 121, after 121
    $ grep -c '^'      MEMORY.md   → before 160, after 160

행동을 바꾸는 절(`작업 규율` ~ `템플릿·문서`, 29-137행)은 **바이트 동일**:

    $ diff <(sed -n '29,137p' MEMORY.md.bak) <(sed -n '29,137p' MEMORY.new)
    (출력 없음, exit 0)

경로: `~/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md`.
**저장소 밖이라 git 커밋 대상이 아니다.** 백업 `/tmp/MEMORY.md.bak`.

### 손실 명시

마지막 351 바이트를 맞추려고 오래된 착지 항목 **19건에서 `(session: <8자>)` 괄호를 제거**했다.
색인에서 세션 상관 추적이 사라진다 — 각 토픽 파일 안에 같은 값이 남아 복구는 가능하나,
색인만 읽는 독자는 못 본다. 최근 5건은 유지했다.

### Gaps / Residual-risk

- 착수 시점 실측은 **31,416** 이었다(배차문의 29,586 이 아님). 그 사이 다른 세션이 항목을 추가한
  것으로 보이며, 그래서 감축 목표가 3,986 이 아니라 5,845 였다.
- `작업 규율` 절 하나가 **14,578 bytes = 예산 25,600 의 57%** 다. 이번엔 나머지를 깎아 맞췄지만
  규율이 몇 줄만 더 늘면 깎을 여지가 없다. 그 절 자체의 압축·분할이 다음 재료다.
- 로더가 실제로 이 파일을 어떻게 자르는지는 관측하지 않았다 — 바이트 수만 맞췄다.
