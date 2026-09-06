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

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
