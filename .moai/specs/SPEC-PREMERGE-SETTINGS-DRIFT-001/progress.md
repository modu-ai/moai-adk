# progress.md — SPEC-PREMERGE-SETTINGS-DRIFT-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-06
artifacts: spec.md + plan.md + acceptance.md + progress.md (Tier M) — v0.6.2
baseline: worktree `.claude/worktrees/t488`, branch `WT-premerge-drift-assert`
spec_id_check: `[[ "SPEC-PREMERGE-SETTINGS-DRIFT-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS` (실행 출력)
open_clarifications: 0 — 미해결 `[NEEDS CLARIFICATION:` 마커 0건. 기본 자세는 운영자 결정으로 계열 (가) 기본 OFF + 켜면 거절(2026-09-06), 킬 스위치 키는 `workflow.settings_drift_gate.enabled` 새 블록. 두 결정에 묶여 있던 다섯 곳 전부 처리됨(plan.md §B D2 표).
origin: SPEC-SETTINGS-ORIGIN-001 (card t487) verdict §Q3 recommendation C5

### 수리 회차 1 (2026-09-06) — plan 감사 FAIL 대응

audit_report: `.moai/reports/t488/plan-audit.md` (판정 FAIL, 종합 0.725 / Tier M 임계 0.80)
repaired_at_head: `615d18c1f` · worktree `.claude/worktrees/t488` · branch `WT-premerge-drift-assert`
artifact_version: v0.1.0 → v0.2.0

| 발견 | 수리 |
|---|---|
| D1 — REQ-PSD-007을 반증하는 AC 없음 | `AC-PSD-007`에 (d) 추가: 실행 전후 `git log --oneline -1` 동일 + 보존 디렉터리 ignored 분류 + 기록된 명령 목록에 `commit`/`push` 0건 |
| D2 — `AC-PSD-005`에 관측 지점 없음 | REQ-PSD-013에 argv 단일 원천 조항 추가, M1 서명에 명령 실행기 노출, `AC-PSD-005`를 "실행 경계에서 기록된 argv"로 재작성(빌더 반환값 단정 금지) |
| D3 — "원리적으로 검출 불가" 과장 | `acceptance.md` AC-PSD-005 / `plan.md` M2 주석 / `plan.md` §E 세 곳을 "신뢰성 있게는 구별되지 않는다(인덱스 쓰기 여부는 stat 캐시 상태에 좌우)"로 정정 |
| D4 — `AC-PSD-007(c)`가 파일명 규칙과 충돌 | D3 파일명 규칙을 밀리초 해상도 + **순번 접미 충돌 규칙**으로 바꾸고 (c)를 그 규칙에 맞춰 재작성 |
| D5 — 뮤턴트 2 포착기 오귀속 | `plan.md` M2 표에 층(술어/게이트) 열 추가, 판정 반전의 포착기를 F1 → `AC-PSD-009`로 정정하고 F1이 초록으로 남는 이유를 명시 |
| D6 — 안전 규칙 2건이 어느 층에도 없음 | REQ-PSD-014(술어 실패 ≠ 통과) / REQ-PSD-015(보존 실패 ≠ 통과) 신설 + 반증 `AC-PSD-011`/`AC-PSD-012` 신설. 두 AC는 `integration-lock.json`이 아니라 `preflight`의 판정 필드에 걸어 기본 자세 결정과 독립시켰다 |
| D7 — 새 Go 코드의 해시 | `spec.md` REQ-PSD-005/008, `plan.md` D3 파일명·원장 필드, `plan.md` D4 `--json` 키를 sha256으로. 추가로 `acceptance.md` AC-PSD-007/008/009의 md5도 같은 이유로 이관(요구 성질 기준 스윕). **`spec.md` §1의 t334 실측 3줄과 t487 §Q3 인용은 실행된 명령의 기록이므로 미수정** |
| D8 — M6 미러 대상 유보 | 실측 판정으로 대체: `kanban-dispatch.md` 미러 존재(Template-First 적용) / `CLAUDE.local.md` 미러 없음(의도적 로컬 전용). 템플릿 중립성 제약 명기 |
| marker 2 — 킬 스위치 키 위치 | 해소. `workflow.settings_drift_gate.enabled` 새 블록(근거: `internal/config/types.go:650-659`의 deny 층 계약 + REQ-PSD-010의 중복 적재 회피) |
| marker 1 — 기본 자세 | **열어 둔다.** 2갈래 → 3계열로 재작성(차단/권고 축 x 기본 ON/OFF 축; 실측 선례 수 4 / 2 / 0). 권고하지 않음 |

### 수리 회차 5 (2026-09-06) — 확인 감사 PASS 후 잔여 종결

audit_report: 확인 감사 **PASS · 0.930** (0.725 → 0.875 → 0.895 → 0.930). must-pass 7/7, 미해결 마커 0.
artifact_version: v0.5.0 → v0.6.0

| 항목 | 처리 |
|---|---|
| **상시 조항 어순**(리드가 가장 무겁게 본 것) | 착지본은 이미 내용 기반이었다 — "성공했고 **그 결과가 퇴화하지 않았음**"이라는 복합문이었고 문단의 결론 문장이 "대조는 종료 코드가 아니라 **사전 측정값의 내용**"이었다. 즉 `err != nil`로 환원되지 않는다. 그럼에도 **어순을 바꿨다**: 내용 대조를 첫 문장에 놓고, 성공 검사(종료 코드·`err != nil`)를 명시적으로 배제했다. 이 조항은 미래의 모든 AC에 일반화되는 장수 객체라, 급히 읽는 사람이 앞머리의 "성공"에 닻을 내리는 것을 막아야 한다. 좁은 예외(빈 모집단)는 **예외로 표시**하고 남용 금지를 붙였다 |
| **R1** `plan.md:180` | `drift`가 `true` → `status`가 `"drift"`. 동반 단정 `match_count: 1`은 **그대로 옳다**(측정은 성공했고 실패한 것은 보존뿐이므로 상태는 `drift`이고, `match_count` 생략은 `undetermined`에만 걸린다) — 그 이유를 문장에 적었다 |
| **R3** `acceptance.md:46` | "거절 판정은 그대로 유지" → "**drift 판정**은 그대로 유지(`status`가 `\"drift\"`로 남는다)". 유지되는 것은 거절이 아니라 판정이며, 거절은 킬 스위치에 달려 기본 설정에서는 애초에 일어나지 않는다 |

**스윕이 잡은 것 — 요구층, R1/R3보다 무겁다.** `spec.md` `REQ-PSD-008`이 "drift가 검출되면 그 명령은 창을 기록하지 않고 **거절한다**"로 무조건 거절을 요구하고 있었다. REQ-PSD-016(킬 스위치는 거절 층만 게이트)과 요구층에서 정면으로 모순이고, REQ-008만 읽은 구현자는 **운영자가 기각한 기본 거절을 그대로 만든다**. 거절만 `Where` 능력 게이트로 조건화하고 보고는 비게이트로 분리했다. 같은 모양으로 `REQ-PSD-009`도 조건화했고(거절 층이 꺼져 있으면 우회 플래그는 아무것도 우회하지 않으며 우회로 기록되지도 않는다 — 우회할 거절이 없는데 우회로 적으면 기록이 거짓말을 한다), 그 조항이 산문으로만 남지 않도록 `AC-PSD-013`에 (e)를 붙여 반증 가능하게 했다.

스윕 범위: 불리언 carrier 잔여(`drift`가/는 true·false, `"drift": bool`)와 거절 기본 전제를 spec.md / plan.md / acceptance.md 전수. plan.md의 `거절` 언급 13건은 전부 조건화돼 있거나 설계 근거·이력이라 손대지 않았다.

budget: 요구 16 = Tier M 상한 16(포화, 신설 없음), 수용 13(신설 없음).

### 수리 회차 4 (2026-09-06) — 픽스처 실수 모양 명시 + 중복 적용 회피 기록

artifact_version: v0.4.0 → v0.5.0

리드가 이 회차를 "운영자의 marker 1 답을 적용하는 회차"로 지시했으나, **그 답은 v0.4.0에서 이미 적용돼 있었다.** 지시에 인용된 측정(`NEEDS CLARIFICATION` 1건, `plan.md:40`이 기본 `true`)은 v0.4.0 편집 **이전** 시점의 것이고, 디스크를 다시 재서 확인했다.

```
$ grep -c 'NEEDS CLARIFICATION' plan.md   → 0
$ grep -n 'settings_drift_gate' plan.md
  :40  ... 계열 (가) 기본 OFF + 켜면 거절 ... 기본값은 `false` ...
  :72  | 기본값 상수 | workflow.settings_drift_gate.enabled 기본 `false` |
```

**양쪽 모두 이미 움직인 상태를 서술했다.** 이 세션의 완료 보고가 "marker 1 열림"이라고 적은 시점에도, 그 뒤 리드가 같은 것을 판독한 시점에도 그 서술은 참이었고 몇 분 뒤 거짓이 됐다. 두 서술이 일치한 것이 교차 확인처럼 보였지만, 같은 순간의 같은 상태를 두 번 읽은 것이다. 살아 있는 작성자가 있는 트리에서는 순간 판독이 어느 쪽에도 확정 상태가 아니며, 자기 작업을 보고하는 작성자에게도 그렇다. 이 문단의 완료 보고 쪽 사실은 아티팩트가 아니라 세션 대화에 있어 트리에서 검증되지 않는다.

그래서 marker 1 관련 편집은 **중복 적용하지 않았다.** 이미 적용된 결정을 다시 쓰면 v0.4.0에서 함께 처리한 다섯 곳(§ marker 1 종결 표)이 어긋날 위험만 생긴다. 같은 이유로 REQ-PSD-014 3상태 명명, REQ-PSD-016, `AC-PSD-013`도 v0.4.0 산물이며 이 회차에서 손대지 않았다.

이 회차의 실제 변경은 하나다 — M3 픽스처 항목에 **"bare 원격은 만들되 브랜치 push를 잊는" 실수 모양**과 구성 직후 `ls-remote` 비어 있지 않음 확인을 명시했다. 원격 생성은 눈에 보이는 단계이고 push는 잊기 쉬운 단계인데 잊어도 오류가 나지 않아서, 이 확인이 없으면 `AC-PSD-007(d-3)`의 push 반증이 조용히 무검사 상태로 남는다.

다만 위 "변경은 하나"라는 범위 주장은 **기계 비교로 뒷받침되지 않는다** — v0.4.0은 커밋된 적이 없어 그 내용이 남아 있지 않고, 따라서 v0.4.0 → v0.5.0 델타는 편집을 수행한 쪽의 판독이다. 커밋 `d02db303b`(plan 산출물 4개 + 판정서 4개) 이후로는 기계 비교가 가능하다.

### marker 1 종결 (2026-09-06) — 운영자 결정 [v0.4.0에서 적용]

decision: 계열 **(가) 기본 OFF + 켜면 거절**. `workflow.settings_drift_gate.enabled` 기본 `false`, 게이트되는 것은 거절 층 하나.
operator_rationale: t334는 인스턴스 2·실측 피해 0이라 배포 기본 거절의 근거로 얇다. 9일 실명의 원인은 "막지 않아서"가 아니라 "아무도 보지 않아서"이고, (가)에서도 검출·보존·원장은 매 `acquire`마다 돈다.

| 묶여 있던 곳 | 처리 |
|---|---|
| 기본값 상수 | `enabled` 기본 `false` (plan.md §B D2) |
| M5 `acquire` 동작 | 기본은 창을 내주고 보고, 켠 설정에서만 거절 — 두 갈래로 재작성 |
| `AC-PSD-009` / `AC-PSD-010` | Given에 `enabled: true` 명시. AC-009에는 양성 대조도 함께 부여(조항 예외 해소) |
| M2 뮤턴트 2 포착 픽스처 | 거절 층을 켠 설정에서 구성 — 확정형으로 |
| 부재 단정 규율의 AC-009 예외 | 삭제. 이제 예외 없음 |

**결정이 새로 요구한 것**: REQ-PSD-016(킬 스위치는 거절 층만 게이트) + `AC-PSD-013`. 운영자 근거의 핵심이 "검출·보존·원장은 계속 돈다"인데, 킬 스위치를 게이트 전체에 건 구현은 그 전제를 무너뜨리면서 기존 AC를 전부 통과한다. 산문으로 두지 않고 요구·반증으로 못박았다.

**REQ-PSD-014 3상태 판정(리드 지시)**: 종전 문안은 부정형뿐이라 두 상태 표면을 허용했다. `--json`의 판정 carrier가 불리언 `drift`였으므로, "통과 아님"을 만족시키는 값이 `drift`밖에 없어 측정 실패가 오탐으로 바뀌거나, 필드를 생략해 소비자 기본값이 통과로 읽는 경로가 열려 있었다. `clean` / `drift` / `undetermined`를 이름으로 요구하도록 REQ-PSD-014를 고치고, REQ-PSD-012에 판정 상태 필드 상시 포함을 넣고, plan.md D4의 `--json`을 `status` carrier로 바꿨다. `AC-PSD-011`은 `status == "undetermined"`를 **적극 단정**한다.

budget: 요구 16 = Tier M 상한 16(포화), 수용 13 ≤ 16.

### 수리 회차 3 (2026-09-06) — 3회차 감사 대응

audit_report: `.moai/reports/t488/plan-audit-iter3.md` (0.895 / 임계 0.80 초과. FAIL 사유는 MP-7 운영자 관문 + 2회차 수리가 심은 N5-N7)
repaired_at_head: `615d18c1f` · worktree `.claude/worktrees/t488` · branch `WT-premerge-drift-assert`
artifact_version: v0.3.0 → v0.4.0

| 발견 | 수리 |
|---|---|
| N5 — (d-3)이 어느 저장소를 재는지 미명시 | 커밋 반증을 **대상 트리 + primary 루트 양쪽**의 HEAD로 확장. 보존 사본이 primary에 살므로 위험한 커밋은 그쪽에서 일어난다. 픽스처에서 두 루트가 일치하면 그 사실을 명시하도록 요구(우연을 성질로 착각하지 않기 위해) |
| N6 — 동일성 단정 자체가 공허할 수 있음 | 부재 단정 규율에 **동일성 계열**을 흡수하고, 대조를 **종료 코드가 아니라 사전 측정값의 내용**으로 규정. (d-3) 두 측정과 `AC-PSD-008`에 각각 적용. 정상 상태에서 비어 있을 수 있는 모집단((d-4))은 실행 성공 대조 + 한계 명기로 분리 |
| N7 — 우회 뮤턴트가 열거에 없음 | `plan.md` M2 표에 한 행 추가(게이트 경로 실행기 우회 / 게이트층 / 포착기 `AC-PSD-007(d-3)`) + 왜 (d-1)/(d-2)로는 안 잡히는지 명시, DoD "뮤턴트 5종" → **6종** |
| N8 (optional) | `AC-PSD-009` 예외를 두 곳에 기록 — 부재 단정 규율 본문과 marker 1의 "답이 걸리는 곳" 목록. 조항이 침묵의 예외로 닳지 않게 한다 |

**실측 근거(이 트리에서 직접 관측, 감사는 재현하지 않았다고 §9에 밝힌 항목).** `git ls-remote --heads`를 세 경우로 재서 N6의 처방을 교정했다.

| 경우 | rc | stdout |
|---|---|---|
| 존재하지 않는 원격 | 128 | 0바이트 (stderr에 `fatal:` 2줄) |
| **브랜치 없는 정상 bare 원격** | **0** | **0바이트, 오류 없음** |
| 대조군(이 저장소) | 0 | 9500바이트 |

두 번째가 결정적이다 — 종료 코드 검사로는 퇴화를 못 잡으므로, 대조는 반드시 **사전 출력의 내용**이어야 한다. 대조군을 함께 잰 이유는 앞의 두 빈 출력이 명령 모양이 틀려서 나온 것이 아님을 보이기 위해서다.

### 수리 회차 2 (2026-09-06) — 2회차 감사 대응

audit_report: `.moai/reports/t488/plan-audit-iter2.md` (0.875 / 임계 0.80 초과, D1-D8 전부 종결. FAIL 사유는 MP-7 운영자 관문 + 수리가 심은 N1-N3)
repaired_at_head: `615d18c1f` · worktree `.claude/worktrees/t488` · branch `WT-premerge-drift-assert`
artifact_version: v0.2.0 → v0.3.0

| 발견 | 수리 |
|---|---|
| N1 — 부재 단정이 빈 기록에서 공허 | `AC-PSD-007(d)`를 (d-1) 양성 대조 / (d-2) 기록 위 부재 / (d-3) **기록과 무관한 직접 관측**(`git rev-parse HEAD` 불변 + bare 원격 `ls-remote` 불변) / (d-4) 스테이징 부재 넷으로 재구성. `AC-PSD-008` 세 번째 단정에도 양성 대조 추가. 감사가 제시한 양성 대조만으로는 닫히지 않아 (d-3)과 REQ-PSD-013 확장을 함께 넣었다 |
| N2 — `ignored 분류` 단정이 픽스처에서 거짓 빨강 | 삭제하고 (d-4) "보존 사본이 스테이징되지 않았다"로 교체. 픽스처 `.gitignore` 유무와 무관하고, 재는 대상이 저장소 설정이 아니라 게이트 행동이다 |
| N3 — `AC-PSD-012` 관측 지점(M4) vs 청구 마일스톤(M3) | M3은 **동작**만 넣고 청구 제거, M4가 `AC-PSD-012`를 청구. 겸해 M4를 Medium → **High**로 올렸다(REQ-PSD-014/015 반증의 유일한 관측 지점) |
| N4 (optional) | REQ-PSD-014/015의 `**When**` 절 제거 — 다른 Unwanted 5건과 형식 일치 |
| iter2 §10 잔여 위험 1 | REQ-PSD-013 확장: 실행기 단일 원천이 술어뿐 아니라 **게이트가 대상 트리·저장소에 실행하는 모든 명령**에 걸린다 |
| 형태 스윕(같은 모양 형제) | `AC-PSD-002`/`-003`(0줄 단정) · `AC-PSD-006`(0히트 단정)에 각각 양성 대조 추가. acceptance.md 머리말에 **부재 단정 규율**을 전 항목 공통으로 명문화 |

unchanged: `AC-PSD-009`/`AC-PSD-010`(기존 거절 전제 얽힘 — marker 1 미결 중 재작성 금지), marker 1(열림), marker 2(닫힘), `spec.md:37/39/41` t487 실측 기록.

unchanged_dimensions: 범위 규율(감시 대상 `.claude/settings.json` 단일 유지), 자동 복원 금지, plan.md 코드 인용 13건 — 감사가 통과시킨 세 차원을 건드리지 않았다.

## §E.2 Run-phase Evidence

측정 트리: `.claude/worktrees/t488` · 브랜치 `WT-premerge-drift-assert` · M1 커밋 `63e8e900d`(그 부모 `256b30fa5`).
아래 모든 판정은 일치 줄 수·파일 바이트·기록된 argv 문자열로 내렸다. 종료 코드로 내린 판정은 하나도 없다.

### AC 매트릭스

| AC | 검증 | Actual Output | Status |
|---|---|---|---|
| AC-PSD-001 적중 | `TestSettingsDriftPredicateHit` (F2) | `ok internal/kanban` — 뮤턴트 1에서 `match count: got 0, want 1` 로 뒤집힘 | PASS |
| AC-PSD-002 통과 | `TestSettingsDriftPredicatePass` (F1) | 양성 대조(예상 술어 호출 정확히 1건) 성립 후 `matchCount==0`, raw `""` | PASS |
| AC-PSD-003 경로 오지정 | `TestSettingsDriftPredicateIgnoresOtherFile` (F3) | 대조 성립 후 0. 뮤턴트 3에서 `recorded 0 occurrences of "…-- .claude/settings.json"` 로 뒤집힘 | PASS |
| AC-PSD-004 경로 누락 | `TestSettingsDriftPredicateScopedToWatchedPath` (F4) | 1. 뮤턴트 4에서 `got 2, want 1`, raw `" M .claude/settings.json\n M README.md"` | PASS |
| AC-PSD-005 argv 고정 | `TestSettingsDriftPredicateArgvRecordedAtExecutionBoundary` | 기록된 argv `git --no-optional-locks status --porcelain -- .claude/settings.json`, 플래그 index < `status` index | PASS |
| AC-PSD-006 종료 코드 미의존 | `TestSettingsDriftVerdictNeverReadsAnExitCode` | 스윕 5파일 전부 판독·비어 있지 않음·구현 심볼 확인 후 `ExitCode`/`ExitError`/`$?` 0건. 대조 뮤턴트에서 `../kanban/settings_drift.go:320 reads an exit code` 로 뒤집힘 | PASS |
| AC-PSD-007 (a) 보존 | `TestAssessSettingsDriftPreservesAndLedgers` | 보존 사본이 `<root>/.moai/state/settings-drift/` 아래, 원본과 바이트 일치(원본 비어 있지 않음을 먼저 확인) | PASS |
| AC-PSD-007 (b) 원장 | 같은 테스트 | `ledger.jsonl` 정확히 1줄, `preserved_path`·`sha256`·`size_bytes`·`worktree`·`match_count`·`card` 일치 | PASS |
| AC-PSD-007 (c) 충돌 접미 | `TestAssessSettingsDriftDoesNotOverwriteOnCollision` | 같은 카드·같은 내용 연속 2회 → 보존 파일 2개, 원장 2줄, 두 경로 상이 | PASS |
| AC-PSD-007 (d-1) 양성 대조 | 같은 테스트 | 기록 목록 비어 있지 않고 예상 술어 호출 정확히 1건 | PASS |
| AC-PSD-007 (d-2) 기록 위 부재 | 같은 테스트 | 기록 목록 토큰에 `commit`/`push` 0건 | PASS |
| AC-PSD-007 (d-3) 직접 관측 | 같은 테스트 | 사전 대조: 두 루트 HEAD 모두 40자·서로 상이, `ls-remote` 사전 출력 비어 있지 않고 `refs/heads/main` 포함. 전후 3값 동일. **뮤턴트 6에서 `primary-root HEAD moved: "7fbd3d1b…" -> "0bad8155…"` 로 뒤집힘** — 대상 트리 HEAD는 움직이지 않았으므로 양쪽 루트를 재라는 요구가 실측으로 필요조건이었음이 확인됨 | PASS |
| AC-PSD-007 (d-4) 미스테이징 | 같은 테스트 | `git diff --cached --name-only` 정상 종료(대조: 미실행/퇴화 구별에 그침), 출력에 `settings-drift` 0건 | PASS |
| AC-PSD-008 원본 불변 | `TestAssessSettingsDriftLeavesOriginalUntouched` | 사전 sha256 64자 hex·사전 파일 목록 비어 있지 않음 확인 후 sha256 불변·파일 수 불변, 기록에 트리 수정 명령 0건 | PASS |
| AC-PSD-009 거절이 창을 안 잡음 | `TestAcquireRefusesOnDriftWithGateEnabled` | 대조: 명령이 실제로 실행돼 거절 출력 생성, sha256·보존 경로 포함. `integration-lock.json` 미생성. **뮤턴트 2에서 `acquire succeeded on a drifted tree`(lock 생성) 로 뒤집힘** | PASS |
| AC-PSD-009 사전 lock 변형 | `TestAcquireRefusalLeavesAnExistingLockByteIdentical` | 사전 lock 비어 있지 않고 보유자 이름 포함 확인 후 전후 바이트 동일 | PASS |
| AC-PSD-010 우회 기록 | `TestAcquireBypassIsRecorded` / `TestAcquireForceIsNotASettingsDriftBypass` | lock에 `settings_drift_bypass`·`settings_drift_preserved`, 출력에 보존 경로 + `bypass`. `--force`만 준 실행은 여전히 거절되고 lock 미생성 | PASS |
| AC-PSD-011 실패 ≠ 통과 | `TestPreflightUndeterminedOnPredicateFailure` (F5) | `status == "undetermined"`(적극 단정), `match_count` 키 부재, `error` 비어 있지 않음 | PASS |
| AC-PSD-012 보존 실패 ≠ 통과 | `TestPreflightKeepsDriftVerdictWhenPreservationFails` | `status == "drift"`, `match_count == 1`, `preserve_error` 비어 있지 않음 | PASS |
| AC-PSD-013 (a)-(d) 기본 자세 | `TestAcquireDefaultPostureObservesWithoutRefusing` | workflow.yaml 자체를 두지 않은 기본 설정. 창 기록됨 + 보존 사본 바이트 일치 + 원장 1줄 + 출력에 drift·보존 경로. lock에 우회 표시 없음 | PASS |
| AC-PSD-013 (e) | `TestAcquireDefaultPostureWithAllowFlagIsIdentical` | 같은 기본 설정 + `--allow-settings-drift` → (a)-(d) 동일, lock에 우회 표시 **없음** | PASS |

### 뮤턴트 6종 — 전부 뒤집어 RED를 실측

출력 원본: `.moai/reports/t488/mutants/`.

| # | 뮤턴트 | 포착기 | 관측된 RED |
|---|---|---|---|
| 1 | 술어 삭제(항상 0) | F2 / AC-PSD-001 | `match count: got 0, want 1` (`m1-predicate-deleted.txt`) |
| 2 | 판정 반전 | AC-PSD-009 (게이트) | `acquire succeeded on a drifted tree with the refusal layer on` (`m2-verdict-inverted-gate.txt`). 같은 파일에 F1이 초록으로 남는다는 plan.md M2의 예측도 실측으로 확인 |
| 3 | 경로를 `.claude/settings.local.json`으로 오지정 | F2/F3 + argv | `recorded 0 occurrences` + argv 불일치 (`m3-path-misspecified.txt`) |
| 4 | 경로 인자 누락 | F4 | `got 2, want 1 (2 means the pathspec was dropped)` (`m4-pathspec-dropped.txt`) |
| 5 | `--no-optional-locks` 제거 | 실행 경계 argv | `--no-optional-locks absent from executed argv` (`m5-no-optional-locks-removed.txt`) |
| 6 | 게이트 경로에서 실행기 우회 | AC-PSD-007(d-3) | `primary-root HEAD moved` (`m6-executor-bypassed.txt`). (d-1)/(d-2)는 통과했다 — 기록을 읽지 않는 단정만이 잡았다 |

추가 대조 1건: AC-PSD-006 스윕 자체를 뮤턴트로 뒤집어 RED 확인(`m7-exitcode-sweep-control.txt`) — 0히트가 공허하지 않음을 보였다.

### 픽스처 결함 1건 — 대조가 잡았다

`TestAssessSettingsDriftPreservesAndLedgers`의 두-루트 대조가 초회 전체 패키지 실행에서 발화했다: `control: the two roots resolve to the same HEAD "dc4cf481…"`. 원인은 두 픽스처 저장소의 트리·저자·메시지·타임스탬프가 모두 같아 커밋 SHA가 일치한 것이다. 대조가 옳았다 — 한 값에 대한 두 단정은 한 단정이다. 픽스처를 내용으로 구분(`newFixtureRepoNamed`)해 수리했고, `-count=5` 반복으로 재현 없음을 확인했다.

### 검증 명령과 결과

| 명령 | 결과 |
|---|---|
| `go test ./internal/kanban/... -count=1` | `ok github.com/modu-ai/moai-adk/internal/kanban 136.700s` |
| `go test ./internal/cli/... -count=1` | 전 하위 패키지 `ok` (비-ok 줄 0건) |
| `go test ./internal/template/... -count=1` | `ok` (중립성 감사 포함) |
| `go test ./internal/config/... -count=1` | `TestAlwaysLoadedTokenBudget` 1건 FAIL — 아래 참조. 그 외 전부 `ok` |
| `go vet ./internal/cli/... ./internal/kanban/... ./internal/config/...` | 출력 없음 |
| `golangci-lint run ./internal/cli/... ./internal/kanban/... ./internal/config/...` | `0 issues.` |
| `GOOS=windows GOARCH=amd64 go build ./...` / `GOOS=linux …` | 출력 없음 |
| `GOOS=windows GOARCH=amd64 go vet ./internal/kanban/... ./internal/cli/` | 출력 없음(테스트까지 컴파일됨) |

커버리지: `internal/kanban` 86.3%, `internal/cli` 80.6%(패키지 기존 수준). 신규 파일 단위 — `internal/kanban/settings_drift.go` 83.2%(109/131), `internal/cli/integration_settings_drift.go` 90.8%(69/76). 미커버 잔여는 전부 I/O 실패 분기다.

### 알려진 red 1건 — 본 카드가 만든 것이 아니다

`TestAlwaysLoadedTokenBudget`은 착수 시점 HEAD `256b30fa5`에서 이미 실패하고 있었다: `always-loaded surface = 77723 tokens (budget 77600, headroom -123)`. 이 측정의 표면은 `.claude/rules/moai/**`(paths 무제한) + `CLAUDE.md` + `AGENTS.md` + 출력 스타일이며, 착수 시점의 본 카드 변경은 전부 `internal/` 아래였으므로 이 수치는 HEAD의 값이다(`budget-baseline-before-m6.txt`).

M6이 always-loaded 파일 하나(`kanban-dispatch.md`)에 [HARD] 한 줄을 더해 77723 → **77801**(+78)이 됐다(`budget-after-m6.txt`). 초과분이 123 → 201로 커졌다. 상세 기술은 always-loaded가 아닌 detail companion(`kanban-dispatch-detail.md`, paths 제한 있음)에 넣어 stub 증가분을 최소화했고, stub 문장은 세 차례 압축했다. 예산 상수(`AlwaysLoadedTokenBudget`)는 **건드리지 않았다** — 본 SPEC이 승인한 범위가 아니고, 내가 만들지 않은 123 토큰 초과를 이 카드 안으로 흡수하는 셈이 되기 때문이다. 리드 판단 사항으로 올린다.

### 잔여 위험 — 한 항목은 재측정으로 사라졌다

**개행이 든 경로는 위험이 아니다(실측).** 초기 보고는 "경로에 개행이 들어가면 줄 수 계수가 어긋날 수 있다"를 잔여 위험으로 적었다. 그것은 추론이었고, 재서 보니 틀렸다. `t.TempDir()` 밖 `/tmp` 픽스처에 문자 그대로 개행이 든 파일명을 만들고 프로덕션과 동일한 argv로 잰 결과:

```
$ git --no-optional-locks status --porcelain -- .claude/settings.json
 M .claude/settings.json/a.txt
 M ".claude/settings.json/we\nird.txt"
LINES=2
```

git이 `core.quotePath` 기본값으로 그 이름을 따옴표로 감싸고 개행을 `\n` **두 글자로** 이스케이프한다. 한 경로가 두 줄이 되는 일은 없으며, **줄 수 계수는 경로 내용과 무관하다.**

이 항목은 목록에서 **빠진 것이 아니라 측정으로 해소된 것**이다. 뒤에 읽는 사람이 짧아진 목록과 쓸어 담은 목록을 구별할 수 있어야 하므로 그 구별을 여기 남긴다.

**그 자리에 들어가는 실제 항목 — 유계다.** 위 측정이 다른 사실을 하나 드러냈다. 감시 경로가 파일이 아니라 **디렉터리**가 되면(누군가 `.claude/settings.json`을 디렉터리로 대체하고 커밋한 경우) pathspec이 그 아래 전부를 잡아 줄 수가 1을 넘는다 — 위 출력이 정확히 그 경우로 2줄이다. 판정에는 영향이 없다: 판정식이 `count > 0`이므로 2도 그대로 `drift`로 읽힌다. 정확한 숫자를 단정하는 유일한 곳은 `AC-PSD-004`의 "정확히 1"이고, 그것은 F4 픽스처 안에서 성립을 주장하는 단정이지 프로덕션 불변식이 아니다. 다만 `match_count`를 "수정된 파일 수"로 읽는 소비자가 미래에 생기면 그때부터 유계가 아니다 — **오늘 그런 소비자는 없다.**

**`undetermined`는 거절하지 않는다 — fail-open이다.** git이 실행되지 않는 환경에서 이 게이트는 아무것도 막지 못하고 출력으로만 알린다. 저장소의 다른 네 가드와 같은 자세이며, 운영자가 고른 계열 (가)에서는 거절 층이 애초에 기본 OFF이므로 배포 설정에서의 동작 변화는 아니다. 그러나 그 사실로 이 항목을 눅잡지 않는다 — 조용한 실패가 아니라 **시끄러운 무력화**라는 것이 이 위험의 정확한 성격이다. `AC-PSD-011`이 고정하는 것은 "통과로 접히지 않는다"까지이고, "막는다"는 고정하지 않는다.

**보존 사본은 무한히 쌓인다.** `<primary>/.moai/state/settings-drift/` 아래 untracked로 남고 정리 정책은 이 카드 범위 밖이다. 충돌 접미 규칙이 "덮어쓰지 않음"을 결정적으로 만든 대가이기도 하다 — 덮어쓰지 않는다는 것은 곧 지우지 않는다는 뜻이다. 적중이 드물다는 전제에서는 실무상 문제가 아니지만, 게이트가 오탐을 내기 시작하면 그 전제가 깨진다.

**관측하지 않은 채 남긴 것.** `--allow-settings-drift`와 `--force`를 **동시에** 준 경로는 테스트하지 않았다. 코드상 두 값은 독립 변수이고 서로의 분기를 건드리지 않지만, 그것은 **읽어서 안 것이지 재서 안 것이 아니다.** 증거를 보고한 뒤에 테스트를 덧붙이는 것보다 미검증으로 이름을 남기는 편이 정직한 마감이라 그대로 둔다. doctrine 문구 3곳도 사람이 읽고 판단할 대상이라 기계 검증이 없고, 미러 동일성(`diff -q`)만 확인했다.

### sync-audit 지적 2건 수리 (PASS 0.906 후속)

**F1 [Medium] — 원장의 sha256이 보존 사본에 없는 바이트를 기술할 수 있었다.** 착지본은 파일을 **두 번** 읽었다: `hashSettingsDriftSource(result.Path)`로 해시를 재고, `preserveSettingsDriftCopy`가 `os.ReadFile(source)`로 다시 읽어 복사했다. 그 사이에 누군가 쓰면 원장의 digest와 보존 사본의 바이트가 갈린다.

이론적 경합이 아니다. **그 파일을 예측 불가능하게 쓰는 무언가가 있다는 것이 이 카드의 전제 자체이며**, REQ-PSD-006이 자동 복원을 금지하는 이유가 그것이다. 보존 사본의 지문은 뒤에 세 번째 인스턴스가 나왔을 때 대조할 증거인데, 그 지문이 조용히 틀리면 증거가 아니라 함정이 된다 — 없는 것보다 나쁘다, 신뢰받기 때문이다.

수리: **한 번만 읽는다.** `readAndHashSettingsDriftSource`가 바이트·digest·크기를 한 버퍼에서 함께 돌려주고, `preserveSettingsDriftCopy`는 경로가 아니라 **바이트를 받는다**(경로를 받으면 두 번째 읽기를 허용하는 서명이 된다).

**반증 가능하게 만들었다.** 정적 픽스처는 한 번 읽기와 두 번 읽기를 구별하지 못한다 — 사이에 아무도 쓰지 않으면 두 값이 일치하기 때문이다. 그래서 간섭을 **구성했다**: 같은 패키지의 `integrationLockMutationTestHook` 선례를 그대로 따라, 측정과 쓰기 사이에 nil 기본값 테스트 훅(`settingsDriftPreserveTestHook`)을 두고 테스트가 그 지점에서 원본을 다른 내용으로 덮어쓴다. 한 번 읽기에서는 두 값이 모두 훅 이전 바이트를 기술해 일치하고, 두 번 읽기에서는 갈린다. 그 갈림이 단정이다.

RED 실측(`f1-double-read.txt`) — 두 번 읽기를 복원한 상태:

```
settings_drift_test.go:780: the ledger digest does not describe the preserved bytes:
     ledger:    8d39bb027e9dc2f980612460729cf754be386c725ff0a300baa217faa1859028
     preserved: 92733cfa5e922781efae8739382ca015c9a34dc6b84a6783b8d0180e50eb5652
    the file was read twice, and something wrote to it in between
settings_drift_test.go:784: ledger size_bytes 22 does not match the preserved copy's 38 bytes
```

테스트에는 대조 3개가 앞선다: 두 내용이 실제로 다르다, 훅이 정확히 1회 발화했다, 간섭 쓰기가 실제로 착지했다. 셋 중 하나라도 빠지면 단정이 공허해진다.

**F2 [Low-Med] — 뮤턴트 로그의 줄 번호가 아무 데도 닿지 않았다.** 로그 5건의 앵커가 배포 트리 대비 +14/+16 어긋나 있었다. 원인은 로그를 만든 뒤 테스트 파일이 ~30줄 늘어난 것이다. RED는 진짜였고 앵커는 아니었다 — plan 단계에서 수리한 dangling-SHA 인용과 같은 계열이다: **검증 가능성을 주장하는 기록은 아직 해소되는 것을 인용해야 한다.**

수리: 테스트 파일이 최종 상태가 된 뒤 **여덟 로그 전부를 재생성**했다. 모든 뮤턴트는 소스 파일에 가하고 발화 단정은 테스트 파일에 있으므로, 소스 편집이 테스트 파일 줄을 밀지 않아 앵커가 정확히 해소된다. 아울러 `.moai/reports/t488/mutants/README.md`를 추가해 뮤턴트별 **안정 앵커**(테스트 함수명 + 단정 문구)를 줄 번호와 나란히 적었다 — 뒤에 줄이 밀려도 인용이 살아남는다.

교차 확인: 감사자가 M6을 독립적으로 재-flip해 `settings_drift_test.go:537`에서 발화시켰고, 재생성한 `m6-executor-bypassed.txt`도 **537**로 일치한다.

### sync-audit optional 2건 수리 (F4·F5, 리드 지시로 승격)

**F4 — `preflight`이 잡은 적 없는 창을 거절했다고 말했다.** `errSettingsDriftRefused` sentinel을 `acquire`와 공유한 탓에, 창을 전혀 건드리지 않는 `preflight`이 `release-integration window refused …`를 냈다. **이 카드가 REQ-PSD-009에 [HARD]로 써 넣은 규칙을 자기 출력에서 어긴 것이다** — "우회할 거절이 없는데 우회로 적으면 기록이 거짓말을 한다". 같은 문장이 이 출력을 그대로 정죄한다. 다른 점은 층뿐이다: lock 레코드의 JSON 필드가 아니라 사람이 읽는 문장이다.

수리: sentinel을 둘로 나눴다. `acquire`는 `errSettingsDriftRefused`(창을 실제로 거절한다), `preflight`은 `errSettingsDriftDetected`(판정만 내고 아무것도 가져가지 않는다).

**단정을 두 겹으로 세웠다.** 문언 검사는 거짓말을 잡지만 그 자체가 근사다 — 다르게 쓴 거짓 주장은 빠져나간다. 그래서 **sentinel 동일성**을 앞에 놓았다: 값이 하나면 두 표면을 동시에 만족시킬 어떤 문구도 없으므로, 이것이 위조 불가능한 축이다. 문언 검사는 그 위에 얹는다 — 스크립트가 키로 삼는 것은 sentinel이고, 거짓말을 할 수 있는 것은 문장 쪽이기 때문이다. 양방향으로 단정한다: `preflight`은 거절을 주장하지 않아야 하고, `acquire`는 여전히 주장해야 한다.

**초안 단정이 과했고, 그것을 고쳤다.** 처음에는 "`window`라는 낱말이 없을 것"으로 적었는데, 수리본의 `no integration window was taken`은 **더 정직한 해명**이지 거짓 주장이 아니다. 즉 초안은 명사를 금지하고 있었고 금지해야 할 것은 **주장**이었다. `refused`를 금지하는 형태로 좁혔고, 이 완화가 근사임을 테스트 주석에 적었다.

RED 실측(`f4-preflight-sentinel.txt`) — 공유 sentinel을 복원한 상태, 최종 단정 4개가 전부 발화:

```
integration_settings_drift_report_test.go:174: preflight did not return its own sentinel: release-integration window refused: …
integration_settings_drift_report_test.go:177: preflight returned the acquire refusal sentinel, so its text describes a refusal that never happened: …
integration_settings_drift_report_test.go:186: preflight's error claims a refusal, but it takes no window to refuse: …
integration_settings_drift_report_test.go:189: preflight's error does not name what it actually found: …
```

**F5 — 종료 코드 스윕 목록에 `internal/cli/integration.go`를 추가했다(1줄).** 거절을 호출자에게 전파하는 acquire 배선이 그 파일에 있으므로 게이트의 일부다. 감사가 카드가 손댄 Go 파일 12개 전량 독립 스윕으로 **살아 있는 결함 0**을 이미 확인했으므로 이것은 버그 수리가 아니라 커버리지 간극 메우기다. 추가 후 스윕은 여전히 통과한다(`integration.go`에 `ExitCode`/`ExitError`/`$?` 0건, 측정 확인).

**F6·F7은 optional 유지에 동의한다.** F6(0바이트 dirty 파일에서 `size_bytes` 키가 빠져 JSON이 비대칭)은 소비자가 없어 유계이고, F7(`--allow-settings-drift` + `--force` 동시 경로 무검증)은 §E.2에 이미 미검증으로 이름을 남긴 항목이라 후속 카드가 맞다 — 증거 보고 뒤에 테스트를 덧붙이는 것보다 이름을 남기는 편이 정직하다는 처분이 그대로 유효하다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-06
run_commit_sha: 63e8e900d          # M1; 본 §E.2/§E.3 기록은 그 뒤 M2 커밋
run_status: complete
ac_pass_count: 20                  # AC-PSD-001..013 (007은 a/b/c/d-1..d-4, 009는 2변형, 013은 2변형으로 행 분해)
sync_audit: PASS 0.906             # F1/F2 수리 완료, F4/F5도 리드 지시로 수리; F3은 sync 소관, F6/F7은 optional 유지
last_code_change_after_sync_close: 1a34c51ef, 318d098c4  # §E.4의 "M1 이후 .go 변경 0" 귀속은 무효화됨 — 재측정 결과는 §E.4 참조
ac_fail_count: 0
preserve_list_post_run_count: 0    # t334 워크트리 무수정 — 읽지도 않았다
l44_pre_commit_fetch: n/a          # push 없음, 원격 접촉 없음(레인 규율)
l44_post_push_fetch: n/a           # push 없음
new_warnings_or_lints_introduced: 0 # golangci-lint 0 issues, go vet 무출력
cross_platform_build:
  darwin_arm64: pass               # make build
  linux_amd64: pass                # GOOS=linux go build ./...
  windows_amd64: pass              # GOOS=windows go build ./... + go vet(테스트 컴파일 포함)
total_run_phase_files: 19          # 신규 7 + 수정 12(테스트·문서·미러 포함)
m1_to_mN_commit_strategy: M1 단일 구현 커밋 + M2 증거 기록 커밋
known_preexisting_red: TestAlwaysLoadedTokenBudget (HEAD 256b30fa5에서 이미 -123; M6이 +78)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-07
sync_commit_sha: 955e5190c86c88605efcaaf41bd002005516ed51   # re-close commit. Superseded prior value: 199d2777be085033a96d89cc45466d11ef4a37b9 (invalidated by post-close code landing 1a34c51ef, 318d098c4 — see § The close ran once, code landed after it, and the close is re-run here)
sync_status: complete
b12_self_test_a: "grep -c 'PREMERGE-SETTINGS-DRIFT-001' CHANGELOG.md → 0 (pre-emission; count taken BEFORE this commit's own entry was appended)"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-PREMERGE-SETTINGS-DRIFT-001/acceptance.md | sort -u | wc -l → 13 (AC-PSD-001..013, distinct identifiers); CHANGELOG entry states '13 acceptance criteria' — counts match"
b12_self_test_c: "every file path cited in the CHANGELOG entry verified to exist via ls: internal/cli/integration_settings_drift.go, internal/config/types.go, internal/config/defaults.go — all present"
changelog_entry_position: "CHANGELOG.md, [Unreleased] > ### Added, first entry (immediately after the ### Added heading, before the SPEC-BINLAG-KEYGUARD-001 entry)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (status: + updated: only; body untouched)"
  plan_md: "no status: field (stateless per spec-frontmatter-schema.md § Artifact Statelessness) — no transition"
  acceptance_md: "no status: field (stateless) — no transition"
  progress_md: "no status: field (progress.md records phase progress in body sections, not frontmatter) — this §E.4 section is the sync-phase signal"
canary_compliance_check:
  mirror_kanban_dispatch_md: "diff -q .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md → exit 0 (identical)"
  mirror_kanban_dispatch_detail_md: "diff -q .claude/rules/moai/workflow/kanban-dispatch-detail.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md → exit 0 (identical)"
  claude_local_md_local_only: "CLAUDE.local.md is documented local-only (no template mirror expected) — the M1 addition there (acquire pre-merge assertion doctrine) carries no counterpart obligation"
docs_site_reach_decision: "NOT extended — see § Documentation reach below"
```

### Documentation reach decision

`docs-site` (tracked at `docs-site/`) carries **zero** existing pages mentioning `moai integration` — none of its three pre-existing verbs (`status` / `acquire` / `release`) are documented there, confirmed by `grep -rl "moai integration" docs-site/content` returning no matches. `moai integration` is a lane/maintainer-internal surface (the pre-merge integration-window protocol used by kanban/factory lanes and the release lead), not an end-user-facing command — end users of `moai init`-scaffolded projects never invoke it. The new `preflight` verb and the `workflow.settings_drift_gate` config block extend that same internal surface and are documented at the maintainer layer instead: `CLAUDE.local.md` §4.1 (the `[HARD]` acquire-precondition paragraph added in this same M1 commit) and the code-level doctrine comment at the top of `internal/cli/integration_settings_drift.go`. No docs-site page is added for this card, consistent with the existing (undocumented) sibling verbs.

### Template-First check

`63e8e900d` (M1) touched two doctrine files under `.claude/rules/moai/workflow/`: `kanban-dispatch.md` (+2 lines, the acquire-precondition mention) and `kanban-dispatch-detail.md` (the fuller mechanism description). Both carry a mirror under `internal/template/templates/.claude/rules/moai/workflow/`, and both mirrors were edited in the same M1 commit (confirmed by `git diff --stat 256b30fa5..63e8e900d`, which lists all four paths, and by the `diff -q` byte-identity check in `canary_compliance_check`). `CLAUDE.local.md` is deliberately local-only per `CLAUDE.local.md` §2 (Local-Only Files list) and carries no template counterpart — its M1 addition needs none. No embedded-template content changed beyond the two rule-file mirrors, so `make build` was not required for this sync commit (the `.claude/rules/` tree is deployed as source files under `internal/template/templates/`, not regenerated by `make build`; `make build` only recompiles the Go binary with the `//go:embed all:templates` directive picking up the current template-source bytes at compile time — verified: `git diff --stat` shows no `internal/template/embed.go` or catalog change, and no new template file was added that would need a fresh embed).

### The close ran once, code landed after it, and the close is re-run here

The first sync close (`199d2777b`) and its two follow-on accuracy commits (`5832f1f51`, `65146f0f6`) were dispatched at the same time the lead dispatched sync-audit findings F1/F2 to a second, code-writing agent working the same tree. Both agents were write-capable and both landed commits: `1a34c51ef` (F1 — read-once fix + mutant-log anchor regeneration) and `318d098c4` (F4 — `preflight` sentinel split + F5 — exit-code sweep list addition) touch five `.go` files, all after the first close had already run. The §E.4 attribution this section originally carried — "zero `.go` files changed since M1" — was measured against the tree as it stood before those two commits landed, and is false against the tree as it now stands. It is corrected here rather than quietly dropped.

**Cause, stated as a fact rather than as blame directed anywhere**: the repository's concurrency rule (`agent-common-protocol.md` § Background Agent Execution — never run two write-capable agents concurrently in one tree) was not observed for this dispatch; the lead has named this in its own message as the dispatch that caused it. **Structural fix**: the sync close goes last — a card's code-fixing work (including sync-audit remediation) completes and lands before the sync close reads the tree to attribute evidence against it, not concurrently with it.

**Re-measured attribution, against current HEAD `318d098c4`.** Rather than re-assert a diff-stat consumption that no longer holds, the affected suites were re-run directly, by this agent, in this sync-phase turn:

```
$ go build ./...
(exit 0)

$ go test ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	139.942s

$ go test ./internal/cli/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	389.135s
ok  	github.com/modu-ai/moai-adk/internal/cli/agentlint	0.591s
ok  	github.com/modu-ai/moai-adk/internal/cli/harness	10.531s
ok  	github.com/modu-ai/moai-adk/internal/cli/pr	2.966s
ok  	github.com/modu-ai/moai-adk/internal/cli/preference	3.037s
ok  	github.com/modu-ai/moai-adk/internal/cli/printer	2.056s
ok  	github.com/modu-ai/moai-adk/internal/cli/specid	2.558s
ok  	github.com/modu-ai/moai-adk/internal/cli/taskledger	4.657s
ok  	github.com/modu-ai/moai-adk/internal/cli/uikit	3.400s
ok  	github.com/modu-ai/moai-adk/internal/cli/update	5.044s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	6.493s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/deploy	1.072s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	5.539s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/plan	5.882s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/report	7.137s
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	9.641s
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	10.807s

$ go test ./internal/config/... -count=1
--- FAIL: TestAlwaysLoadedTokenBudget (0.01s)
    always-loaded surface = 77801 tokens (budget 77600, headroom -201, 17 entries)
FAIL	github.com/modu-ai/moai-adk/internal/config	2.649s
ok  	github.com/modu-ai/moai-adk/internal/config/atomicfile	0.840s
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.520s

$ go vet ./internal/cli/... ./internal/kanban/... ./internal/config/...
(no output — exit 0)

$ golangci-lint run --timeout=2m ./internal/cli/... ./internal/kanban/... ./internal/config/...
0 issues.
```

`internal/kanban` and `internal/cli` (all sub-packages) are green — the F1 read-once fix and the F4 sentinel split introduced no regression. `internal/config`'s single failure is `TestAlwaysLoadedTokenBudget`, and its numbers — surface 77801, budget 77600, overflow 201 — are byte-for-byte the same as the pre-existing red already carried in §E.2 and §E.3 (`known_preexisting_red`); this run does not add a new failure, it reproduces the one already on record. This is a genuine re-execution, not the consumed-evidence path this section originally claimed — the attribution above states plainly that it was re-run, and by whom.

**Self-sufficiency evidence — no further code landed after this measurement, either.** This test run was taken against `318d098c4`. Two commits followed it before this record was written (the auditor's verdict commit and its delta-confirmation addendum), and both are confirmed docs-only rather than merely asserted:

```
$ git diff --name-only 318d098c4..78b3211bb
.moai/reports/t488/sync-audit.md
.moai/specs/SPEC-PREMERGE-SETTINGS-DRIFT-001/progress.md
```

Zero `.go` files. Pinned to `78b3211bb` (the tree this section is being written against) rather than to `HEAD` — a moving anchor here would let this very claim go stale the next time a commit lands, which is the exact defect class this re-close exists to correct.

### Carried-forward items (not resolved by sync — recorded, not silently dropped)

- **`TestAlwaysLoadedTokenBudget`** — pre-existing red (surface 77801 / budget 77600, overflow 201), NOT this card's origin (123 of 201 inherited at base commit `256b30fa5`; this card's `kanban-dispatch.md` addition contributed 78). The budget constant `AlwaysLoadedTokenBudget` in `internal/config` is left untouched, per the run-phase decision recorded in §E.2 — raising it would silently absorb overflow this card did not create. Left for the lead / a follow-up card to resolve.
- **Five residual risks** recorded in §E.2 (undetermined is fail-open by design; preserved copies accumulate unboundedly with no cleanup policy in scope; `--allow-settings-drift` + `--force` combined is untested; directory-shaped watched path is bounded but the `match_count` semantic contract is undocumented for future consumers; three doctrine-text locations rely on human review, not mechanical verification) — none resolved at sync; all carried forward as-is.
