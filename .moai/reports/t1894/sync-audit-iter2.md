auditor-model: claude-opus-5-5
verdict: PASS-WITH-DEBT
overall_score: 0.89
must_pass_failed: 0
blocking_count: 0
audited_sha: 0ca3d59b2e85a801b456a60e3972f91f08ebbadf
required_backend: codex pass
convergence_overall: pass

# 동기화 단계 델타 재감사 (2회차) — SPEC-CARD-MODE-JUDGMENT-001 (카드 t1894, Tier L, 문서 전용)

감사 대상은 `WT-fail-safe-serial` 브랜치 HEAD `0ca3d59b2e85a801b456a60e3972f91f08ebbadf`이며 작업 트리는 깨끗했다. 델타 범위는 1회차 감사 SHA `f7cc3f8ee8`부터 HEAD까지의 커밋 여덟 개(`88bc5bcfb4` 1회차 판정 파일 포함)다. 호출자의 설명은 검증할 주장 목록으로만 다뤘고, 판정은 이번 실행에서 이 트리를 다시 측정한 결과에만 근거한다. Tier L 통과 기준선은 0.85다.

## 1. 판정 요약

판정은 PASS-WITH-DEBT, 점수는 0.89다.

- 1회차 차단 결함 S1은 해소됐다. D의 전환 경계 줄(139행)이 새 빈 레인 정의와 같은 조건("실행을 멈춘 때")을 쓰고, 다시 들어오는 경로(`moai factory next --card <id>` + `EnterWorktree`)도 M의 임대 갱신 절과 맞는다. 계획 산출물 개정(`0fca8e0e90`)이 실행 정정(`45270e3ef3`)보다 커밋 그래프에서 앞선다.
- S2는 해소됐다. AC-MJ-020(b)가 닫는 트리에서 초록이고, 허용은 정확한 경로 하나뿐임을 대조 셀 세 개로 직접 재현했다.
- S3는 정정됐다. 핵심 사실은 디스크와 맞지만 새로 들어간 문구 두 곳이 부정확하다(N2, 경미).
- S4는 미리 공개된 운영자 보류 잔여(R-9)로 정직하게 남겨졌고, S5는 경미 항목으로 수용됐다.
- 필수 백엔드 codex는 델타에 대해 pass(지적 0건)를 냈다. 영수증 `rcpt-a4b24c097a712091ae9c1d06`.
- 새 결함은 비차단 세 건이다. 그중 N1(SHOULD-FIX)은 S1과 같은 계열이다. `factory-dispatch.md` 177행 [HARD] "stays while it carries the card"를 같은 문서 20행의 정의("A lane carries its card from dispatch to landing")로 읽으면, merge-ready 카드를 쥔 레인이 그 트리를 떠나라는 D:139와 맞지 않는다. 다른 독해("carries" = 실제 작업 중)로는 충돌이 없고, 라우팅 조건이나 `done`을 직접 말하지도 않아 차단으로 올리지 않았다. 부채로 남긴다.

부채(PASS-WITH-DEBT의 근거): N1, N2, N3, 그리고 운영자 판정을 기다리는 R-9(S4).

## 2. 4차원 평가

| 차원 | 점수(0..1) | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 0.88 | PASS (must-pass) | 원장 65개 실행이 모두 초록 경로 값과 일치(E-005e `0 1`/`0 1`, E-020a·E-020i 열여섯 행 전부 `0 0 0 0 0`, E-020b `unexpected-paths 0`). S1 해소. 남은 것은 해석에 따라 갈리는 N1 |
| Security (25%) | 0.95 | PASS (must-pass) | 문서·점검 스크립트만 변경, Go 소스 0개. 델타 추가 행 비밀정보 탐침 0건, 템플릿 추가 행 카드·SPEC id·날짜·긴 hex 탐침 0건. 위임 승격 예외는 E-006 `8 8 1 1`로 문언 그대로 |
| Craft (20%) | 0.88 | PASS | 이름 지정 Go 가드 14+6개 통과, 빈 스윕 0건, `pairs` 다섯 커밋 모두 기대 쌍 수. 점검 스크립트 변경은 한 줄이고 docstring이 정직함. 감점: CHANGELOG 새 문구 부정확(N2), progress.md §E.4 기록 일부의 순서 오기와 이월된 자기 검증 문구(N3) |
| Consistency (15%) | 0.85 | PASS | 일곱 미러 쌍 `cmp` 일치, D 쌍 바이트 동일, `factory-dispatch.md` 포크는 기준 커밋과 같은 한 줄, `git diff --check` 깨끗. 감점: N1, R-9(S4) 미해결 |

가중 합은 0.4×0.88 + 0.25×0.95 + 0.2×0.88 + 0.15×0.85 = 0.893이고, 소수 둘째 자리로 0.89다. must-pass 차원(Functionality, Security)은 둘 다 기준을 넘는다.

## 3. 1회차 결함 처분

| 결함 | 1회차 등급 | 처분 | 재현 근거 |
|---|---|---|---|
| S1 빈 레인 정의 대 전환 경계 | BLOCKING | 해소 | D:139가 "once it stops executing the card — the card reached `done`, or is merge-ready, held (operator-wait), or serial-waiting — and before the next card is routed to it"로 바뀌었고 두 쌍이 바이트 동일(`cmp` exit 0). E-005e `0 1`/`0 1` exit 0. F:158·F:164·F:168·K:58·C:52-56은 `done` 때 일어나는 일만 말하고 라우팅 조건을 걸지 않는다. M:30의 실행 정의(대기 중이면 실행 아님)와 D:139의 상태 목록이 맞는다. 남은 의역 후보는 F:177(N1) |
| S2 AC-MJ-020(b) 빨강 | SHOULD-FIX | 해소 | E-020b `unexpected-paths 0` exit 0, E-020j `unexpected-paths 0` exit 0, E-020k `CHANGELOG.md.bak` exit 1, E-020l `docs/CHANGELOG.md` exit 1, E-020c 199 exit 1. 코드 변경은 `allowed = set(also) \| {'CHANGELOG.md'}` 한 줄이며 set 멤버십이라 접두사나 기본 이름 일치가 아니다. docstring은 "exact path"와 함께 내용은 보지 않는다는 한계까지 적었다. E-020d `same same`이 새 해시를 고정 |
| S3 CHANGELOG 과장 | SHOULD-FIX | 정정(문구 경미 결함 N2 남음) | `Mode judgment` 개수: core `:0`, factory-dispatch.md `:1`, gtd.md `:2`. 이제 항목은 core에 포인터가 없다고 적는다. core 차이는 재조정 절과 크론 범위 꼬리 두 줄뿐. REQ 17, AC-MJ 20, 실행 커밋 네 개 `pairs` 3/5/4/1. CHANGELOG에 SPEC id는 한 줄(`grep -c` → 1) |
| S4 폴링 크론 문구 대 A §14 | SHOULD-FIX | R-9 잔여로 유지(정직) | spec.md R-9가 원래 문구 그대로 남아 있고 §E.4 `s4`가 같은 판단을 기록. 문서 텍스트는 바뀌지 않았다. 운영자 판정 전까지 부채 |
| S5 실행 정의 제외 목록 | MINOR | 수용 | §E.4 `s5: "minor, accepted"`. 1회차에서 선택 수정이었으므로 수용은 정당하다 |

## 4. 새 결함 (델타 한정)

### N1 [SHOULD-FIX] [optional] F:177 "stays while it carries the card"와 D:139의 긴장 — 신뢰도 중간

- 위치: `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md` 177행(배포 쌍 동일, 기준 커밋 173행에 이미 있던 줄)과 같은 파일 20행, D 139행.
- 내용: F:20(및 K:12)은 "A lane carries its card from dispatch to landing"이라고 정의한다. F:177 [HARD]는 "A card session starts in its worktree (`moai cc -w`) and stays while it carries the card"라고 한다. merge-ready 카드는 아직 착지 전이므로 이 정의로는 레인이 계속 "carries" 상태인데, D:139는 merge-ready에서 그 트리를 떠나라고 하고, 빈 레인 정의는 그 레인에 다른 카드를 보내라고 한다. 한 세션이 두 트리에 동시에 머물 수는 없다.
- 차단이 아닌 이유: F:177은 `done`이나 라우팅 조건을 말하지 않으며, "carries"를 "실제로 작업 중"으로 읽으면 D:139와 맞는다. REQ-MJ-006의 새 절은 "`done` 도달"을 기다리게 하는 문장을 금하는데, F:177은 그 말을 쓰지 않는다. codex(필수)도 같은 델타에서 지적하지 않았다. design.md §D-3 S-행은 F:158·F:164·F:168·K:58만 검토했고 F:177은 언급하지 않는다.
- 필수 수정(선택, 후속 카드 권장): manager-spec이 매니페스트를 개정해 F:177의 "while it carries the card"를 "while it executes the card"로 바꾸거나, D:139처럼 merge-ready·held·serial-waiting 이탈을 예외로 명시한다. 같은 판단을 K:62(always-loaded, "stays in its worktree")에도 적용할지 함께 정한다.

### N2 [MINOR] [optional] CHANGELOG 새 문구 두 곳이 부정확 — 신뢰도 높음

- 위치: `CHANGELOG.md` [Unreleased] > Added 첫 항목(정정 커밋 `6e74de5f4e`가 바꾼 문장).
- 내용 (a): "within the limits stated once in `factory-dispatch-detail.md` § Delegated-dispatch promotion, where the transition boundary follows the free-lane definition" — 전환 경계 줄은 § Factory in-lane 3-stage(D:139)에 있고 § Delegated-dispatch promotion(D:142 이하)에는 없다.
- 내용 (b): "Three mid-run SPEC amendments … (follow-up commits `45270e3ef3`, `80a7dc041b`)" — 세 개정은 `8b9f610b91`, `0fca8e0e90`, `80a7dc041b`다. `45270e3ef3`은 개정이 아니라 실행 정정이며 `0fca8e0e90`은 빠졌다. 또 spec.md 이력은 0.1.8·0.1.9를 "Sync-audit amendment"라 부르는데 CHANGELOG는 셋 모두 "during the run"이라고 적는다.
- 필수 수정(선택): 문장을 "the transition boundary in § Factory in-lane 3-stage follows the free-lane definition"으로 옮기고, 커밋 목록을 개정 세 개와 정정 하나로 나눠 적는다.

### N3 [MINOR] [optional] progress.md §E.4 기록의 순서 오기와 이월 문구 — 신뢰도 높음

- 위치: `progress.md` §E.4.
- 내용 (a): `s1: "fixed by the run phase in 45270e3ef3 …, after SPEC amendments 8b9f610b91, 0fca8e0e90, 80a7dc041b"` — `git log`는 `45270e3ef3`이 `80a7dc041b`보다 먼저임을 보인다(`0fca8e0e90` → `45270e3ef3` → `120f9eeb6e` → `80a7dc041b`). S1 정정은 `80a7dc041b` 뒤가 아니다.
- 내용 (b): `b12_self_test_a/b/c` 세 줄은 첫 동기화 커밋의 것이 그대로 남았다(정정 커밋 diff에 변경 없음). `_a`의 "printed 0 before the entry was written"은 정정 시점에는 1이었을 것이고, `_c`의 "(… verdict.md)"는 항목이 `verdict.md`를 명시하지 않는다. 정정 동기화가 새로 넣은 주장(커밋 sha 두 개, core 꼬리)에 대한 재검증이라고는 적혀 있지 않다. 이번 감사는 그 주장을 직접 확인했다(§3 S3 행).
- 필수 수정(선택): 다음 기록 갱신 때 `s1`의 "after"를 "after amendment 0fca8e0e90; S2 amendment 80a7dc041b followed"로 고치고, b12 줄에 정정 시점 재검증 결과를 덧붙인다.

### 확인했지만 결함이 아닌 것

- 위임 승격 예외: 델타가 D § Delegated-dispatch promotion을 건드리지 않았다(E-006 `8 8 1 1`/`8 8 1 1` exit 0). d314 제한 (a)-(g) 대조는 1회차 §4와 같다. D:139는 merge-ready·held·serial-waiting 카드를 쥔 레인을 "실행 카드 없음"으로 다루므로 d314 (b) "live lane on which no card is actually executing"과 같은 범위이며 넓히지 않는다.
- 누출: 템플릿·배포 추가 행에서 카드 id, `SPEC-` 접두, 날짜, 10자 이상 hex 탐침 0건. E-013a 여덟 파일 모두 `:0`.
- 회귀: E-020a 열여섯 행 `0 0 0 0 0`, 일곱 미러 쌍 `cmp` exit 0, 포크 차이 줄의 md5가 기준 커밋과 HEAD에서 같음.
- 커밋 순서: 개정 `0fca8e0e90`(manifest D 행 추가)이 정정 `45270e3ef3`보다 먼저다. 계획이 먼저 고정되고 실행이 따랐음을 커밋 그래프가 보인다.

## 5. 증거 기반 보고

### Claim

| # | 주장 | 결과 |
|---|---|---|
| C1 | 원장 59블록(65회 실행) 전부 초록 경로 또는 알려진 실패 입력 값과 일치 | 확인 |
| C2 | 새 셀 E-005e·E-020i·E-020j/k/l과 재고정된 E-020b/c/d/e가 기록대로 재현 | 확인 |
| C3 | `pairs` 정정 커밋 `45270e3ef3:1`과 마일스톤 네 커밋 | 확인 |
| C4 | AC-MJ-014 Go 가드 14+6개가 이 트리에서 통과, 빈 스윕 없음 | 확인 |
| C5 | S1-S5 처분이 디스크와 일치 | 확인(기록 오기 N3) |
| C6 | 델타가 새 모순·과장·예외 확대·누출·회귀를 넣지 않음 | 부분 반증 — N1(이전부터 있던 줄과의 긴장), N2, N3. 예외 확대·누출·회귀는 없음 |
| C7 | 필수 백엔드 codex 통과 | 확인 |

### Evidence

원장 재실행은 `acceptance.md` §2의 `command:`/`commands:` 셀을 수정 없이 `/bin/sh -c`로 돌렸다(PATH 앞에 `/usr/bin`, cwd 워크트리 루트, 스크립트 `run_ledger.py`). 결과 `59 blocks 65 runs`. 판정에 쓴 줄:

```text
E-001  exit=0 → 8 / 8
E-002a exit=0 → (빈 출력) ; E-002d exit=0 → gtd.md:1 ×2 ; E-003 exit=0 → :1 ×2
E-004a exit=0 → :1 ×2 ; E-004b exit=1 → :0 ×2 ; E-004c exit=0 → :2 ×2 ; E-004d exit=0 → :1 ×2 ; E-004e exit=0 → :2 ×2
E-005a exit=0 → :1 ×2 ; E-005b exit=1 → :0 ×2 ; E-005c exit=0 → :1 ×2 ; E-005d exit=0 → :3 ×2
E-005e exit=0 → 0 1 / 0 1
E-006  exit=0 → 8 8 1 1 / 8 8 1 1 ; E-007 exit=0 → 4 4 1 / 4 4 1
E-008a exit=0 → (빈 출력) ; E-008b exit=0 → core:1 F:2 A:2 W:1 (양쪽) ; E-008c..f exit=0
E-009  exit=0 → 1 1 0 / 1 1 0 ; E-010a exit=1 (빈 출력) ; E-010b exit=0 → :1 ×2 ; E-010c exit=0 → 1
E-011#1..#7 exit=0 (출력 없음) ; E-011f exit=1 → differ: char 23761, line 226
E-011g exit=1 → 9579676013 pairs 0 unpaired 7 ; E-011h exit=0 → 1969709ce6 pairs 2 unpaired 0 expected 2
E-012a exit=0 → OK 9997 ; E-012b exit=1 → OVER 26434
E-013a exit=1 → 여덟 파일 :0 ; E-013b exit=0 → 8 ; E-013c exit=0 → A:2 D:2 G:1
E-016  exit=0 → cards, detail, mechanics ; E-017 exit=1 → 1 file changed, 1 insertion(+), 1 deletion(-)
E-018a exit=0 → :1 ×2 ; E-018b exit=1 → :0 ×4 ; E-018c exit=0 → :1 ×4 ; E-018d/e exit=0 → :1 ×2
E-019a exit=0 → :1 ×2 ; E-019b exit=1 → 0 ; E-019c exit=0 → :1 ×2
E-020a exit=0 → C K F D M A G W 모두 "template 0 0 0 0 0 deployed 0 0 0 0 0"
E-020b exit=0 → unexpected-paths 0
E-020c exit=1 → unexpected-paths 199 (첫 여덟 경로 기록과 동일)
E-020d exit=0 → same same ; E-020e exit=1 → DIFFERENT same
E-020f exit=1 → unexpected-paths 1 / .moai/policies/workflow/factory-dispatch-detail.md
E-020g exit=1 → unexpected-paths 1 / .moai/workflows/gtd.md
E-020h exit=1 → C good 0 0 pair-move 44 44
E-020i exit=0 → 열여섯 행 모두 0 0 0 0 0
E-020j exit=0 → unexpected-paths 0
E-020k exit=1 → unexpected-paths 1 / CHANGELOG.md.bak
E-020l exit=1 → unexpected-paths 1 / docs/CHANGELOG.md
```

원장 값과 다른 칸은 없다. 원장에 적힌 stdout은 RED-now 값이고, 위 출력은 각 기준의 초록 경로 문장과 progress.md §E.2 표가 말하는 값과 같다. E-020i는 정정 커밋 뒤 exit 0으로 뒤집혔다고 원장이 예고한 그대로다.

커밋 그래프 쌍:

```text
$ python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py pairs .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt 45270e3ef3:1 2fb0be5ed8:3 89eb51de2f:5 770b5ae654:4 05bcdc3a8f:1
45270e3ef3 pairs 1 unpaired 0 expected 1
2fb0be5ed8 pairs 3 unpaired 0 expected 3
89eb51de2f pairs 5 unpaired 0 expected 5
770b5ae654 pairs 4 unpaired 0 expected 4
05bcdc3a8f pairs 1 unpaired 0 expected 1
exit=0
```

선언된 포크와 미러:

```text
$ git diff 28c7bb8d36:<template F> 28c7bb8d36:<deployed F> | grep -E '^[-+][^-+]' | md5 → 6522b0a6755c3bc0af228fe59c14bdf8
$ git diff HEAD:<template F> HEAD:<deployed F> | grep -E '^[-+][^-+]' | md5        → 6522b0a6755c3bc0af228fe59c14bdf8
  (-/+ 한 쌍: "[HARD] A lane whose card has passed verification does not wait for the leader to integrate it: …")
$ cmp <template D> <deployed D> → exit 0
$ git diff --check 28c7bb8d36 HEAD → 출력 없음, exit 0
$ git diff --name-only 28c7bb8d36 HEAD | grep -c '\.go$' → 0
```

S1 재현(정정 뒤):

```text
$ awk 'NR>=133 && NR<=139' <template D>
133: - **Routing.** The leader routes a card WHOLE to a free lane. …
134:   - A free lane is a lane on which no card is actually executing; a lane that holds only merge-ready, held (operator-wait), or serial-waiting cards is a free lane.
139: - **The transition boundary is between cards.** A factory lane leaves its card worktree (`ExitWorktree(keep)`) once it stops executing the card — the card reached `done`, or is merge-ready, held (operator-wait), or serial-waiting — and before the next card is routed to it, …; a kept tree is re-entered when its card resumes, by `moai factory next --card <id>` and `EnterWorktree` of the leased tree (`factory-dispatch-mechanics.md` § Lease renewal — only the card being executed).
$ M:30  Executing means the lane is carrying out a stage of the card — plan, run, sync, the card-review, or the landing sequence — and is not waiting on a slot, a trigger, or the operator.
$ M:34  … at the landing slot the leader reassigns the card or the lane re-leases it with `factory next --card`, the stage preserved.
```

N1 재현:

```text
$ /usr/bin/grep -rn -E 'carr(y|ies) (its|the) card|while it carries' internal/template/templates/.claude/rules/moai/workflow/
factory-dispatch.md:20:A lane carries its card from dispatch to landing autonomously: …
factory-dispatch-core.md:12:A lane carries its card from dispatch to landing autonomously: …
factory-dispatch.md:177:[HARD] A card session starts in its worktree (`moai cc -w`) and stays while it carries the card; between cards it moves by the in-session transition (`ExitWorktree(keep)`, lease, `EnterWorktree` of the leased tree); …
$ git show 28c7bb8d36:<template F> | /usr/bin/grep -n 'stays while it carries' → 173: (기존 줄)
```

N2·S3 재현:

```text
$ /usr/bin/grep -c 'Mode judgment' <template K> <template F> <template G> → core:0 F:1 gtd:2
$ git diff 28c7bb8d36 HEAD -- <template K> 의 -/+ 줄 → 재조정 절 첫 구("The one reconciliation" → "A reconciliation") 한 쌍, 크론 범위 꼬리 한 쌍
$ grep -o REQ-MJ-[0-9]+ spec.md | sort -u | wc -l → 17 ; AC-MJ (acceptance.md) → 20 ; grep -c SPEC-CARD-MODE-JUDGMENT-001 CHANGELOG.md → 1
$ git log --oneline f7cc3f8ee8..HEAD → 0ca3d59b2e, 6e74de5f4e, ff41cde331, 80a7dc041b, 120f9eeb6e, 45270e3ef3, 0fca8e0e90, 88bc5bcfb4
  (D 전환 줄은 D:139, § Factory in-lane 3-stage 안; § Delegated-dispatch promotion은 D:142부터)
```

N3 재현: `git show 6e74de5f4e -- progress.md`의 +/- 줄은 `sync_commit_sha`, `sync_status`, `corrective_sync` 다섯 줄뿐이고 `b12_self_test_*`는 바뀌지 않았다. 커밋 순서는 위 `git log` 출력.

보안·누출 탐침:

```text
$ git diff f7cc3f8ee8 HEAD -- internal .claude CHANGELOG.md .moai/specs | grep '^+' | /usr/bin/grep -c -i -E 'api[_-]?key|secret|password|token=|ghp_|sk-[A-Za-z0-9]{8}' → 0
$ git diff f7cc3f8ee8 HEAD -- internal .claude | grep '^+[^+]' | /usr/bin/grep -c -E '(카드 id|SPEC-[A-Z]|10자 이상 hex|YYYY-MM-DD 패턴)' → 0
```

AC-MJ-014 Go 가드(직전 `uptime` 1분 부하 6.14로 32 이하; `moai slot acquire --resource go-heavy --max-duration 15m` → `slot go-heavy acquired by e88b03dd-… until 2026-10-22T20:20:20Z`):

```text
$ go test ./internal/template -count=1 -v -run '^(TestBindingLedgerIntegrity|TestRoleCoreStubGate|TestDeployedAlwaysLoadedCharBudget|TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks|TestReviewOwnership_CardReviewDoctrine|TestFactoryDispatchBodyCarriesRenewIntervalLinkage|TestLaneRecheckDoctrine|TestLaneWatchdogRenewsFirst|TestLaneWatchdogRecordsRefusalAndMeasures|TestWorkflowRulePathsPinned|TestTemplateNoInternalContentLeak|TestCardIDBaselineHasNoStaleEntries|TestCatalogHashCoversSkillSubfiles)$'   exit=0
--- PASS: TestDeployedAlwaysLoadedCharBudget (0.10s)
--- PASS: TestBindingLedgerIntegrity (5.85s)
--- PASS: TestTemplateNoInternalContentLeak (0.69s)
--- PASS: TestLaneRecheckDoctrine (0.00s)
--- PASS: TestLaneWatchdogRenewsFirst (0.00s)
--- PASS: TestLaneWatchdogRecordsRefusalAndMeasures (0.00s)
--- PASS: TestFactoryDispatchBodyCarriesRenewIntervalLinkage (0.00s)
--- PASS: TestRoleCoreStubGate (0.00s)
--- PASS: TestReviewOwnership_CardReviewDoctrine (0.00s)
--- PASS: TestDeclaredRuleMirrorForks (0.00s)
--- PASS: TestWorkflowRulePathsPinned (0.00s)
--- PASS: TestRuleTemplateMirrorDrift (0.00s)
--- PASS: TestCatalogHashCoversSkillSubfiles (0.01s)
--- PASS: TestCardIDBaselineHasNoStaleEntries (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/template	7.141s
    binding_ledger_test.go:1084: [AC-ALB-021(2)] fragment sweep: binding/normative qualifying lines=560 on-demand rule files=100 .claude skill files=248 .agents skill files=17 violations=0 …

$ go test ./internal/cli -count=1 -v -run '^(TestAutoPickMirrorParity|TestAutoPickDocDoctrine|TestTodoSkillDocumentsClassification|TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestAutoRankMarkerDisclosure)$'   exit=0
--- PASS: TestAutoRankDoctrineAmendment (0.00s)
--- PASS: TestAutoRankMirrorParity (0.00s)
--- PASS: TestAutoRankMarkerDisclosure (0.00s)
--- PASS: TestAutoPickDocDoctrine (0.01s)
--- PASS: TestAutoPickMirrorParity (0.00s)
--- PASS: TestTodoSkillDocumentsClassification (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.232s

빈 스윕: grep -c -E 'no tests to run|no test files' → go-template.log:0, go-cli.log:0
$ moai slot release --resource go-heavy → slot go-heavy released (was e88b03dd-…) ; moai slot status → slot go-heavy: free
```

교차 모델 감사(`mcp__moai__audit_multi`, project_root=이 워크트리, target=baseBranch, delta_base=`f7cc3f8ee8c65061382a29c40cf7cc809c1cb7b2`, gates 인자 없음). 감사 계획(`moai verify audit-plan --project-root <toplevel>`)은 `config_status ok`, model codex, gate codex required / claude off / glm off, `cross_model_active true`. 결과 다이제스트:

| 항목 | 값 |
|---|---|
| overall_verdict | pass |
| gate_unmet | 결과에 필드 없음 |
| plan_source | config |
| per_backend: codex | gate required, 판정 pass, 지적 0건 |
| audit_receipt | rcpt-a4b24c097a712091ae9c1d06 |
| build_commit | b1806498fe |

codex가 보고한 지적이 없으므로 재현할 대상도 없다. 이 판정은 오케스트레이터가 결과를 `<toplevel>/.moai/state/audit-plan-result.json`에 쓰고 `moai verify audit-plan --project-root <toplevel> --result-file <그 경로>` 확인을 마칠 때까지 최종이 아니다.

### Baseline-attribution

- 트리: HEAD `0ca3d59b2e85a801b456a60e3972f91f08ebbadf`, 작업 트리 깨끗, 델타 기준 `f7cc3f8ee8`, 문서 기준 `28c7bb8d36`. 모든 측정은 이번 실행에서 이 트리를 대상으로 했다.
- Go: `go version go1.26.0 darwin/arm64`가 이 트리 소스에서 테스트 바이너리를 컴파일했다. 판정 빌드는 곧 이 트리다.
- 원장 셀: `/usr/bin/grep`(BSD), `awk`, `cmp`, `git`, 트리 안의 `closed-world-check.py`(`python3 -I`).
- 설치된 `moai`: v3.2.0-rc.38, 커밋 `b1806498fe`(`moai_cp/20261007_195927-2076-gb1806498fe`). `git merge-base --is-ancestor b1806498fe HEAD` → exit 1로 HEAD의 엄밀한 조상이 아니다(merge-base `74ad197380a0`, `b1806498fe..HEAD` 37커밋). 지연 진술이 필요 없다. 설치 빌드는 `moai slot`, `moai verify audit-plan`, MCP 감사 서버(같은 `b1806498fe`)에만 쓰였고 교리 문서에 대한 측정값은 내지 않았다.

### Gaps

- AC-MJ-015(레인 수령 확인)는 리더 몫이라 관찰하지 않았다.
- Go는 이름 지정 14+6 테스트만 돌렸다. 두 패키지의 나머지, `go test ./...`, 저장소 전체 CI는 관찰하지 않았다.
- E-020j/k/l의 "접두사·기본 이름·허용 없음" 변이 세 개는 다시 만들지 않았다. 대신 현재 점검 스크립트에 대해 세 대조 셀을 직접 실행했고, 코드가 set 멤버십 한 줄임을 읽어 확인했다.
- codex의 델타 범위는 프롬프트로만 전달됐고 기계적으로 묶이지 않았다(residual_risk_note). claude·glm 레그는 계획상 꺼져 있다.
- 거부된 도구 호출 1건: 초기 셸 명령에서 `$(git rev-parse --show-toplevel)` 치환을 쓴 복합 명령이 워크트리 가드에 거부됐다. 경로를 직접 적은 단일 명령으로 다시 실행했고, 측정값을 다른 것으로 바꾸지 않았다.
- `moai slot release`를 처음에는 스크래치 디렉터리에서 실행해 "resource is not held"로 실패했고(다른 프로젝트 루트로 해석됨), 워크트리에서 다시 실행해 해제했다. 임대가 남지 않았음을 `slot go-heavy: free`로 확인했다.
- 원장 실행 전문과 `go test -v` 원본 로그는 세션 스크래치에만 있다. 여기 옮긴 줄만 인용할 수 있다.

### Residual-risk

- 원장은 어휘 기반이다. E-005e는 D 한 곳만 보며, "merge-ready"라는 낱말이 들어간 줄은 `done`과 함께 있어도 세지 않는다. N1처럼 다른 낱말로 같은 조건을 거는 문장은 여전히 감사의 읽기에만 의존한다.
- N1이 그대로면, F:20의 "carries" 정의를 엄격히 따르는 레인은 merge-ready 카드의 트리를 떠나지 않으려 할 수 있고, 그러면 빈 레인 정의에 따른 라우팅이 막힌다.
- D:139의 `ExitWorktree(keep)` 요구는 "A factory lane" 전체에 걸린다. F:180은 Codex 레인의 `ExitWorktree` 호출을 금한다. 이 범위 문제는 기준 커밋의 원래 줄에도 있었고 이번 델타가 만든 것은 아니지만, 이탈 시점이 늘어나 마주칠 일이 많아졌다.
- R-9(S4): 슬롯 대기의 일회성 재무장이 사실상 폴링이 될 가능성은 운영자 판정 전까지 남는다.
- R-1(SessionStart 레인 공지의 범위 없는 크론 문구), R-5(위임 성립 장치 없음)는 1회차와 같다.

## 6. 반복 이력

- 1회차(`sync-audit.md`, audited `f7cc3f8ee8`): FAIL 0.80, 차단 S1, 비차단 S2-S5, 필수 codex fail(영수증 `rcpt-aeda7d3bfadbf2012f680998`).
- 2회차(이 파일, audited `0ca3d59b2e`): PASS-WITH-DEBT 0.89, 차단 0건. S1·S2 해소, S3 정정, S4 R-9 유지, S5 수용. 새 결함 N1(SHOULD-FIX), N2·N3(MINOR), 모두 비차단. 필수 codex pass(영수증 `rcpt-a4b24c097a712091ae9c1d06`).
- 부채 처리 제안: N1은 매니페스트 개정이 필요하므로 후속 카드로 분리한다. N2·N3은 다음 문서 갱신 때 함께 고친다. R-9는 운영자 판정을 기다린다.
