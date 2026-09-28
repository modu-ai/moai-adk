auditor-model: claude-opus-5-5[1m]

# t1226 sync-audit — SPEC-ALWAYS-LOADED-HEADROOM-001

- 감사 대상: 워크트리 `.claude/worktrees/t1226`, 브랜치 `WT-always-loaded-headroom`, HEAD `fcae46594447b4859aad8505a0820ab0d94dbe6a`, 기준 `7fe658815`
- 감사자: sync-auditor (독립 재실행, 기록된 출력은 신뢰하지 않음)
- 모델 출처: 이 감사 세션 시스템 프롬프트의 모델 식별 문장("The exact model ID is claude-opus-5-5[1m]") — 런타임 자기 보고
- 임시 경로 표기: 감사 세션 scratchpad 는 `$SCRATCH` 로 적는다

## 판정

**Overall Verdict: PASS-WITH-DEBT — 90.6 / 100** (가중 조화평균; 비가중 조화평균 90.5)

| Dimension | Score | Verdict | Evidence (재실행 원문 요지) |
|---|---|---|---|
| Functionality (40%) | 92 | PASS | AC-ALH-001~010 전 명령 재실행, 기대값 전부 일치(아래 §1). `CHECK-init=OK`·`CHECK-live=OK`·`CHECK-init_17=OK`, `TOKEN-live=OK`·`TOKEN-init_17=OK`·`TOKEN-init=OK`, `HASH-BAD=0` |
| Security (25%) | 88 | PASS | 하네스 재실행 후 실제 홈 지문 `HOME-UNCHANGED`; 하네스 자체 가드 통과(`--- PASS`). 감점: 플래그 경로에 대한 무방비 `os.RemoveAll`(F1) |
| Craft (20%) | 90 | PASS | `golangci-lint run ./internal/cli/...` → `0 issues.`(v2.1.6), `go vet` 무출력, `gofmt -l` 무출력, 플래그 없는 실행 `--- SKIP` 0.00s. 감점: 탐욕 해제 수치의 낙관 편향(F3) |
| Consistency (15%) | 92 | PASS | `moai spec lint --strict` → `✓ No findings`, 카드 id 가 모든 커밋에 존재, 상태 `completed`. 감점: HISTORY 완료 행 부재·`sync_commit_sha` 미보충(F5) |

MUST-PASS(Functionality·Security) 둘 다 임계 이상. 차단(blocking) 결함 0건, 선택(optional) 결함 6건. 선택 결함만으로는 PASS 를 FAIL 로 바꾸지 않는다.

판정 토큰 `STRUCTURALLY-INFEASIBLE-UNDER-FREEZE`(S_init 18·17 집합, S_live)는 이 감사가 독립 재계산으로 재현했고, 반증 시도(§3~§5)에서 뒤집히지 않았다. 다만 그 견고성은 이 SPEC 이 정의한 `A_adm` 범위와 한 번의 압축 시도 수율에 조건부이며, 그 조건이 판정서에 완전히 드러나 있지 않다(F2).

## §1 AC 재실행 (acceptance.md 명령 원문, `$SCRATCH/ac.sh`)

`ac-run.sh` 는 먼저 읽었고(읽기 전용, `$SCRATCH` 만 씀) 그대로 쓰지 않았다. acceptance.md 명령을 새 스크립트로 옮겨 재실행했다. 두 표면 트리는 감사자가 새로 만들었다 — `S_live` 는 `git archive 05d79c8b2 … -o $SCRATCH/live.tar` 후 해제, `S_init` 은 하네스 재실행(§2) 산출물.

| AC | 재실행 핵심 출력 | 결과 |
|---|---|---|
| 001 | `1` `1` `2` `1` / `0` `0` | PASS |
| 002 | `UTF-8`, `1 1 1`, `PIN-live-OK`, `PIN-init-OK`, `remeasured live: 199111 total`, `remeasured init: 203413 total`, 판정서 `total_init = 203413`·`total_live = 199111`, `_미측정` `0` `0`, `missing_init = -` | PASS |
| 003 | `KEYS-init-OK` `KEYS-live-OK` `SPLIT-init-OK` `SPLIT-live-OK`; `ROWS/OVERLAP/DESTIN/M1P/NETNEG BAD=0`, `EVID MISSING=0`, `DEST-init BAD=0` `DEST-live BAD=0`(두 표면); `dest-sizes.tsv` 14개 목적지를 두 트리에서 다시 `wc -m` — 불일치 0 | PASS |
| 004 | `hash_init remeasured: 93de7321ea747c584768af09d1908ad3e36fd076073a6f2a9d6fb5b7a22b4477`(구속 줄 `170`), `hash_live remeasured: d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3`(`170`), 해시 줄 `2`, `HASH-BAD=0` | PASS |
| 005 | `d0e61541367abb06a170bd36b6376e51d51882380ca9f899f0d2934016a78547`, `1 1 1`, 파일 존재, `/tmp` 인용 `0` `0`; `172ef22eb` 아카이브에서 `sec.py 3` 재실행 → `SEC-RERUN-IDENTICAL`, `psection.py` 재실행 → `PSECTION-RERUN-IDENTICAL`, `246943−73126−10522+8400 = 171695` | PASS |
| 006 | `init recomputed: C=203413 A=14201 R=477 U=79 T_min=189689`, `live … C=199111 A=13716 R=477 U=79 T_min=185872`, `init_17 … C=198447 A=13717 R=454 U=79 T_min=185184`; `PAIRS-init=OK` `PAIRS-live=OK`; 포인터 원문 `wc -m` 다중집합 `{64,87,89,146,23,68}` = TSV `pointer_chars` 다중집합 | PASS |
| 007 | `TOKEN-live=OK` `TOKEN-init_17=OK` `TOKEN-init=OK`, 상신 절 `4`, `RECOMMEND:` `1`, 결정 서술 `0` | PASS |
| 008 | `git log --first-parent --no-merges --format=%H 7fe658815..HEAD -- <18경로 + 미러(템플릿 core·workflow 디렉터리 전체로 확장)>` → 무출력(0행). 양성 대조: 같은 형식으로 `verdict.md internal/cli` 를 주면 5행. 범위 안 병합 커밋 0개(`git log --format='%h %p'` 전부 단일 부모) — `--first-parent` 가 가리는 것이 없음 | PASS |
| 009 | 기록본 `--- PASS` `1`, `--- SKIP` `0`, `harness_head = 05d79c8b27b1aacfc56946672e5187b0fcabf96e` = `build_head`; `shasum -a 256 internal/cli/init_headroom_export_test.go` → `8776d7b5809b47a6215a79497f6036414ceb2967d32067db612d07289b49e9c5` = `harness_sha256`; `prepareSafeInitHome` `2`; `LOG-OK`; `moai … init` 정규식 `0` | PASS |
| 010 | 형태 1 줄 `0 0 0 0 0`, 형태 2 `runtime_observed = no` `1`, `count_set_init = 18` `1`, 17집합 줄 `6`, `verdict_init_17` `1` | PASS (형태 2) |

품질 게이트: `$SCRATCH/moai-audit spec lint --strict SPEC-ALWAYS-LOADED-HEADROOM-001` → `✓ No findings — all SPEC documents are valid`. `golangci-lint version` → `v2.1.6`(CI 판), `golangci-lint run ./internal/cli/...` → `0 issues.`

## §2 하네스 재실행과 실제 홈 비접촉

- `BH`(05d79c8b2) 와 HEAD 사이의 변경은 `.moai/reports/t1226/**` 와 SPEC 문서뿐이다(`git diff --name-only 05d79c8b2 HEAD` 전수 확인, `internal/`·`cmd/`·`go.mod` 변화 0). 그래서 HEAD 에서 돌린 하네스는 `BH` 코드와 같다.
- 명령(한 번의 복합 호출): `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_DISTRIBUTE_ALL && go test ./internal/cli/ -run '^TestHeadroomInitSurfaceExport$' -count=1 -v -args -headroom-export=$SCRATCH/init-surface`
- 출력: `--- PASS: TestHeadroomInitSurfaceExport (1.19s)`, `ok  github.com/modu-ai/moai-adk/internal/cli 2.087s`, `headroom-path` 18줄 전부 `present`. 18개 sha256 을 기록본 `harness-run.txt` 와 대조 → `HARNESS-HASHES-IDENTICAL`.
- 실제 홈 독립 지문(실행 전후, 감사자 스크립트): `~/.claude/settings.json`·`~/.zshrc`·`~/.bashrc`·`~/.profile`·`~/.zprofile`·`~/.bash_profile` sha256 과 `~/.claude/hooks` 상태 → `HOME-UNCHANGED`. `~/.moai/claude-profiles` depth 2 파일 sha256: `preferences.yaml`·`launch.yaml` 은 해시 동일(나열 순서만 다름), 변한 것은 `moai-adk/.claude.json` 하나(mtime 23:44:03). 이 파일은 Claude Code 런타임 상태 파일이고 내용에 `headroom`·`TestHeadroom`·`/var/folders/kt` 가 0건이며, 하네스는 `CLAUDE_CONFIG_DIR` 을 비운다 — 하네스 귀속 근거 없음(Gap G2). `~` 최상위의 marker 이후 변경은 `~/.codex`(23:49:30, 하네스 종료 5분 뒤, 동시 codex 세션)뿐.
- 플래그 없는 실행(CI 경로): `--- SKIP: TestHeadroomInitSurfaceExport (0.00s)`, `ok … 0.782s`.

## §3 DEBT-1~3 변이 검사 (`$SCRATCH/mut.sh`, TSV 사본만 변조)

대조(무변조): `ROWS BAD=0 OVERLAP BAD=0 DESTIN BAD=0 M1P BAD=0 NETNEG BAD=0 EVID MISSING=0 HASH-BAD=0`.

| 변이 | 표적 | 발화 |
|---|---|---|
| ADMIT M2 `chars = gross+1` | DEBT-1 (≤ gross) | `ROWS BAD=1`, `HASH-BAD=1` |
| ADMIT M1 `chars = gross−1` | DEBT-1 (M1 == gross) | `ROWS BAD=1` |
| ADMIT M2 `chars+1` (여전히 ≤ gross) | DEBT-1 (pre−post) | `HASH-BAD=1` |
| ADMIT M2 `gross+5` | DEBT-1 (pre == gross) | `HASH-BAD=1` |
| `CLAUDE.md · Selection Decision Tree` 절 행을 ADMIT 로(문단 행 ADMIT 유지) | DEBT-2 | `ROWS BAD=1`, `OVERLAP BAD=1` |
| 같은 변이에 `bind=0 gov=N c1=Y` 까지 위장 | DEBT-2 단독 | `OVERLAP BAD=1` |
| ADMIT M1 목적지 → `workflow/skill-routing.md` | DEBT-3 | `DESTIN BAD=1` |
| ADMIT M1 목적지 → `core/moai-constitution.md` | DEBT-3 | `DESTIN BAD=1` |
| ADMIT M1 → `M1p`(dup_source 없음) | M1p 증거 | `M1P BAD=1` |
| `citation-breaks` 행 `c2=Y` | 사유↔조건 | `ROWS BAD=1` |
| ADMIT M1 → `net-negative` 위장 | net-negative 증거 | `NETNEG BAD=1` |
| ADMIT M2 → `rewraps-binding-line` 위장 | 해시 방향 | `HASH-BAD=1` |
| M1 행 5,000자로 부풀림(목적지 acp-reference) | 수용량 | `DEST-cap BAD=1` |

13개 변이 전부 사살. DEBT-1~3 검사는 공허하지 않다.

## §4 기각 판단 표본 검토 (14행, seed 1226, 큰 행 8 + 나머지 6)

모집단: init `governs-scope` 148행 48,990자(그중 문단 행 117개 33,498자), `narrows-scope` 30행 14,235자, `citation-breaks` 3행 90자 — 합 181행 63,315자(판정서 표와 일치).

| 행 | gross | 판단 | 감사 의견 |
|---|---:|---|---|
| kanban · Integration … self-served ¶3 | 1,114 | a | 방어 가능 — [HARD] 줄이 요구하는 통합 규칙 본체 |
| goal-directive · Hard Preconditions | 960 | b | 방어 가능 — Kickoff 게이트 불변 단서 |
| cache-aware · (H1 서문) | 865 | b | 부분적 — "change no gate semantics" 는 b 이나 캐시 설명 대부분은 압축 여지 있음 |
| acp · Ledger Closure ¶3 | 865 | a | 방어 가능 — MUST 가 열거하는 (a)~(d) 본체 |
| session-handoff · Invariants | 835 | b | 방어 가능 |
| CLAUDE.md · HARD Rules (Mandatory) | 852 | a | 논쟁적 — constitution 의 요약 중복, `§1` 인용 앵커라 제거는 어려우나 압축 여지 큼 |
| session-handoff · Output Surface ¶2 | 710 | a | 방어 가능 — 비-방출 정의 |
| VCI · §3 ¶3 (5절 표) | 647 | a | 방어 가능 — [HARD] 줄이 이름 붙인 표 |
| kanban · Isolation ¶3 | 260 | a | 방어 가능 |
| kanban · Isolation ¶15 | 266 | b | 방어 가능 — 카드 PR 한정 |
| askuser · Structural Constraints ¶2 | 224 | a | 논쟁적 — 용어 주석, companion 이동 가능해 보임 |
| native-idiom · The Invariant ¶2 | 180 | b | 방어 가능 — 영어 세션 면제 |
| constitution · Enforce Simplicity ¶2 | 94 | a | 방어 가능(사소) |
| branch-guard · Staleness Rule ¶2 | 65 | a | 방어 가능 |

14행 중 11행은 "존재해야 한다"는 판단이 방어 가능, 3행(약 1.9k자)은 논쟁적. 표본 비율(≈20%)을 모집단 63,315자에 외삽해도 재분류 가능량은 약 1.3만자이고, 그 전부를 M1 로 완전 제거해도 필요량 39,690 에 한참 못 미친다. **개연성 있는 재분류로는 `T_min` 이 150,000 아래로 내려가지 않는다.**

판정이 뒤집히는 조건은 재분류가 아니라 정의와 수율이다(F2): 기각 풀과 M2 풀(56,179)을 합친 119,494자에 제자리 압축을 허용하고 균일 수율이 **40.9%** 를 넘으면 `T_min` 이 150,000 아래가 된다(`(189689 + 9179 − 150000) / (63315 + 56179) = 0.409`, `$SCRATCH/sample.py`·계산 재현). 실측 수율은 16.3%(9,179 / 56,179), 표본으로 본 압축본(`99347713b2` 33%, `6f1889ee2d` 15%)은 의미를 지우지 않은 유능한 압축이었다. 구조적 바닥도 쟀다 — `bind>0` 절에서 문단 행을 뺀 잔여 77,335자, 구속 줄 170개 자체 55,397자.

계수 집합 민감도(`$SCRATCH/cs.py`): 18집합 `T_min 189689`, 17집합 `185184`, AGENTS.md·yaml 2개·skill-routing 을 뺀 14집합에서도 `166549` — 런타임 계수 집합이 무엇으로 관측되든 판정 토큰은 같다.

## §5 해제 산술·상신 절 검토

- `unfreeze.py candidates-init.tsv 189689` 재실행 → `UNFREEZE-RERUN-IDENTICAL`. `need = 39690 (T_min 189689 - 149999)`, `T_min_after = 148851`, `all_bind_rows = 99`, `all_bind_lines = 170`, `all_bind_gain = 131257`. 줄당 이득 내림차순 정렬 확인(순위 12→13: 1,499 > 1,414, 15→16: 1,377 > 1,360.5, 18→19: 1,296 > 1,280.5).
- U1 가정 명시: (b) 절 끝 "해제된 절 전체를 목적지로 옮길 수 있다는 가정(아래 U1)을 둔 상한", (c) 에 U1·U2·U3 정의, 포인터 재유입 미포함은 Gaps 에 있음 — 요구 충족.
- 상신 절은 결정하지 않는다: (d) "해제 여부는 운영자가 정한다 … 레인은 결정하지 않고 권고만", `RECOMMEND:` 1줄은 세 선택지를 함께 올리자는 권고와 순서 선호뿐. 결정 서술 grep `0`.

## Findings (structured defect-list)

- **F1** [Medium] [optional] `internal/cli/init_headroom_export_test.go:125` — `os.RemoveAll(exportDir)` 가 `-headroom-export` 플래그 값에 아무 제한 없이 실행된다. `-headroom-export=.` 이나 홈 디렉터리를 잘못 넘기면 그 트리를 지운다. `prepareSafeInitHome` 의 지문 가드는 8개 항목만 보므로 이 삭제를 막지 못한다. 신뢰 경계(사용자 입력) 위의 파괴적 원시 연산이다. Confidence: high. Required fix: 삭제 전에 경로가 존재하지 않거나 빈 디렉터리이거나, 이전 실행이 남긴 표식 파일을 가진 경우에만 지우고, 그 밖에는 `t.Fatalf` 로 거부한다(또는 `os.TempDir()` 하위만 허용).
- **F2** [Medium] [optional] `.moai/reports/t1226/verdict.md` § Residual-risk / § 동결 해제 상신 절차 — 판정의 조건부성이 완전히 드러나지 않았다. (i) REQ-ALH-005 는 조건 1 단서 (a)·(b) 에 걸린 63,315자를 M2(제자리 압축)에서도 배제하는데, 선행 `SPEC-ALWAYS-LOADED-DIET-002/plan.md:80` 은 그 단서를 "옮기지 않는다"(M1)로만 규정했고, 이 배제는 동결 해시(구속 줄 다중집합)와 무관하다. (ii) 그 풀까지 제자리 압축을 허용하면 균일 수율 40.9% 에서 150,000 을 넘는다. 판정서는 (i)을 명명하지 않고 16.3% 외삽(≈179,000)만 적었다. 운영자에게 "구속 줄 해제" 말고 "단서 행의 신중한 제자리 압축 허용 + 수율 측정"이라는 동결 비해제 경로가 있음을 알려야 상신 자료가 완결된다. Confidence: high(산술), medium(해석). Required fix: 리드 상신 시 이 교차 수율과 정의 의존을 한 줄로 함께 제시한다(판정서 수정은 선택).
- **F3** [Low] [optional] `verdict.md` § (b) 탐욕 해제 집합 — 머리 수치 "24절·27줄·40,838 → 148,851" 은 순위 1 `AGENTS.md`(U3 필요)를 포함한다. U1 만 허용한 변형도 24절·27줄이지만 `T_min_after = 149997`로 여유가 3자뿐이고(감사자 재계산), 절마다 생길 포인터 재유입(24개 이상)을 넣으면 150,000 을 넘는다. 또한 `agent-common-protocol-reference.md`(39,227자)는 이 집합의 acp 이득 9,623자를 받을 수 없어 새 companion 이 필요하다. Required fix: 상신 시 "27줄은 하한, 재유입·수용량 반영 시 더 늘어난다"로 표기한다.
- **F4** [Low] [optional] Gap 표기 — 런타임 계수 집합 미관측(AC-ALH-010 형태 2)은 판정서가 정직하게 Gap 으로 적었고, §4 민감도가 보이듯 판정 토큰에 영향이 없다. `spec.md §A` 의 t1184 관측("18파일 249.2k", 격리 init 트리)이 18집합 가설을 지지하는 기존 관측이라는 점을 AC-ALH-010 절에서 인용하면 좋다. Required fix: 없음(선택적 인용).
- **F5** [Low] [optional] `.moai/specs/SPEC-ALWAYS-LOADED-HEADROOM-001/spec.md` HISTORY 에 run/sync 완료(`completed`) 행이 없고, `progress.md §E.4` `sync_commit_sha: pending-backfill-sync` 가 미보충이다. Required fix: 병합 시 HISTORY 한 행 추가와 `fcae46594` 보충.
- **F6** [Low] [optional] `internal/cli/init_headroom_export_test.go` — 플래그가 없으면 SKIP 하므로 CI 비용은 사실상 0(0.00s)이고 플레이크 위험도 없지만, 스위트에 대한 회귀 가치도 없다(`--- SKIP` 은 초록으로 보인다). 측정 하네스를 테스트 스위트에 상주시키는 선택 자체는 수용 가능하다. Required fix: 없음. 원하면 빌드 태그(`//go:build headroom`)로 분리해 일반 `go test` 목록에서 제외.

## Recommendations

- 리드는 운영자 상신 때 판정서 (a)~(d) 에 F2·F3 의 두 줄(교차 수율 40.9%, 탐욕 27줄은 하한)을 덧붙여 올린다. 둘 다 판정 토큰을 바꾸지 않는다.
- F1 은 후속 카드에서 두어 줄로 막을 수 있다. 이 카드의 병합을 막을 사안은 아니다.

## Gaps

- **G1** 교차 모델 감사(`audit_multi`/`codex_audit`/`glm_audit`)는 실행하지 않았다. 이 카드의 범위는 `7fe658815..HEAD` 인데 도구의 `baseBranch` 대상은 원격 기본 브랜치(`main`) 기준 diff 를 보내고, `uncommittedChanges` 는 깨끗한 트리라 비어 있다 — 어느 대상도 이 카드의 변경만을 가리키지 못한다. Claude 단독 감사다.
- **G2** 하네스 실행 창(23:43:34~23:44:02) 안에 `~/.moai/claude-profiles/moai-adk/.claude.json` 이 바뀌었다(23:44:03). 내용에 하네스 흔적이 없고 동시 세션이 쓰는 런타임 파일이라 하네스 귀속 근거는 없지만, 귀속을 배제하는 관측도 없다.
- **G3** 런타임이 보고하는 always-loaded 파일 수·합계는 관측하지 않았다(판정서와 같은 Gap).
- **G4** M2 압축본의 의미 보존은 표본 2건만 읽었다. 110행 전체의 의무·예외·수치 보존은 검증하지 않았다.
- **G5** 기각 판단은 181행 중 14행만 표본 검토했다.

## Residual-risk

- 판정은 한 작성자의 한 번 압축 시도(수율 16.3%)와 SPEC 의 보수적 `A_adm` 정의에 기댄다. 둘 다 바뀌면(단서 행 압축 허용 + 수율 41% 이상) 판정이 뒤집힐 수 있다. 개연성 있는 재분류만으로는 뒤집히지 않는다.
- `commands.log` 는 자기 신고이고, `moai init` 비격리 실행 부재는 그 기록 범위 안에서만 확인된다. 이 감사 세션은 `moai init` 을 실행하지 않았다.
- 하네스는 테스트 안에서 init 명령 경로를 탄다. 설치 바이너리의 셸 진입점과 플래그 파싱 이전 단계는 거치지 않는다.
