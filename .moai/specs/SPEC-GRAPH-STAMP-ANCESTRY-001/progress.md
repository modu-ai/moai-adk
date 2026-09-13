# SPEC-GRAPH-STAMP-ANCESTRY-001 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

- card: `t688`
- plan_status: `audit-ready`
- plan_complete_at: `2026-09-13T23:31:05+0900`
- worktree: `.claude/worktrees/t688`
- branch: `WT-graph-stamp-freshness`
- baseline_head: `30d505ac4`
- tier: `M`
- artifacts: `spec.md` (v0.2.0, `status: draft`) · `plan.md` · `acceptance.md` · `progress.md` — Tier M 산출물 3종 + 이 진행 기록이 모두 존재한다. 종전 이 절이 "`spec.md`만 생성됨"이라고 적고 있었으나 그 서술은 `b4626c042` 시점 이후로 거짓이었다.
- next_owner: `plan-auditor`
- next_action: 독립 plan-audit를 실행한다. Tier M PASS 임계는 `0.80`이며, `acceptance.md AC-GSA-011`이 네 렌즈 교차 모순 검사를 감사 항목으로 지정한다.

### 이번 보강에서 메운 것

산출물 네 개를 함께 읽고 교차 불일치만 겨냥해 고쳤다. 재작성은 하지 않았다.

- **라벨 충돌** — `spec.md §B.1`의 목표가 `M1~M3`이고 `plan.md §F`의 마일스톤이 `M1~M5`여서, `M3`이 두 문서에서 서로 다른 일을 가리켰다. 목표를 `G1~G3`으로 분리하고 두 축이 1:1이 아님을 명시했다. `REQ-GSA-011`도 마일스톤 이름을 걷어냈다 — 요구사항 층이 계획 층 식별자를 참조하고 있었다.
- **규범 표의 누락 행** — `spec.md §D` 판정 순서표에 본문 부재(C1) 행이 없었다. `REQ-GSA-003`이 막으려는 오독이 정확히 "`VerdictAbsent` 문자열만 보고 C1과 비조상을 합치는 것"인데, 그 구분이 규범 표에서는 검증할 수 없었다. C1 행과 `system error` 열을 추가했고, 종료코드만으로는 갈라지지 않는다는 사실(C1과 reachable stale이 둘 다 exit 1)을 적었다.
- **생성기 5문서 미명시** — `docs-truth.md` 제외가 세 문서에 걸쳐 있었는데 다섯이 무엇인지는 어디에도 없었다. `.claude/skills/moai/workflows/codemaps.md` § Output Files를 읽어 `overview`/`modules`/`dependencies`/`entry-points`/`data-flow`로 확정하고 `spec.md §B.1 G3`·`plan.md §A.3`·`§F M4`에 명시했다. `fold-judgments.txt`와 `provenance.json`도 생성기 밖임을 함께 적었다.
- **제외 범위의 소재** — warning-only 완화 제외가 이 진행 기록의 메모에만 있었다. 제외 범위는 SPEC이 소유해야 진행 기록과 함께 사라지지 않으므로 `spec.md §G`로 올렸다.
- **문서 수준 트리 핀** — `acceptance.md §A`가 단일 핀 `7097e6e21`을 제시했지만 `RED-GSA-008`은 `b4626c042`에서 측정됐다. 핀 우선순위(자기 핀이 이긴다)를 명시해 두 값의 차이가 드리프트로 읽히지 않게 했다.

### 보존된 두 결함 분리

이 SPEC의 핵심은 두 적색 원인을 서로 다른 실패 종류로 유지하는 것이며, 보강 과정에서 흐리지 않았다.

- **스탬프 도달 불가** — 저장된 clean `commit_sha`가 checkout `HEAD`의 조상이 아니다. 구조적/system error, exit 2, 숫자 freshness 없음.
- **freshness 값 초과** — 스탬프는 조상이지만 `described-source-diff`가 임계 40 이상이다. 진짜 stale이며, 재생성과 도달 가능한 재스탬핑이 필요하다. exit 1, 숫자와 귀속 유지.

warning-only 모드와 자동·맨손 재스탬프는 양쪽 모두에서 제외다.

### 확인된 계획 근거

- 읽기 전용 조사 4개 렌즈가 producer/consumer 경로, release merge 위상, 기존 SPEC 계약, 회귀 위험을 각각 조사했다. 합성 기록은 `plan.md §B`가 대신한다.
- 현재 기준에서 develop 스탬프 `f7b4919541f10b6415173b6bc9fb7192e8824450`은 `origin/develop`의 조상이지만 `origin/main`의 조상은 아니다.
- 기준 커밋 `7097e6e21`의 CI 근거는 `value=254`, `threshold=40`, `verdict=stale`, exit 1이다. 카드의 `value=224`는 같은 기준에서 재현되지 않았으므로 인수 기준의 고정값으로 쓰지 않는다.

### Evidence

이 보강 turn에서 실제로 실행한 명령과 원문 출력이다.

```text
command: moai spec lint SPEC-GRAPH-STAMP-ANCESTRY-001 --strict
tree_sha: 30d505ac4 (보강 편집 후, 커밋 전)
exit_code: 0
stdout:
0 error(s), 0 warning(s)
INFO  OwnershipTransitionUnmeasured  spec.md:1  transition "(none)" → "draft" expected owner
      "manager-spec" but commit b4626c0421d3b21564fcf7ba08967889af52a278 has no
      Authored-By-Agent trailer — ownership transition unmeasured
```

```text
command: moai spec audit --filter-spec SPEC-GRAPH-STAMP-ANCESTRY-001
tree_sha: 30d505ac4
exit_code: 0
stdout:
Total SPECs: 1 / Grandfathered: 0 / Modern-era clean: 1 / Drift findings: 0
No drift findings — all modern-era SPECs clean.
```

`OwnershipTransitionUnmeasured`는 INFO이며 `--strict`의 error/warning 집계에 들어가지 않는다. 이 보강 커밋은 `Authored-By-Agent: manager-spec` trailer를 담으므로, 다음 전이 측정부터는 귀속이 성립한다. 선행 커밋 `b4626c042`의 미측정 상태는 소급 수정하지 않는다.

- RED 원장 스크립트: `.moai/reports/t688/red-{ancestry,cli-unreachable,history-topologies}.sh`, `red-push-guard.py`
- 중단 기록: `.moai/reports/t688/gateway-502-recurrence.md`, `gateway-400-20260913T133738Z.md`, `gateway-400-20260913T141256Z.md`

### Gaps

이 보강 turn이 **관측하지 않은** 것.

- RED 원장 4건(`RED-GSA-001/002/006/008`)을 재실행하지 않았다. `acceptance.md §D`에 담긴 출력은 선행 turn의 관측이며, 이 turn은 그 기록의 존재만 확인했다.
- `BASE-GSA-001`(`moai graph check --root .`)을 재측정하지 않았다. `value=254`는 `7097e6e21`에 귀속된 값이고 현재 트리 `30d505ac4`에서 다시 재지 않았다.
- Go 테스트, `go vet`, `golangci-lint`를 실행하지 않았다. plan 단계이므로 코드 변경이 없다.
- `plan-auditor` 독립 감사를 실행하지 않았다 — 이것이 `next_action`이다.
- 두 gateway 중단(502 1건, 400 2건)의 하위 원인은 이 카드에서 진단하지 않았다. `t707`/`t708` 소관이다.
- push하지 않았다. 원격 반영은 리드의 일괄 소관이다.

### Residual risk

산출물은 자기 정합성이 올라갔지만 독립 감사를 통과한 적은 없다. 특히 `AC-GSA-011`의 네 렌즈 교차 모순 검사는 정의상 저자가 스스로 만족시켰다고 주장할 수 없는 항목이다 — `plan-auditor`가 각 행의 반증 조건을 원본 인용에 대조해야 성립한다. 또한 `value=254`가 현재 트리에서 재현되는지 확인하지 않았으므로, run 시작 시 `plan.md §C` 실행 전 점검이 이 값을 다시 재는 것이 전제다.

## §E.2 Run-phase Evidence

_대기 중 — Implementation Kickoff Approval 전에는 run을 시작하지 않는다._

## §E.3 Run-phase Audit-Ready Signal

_대기 중._

## §E.4 Sync-phase Audit-Ready Signal

_대기 중._
