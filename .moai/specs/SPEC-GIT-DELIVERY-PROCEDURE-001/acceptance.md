# Acceptance — SPEC-GIT-DELIVERY-PROCEDURE-001

> 기준마다 Given-When-Then과 판정 명령을 둔다. 모든 부재 기준은 같은 검출기로 알려진 적중을 먼저 잡는 **양성 대조**를 가진다. 검증 출력은 파일로 보내고 exit code를 파이프 없이 따로 읽는다(`| head`·`| tail`·`| grep` 금지).

## 공통 변수와 명령 관례

```bash
BASE=255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089   # R1 고정(0.2.5 재고정): develop f1f034bb4 두 번째 흡수 병합, 2026-09-11 이 워크트리에서 해석. 고정 스냅숏 전용(0.2.6)
CARD_BASE=<git merge-base develop HEAD>          # 핀하지 않음(0.2.6): 판정할 때마다 $E/card-base.txt 로 다시 구해 글자 그대로 넣는다
E=.moai/reports/t622/run                         # 추적 증거 경로
T=internal/template/templates                    # 템플릿 루트
```

- 명령은 워크트리 루트에서 실행한다. "exit" 는 직전 명령의 exit code를 `echo "exit=$?"` 로 따로 기록한 값이다.
- **값은 판정 명령에 글자 그대로 넣는다.** 이 파일의 명령에 적힌 `$BASE`·`$E`·`$T` 는 위 세 값을 가리키는 표기다. 실행할 때는 그 자리에 `255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089`·`.moai/reports/t622/run`·`internal/template/templates` 를 글자 그대로 바꿔 넣는다(예: `git show 255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089:.claude/rules/moai/core/agent-common-protocol.md > .moai/reports/t622/run/base-agent-common-protocol.md`). `$CARD_BASE` 는 판정 직전에 `git merge-base --all develop HEAD > $E/card-base.txt` 로 구한 한 줄 값을 같은 방식으로 넣는다(이 파일에 값을 적지 않는다). **git 명령과 `manager-git.md` 경로를 담은 명령에는 `BASE=…; E=…; T=…; <판정 명령>` 처럼 변수를 먼저 지정하는 형태를 쓰지 않는다** — 워크트리 가드가 `E=…; … > $E/…` 형태를 거부한다(축소판 plan-audit 3회차 관측: 이 형태의 git 판정과 `manager-git.md` 를 읽는 awk·grep 판정이 거부됐고, 같은 판정을 값을 글자 그대로 넣어 실행하면 돌았다). Bash 호출은 매번 새 프로세스라서 앞 호출에서 지정한 변수는 다음 호출에 남지 않는다. 치환하지 않은 채 돌린 호출에서는 `git show $BASE:<경로>` 가 `git show :<경로>`(인덱스 사본)로, `git diff --name-only $BASE $CARD_BASE -- <경로>` 가 작업 트리 차이로 바뀌므로 그 출력은 판정 근거가 아니다. `manager-git.md` 경로를 담은 복합 스크립트도 가드가 거부하므로, 한 호출에는 판정 명령 하나만 둔다.
- 기준 트리 사본은 사전 점검에서 `git show $BASE:<경로> > $E/base-<이름>` 으로 반출해 둔다(대상 목록은 plan.md §C 2단계).
- **`$BASE` 는 고정 스냅숏, `$CARD_BASE` 는 읽는 시점 범위다 (0.2.6).** `$BASE` 는 반출본·양성 대조·사본 덩어리·미러 기준선·AC-GDP-002·003 기준 절처럼 움직이면 안 되는 값에만 쓴다. "이 카드가 무엇을 바꿨는가·커밋했는가" 를 재는 판정(AC-GDP-014·015·016·025·030)은 리터럴 `$BASE` 를 쓰지 않는다 — 통합 창에서 로컬 develop 을 다시 흡수하면 `$BASE..HEAD` 에 develop 커밋이 섞이기 때문이다(`.claude/rules/local/gitflow-lane-protocol.md` §8, spec.md §E.2). 그 판정의 형태는 둘이다:
  - 차이: `git diff develop...HEAD -- <경로>`. 세 점 형태는 `git diff $CARD_BASE HEAD -- <경로>` 와 같은 결과를 내며(0.2.6 측정: 두 형태의 `--name-only` 출력 `cmp` exit 0, `.moai/reports/t622/reanchor-fix/card-range-names-current*.txt`), 가드가 거부하는 `$(…)` 없이 한 명령으로 돈다.
  - 커밋 목록: `git log --no-merges --format=%H HEAD --not develop -- <경로>` — develop 에서 닿는 커밋을 모두 빼므로 흡수된 develop 커밋은 들어오지 않고 카드의 비병합 커밋은 모두 남는다(0.2.6 측정: 현재 트리와 흉내 낸 흡수 커밋에서 같은 23줄, `diff` exit 0 — `reanchor-fix/card-commits-*.txt`).
  - 범위 대조(판정마다 한 번): `git merge-base --all develop HEAD > $E/card-base.txt` 는 정확히 1줄이어야 하고, `git diff --name-only develop...HEAD > $E/card-range-names.txt` 뒤 `test -e` exit 0·`test -s` exit 0 이어야 한다. `test -s` 가 exit 1 이면 "변경 없음" 이 아니라 **"측정 불가"** 로 기록한다(PASS 아님). 형태의 양성 대조(0.2.6, 흡수를 거친 실제 카드 t645): `git diff --name-only 0db675bedcae69c7ade2f02f106649a715fb14f6...c7874923a4059c8b2bffa955c92a3e9f1230d118` → 그 카드 자기 파일 8개, 흡수 전 분기점 `dae1b070b44b6d9ef2f4334372467ce570ce9202` 에 고정해 재면 19개(`reanchor-fix/control-t645-*.txt`).
  - 한계: 이 판정은 **병합 전 전용**이다. 카드가 develop 에 병합되면 merge-base 가 카드 tip 자신이 되어 범위가 비고 공허하게 통과한다. 병합 뒤 근거는 병합 트리 동일성(병합 커밋의 `git rev-parse <merge>^{tree}` 가 재측정한 트리와 같음)이다.
- **판정 전 스냅숏 신선도 점검 (0.2.6).** `$CARD_BASE` 를 구한 직후, 어느 판정보다 먼저 실행한다:

```bash
git diff --name-only $BASE $CARD_BASE -- .claude/agents/moai/manager-git.md .claude/rules/moai/core/agent-common-protocol.md .claude/skills/moai/workflows/sync/delivery.md .claude/skills/moai/workflows/sync/doc-execution.md .claude/skills/moai/SKILL.md .claude/skills/moai/references/reference.md .claude/skills/moai/workflows/sync/quality-gates-context.md .claude/skills/moai/workflows/sync.md .claude/commands/moai/sync.md .agents/skills/moai-sync/SKILL.md $T/.claude/agents/moai/manager-git.md $T/.claude/rules/moai/core/agent-common-protocol.md $T/.claude/skills/moai/workflows/sync/delivery.md $T/.claude/skills/moai/workflows/sync/doc-execution.md $T/.claude/skills/moai/SKILL.md $T/.claude/skills/moai/references/reference.md $T/.claude/skills/moai/workflows/sync/quality-gates-context.md $T/.claude/skills/moai/workflows/sync.md $T/.claude/commands/moai/sync.md.tmpl $T/.agents/skills/moai-sync/SKILL.md $T/.codex/agents/moai/manager-git.toml > $E/snapshot-stale.txt
test -e $E/snapshot-stale.txt
# 기대: exit 0 — 파일이 없으면 점검이 돌지 않은 것이다(판정 불가)
test -s $E/snapshot-stale.txt
# 기대: exit 1 — 흡수가 스냅숏 파일을 바꾸지 않았다. exit 0 이면 목록의 파일마다 고정 스냅숏이 낡았다:
#       그 파일의 반출본을 $CARD_BASE 에서 다시 만들고, 그 파일이 쓰는 기준선(spec.md §A.1 줄, §A.4 덩어리, AC-GDP-002·003 기준 절,
#       AC-GDP-025 clause 개수)을 다시 재어 기록한 뒤 리드에 올린다. 낡은 스냅숏으로 판정한 결과는 근거가 아니다.
git diff --name-only $BASE $CARD_BASE -- internal/template/ .claude/rules/moai/ > $E/snapshot-stale-mirror.txt
# 기록 — 비어 있지 않으면 미러 테스트 기준선이 낡았을 수 있다(AC-GDP-013 (e) 귀속 규칙)
```

  0.2.6 측정(`.moai/reports/t622/reanchor-fix/`): 현재 merge-base `f1f034bb4` 에서 `snapshot-stale-current.txt` `test -e` 0·`test -s` 1, 흉내 낸 흡수 뒤 merge-base `00ae57ad7` 에서 `snapshot-stale-emulated.txt` 같은 값, 같은 형태를 `b412f8a33..$BASE` 에 돌린 양성 대조 `snapshot-stale-control-b412.txt` 8줄. 미러 기준선용 점검은 현재 빈 결과, 흉내 낸 흡수 뒤 8줄(`snapshot-stale-mirror-emulated.txt` — 기준선 FAIL 쌍인 `spec-workflow.md` 두 사본 포함. 그 FAIL 이 사라지는 것은 (e) 위반이 아니다).
- 검출식 안의 `git` 은 `[g]it` 으로, `parallel` 은 `para[l]lel` 로, perl 코드 안의 `git` 은 `\x67it` 로 쓴다. 셸 변수를 받는 `sed`·`perl` 과 경로를 만드는 반복문은 가드가 거부할 수 있으므로 경로를 글자 그대로 쓴다.
- 판정용 grep은 `/usr/bin/grep` 으로 실행한다. 개수가 찍히지 않은 결과는 판정 불가로 기록한다.
- **빈 결과가 PASS인 판정은 파일 존재부터 확인한다.** 기대가 `test -s <파일>` exit 1 이거나 "빈 파일" 인 판정(그리고 빈 결과로 경우를 가르는 AC-GDP-030 판정)은 바로 앞에 `test -e <파일>` 을 두고 exit 0 을 기대한다. `test -e` 가 exit 1 이면 판정 명령이 파일을 만들지 않은 것이므로 판정 불가로 기록한다(PASS 아님). 존재 확인이 없으면 돌지 않은 판정과 돌아서 아무것도 찍지 않은 판정이 같은 exit 1 을 낸다 — 3회차 감사에서 완료로 보고된 호출이 출력 파일을 만들지 않은 사례가 관측됐다. 뮤턴트는 §D.2.
- **범위 파일 집합(아홉 개, 로컬·템플릿)** — 로컬·템플릿 사본 한 쌍을 한 파일로 세고, 이름이 다른 명령 원본 `sync.md`/`sync.md.tmpl` 도 한 쌍으로 센다. 생성물과 게시본은 범위 파일로 세지 않는다: `.claude/agents/moai/manager-git.md`, `.claude/rules/moai/core/agent-common-protocol.md`, `.claude/skills/moai/workflows/sync/delivery.md`, `.claude/skills/moai/workflows/sync/doc-execution.md`, `.claude/skills/moai/SKILL.md`, `.claude/skills/moai/references/reference.md`, `.claude/skills/moai/workflows/sync/quality-gates-context.md`, `.claude/skills/moai/workflows/sync.md`, 명령 원본 `.claude/commands/moai/sync.md`(로컬) / `.claude/commands/moai/sync.md.tmpl`(템플릿). 생성물 `$T/.codex/agents/moai/manager-git.toml`, 게시본 `$T/.agents/skills/moai-sync/SKILL.md`(로컬 사본 `.agents/skills/moai-sync/SKILL.md`).
- **기준 트리 측정 (0.2.5)**: 0.2.5 작성 시점 HEAD `b24f2e184` 와 `$BASE` 의 차이는 보고서 두 파일뿐이고(`git diff --stat 255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089 HEAD`), 범위 루트(`internal/template/templates/`, `.claude/`, `.agents/`, `Makefile`, `docs-site/content`) 작업 트리는 `$BASE` 와 차이가 없다(빈 출력; 같은 형태를 `b412f8a33..$BASE` 에 돌리면 113개 파일 — 양성 대조). 증거 `.moai/reports/t622/reanchor/`(명령과 관측 값 목록은 `index.md`). "plan 작성 시점 측정" 값 가운데 0.2.5 에서 다시 잰 값은 그렇게 표시했고, 표시가 없는 값은 `b412f8a33` 에서 잰 값이다 — 해당 파일·절이 `b412f8a33..$BASE` 에서 바뀌지 않았거나(템플릿 사본, 생성물, 게시본, 발행기, 테스트 파일), 로컬 사본은 줄만 밀리고 절 본문은 같다(아래 기준마다 적음).
- **줄번호**: 기대값에 적힌 파일 줄번호는 반출 사본 기준이다. `$E/base-<이름>.md` 는 로컬 사본을 반출하므로 로컬 줄이다(`manager-git.md` 6행부터 템플릿 +2, `delivery.md` 템플릿 279~421행 구간 +25). 템플릿 줄과 다르면 괄호에 템플릿 줄을 적었다(spec.md §A.1 표).

## §D AC 표

**번호 방식**: 0.1.3 번호를 유지한다(spec.md §C.2). 카드 t658로 옮긴 번호와 철회된 번호는 판정하지 않는 자리표시로 남기고 spec.md §G에서 추적한다. 새 기준은 AC-GDP-025~030.

| AC ID | REQ | 등급 | 상태 | 요약 |
|---|---|---|---|---|
| AC-GDP-001 | REQ-GDP-001 | MUST-PASS | 활성 | `manager-git.md`·`.toml` 동기화 절에서 fetch 가 rev-list 와 같은 배치로 묶이지 않고 순서가 지시됨 |
| AC-GDP-002 | REQ-GDP-002 | MUST-PASS | 활성 (0.2.5: 회귀 방지, 0.2.6: 두 부분) | (a) 순서 — 두 사본 Pre-Spawn 코드 블록에서 fetch 가 끝난 뒤 rev-list 가 시작됨(줄 끝에 고정한 fetch 줄 → 바로 다음 코드 줄 `fetch_status=$?` → 그 뒤 rev-list, 또는 `;`·`&&` 로 이은 한 줄). (b) 보존 — 두 사본 Pre-Spawn 절 전체가 고정 기준 절과 diff exit 0(Lane B 병렬 허용, 두 해석 표, 착지 블록의 상태 관측·실패 시 exit 포함). BASE 에서 이미 초록 — 이 카드의 작업을 재지 않고 보존을 지킨다. MUST-PASS 근거는 `verification-completeness.md` §4 보존 단언(AC-GDP-003 과 같음) |
| AC-GDP-003 | REQ-GDP-003 | MUST-PASS | 활성 | Pre-Edit Sync Check 절 불변 |
| AC-GDP-004 | REQ-GDP-004 | MUST-PASS | 활성 | `delivery.md` 병합 명령이 `--<merge_method>` 로 해석 |
| AC-GDP-005 | REQ-GDP-005 | MUST-PASS | 활성 | 범위 파일과 `.toml` 에서 `--squash` 고정 `gh pr merge` 가 기본값 설명 문장뿐 |
| AC-GDP-006 | REQ-GDP-006 | MUST-PASS | 활성 | `delivery.md`·`doc-execution.md` 에 워크트리 기본 병합 문구가 없고 `manager-git.md` 를 기준으로 밝힘 |
| AC-GDP-007 | — | — | 카드 t658로 이동 | 명령 검출식 분류 장부 |
| AC-GDP-008 | — | — | 카드 t658로 이동 | 절차 참조 유지 |
| AC-GDP-009 | — | — | 카드 t658로 이동 | 흐름 표지·접두 합집합 |
| AC-GDP-010 | — | — | 카드 t658로 이동 | 워크트리 흐름·Frozen 기록 |
| AC-GDP-011 | — | — | 철회(0.1.2) | OD-1 선택지 2 경로 |
| AC-GDP-012 | — | — | 철회(0.1.2) | OD-1 선택지 3 경로 |
| AC-GDP-013 | REQ-GDP-013 | MUST-PASS | 활성 (파일 확장, 0.2.5 기준선 재측정) | 범위 파일 아홉 개 사본 일치(바이트 동일 2쌍, BASE 차이 본문 보존 8쌍), 범위 파일을 덮는 테스트 2개 실행·통과, 미러 테스트가 BASE 기준선 대비 새 FAIL·잃은 PASS 없음 |
| AC-GDP-014 | REQ-GDP-014 | MUST-PASS | 활성 | `.toml` 재생성, `agents-emit-check` exit 0, `.toml` 두 절이 템플릿 `manager-git.md` 와 같음 |
| AC-GDP-015 | REQ-GDP-015 | MUST-PASS | 활성 | 템플릿 추가 줄에 SPEC ID·REQ 토큰·날짜·SHA·`CLAUDE.local` 없음 |
| AC-GDP-016 | 없음 — 절차 점검(plan.md §D 제약) | SHOULD-PASS | 활성 (0.2.5 재작성, 0.2.6 범위 수리) | 이 카드의 커밋(`HEAD --not develop`, 병합 제외) 가운데 `agent-common-protocol.md` 두 사본을 바꾼 커밋이 0개(양성 대조: 같은 형태가 `$BASE --not b412f8a33` 에서 2커밋을 찾음). 병합 전 전용 |
| AC-GDP-017 ~ AC-GDP-024 | — | — | 카드 t658로 이동 | spec.md §G.1 표 |
| AC-GDP-025 | REQ-GDP-024 | MUST-PASS | 활성 (파일 확장) | 범위 파일의 `[ZONE:Frozen]` 줄과 등록 Frozen clause 불변 |
| AC-GDP-026 | REQ-GDP-025 | MUST-PASS | 활성 (조각 확장) | 플래그 표면 아홉 조각이 `--auto-merge` 를 노출하고, `--merge` 를 네 조각에 남긴 채 `--auto-merge` 의 폐기된 별칭으로만 서술(방향 검사, 읽기 기록) |
| AC-GDP-027 | REQ-GDP-025 | MUST-PASS | 활성 (조각 확장) | `--no-merge` 는 폐기된 no-op으로만 서술되고 병합 조건·동작(건너뜀·끔·덮어씀 등)에 쓰이지 않음 |
| AC-GDP-028 | REQ-GDP-026 | MUST-PASS | 활성 | team 모드: `--auto-merge` 가 전원 승인(all … approvals) 조건과 함께 적힘, 승인 없음·부분 승인 문장 없음, 읽기 기록 |
| AC-GDP-029 | REQ-GDP-026 | MUST-PASS | 활성 | personal·manual 모드: `--auto-merge` 가 승인 조건 없이 병합한다고 적힘, 승인·리뷰를 요구하는 문장 없음, 읽기 기록(028과 공유) |
| AC-GDP-030 | REQ-GDP-014 | MUST-PASS | 활성 (신규) | 명령 원본 편집 뒤 `commands-emit`·`commands-emit-check` exit 0, 게시본 변화 여부를 두 경우 모두 판정, 바뀐 게시본은 원본과 같은 커밋 |

판정 대상: 16개(AC-GDP-001~006, 013~016, 025~030). AC-GDP-016은 요구사항 추적 밖의 절차 점검이다.

**감사가 지목한 잘못된 구현과 이를 잡는 기준**

| 잘못된 구현 | 잡는 기준 |
|---|---|
| `--merge` 를 별개이거나 폐기되지 않은 auto-merge 플래그로 서술 (결정 목록 줄, X1 사용법 줄, X2 `argument-hint`, X3 다음 단계 선택지 포함) | AC-GDP-026 (ii) 방향 검사와 읽기 기록 — 그리고 `--auto-merge` 누락은 (i) |
| 별칭 방향을 뒤집음(`--auto-merge` 를 폐기됐다고 적거나 `--auto-merge` 를 `--merge` 의 별칭으로 서술) | AC-GDP-026 (ii) 방향 검사와 읽기 기록 |
| `--merge` 를 모든 표면에서 지움(폐기된 별칭을 문서화하지 않음) | AC-GDP-026 (iii) |
| `--no-merge` 가 여전히 동작을 바꿈(건너뜀·트리거 조건, "no-op" 이라 적고 끔·덮어씀) | AC-GDP-027 (i)·(ii) |
| 워크트리 문맥에서 여전히 기본 병합 | AC-GDP-006 (확장 검출식), AC-GDP-027 (ii) 트리거 조건 |
| team 모드가 승인 없이 병합 | AC-GDP-028 (a)·(b) |
| personal·manual 모드가 승인을 요구 | AC-GDP-029 (a)·(b) |
| 명령 원본을 고치고 게시본을 다시 만들지 않거나 다른 커밋에 넣음 | AC-GDP-030 |

## §D.1 Given-When-Then과 판정 명령

### AC-GDP-001 — fetch 와 rev-list 가 같은 배치로 묶이지 않고 순서가 지시됨

```
GIVEN manager-git.md 로컬·템플릿 사본과 생성물 manager-git.toml 의 "## Synchronization" 절
WHEN 절을 빈 줄로 나눈 문단(목록 묶음은 한 문단)마다 검사하면
THEN fetch 와 rev-list 를 함께 담은 문단이 1개 이상 있고
 AND 그중 배치·병렬 낱말을 담으면서 순서 낱말이 없는 문단이 0개이고
 AND 읽기 기록 $E/ac001-reading.md 가 존재하며, fetch 와 rev-list 를 함께 담은 모든 문단에 대해
     (1) fetch 가 끝난 뒤 rev-list 를 실행한다고 말하는지
     (2) fetch 와 rev-list 를 같은 배치·같은 목록·표·병렬 묶음에 넣지 않는지
     를 문단마다 예/아니오로 답하고 모두 예다
```

판정은 줄이 아니라 문단 단위다. 자동 검출은 보조이고, 읽기 기록이 PASS의 전제다(§D.3).

대조(기준 트리):

```bash
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' $E/base-manager-git.md > $E/ac001-base-sync.md
awk 'BEGIN{RS=""} /fetch/ && /rev-list/ {c++} END{print c+0}' $E/ac001-base-sync.md > $E/ac001-base-paras.txt
# 기대: 1
awk 'BEGIN{RS=""; ORS="\n\n"} /fetch/ && /rev-list/ && /(batch|para[l]lel|single-turn|multi-Bash|independent)/ && !/(first|before|once|after|completes|wait)/' $E/ac001-base-sync.md > $E/ac001-base-autofail.md
test -s $E/ac001-base-autofail.md
# 기대: exit 0 — 자동 실패 검출기는 기준 트리에서 빨강(로컬 158행 문단, 템플릿 156행)
```

판정(로컬·템플릿·생성물 각각 — 아래는 로컬·생성물 명령, 템플릿은 경로만 `$T/.claude/agents/moai/manager-git.md` 로 바꾼다):

```bash
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' .claude/agents/moai/manager-git.md > $E/ac001-local-sync.md
awk 'BEGIN{RS=""} /fetch/ && /rev-list/ {c++} END{print c+0}' $E/ac001-local-sync.md > $E/ac001-local-paras.txt
# 기대: 1 이상
awk 'BEGIN{RS=""; ORS="\n\n"} /fetch/ && /rev-list/' $E/ac001-local-sync.md > $E/ac001-local-paras.md
awk 'BEGIN{RS=""; ORS="\n\n"} /fetch/ && /rev-list/ && /(batch|para[l]lel|single-turn|multi-Bash|independent)/ && !/(first|before|once|after|completes|wait)/' $E/ac001-local-sync.md > $E/ac001-local-autofail.md
test -e $E/ac001-local-autofail.md
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
test -s $E/ac001-local-autofail.md
# 기대: exit 1
awk 'BEGIN{RS=""; ORS="\n\n"} /(^|\n)[-*|] [^\n]*fetch/ && /(^|\n)[-*|] [^\n]*rev-list/' $E/ac001-local-sync.md > $E/ac001-local-listgroup.md
# 비어 있지 않으면 목록·표 안에 fetch 와 rev-list 가 함께 있다는 뜻 — 읽기 기록 (2)에서 반드시 판정
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' $T/.codex/agents/moai/manager-git.toml > $E/ac001-toml-sync.md
test -s $E/ac001-toml-sync.md
# 기대: exit 0 — 절이 비면 판정 불가
awk 'BEGIN{RS=""} /fetch/ && /rev-list/ {c++} END{print c+0}' $E/ac001-toml-sync.md > $E/ac001-toml-paras.txt
# 기대: 1 이상
awk 'BEGIN{RS=""; ORS="\n\n"} /fetch/ && /rev-list/ && /(batch|para[l]lel|single-turn|multi-Bash|independent)/ && !/(first|before|once|after|completes|wait)/' $E/ac001-toml-sync.md > $E/ac001-toml-autofail.md
test -e $E/ac001-toml-autofail.md
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
test -s $E/ac001-toml-autofail.md
# 기대: exit 1
test -s $E/ac001-reading.md
# 기대: exit 0 — 읽기 기록이 없으면 PASS 불가
```

plan 작성 시점 측정: 기준 트리 절 문단 1·autofail 1·목록 묶음 0(0.2.5 에서 BASE 로컬 반출본에 다시 잼: 문단 1, autofail `test -s` exit 0 — `.moai/reports/t622/reanchor/ac001-base-sync.md`). 기준 트리 `.toml` 의 `## Synchronization` 절은 템플릿 `manager-git.md` 의 같은 절과 diff exit 0(11줄)이므로 같은 값을 낸다.

뮤턴트 재실행:

```
"`git fetch` then … all in parallel"                               → 문단 1, autofail 1 → FAIL
2회차 목록 뮤턴트                                                   → autofail 0, listgroup 1 → 읽기 기록 (2)에서 FAIL
감사 뮤턴트 "… ONE single-turn multi-Bash batch after the checkpoint" → autofail 0 → 읽기 기록 (1)·(2)에서 FAIL
감사 표 뮤턴트(| remote | `fetch` | / | divergence | `rev-list …` |)   → autofail 0, listgroup(표 행 포함) 비어 있지 않음 → 읽기 기록에서 FAIL
2회차 올바른 문장                                                   → autofail 0, listgroup 0 → 빨강 아님
```

### AC-GDP-002 — Pre-Spawn 코드 블록의 순서와 Pre-Spawn 절 보존 (0.2.5: 회귀 방지 판정, 0.2.6: 두 부분)

```
GIVEN agent-common-protocol.md 로컬·템플릿 사본의 Pre-Spawn Sync Check 절과 그 절의 bash 코드 블록,
      그리고 고정 기준 사본($BASE 반출본 — 스냅숏 신선도 점검이 빈 결과일 때)의 같은 절
WHEN 이 카드의 run-phase 가 끝난 뒤 블록과 절을 추출해 검사하면
THEN (a) 순서: 주석 줄을 뺀 블록에서 "git fetch" 를 담은 줄과
         "git rev-list --count --left-right origin/main...HEAD" 를 담은 줄이 각각 정확히 1개이고,
         둘이 한 줄에 있으면 그 줄은 fetch 뒤에 ";" 또는 "&&" 로 rev-list 가 이어지는 한 줄이며,
         둘이 다른 줄에 있으면 fetch 줄은 "git fetch origin main"(뒤에 " 2>&1" 가능)으로 줄이 끝나고
         (뒤에 "&"·"|" 없음) 그 바로 다음 코드 줄이 fetch 의 완료 상태를 읽는 "fetch_status=$?" 이며
         rev-list 줄이 그보다 뒤에 있다
 AND (b) 보존: 두 사본의 Pre-Spawn 절 전체("### Pre-Spawn Sync Check" 제목부터 "### Pre-Edit Sync Check" 제목까지)가
         기준 절과 diff exit 0 이고, 그 절의 두 해석 표 행이 기준 표 행과 diff exit 0 이다
```

**(a)는 REQ-GDP-002 의 순서 문장을, (b)는 나머지 두 문장(세 번째 명령의 병렬 허용, 두 해석 표 불변)과 착지한 절 전체를 판정한다.** (a)의 모양 규칙:

- 한 줄 모양은 `;`·`&&` 만 받는다. 두 연산자는 fetch 가 끝난 뒤에 rev-list 를 시작시킨다. `&`(백그라운드)나 `|`(파이프)로 이으면 fetch 가 끝나기 전에 rev-list 가 돌 수 있어 FAIL 이다. spec.md §A.2 의 Pre-Edit 한 줄 형태가 "정상" 인 것과 같은 이유다.
- 다른 줄 모양은 fetch 줄이 줄 끝에서 끝나야 한다(뒤에 붙은 ` &`·`| tee …` 는 FAIL — 재고정 감사 뮤턴트 v·vi). 그리고 블록이 fetch 의 완료를 바로 다음 코드 줄에서 읽어야 한다(`fetch_status=$?`). 두 명령이 그냥 나란히 놓인 블록은 원래 결함(AC-11) 모양 그대로다 — 블록이 fetch 의 완료를 소비해야 rev-list 가 독립 병렬 항목으로 떨어져 나갈 수 없다. 이 규칙이 착지한 모양과 `b412f8a33` 모양(뮤턴트 i)을 가르는 기계 표지다. 산문으로만 순서를 적은 다른 모양은 (a)가 받지 않는다. 이 카드는 이 파일을 고치지 않으므로 그런 모양은 흡수로만 들어오고, 그때는 스냅숏 신선도 점검이 먼저 잡는다.
- 착지한 블록의 추가 성질(실패 시 `exit` 로 멈춤)은 REQ-GDP-002 가 요구하지 않는다(spec.md REQ-GDP-002 의 0.2.6 주석). 판정기는 그 성질을 `shape` 로 따로 찍는다 — 빨강이 났을 때 무엇이 사라졌는지 가리키는 진단이며, 보존 판정은 (b)가 한다.

**성격 (0.2.5, 0.2.6 보강).** 이 기준은 `$BASE` 에서 이미 초록이다 — develop 카드 t635 가 Lane A/B 형태로 고쳤고 dr0911 이 템플릿에 미러했으며, 리드 판단 (a)에 따라 이 카드는 이 파일을 편집하지 않는다. `.claude/rules/moai/development/verification-completeness.md` §2 기준으로 RED-now 셀이 없으므로 이 기준은 **이 카드의 작업을 재지 않는다**. 그 대신 이 카드의 작업(M1~M4 편집, 흡수, 생성물 재생성)이 끝난 뒤에도 REQ-GDP-002 의 성질이 남아 있는지 보는 **회귀 방지 판정**이다.

**MUST-PASS 근거 (0.2.6, 재고정 감사 D4).** §2 의 두 칸 채택은 이 기준의 근거가 아니다 — RED-now 셀이 없어 §2 로는 채택되지 않고, §2.1 의 판정 불가 처리는 그런 기준을 regression-guard 로 두고 출시 차단 자격을 뺀다. 이 기준의 근거는 §4 증거 고정이다. (b)는 고정 트리 `$BASE` 에 핀한 보존 단언(preserved-surface)으로 AC-GDP-003 과 같은 성격이다. (b)가 빨강이면 이 카드가 바꾸면 안 되는 항상 로드되는 규칙이 카드 트리에서 달라졌다는 뜻이고, 원인이 카드 편집(AC-GDP-016)이든 흡수(스냅숏 신선도 점검)든 카드 착지를 막아야 하므로 MUST-PASS 로 둔다. (a)의 채택 근거는 §1.1 의 관측된 실패다 — 아래 뮤턴트에서 FAIL, 현재 두 블록에서 PASS 를 관측했다. green path: 이 카드는 파일을 고치지 않으므로 M5·M6 에서 두 사본 모두 `order=PASS` 와 절 diff exit 0 이 그대로 나오는 것이 기대값이다. 빨강이 나면 원인을 이 카드 편집과 흡수 중에서 가린다(plan.md M5).

블록 추출(대조·판정 공통, 0.2.4 와 같음):

```bash
awk '/^### Pre-Spawn Sync Check/{s=1} s && /^```bash/{b=1; next} b && /^```/{exit} b' <파일> > <블록 파일>
```

순서 판정 검출기(대조·판정·뮤턴트 공통 — 한 줄로 실행한다. 0.2.6 에서 재고정 감사 D2 의 보강안을 반영해 다시 썼다). 주석 줄은 건너뛴다. `fetch`=`git fetch` 를 담은 줄 수(모양과 무관), `revlist`=rev-list 줄 수, `joined`=fetch 와 rev-list 를 한 줄에 담은 줄 수, `J`=그 한 줄이 `;`·`&&` 로 이은 순서 모양일 때 그 줄, `F`=줄 끝까지 고정한 fetch 줄(`git fetch origin main` 뒤에 ` 2>&1` 만 허용 — ` &`·`| …` 가 붙으면 `F=0`), `S`=`F` 바로 다음 코드 줄이 `fetch_status=$?` 일 때 그 줄, `C`=`fetch_status` 와 `-ne 0` 을 담은 첫 줄, `X`=`C` 뒤 첫 `exit` 명령 줄(`exit` 뒤가 공백이나 줄 끝 — `exit_note=` 같은 낱말은 걸리지 않음), `R`=rev-list 줄.

- `order=PASS` (판정 (a)) — `fetch=1`·`revlist=1` 이고, 한 줄 모양(`joined=1` 이며 `J` 있음) 또는 다른 줄 모양(`joined=0`, `F`·`S` 있음, `R > S`).
- `shape=PASS` (진단) — `F < S < C < X < R`. 착지한 블록의 상태 관측·실패 시 exit 순서다. 판정 조건이 아니다.

```bash
awk '/^[[:space:]]*#/ {next} /[g]it fetch/ {nf++} /[g]it rev-list --count --left-right origin\/main[.][.][.]HEAD/ {nr++; if(!R) R=NR} /[g]it fetch/ && /[g]it rev-list/ {j++; if ($0 ~ /^[[:space:]]*[g]it fetch origin main( 2>&1)?[[:space:]]*(;|&&)[[:space:]]*[g]it rev-list --count --left-right origin\/main[.][.][.]HEAD[[:space:]]*$/) J=NR} /^[[:space:]]*[g]it fetch origin main( 2>&1)?[[:space:]]*$/ {if(!F) F=NR; w=1; next} w && NF {w=0; if ($0 ~ /^[[:space:]]*fetch_status=[$][?][[:space:]]*$/) S=NR} /fetch_status/ && /-ne 0/ && !C {C=NR} /^[[:space:]]*exit([[:space:]]|$)/ && C && !X {X=NR} END {o="FAIL"; if (nf==1 && nr==1 && ((j==1 && J) || (j==0 && F && S && R>S))) o="PASS"; s="FAIL"; if (F && S>F && C>S && X>C && R>X) s="PASS"; printf "fetch=%d revlist=%d joined=%d J=%d F=%d S=%d C=%d X=%d R=%d order=%s shape=%s\n",nf,nr,j,J,F,S,C,X,R,o,s}' <블록 파일> > <판정 파일>
```

대조(기준 트리 — `$BASE` 에서 이미 초록이므로 대조는 "검출기가 옳은 블록을 받는가"를 본다):

```bash
awk '/^### Pre-Spawn Sync Check/{s=1} s && /^```bash/{b=1; next} b && /^```/{exit} b' $E/base-agent-common-protocol.md > $E/ac002-base-block.md
# 위 순서 판정 검출기를 $E/ac002-base-block.md 에 실행 → $E/ac002-base-order.txt
/usr/bin/grep -c 'order=PASS shape=PASS$' $E/ac002-base-order.txt > $E/ac002-base-order-pass.txt
# 기대: 1 (0.2.6 측정: fetch=1 revlist=1 joined=0 J=0 F=2 S=3 C=4 X=6 R=8 order=PASS shape=PASS — 기준·로컬·템플릿 블록 모두,
#       .moai/reports/t622/reanchor-fix/ac002/judge-results.txt)
/usr/bin/grep -c 'moai session list --json --filter-spec=' $E/ac002-base-block.md > $E/ac002-base-session.txt
# 기대: 1
sed -n '/^### Pre-Spawn Sync Check/,/^### Pre-Edit Sync Check/p' $E/base-agent-common-protocol.md > $E/ac002-base-section.md
test -s $E/ac002-base-section.md
# 기대: exit 0 — 기준 절이 비면 (b)의 diff 가 공허하다(0.2.6 측정: 52줄, 로컬·템플릿 기준 절끼리 diff exit 0)
```

판정(로컬·템플릿 각각 — 템플릿은 경로만 `$T/.claude/rules/moai/core/agent-common-protocol.md` 로, 출력 이름은 `ac002-template-*` 로 바꾼다):

```bash
awk '/^### Pre-Spawn Sync Check/{s=1} s && /^```bash/{b=1; next} b && /^```/{exit} b' .claude/rules/moai/core/agent-common-protocol.md > $E/ac002-local-block.md
test -s $E/ac002-local-block.md
# 기대: exit 0 — 블록이 비면 판정 불가
# 위 순서 판정 검출기를 $E/ac002-local-block.md 에 실행 → $E/ac002-local-order.txt
test -s $E/ac002-local-order.txt
# 기대: exit 0 — 판정 줄이 없으면 검출기가 돌지 않은 것이다(판정 불가, PASS 아님)
/usr/bin/grep -c ' order=PASS ' $E/ac002-local-order.txt > $E/ac002-local-order-pass.txt
# 판정 (a) 기대: 1 — 개수가 파일에 찍히지 않으면 판정 불가. shape 값은 기록만 한다(진단)
/usr/bin/grep -c 'moai session list --json --filter-spec=' $E/ac002-local-block.md > $E/ac002-local-session.txt
# 진단 기대: 1
sed -n '/^### Pre-Spawn Sync Check/,/^### Pre-Edit Sync Check/p' .claude/rules/moai/core/agent-common-protocol.md > $E/ac002-local-section.md
test -s $E/ac002-local-section.md
# 기대: exit 0 — 시작 표지가 사라져 절이 비면 판정 불가
diff $E/ac002-base-section.md $E/ac002-local-section.md > $E/ac002-local-section.diff
# 판정 (b) 기대: exit 0 — Pre-Spawn 절 전체가 기준 절과 같다(Lane A/B 문장, 블록, 상태 관측·실패 시 exit, Lane B 병렬 허용 주석, 두 해석 표).
#               이 카드는 이 파일을 고치지 않으므로 절 전체 보존이 곧 REQ-GDP-002 의 불변 문장 보존이다
/usr/bin/grep -E '^[|] ' $E/ac002-base-section.md > $E/ac002-base-matrix.txt
/usr/bin/grep -E '^[|] ' $E/ac002-local-section.md > $E/ac002-local-matrix.txt
test -s $E/ac002-base-matrix.txt
# 기대: exit 0 — 기준 표 행이 비면 아래 비교가 공허하다
diff $E/ac002-base-matrix.txt $E/ac002-local-matrix.txt > $E/ac002-matrix.diff
# 판정 (b) 기대: exit 0 (절 diff 의 부분집합이지만 표 불변을 따로 가리킨다. 기준 트리의 표 행은 8개. 0.2.5 측정: `$BASE` 두 사본의 표 행 8개가 `b412f8a33` 의 8개와 diff exit 0 — t635 는 표를 바꾸지 않았다)
```

템플릿 사본의 (b) 기준 절은 `$E/base-agent-common-protocol-template.md`(plan.md §C 2단계에 더한 반출본 — `git show $BASE:$T/.claude/rules/moai/core/agent-common-protocol.md`)에서 같은 `sed` 로 뽑는다. 0.2.6 측정: 로컬·템플릿 기준 절 diff exit 0, 현재 로컬·템플릿 절과 각 기준 절 diff exit 0(`.moai/reports/t622/reanchor-fix/ac002/section-*.md`·`section-*.diff`).

뮤턴트(0.2.6, `.moai/reports/t622/reanchor-fix/ac002/` 에서 실행 — 레인이 같은 명령으로 다시 돌린다). 현재 블록 `B` = `$BASE` 로컬 사본의 블록(12줄: 1 주석, 2 fetch, 3 `fetch_status=$?`, 4-7 검사와 exit, 8 rev-list, 9 빈 줄, 10-11 주석, 12 session list). 블록 뮤턴트 i~viii 는 (b)용 절 픽스처도 만든다 — 기준 절의 bash 블록 줄을 뮤턴트 블록으로 바꿔 끼운 절(`<뮤턴트>.section.md`, 현재 블록을 끼운 대조 `ctl-section-current.md` 는 기준 절과 `cmp` exit 0).

```
(i)    옛 블록 — git show b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0:.claude/rules/moai/core/agent-common-protocol.md 반출본에서 같은 awk 로 추출(9줄: 단독 fetch, 단독 rev-list, 상태 검사 없음)
(ii)   fetch 와 rev-list 를 ";" 로 한 줄에 — awk 'NR==FNR{if(FNR==8) r=$0; next} FNR==2{print $0 "; " r; next} FNR==8{next} {print}' B B
(iii)  상태 검사 삭제 — awk 'NR>=3 && NR<=7 {next} {print}' B
(iv)   rev-list 를 fetch 앞으로 — awk 'NR==FNR{if(FNR==8) r=$0; next} FNR==1{print; print r; next} FNR==8{next} {print}' B B
(v)    fetch 를 백그라운드로 — 2행 끝에 " &"            (재고정 감사 reanchor-audit/mut-v-background.md)
(vi)   fetch 를 파이프로 — 2행 끝에 " | tee <파일>"      (reanchor-audit/mut-vi-pipe.md)
(vii)  실패 시 멈추지 않음 — exit 줄을 exit_note="…" 로  (reanchor-audit/mut-vii-noabort.md)
(viii) Lane B 병렬 허용 삭제 — Lane B 주석을 "MUST NOT start until Lane A has fully finished" 로 (reanchor-audit/mut-viii-laneb-serial.md)
(ix)   절 밖 뒤집기 — Pre-Spawn 절 앞 "### Read-only verification batching" 문단 뒤에 "Lane A 의 fetch 와 rev-list 는 독립 읽기 검증이라 한 턴에 병렬 Bash 호출로 내도 된다" 는 문장을 넣은 파일 전체 픽스처 (mut-ix-outside-override-acp.md)
관측(순서 판정 검출기 judge-results.txt, 절 diff section-results.txt, 뮤턴트 ix 는 파일 전체 픽스처에서 같은 추출):
                   판정기 출력                                                             (a)    (b) 절 diff  AC-GDP-002
  현재 로컬 블록   fetch=1 revlist=1 joined=0 J=0 F=2 S=3 C=4 X=6 R=8 order=PASS shape=PASS   PASS   exit 0       PASS
  현재 템플릿 블록 fetch=1 revlist=1 joined=0 J=0 F=2 S=3 C=4 X=6 R=8 order=PASS shape=PASS   PASS   exit 0       PASS
  (i)    fetch=1 revlist=1 joined=0 J=0 F=2 S=0 C=0 X=0 R=5 order=FAIL shape=FAIL              FAIL   exit 1       FAIL
  (ii)   fetch=1 revlist=1 joined=1 J=2 F=0 S=0 C=4 X=6 R=2 order=PASS shape=FAIL              PASS   exit 1       FAIL
  (iii)  fetch=1 revlist=1 joined=0 J=0 F=2 S=0 C=0 X=0 R=3 order=FAIL shape=FAIL              FAIL   exit 1       FAIL
  (iv)   fetch=1 revlist=1 joined=0 J=0 F=3 S=4 C=5 X=7 R=2 order=FAIL shape=FAIL              FAIL   exit 1       FAIL
  (v)    fetch=1 revlist=1 joined=0 J=0 F=0 S=0 C=4 X=6 R=8 order=FAIL shape=FAIL              FAIL   exit 1       FAIL
  (vi)   fetch=1 revlist=1 joined=0 J=0 F=0 S=0 C=4 X=6 R=8 order=FAIL shape=FAIL              FAIL   exit 1       FAIL
  (vii)  fetch=1 revlist=1 joined=0 J=0 F=2 S=3 C=4 X=0 R=8 order=PASS shape=FAIL              PASS   exit 1       FAIL
  (viii) fetch=1 revlist=1 joined=0 J=0 F=2 S=3 C=4 X=6 R=8 order=PASS shape=PASS              PASS   exit 1       FAIL
  (ix)   fetch=1 revlist=1 joined=0 J=0 F=2 S=3 C=4 X=6 R=8 order=PASS shape=PASS              PASS   exit 0       PASS (아래 한계)
(a) 탐침(판정 모양 확인, 뮤턴트 아님): 한 줄 "fetch … & rev-list …" → joined=1 J=0 order=FAIL. 한 줄 "fetch … && rev-list …" → joined=1 J=2 order=PASS.
```

(ii)는 REQ-GDP-002 의 순서 문장을 어기지 않는다(`;` 는 fetch 가 끝난 뒤 rev-list 를 시작시킨다) — (a) PASS 가 맞고, 착지 블록이 바뀐 것을 (b)가 잡는다. (vii)도 같다 — 실패 시 멈춤은 REQ-GDP-002 의 요구가 아니고 착지 블록의 성질이라 (b)가 잡는다. (viii)은 REQ-GDP-002 의 "세 번째 명령의 병렬 실행 허용" 문장을 어기며 (b)가 잡는다. 0.2.5 판정기는 (v)~(viii) 네 개를 모두 PASS 로 읽었다(재고정 감사 `reanchor-audit/judge-new-mutants.txt`).

**다시 시도한 뮤턴트 — 요구를 어기면서 통과하는 것 (0.2.6).** 절 안의 어떤 편집도 (b)가 잡으므로, 통과하려면 절을 그대로 두어야 한다. 시도한 모양(`reanchor-fix/ac002/pass-attempts.txt`): 절 제목을 바꿈(절 추출 0줄, (b) exit 1), 파일 뒤쪽에 두 번째 `### Pre-Spawn Sync Check` 절을 넣음(`sed` 범위가 두 범위를 모두 찍어 55줄, (b) exit 1), 이 절의 bash 펜스를 ` ```sh ` 로 바꿈(블록 추출이 Pre-Edit 절의 블록으로 넘어가 그 `;` 한 줄로 (a)는 PASS 지만 (b) exit 1), 템플릿 사본만 바꿈(템플릿 (b) exit 1). 통과한 것은 (ix) 하나다 — 절 **밖**에 순서를 뒤집는 문장을 넣으면 블록과 절이 모두 그대로라 (a)·(b) 모두 PASS 다. REQ-GDP-002 는 "Pre-Spawn Sync Check 코드 블록" 을 묶으므로 글자로는 어기지 않지만, 에이전트가 받는 지시로는 순서를 뒤집는다. AC-GDP-002 는 이 모양을 보지 않는다. 막는 곳은 둘이다: 이 카드의 커밋이 넣으면 AC-GDP-016(두 사본을 바꾼 카드 커밋 0개), 흡수로 들어오면 스냅숏 신선도 점검(`agent-common-protocol.md` 가 목록에 나옴 → 리드 보고). 같은 픽스처를 기준 파일 전체와 `diff` 하면 exit 1 이다(`mut-ix-wholefile.diff`) — 파일 전체 비교를 AC-GDP-002 에 넣지 않은 것은 그 비교가 REQ-GDP-002 가 아니라 "이 카드는 이 파일을 고치지 않는다" 를 재기 때문이며, 그 성질은 AC-GDP-016 의 몫이다.

0.2.4 의 판정(단독 fetch 줄 0·단독 rev-list 줄 0·결합 줄 1)은 "한 명령" 모양을 요구해 t635 형태를 FAIL 로 읽으므로 0.2.5 에서 버렸다. 한계: 판정기는 줄 단위다. `exit` 가 실제로 `if` 블록 안에 있는지, 검사 줄이 `fetch_status` 를 올바른 방향으로 비교하는지까지는 보지 않는다 — (b)가 절 전체를 기준과 비교하므로 이 카드가 고치지 않는 한 문제되지 않지만, 기준 절을 `$CARD_BASE` 에서 다시 반출한 경우에는 한 번 읽어 판정 기록에 적는다.

### AC-GDP-003 — Pre-Edit Sync Check 절 불변

```
GIVEN 기준 트리와 편집 뒤의 Pre-Edit Sync Check 절(제목부터 #### The sweep prohibition 앞까지)
WHEN 두 절을 추출해 비교하면
THEN 차이가 없다
```

대조(0.2.5 — 이 카드가 `agent-common-protocol.md` 를 고치지 않으므로 0.2.4 의 "편집 뒤 Pre-Spawn 절 diff exit 1" 대조는 성립하지 않는다): 같은 `sed` 추출 + `diff` 형태가 바뀐 절을 실제로 잡는지, `b412f8a33` 반출본과 `$BASE` 의 Pre-Spawn 절로 확인한다 — `sed -n '/^### Pre-Spawn Sync Check/,/^### Pre-Edit Sync Check/p' $E/b412-agent-common-protocol.md > $E/ac003-control-b412-section.md` 뒤 `diff $E/ac003-control-b412-section.md $E/ac002-base-section.md > $E/ac003-control.diff` → exit 1 기대(0.2.5 측정 exit 1 — t635 가 절을 다시 썼다, `reanchor/ac003-control-section.diff`). Pre-Edit 절 자체는 `b412f8a33` 과 `$BASE` 에서 diff exit 0(35줄, `reanchor/ac003-b412-vs-base.diff` 빈 파일)이다.

```bash
sed -n '/^### Pre-Edit Sync Check/,/^#### The sweep prohibition/p' $E/base-agent-common-protocol.md > $E/ac003-base.md
sed -n '/^### Pre-Edit Sync Check/,/^#### The sweep prohibition/p' .claude/rules/moai/core/agent-common-protocol.md > $E/ac003-local.md
diff $E/ac003-base.md $E/ac003-local.md > $E/ac003.diff
# 기대: exit 0 (템플릿 사본도 같은 방식)
```

### AC-GDP-004 — `delivery.md` 병합 명령 해석

```
GIVEN delivery.md 로컬·템플릿 사본의 Step 3.4 Auto-Merge Behavior
WHEN 편집 뒤 검사하면
THEN "gh pr merge --squash --delete-branch" 가 0회이고
 AND "gh pr merge --<merge_method> --delete-branch" 가 2회 이상이며
 AND merge_method 의 해석 출처(git_strategy 모드 설정, 기본 squash)가 적혀 있다
```

```bash
/usr/bin/grep -c 'gh pr merge --squash --delete-branch' $E/base-delivery.md > $E/ac004-control-squash.txt
# 대조 기대: 2 (로컬 반출본 368·380행, 템플릿 343·355행 — 0.2.5 측정 2)
/usr/bin/grep -n 'merge_method' $E/base-delivery.md > $E/ac004-control-source.txt
# 대조 기대: exit 1
/usr/bin/grep -c 'gh pr merge --squash --delete-branch' .claude/skills/moai/workflows/sync/delivery.md > $E/ac004-local-squash.txt
# 기대: 0
/usr/bin/grep -c 'gh pr merge --<merge_method> --delete-branch' .claude/skills/moai/workflows/sync/delivery.md > $E/ac004-local-resolved.txt
# 기대: 2 이상
/usr/bin/grep -n -E 'merge_method.*(squash|default)' .claude/skills/moai/workflows/sync/delivery.md > $E/ac004-local-source.txt
# 기대: exit 0, 적중 줄이 해석 출처를 설명하는지 읽어 기록
```

템플릿 사본도 같은 방식으로 판정한다.

### AC-GDP-005 — 고정 `--squash` 병합 예시 부재, 기본값 설명 유지

```
GIVEN 범위 파일 아홉 개(로컬·템플릿)와 생성물 manager-git.toml
WHEN 편집 뒤 --squash 가 붙은 gh pr merge 를 모두 찾으면
THEN manager-git.md 와 .toml 에서만 1줄씩 나오고, 그 줄은 기본값 설명 문장이며
 AND 나머지 여덟 파일에서는 0줄이다
```

대조(기준 트리):

```bash
/usr/bin/grep -n -E 'gh pr merge[^|]*--squash' $E/base-manager-git.md $E/base-delivery.md $E/base-manager-git.toml > $E/ac005-control.txt
# 기대: 6줄 — manager-git 34·116(로컬 반출본; 템플릿 32·114), delivery 368·380(로컬 반출본; 템플릿 343·355), .toml 26·108 (0.2.5 측정 6줄)
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' $E/base-agent-common-protocol.md $E/base-doc-execution.md $E/base-skill.md $E/base-reference.md $E/base-qgc.md $E/base-sync.md $E/base-command-sync.md > $E/ac005-control-zero.txt
# 기대: 파일마다 0
```

판정(로컬·템플릿 각각, 파일마다 따로 셈):

```bash
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' .claude/agents/moai/manager-git.md > $E/ac005-mg.txt
# 기대: 1
/usr/bin/grep -c -F 'which under the squash default renders `gh pr merge --squash --delete-branch`' .claude/agents/moai/manager-git.md > $E/ac005-mg-default.txt
# 기대: 1
/usr/bin/grep -c -F 'gh pr merge <PR> --<merge_method> --delete-branch' .claude/agents/moai/manager-git.md > $E/ac005-mg-example.txt
# 기대: 1 — 114행 예시가 치환됨
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' .claude/skills/moai/workflows/sync/delivery.md .claude/rules/moai/core/agent-common-protocol.md .claude/skills/moai/workflows/sync/doc-execution.md .claude/skills/moai/SKILL.md .claude/skills/moai/references/reference.md .claude/skills/moai/workflows/sync/quality-gates-context.md .claude/skills/moai/workflows/sync.md .claude/commands/moai/sync.md > $E/ac005-others.txt
# 기대: 파일마다 0
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' $T/.codex/agents/moai/manager-git.toml > $E/ac005-toml.txt
# 기대: 1 (기본값 설명 문장)
```

1회차 감사 뮤턴트(`gh pr merge 42 --squash --delete-branch`, `gh pr merge --squash <PR> --delete-branch`)는 이 검출식에 걸려 FAIL로 판정된다.

### AC-GDP-006 — auto-merge 기본값 단일 기준 (OD-2 = B)

```
GIVEN delivery.md 로컬·템플릿 사본의 Step 3.4 절과 doc-execution.md 로컬·템플릿 사본의 "Worktree Context Detection" 소절
WHEN 편집 뒤 두 절을 추출해 검사하면
THEN 워크트리 문맥을 기본 병합과 묶는 문구(확장 검출식)가 두 절 모두 0개이고
 AND 두 절이 각각 manager-git.md 를 기준으로 1회 이상 이름으로 밝히고
 AND manager-git.md 의 옵트인 문장 두 개(148행, 166행 — 로컬 150행, 168행)의 --auto-merge 조건이 남아 있고
 AND 읽기 기록 $E/ac006-reading.md 가 존재하며, 두 절의 모든 문장에 대해 "워크트리 문맥만으로 병합이 일어난다고 말하는가" 에 아니오로 답하고, 병합 조건이 --auto-merge 로 적혀 있음을 확인한다
```

확장 검출식(대조·판정 공통):

```
default for worktree contexts|worktree contexts default|no-merge.{1,3}flag NOT set|merges? (automatically|by default)|auto-merge (is )?(the )?default|no-merge.{0,12}(absent|not set|missing|NOT set)
```

doc-execution 소절에는 기존 검출식 `default (to )?auto-merge|worktree contexts default` 도 함께 쓴다.

절 추출(대조·판정 공통):

```bash
awk '/^#### Step 3\.4/{s=1; print; next} s && /^(####|###) /{exit} s' <delivery.md 경로> > <delivery 절 파일>
awk '/^##### Worktree Context Detection/{s=1; print; next} s && /^#/{exit} s' <doc-execution.md 경로> > <doc-execution 절 파일>
```

대조(기준 트리):

```bash
awk '/^#### Step 3\.4/{s=1; print; next} s && /^(####|###) /{exit} s' $E/base-delivery.md > $E/ac006-base-dl.md
awk '/^##### Worktree Context Detection/{s=1; print; next} s && /^#/{exit} s' $E/base-doc-execution.md > $E/ac006-base-de.md
awk '/default for worktree contexts|worktree contexts default|no-merge.{1,3}flag NOT set|merges? (automatically|by default)|auto-merge (is )?(the )?default|no-merge.{0,12}(absent|not set|missing|NOT set)/ {print FNR}' $E/ac006-base-dl.md > $E/ac006-base-dl-default.txt
# 기대: 8·20 (로컬 반출본 파일 기준 362·374행, 템플릿 337·349행 — 0.2.5 측정 8·20)
/usr/bin/grep -n -i -E 'default (to )?auto-merge|worktree contexts default|merges? (automatically|by default)' $E/ac006-base-de.md > $E/ac006-base-de-default.txt
# 기대: exit 0 — 절 안 8행, 파일 기준 36행(두 사본 같음, 0.2.5 측정)
/usr/bin/grep -c 'manager-[g]it[.]md' $E/ac006-base-dl.md $E/ac006-base-de.md > $E/ac006-base-source.txt
# 기대: 파일마다 0
/usr/bin/grep -c -F 'Execute only with `--auto-merge` flag AND all approvals obtained' $E/base-manager-git.md > $E/ac006-base-optin1.txt
/usr/bin/grep -c -F 'Auto-merge: only with the `--auto-merge` flag' $E/base-manager-git.md > $E/ac006-base-optin2.txt
# 기대: 각각 1
```

판정(로컬·템플릿 각각):

```bash
awk '/^#### Step 3\.4/{s=1; print; next} s && /^(####|###) /{exit} s' .claude/skills/moai/workflows/sync/delivery.md > $E/ac006-local-dl.md
awk '/^##### Worktree Context Detection/{s=1; print; next} s && /^#/{exit} s' .claude/skills/moai/workflows/sync/doc-execution.md > $E/ac006-local-de.md
test -s $E/ac006-local-dl.md
test -s $E/ac006-local-de.md
# 기대: 둘 다 exit 0
awk '/default for worktree contexts|worktree contexts default|no-merge.{1,3}flag NOT set|merges? (automatically|by default)|auto-merge (is )?(the )?default|no-merge.{0,12}(absent|not set|missing|NOT set)/ {print FNR}' $E/ac006-local-dl.md > $E/ac006-local-dl-default.txt
test -e $E/ac006-local-dl-default.txt
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
test -s $E/ac006-local-dl-default.txt
# 기대: exit 1
/usr/bin/grep -n -i -E 'default (to )?auto-merge|worktree contexts default|merges? (automatically|by default)' $E/ac006-local-de.md > $E/ac006-local-de-default.txt
# 기대: exit 1
/usr/bin/grep -c 'manager-[g]it[.]md' $E/ac006-local-dl.md $E/ac006-local-de.md > $E/ac006-local-source.txt
# 기대: 파일마다 1 이상
/usr/bin/grep -c -e '--auto-merge' .claude/agents/moai/manager-git.md > $E/ac006-local-optin.txt
# 기대: 2 이상 (148·166행 옵트인 조건 유지 — 로컬 150·168행)
test -s $E/ac006-reading.md
# 기대: exit 0 — 읽기 기록이 없으면 PASS 불가
```

뮤턴트 재실행(스크래치 픽스처 4줄): "Merge policy follows manager-git.md." · "When the session runs inside a worktree and the no-merge flag is absent, merge once checks pass" · "Worktree sessions merge automatically." · "Auto-merge is not the default; pass `--auto-merge` per manager-git.md." → 확장 검출식이 2·3줄을 잡고 올바른 4줄은 잡지 않음.

### AC-GDP-007 ~ AC-GDP-012 — 자리표시

AC-GDP-007·008·009·010은 카드 t658로 옮겼고, AC-GDP-011·012는 0.1.2에서 철회됐다. 판정하지 않으며 원문은 커밋 `87988e946` 의 acceptance.md, 추적 표는 spec.md §G.1이다.

### AC-GDP-013 — 사본 일치와 의도된 차이 보존

```
GIVEN 범위 파일 아홉 개의 로컬·템플릿 사본과 게시본 moai-sync/SKILL.md 의 로컬·템플릿 사본
WHEN 편집 뒤 비교하면
THEN (a) agent-common-protocol.md, 게시본 SKILL.md 는 diff exit 0 이고
 AND (b) manager-git.md, delivery.md, doc-execution.md, quality-gates-context.md, moai/SKILL.md, references/reference.md, workflows/sync.md,
     명령 원본(sync.md 대 sync.md.tmpl)은 줄번호 머리를 뺀 차이 본문이 기준 트리($BASE)와 같고
 AND (c) 명령 원본 두 사본의 argument-hint 줄이 서로 같고
 AND (d) 범위 파일을 덮는 테스트 두 개(TestSanitizedPairParity, TestTemplateNoInternalContentLeak)가 각자 최상위 PASS 줄을 내며 exit 0 이고
 AND (e) 미러 테스트 비회귀: TestSanitizedPairParity·TestRuleTemplateMirrorDrift 실행의 FAIL 이 모두 $BASE 기준선 FAIL 집합 안에 있고,
     기준선 PASS 가 하나도 빠지거나 FAIL 로 바뀌지 않았다(개수가 아니라 집합으로 비교)
```

**0.2.5 기준선 재측정.** develop 흡수로 `$BASE` 에서 사본 관계가 바뀌었다(`.moai/reports/t622/absorb-ee99507fb.md` §2, `reanchor/pair-hunks.txt`). `manager-git.md`·`quality-gates-context.md` 는 바이트 동일 쌍에서 의도된 차이 쌍으로 옮겼고, `delivery.md`·`doc-execution.md`·`workflows/sync.md` 는 덩어리가 늘었다. (b)의 차이 본문 비교는 `$BASE` 의 차이를 기준으로 삼으므로, develop 이 로컬에만 넣은 차이는 이 카드가 만든 차이로 읽히지 않고, 이 카드가 한쪽 사본에만 넣은 편집은 새 본문 줄이 되어 잡힌다(아래 뮤턴트).

**파일별 사본 가드** (`$BASE` 에서 잰 L·T diff, 테스트 파일은 `b412f8a33..$BASE` 에서 바뀌지 않음):

| 범위 파일 | 기준 트리 L·T diff (`$BASE`) | 사본 가드 | 근거 |
|---|---|---|---|
| `agent-common-protocol.md` | exit 0 | `diff` + `TestSanitizedPairParity` | `sanitized_pair_parity_test.go:71`. 이 카드는 편집하지 않는다 |
| 게시본 `.agents/skills/moai-sync/SKILL.md` (로컬) 대 템플릿 게시본 | exit 0 | `diff` | 발행기는 템플릿 게시본만 쓴다(`golden_test.go` `templatesDir`). 로컬 사본은 추적 파일이다 |
| `manager-git.md` | exit 1 (`5,7c5` — 로컬 프론트매터 설명 +2줄) | 차이 본문 `diff` 만 | `rule_template_mirror_test.go` 주석이 바이트 동일 목록에서 제거를 명시 |
| `quality-gates-context.md` | exit 1 (`9,11c9,11`, `48,49c48,49`, `51c51`, `138c138`, `142,149c142`, `151c144`, `165c158`) | 차이 본문 `diff` 만 | 어떤 사본 테스트에도 없음. 편집 자리 30·101행은 덩어리 밖 |
| `delivery.md` | exit 1 (`9,11c9,11`, `49,54c49,50`, `154,163c150,152`, `169,188c158,163`, `300c275`, `303c278`, `447,448c422`, `450,458c424,427`, `510,511c479`) | 차이 본문 `diff` 만 | 어떤 사본 테스트에도 없음. 편집 자리(템플릿 330-356·396-404, 로컬 355-381·421-429)는 `303c278` 과 `447,448c422` 사이. X3 도 이 판정에 들어간다 |
| `doc-execution.md` | exit 1 (`9,11c9,11`, `79,91d78`, `118c105`, `128,134c115,116`, `144c126`, `156,161d137`, `173,174c149`, `180c155`, `182,189d156`) | 차이 본문 `diff` 만 | 어떤 사본 테스트에도 없음. 편집 자리 34-36행은 덩어리 밖 |
| `moai/SKILL.md` | exit 1 (20개 덩어리, `125c125` … `392d391` — `b412f8a33` 과 같음) | 차이 본문 `diff` 만 | `backlog_json_disclosure_mirror_test.go:24` 는 임베드 사본 = 템플릿 원본을 볼 뿐 로컬 사본을 보지 않는다 |
| `references/reference.md` | exit 1 (`229d228`) | 차이 본문 `diff` 만 | `agent_frontmatter_audit_test.go:407` 은 프론트매터만 본다 |
| `workflows/sync.md` | exit 1 (`29,31c29,31`, `65,74d64`, `81c71`) | 차이 본문 `diff` 만 | `agentless_audit_test.go:44` 는 사본 일치가 아닌 지침 내용을 본다. X1 사용법 줄(로컬 95 / 템플릿 85)도 이 판정에 들어간다 |
| 명령 원본 `.claude/commands/moai/sync.md` (로컬) 대 `sync.md.tmpl` (템플릿) | exit 1 (`2c2` — `description` 줄만 다름) | 차이 본문 `diff` + `argument-hint` 줄 비교 | 파일 형식이 다르다: 템플릿은 `moai init` 때 `ConversationLanguage` 로 렌더링되는 Go 템플릿이고 2행 `description` 이 로케일 조건문이다. 로컬은 렌더링된 영어 사본이다. X2가 고치는 3행에는 템플릿 액션이 없어 두 사본에서 글자 그대로 같아야 한다. `commandemit/golden_test.go` 는 템플릿 원본과 게시본의 관계를 보며 로컬 원본을 보지 않는다 |
| 템플릿 사본 전체 | — | `TestTemplateNoInternalContentLeak` (사본 일치가 아니라 템플릿 청결) | `internal_content_leak_test.go:1535`, 템플릿 루트 전체를 걷는다 |

`TestRuleTemplateMirrorDrift`·`TestLateBranchTemplateMirror` 는 허용 목록에 범위 파일이 하나도 없어 사본 일치 증거로 쓰지 않는다. (e)는 `TestRuleTemplateMirrorDrift` 를 사본 일치가 아니라 **비회귀** 점검으로만 돌린다 — 이 카드가 템플릿 트리를 바꾸는 동안 다른 사본 쌍을 깨지 않았는지를 본다. `TestHookWrapperCopiesStayIdentical`(`internal/hook/wrapper_copies_contract_test.go:73`)은 `.claude/hooks/moai` 의 훅 래퍼 스크립트 여섯 개와 그 템플릿 사본만 읽고, 이 카드의 파일(마크다운 지침·명령 원본·`.toml`·게시본)은 그 목록에 없어 넣지 않는다.

대조: 위 표의 기준 트리 diff 결과(0.2.5 측정, `reanchor/pair-*.diff`), 명령 원본 `argument-hint` 줄 L·T diff exit 0. 본문 비교 검출기는 빈 파일과 차이 본문을 `diff` 하면 exit 1.

판정 (a) — 바이트 동일 두 쌍:

```bash
diff .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac013-acp.diff
# 기대: exit 0
diff .agents/skills/moai-sync/SKILL.md $T/.agents/skills/moai-sync/SKILL.md > $E/ac013-published.diff
# 기대: exit 0
```

판정 (b) — 의도된 차이 여덟 쌍(파일마다 아래 다섯 줄을 경로만 바꿔 실행한다. 기준 트리 반출 이름은 `base-<이름>.md`·`base-<이름>-template.md`, plan.md §C 2단계):

```bash
diff $E/base-delivery.md $E/base-delivery-template.md > $E/ac013-delivery-base.diff
diff .claude/skills/moai/workflows/sync/delivery.md $T/.claude/skills/moai/workflows/sync/delivery.md > $E/ac013-delivery-post.diff
/usr/bin/grep -v -E '^[0-9]+(,[0-9]+)?[acd][0-9]+(,[0-9]+)?$' $E/ac013-delivery-base.diff > $E/ac013-delivery-base.body
/usr/bin/grep -v -E '^[0-9]+(,[0-9]+)?[acd][0-9]+(,[0-9]+)?$' $E/ac013-delivery-post.diff > $E/ac013-delivery-post.body
diff $E/ac013-delivery-base.body $E/ac013-delivery-post.body > $E/ac013-delivery-body.diff
# 기대: exit 0
# 같은 형태: manager-git.md(base-manager-git 대 base-manager-git-template; 경로에 manager-git.md 가 들어가므로 한 호출에 한 명령),
#            quality-gates-context.md(base-qgc), doc-execution.md(base-doc-execution), moai/SKILL.md(base-skill),
#            references/reference.md(base-reference), workflows/sync.md(base-sync), 명령 원본(base-command-sync 대 base-command-sync-template;
#            로컬 .claude/commands/moai/sync.md 대 $T/.claude/commands/moai/sync.md.tmpl)
test -s $E/ac013-delivery-base.body
# 기대: exit 0 — 여덟 쌍 모두 기준 본문이 비어 있지 않다(바이트 동일 쌍을 이 형태로 비교하면 빈 본문끼리 비교되어 공허하게 통과한다)
```

판정 (c) — 명령 원본 `argument-hint` 줄:

```bash
/usr/bin/grep -E '^argument-hint:' .claude/commands/moai/sync.md > $E/ac013-hint-local.txt
/usr/bin/grep -E '^argument-hint:' $T/.claude/commands/moai/sync.md.tmpl > $E/ac013-hint-template.txt
test -s $E/ac013-hint-local.txt
test -s $E/ac013-hint-template.txt
# 기대: 둘 다 exit 0
diff $E/ac013-hint-local.txt $E/ac013-hint-template.txt > $E/ac013-hint.diff
# 기대: exit 0
```

판정 (d) — 테스트:

```bash
go test ./internal/template/ -run '^(TestSanitizedPairParity|TestTemplateNoInternalContentLeak)$' -v -count=1 > $E/ac013-gotest.txt 2>&1
# 기대: exit 0 (0.2.5 에서 $BASE 트리에 실행: exit 0, 최상위 PASS 2 — reanchor/ac013-gotest-at-base.txt)
/usr/bin/grep -c -E '^--- PASS: (TestSanitizedPairParity|TestTemplateNoInternalContentLeak) ' $E/ac013-gotest.txt > $E/ac013-pass-count.txt
# 기대: 정확히 2
/usr/bin/grep -c -F 'agent-common-protocol.md' $E/ac013-gotest.txt > $E/ac013-acp-subtest.txt
# 기대: 1 이상 — TestSanitizedPairParity 가 범위 파일 하위 테스트를 실제로 돌렸다는 기록(빈 선택 방지)
```

판정 (e) — 미러 테스트 비회귀(M4 7단계, M6 4단계에서 실행. 기준선 `.moai/reports/t622/reanchor/mirror-baseline-sets.txt` · `mirror-baseline-fail.txt` · `mirror-baseline-pass.txt`, 원본 출력 `.moai/reports/t622/absorb2-mirror-baseline.txt`):

```bash
go test ./internal/template/ -count=1 -run 'TestSanitizedPairParity|TestRuleTemplateMirrorDrift' -v > $E/mirror-post.txt 2>&1
# exit 는 기록만 한다 — 기준선이 exit 1(범위 밖 spec-workflow.md 하위 테스트)이라 exit 로 판정하지 않는다
/usr/bin/grep -E '^[[:space:]]*--- (PASS|FAIL): ' $E/mirror-post.txt > $E/mirror-post-lines.txt
sed -E 's/^[[:space:]]*--- (PASS|FAIL): ([^ ]+).*/\1 \2/' $E/mirror-post-lines.txt > $E/mirror-post-sets.txt
/usr/bin/grep '^FAIL ' $E/mirror-post-sets.txt > $E/mirror-post-fail.txt
/usr/bin/grep '^PASS ' $E/mirror-post-sets.txt > $E/mirror-post-pass.txt
/usr/bin/grep -c '' $E/mirror-post-pass.txt > $E/mirror-post-pass-count.txt
# 기대: 17 이상 — 컴파일 실패나 빈 선택이면 PASS 줄이 사라진다(아래 잃은 PASS 판정도 같이 빨강)
/usr/bin/grep -v -x -F -f .moai/reports/t622/reanchor/mirror-baseline-fail.txt $E/mirror-post-fail.txt > $E/mirror-new-fail.txt
test -e $E/mirror-new-fail.txt
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
test -s $E/mirror-new-fail.txt
# 기대: exit 1 — 기준선 FAIL 집합 밖의 FAIL 이 없다
/usr/bin/grep -v -x -F -f $E/mirror-post-pass.txt .moai/reports/t622/reanchor/mirror-baseline-pass.txt > $E/mirror-lost-pass.txt
test -e $E/mirror-lost-pass.txt
# 기대: exit 0
test -s $E/mirror-lost-pass.txt
# 기대: exit 1 — 기준선 PASS 가 하나도 빠지거나 FAIL 로 바뀌지 않았다
```

기준선(0.2.5, `$BASE` 트리에 재실행 — `reanchor/mirror-baseline-rerun.txt`, 원본 `absorb2-mirror-baseline.txt` 와 PASS/FAIL 집합 `diff` exit 0): exit 1, 집합 19줄 — PASS 17(최상위 `TestSanitizedPairParity` 1 + 하위 16: `TestRuleTemplateMirrorDrift/{default,frontend,hooks-system,model-policy,session-handoff-examples,session-handoff,worktree-integration}.md`, `TestSanitizedPairParity/{agent-common-protocol,askuser-protocol,main-checkout-branch-guard,manager-develop-prompt-template,plan-auditor,runtime-recovery-doctrine,verification-batch-pattern,verification-claim-integrity,zone-registry}.md`), FAIL 2(`TestRuleTemplateMirrorDrift` 최상위와 `TestRuleTemplateMirrorDrift/spec-workflow.md`). `spec-workflow.md` 는 이 카드 범위 밖이고 다른 카드가 고친다 — 그 카드가 먼저 착지해 흡수되면 FAIL 이 사라지는 것은 (e)를 어기지 않는다(빠지는 것은 FAIL 이고, PASS 는 늘어도 된다). **흡수 뒤 귀속 규칙 (0.2.6).** 이 기준선은 `$BASE` 트리에서 잰 고정 스냅숏이다. 판정 전 신선도 점검의 `snapshot-stale-mirror.txt` 가 비어 있지 않으면(흡수가 미러 대상 파일을 바꿈) 새 FAIL 이나 잃은 PASS 가 흡수 탓일 수 있다. 그때 해당 하위 테스트가 읽는 두 사본 경로가 `git diff --name-only develop...HEAD` 에 없으면 흡수 탓으로 기록하고 리드에 올린다 — 이 카드의 FAIL 로 세지 않되, 기록 없이 넘기지도 않는다. 두 사본 경로가 그 목록에 있으면 이 카드의 FAIL 이다.

뮤턴트:

```
(b) 한쪽 사본만 편집 — reanchor/mut013/: 로컬 manager-git.md 116행 예시만 --<merge_method> 로 바꾸고 템플릿은 그대로 둔 픽스처
    → 본문 비교 exit 1(FAIL 검출, 새 본문 줄 "< gh pr merge <PR> --<merge_method> --delete-branch" / "> … --squash …")
    두 사본을 같게 바꾼 픽스처 → 본문 비교 exit 0(PASS). 관측 reanchor/mut013/result.txt
(c) 로컬 명령 원본의 argument-hint 만 --auto-merge 로 바꾸고 템플릿은 그대로 둔 상태 → ac013-hint.diff exit 1,
    차이 본문에 argument-hint 줄 쌍이 더해져 본문 비교 exit 1 → FAIL
(e) 집합 비교 — reanchor/mutmirror/result.txt:
    기준선 그대로                                   → new-fail test -s exit 1, lost-pass test -s exit 1 → PASS
    PASS TestSanitizedPairParity/agent-common-protocol.md 를 FAIL 로 바꾼 집합 → new-fail 1줄, lost-pass 1줄 → FAIL
    PASS TestSanitizedPairParity/zone-registry.md 를 뺀 집합                   → new-fail 0줄, lost-pass 1줄 → FAIL
```

### AC-GDP-014 — 에이전트 생성물 재생성

```
GIVEN 템플릿 manager-git.md 편집(템플릿 114·156행과 PR Auto-Merge 절)이 끝났고 .toml 은 아직 재생성하지 않은 상태
WHEN agents-emit-check → agents-emit → agents-emit-check 순서로 실행하면
THEN 첫 점검은 exit 1, 재생성은 exit 0, 두 번째 점검은 exit 0 이고
 AND 이 카드의 범위(develop...HEAD — 읽는 시점의 merge-base 부터) 에서 .codex/agents/moai/ 아래 바뀐 파일은 manager-git.toml 뿐이고
 AND .toml 의 병합 예시가 --<merge_method> 이고
 AND .toml 의 "## Synchronization" 절과 "## PR Auto-Merge" 절이 템플릿 manager-git.md 의 같은 절과 diff exit 0 이다
```

```bash
make agents-emit-check > $E/ac014-red.txt 2>&1
# 기대: exit 1 (재생성 전 RED — 검출기 양성 대조)
make agents-emit > $E/ac014-emit.txt 2>&1
# 기대: exit 0
make agents-emit-check > $E/ac014-green.txt 2>&1
# 기대: exit 0
git diff --name-only develop...HEAD -- $T/.codex/agents/moai/ > $E/ac014-changed.txt
# 기대: 한 줄, internal/template/templates/.codex/agents/moai/manager-git.toml — 범위 대조(관례의 card-range-names.txt 1줄 이상)가 먼저 성립해야 한다.
# 0.2.5 까지의 `git diff --name-only $BASE -- …` 는 흡수 뒤 develop 쪽 생성물 변경까지 담는다(0.2.6, 재고정 감사 D1).
# 0.2.6 측정: 현재 트리·흉내 낸 흡수 커밋 모두 빈 파일(아직 편집 전), 같은 형태의 양성 대조
# `git diff --name-only ee99507fbe3b4a22c6a0a74815723d222dfdc04d...97ef8e3023e9e7a29e7478289b69d28796dddbd7 -- $T/.codex/agents/moai/`
# (카드 dr0911 이 병합될 때의 develop 쪽 부모와 카드 tip) → 2줄 — reanchor-fix/ac014-*.txt
/usr/bin/grep -c -F 'gh pr merge <PR> --<merge_method> --delete-branch' $T/.codex/agents/moai/manager-git.toml > $E/ac014-toml-example.txt
# 기대: 1 (기준 트리 .toml 108행은 --squash, 이 개수 0)
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' $T/.claude/agents/moai/manager-git.md > $E/ac014-md-sync.md
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' $T/.codex/agents/moai/manager-git.toml > $E/ac014-toml-sync.md
diff $E/ac014-md-sync.md $E/ac014-toml-sync.md > $E/ac014-sync.diff
# 기대: exit 0 — 156행 새 순서 문장이 .toml 에 같은 문장으로 들어감
awk '/^## PR Auto-Merge/{s=1; print; next} s && /^## /{exit} s' $T/.claude/agents/moai/manager-git.md > $E/ac014-md-pram.md
awk '/^## PR Auto-Merge/{s=1; print; next} s && /^## /{exit} s' $T/.codex/agents/moai/manager-git.toml > $E/ac014-toml-pram.md
test -s $E/ac014-toml-pram.md
# 기대: exit 0
diff $E/ac014-md-pram.md $E/ac014-toml-pram.md > $E/ac014-pram.diff
# 기대: exit 0 — 모드별 승인 규칙이 .toml 에 같은 문장으로 들어감
```

plan 작성 시점 측정: 기준 트리에서 두 절의 diff 는 exit 0(11줄, 9줄). 뮤턴트: 기준 트리 `.toml` 절에서 "AND all approvals obtained" 를 "once checks pass" 로 바꾼 픽스처 → diff exit 1.

### AC-GDP-015 — 템플릿 중립성

```
GIVEN 이 카드의 템플릿 변경분(develop...HEAD — 읽는 시점의 merge-base 부터. 범위 파일 아홉 개의 템플릿 사본, 명령 원본 sync.md.tmpl, 게시본 포함)
WHEN 추가 줄(+++ 머리 줄 제외)을 검사하면
THEN SPEC ID, REQ 토큰, 날짜, CLAUDE.local 참조가 추가 줄에 없고
 AND 7~40자 16진 낱말 가운데 a-f 문자를 담은 것이 없으며
 AND 숫자로만 된 7~40자 낱말은 목록으로 뽑혀 읽기 단계에서 커밋 SHA가 아님이 기록된다
```

`git diff develop...HEAD -- $T/` 는 템플릿 루트 전체를 담으므로 X1~X3의 템플릿 사본과 게시본 변화도 이 판정에 들어간다. 흡수된 develop 의 템플릿 변경은 merge-base 에 이미 들어 있어 이 변경분에 섞이지 않는다(0.2.6 측정: 흉내 낸 흡수 커밋에서 이 형태는 빈 결과인데, 리터럴 `git diff $BASE <흡수 커밋> -- $T/` 는 develop 의 템플릿 6개 파일·추가 줄 146줄을 담았다 — `reanchor-fix/ac015-*.diff`). 프로그래밍 언어 편향은 추가 줄을 읽어 기록하고 CI의 `template-neutrality-check` 결과를 함께 적는다.

대조:

```bash
/usr/bin/grep -c -E 'SPEC-([A-Z][A-Z0-9]*-)+[0-9]{3}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-specid.txt
/usr/bin/grep -c -E 'REQ-([A-Z][A-Z0-9]*-)+[0-9]{3}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-req.txt
/usr/bin/grep -c -E '20[0-9]{2}-[0-9]{2}-[0-9]{2}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-date.txt
# 기대: 세 값 모두 1 이상
perl -ne 'while (/(?<![0-9A-Za-z])([0-9a-f]{7,40})(?![0-9A-Za-z])/g) { print "$.:$1\n" }' -- .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-sha-tokens.txt
/usr/bin/grep -c -E ':980ccdc56$' $E/ac015-control-sha-tokens.txt > $E/ac015-control-sha-980.txt
# 기대: 1 이상. 전체 낱말 수는 progress.md §E.1 에 기록
perl -e 'print "+++ b/file.md\n+sha 538684c47 here\n+one line 7374b183e 2213871af 980ccdc56 b412f8a33\n+all digits 647460835 and 1000000 here\n context 02aca7afe not added\n"' > $E/ac015-fixture.diff
perl -ne 'next unless /^\+(?!\+\+)/; while (/(?<![0-9A-Za-z])([0-9a-f]{7,40})(?![0-9A-Za-z])/g) { print "$.:$1\n" }' -- $E/ac015-fixture.diff > $E/ac015-fixture-tokens.txt
/usr/bin/grep -c -v -E ':[0-9]+$' $E/ac015-fixture-tokens.txt > $E/ac015-fixture-letter.txt
# 기대: 5
/usr/bin/grep -c -E ':[0-9]+$' $E/ac015-fixture-tokens.txt > $E/ac015-fixture-digits.txt
# 기대: 2
```

판정:

```bash
git diff develop...HEAD -- $T/ > $E/ac015-template.diff
test -s $E/ac015-template.diff
# 기대: exit 0 — 변경분이 비면 아래 부재 판정이 모두 공허하게 통과한다. 편집 전에는 변경분이 없어 exit 1(0.2.6 측정: 현재 트리·흉내 낸 흡수 커밋 모두 test -e 0·test -s 1 — reanchor-fix/ac015-card-template-*.diff).
#       develop...HEAD 는 마지막으로 흡수한 develop 커밋과의 merge-base 부터 재므로, 몇 번을 다시 흡수해도 이 변경분은 이 카드의 템플릿 편집만 담는다(병합 전 전용 — 관례의 한계)
/usr/bin/grep -c -E '^[+][^+]' $E/ac015-template.diff > $E/ac015-added-count.txt
# 기대: 1 이상 — 검사할 추가 줄이 실제로 있다
/usr/bin/grep -n -E '^[+][^+].*SPEC-([A-Z][A-Z0-9]*-)+[0-9]{3}' $E/ac015-template.diff > $E/ac015-specid.txt
# 기대: exit 1
/usr/bin/grep -n -E '^[+][^+].*REQ-([A-Z][A-Z0-9]*-)+[0-9]{3}' $E/ac015-template.diff > $E/ac015-req.txt
# 기대: exit 1
/usr/bin/grep -n -E '^[+][^+].*20[0-9]{2}-[0-9]{2}-[0-9]{2}' $E/ac015-template.diff > $E/ac015-date.txt
# 기대: exit 1
/usr/bin/grep -n -E '^[+][^+].*CLAUDE[.]local' $E/ac015-template.diff > $E/ac015-local-ref.txt
# 기대: exit 1
perl -ne 'next unless /^\+(?!\+\+)/; while (/(?<![0-9A-Za-z])([0-9a-f]{7,40})(?![0-9A-Za-z])/g) { print "$.:$1\n" }' -- $E/ac015-template.diff > $E/ac015-hex-tokens.txt
/usr/bin/grep -v -E ':[0-9]+$' $E/ac015-hex-tokens.txt > $E/ac015-sha-letter.txt
test -e $E/ac015-sha-letter.txt
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
test -s $E/ac015-sha-letter.txt
# 기대: exit 1. 비어 있지 않으면 낱말마다 읽어 커밋 SHA인지 판정
/usr/bin/grep -E ':[0-9]+$' $E/ac015-hex-tokens.txt > $E/ac015-sha-digits.txt
# 읽기 단계: 숫자로만 된 낱말마다 커밋 SHA가 아닌지 $E/ac015-reading.md 에 기록
```

검출 한계: 대문자 16진, 7자 미만 약식 SHA, 영숫자에 바로 붙은 16진 낱말은 잡지 않는다.

### AC-GDP-016 — 이 카드의 커밋이 `agent-common-protocol.md` 를 바꾸지 않음 (SHOULD, 절차 점검 — 0.2.5 재작성)

```
GIVEN 이 카드의 커밋들(HEAD --not develop, 병합 커밋 제외 — plan·run·sync 커밋 모두)
WHEN agent-common-protocol.md 의 로컬·템플릿 두 사본을 바꾼 커밋을 찾으면
THEN 결과가 비어 있다 — 이 카드는 두 사본 어느 쪽도 고치지 않는다(REQ-GDP-002 는 develop 카드 t635·dr0911 로 충족, 리드 판단 (a))
 AND 같은 명령 형태가 그 파일을 바꾼 커밋이 있는 구간에서는 커밋을 찾는다(양성 대조)
```

0.2.4 의 판정("`agent-common-protocol.md` 커밋이 마지막 지침 편집 커밋")은 이 카드가 그 파일을 고친다는 전제였다. 0.2.5 에서 그 편집이 없어져 판정을 "편집 커밋 없음"으로 바꿨다. 0.2.6 에서 범위를 리터럴 `$BASE..HEAD` 에서 `HEAD --not develop` 으로 바꿨다(재고정 감사 D1) — 관례의 "카드 변경 범위 판정" 을 따른다. GIVEN 의 범위는 plan 커밋까지 포함한 이 카드의 모든 비병합 커밋이다(D6).

```bash
git log --no-merges --format=%H 255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089 --not b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0 -- .claude/rules/moai/core/agent-common-protocol.md internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md > $E/ac016-control.txt
/usr/bin/grep -c '' $E/ac016-control.txt > $E/ac016-control-count.txt
# 대조 기대: 1 이상 — 판정과 같은 `<tip> --not <ref>` 형태와 경로 지정이 이 파일을 바꾼 커밋을 실제로 찾는다(고정 과거 구간이라 리터럴 SHA 를 쓴다).
# 0.2.6 측정: 2줄 — 97ef8e3023e9e7a29e7478289b69d28796dddbd7(dr0911 템플릿 미러), 6896eef3766a5265ac32b21154e95907e1173b54(t635)
#   — reanchor-fix/ac016-control-notform.txt. 0.2.5 의 `b412f8a33..$BASE` 형태 측정(reanchor/ac016-control-b412-to-base.txt)과 같은 집합
git log --no-merges --format=%H HEAD --not develop -- .claude/rules/moai/core/agent-common-protocol.md internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md > $E/ac016-acp-commits.txt
test -e $E/ac016-acp-commits.txt
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
test -s $E/ac016-acp-commits.txt
# 기대: exit 1 — 빈 파일. 관례의 범위 대조(card-range-names.txt 1줄 이상)가 먼저 성립해야 한다 — 범위가 비면 빈 파일은 "측정 불가" 다.
# 0.2.6 측정: 현재 트리(HEAD f2fa64e08)와 흉내 낸 흡수 커밋 7fae4764e 모두 test -e 0 · test -s 1(reanchor-fix/ac016-judge-*.txt).
#   같은 두 지점에서 경로 없는 `HEAD --not develop` 카드 커밋 목록은 23줄로 같았다(diff exit 0).
#   리터럴 `$BASE..<흡수 커밋>` 은 3줄 → 12줄로 늘었고 늘어난 9줄은 모두 develop 커밋이었다(literal-base-range-*.txt).
```

전제와 한계:

- **재흡수.** 흡수된 develop 커밋은 develop 에서 닿으므로 `--not develop` 이 모두 뺀다. 흡수를 몇 번 해도 판정이 바뀌지 않는다. 0.2.5 의 "커밋 메시지의 카드 id 로 가림" 과 "BASE 재고정 사유" 는 폐기했다 — 메시지 문자열에 기대면 카드 id 를 언급하는 develop 커밋(리드 배치 커밋 등)을 잘못 분류한다.
- **병합 전 전용.** 카드가 develop 에 병합되면 카드 커밋이 develop 에서 닿게 되어 목록이 빈다. 병합 뒤 근거는 병합 트리 동일성이다(관례의 한계).
- **병합 커밋 안의 편집.** `--no-merges` 는 병합 커밋의 충돌 해결로 들어간 편집을 보지 않는다. 이 카드가 흡수 병합에서 이 파일의 충돌을 손으로 풀었다면 그 병합을 따로 기록한다. 흡수 병합이 이 파일을 바꿨는지는 스냅숏 신선도 점검이 따로 알린다.

뮤턴트(판정 논리 — 이 카드 이력에 커밋을 만드는 픽스처는 실행하지 않았다. 검출 능력은 위 양성 대조가 실측으로 보인다):

```
이 카드가 로컬 사본만 고친 커밋 C1   → ac016-acp-commits.txt 에 C1 → test -s exit 0 → FAIL
이 카드가 템플릿 사본만 고친 커밋 C1 → 같음 → FAIL (두 경로를 한 명령에 담아 어느 쪽 편집이든 잡는다)
이 카드 커밋이 두 사본을 건드리지 않음 → 빈 파일 → PASS
흡수된 develop 커밋이 두 사본을 바꿈 → --not develop 이 빼서 빈 파일 → PASS (카드 위반이 아니다. 신선도 점검이 리드 보고로 돌린다)
판정 명령이 돌지 않음               → test -e exit 1 → 판정 불가(PASS 아님)
범위 대조가 0줄                     → 측정 불가(PASS 아님)
```

### AC-GDP-017 ~ AC-GDP-024 — 자리표시

카드 t658로 옮겼다. 판정하지 않으며 원문은 커밋 `87988e946` 의 acceptance.md, 추적 표는 spec.md §G.1이다.

### AC-GDP-025 — Frozen 줄과 등록 Frozen clause 불변

```
GIVEN 범위 파일 아홉 개의 로컬·템플릿 사본
WHEN 이 카드의 변경분(develop...HEAD — 읽는 시점의 merge-base 부터)과 등록 Frozen clause 를 검사하면
THEN 변경분에 [ZONE:Frozen] 을 담은 추가·삭제 줄이 없고
 AND agent-common-protocol.md 두 사본 모두 등록 Frozen clause 네 문장이 기준 트리와 같은 개수로 남아 있고
 AND zone-registry.md 에서 나머지 여덟 파일을 가리키는 항목이 여전히 0개이며, 같은 검출식 형태로 센 agent-common-protocol.md 항목이 13개다(양성 대조)
```

등록 Frozen clause (`zone-registry.md`, 모두 `file: .claude/rules/moai/core/agent-common-protocol.md`, `#user-interaction-boundary`):

| ID | clause | 기준 트리 위치 |
|---|---|---|
| `CONST-V3R2-006` | `` `AskUserQuestion` is the **only** user-facing question channel `` | 13행 |
| `CONST-V3R2-036` | `Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.` | 17행 |
| `CONST-V3R2-037` | `` Preload `AskUserQuestion` via `ToolSearch(query: `` | 52행 |
| `CONST-V3R2-038` | `AskUserQuestion is reserved exclusively for the MoAI orchestrator` | 17행 |

대조(plan 작성 시점 측정, 템플릿·로컬 사본 — 0.2.5 에서 `$BASE` 에 다시 재 같은 값: `reanchor/zone-lines-others.txt`·`registry-others.txt`·`lines-acp.txt`·`clause-counts.txt`): `[ZONE:Frozen]` 줄 — `agent-common-protocol.md` 17행 1줄, 나머지 여덟 파일 0줄(명령 원본 `sync.md`·`sync.md.tmpl` 포함). 네 clause 개수 — 사본마다 1·1·1·1. 레지스트리에서 나머지 여덟 파일을 가리키는 `file:` 항목 0개(명령 원본 `file: .claude/commands/moai/sync.md` 0개 포함), `agent-common-protocol.md` 를 가리키는 항목 13개(로컬·템플릿). 뮤턴트: `agent-common-protocol.md` 템플릿 사본의 "MUST NOT prompt" 를 소문자로 바꾼 픽스처 → `CONST-V3R2-036` 개수 0, 기준 사본과의 `diff` 에 `[ZONE:Frozen]` 을 담은 줄 2개 → FAIL. Frozen 줄 판정 뮤턴트: `-[ZONE:Frozen] a` · `+[ZONE:Frozen] b` · ` [ZONE:Frozen] c`(문맥) 세 줄 diff 픽스처 → 2줄 적중, 문맥 줄 미적중. 레지스트리 "나머지 여덟 파일" 검출식 뮤턴트(0.2.2 측정): `file: .claude/commands/moai/sync.md` · `file: .claude/skills/moai/workflows/sync.md` · `file: .claude/rules/moai/core/agent-common-protocol.md` 세 줄 픽스처 → 2 — 검출식이 명령 원본과 새 스킬 경로를 실제로 잡고 `agent-common-protocol.md` 는 잡지 않는다. 이 검출식은 워크트리 가드 때문에 `manager-[g]it` 으로 쓴다.

판정:

```bash
git diff develop...HEAD -- .claude/agents/moai/manager-git.md .claude/rules/moai/core/agent-common-protocol.md .claude/skills/moai/workflows/sync/delivery.md .claude/skills/moai/workflows/sync/doc-execution.md .claude/skills/moai/SKILL.md .claude/skills/moai/references/reference.md .claude/skills/moai/workflows/sync/quality-gates-context.md .claude/skills/moai/workflows/sync.md .claude/commands/moai/sync.md $T/.claude/agents/moai/manager-git.md $T/.claude/rules/moai/core/agent-common-protocol.md $T/.claude/skills/moai/workflows/sync/delivery.md $T/.claude/skills/moai/workflows/sync/doc-execution.md $T/.claude/skills/moai/SKILL.md $T/.claude/skills/moai/references/reference.md $T/.claude/skills/moai/workflows/sync/quality-gates-context.md $T/.claude/skills/moai/workflows/sync.md $T/.claude/commands/moai/sync.md.tmpl > $E/ac025-scope.diff
test -s $E/ac025-scope.diff
# 기대: exit 0 — 범위 변경분이 비면 아래 Frozen 줄 판정이 공허하게 통과한다(편집 전에는 변경분이 없어 exit 1).
# 0.2.6: 0.2.5 까지의 `git diff $BASE -- …` 는 흡수 뒤 develop 의 범위 파일 변경까지 이 카드 것으로 읽는다(재고정 감사 D1 과 같은 형태).
#   0.2.6 측정: 현재 트리·흉내 낸 흡수 커밋 모두 test -e 0 · test -s 1(reanchor-fix/ac025-scope-*.diff)
/usr/bin/grep -F '[ZONE:Frozen]' $E/ac025-scope.diff > $E/ac025-frozen-any.txt
/usr/bin/grep -n -E '^[-+].*\[ZONE:Frozen\]' $E/ac025-frozen-any.txt > $E/ac025-frozen-lines.txt
# 기대: exit 1 (문맥 줄 ' …[ZONE:Frozen]' 은 허용, 추가·삭제 줄은 불허).
# 첫 grep 에 -n 을 붙이면 줄번호 머리 때문에 둘째 grep 의 ^[-+] 가 아무 줄도 잡지 못해 공허하게 통과한다 — 붙이지 않는다.
/usr/bin/grep -c -F '`AskUserQuestion` is the **only** user-facing question channel' .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac025-const006.txt
/usr/bin/grep -c -F 'Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.' .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac025-const036.txt
/usr/bin/grep -c -F 'Preload `AskUserQuestion` via `ToolSearch(query:' .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac025-const037.txt
/usr/bin/grep -c -F 'AskUserQuestion is reserved exclusively for the MoAI orchestrator' .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac025-const038.txt
# 기대: 네 파일 모두 사본마다 1
/usr/bin/grep -c -E 'file: \.claude/rules/moai/core/agent-common-protocol\.md' .claude/rules/moai/core/zone-registry.md $T/.claude/rules/moai/core/zone-registry.md > $E/ac025-registry-control.txt
# 기대: 사본마다 13 — 검출식 형태가 레지스트리 항목을 실제로 잡는다는 양성 대조
/usr/bin/grep -c -E 'file: \.claude/(agents/moai/manager-[g]it\.md|skills/moai/workflows/sync/(delivery|doc-execution|quality-gates-context)\.md|skills/moai/SKILL\.md|skills/moai/references/reference\.md|skills/moai/workflows/sync\.md|commands/moai/sync\.md)' .claude/rules/moai/core/zone-registry.md $T/.claude/rules/moai/core/zone-registry.md > $E/ac025-registry-others.txt
# 기대: 사본마다 0
```

### AC-GDP-026 — 플래그 표면: `--auto-merge` 노출, `--merge` 는 폐기된 별칭으로만

```
GIVEN 플래그 표면 아홉 조각의 로컬·템플릿 사본
      (1) moai/SKILL.md 의 "Modes: auto, force, status, project. Flags:" 줄
      (2) references/reference.md 의 "- Modes (positional): auto (default), force, status, project" 로 시작하는 목록(빈 줄까지)
      (3) quality-gates-context.md 의 "- $ARGUMENTS: Mode and optional path" 로 시작하는 목록(빈 줄까지)
      (4) quality-gates-context.md 의 "## Supported Flags" 절
      (5) workflows/sync.md 의 "**Flags**:" 줄
      (6) delivery.md 의 "#### Step 3.4" 절
      (7) workflows/sync.md 의 "/moai sync [mode]" 로 시작하는 사용법 줄 (X1)
      (8) 명령 원본의 "argument-hint:" 줄 — 로컬 .claude/commands/moai/sync.md, 템플릿 .claude/commands/moai/sync.md.tmpl (X2)
      (9) delivery.md 의 "#### Context-Aware Next Steps" 절 (X3, 404행 선택지를 담음)
WHEN 편집 뒤 조각을 추출해 검사하면
THEN (i) 조각 (1)·(2)·(5)·(6)·(7)·(8)·(9)와 조각 (3)·(4)를 합친 quality-gates-context 조각이 각각 --auto-merge 를 1회 이상 담고
 AND (ii) --merge 낱말(앞뒤가 영문자·하이픈이 아닌 --merge)을 담은 줄은 모두, 그 낱말 뒤에
     "deprecated alias of --auto-merge"(또는 "deprecated alias for", --auto-merge 앞에 백틱·따옴표 한 글자 허용) 구절을 담고
     "not/no longer/never deprecated" 나 "un-deprecated" 를 담지 않는다 (위반 줄 0개)
     — --merge 가 --auto-merge 의 폐기된 별칭이라는 방향만 통과한다. --auto-merge 를 폐기됐다고 적거나 별칭 관계를 뒤집은 줄은 위반이다
 AND (iii) --merge 낱말이 조각 (1) SKILL Flags 줄, (5) workflows/sync.md 의 **Flags**: 줄, (4) Supported Flags 절, (6) Step 3.4 절에
     각각 1줄 이상 남아 있다 — 폐기된 별칭은 지우지 않고 문서화한다(운영자 결정, spec.md §C.1). 조각 (2)·(3)·(7)·(8)·(9)에서는 빼도 된다
 AND 읽기 기록 $E/ac026-reading.md 가 존재하며, ac026-merge-lines.txt 의 모든 줄(로컬·템플릿)에 대해
     "이 줄은 --merge 가 --auto-merge 의 폐기된 별칭이라고 말하는가 — 뒤집힌 관계, 별개 플래그, 폐기되지 않은 플래그, 자체 동작을 가진 플래그가 아닌가"
     에 예로 답한다 (자동 검출 (ii)는 보조이고 읽기 기록이 PASS의 전제다, §D.3)
```

조각 추출(대조·판정 공통, `<SKILL>` 등은 로컬·템플릿 경로):

```bash
/usr/bin/grep -E '^Modes: auto, force, status, project\. Flags:' <SKILL> > $E/ac026-skill.md
awk '/^- Modes \(positional\): auto \(default\), force, status, project/{s=1; print; next} s && /^$/{exit} s' <reference> > $E/ac026-ref.md
awk '/^- \$ARGUMENTS: Mode and optional path/{s=1; print; next} s && /^$/{exit} s' <quality-gates-context> > $E/ac026-qgc-args.md
awk '/^## Supported Flags/{s=1; print; next} s && /^## /{exit} s' <quality-gates-context> > $E/ac026-qgc-flags.md
/usr/bin/grep -E '^\*\*Flags\*\*: ' <sync.md> > $E/ac026-sync.md
awk '/^#### Step 3\.4/{s=1; print; next} s && /^(####|###) /{exit} s' <delivery> > $E/ac026-dl.md
/usr/bin/grep -E '^/moai sync \[mode\]' <sync.md> > $E/ac026-sync-usage.md
/usr/bin/grep -E '^argument-hint:' <command source> > $E/ac026-hint.md
awk '/^#### Context-Aware Next Steps/{s=1; print; next} s && /^(####|###|##) /{exit} s' <delivery> > $E/ac026-dl-next.md
```

조각의 추출 표지(`Modes: auto, force, status, project. Flags:`, `- Modes (positional): …`, `- $ARGUMENTS: Mode and optional path`, `## Supported Flags`, `**Flags**:`, `#### Step 3.4`, `/moai sync [mode]`, `argument-hint:`, `#### Context-Aware Next Steps`)는 편집 뒤에도 남아야 한다.

대조(기준 트리, 템플릿 사본; 조각 (7)·(8)·(9)는 로컬 사본도 같은 값):

```
조각 줄 수: skill 1, ref 3, qgc-args 4, qgc-flags 6, sync 1, dl 52, sync-usage 1, hint 1, dl-next 24
(i) --auto-merge 개수: 아홉 조각 모두 0 → 빨강
(ii) 위반 줄: skill 1(140행), ref 2(161행), qgc-args 4(30행), qgc-flags 4(101행), sync 1(템플릿 104 / 로컬 114), dl 9(템플릿 338 / 로컬 363)·20(템플릿 349 / 로컬 374),
     sync-usage 1(템플릿 85 / 로컬 95), hint 1(3행), dl-next 9(템플릿 404 / 로컬 429) → 10줄 → 빨강
     (0.2.5: $BASE 에서 아홉 조각을 로컬·템플릿 모두 다시 뽑아 조각마다 L·T diff exit 0, 줄 수와 위반 조각 줄이 위와 같음 — reanchor/frag/)
     (0.2.3 방향 검사로 다시 잰 값도 같은 10줄)
(iii) --merge 존재 개수: skill 1, sync 1, qgc-flags 1, dl 2 → 기준 트리에서도 성립(편집이 지우면 빨강)
```

판정(로컬·템플릿 각각):

```bash
test -s $E/ac026-skill.md
test -s $E/ac026-ref.md
test -s $E/ac026-qgc-args.md
test -s $E/ac026-qgc-flags.md
test -s $E/ac026-sync.md
test -s $E/ac026-dl.md
test -s $E/ac026-sync-usage.md
test -s $E/ac026-hint.md
test -s $E/ac026-dl-next.md
# 기대: 모두 exit 0 — 조각이 비면 판정 불가
/usr/bin/grep -c -e '--auto-merge' $E/ac026-skill.md $E/ac026-ref.md $E/ac026-sync.md $E/ac026-dl.md $E/ac026-sync-usage.md $E/ac026-hint.md $E/ac026-dl-next.md > $E/ac026-i.txt
/usr/bin/grep -c -e '--auto-merge' $E/ac026-qgc-args.md $E/ac026-qgc-flags.md > $E/ac026-i-qgc.txt
# 기대: ac026-i.txt 파일마다 1 이상, ac026-i-qgc.txt 합계 1 이상
awk '/(^|[^A-Za-z-])--merge([^A-Za-z-]|$)/ && (!/--merge([^A-Za-z-].*)?[Dd]eprecated alias (of|for) .?--auto-merge([^A-Za-z-]|$)/ || /(not|no longer|never) (a |an |the )?[Dd]eprecated/ || /[Uu]n-?deprecat/) {print FILENAME ":" FNR ": " $0}' $E/ac026-skill.md $E/ac026-ref.md $E/ac026-qgc-args.md $E/ac026-qgc-flags.md $E/ac026-sync.md $E/ac026-dl.md $E/ac026-sync-usage.md $E/ac026-hint.md $E/ac026-dl-next.md > $E/ac026-ii.txt
test -e $E/ac026-ii.txt
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
test -s $E/ac026-ii.txt
# 기대: exit 1
/usr/bin/grep -c -E '(^|[^A-Za-z-])--merge([^A-Za-z-]|$)' $E/ac026-skill.md $E/ac026-sync.md $E/ac026-qgc-flags.md $E/ac026-dl.md > $E/ac026-merge-presence.txt
# 기대: (iii) 네 파일 모두 1 이상 — exit code가 아니라 파일에 찍힌 개수로 판정한다. 하나라도 0이면 FAIL
awk '/(^|[^A-Za-z-])--merge([^A-Za-z-]|$)/ {print FILENAME ":" FNR ": " $0}' $E/ac026-skill.md $E/ac026-ref.md $E/ac026-qgc-args.md $E/ac026-qgc-flags.md $E/ac026-sync.md $E/ac026-dl.md $E/ac026-sync-usage.md $E/ac026-hint.md $E/ac026-dl-next.md > $E/ac026-merge-lines.txt
test -s $E/ac026-merge-lines.txt
# 기대: exit 0 — (iii)가 성립하면 비지 않는다. 로컬·템플릿 각각의 이 목록 줄이 모두 읽기 기록의 대상이다
test -s $E/ac026-reading.md
# 기대: exit 0 — 읽기 기록이 없거나 목록의 한 줄이라도 답이 비면 PASS 불가
```

(ii)의 방향 검사가 받는 모양: `--merge` 낱말 뒤 같은 줄에 "deprecated alias of `--auto-merge`" 또는 "deprecated alias for `--auto-merge`". "`--auto-merge` (deprecated alias: `--merge`)" 처럼 뜻은 맞지만 어순이 다른 문구는 위반으로 잡힌다(FAIL 쪽으로 기운다) — plan.md §B.14가 받는 문구를 적어 둔다.

뮤턴트(plan 작성 시점, 스크래치 픽스처):

```
결정 목록 조각 픽스처 5줄: "Modes: … Flags: --merge, --skip-mx"(그대로) · "Modes: … Flags: --auto-merge, --merge, --skip-mx"(별개·폐기 표시 없음)
  · "- `--merge` (deprecated)"(별칭 관계 없음) · "- `--merge`: deprecated alias of `--auto-merge` (logs a warning)"(올바름)
  · "moai worktree clean --merged-only"(다른 플래그)
  → (ii)가 1·2·3줄을 잡고 4·5줄은 잡지 않음
X1~X3 픽스처 6줄: "/moai sync [mode] [--pr] [--auto-merge] [--merge] [--skip-mx]" · "/moai sync [mode] [--pr] [--auto-merge] [--skip-mx]"
  · "argument-hint: \"[SPEC-XXX] [--merge] [--skip-mx]\"" · "argument-hint: \"[SPEC-XXX] [--auto-merge] [--skip-mx]\""
  · "- Auto-Merge PR (/moai sync --merge)" · "- Auto-Merge PR (/moai sync --auto-merge)"
  → (ii)가 1·3·5줄(옛 별칭을 폐기 표시 없이 남긴 줄)을 잡고, --auto-merge 를 담은 줄은 1·2·4·6줄
    → 3줄이나 5줄만 남은 조각은 (i) 0 과 (ii) 적중으로 FAIL, 2·4·6줄만 남은 조각은 PASS
```

뮤턴트(0.2.3, 스크래치 픽스처 — 레인이 같은 줄로 다시 돌린다):

```
m026.md 10줄 (감사 2회차 줄은 글자 그대로):
  1 "- `--auto-merge`: deprecated alias of `--merge`; use `--merge` to auto-merge the PR"            (뒤집힌 별칭, 감사)
  2 "- `--merge`: auto-merge the PR after sync (the older `--auto-merge` spelling is deprecated)"     (뒤집힌 별칭·폐기 해제, 감사)
  3 "- `--merge`: alias of `--auto-merge`"                                                           (폐기 표시 없음, 감사)
  4 "- `--merge`: deprecated alias of `--auto-merge` (logs a warning)"                               (올바름)
  5 "- `--merge` is no longer a deprecated alias of `--auto-merge`; it is the primary flag."         (부정)
  6 "Modes: auto, force, status, project. Flags: --auto-merge, --merge (deprecated alias of --auto-merge), --skip-mx" (올바름, 한 줄 목록)
  7 "moai clean --merged-only"                                                                       (다른 플래그)
  8 "Modes: … Flags: --merge, --skip-mx"   9 "Modes: … Flags: --auto-merge, --merge, --skip-mx"   10 "- `--merge` (deprecated)"
명령: 위 판정의 (ii) awk 를 m026.md x13.md 에 실행 (x13.md = 위 X1~X3 픽스처 6줄)
관측: m026.md:1 2 3 5 8 9 10, x13.md:1 3 5 적중 / m026.md 4·6·7 미적중 → 올바른 줄만 통과

제거 픽스처 rm026/ (감사 줄 "Modes: auto, force, status, project. Flags: --auto-merge, --skip-mx" 를 skill.md 로,
  "**Flags**: --pr, --auto-merge, --skip-mx" 를 sync.md 로, "## Supported Flags" + "- `--auto-merge`: merge the PR after sync" 를 qgc-flags.md 로,
  "#### Step 3.4" + 같은 목록 줄을 dl.md 로):
명령: /usr/bin/grep -c -E '(^|[^A-Za-z-])--merge([^A-Za-z-]|$)' skill.md sync.md qgc-flags.md dl.md
관측: skill.md:0 sync.md:0 qgc-flags.md:0 dl.md:0 (exit 1) → (iii) FAIL

양성 대조 pos026/ (skill.md = 6번 줄, sync.md = "**Flags**: --pr, --auto-merge, --merge (deprecated alias of --auto-merge), --skip-mx",
  qgc-flags.md = "## Supported Flags" + 4번 줄, dl.md = "#### Step 3.4" + 4번 줄):
관측: (iii) 1 1 1 1 · (ii) 출력 파일 test -s exit 1 · (i) --auto-merge 개수 1 1 1 1 → PASS
```

### AC-GDP-027 — `--no-merge` 는 폐기된 no-op으로만

```
GIVEN AC-GDP-026 의 아홉 조각(로컬·템플릿)
WHEN 편집 뒤 --no-merge 를 담은 줄을 검사하면
THEN (i) --no-merge 를 담은 줄은 모두 "no-op" 과 "deprecat" 를 함께 담고
 AND (ii) --no-merge 를 담은 줄 가운데 효과·조건 낱말(skip, not set, prevent, unless, disable, override, turn(s) off, suppress, bypass, cancel)을 담은 줄이 0개다
```

(ii)는 `--no-merge` 가 병합을 건너뛰게 하거나 트리거 조건(`--no-merge flag NOT set`)으로 쓰이는 서술을 잡는다. `--no-merge` 가 어느 조각에도 없으면 두 조건은 공허하게 참이 되므로, 판정 기록에 조각별 `--no-merge` 줄 수를 함께 적는다(REQ-GDP-025는 `--no-merge` 를 호환용 no-op으로 서술할 것을 요구하므로 `delivery.md` Step 3.4 조각에는 1줄 이상 있어야 한다).

대조(기준 트리, 템플릿 사본 — 로컬 사본 조각도 같음, 0.2.5 측정): `--no-merge` 줄 — dl 조각 8(템플릿 337행 / 로컬 362행)·19(템플릿 348행 / 로컬 373행), 나머지 여덟 조각 0(X1~X3 조각 포함). (i) 위반 8·19 → 빨강. (ii) 위반 8("NOT set")·19("Skip") → 빨강.

판정(로컬·템플릿 각각):

```bash
/usr/bin/grep -c -e '--no-merge' $E/ac026-dl.md > $E/ac027-dl-count.txt
# 기대: 1 이상
awk '/--no-merge/ && !(/no-op/ && /[Dd]eprecat/) {print FILENAME ":" FNR ": " $0}' $E/ac026-skill.md $E/ac026-ref.md $E/ac026-qgc-args.md $E/ac026-qgc-flags.md $E/ac026-sync.md $E/ac026-dl.md $E/ac026-sync-usage.md $E/ac026-hint.md $E/ac026-dl-next.md > $E/ac027-i.txt
test -e $E/ac027-i.txt
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
test -s $E/ac027-i.txt
# 기대: exit 1
awk '/--no-merge/ && (/[Ss]kip/ || /[Nn][Oo][Tt] set/ || /[Pp]revent/ || /unless/ || /[Dd]isabl/ || /[Oo]verrid/ || /[Tt]urns? off/ || /[Ss]uppress/ || /[Bb]ypass/ || /[Cc]ancel/) {print FILENAME ":" FNR ": " $0}' $E/ac026-skill.md $E/ac026-ref.md $E/ac026-qgc-args.md $E/ac026-qgc-flags.md $E/ac026-sync.md $E/ac026-dl.md $E/ac026-sync-usage.md $E/ac026-hint.md $E/ac026-dl-next.md > $E/ac027-ii.txt
test -e $E/ac027-ii.txt
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
test -s $E/ac027-ii.txt
# 기대: exit 1
```

(ii)의 효과 낱말은 "no-op" 이라 적고도 효과를 말하는 줄을 잡는다. 올바른 줄이 "does not disable anything" 처럼 부정문으로 효과 낱말을 쓰면 위반으로 잡힌다(FAIL 쪽으로 기운다) — 받는 문구는 plan.md §B.14.

뮤턴트(0.1.x 4줄에 감사 2회차 3줄을 더한 스크래치 픽스처 m027.md 7줄 — 워크트리 가드가 픽스처 쓰기 명령 속 `worktree` 낱말을 거부해 1·3줄의 해당 낱말만 줄였고 검출과 무관하다):
  1 "- `--no-merge`: Skip auto-merge even in context." · 2 "- `--no-merge`: Deprecated no-op; skips auto-merge." · 3 "- `is_wt_context == true` AND `--no-merge` flag not set"
  · 4 "- `--no-merge`: Deprecated no-op kept for compatibility (logs a warning); not merging is already the default."(올바름)
  · 5 "- `--no-merge`: Deprecated no-op; disables auto-merge for this run."(감사) · 6 "- `--no-merge`: Deprecated no-op that overrides `--auto-merge`."(감사)
  · 7 "- `--no-merge`: Deprecated no-op; turns off merging even when `--auto-merge` is given."(감사)
명령: 위 (i)·(ii) awk 를 m027.md 에 실행. 관측: (i) 1·3줄, (ii) 1·2·3·5·6·7줄 적중, 4줄은 둘 다 미적중 → 올바른 줄만 통과. 기준 트리 dl 조각: (i) 8·19, (ii) 8·19(0.2.2와 같음).

### AC-GDP-028 — team 모드: 전원 승인 조건

```
GIVEN manager-git.md 로컬·템플릿 사본의 "## PR Auto-Merge" 로 시작하는 절(다음 "## " 제목 앞까지)과 delivery.md 로컬·템플릿 사본의 Step 3.4 절
WHEN 편집 뒤 두 절을 검사하면
THEN (a) 두 절 각각에 "team mode"·"--auto-merge" 와 전원 승인 구절 — "all" 뒤 낱말 두 개 안에 "approv" 가 오는 구절(all approvals, all required approvals) — 을 함께 담은 줄이 1개 이상 있고
 AND (b) 두 절 어디에도 "team mode" 를 담으면서 승인 없이 또는 일부 승인만으로 병합한다는 문구
     (without approval, no approval, approvals not required·not needed·optional·unnecessary, optional approval, regardless of approval,
      at least one approval, a single approval, after·on·once·upon·with + any·one·an·a·some·the first + approval, majority)를 담은 줄이 없고
 AND 읽기 기록 $E/ac028-reading.md 가 존재하며, ac028-mode-lines.txt 의 모든 줄(로컬·템플릿)에 대해
     "team 모드 줄이면 전원 승인을 병합 조건으로 두고 그 조건을 낮추거나 빼는 말이 없는가 / personal·manual 줄이면 승인·리뷰 조건 없이 병합한다고 말하는가 /
      그 밖의 줄이면 어느 모드의 승인·리뷰 조건도 더하거나 빼지 않는가" 에 예로 답한다 (AC-GDP-029도 이 기록을 쓴다, §D.3)
```

전원 승인 구절을 (a)에 넣은 까닭은 "approvals are optional" 이나 "at least one approval" 처럼 "approv" 를 담기만 한 약한 문장이 (a)를 채우지 못하게 하려는 것이다. (b)의 낱말 목록은 부분 승인 표현을 모두 담지 못하므로 읽기 기록이 PASS의 전제다.

검출식은 "team mode" 두 낱말로 판정한다. "team" 한 낱말은 "teammates" 에 걸려 personal·manual 문장(승인할 팀원 없음)을 team 규칙으로 잘못 읽는다.

절 추출(대조·판정 공통):

```bash
awk '/^## PR Auto-Merge/{s=1; print; next} s && /^## /{exit} s' <manager-git.md> > <mg 절 파일>
awk '/^#### Step 3\.4/{s=1; print; next} s && /^(####|###) /{exit} s' <delivery.md> > <dl 절 파일>
```

대조(기준 트리, 템플릿 사본): mg 절 9줄(0.2.5: `$BASE` 로컬 절 — 로컬 166-174행 — 도 9줄이고 템플릿 절과 diff exit 0, `reanchor/mg-pram-*.md`). (a) mg 0(승인 조건 줄에 모드 이름이 없고 절 제목만 team을 말함), dl 0. (b) 0·0. 0.2.3 검출식으로 다시 재도 (a) 0·0, (b) 0·0.

판정(로컬·템플릿 각각):

```bash
awk '/^## PR Auto-Merge/{s=1; print; next} s && /^## /{exit} s' .claude/agents/moai/manager-git.md > $E/ac028-mg.md
test -s $E/ac028-mg.md
# 기대: exit 0
awk '/[Tt]eam mode/ && /--auto-merge/ && /(^|[^A-Za-z])[Aa]ll( [A-Za-z-]+)?( [A-Za-z-]+)? approv/ {c++} END{print c+0}' $E/ac028-mg.md > $E/ac028-a-mg.txt
awk '/[Tt]eam mode/ && /--auto-merge/ && /(^|[^A-Za-z])[Aa]ll( [A-Za-z-]+)?( [A-Za-z-]+)? approv/ {c++} END{print c+0}' $E/ac026-dl.md > $E/ac028-a-dl.txt
# 기대: 각각 1 이상
awk '/[Tt]eam mode/ && (/without (any |an )?approv/ || /no approv/ || /approv[a-z]* (are |is )?(not required|not needed|optional|unnecessary)/ || /optional approv/ || /regardless of approv/ || /at least one approv/ || /(a|one) single approv/ || /(after|on|once|upon|with) (any|one|an|a|some|the first) approv/ || /majority/) {print FILENAME ":" FNR ": " $0}' $E/ac028-mg.md $E/ac026-dl.md > $E/ac028-b.txt
test -e $E/ac028-b.txt
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
test -s $E/ac028-b.txt
# 기대: exit 1
awk '/[Aa]pprov/ || /[Rr]eview/ || /[Tt]eam mode/ || /[Pp]ersonal/ || /[Mm]anual/ || /--auto-merge/ {print FILENAME ":" FNR ": " $0}' $E/ac028-mg.md $E/ac026-dl.md > $E/ac028-mode-lines.txt
test -s $E/ac028-mode-lines.txt
# 기대: exit 0 — (a)와 AC-GDP-029 (a)가 성립하면 비지 않는다. 로컬·템플릿 각각의 이 목록 줄이 모두 읽기 기록의 대상이다
test -s $E/ac028-reading.md
# 기대: exit 0 — 읽기 기록이 없거나 목록의 한 줄이라도 답이 비면 AC-GDP-028·029 모두 PASS 불가
```

뮤턴트(0.2.3, 스크래치 픽스처 m028.md 15줄 — 1~6줄은 0.2.2 픽스처, 7~11줄은 감사 2회차 줄 글자 그대로, 12~15줄은 이번에 더한 줄):

```
  1 "In team mode, `--auto-merge` merges once checks pass."                          (승인 없음)
  2 "In team mode, `--auto-merge` merges regardless of approvals."                   (승인 무시)
  3 "In team mode, `--auto-merge` merges only after all approvals are obtained."     (올바름 — team 양성 대조)
  4 "In personal and manual modes, `--auto-merge` merges after all approvals."       (personal·manual 에 승인 조건)
  5 "In personal mode, `--auto-merge` merges without an approval condition."         (manual 누락)
  6 "In personal and manual modes, `--auto-merge` merges without an approval condition (no teammates to approve)." (올바름 — personal·manual 양성 대조)
  7 "In team mode, `--auto-merge` merges once CI passes; approvals are optional."    (감사)
  8 "In team mode, `--auto-merge` merges after at least one approval."              (감사)
  9 "In personal and manual modes, `--auto-merge` merges without requiring approval." (감사, 올바름)
 10 "In personal and manual modes, `--auto-merge` merges; approval is not needed."   (감사, 올바름)
 11 "In personal and manual modes, `--auto-merge` merges after a code review."       (감사, 리뷰 조건)
 12 "In team mode, `--auto-merge` merges after a single approval."                   (부분 승인)
 13 "In team mode, `--auto-merge` merges once any approval is recorded."             (부분 승인)
 14 "In personal and manual modes, `--auto-merge` merges after a code review; approval not required." (리뷰 조건 + 없음 표현)
 15 "In team mode, `--auto-merge` waits for all required approvals before merging."  (올바름 — team 변형)
명령: 위 판정의 AC-GDP-028 (a)·(b) awk 와 AC-GDP-029 (a)·(b) awk 를 {print FILENAME ":" FNR} 로 m028.md 에 실행
관측: 028 (a) 3·15 · 028 (b) 2·7·8·12·13 · 029 (a) 4·6·9·10·11·14 · 029 (b) 4·11·14
판정: 7줄이나 8줄만 담은 절은 028 (a) 0 이고 (b) 적중 → FAIL. 3줄만 담은 절은 028 (a) 1·(b) 0 → PASS.
      11줄만 담은 절은 029 (b) 적중 → FAIL. 6·9·10줄은 029 (b)에 걸리지 않음 → 이 줄만 담은 절은 PASS. 5줄만 담은 절은 029 (a) 0 → FAIL.
```

뮤턴트(0.2.4, 읽기 목록 선택자 — 스크래치 픽스처 m-n3.md 3줄):

```
  1 "In team mode, `--auto-merge` merges only after all approvals are obtained."  (양성 대조 — 두 선택자 모두 선택)
  2 "`--auto-merge` merges as soon as CI checks pass."                              (모드 이름·승인 낱말이 없는 병합 조건)
  3 "Unrelated line about the changelog."                                           (음성 대조)
명령: 0.2.3 선택자(끝 항목 /[Mm]anual/)와 위 판정의 선택자(|| /--auto-merge/ 추가)를 {print FNR ": " $0} 로 m-n3.md 에 실행
관측: 0.2.3 선택자 1 · 0.2.4 선택자 1·2 · 3줄은 둘 다 선택하지 않음
판정: 2줄은 이제 $E/ac028-mode-lines.txt 에 들어가 읽기 기록의 대상이 된다(0.2.3 선택자에서는 빠져 028·029 (a)·(b)만으로 통과할 수 있었다).
```

### AC-GDP-029 — personal·manual 모드: 승인 조건 없음

```
GIVEN AC-GDP-028 과 같은 두 절(로컬·템플릿)
WHEN 편집 뒤 검사하면
THEN (a) 두 절 각각에 "personal"·"manual"·"--auto-merge" 를 함께 담은 줄이 1개 이상 있고
 AND (b) 두 절 어디에도 personal 또는 manual 을 담으면서 다음 어느 하나에 해당하는 줄이 없다
     (1) "approv" 나 "review" 를 담는데 없음 표현(without [requiring] [any|an] approval·review, no approval·review, not required, not needed, no teammates)이 없는 줄
     (2) 없음 표현과 상관없이 조건 구절(after·once·until·pending·require·requires + [a|an|the|all|one|at least one] [code] + approval·review)을 담은 줄
 AND AC-GDP-028의 읽기 기록 $E/ac028-reading.md 가 존재하고 모든 질문에 답했다
```

대조(기준 트리, 템플릿 사본): (a) mg 절 0, dl 절 0. (b) 0·0. 0.2.3 검출식으로 다시 재도 (a) 0·0, (b) 0·0.

판정(로컬·템플릿 각각):

```bash
awk '/[Pp]ersonal/ && /[Mm]anual/ && /--auto-merge/ {c++} END{print c+0}' $E/ac028-mg.md > $E/ac029-a-mg.txt
awk '/[Pp]ersonal/ && /[Mm]anual/ && /--auto-merge/ {c++} END{print c+0}' $E/ac026-dl.md > $E/ac029-a-dl.txt
# 기대: 각각 1 이상
awk '(/[Pp]ersonal/ || /[Mm]anual/) && (((/[Aa]pprov/ || /[Rr]eview/) && !(/without (requiring )?(any |an )?(approv|review)/ || /no (approv|review)/ || /not required/ || /not needed/ || /no teammates/)) || /(after|once|until|pending|requires?) (a |an |the |all |one |at least one )?(code )?(approv|review)/) {print FILENAME ":" FNR ": " $0}' $E/ac028-mg.md $E/ac026-dl.md > $E/ac029-b.txt
test -e $E/ac029-b.txt
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
test -s $E/ac029-b.txt
# 기대: exit 1
test -s $E/ac028-reading.md
# 기대: exit 0
```

(b)의 (2)는 "approval not required" 같은 없음 표현을 덧붙여 리뷰·승인 조건을 숨긴 줄(14줄)을 잡는다. "without waiting for approval" 처럼 없음 표현 목록 밖의 올바른 문구는 위반으로 잡힌다(FAIL 쪽으로 기운다) — 받는 문구는 plan.md §B.14.

뮤턴트: AC-GDP-028의 m028.md 15줄과 관측을 함께 쓴다(위 코드 블록).

### AC-GDP-030 — 명령 원본 편집 뒤 게시본 발행

```
GIVEN 명령 원본 argument-hint 편집(로컬 .claude/commands/moai/sync.md, 템플릿 .claude/commands/moai/sync.md.tmpl)
WHEN run-phase 가 사전 점검에서 발행 점검의 양성 대조를 실행하고,
     원본 편집 직후 make commands-emit 과 make commands-emit-check 를 실행하고,
     원본 편집 커밋이 만들어진 뒤 게시본 변화를 판정하면
THEN (c) 양성 대조: 템플릿 게시본 한 파일을 잠시 바꾼 상태에서 make commands-emit-check 가 exit 1 이고,
     백업으로 되돌린 뒤 cmp 가 exit 0, 다시 실행한 make commands-emit-check 가 exit 0 이며, 게시본 경로의 git status 가 비어 있고
 AND make commands-emit 이 exit 0, 이어서 make commands-emit-check 가 exit 0 이고
 AND 게시본 변화가 아래 두 경우 중 하나로 판정되어 progress §E.2 에 경우 이름과 근거 출력과 함께 기록된다:
     (A) 변화 없음 — 기준 트리 대비 템플릿·로컬 게시본(.agents/skills/) 변경 경로 목록이 비어 있다
     (B) 변화 있음 — 변경 경로가 템플릿 게시본 moai-sync/SKILL.md 와 로컬 사본 .agents/skills/moai-sync/SKILL.md 뿐이고,
         두 게시본 파일을 바꾼 커밋이 모두 템플릿 명령 원본 sync.md.tmpl 을 바꾼 커밋과 같으며,
         로컬 게시본 사본이 템플릿 게시본과 diff exit 0 이고,
         게시본이 AC-GDP-026 (ii)·AC-GDP-027 규칙을 어기는 줄을 담지 않는다
```

**plan 작성 시점 예상: 경우 (A).** 근거: 발행기는 명령 원본의 `argument-hint`·`allowed-tools` 를 게시본에 옮기지 않는다("Claude-only keys and are NOT carried into the published skill", `internal/template/commandemit/loader.go:4-6`). 게시본은 생성 머리말·`name`·영어 `description`·원본 본문만 담는다(`emit.go:122-130` `renderSkill`). X2 편집은 3행만 바꾸고 2행 `description` 과 본문을 바꾸지 않는다. 기준 트리의 템플릿·로컬 게시본은 서로 바이트 동일하고(각 7줄) `merge` 를 담지 않는다(`grep -ci merge` 0·0). 예상은 판정을 대신하지 않는다 — 두 경우를 모두 명령으로 판정한다. 발행기는 템플릿 게시본만 쓰므로(`golden_test.go:28` `templatesDir = "../templates"`, 갱신 분기 `:57-74`) 경우 (B)가 되면 로컬 게시본 사본은 템플릿 게시본을 그대로 복사해 같은 커밋에 넣는다.

발행 명령 근거: `make commands-emit` = `COMMAND_EMIT_UPDATE=1 go test ./internal/template/commandemit/... -run TestGoldenCommittedArtifactsMatchEmission`(`Makefile:51-52`), `make commands-emit-check` = 같은 테스트를 `COMMAND_EMIT_UPDATE=` 로 비워 읽기 전용 실행(`Makefile:58-60`, 실패 시 "command-skill drift" 를 내고 exit 1). `TestCommandSourcesUnmodified`(`golden_test.go:94-106`)는 한 실행 안의 발행 전후 해시 비교라 원본 편집만으로는 실패하지 않는다 — 원본 편집 뒤 빨강이 날 수 있는 점검은 게시본 대조뿐이고, 그래서 아래 양성 대조가 게시본을 바꿔 빨강을 확인한다.

양성 대조(run-phase 사전 점검, 원본 편집 전):

```bash
cp $T/.agents/skills/moai-sync/SKILL.md /tmp/t622-moai-sync-SKILL.backup.md
printf '\n' >> $T/.agents/skills/moai-sync/SKILL.md
make commands-emit-check > $E/ac030-red.txt 2>&1
# 기대: exit 1 — 발행 점검이 게시본 드리프트를 실제로 잡는다
cp /tmp/t622-moai-sync-SKILL.backup.md $T/.agents/skills/moai-sync/SKILL.md
cmp /tmp/t622-moai-sync-SKILL.backup.md $T/.agents/skills/moai-sync/SKILL.md
# 기대: exit 0
make commands-emit-check > $E/ac030-control-green.txt 2>&1
# 기대: exit 0
git status --porcelain -- $T/.agents/skills/ .agents/skills/ > $E/ac030-control-clean.txt
test -e $E/ac030-control-clean.txt
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
test -s $E/ac030-control-clean.txt
# 기대: exit 1 — 되돌림 뒤 게시본 경로에 변경 없음
git ls-files -- $T/.agents/skills/moai-sync/SKILL.md .agents/skills/moai-sync/SKILL.md > $E/ac030-tracked.txt
/usr/bin/grep -c '' $E/ac030-tracked.txt > $E/ac030-tracked-count.txt
# 기대: 2 — 아래 git status·git diff·git log 의 경로 지정이 추적 파일을 실제로 가리킨다.
# 경로가 틀리면 빈 목록이 "변화 없음"(경우 A)으로 읽히므로 이 개수가 2가 아니면 판정 불가. plan 작성 시점 측정: 두 경로 모두 출력
```

판정(원본 편집 직후, 커밋 전):

```bash
make commands-emit > $E/ac030-emit.txt 2>&1
# 기대: exit 0
make commands-emit-check > $E/ac030-check.txt 2>&1
# 기대: exit 0
git status --porcelain -- $T/.agents/skills/ .agents/skills/ > $E/ac030-emit-status.txt
test -e $E/ac030-emit-status.txt
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
# 비어 있으면 경우 (A) 후보, 비어 있지 않으면 경우 (B) 후보 — 비어 있지 않으면 바뀐 게시본을 원본과 함께 스테이징한다
```

판정(원본 편집 커밋 뒤):

```bash
git diff --name-only develop...HEAD -- $T/.agents/skills/ .agents/skills/ > $E/ac030-changed.txt
# (A): 빈 파일. (B): 두 줄 이하이며 internal/template/templates/.agents/skills/moai-sync/SKILL.md, .agents/skills/moai-sync/SKILL.md 만
test -e $E/ac030-changed.txt
# 기대: 두 경우 모두 exit 0 — 파일이 없으면 판정 불가(PASS 아님). (A)이면 이어서 test -s $E/ac030-changed.txt → exit 1
git log --format=%H HEAD --not develop -- $T/.claude/commands/moai/sync.md.tmpl > $E/ac030-src-commits.txt
git log --format=%H HEAD --not develop -- $T/.agents/skills/moai-sync/SKILL.md .agents/skills/moai-sync/SKILL.md > $E/ac030-artifact-commits.txt
# 0.2.6: 범위는 관례의 "카드 변경 범위 판정" 을 따른다(리터럴 $BASE..HEAD 는 흡수된 develop 커밋을 담는다 — 재고정 감사 D1).
#   흡수 병합은 경로 기준 이력 단순화에서 한쪽 부모와 같으면 빠진다. 흡수 병합에서 이 경로의 충돌을 풀었다면 병합이 목록에 오를 수 있고,
#   그때는 두 목록 모두에 오르므로 아래 부분집합 판정은 그대로 성립한다.
#   0.2.6 측정: 현재 트리·흉내 낸 흡수 커밋 모두 ac030-src-commits 빈 목록(편집 전), 같은 형태 양성 대조
#   `git log --format=%H 97ef8e3023e9e7a29e7478289b69d28796dddbd7 --not ee99507fbe3b4a22c6a0a74815723d222dfdc04d -- $T/.claude/rules/moai/core/agent-common-protocol.md` → 1줄(reanchor-fix/ac030-*.txt)
# (A): ac030-artifact-commits.txt 빈 파일. (B): ac030-artifact-commits.txt 의 모든 줄이 ac030-src-commits.txt 에 있음(같은 커밋)
test -e $E/ac030-artifact-commits.txt
# 기대: 두 경우 모두 exit 0 — 파일이 없으면 판정 불가(PASS 아님). (A)이면 이어서 test -s $E/ac030-artifact-commits.txt → exit 1
test -s $E/ac030-src-commits.txt
# 기대: 두 경우 모두 exit 0 — 원본 편집 커밋이 실제로 있다(없으면 아래 부분집합 판정이 공허해진다)
git diff --name-only develop...HEAD -- $T/.claude/commands/moai/sync.md.tmpl .claude/commands/moai/sync.md > $E/ac030-src-changed.txt
/usr/bin/grep -c '' $E/ac030-src-changed.txt > $E/ac030-src-changed-count.txt
# 기대: 2 — 같은 git diff 경로 형태가 바뀐 파일을 실제로 본다(ac030-changed.txt 가 빈 것이 경로 오류가 아니라는 양성 대조)
/usr/bin/grep -v -x -F -f $E/ac030-src-commits.txt $E/ac030-artifact-commits.txt > $E/ac030-orphan-artifact-commits.txt
test -e $E/ac030-orphan-artifact-commits.txt
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
test -s $E/ac030-orphan-artifact-commits.txt
# 기대: 두 경우 모두 exit 1 — 원본 편집 커밋에 없는 게시본 커밋이 없다. (B)이면 추가로 test -s $E/ac030-artifact-commits.txt → exit 0
# 뮤턴트(0.2.2 스크래치 측정): 원본 {aaa1111}·게시본 {bbb2222} → 위 grep exit 0(FAIL 검출); 게시본 {aaa1111} → exit 1; 빈 원본 목록 → exit 0, 1줄 출력(FAIL)
test -s $T/.agents/skills/moai-sync/SKILL.md
# 기대: exit 0 — 아래 플래그 검사가 빈 파일 위에서 통과하지 않는다
diff .agents/skills/moai-sync/SKILL.md $T/.agents/skills/moai-sync/SKILL.md > $E/ac030-published-lt.diff
# 기대: 두 경우 모두 exit 0
awk '/(^|[^A-Za-z-])--merge([^A-Za-z-]|$)/ && (!/--merge([^A-Za-z-].*)?[Dd]eprecated alias (of|for) .?--auto-merge([^A-Za-z-]|$)/ || /(not|no longer|never) (a |an |the )?[Dd]eprecated/ || /[Uu]n-?deprecat/) {print FNR ": " $0}' $T/.agents/skills/moai-sync/SKILL.md > $E/ac030-published-flags.txt
test -e $E/ac030-published-flags.txt
# 기대: exit 0 — 파일이 없으면 판정 명령이 돌지 않은 것이다(판정 불가, PASS 아님)
test -s $E/ac030-published-flags.txt
# 기대: 두 경우 모두 exit 1
```

판정 기록: `$E/ac030-outcome.md` 에 경우 이름(A 또는 B), `ac030-changed.txt`·`ac030-src-commits.txt`·`ac030-artifact-commits.txt` 내용을 적는다. 경우를 가르는 명령 출력 없이 경우 이름만 적은 기록은 PASS가 아니다.

뮤턴트(판정 논리): 원본을 커밋한 뒤 별도 커밋으로 게시본만 바꾼 이력 → `ac030-artifact-commits.txt` 의 SHA가 `ac030-src-commits.txt` 에 없음 → FAIL. 게시본을 다시 만들지 않고 원본의 `description` 을 바꾼 상태 → `make commands-emit-check` exit 1 → FAIL. 템플릿 게시본만 바뀌고 로컬 사본을 두고 온 상태 → `ac030-published-lt.diff` exit 1 → FAIL.

## §D.2 경계 사례

- `grep -c` 는 줄 수를 센다. SHA는 AC-GDP-015에서 낱말 단위로 센다.
- 빈 결과 판정의 존재 확인 뮤턴트(0.2.4, 스크래치 실행): 판정 명령이 돌지 않아 `ac028-b.txt` 가 없는 상태. 명령 `rm -f <스크래치>/ac028-b.txt` → `test -e <스크래치>/ac028-b.txt; echo "exit=$?"` → `test -s <스크래치>/ac028-b.txt; echo "exit=$?"`. 관측 `exit=1` · `exit=1`. 판정: 존재 확인이 exit 1 이라 FAIL(판정 불가, PASS 아님). 0.2.3의 빈 결과 확인만 있었다면 `test -s` exit 1 이 기대값과 같아 PASS로 읽혔다.
- `sed -n '/A/,/B/p'` 의 끝 제목이 편집으로 바뀌면 절이 파일 끝까지 늘어난다. `awk` 절 추출은 시작 표지가 사라지면 빈 파일을 낸다 — 빈 파일은 판정 불가로 기록한다. 추출 표지(`## Synchronization`, `## PR Auto-Merge`, `### Pre-Spawn Sync Check`, `### Pre-Edit Sync Check`, `#### The sweep prohibition`, `#### Step 3.4`, `##### Worktree Context Detection`, AC-GDP-026의 아홉 표지)는 편집 뒤에도 남아야 한다. `manager-git.md` 절 제목은 "## PR Auto-Merge" 로 시작하기만 하면 뒤의 괄호를 바꿔도 된다.
- 자리표시 기준(AC-GDP-007~012, 017~024)은 판정하지 않고 N/A로 기록한다.
- 의도된 사본 차이가 있는 여덟 파일(0.2.5 `$BASE` 측정)은 편집으로 줄 수가 바뀌면 차이 줄번호가 밀린다. AC-GDP-013은 줄번호 머리를 빼고 본문만 비교한다. 명령 원본은 로컬 `.md` 와 템플릿 `.md.tmpl` 로 파일 이름 자체가 다르므로 두 경로를 글자 그대로 짝지어 비교한다.
- 검출식에 `\b` 를 쓰지 않는다(POSIX ERE에서 단어 경계가 아니다). AC-GDP-026의 `--merge` 낱말 경계는 앞뒤 문자 클래스로 표현하며 `--merged-only`·`--auto-merge` 는 걸리지 않고, `[--merge]` 처럼 대괄호에 둘러싸인 형태는 걸린다.
- AC-GDP-025의 Frozen 줄 판정은 `[ZONE:Frozen]` 을 담은 diff 줄을 먼저 모은 뒤 `-`·`+` 로 시작하는 줄만 고른다. 첫 grep 에 `-n` 을 붙이면 공허하게 통과한다.
- AC-GDP-026~029는 줄 단위다. 한 조건을 여러 줄에 나눠 적으면 (a)·(i)가 0이 되어 FAIL로 기울고, 검출식이 예상하지 않은 표현은 통과할 수 있다(spec.md §E.2). 그 틈 때문에 AC-GDP-026·028·029는 읽기 기록을 PASS 전제로 둔다(§D.3). AC-GDP-027은 자동 검출만으로 판정한다. 검출식이 받는 문구와 떨어뜨리는 문구의 경계는 plan.md §B.14에 적었다.
- AC-GDP-030의 양성 대조는 추적 파일을 잠시 바꾼다. 되돌림을 `cmp` 와 `git status` 로 확인하지 못하면 사전 점검을 멈추고 보고한다.
- `go test -run` 선택자는 `^…$` 로 고정하고, 최상위 PASS 줄 수와 범위 파일 하위 테스트 흔적을 함께 본다. AC-GDP-013 (e)의 미러 테스트 선택자는 기준선 측정과 같은 문자열(`'TestSanitizedPairParity|TestRuleTemplateMirrorDrift'`, 고정 없음)을 그대로 쓴다 — 기준선과 다른 선택자로 돌리면 집합 비교의 기준이 달라진다.
- (0.2.5, 0.2.6) AC-GDP-002 는 `$BASE` 에서 이미 초록인 회귀 방지 판정이다. 0.2.6 에서 (a) 순서와 (b) 절 보존으로 나눴고, 뮤턴트 i~viii 는 모두 AC-GDP-002 FAIL 이다 — (a)가 i·iii·iv·v·vi 를, (b)가 여덟 개 모두를 잡는다(§D.1 AC-GDP-002 표). 절 밖 뒤집기(ix)는 AC-GDP-002 가 보지 않는다(AC-GDP-016·신선도 점검 몫). 이 기준이 초록이라는 사실은 이 카드가 REQ-GDP-002 를 이뤘다는 증거가 아니다 — 이룬 것은 t635·dr0911 이다.
- (0.2.6) "이 카드가 바꾼 것" 을 재는 판정(AC-GDP-014·015·016·025·030)은 리터럴 `$BASE` 가 아니라 `develop...HEAD`·`HEAD --not develop` 으로 잰다. 0.2.5 의 "BASE 이후 develop 재흡수가 없다" 는 전제는 계획 시점에 이미 거짓이었고(로컬 develop 이 카드보다 앞서 있었다) 지웠다. 이 판정들은 병합 전 전용이며, 범위 대조가 0줄이면 "측정 불가" 다(공통 변수와 명령 관례).
- (0.2.6) 판정 전 스냅숏 신선도 점검이 비어 있지 않으면, 목록의 파일에 기댄 판정은 그 파일의 기준 사본을 `$CARD_BASE` 에서 다시 반출하고 기준선을 다시 잰 뒤에 한다. 낡은 스냅숏으로 낸 결과는 PASS 도 FAIL 도 아니다.

## §D.3 품질 게이트

- 사전 점검의 양성 대조가 모두 기대값을 냈다는 기록이 있어야 판정이 유효하다(AC-GDP-030의 발행 점검 대조 포함).
- (0.2.6) 판정을 결정하는 실행마다 `$E/card-base.txt`(1줄), `$E/card-range-names.txt`(1줄 이상), `$E/snapshot-stale.txt`(빈 파일, 아니면 재반출 기록)가 남아 있어야 한다. 통합 창에서 develop 을 흡수한 뒤 다시 재면 이 세 파일도 그 실행에서 다시 만든다 — 흡수 전 파일을 흡수 뒤 판정의 근거로 쓰지 않는다.
- **읽기 단계가 있는 기준(AC-GDP-001, AC-GDP-006, AC-GDP-026, AC-GDP-028, AC-GDP-029)은 읽기 기록 파일(`$E/ac001-reading.md`, `$E/ac006-reading.md`, `$E/ac026-reading.md`, `$E/ac028-reading.md` — AC-GDP-029는 `ac028-reading.md` 를 함께 쓴다)이 존재하고 대상 문단·절·줄의 모든 질문에 답했을 때만 PASS다.** 자동 검출이 통과해도 읽기 기록이 없거나 한 줄이라도 답이 비면 PASS가 아니다. AC-GDP-026의 대상 줄은 `$E/ac026-merge-lines.txt`, AC-GDP-028·029의 대상 줄은 `$E/ac028-mode-lines.txt` 에 로컬·템플릿 사본마다 뽑은 줄 전부다.
- `go test` 선택 실행이 최상위 PASS 줄 2개를 내지 않으면 합격이 아니다. 로컬 전체 스위트는 돌리지 않는다.
- 미러 테스트 비회귀 점검(AC-GDP-013 (e))은 M4 와 M6 에서 각각 한 번씩 돌고, 두 번 모두 새 FAIL 0·잃은 PASS 0 이어야 한다. 미러 테스트의 exit 1 은 기준선과 같은 이유(범위 밖 `spec-workflow.md`)일 때 합격을 막지 않는다.

## §D.4 완료 정의 (Definition of Done)

- 판정 대상 기준 AC-GDP-001~006, 013~015, 025~030이 PASS이고 AC-GDP-016이 PASS 또는 사유 기록. 자리표시 기준은 N/A. AC-GDP-013 은 (a)~(e) 모두 PASS 여야 한다.
- AC-GDP-030의 경우 이름(A 또는 B)과 근거 출력이 progress 기록에 남는다.
- 모든 증거 파일이 `.moai/reports/t622/run/` 에 커밋되어 인용 경로가 해석된다.
