auditor-model: claude-opus-5-5

# SPEC Review Report: SPEC-ALWAYS-LOADED-HEADROOM-001
Iteration: 2/2 (Tier M — `harness.yaml` `plan_audit_tier_ceilings.M: 2`, 이번이 상한 회차)
Verdict: FAIL
Overall Score: 0.77 (1회차 0.62 → 상승, STOP 신호 없음. Tier M PASS 문턱 0.80 미달)

카드 t1226 · 감사 대상 커밋 `b231487ab` (spec 0.2.0) · 브랜치 `WT-always-loaded-headroom` · 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1226`

Reasoning context ignored per M1 Context Isolation. 오케스트레이터 지시문의 과제 목록만 범위로 받았고, 수리 커밋 메시지와 `progress.md` 대응표는 주장으로만 읽었다.

교차 모델 감사는 호출하지 않았다. `.moai/config/sections/` 에 `audit_model` 키가 없다(`grep -rn audit_model` 0행). 1회차와 같은 이유로 Claude 단독 감사다.

## 감사자 재측정 (이 실행, 이 트리 `b231487ab`)

| 항목 | 명령 | 관측 |
|---|---|---|
| HEAD·브랜치 | `git rev-parse --short HEAD` · `git branch --show-current` | `b231487ab` · `WT-always-loaded-headroom`, 작업 트리 깨끗함 |
| 1회차 이후 변경 | `git diff --stat 0a4ff87dd b231487ab` | SPEC 5파일 + `verdict.md` + `sec.py` 신규(39줄). 7 files, +526 −202 |
| `S_live` 합계 | 18경로 `wc -m … \| tail -1` | `199111 total` — 변함없음 |
| 18경로 무수정 | `git diff --quiet 7fe658815 HEAD -- <18경로·미러 루트>` | `exit=0` |
| `develop` 이동 | `git rev-parse develop origin/develop` | 둘 다 `7fe658815eb0…` |
| `skill-routing.md` | 두 트리 머리 5줄 `grep paths:` | 두 트리 모두 `paths: ".claude/agents/**,.claude/skills/**,.moai/config/sections/delegation.yaml"` |
| `sec.py` 출처 | `shasum -a 256 .moai/reports/t1226/sec.py /tmp/claude-501/sec.py` | 두 줄 모두 `d0e61541…78547` — 일치 |
| REQ 번호 | `grep -oE 'REQ-ALH-[0-9]{3}' spec.md \| sort -u` | 001~017, 빈틈·중복 없음 |
| clarification 마커 | `grep -rn 'NEEDS CLARIFICATION' <SPEC 디렉터리>` | 0행(`exit=1`) |
| `syscall` | `grep -c syscall` 5파일 | 전부 0 |
| D7 참조 | `grep -Eo 'SPEC-…'` 5파일 → `grep '^status:'` | DIET-001·DIET-002 모두 `status: completed` |
| 린트 | `go build -o $SCRATCH/moai ./cmd/moai` → `$SCRATCH/moai spec lint --strict SPEC-ALWAYS-LOADED-HEADROOM-001; echo exit=$?` | `built` / `0 error(s), 0 warning(s)` / `lint exit=0`. INFO 1건 `OwnershipTransitionUnmeasured` — 대상은 최초 커밋 `10281a857`(트레일러 없음) |
| 수리 커밋 트레일러 | `git log -1 --format=%B b231487ab` | `Authored-By-Agent: manager-spec` 있음 |
| AC 코퍼스 게이트 | `go test ./internal/spec -run TestACCounterFullCorpusMatchesBaseline -count=1 -v` | `exit=0`, `--- PASS … (10.18s)`. 부재 행 3건 보고, 그중 `absent-from-snapshot .moai/specs/SPEC-ALWAYS-LOADED-HEADROOM-001/acceptance.md: COUNT 11` |
| AC 식별자 | `manager-docs.md` 카운터 본문과 같은 정규식을 인라인 awk 로 실행 | 현재 `AC-ALH-001`~`010` + `AC-ALD2-002`(L35) = 11. v0.1.0 은 9(`AC-ALH-001`~`008` + `AC-ALD2-002`) |
| 하네스 헬퍼 | `grep -rn 'func prepareSafeInitHome\|func runInitWithFlags' internal/` | `internal/cli/init_home_guard_test.go:145`, `internal/cli/init_deploy_exit_test.go:50` |
| 테스트 플래그 선례 | `grep -rn 'flag\.\(String\|Bool\)(' internal/cli/*_test.go` | `update_version_downgrade_test.go:26` `-update-golden` |
| 실 `runInit` 선례 | `init_quiet_wizard_test.go:90` | `prepareSafeInitHome(t)` 뒤 실제 `runInit` 구동 |
| 흡수 내성 양성 대조 | t1175 흡수 병합 `8fb81c948` 에서 `git log --no-merges 8fb81c948^1..8fb81c948` | **53행**(다른 카드의 비병합 커밋). 룰 경로 pathspec 적용 시 `43697af85`·`237d5e3e7`(t1064, 템플릿 미러·독트린) 출력. `--first-parent` 를 붙이면 **0행** |
| AC-ALH-008 pathspec | `git ls-files <36경로> \| wc -l` | `36` — 오타 없음 |
| 행 규칙 변이 | AC-ALH-003 (2) awk 를 `bind=0·gov=N·c1~c4=Y` 행 둘을 `REJECT net-negative` 로 채운 TSV 에 실행 | `mut.tsv BAD=0` |
| 백엔드 줄 정규식 | AC-ALH-001 첫 grep 을 자리표시 줄 + 정상 줄에 실행(`locale charmap`=UTF-8) | `1` — 자리표시는 거부, 정상 줄만 통과 |

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: `spec.md:L72-88` 에 REQ-ALH-001 ~ REQ-ALH-017 이 빈틈·중복 없이 3자리로 이어진다.
- [PASS] MP-2 GEARS 준수(요구사항 층 `spec.md §C` 대상): 17개 REQ 모두 `shall`/`shall not` 을 가진다. 신설 REQ-ALH-016(L87)은 Ubiquitous, REQ-ALH-017(L88)은 Ubiquitous 에 금지문 한 줄이 붙은 형태로 허용 범위 안이다. REQ-ALH-013(L84)은 Unwanted 로 바뀌었다. AC 는 검증 층이라 여기서 채점하지 않았다.
- [PASS] MP-3 frontmatter: `spec.md:L2-14` — 12개 필드 모두 있음, `version: "0.2.0"` 따옴표, 거부 별칭 없음.
- [N/A] MP-4 언어 중립성: 다언어 도구 SPEC 이 아니다. 추가될 Go 테스트는 이 저장소 내부 도구이며 템플릿 경로가 아니다.
- [PASS] MP-5 D7: 참조 SPEC 둘 다 `completed`. BLOCKING 0.
- [PASS] MP-6 D8: `syscall` 0회.
- [PASS] MP-7 clarification gate: `plan.md`·`research.md` 에 마커 0행. 1회차 D20 의 리터럴 표기도 사라졌다.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 대부분 해소. 남은 것: 관측 집합이 18경로와 다를 때 `total_init`·`current_init` 의 정의가 AC-ALH-006 과 AC-ALH-010 사이에서 충돌(N5), 17경로 집합의 `A_adm`·`T_min` 산출법 미정(N4), AC-ALH-003 (1)의 `$SCRATCH/init-surface` 출처 미기재(N8) |
| Completeness | 0.75 | 0.75 | 모든 절 존재(`### Out of Scope — …` 5개, `spec.md:L94-114`). Tier M REQ 상한 16 초과(17, N6) |
| Testability | 0.65 | 0.50–0.75 사이 | 뼈대 공허 통과는 해소(verdict.md:L3 자리표시는 정규식이 거부). 그러나 1차 판정 표면의 수치 `total_init`·`hash_init`·`hash_live` 가 형식만 검사되고(N1), `net-negative` 기각이 무조건 통과하며(N2, 변이 BAD=0 실측), `verdict_init_17` 은 재계산 수단이 없고(N4), AC-ALH-008 은 흡수 뒤 거짓 FAIL 을 낸다(D14-R, 양성 대조 실측) |
| Traceability | 1.00 | 1.00 | `acceptance.md:L413-433` — REQ 17개 모두 AC 에 매핑, AC 10개 모두 실재 REQ 를 가리킴. §D 매트릭스(L11-22)와 §D.2 일치 |

총점은 네 차원의 조화평균(0.768)을 반올림했다.

## 1회차 결함 판정 (D1–D21)

| 결함 | 판정 | 근거 |
|---|---|---|
| D1 계수 집합 대조 | 요구 층 해소 · 검증 설계에 신규 결함 | REQ-ALH-016(`spec.md:L87`)·AC-ALH-010(`acceptance.md:L383-403`) 신설. 그러나 관측 대상 디렉터리와 관측 격리에 결함(N3), 17 토큰 미검증(N4), 관측 집합 분기와 AC-ALH-006 충돌(N5) |
| D2 `S_init` 실행 경로 | 해소 | 헬퍼 둘이 실재하고 SPEC 서술대로 동작한다(아래 §신규 표면 검토). 비격리 실행 금지 명문화(`spec.md:L88`, `plan.md:L97`) |
| D3 바이너리 출처 | 해소(보조 지적 N8) | 바이너리 대신 `build_head` + 커밋된 하네스 파일 해시(`acceptance.md:L32-33`, L369-372). 하네스가 트리에서 컴파일되므로 출처로 성립 |
| D4 공허한 머리 검사 | **부분 해소** | `S_live` 는 재실행 대조(L143), 절 내 `_미측정` 0건(L144-145), 뼈대 자리표시 거부 확인. 그러나 1차 표면 `total_init` 은 「정수」만 검사(L142) — N1 |
| D5 백엔드 값 고정 | 해소 | 형식 정규식(L96)이 자리표시 거부·정상 줄 통과를 실측. 뼈대 L3 은 `_미관측_` |
| D6 표 완결성·임의 기각 | **부분 해소** | 키 재생성 diff(L159-165)와 사유 열거(L66-81)는 들어왔다. 그러나 `net-negative` 는 L187 에서 조건 없이 `ok=1` — 1회차가 지적한 「`bind=0`·`c1~c4=Y` 행을 임의 사유로 기각」 경로가 그대로 열려 있다(변이 `BAD=0`) — N2 |
| D7 조건 1 단서 | 해소 | 용어 표 `spec.md:L60`, `gov` 열(`acceptance.md:L56`), `ADMIT` 은 `$7=="N"` 요구(L172), §4.3 기각 목록 검토 항목(L207) |
| D8 목적지 누적 | 해소 | `dest` 열(L61), `dest-sizes.tsv`(L84), 두 트리 누적 awk(L199-204), REQ-ALH-015 |
| D9 UNDETERMINED 우회 | 해소 | `gross` 열, `U = Σ gross(UNTRIED)`(L268), `untried:<사유>` 필수(L176), 토큰 재계산(L295-301), UNDETERMINED 도 상신 필수(`spec.md:L82`, `acceptance.md:L411`) |
| D10 재조정 선택지 | 해소 | `(a)\|(b)\|(공표값)\|(기타)` + `J_includes_kanban_scope` 줄(L246-247) |
| D11 F 정의 | 해소 | `F(172ef22eb) = N` 서술값(`spec.md:L66`, L78), `^F = ` 0건 검사(L101) |
| D12 `/tmp` 금지 충돌 | 해소 | 금지를 `evidence` 열·`evidence:` 줄로 한정(L250-254) |
| D13 `sec.py` 출처 | 해소 | 커밋됨, 해시 일치 실측, AC-ALH-005 해시 검사(L243-244) |
| D14 흡수 내성 | **미해소 — 수리 형식이 틀렸다** | L319 의 `git log --no-merges 7fe658815..HEAD` 는 develop 흡수 뒤 다른 카드의 비병합 커밋을 모두 포함한다. 양성 대조: t1175 흡수 병합에서 같은 형식이 53행, 룰 pathspec 으로 t1064 의 `43697af85`·`237d5e3e7` 을 낸다. **이 형식은 1회차 보고서의 Required fix 가 제시한 것이다 — 감사자 제안 자체가 결함이었다.** `--first-parent` 형식은 같은 범위에서 0행 — D14-R |
| D15 `AGENTS.md` M1p | 해소 | L174 `$1 ~ /AGENTS\.md/ && $5!="M2" && $13=="ADMIT"` |
| D16 표면별 해시 | 해소(값 검증은 N1) | `hash_init`·`hash_live` 분리, AC-ALH-004 가 표면별 `H` 사용(L220) |
| D17 증거 존재 | 해소 | 모든 행 존재 검사(L193-196) |
| D18 최소 해제 집합 | 해소 | 「탐욕 해제 집합」 개명·겹침 규칙(`plan.md:L83`, `spec.md:L82`) |
| D19 분할 범위·로케일 | 해소 | 서문 행(`acceptance.md:L440`), yaml `config-data`(`spec.md:L74`), `charmap = UTF-8`(L140) |
| D20 마커 리터럴 | 해소 | grep 0행 |
| D21 트레일러 | 해소(잔여 INFO 수용) | `b231487ab` 에 트레일러. `10281a857` 의 INFO 는 이력 재작성 없이는 남으며 `--strict` 결과를 바꾸지 않는다 |

## 신규 표면 검토

### REQ-ALH-017 / AC-ALH-009 — 격리 하네스

- **헬퍼 실재와 동작.** `prepareSafeInitHome`(`internal/cli/init_home_guard_test.go:145-219`)은 실제 홈 지문(8항목)을 먼저 뜨고, `userHomeDirFn`·`profile.BaseDirOverride`·`MOAI_HOME` 을 `t.TempDir()` 아래로 돌리고, `CLAUDE_CONFIG_DIR` 등을 비우고, 셸 설정 seam 을 호출만 세는 스파이로 바꾸고, 돌린 홈이 실제 홈 안이면 `t.Fatalf` 로 init 을 거부하며, 종료 시 지문을 비교한다. SPEC 의 「홈 쓰기 네 곳을 돌린다」(`spec.md:L88`)는 정확하다(셋은 경로 이동, 하나는 스파이).
- `runInitWithFlags`(`internal/cli/init_deploy_exit_test.go:50-86`)는 전역 `initCmd` 에 `root`·`name`·`language`·`mode`·`non-interactive` 를 세팅하고 `initCmd.RunE` 를 직접 부른다. `extra` 로 `name`·`llm` 을 덮을 수 있다. 계약의 고정 플래그(`acceptance.md:L379`)는 모두 `init.go:73-137` 에 정의된 플래그다.
- 실제 `runInit` 을 이 헬퍼 아래에서 돌리는 선례가 있다(`init_quiet_wizard_test.go:90`). 테스트 전용 플래그 선례도 있다(`-update-golden`). 따라서 하네스 경로는 실행 가능성이 높다 — 다만 이것은 선례에서 추론한 것이며 하네스 자체는 아직 존재하지 않는다(Gap).
- **SKIP=FAIL 은 강제 가능하다.** `-v` 가 있고(L360), `--- PASS: TestHeadroomInitSurfaceExport ` 1건과 `--- SKIP` 0건을 요구한다(L367-368). 테스트가 없거나 셀렉터가 아무것도 고르지 않으면 PASS 줄이 0 이 되어 FAIL — 빈 스윕이 드러난다.
- **`-run` 앵커**는 `'^TestHeadroomInitSurfaceExport$'`(L360)로 있다.
- **「18개 파일을 고치지 않는다」·템플릿 우선 규칙과의 충돌 없음.** REQ-ALH-013 의 pathspec(L319-344)에 `internal/cli` 가 없고, `spec.md:L97` 이 하네스 한 파일을 명시적으로 범위에 넣는다. Template-First 규칙은 `.claude/`·`.moai/`·`.agency/` 신규 파일에 걸리며 Go 테스트는 대상이 아니다.
- 잔여 지적은 N8(하네스 출력의 HEAD 귀속)과 N11(린트 게이트) — 둘 다 optional.

### REQ-ALH-016 / AC-ALH-010 — 런타임 관측 위임과 18/17 대체 판정

- **대체 판정의 방향은 건전하다.** 17경로 집합은 18경로에서 `skill-routing.md` 를 뺀 것이므로 `T_min_17 = T_min_18 − (그 파일의 잔여 기여 ≥ 0)` 이고, 따라서 18 이 `ACHIEVABLE` 이면 17 도 `ACHIEVABLE` 이다. 갈림은 18 이 비-ACHIEVABLE 일 때만 생기고, 그때 규칙은 `UNDETERMINED` 로 보내 상신을 요구한다(`spec.md:L82`). 두 토큰을 조작해도 상신은 피할 수 없다 — 상신 회피 방향으로는 비조작적이다.
- **그러나 `verdict_init_17` 은 검증 수단이 없다.** AC-ALH-010 은 형식만 본다(L395). `T_min_init_17`·`U_init_17` 줄도, 17집합의 `A_adm` 산출 규칙도 없다. 레인이 17 을 18 과 같게 적으면, 17 로는 `ACHIEVABLE` 인 상황이 `STRUCTURALLY-INFEASIBLE-UNDER-FREEZE` 로 상신된다 — 운영자에게 불필요한 동결 해제를 묻는 오도된 상신이다 — N4.
- **관측 설계 결함** — N3.
- **관측 집합 분기가 AC-ALH-006 과 모순** — N5.

### `candidates.py` · TSV 재생성 diff

키 `(file, section, gross, bind)` 재생성 diff(L159-165)는 「TSV 가 스크립트 출력과 같다」를 증명한다. 「스크립트가 파일을 빠짐없이 나눴다」는 증명하지 않는다 — 스크립트도 레인이 run 단계에서 쓴다. 파일별로 최상위 분할 행(서문 + 절, `¶` 행 제외)의 `Σ gross` 가 그 파일의 `wc -m` 과 같다는 검사를 넣으면 분할 완결성과 `total_*` 값 검증을 한 번에 얻는다(N1 수리안에 포함).

### AC-ALH-007 토큰 재계산

L295-301 의 awk 는 REQ-ALH-010 의 경계(`T_min < 150000` / `T_min − U ≥ 150000` / 그 사이)를 그대로 옮겼고, `verdict_<s>` 줄이 정확히 하나(`n==1`)이길 요구한다. `verdict_init_18`·`verdict_init_reason` 줄은 `$1` 이 달라 섞이지 않는다. 예외(L308)는 `verdict_init_reason = count-set` 줄을 요구해 기계적으로 식별된다. 건전하다.

### AC 개수 11 대 10 — 어느 쪽이 맞는가, 스냅숏이 필요한가

- **둘 다 자기 기준으로 맞다.** 정본 카운터(`manager-docs.md` 의 `MOAI-AC-COUNTER` 본문)는 왼쪽 경계 없이 `AC-([A-Z0-9]+-)*[0-9]+` 를 모두 센다. 이 파일에서 그것은 자기 AC 10개 + `acceptance.md:L35` 의 선행 SPEC 식별자 `AC-ALD2-002` = **11** 이다. 코퍼스 게이트도 `COUNT 11` 로 보고했다. 문서가 정의하는 인수 조건은 **10개**(§D 매트릭스, §D.1)다. 차이는 표시 없는 외부 참조 하나다.
- **지금 스냅숏을 커밋할 의무는 없다.** 이 파일은 스냅숏 `.moai/reports/t338/ac-count-baseline.txt` 에 기록돼 있지 않다(`grep HEADROOM` 0행). 게이트는 부재 행을 「report, not fail」로 다루며(실측: `--- PASS`), `.moai/docs/ac-count-baseline-refresh.md §2` 는 「새 `acceptance.md` 의 추가는 방아쇠가 아니다」라고 적는다. 네 번째 방아쇠(제자리 개정으로 개수 변경)는 **기록된** 개수가 움직일 때의 이야기다 — v0.1.0(9) → v0.2.0(11) 의 변화는 기록된 행이 없어 게이트를 붉히지 않는다. 저장소 교훈 「acceptance.md 를 제자리 개정하면 AC 스냅숏을 같은 커밋에」는 기록된 파일에 대한 규칙이다.
- **그래도 고쳐 둘 가치가 있다.** 다음 소모성 재생성이 이 파일을 `COUNT 11` 로 흡수하고, sync 단계의 B12 계수도 live 식별자 11 을 낼 것이다. L35 를 `AC-ALD2-002 [REF]` 로 적으면 카운터는 `live=10 excluded=1` 을 낸다 — N7(optional). 이 수정 자체도 기록된 행이 없으므로 스냅숏 커밋을 요구하지 않는다.

## Defects Found (structured defect-list)

D1. D14-R ABSORB-FORM — acceptance.md:L318-347 — AC-ALH-008 의 `git log --no-merges 7fe658815..HEAD -- <36경로>` 는 「카드가 작성한 비병합 커밋」이 아니라 「7fe658815 에서 HEAD 로 도달 가능한 모든 비병합 커밋」을 센다. 레인 규약상 병합 전에 `git merge develop` 을 하므로(`CLAUDE.local.md §4.1`), 흡수된 다른 카드의 커밋이 18경로나 미러를 건드렸다면 이 카드가 아무것도 고치지 않았어도 FAIL 이다. 실측 양성 대조: t1175 의 흡수 병합 `8fb81c948` 에서 `git log --no-merges 8fb81c948^1..8fb81c948` 가 53행, 룰 경로로 좁히면 t1064 의 `43697af85`·`237d5e3e7` 이 나오고, 같은 범위에 `--first-parent` 를 붙이면 0행이다. 이 형식은 1회차 보고서 D14 의 Required fix 가 제시한 것으로, 감사자 제안이 틀렸다. — Severity: major — Class: blocking — Required fix: 명령을 `git log --first-parent --no-merges 7fe658815..HEAD -- <36경로> | wc -l` 로 바꾼다(레인은 develop 을 카드 브랜치로 흡수하므로 카드 계보가 제1 부모다). 대안은 병합 전 전용 `git log --no-merges "$(git merge-base develop HEAD)"..HEAD -- <36경로>` 이며, 이 경우 「병합 뒤에는 쓸 수 없다」는 한계를 AC 에 적는다. `acceptance.md:L347` 의 설명도 새 형식에 맞춘다.

D2. N1 INIT-VALUES-UNVERIFIED — acceptance.md:L142·L229, L34-35 — 한도 판정의 1차 표면 `S_init` 의 합계 `total_init` 은 「정수」만(L142), `hash_init`·`hash_live` 는 「64자리 hex 두 줄」만(L229) 검사한다. 변이: `total_init = 140000` 을 적으면 AC-ALH-002 는 통과하고, AC-ALH-006 은 같은 값을 `C` 로 받아 일관되게 통과하며, AC-ALH-007 은 `ACHIEVABLE` 을 요구해 통과한다. 1회차 D4 가 `S_live` 에서는 고쳐졌지만 판정이 실제로 걸리는 표면에서는 고쳐지지 않았다. — Severity: major — Class: blocking — Required fix: AC-ALH-002 에 「AC-ALH-009 의 하네스 명령을 `build_head` 에서 다시 돌려 `$SCRATCH/init-surface` 를 만들고, 그 18경로의 `wc -m` 합계가 `total_init` 과 같다」를 넣는다. AC-ALH-004 에 「`hash_init` = 그 내보낸 트리에서 AC-ALD2-002 파이프라인 재실행 값, `hash_live` = `git archive <build_head>` 트리에서 같은 파이프라인 값(18경로가 7fe658815 이후 불변이면 `d97b33d9…c6c3`)」을 넣는다. 함께 파일별 `Σ gross(서문 + 최상위 절 행) == wc -m` 을 AC-ALH-003 에 넣어 분할 완결성을 확립한다.

D3. N2 NET-NEGATIVE-ESCAPE — acceptance.md:L79·L187 — `REJECT` 사유 `net-negative` 는 awk 에서 조건 없이 `ok=1` 이다. 「증거 파일에 순감 산술」(L79)은 검사되지 않는다. 실측: `bind=0`·`gov=N`·`c1~c4=Y`·`mech=M1` 행 둘을 `REJECT net-negative` 로 둔 TSV 에 L169-191 awk 를 돌리면 `BAD=0`. 모든 허용 가능 행을 이렇게 기각하면 `A_adm` 이 0 이 되어 `STRUCTURALLY-INFEASIBLE-UNDER-FREEZE` 로 거짓 상신할 수 있다 — 1회차 D6 이 지적한 경로다. — Severity: major — Class: blocking — Required fix: `net-negative` 에 `$5=="M1"` 을 요구하고, 그 행의 (file, dest) 가 `pointers-<s>.tsv` 에 있고 `pointer_chars ≥ gross` 임을 awk 로 검사한다(또는 증거 파일의 `net = <음수>` 줄을 `grep -E '^net = -[0-9]+$'` 로 검사). 같은 방식으로 `dest-over-40k` 의 `c4=N` 을 `dest-sizes.tsv` 와 대조하면 더 좋다(optional).

D4. N3 RUNTIME-OBS-DESIGN — acceptance.md:L379·L398-401, spec.md:L87 — 런타임 관측의 대상과 조건이 대조의 목적을 이루지 못한다. (i) 관측됨: 하네스 계약은 **18경로만** 내보내고(L379) `t.TempDir()` 의 init 트리는 테스트 종료와 함께 사라진다. 그 사본에서 런타임을 띄우면 보고 파일 수는 구성상 18 이하이고, 실제 init 트리에 18경로 밖의 상시 로드 파일이 있는지는 원리상 드러나지 않는다. (ii) 가설(미관측): 관측자(리드·운영자)의 사용자 범위 지시문(예: `~/.claude/CLAUDE.md`)이 런타임 합계에 섞일 수 있는데, 관측 격리 요구가 없다. 선행 관측(t1184)은 격리된 홈에서 이루어졌다(`plan.md:L24`). — Severity: major — Class: blocking — Required fix: 하네스가 init 트리 전체(적어도 `CLAUDE.md`·그 `@import` 대상·`.claude/rules/**/*.md` 전부)를 내보내게 하고, 관측은 사용자 지시문이 없는 격리 `CLAUDE_CONFIG_DIR`·`HOME` 에서 하며 그 사실을 `runtime_source` 에 적게 한다. 격리 관측이 불가하면 형태 2 로 닫는다.

D5. N5 OBSERVED-SET-CONTRADICTION — acceptance.md:L34·L270-273·L399 — AC-ALH-010 형태 1 규칙은 관측 집합이 18경로와 다르면 「`T_min_init` 을 관측된 집합으로 계산」하라고 한다(L399). 그런데 `total_init` 은 「18경로 `wc -m` 합계」로 정의되고(L34), AC-ALH-006 은 `current_init == total_init` 과 `A_adm_init == Σ chars(ADMIT, 전체 TSV)` 를 요구한다(L267-273). 관측 집합 분기에서는 두 AC 를 동시에 만족할 수 없다. 형태 2 의 17경로 판정도 같은 문제를 가진다. — Severity: major — Class: blocking — Required fix: `count_set_init = 18|17|observed` 줄을 두고, `total_init`·`current_init`·`A_adm_init`·`R_init`·`U_init` 을 그 집합 기준으로 정의하며, AC-ALH-006 의 awk 가 `count_set_init` 에 해당하는 행만(`skill-routing.md` 등 제외 경로 필터) 합산하게 한다. 또는 관측 집합용 별도 줄(`*_init_obs`)을 두고 AC-ALH-006/007 이 그것을 소비하게 한다.

D6. N4 TOKEN17-UNVERIFIED — acceptance.md:L42·L395·L400, spec.md:L87 — `verdict_init_17` 은 형식만 검사되고, 17경로 집합의 `A_adm`·`R`·`U`·`T_min` 산출 규칙과 기계 줄이 없다. 위 「신규 표면 검토」의 단조성 때문에 상신 회피는 불가하지만, 17 로는 `ACHIEVABLE` 인 상황을 `STRUCTURALLY-INFEASIBLE-UNDER-FREEZE` 로 적어 운영자에게 불필요한 동결 해제를 상신할 수 있다. — Severity: minor — Class: blocking — Required fix: `T_min_init_17`·`U_init_17` 줄을 요구하고, 17집합 값은 TSV 에서 `$1 != ".claude/rules/moai/workflow/skill-routing.md"` 행만 합산해 재현한다고 규정한 뒤, AC-ALH-007 과 같은 토큰 awk 를 17집합에도 돌린다. D5 수리와 한 번에 처리할 수 있다.

D7. N6 TIER-M-REQ-BUDGET — spec.md:L72-88, spec.md:L14 — `tier: M` 인데 REQ 가 17개다. `.claude/rules/moai/workflow/spec-workflow.md § SPEC Complexity Tier` 는 Tier M 의 요구사항 상한을 16 으로 두고 「초과는 tier 상향이나 분할의 신호이지 예산 완화가 아니다」라고 적는다. 린트는 이를 기계 검사하지 않는다(0 warning). — Severity: minor — Class: blocking — Required fix: REQ 하나를 병합해 16 이하로 줄이고 번호를 다시 매긴다(예: REQ-ALH-008 의 스크립트 커밋 요구를 REQ-ALH-003 에, 또는 REQ-ALH-014 를 REQ-ALH-001 에 흡수). §D.2 추적성 표와 AC 매트릭스를 함께 고친다. 또는 `tier: L` 로 올리고 `design.md` 를 갖춘다.

D8. N7 AC-COUNT-FOREIGN-REF — acceptance.md:L35 — 선행 SPEC 식별자 `AC-ALD2-002` 가 표시 없이 있어 정본 카운터가 11 을 센다(코퍼스 게이트 실측 `COUNT 11`). 문서의 AC 는 10개다. 스냅숏 커밋 의무는 지금 없다(부재 행, 방아쇠 아님). — Severity: minor — Class: optional — Required fix: L35 를 `AC-ALD2-002 [REF]` 로 적어 카운터가 `live=10 excluded=1` 을 내게 한다.

D9. N8 HARNESS-SURFACE-SOURCE — acceptance.md:L159-160·L357-361 — AC-ALH-003 (1)의 `$SCRATCH/init-surface` 는 커밋되지 않은 임시 경로이며, 그것을 어떻게 다시 만드는지(AC-ALH-009 명령을 `build_head` 에서 실행) 적혀 있지 않다. `harness-run.txt` 에는 실행 시점 HEAD 가 기록되지 않는다. — Severity: minor — Class: optional — Required fix: AC-ALH-003 (1) 앞에 「`build_head` 체크아웃에서 AC-ALH-009 명령을 재실행해 `$SCRATCH/init-surface` 를 만든다」를 적고, run 단계가 하네스 실행 직전 `git rev-parse HEAD` 를 `harness-run.txt` 머리에 남기게 한다.

D10. N9 COMMANDS-LOG-REGEX — acceptance.md:L373-377 — `commands.log` 는 레인 자기 신고이고, L374 정규식은 `HOME=… moai init`, `moai --debug init`, `moai  init`(공백 둘) 같은 형태를 놓친다. — Severity: minor — Class: optional — Required fix: 정규식을 `moai([[:space:]]+-[^[:space:]]+)*[[:space:]]+init` 계열로 넓히고, 이 검사가 자기 신고에 기댄다는 한계를 적는다.

D11. N10 REQ-IMPL-DETAIL — spec.md:L88 — REQ-ALH-017 이 테스트 헬퍼 함수명 `prepareSafeInitHome` 을 요구사항 층에 담는다(RQ-4). 격리 보장이 요구의 실체이므로 이름은 plan/acceptance 로 내리는 편이 낫다. — Severity: minor — Class: optional — Required fix: REQ 는 「실제 홈의 네 쓰기 지점을 임시 디렉터리로 돌리고 실제 홈 불변을 검증하는 커밋된 하네스」로 쓰고 함수명은 `plan.md` D2·AC-ALH-009 에 둔다.

D12. N11 LINT-GATE — plan.md:L98 — Go 테스트 파일을 추가하는데 Go 검증을 `go vet ./internal/cli/` 로 한정한다. CI 는 golangci-lint(v2.1.6)를 돌린다. — Severity: minor — Class: optional — Required fix: M7 에 `golangci-lint run ./internal/cli/...`(CI 판)를 넣는다.

## Regression Check (Iteration 2)

Defects from previous iteration:
- D1: 요구 층 [RESOLVED] — REQ-ALH-016·AC-ALH-010 신설. 검증 설계의 새 결함은 N3·N4·N5 로 분리.
- D2: [RESOLVED] — 헬퍼 실재·동작 확인(`init_home_guard_test.go:145`, `init_deploy_exit_test.go:50`), 비격리 금지 명문화.
- D3: [RESOLVED] — `build_head` + `harness_sha256`.
- D4: [UNRESOLVED — 부분] — `S_live` 해소, `S_init` 값 미검증(N1).
- D5: [RESOLVED] — 정규식이 자리표시 거부를 실측.
- D6: [UNRESOLVED — 부분] — `net-negative` 무조건 통과(N2, 변이 `BAD=0`).
- D7–D13: [RESOLVED] — 위 표의 행 번호.
- D14: [UNRESOLVED] — 채택한 형식이 흡수 커밋을 포함(양성 대조 53행). 1회차 감사자 제안의 결함이며 manager-spec 은 제안을 그대로 반영했다. 정체(stagnation)가 아니라 잘못된 처방의 결과로 기록한다.
- D15–D20: [RESOLVED].
- D21: [RESOLVED] — 새 커밋에 트레일러. 최초 커밋의 INFO 는 수용.

## Recommendation

FAIL. 필수 통과 7개는 모두 통과했고 점수는 0.62 → 0.77 로 올랐지만(STOP 없음), 문턱 0.80 아래이며 blocking 결함 7건(D1–D7)이 남았다. 1회차에서 이어진 미해소 3건(D14·D4·D6)은 규칙상 자동 FAIL 이다.

이번이 Tier M 상한 회차(2/2)이므로 오케스트레이터는 사용자에게 에스컬레이션한다(PASS-with-debt / 범위 축소 / 명시적 상한 연장). 판단 자료:

1. 남은 blocking 결함은 모두 AC 명령과 기계 줄 규약의 국소 수정이며 요구의 방향을 바꾸지 않는다. D1(`--first-parent` 한 단어), D7(REQ 하나 병합)은 기계적이다.
2. 무게가 큰 것은 D2(`total_init` 재현)와 D4(관측 대상·격리)다. 둘 다 1차 판정 표면의 수치를 확립하는 문제라서, 고치지 않은 채 run 에 들어가면 판정 토큰이 검증 불가능한 값 위에 선다.
3. PASS-with-debt 를 택한다면 D1–D7 을 run 단계 M1 착수 전 선행 조건으로 두고, 그 수정이 manager-spec 의 SPEC 개정(소유권: spec/acceptance 본문은 manager-spec)으로 이루어져야 한다.

수정 지시(manager-spec):

1. AC-ALH-008 을 `git log --first-parent --no-merges 7fe658815..HEAD -- <36경로>` 로 바꾼다(D1).
2. `total_init`·`hash_init`·`hash_live` 를 하네스 재실행과 파이프라인 재실행으로 재현하는 검사, 파일별 `Σ gross == wc -m` 검사를 넣는다(D2).
3. `net-negative` 에 M1 + 포인터 표 대조 조건을 붙인다(D3).
4. 하네스가 init 트리 전체를 내보내고 관측을 격리 설정에서 하도록 규정한다(D4).
5. `count_set_init` 줄로 집합 기준을 한 곳에서 정하고, AC-ALH-006/007 과 17집합 재계산이 그것을 소비하게 한다(D5·D6).
6. REQ 를 16개 이하로 줄인다(D7).

optional D8–D12 는 오케스트레이터 재량이다. D8(`[REF]` 표시)은 한 토큰이므로 D1–D7 수정과 함께 처리하는 편이 싸다.
