---
id: SPEC-TUI-CHECKLIST-001
updated: 2026-10-27
---

# SPEC-TUI-CHECKLIST-001 — 실행 계획

마일스톤은 바뀌기 쉬운 결정부터 놓았다. 앞쪽(M0~M2)은 기준선·어휘·렌더러 모양처럼 리뷰에서 뒤집힐 수 있는 결정이고, 뒤쪽(M4·M6)은 연결과 정리처럼 기계적인 일이다. M5(update 연결)는 v0.2.0 에 카드 t1978 로 넘어갔다 — 번호는 다른 문서의 참조를 지키려고 비워 둔다. 시간 추정은 쓰지 않는다.

## §A 맥락

- 카드 t1969, Class C, P1. 워크트리 `.moai/worktrees/t1969`, 브랜치 `WT-moai-init-update`, plan 시점 HEAD `cb30e05b3e`(로컬 `main` 과 같음). 범위 축소 회차의 측정 HEAD `6f05252d6ddf62362a60cadbfc3350e0d068dbc4`(코드는 `cb30e05b3e` 와 같음 — acceptance.md 원장 L32).
- 범위(v0.2.0, 리더 결정 기록 §15.98): 렌더러(`internal/tui/checklist.go`, 네 언어 카탈로그 키, 골든 기준선) + `moai init` 연결. update 연결은 카드 t1978(이 카드 착지 뒤 시작). 이월 목록은 §K.
- 산출물: spec.md(REQ 23개 — id 그대로, update 조각만 걷어냄), acceptance.md(AC 23개 — id 그대로, v0.2.7 에 AC-TCL-008 을 t1989 로 넘김, 증거 원장 L1~L41), design.md, research.md, progress.md. 델타 4회차(v0.2.1)의 측정 HEAD 는 `274441db21a4be3cec01b78c0f9d496893d12479`(코드는 `cb30e05b3e` 와 같음).
- 개발 방식: TDD(RED-GREEN-REFACTOR). 단, M0 은 특성화(기준선) 커밋이다.

### §A.0 회차 기록

| 회차 | 측정 HEAD | 결과 | 바뀐 것 |
|---|---|---|---|
| 초안 v0.1.0 | `cb30e05b3e` | plan 감사 1회차 FAIL 0.71(감사 HEAD `d241917d17`, 차단 6·권장 6·경미 8, MP-8 실패) | — |
| 델타 1 v0.1.1 | `d241917d17` | plan 감사 2회차 FAIL 0.75(감사 HEAD `b422a13d53`, 차단 6·권장 4·경미 3), 리더 codex_audit 실패(P2 1건 = N8) | 결함 D1~D20 수리. REQ 20→23(REQ-TCL-021, REQ-TCL-022, REQ-TCL-023), AC 23→24(AC-TCL-024), 원장 L15~L19, 대체 골든 하위 디렉터리 `fallback/`, 순서 증거 O1~O3, 경고 수 정의 D-9, 보존 줄 목록(design.md §3.3.1), 레인 결정 Q1~Q3 반영 |
| 델타 2 v0.1.2 | `b422a13d53` | plan 감사 3회차 FAIL 0.80(감사 HEAD `69358ecba8`, 차단 3), 리더 codex_audit PASS | 결함 N1~N13·codex P2 수리. 보존 줄 목록 고정 검색(원장 L20), AC-TCL-024 행마다 하위 테스트, 재개 경로 실제 순서(원장 L21), O4 와 카드 id 완전성 검사, 결정 D-10 |
| 델타 3 v0.1.3 | `69358ecba8` | 델타 감사 4회차 FAIL 0.80(감사 HEAD `6f05252d6d`, 차단 1 — N1 update 24줄 예산 불성립, 델타 자격 불충족), 리더 codex_audit PASS | 3회차 fix_scope 결함 M1~M4 수리 |
| 범위 축소 v0.2.0 | `6f05252d6d` | 확인 plan 감사 5회차 FAIL 0.73(감사 HEAD `274441db21`, 차단 B1·B2), 리더 codex_audit PASS | 리더 SPLIT 판정(결정 기록 §15.98). update 연결을 t1978 로 이월(§K). REQ·AC id 유지. 3회차 감사 M5~M9 와 4회차 감사 N3·N5·N6 수리. 원장 L26~L33 |
| 델타 4 v0.2.1 | `274441db21` | 최종 확인 plan 감사 대기(scope full) | 리더 결정 기록 §15.102. B1: REQ-TCL-010 을 "체크리스트 줄 ≤ 24 + 직접 출력 줄 증가 0" 으로 재정의, 깨끗한 init 정의·줄 단위(M2) 수리. B2: REQ-TCL-015 의 두 조항을 AC-TCL-017 (a1)·(c)·(e) 로 판정, K-3 철회. S1~S3, M1, M3. REQ·AC id 유지(23·24). 원장 L34~L41 |

### §A.1 Tier 판정 — L 유지 (측정 근거)

리더는 축소 뒤 Tier M 을 예상했다. 이 카드가 만들거나 고치는 파일을 다시 세면 L 기준(15 파일 초과, 1000줄 초과)을 두 축 모두 넘으므로 L 을 유지한다.

| 근거 | 측정·추정 |
|---|---|
| 새로 만드는 코드·테스트 파일 6개 | `internal/tui/checklist.go`, `internal/tui/checklist_test.go`, `internal/cli/checklist_session.go`, `internal/cli/checklist_session_test.go`, `internal/cli/checklist_plan.go`, `internal/cli/checklist_fallback_test.go` — 여섯 모두 지금 없음(원장 L27, 종료 1) |
| 반드시 고치는 기존 파일 9개 | 카탈로그 `internal/tui/messages/{en,ko,ja,zh}.yaml`, `internal/tui/testdata/MANIFEST.txt`, `internal/cli/init.go`, `init_warnings.go`, `reporter.go`, `severity_line.go` — 모두 있음(이번 회차 `ls`, 종료 0) |
| 보존 줄 알림 때문에 고칠 수 있는 기존 파일 6개 | `init_update_notice.go`(P4), `worktree_advisory.go`(P5), `hook_install.go`(P20), `hook_install_precommit.go`(P21), `init_resume.go`(P23), `update/backup/file_snapshot.go`(P1·P2) — 모두 있음(같은 `ls`). 어느 쪽에서 알릴지는 run 단계가 고른다(design.md §3.3.1) |
| 새 골든 데이터 파일 20개 | 렌더러 6(`checklist-init-{clean,nocolor,reduced,ko}.golden`, `checklist-probe-{fail,warn}.golden` — v0.2.1 에 fail·warn 을 init 계획표 밖 키의 probe 계획으로 바꿈, design.md §7), 대체 12(`fallback/{init,update}-{clean,warn,fail}-{stdout,stderr}.golden`), 리치 2(`rich/init-{clean,warn}.golden`) |
| 파일 수 | 코드·테스트 15~21개 + 골든 20개 = 35~41개(15 초과). 골든을 빼고도 하한이 15 로 경계에 닿는다 |
| 줄 수(추정) | 비교 기준을 이번 회차에 쟀다: `wc -l` → `internal/tui/progress_line.go` 224, `progress_line_test.go` 316, `internal/cli/init_warnings.go` 130, `reporter.go` 118. 추정: 렌더러 ~300 + 테스트 ~400, 세션·계획표 ~300, init 연결 ~100, 카탈로그 4×~30, CLI 테스트 ~600, 골든 ~300 → 약 2,100(1000 초과) |

줄 수는 추정이다(Gap G-7). Tier 를 M 으로 내리려면 REQ·AC 를 16개 이하로 줄여야 하고(spec-workflow.md § SPEC Complexity Tier), 그러려면 id 를 지우거나 합쳐 모든 교차 참조를 다시 매겨야 한다. 측정이 L 이므로 그 위험을 지지 않는다. design.md·research.md 는 Tier L 산출물로 그대로 남는다.

## §B init.go 변경 영역과 t1966 조율

영역 기준 SHA: `cb30e05b3e`(REQ-TCL-021·AC-TCL-022 가 이 값을 읽는다). 행번호는 이 SHA 기준. t1966 hunk 목록은 research.md §7.

| 영역 | 행 | 하는 일 | t1966 hunk 와의 거리 |
|---|---|---|---|
| R1 | 383-389 (출력 관문) | `--verbose` 를 읽고 세션을 시작하는 줄과, 세션 종료 `defer` 줄을 printer 생성 줄 **앞**에 넣는다(가장 먼저 등록되어 가장 나중에 실행). 수집기에 세션을 넘긴다(388 수정 가능). | 332 뒤 삽입과 51줄, 561 뒤 삽입과 172줄 |
| R3 | 919-1075 (꼬리 단계) | 단계 앞마다 `cs.Item(…)` 한 줄을 **삽입만** 한다(기존 줄은 고치지 않음 — AC-TCL-022 가 `b=0` 으로 검사). 항목 키는 design.md §4 표. | 866 과 53줄 |
| R2 | 1077-1087 (완료 카드) | 세션이 켜져 있으면 카드 대신 요약 블록을 그리는 분기를 넣는다. 꺼져 있으면 지금 줄이 그대로 돈다. | 866 과 211줄 |

손대지 않는 줄: 728(보고기 연결 — t1966 의 730-749 삭제와 1줄 간격), 750(첫 Info — 삭제 구간 바로 다음), 780-790(자산 보장 — t1966 이 780 을 바꿈), 858-866(재개 경로 — t1966 의 866 수정과 겹침. REQ-TCL-023 은 세션 종료 `defer` 가 명명된 반환값으로 처리하므로 이 줄을 고칠 필요가 없다), 103-107 근처의 플래그 등록(t1966 이 바꿈). init 의 `--verbose` 등록은 새 파일 `checklist_session.go` 의 `init()` 함수에서 한다(패키지 변수 `initCmd` 는 모든 `init()` 보다 먼저 초기화된다). 보존 줄(design.md §3.3.1) 가운데 init.go 안의 지점(862, 863, 885, 890, 944, 945, 1034, 1058)은 init.go 를 고치지 않고 호출되는 함수 쪽에서 세션에 알린다.

조율 순서:

1. **t1966 이 먼저 착지하면** 이 카드는 M3 시작 전(= `internal/cli` 의 testdata 밖 파일을 처음 고치기 전)에 `main` 을 흡수한다. 흡수 뒤:
   - (a) research.md §7 표를 새 행번호로 다시 재고, 영역 기준 SHA 를 흡수 커밋으로 바꾸고 R1·R2·R3 경계를 갱신한다(manager-spec 재위임 — 본문 수정이므로).
   - (b) `go test ./internal/cli -run '^TestChecklistFallbackMatchesBaseline$' -count=1 -v` 를 흡수된 트리에서 돌린다. 녹색이면 기준선은 그대로 유효하다.
   - (c) 빨강이면(t1966 이 init·update 출력 문구를 바꿈) **재촬영 절차**를 탄다: manager-spec 델타 회차가 AC-TCL-013 을 다시 쓰고, 흡수 커밋 그 자체에서 골든을 다시 찍어 골든만 담은 커밋 하나로 넣는다. 흡수 커밋에는 이 카드의 코드가 **있다** — M0(대체 테스트와 골든), M1(`internal/tui/messages/*.yaml` 카탈로그), M2(`internal/tui/checklist.go` 렌더러)가 M3 전에 커밋되기 때문이다. 다만 그 코드는 아직 어떤 CLI 경로에도 연결되지 않아(연결은 M3·M4) init·update 의 출력 바이트를 바꾸지 않는다. 그래서 흡수 커밋의 출력은 "카드 코드가 출력에 닿지 않은 기준선"으로 쓸 수 있다. REQ-TCL-012 의 "기준선 트리"는 그 흡수 커밋이 되고, 순서 증거 O1 은 "M0 과 재촬영 커밋 두 줄"로 바뀐다. O3 은 범위를 `internal/cli`(`internal/cli/testdata` 와 `internal/cli/checklist_fallback_test.go` 제외)로, 조상 기준을 재촬영 커밋으로 바꾼다 — 범위를 `internal` 그대로 두면 재촬영 커밋보다 앞선 M1·M2 의 `internal/tui` 커밋이 빨강을 낸다. O4b 의 `<기준선 트리>` 는 흡수 커밋이다.
   - M3 이후에 흡수해야 하는 상황이 오면 이 카드의 코드가 섞인 트리에서는 기준선을 찍을 수 없으므로, 진행을 멈추고 리더에게 막힘 보고를 낸다.
2. **이 카드가 먼저 착지하면** t1966 이 흡수한다. 겹치는 줄이 없으므로 텍스트 충돌은 없을 것으로 본다. 다만 t1966 이 `resuming` 계산을 위저드 앞으로 옮기므로(hunk `@@ -561,0 +586,20 @@`, `@@ -775,6 +800 @@`) 의미 충돌이 없는지 흡수 쪽이 확인한다.
3. 어느 쪽이든 흡수 뒤 `go build ./...` 와 AC-TCL-013·AC-TCL-017 을 다시 돌린다.

범위 검사(AC-TCL-019 뒤 절, AC-TCL-021, AC-TCL-013 O3)는 이 카드의 커밋만 센다(`git log --no-merges --grep=t1969 <기준>..HEAD ^main`). 그래서 흡수만으로 빨개지지 않고, 흡수한 다른 카드의 커밋이 메시지에서 `t1969` 를 언급해도 세지 않는다. 카드 id 를 빠뜨린 이 카드의 커밋은 acceptance.md §A 의 카드 id 완전성 검사가 잡는다. AC-TCL-022 는 커밋 선택 대신 영역 기준과 HEAD 의 누적 diff 를 잰다.

## §C 사전 점검 (run 시작 시)

각 명령은 한 번의 호출이다. 출력을 progress.md §E.2 에 그대로 남긴다.

```bash
git branch --show-current
git rev-parse --short HEAD
git log --oneline -3 origin/main -- internal/cli/init.go
git diff --stat cb30e05b3e -- internal/cli/init.go
git rev-parse WT-moai-init-moai
git diff -U0 --output=<스크래치 경로>/t1966-init.diff cb30e05b3e...WT-moai-init-moai -- internal/cli/init.go
/usr/bin/grep -n "^@@" <스크래치 경로>/t1966-init.diff
go build ./...
```

- 넷째 줄 출력이 비어 있지 않으면(이 카드 밖에서 init.go 가 바뀜) §B 1번 절차로 간다.
- 다섯째~일곱째 줄은 t1966 진행을 다시 잰다. ref SHA 가 research.md §7 의 기록(델타 2회차 측정 `4e3b0a8714`)과 다르거나 hunk 머리가 바뀌었으면 research.md §7 을 갱신하고, R1·R2·R3 와 최소 간격(현재 53줄)을 다시 계산한다. 간격이 0 이하가 되면(겹침) 진행을 멈추고 리더에게 막힘 보고를 낸다.
- **재측정은 한 번으로 끝내지 않는다**(2회차 감사 남은 위험 1). 셋째·다섯째·여섯째·일곱째 줄(`origin/main` 의 init.go 이력, t1966 ref, t1966 hunk)을 run 시작 때, M3·M4 시작 때, 착지 직전에 다시 돌려 progress.md §E.2 에 남긴다. 특히 M4 는 init.go 를 처음 고치는 마일스톤이라, 그 시작 측정이 R1~R3 근처 새 hunk 를 보면 M4 를 시작하지 않고 막힘 보고를 낸다. 착지 직전 측정에서 겹침이 보이면 착지하지 않는다.
- **t1966 레인 알림.** run 시작 때 한 번, t1966 레인에 사실 통지를 보낸다: "t1969 가 init.go 의 R1 383-389·R3 919-1075(삽입만)·R2 1077-1087 을 고친다(기준 `cb30e05b3e`). 728·750·780-790·858-866 은 건드리지 않는다." 통지는 사실 전달이며 요청이 아니다(cross-session-messaging 규칙). 전달 결과(큐 적재·보류·분실)를 progress.md 에 남기고, 완료 판단은 통지 응답이 아니라 위 재측정으로 한다.
- 공유 ref 는 커밋된 것만 보인다. t1966 워크트리의 미커밋 변경은 이 점검으로 보이지 않는다(Gap G-1).

## §D 제약과 보존 목록

- 고치지 않는 것: `internal/cli/wizard/**`, `internal/template/templates/**`, `.github/workflows/**`, `internal/tui/progress_line.go`, `internal/tui/status.go`, `internal/tui/glyphs.go`, `internal/cli/printer/**`, 기존 골든(`internal/cli/testdata/tuxiu/**`, 기존 `internal/tui/testdata/*.golden`), M0 이후의 `internal/cli/testdata/checklist/fallback/**` 와 `internal/cli/checklist_fallback_test.go`.
- update 연결 코드를 넣지 않는다(spec.md §4 C-7): `runUpdate` 의 세션 생성, `update_template_sync.go` 의 항목 삽입, update 요약. init 과 함께 쓰는 파일을 고칠 때는 세션이 없을 때 출력이 그대로인 형태로만 고친다 — AC-TCL-013 의 `update_*` 하위 테스트가 회귀 경계다(세 고정 실행이 닿는 줄에 한함, §K K-16).
- `internal/tui/testdata/MANIFEST.txt` 는 `UPDATE_GOLDEN=1` 로만 재생성한다.
- 테스트 시접을 바꾸는 테스트는 `t.Parallel` 없이 `t.Cleanup` 으로 되돌린다.
- 전체 `go test ./...` 금지. 바뀐 패키지 범위만 돌리고, 전체 `internal/cli` 실행은 `moai slot acquire` 아래에서 한 번 또는 CI 에 맡긴다.
- 커밋은 경로를 지정해 스테이징한다. 커밋 메시지마다 카드 id `t1969` 를 넣는다 — 범위 검사가 이 id 로 이 카드의 커밋을 고른다.
- 이 카드의 테스트에 `[DEFERRED → t1978]` 하위 테스트를 두지 않는다(건너뜀 하위 테스트로도 남기지 않는다 — acceptance.md §F).

## §E 마일스톤

### M0 — 기준선 고정 (먼저, 단독 커밋) — Priority High

- `internal/cli/checklist_fallback_test.go` 와 `internal/cli/testdata/checklist/fallback/` 골든 12개(design.md §7: init·update 각각 clean·warn·fail 의 stdout·stderr)를 **현재 코드**로 찍어 넣는다. 주입은 모두 지금 코드의 시험 경로로 한다(AC-TCL-013 의 여섯 고정): init `warn` = `applyAutonomyTierBundleFn` 오류, init `fail` = `newInitPhaseExecutorFn` 실패 실행기, update `clean` = `--force --yes`, update `warn` = 사용자가 고친 pre-push 훅 + `--force --yes`, update `fail` = `--force`(확인 단계) + `confirmViaPreviewFn` 오류. update verbose 골든 2개는 t1978 몫이다.
- `init_clean` 고정은 spec.md §3.3 의 깨끗한 init 과 같다(임시 홈, `--all` 없음, 실제 사용자 자산 보장, git 있음) — M4 의 AC-TCL-011 이 이 골든의 직접 출력 줄 다섯 개를 기준선으로 쓴다. 다섯 조각이 정확히 한 번씩 있지 않으면 M0 을 커밋하지 않고 막힘 보고를 낸다.
- 정규화는 acceptance.md AC-TCL-013 에 열거한 정규식만 쓴다. 정규화가 줄 수를 바꾸지 않음을 테스트가 단언한다. `warn`·`fail` 골든에 주입 문장이 있음을 단언한다(양성 대조). `update_clean` stdout 골든에 훅 설치 줄이 정확히 1번 있음을 단언한다(회귀 경계가 공유 파일 줄을 담는지의 양성 대조).
- 이 커밋은 위 두 경로 밖의 파일을 담지 않는다(AC-TCL-013 O2). 대체 테스트는 git 이력을 읽지 않는다(O4b 가 저장소가 아닌 뽑아낸 트리에서 돈다).
- **커밋 직후, 다음 커밋 전에 O4a 를 기록한다**: `git rev-parse HEAD` 와 `go test ./internal/cli -run '^TestChecklistFallbackMatchesBaseline$' -count=1 -v` 를 각각 한 번 돌려, 측정 HEAD(40자)와 여섯 `--- PASS:` 줄을 progress.md §E.2 에 그대로 남긴다. 빨강이면 다음 마일스톤으로 가지 않고 M0 을 고친다 — 아직 뒤 커밋이 없으므로 M0 커밋 자체를 고쳐 쓰고(두 번째 커밋을 더하면 O1 이 두 줄이 된다), 새 HEAD 로 O4a 를 처음부터 다시 기록한다. 빨간 기록도 지우지 않고 남긴다. 이 기록은 레인의 주장이고, sync 감사의 O4b 재실행이 증거다(acceptance.md AC-TCL-013 O4).
- 이유: design.md §8.

### M1 — 어휘: 카탈로그 키 — Priority High

- design.md §6 의 이 카드 범위 키를 en·ko·ja·zh 에 넣는다(update 전용 키는 t1978). ja·zh 는 초안이다.

### M2 — 렌더러 — Priority High

- `internal/tui/checklist.go`, 단위 골든(init 의 clean·nocolor·reduced·ko, probe 의 fail·warn), `MANIFEST.txt` 재생성. 렌더러는 계획표·키·요약 필드를 호출자에게서 받는다 — init 전용 지식을 넣지 않는다(spec.md §3.5). probe 계획의 라벨은 테스트가 카탈로그 캐시에 시험 전용 언어로 넣고 지운다(design.md §7) — 생산 카탈로그 yaml 에는 넣지 않는다.
- AC-TCL-001, AC-TCL-002, AC-TCL-003, AC-TCL-004, AC-TCL-005, AC-TCL-006, AC-TCL-007, AC-TCL-009, AC-TCL-023.
- `@MX:ANCHOR` 후보: `NewChecklist`(init 의 세션·보고기, 그리고 t1978 의 update 가 부름). 고루틴이 없으므로 `@MX:WARN` 은 필요 없다.

### M3 — CLI 세션과 계획표 — Priority Medium

- 시작 전 §B 1번 조건(t1966 착지 여부)을 확인한다.
- `internal/cli/checklist_session.go`: 입력 수집 함수 `checklistEnv`, 활성 판정과 시접 세 개(design.md §3.1), 캡처 작성기, 두 작성기 모두의 경고 수신, 보존 줄 수신, 종료(오류·취소·`nil` 조기 반환 — REQ-TCL-015 의 요약 없는 실행 조항 두 개 포함), init `--verbose` 등록, 그리고 시험 전용 시접 `checklistSessionStartFn`(생산 값 `nil`, 수집기가 만들어진 직후 활성·비활성 모두에서 수집기와 `cmd` 로 한 번 부름 — design.md §3.7, AC-TCL-016(비활성 대조 포함)·AC-TCL-017 (c) 의 항목 밖 주입). 세션은 명령 중립이다 — t1978 이 같은 세션을 update 에 쓴다.
- `internal/cli/checklist_plan.go`: design.md §4 의 init 표.
- 보고기 이벤트 대응(design.md §3.6 표)과 항목 경계 네 경우(design.md §3.2, 결정 D-10).
- AC-TCL-010, AC-TCL-012(판정 단위), AC-TCL-014(init), AC-TCL-015, AC-TCL-017(error 부분).

### M4 — init 연결 — Priority Medium

- 시작 전 §C 의 t1966 재측정을 다시 돈다(init.go 를 처음 고치는 마일스톤).
- init.go R1·R2·R3(§B), `init_warnings.go`(세션에 경고 알림·활성 시 패널 생략·항목 밖 경고를 요약에 넘김), `reporter.go`(세션이 있으면 체크리스트 보고기), `severity_line.go`(경고 수신 인터페이스), 보존 행 P1~P5, P20·P21(init 지점), P23(design.md §3.3.1), 목록 완전성 하위 테스트(14개 모양, X11 분류 포함, 궤도 항목 열여섯 개).
- 리치 골든 `rich/init-clean.golden`, `rich/init-warn.golden`.
- AC-TCL-001(명령 수준 `TestChecklistInitConditionalSkip` — R3 여는 줄의 조건부 `Skip`, design.md §4), AC-TCL-011, AC-TCL-012(init 명령 수준, `init_writer_replaced` 포함), AC-TCL-013(`update_*` 회귀 경계 재실행), AC-TCL-016, AC-TCL-017(cancel, resumed_after_validate_fail — 실제 실행기, nil_return_open_item), AC-TCL-018, AC-TCL-019, AC-TCL-020, AC-TCL-022, AC-TCL-024.

### M5 — update 연결 — [DEFERRED → t1978]

- 이 카드에서 하지 않는다. v0.1.3 의 M5 내용(세션 시작·종료, 템플릿 동기화 항목, update 요약 위치 D-10, update 보존 행 P6~P19·P22, 리치 골든 `rich/update-*.golden`, AC 의 update 하위 테스트)은 §K 와 design.md §4.1 이 t1978 에 넘긴다.

### M6 — 마감 — Priority Low

- AC 전체 재실행과 progress.md §E.2·§E.3 기록.
- AC-TCL-013 순서 증거 O1·O2·O3 실행과 기록, O4a 기록의 측정 HEAD 와 O1 SHA 대조, acceptance.md §A 카드 id 완전성 검사 실행과 기록. O4b 는 sync 감사가 돈다.
- `TestChecklistInitHunkConfinement` 의 마지막 실행(두 명령·그대로의 출력·종료 코드·측정 SHA)을 기록한 **뒤** 제거(AC-TCL-022 의 일회성 성격).
- lint·커버리지·교차 빌드.

## §F 자체 검증 산출물 (run 단계 보고)

각 항목은 명령·그대로의 출력·측정 HEAD 를 함께 적는다.

- E1 AC 표(AC-TCL-001 부터 AC-TCL-024 까지 가운데 AC-TCL-008 을 뺀 23개, §A 의 `--- PASS:` 판정. AC-TCL-022 는 제거 전 기록. `[DEFERRED → t1978]` 조각은 표에 넣지 않는다)
- E2 `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`
- E3 `go test -cover ./internal/tui`
- E4 `golangci-lint run --timeout=2m` 새 지적
- E5 AC-TCL-013 순서 증거 O1·O2·O3 원문, O4a 기록(M0 직후의 측정 HEAD 와 여섯 `--- PASS:` 줄), 카드 id 완전성 검사 원문
- E6 AC-TCL-022 의 누적 diff `@@` 줄 원문과 병합 검사 출력
- E7 §C 사전 점검 출력(run 시작, M3·M4 시작, 착지 직전)과 t1966 레인 알림의 전달 결과

## §G 레인 결정 (Kickoff 기록에서 리더에게 알림)

세 결정은 레인 판단(사다리 마지막 단계)으로 정했다. 근거 기록: `.moai/reports/t1969/progress.md` 결정 기록. 모두 되돌릴 수 있다.

- **Q1 (결정 D-1) 비TTY·CI 출력 = 기준선과 같은 바이트.** 독해: 카드 문구 "TTY 아님/CI/--verbose 면 평문 줄 출력으로 폴백"을 "오늘의 평문 줄 출력으로 폴백"으로 읽었다. 오늘 비TTY 출력은 이스케이프 없는 평문 줄이라 문구를 문자 그대로 충족하고, 완료 조건의 "비TTY 골든"은 M0 기준선이 채운다. 공개한 가정이며 부채로 들고 간다. **리더가 다른 독해를 택하면**(카드가 `--verbose` 만 "현재 상세 그대로"라 따로 적었으니 비TTY·CI 는 평문 체크리스트로 바꾸라는 뜻): manager-spec 델타 회차가 REQ-TCL-012·AC-TCL-013·AC-TCL-014 와 M0 을 다시 쓰고, 렌더러에 평문 모드를 더하며(design.md §2 의 YAGNI 결정 철회), 기존 버퍼 기반 테스트 가운데 init 출력을 단언하는 것들을 새 모양으로 옮긴다.
- **Q2 (결정 D-6) 체크리스트가 켜지면 init 경고 패널 생략.** 항목 안 경고는 항목 아래 한 번 펼치고, 항목 밖 경고는 요약이 다시 나열한다(REQ-TCL-015, design.md §3.3). 경고 수는 수집기 수 하나로 정의했다(결정 D-9). **리더가 패널 유지를 택하면**: REQ-TCL-015 와 AC-TCL-016 의 패널 0번 조건을 1번으로 바꾸고, 체크리스트 줄 24 예산(AC-TCL-011)은 깨끗한 실행 기준이라 그대로 두되 경고가 있는 실행은 예산 밖임을 명시한다.
- **Q3 사용자 자산 보장 단계(`init.go:780-790`) 항목화는 후속 카드.** t1966 이 `:775-780` 을 바꾸므로 그 착지 뒤로 미룬다. 이 단계는 `executor.Execute`(`init.go:829`)보다 먼저 돌아 열린 항목이 없으므로 출력이 넘김 규칙으로 그대로 보인다. **리더가 이번 카드에 넣기를 택하면**: t1966 착지 뒤 §B 1번 절차로 흡수하고, 계획표에 `assets` 항목을 더하는 델타 회차를 연다.
- **글리프 (결정 D-2).** 카드의 `➤`·`–` 대신 `◆`·`○`. 이모지 검사와 허용 목록 충돌 때문이다(design.md §1.1).
- **결정 D-10 (update 요약 위치).** v0.1.3 의 레인 결정은 t1978 로 넘어간다(§K). 이 카드에서는 항목 경계 네 경우만 계약이다.
- **줄 예산 셈 규칙 (리더 결정 기록 §15.102, v0.2.1).** 리더 판정이며 레인 결정이 아니다. REQ-TCL-010 은 체크리스트가 그리는 줄(머리·환경·그룹 제목·항목·요약) ≤ 24, 그룹당 항목 ≤ 8 이고, 기존 직접 출력 줄(`init.go:708`·`:750`·`:784`·`:1034`)은 바꾸지 않고 따로 세며 M0 기준선 대비 증가 0 이다. 근거: 카드 완료 문구 "1화면 이내 (항목 수 ≤ 그룹당 8)", 그리고 `init.go:750` 이 허용 영역 밖이다(AC-TCL-022).
- **열린 운영자 질문 (t1978 과 공유, 결정 기록 §15.102).** 직접 출력 줄까지 합친 화면 전체의 줄 상한 — 24줄 고정인가, 현실값(지금 셈 28줄)인가. 리더가 정하지 않았다. 이 카드는 위 셈 규칙으로 닫으며, 운영자가 화면 전체 상한을 정하면 직접 출력 줄의 처리(축약·항목화)는 별도 카드가 된다(spec.md §5).

## §H 위험

| ID | 위험 | 영향 | 대응 |
|---|---|---|---|
| R-1 | t1966 과 init.go 동시 변경 | 흡수 때 충돌·의미 충돌 | §B 영역 분리(최소 53줄 간격), §C 재측정(run 시작, M3·M4 시작, 착지 직전)과 t1966 레인 알림, 흡수 뒤 AC-TCL-013·AC-TCL-017 재실행, AC-TCL-022. 공유 ref 는 커밋된 변경만 보이므로 측정 사이에 커밋된 t1966 변경은 다음 측정에서 드러난다 |
| R-2 | 캡처가 대화형 화면을 삼킴 | 사용자가 질문을 못 봄 | 위저드는 `cmd` 작성기를 쓰지 않음(research.md §4). AC-TCL-019 가 열린 줄 0 을 확인. 실제 TTY 육안 확인은 Gap G-2 |
| R-3 | 접힌 상세에 경고·오류·안내가 섞여 사라짐 | 정보 손실 | 경고는 신호(수집기·`emitSeverityLine`)로, 직접 쓰기 줄은 보존 줄 목록(design.md §3.3.1, REQ-TCL-022)으로 따로 알린다. 목록은 `internal/cli` 전체 고정 검색 14개 모양(원장 L20·L26)의 모든 일치를 분류하고, AC-TCL-024 가 init 행을 하나씩 주입하며 `inventory_complete` 가 새 일치를 바로 잡는다. 남은 위험 셋: (가) 14개 검색 모양에 걸리지 않는 새 직접 쓰기(예: 새 `p.Info` 안내), (나) 궤도 밖 파일이 새로 init 에서 불리게 되는 변경(궤도는 한 단계 호출 추적과 출력 타입 파일로 정했다 — research.md §9), (다) `internal/cli` 밖 패키지가 명령 작성기를 받아 쓰는 경우(이번 검색 범위 밖) |
| R-4 | init 체크리스트 줄 24 초과, 또는 직접 출력 줄 횟수가 기준선과 다름 | 한 화면 목표 위반 또는 정보 손실 | 항목 묶음(design.md §4), AC-TCL-011(체크리스트 줄 ≤ 24, 직접 출력 줄 다섯의 기준선 대비 증가 0·감소 0). 깨끗한 init 은 체크리스트 23줄 + 직접 출력 5줄(spec.md §3.3, 코드 읽기). M0 기준선에서 다섯 조각이 정확히 한 번씩 나오지 않거나 M4 실측이 넘으면 기준을 느슨하게 하지 않고 manager-spec 에 되돌린다. 화면 전체 상한은 운영자 질문(§G). update 예산은 t1978(§K) |
| R-5 | ja·zh 초안 품질 | 어색한 문구 | 원어민 검수는 범위 밖(Gap G-5), 키 동치·en 과 다름은 카드 t1989 |
| R-6 | 활성 판정 오탐 — 터미널인데 출력이 리디렉트된 경우 | 체크리스트가 파일에 기록됨 | 입력 수집을 한 함수로 모으고 fd 를 따로 판정(design.md §3.1). AC-TCL-012 명령 수준 하위 테스트 `init_stdout_not_tty`(`moai init > log`)·`init_stderr_not_tty`·`init_writer_replaced` 가 호출 지점 오류를 잡는다 |
| R-7 | `ProgressLine` 의 비TTY 줄이 상세로 접힐 때 형식 변화 | 실패 펼침 모양이 예상과 다름 | AC-TCL-003 골든. `/usr/bin/grep -ln "ProgressLine(" internal/cli/init.go internal/cli/init_warnings.go internal/cli/init_resume.go internal/cli/reporter.go internal/cli/init_update_notice.go` → 빈 출력, 종료 1(이번 회차, `6f05252d6d`) — 이 다섯 파일에는 호출이 없다. 그 밖의 init 호출 경로는 재지 않았다. update 쪽 위험은 t1978 |
| R-8 | 렌더러가 하나라는 제약(카드)과 보고기 어댑터 | 사실상 두 번째 렌더러처럼 보임 | 어댑터는 이벤트만 넘기고 그리지 않는다. 그리는 코드는 `checklist.go` 하나 |
| R-9 | 세션 종료 `defer` 순서 착오 | 패널이 캡처에 갇힘 | design.md §3.4, AC-TCL-017 |
| R-10 | `TERM=dumb`·Windows 콘솔의 줄 지우기 지원 | 잔재 문자 | `TERM=dumb` 는 대체 경로. Windows 실제 콘솔 확인은 Gap G-3 |
| R-11 | AC-TCL-022 테스트가 git 이력을 읽음 | 얕은 클론 CI 에서 M4~M6 사이 빨강 | 테스트는 이력 부족을 분명한 메시지로 실패시킨다. M6 에서 지운다. M4~M6 사이 원격 CI 후보 푸시가 있으면 그 실패를 이 사유로 분류한다 |
| R-12 | 공유 파일 변경이 update 출력을 바꿈 | update 사용자에게 의도치 않은 출력 변화 | spec.md §4 C-7, AC-TCL-013 의 `update_*` 회귀 경계(`--force --yes` 고정 실행으로 훅·스냅숏 꼬리까지 담음). `internal/cli/update/backup` 의 프로세스 stderr 직접 쓰기 10곳(X10)은 update 쪽 위험이라 t1978 로 넘긴다(§K) |

## §I 가정과 Gap

가정:

- A-1 huh 폼과 Bubble Tea 미리보기는 `cmd` 작성기가 아닌 자기 출력(터미널)에 쓴다(research.md §4 의 코드 읽기에 근거, 실행 관측 아님).
- A-2 init 위저드 답의 언어를 세션이 받을 수 있다. 받지 못하면 프로필 로케일, 그다음 `en`.

확인된 사실(이전 가정 A-3):

- `TestManifest` 는 골든이 새로 생긴 경우(`internal/tui/golden/index_test.go:73` `golden file %s is new`)와 지워진 경우(`:80` `manifest entry %s is missing on disk`)를 모두 실패시킨다. 델타 1회차에서 HEAD `d241917d17` 의 `:68-83` 을 읽어 확인했다.

Gap:

- G-1 트리 격리 가드가 `git -C <t1966 워크트리>` 를 거부했다. 대신 공유 ref `WT-moai-init-moai` 를 이 트리에서 읽었다(델타 2회차 재측정 `4e3b0a871488d45f2b56e76dbf29958abd3317cb`, hunk 머리 10개 그대로). t1966 워크트리의 커밋되지 않은 변경은 보지 못했다. §C 가 run 시작과 M3·M4 시작, 착지 직전에 다시 잰다. 범위 축소 회차에서는 t1966 ref 를 다시 재지 않았다.
- G-2 실제 TTY 에서 위저드 화면과 체크리스트가 섞이지 않는지 육안 확인을 하지 않았다(plan 단계라 코드가 없음).
- G-3 Windows 콘솔에서의 `\r\x1b[2K` 동작을 확인하지 않았다.
- G-4 이 계획의 `moai spec lint` 판정은 `go run ./cmd/moai` 로 이 트리에서 빌드한 바이너리가 냈다(progress.md §E.1).
- G-5 ja·zh 문구 원어민 검수 없음.
- G-6 긴 라벨 자르기 동작은 정하지 않았다.
- G-7 §A.1 의 줄 수는 추정이다(비교 파일 네 개의 `wc -l` 만 실측).
- G-8 [DEFERRED → t1978] update 깨끗한 실행의 24줄 적합성 — 4회차 감사 N1 이 코드 읽기로 불성립을 보였다(design.md §4.1). init 의 체크리스트 23줄과 직접 출력 5줄도 코드 읽기이며, 직접 출력 줄은 M0 이, 체크리스트 줄은 M4 가 처음 실측한다.
- G-9 O4b 의 `git archive` → `tar -xf` → `go test -C <뽑아낸 트리>` 절차는 plan 단계에서 돌려 보지 않았다(M0 이 아직 없다). sync 감사가 처음 돈다.
- G-10 보존 줄 목록의 궤도(design.md §3.3.1)는 궤도 파일에서 작성기를 넘기는 호출을 한 단계만 따라가 정했고, v0.2.0 에 출력 타입 파일 두 개를 손으로 더했다. 두 단계 이상 떨어진 호출과 다른 출력 타입은 추적하지 않았다(§H R-3 (나)).
- G-11 재개 경로의 이벤트 순서(design.md §3.4)는 `phase.go`·`init.go`·`reporter.go` 를 읽어 정했고 실행으로 관측하지 않았다. AC-TCL-017 (c) 가 실제 실행기로 처음 관측한다.
- G-12 범위 축소 회차는 `go test` 를 돌리지 않았다(기계 부하 지시). 원장 L11(테스트 선택자 대조군)은 다시 재지 않았다.
- G-13 update 회귀 경계의 고정 실행(`--force --yes`, `git init` 한 고정 프로젝트)이 내는 줄 — `--force` 가 더하는 백업·확인 줄, 저장소가 아닐 때 훅 설치기 동작 — 은 이 회차에 읽지 않았다. M0 이 골든으로 처음 찍는다.
- G-14 (v0.2.1) 코드 읽기로 정한 것: (가) update `warn` 고정에서 끝에 한 줄을 덧붙인 MoAI pre-push 훅을 `classifyHook` 이 "사용자 수정" 으로 판정해 `hook_install.go:268` 경고가 나온다는 것(`:202-225` 만 읽고 `classifyHook` 본문은 읽지 않음), (나) update `fail` 고정에서 `--yes` 없는 `--force` 실행이 `update_template_sync.go:1287` 의 `confirmViaPreviewFn` 에 닿기 전에 다른 확인이나 오류로 멈추지 않는다는 것, (다) 깨끗한 init 에서 직접 출력 줄 다섯이 정확히 한 번씩 나온다는 것. 셋 다 M0 의 양성 대조와 기준선 골든이 처음 실측하며, 어긋나면 M0 커밋 전에 막힘 보고가 된다.
- G-15 (v0.2.1) 반례 `init.go:427`(git 미설치 경고)은 실제로 git 을 빼서 재현하지 않고, 같은 모양(항목 밖·즉시 줄 있는 `Warn`)의 주입으로 판정한다(AC-TCL-017). `PATH` 에서 git 을 빼면 다른 단계의 동작이 바뀌는데 그 영향을 재지 않았다.

## §J 참고

- SPEC-V3R6-UPDATE-PROGRESS-001 — `ProgressLine`
- SPEC-CLI-TUX-INIT-UPDATE-001 — 글리프 SSOT, update 카드, 정체성 띠
- SPEC-V3R6-UPDATE-NOISE-001 — update `--verbose` 의 뜻
- SPEC-INIT-QUIET-WIZARD-001 — init 위저드 질문 범위

## §K t1978 로 넘긴 항목 (리더 결정 기록 §15.98)

이 SPEC 은 아래를 판정하지 않는다(K-3 은 철회된 행 — 이 SPEC 이 판정한다). t1978 의 plan 단계가 자기 SPEC 으로 다시 쓴다. "이 SPEC 의 앵커" 는 근거가 남아 있는 곳이다.

| # | 무엇 | 왜 넘기나 | 이 SPEC 의 앵커 |
|---|---|---|---|
| K-1 | 델타 4회차 감사 N1 — 깨끗한 update 의 24줄 예산이 설계로 성립하지 않음(동기화 건너뜀 + 종료 블록·자산 요약 ≥32줄) | update 줄 예산 자체가 update 연결의 결정이고, 24줄 고정인지 "한 화면"인지는 운영자 질문 | spec.md §3.3 마지막 항목(깨끗한 update), §3.4 F-1 행; design.md §4.1 |
| K-2 | 델타 4회차 감사 N2 — AC-TCL-016 의 `:961` 항목 위치(`deploy.restore` 대 `verify.retention`) 불일치 | update 행의 항목 위치 | acceptance.md AC-TCL-016 의 `[DEFERRED → t1978]` 근거 지점 문단; design.md §4.1 표 `verify` 행 |
| K-3 [WITHDRAWN — 리더 결정 기록 §15.102, v0.2.1] | 델타 4회차 감사 N4 — REQ-TCL-015 의 "요약 없이 끝나는 실행의 `nil` 반환 쓰기"·"이미 실제 작성기에 나간 경고는 다시 쓰지 않음" 두 조항을 보는 기준이 없음 | **이월하지 않는다.** v0.2.0 은 §15.98 을 따라 t1978 로 넘겼으나, 두 조항의 실제 경로는 init(오류 경로 `init.go:866`·`:877`·`:879`, 재개 `nil` 반환 `:864`)이고 t1978 은 update 만 다룬다 — 넘기면 어느 카드도 판정하지 않는다(확인 감사 5회차 B2). §15.102 가 이월을 철회했다. 이 SPEC 의 AC-TCL-017 (a1)·(a2)·(c)·(e) 가 판정하고(반례 `init.go:427` 두 번 보임), spec.md §3.5 의 "세션 종료 계약은 이 SPEC 이 지킨다" 와 맞는다 | spec.md REQ-TCL-015, §3.5; acceptance.md AC-TCL-017; design.md §3.4 |
| K-4 | 델타 4회차 감사 N7 — AC-TCL-016 update 하위 테스트의 꼬리 주입 시접이 이름으로 없음 | update 꼬리 시접은 update 연결의 일 | acceptance.md AC-TCL-016 의 `[DEFERRED → t1978]` 문단 |
| K-5 | 델타 4회차 감사 N3 의 대상인 P22(이미 최신 알약)와 AC-TCL-017 (d) | P22 는 update 행 | design.md §3.3.1 P22 행, §3.4 update 이미 최신 문단; acceptance.md AC-TCL-017 (d) |
| K-6 | update 쪽 보존 행 P6~P19, P22 와 P20·P21 의 update 지점 — 이 SPEC 에서는 X11 로 줄 단위 분류 | update 세션이 없으면 보존 행 판정이 성립하지 않음 | design.md §3.3.1 P 표·X11 행; acceptance.md AC-TCL-024 |
| K-7 | 3회차 감사 M8 의 `update_tux.go:120`(`! N conflict` 알약)과 L26 의 `update_wizard.go:39`·`:102`·`:105` 의 보존 여부 | update 행 | design.md §3.3.1 X11 행; acceptance.md 원장 L26 |
| K-8 | update 연결 REQ 조각(활성, 대체 강제, `--verbose`, 채널 stdout, 병합 확인 경계, 숫자 동치, 심각도 줄 펼침) | 리더 SPLIT | spec.md §3.4 표 |
| K-9 | update 계획표·항목 순서·update 요약 위치(결정 D-10 의 update 절)·update 줄 예산 | 리더 SPLIT, K-1 | design.md §4.1; spec.md §2 D-10 |
| K-10 | update `--verbose` 골든 2개와 AC-TCL-014 의 `update`·`update_verbose_differs` | update `--verbose` 의 대체 강제는 update 연결 | acceptance.md AC-TCL-014; design.md §7 |
| K-11 | AC-TCL-011·AC-TCL-012·AC-TCL-016·AC-TCL-017·AC-TCL-018·AC-TCL-019·AC-TCL-020·AC-TCL-024 의 update 하위 테스트, 리치 골든 `rich/update-{clean,fail}.golden`, 렌더러 골든 `checklist-update-clean.golden` | update 연결 | 각 기준의 `[DEFERRED → t1978]` 표시 |
| K-12 | 3회차 잔여 R-M4(깨끗한 update 정의에 `--yes` 없음)와 R-M4b(design.md §3.5 의 update 확인 문구) | update 정의·update 경계 | design.md §3.5 셋째 항목, §4.1 |
| K-13 | `internal/cli/update/backup` 의 프로세스 stderr 직접 쓰기 10곳(X10, 옛 위험 R-12) | update 경로에서만 실행 중 줄과 섞일 수 있음 | design.md §3.3.1 X10 행 |
| K-14 | update 전용 카탈로그 키(`checklist.header.update`, `checklist.group.backup`·`deploy`, update 항목 키) | update 계획표에 묶임 | design.md §6 표 |
| K-15 | research.md 의 update 사실(§3, §8 의 update 문단, 4회차 감사 N5 로 정정한 `:127`) | 배경 증거 | research.md §3 머리 표시, §8 |
| K-16 | (v0.2.1, 확인 감사 5회차 S2) REQ-TCL-012 의 update 절 — 이 SPEC 은 AC-TCL-013 의 세 고정 실행으로 좁혔다. 세 실행이 닿지 않는 공유 파일의 실패 경로 줄(`hook_install.go:283`·`:291`·`:293`, `update/backup/settings_snapshot.go:115` `warnLine` 등)을 지키는 일, 그리고 update 연결 뒤 TTY update 에서 거짓이 되는 그 절의 개정 | update 연결이 그 절의 뜻을 바꾼다 | spec.md REQ-TCL-012, §3.4 F-10; acceptance.md AC-TCL-013 |
