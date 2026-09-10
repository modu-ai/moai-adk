# t541 재현 기록 — plan-auditor 교차 산출물 순서 모순

측정 트리: `39b32b11dbc29d36f1b4f68401b82dea6beec2fd` (워크트리 `.claude/worktrees/t541`, 브랜치 `WT-auditor-order-conflict`)
도구: 모든 grep 은 `/usr/bin/grep`(셸 래퍼 아님). 판정 동사 시제품은 awk 만 쓴다.

## 1. 주장

1. 현행 plan-auditor 교리(템플릿·로컬 두 사본)에는 교차 산출물 순서 모순 감사 항목이 없다.
2. 실제 SPEC(`SPEC-CODEX-STALE-SPLIT-FOURTH-001`, t534)에 동시에 따를 수 없는 순서 의무 한 쌍이 있다.
3. 시제품 동사는 그 쌍을 CONFLICT 로 잡고, 방향만 뒤집은 픽스처에서는 잡지 않는다.

## 2. 증거

### 2.1 교리 부재 (주장 1)

`/usr/bin/grep -c -i -E 'ordering conflict|cross-artifact|jointly satisfiable' <템플릿> <로컬>` → `repro/doctrine-absence.txt`

```
internal/template/templates/.claude/agents/moai/plan-auditor.md:0
.claude/agents/moai/plan-auditor.md:0
exit=1
```

대조군 — 같은 바이너리로 `'^### Group [0-9]'` 계수 → `repro/doctrine-control.txt`: 두 사본 모두 `8`, exit=0. 0 은 읽힌 파일 위의 관측된 부재다.

### 2.2 실제 충돌 쌍 (주장 2) — 나란히 인용

`plan.md` §F:

> L87 `### M1 — the render and the bucket (Priority High)` … L111 `Exit: AC-SSF-002, 003, 004, 005, 007 green.`
> L113 `### M2 — the guard and its controls (Priority High)` … L129 `Exit: AC-SSF-001 green, with its RED recorded.`

`acceptance.md` §D.4 Definition of Done:

> L247 `- AC-SSF-001's RED output recorded verbatim in progress.md §E.2, captured BEFORE the M1 render change, with the command and its exit code.`

AC-SSF-001 은 M2 의 종료 기준이고 plan 은 M2 를 M1 뒤에 둔다. §D.4 는 그 RED 를 M1 앞에 요구한다. 구현자는 실제로 `progress.md` L114-116 에 "M2-RED → M1 → M3 로 순서를 바꿨고, §D.4 가 이긴다"고 기록했다 — 충돌이 기록에 남은 유일한 이유다.

### 2.3 시제품 동사 (주장 3) — `repro/order-verb-prototype.sh`

| 입력 | 결과 파일 | 핵심 출력 |
|---|---|---|
| t534 plan + acceptance (실사례) | `repro/run-t534-real.txt` | `CONFLICT: …acceptance.md:247 orders AC-SSF-001 before M1, but …plan.md binds AC-SSF-001 to the exit of M2, which the plan places after M1` / `COLLECTED: 3 milestones (M1 M2 M3), 8 exit bindings, 5 ordering candidates` / exit=0 |
| 같은 plan + L247 만 `AFTER` 로 뒤집은 픽스처 | `repro/run-fixture-after.txt` | CONFLICT 없음, CANDIDATE 5 / exit=0 |
| `SPEC-TODO-ENABLE-FLAG-001` (대조, 순서 절 없음) | `repro/run-control-todoflag.txt` | `COLLECTED: 7 milestones, 9 exit bindings, 0 candidates` / `NONE: 0 records … match (before\|after\|first\|prior to\|pre-change)` |
| `SPEC-STOPCHAIN-TRIM-001` (대조) | `repro/run-control-stopchain.txt` | `GAP: 0 milestone headings collected` — 제목이 `### Milestone M1` 형식이라 시제품이 못 읽음 |

## 3. 기준선 귀속

모든 측정은 이번 실행, 위 트리에서. 코퍼스 수치(`.moai/specs/`, 이번 실행):

- SPEC 디렉터리 830, plan.md 765
- `^#{2,4} M[0-9]` 제목을 가진 plan 512, `### Milestone M[0-9]` 형식 26, 합집합 520
- `Exit:` 줄에 AC 를 묶은 plan 19 — 기계적 CONFLICT 판정이 가능한 모양은 이만큼뿐
- `§D.4` 를 가진 acceptance 204

## 4. 미검증

- 시제품은 `Milestone M1` 제목 형식(26건)을 놓친다 — 본 구현에서 고칠 대상.
- `Exit:` 결속이 없는 plan(746건)에서는 CONFLICT 를 낼 수 없고 CANDIDATE 만 준다. 그 경우 판정은 감사자의 읽기다.
- 코퍼스 전체에 시제품을 돌린 오탐·미탐 측정은 하지 않았다.
- 한 줄에 AC·키워드·마일스톤이 모두 있지 않은 절(여러 문장에 걸친 의무)은 CONFLICT 로 못 잡는다.

## 5. 잔여 위험

- 키워드가 소문자 `before`/`after` 도 받으므로 CANDIDATE 잡음이 있다(t534 에서 5건 중 1건만 실제).
- "blocking" 을 판정 강제(M5 must-pass)로 연결할지 여부는 설계 결정이며 리드 보고 대상이다.
