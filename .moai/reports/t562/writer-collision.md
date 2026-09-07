# t562 — 워크트리 동시 작성자 사건 기록

카드 t562 · 워크트리 `.claude/worktrees/t562` · 브랜치 `WT-codex-read-inverse`
기록 주체: lane-2 (레인 오케스트레이터, 사건의 원인 제공자)
기록 시점 트리 상태: `bbba4f672`
리드 지시(2026-09-08)에 따라 완료 보고에 접지 않고 별도 파일로 남긴다.

---

## 1. 무슨 일이 있었는가 (인과)

run 단계 작업자 `manager-develop`(t562-m3)가 M3 완료 보고를 보낸 뒤, 레인이 리드의
CWD 경고와 그 정정을 "참고하라"는 뜻으로 그 에이전트에 `SendMessage` 했다.

**완료된 에이전트에 보낸 메시지는 그 에이전트를 transcript 에서 재기동한다.** 레인은
그 성질을 계산에 넣지 않았고, m3 가 종료된 것으로 보고 sync 단계 작업자
`manager-docs`(t562-sync)를 스폰했다. 그 결과 같은 워크트리에 쓰기 가능한 에이전트
둘이 동시에 돌았다.

인과 사슬은 세 단계다:

1. m3 완료 보고 도착 → 레인이 m3 를 종료 상태로 간주
2. 레인이 m3 에 정정 메시지 발신 → **m3 재기동 (두 번째 작성자 발생)**
3. 레인이 sync 스폰 → 두 작성자가 run→sync 경계에서 겹침

"쓰기 가능한 에이전트 둘을 동시에 돌리지 않는다"에 대한 위반이며, 판단 주체는
레인이다. m3 도 sync 도 지시받은 범위를 벗어나지 않았다.

---

## 2. 유실 실측

### 2.1 사건 인지 직후 측정 (트리 `37a631a70`)

```
$ git rev-parse --short HEAD
37a631a70
$ git status --short
 M .moai/specs/SPEC-CODEX-SKILL-PATH-READBACK-001/progress.md
 M .moai/specs/SPEC-CODEX-SKILL-PATH-READBACK-001/spec.md
 M CHANGELOG.md
$ git show 37a631a70 -- .moai/specs/SPEC-CODEX-SKILL-PATH-READBACK-001/progress.md | /usr/bin/grep -c '§E.4'
0
$ git show HEAD:.moai/specs/SPEC-CODEX-SKILL-PATH-READBACK-001/progress.md | /usr/bin/grep -c '§E.4'
0
$ /usr/bin/grep -n '^## §E' .moai/specs/SPEC-CODEX-SKILL-PATH-READBACK-001/progress.md
7:## §E.1 Plan-phase Audit-Ready Signal
18:## §E.2 Run-phase Evidence
192:## §E.3 Run-phase Audit-Ready Signal
211:## §E.4 Sync-phase Audit-Ready Signal
$ git show --stat --format='' 37a631a70
 .moai/reports/t562/ac-csrb-006-remeasure.log       |   6 +
 .moai/reports/t562/run-m3.md                       | 128 +++++++++++++++++++++
 .moai/reports/t562/tree-attribution-remeasure.log  |  62 ++++++++++
 .../SPEC-CODEX-SKILL-PATH-READBACK-001/progress.md |  17 +++
 4 files changed, 213 insertions(+)
$ git status --porcelain -- internal/
(빈 출력)
```

판독:

- **m3 의 커밋이 sync 의 `§E.4` 를 삼키지 않았다** — 커밋 diff 내 `§E.4` 0건.
- **그 시점 `§E.4` 는 미커밋이었다** — `HEAD:progress.md` 에 0건이지만 작업트리
  211행에 존재. 즉 sync 의 진행분이 작업트리에 온전히 살아 있었다.
- **m3 의 `§E.2` M3 절도 온전** — 작업트리에 존재.
- **생산 파일 무변경** — `git status --porcelain -- internal/` 빈 출력.

### 2.2 현재 트리에서의 재확인 (트리 `bbba4f672`)

```
$ git show HEAD:.moai/specs/SPEC-CODEX-SKILL-PATH-READBACK-001/progress.md | /usr/bin/grep -c '§E.4'
3
$ git show 37a631a70 -- .moai/specs/SPEC-CODEX-SKILL-PATH-READBACK-001/progress.md | /usr/bin/grep -c '§E.4'
0
$ /usr/bin/grep -c '^### M3 —' .moai/specs/SPEC-CODEX-SKILL-PATH-READBACK-001/progress.md
1
```

**두 시점의 `§E.4` 개수 0 과 3 은 모순이 아니다.** 2.1 의 `0` 은 "그 시점 sync 의
작업이 아직 커밋 전"이라는 사실이고, 2.2 의 `3` 은 "sync 가 `18bf8cc06` 로 커밋을
마쳤다"는 사실이다. 같은 사실을 두 시점에 잰 것이다 — 뒤에 이 기록을 읽는 사람이
둘 중 하나를 오측으로 읽지 않도록 명시한다.

`37a631a70` 의 `§E.4` 0건은 두 시점 모두 불변이며, 그것이 "삼키지 않았다"의 근거다.

### 판정

**유실 0.** m3 의 산출물과 sync 의 산출물이 각각 온전하며, 어느 쪽도 상대의 작업을
덮거나 인덱스에 흡수하지 않았다.

---

## 3. `37a631a70` 을 되돌리지 않은 판단

되돌리지 않았다. 근거 셋:

1. **유효한 작업이다.** 그 커밋은 CWD 정정을 반영한 트리귀속 재측정으로, 기록된
   M3 값 15건을 절대경로 고정으로 다시 재어 전부 동일 재현한 결과다. 생산 코드는
   0줄이고 만진 파일은 4개(증거 3 + 자기 `§E.2` 절)뿐이다.
2. **되돌리면 실측 15건이 사라진다.** 이 커밋을 revert 하면 CWD 드리프트가 없었다는
   독립 확인 근거가 함께 없어진다 — 사건 대응이 사건보다 큰 손실을 만든다.
3. **되돌릴 결함이 없다.** §2 의 실측대로 이 커밋은 남의 작업을 훼손하지 않았다.
   되돌림의 대상은 훼손이지 동시성 자체가 아니다.

---

## 4. 유실 0 의 원인

**유실이 0이었던 원인은 m3 가 명시 pathspec 으로만 스테이징했기 때문이다.**

m3 는 커밋 직전 스테이징 단계에서 자기가 쓰지 않은 `CHANGELOG.md` 수정과 남의
`progress.md` 수정분을 관측했고, 인덱스에 넣지 않았다. `git add -A` / `git add .` /
`git commit -a` 중 어느 하나였다면 sync 의 미완성 작업이 m3 의 커밋에 섞여 들어갔을
것이고, 그 경우 손실은 삭제가 아니라 **미완성 상태의 조기 고착**이라 삭제 검사
(`git status | grep '^ D'`)로는 잡히지 않는다.

> **규율이 내 실수를 막았지 내가 막은 게 아니다.**

이 문장을 기록에 남기는 이유는, 이 사건을 「유실 0 이었으니 별일 아니었음」으로 읽는
판독을 막기 위해서다. 결과가 무사했던 것과 판단이 옳았던 것은 다른 사실이다. 같은
실수를 sweep 스테이징을 쓰는 작업자와 함께 저질렀다면 결과는 달랐다.

---

## 5. 채택된 조치

- **m3 에 더 이상 메시지를 보내지 않는다.** 보내는 행위 자체가 재기동이므로 침묵이
  곧 stand-down 유지다. 「참고하라」는 무해한 전달은 존재하지 않는다.
- **정정을 전달해야 하면 다음 스폰의 프롬프트에 싣는다.** 살아 있는(또는 되살릴 수
  있는) 에이전트에 부가 정보를 보내는 경로를 쓰지 않는다.
- sync 를 단독 작성자로 두고 진행했다. 이후 감사(`sync-auditor`)도 단독이다.
- `37a631a70` 은 유지한다(§3).

---

## 6. 이 기록이 판정에 갖는 지위

이 파일은 판정 근거가 아니라 **공정 결함의 관측 기록**이다. 카드의 AC 판정에는
영향을 주지 않으며, 영향을 주지 않는다는 것이 §2 의 실측 내용이다. sync-audit 은
이 파일을 읽고 "유실 0"이 실제로 성립하는지 독립 판정한다 — 레인의 자기 보고를
근거로 채택하지 않는다.

미관측으로 남는 것: 두 작성자가 동시에 **같은 파일의 같은 구간**을 쓰는 경합은
발생하지 않았으나(§E.2 와 §E.4 는 서로 다른 절), 그것이 우연인지 구조적으로
불가능했는지는 판정하지 않았다. 다음 사건에서 같은 결과를 기대할 근거로 쓰지 말 것.
