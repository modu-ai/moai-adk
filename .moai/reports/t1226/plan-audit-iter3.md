auditor-model: claude-opus-5-5

# SPEC Review Report: SPEC-ALWAYS-LOADED-HEADROOM-001
Iteration: 3/3 (최종 회차. `harness.yaml` `plan_audit_tier_ceilings.M` 은 2 이며, 이번 회차는 오케스트레이터가 명시한 상한 연장 `max_iterations 3` 아래에서 수행했다)
Verdict: PASS-WITH-DEBT
Overall Score: 0.86 (1회차 0.62 → 2회차 0.77 → 3회차 0.86. 상승 추세라 STOP 신호 없음. Tier M 문턱 0.80 이상)

카드 t1226 · 감사 대상 커밋 `7ae1a3b86` (spec 0.3.0) · 브랜치 `WT-always-loaded-headroom` · 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1226`

Reasoning context ignored per M1 Context Isolation. 수리 커밋 메시지, spec HISTORY 0.3.0 행, `progress.md` 대응표는 주장으로만 읽고 전부 재측정했다.

교차 모델 감사는 호출하지 않았다. `.moai/config/sections/` 에 `audit_model` 키가 없다(`grep -rn audit_model` 0행). 이전 두 회차와 같은 Claude 단독 감사다.

깨끗한 PASS 를 주지 않는 이유는 하나다. 판정의 핵심 입력인 `ADMIT` 행의 `chars` 값에 상한이 없어, AC 전부를 통과하면서 `A_adm` 을 부풀려 거짓 `ACHIEVABLE` 을 낼 수 있다(아래 N1, 변이 실측). 이 결함은 2회차 결함 목록 밖에 있던 것으로, 2회차 감사가 놓쳤다. 고치는 데 awk 조건 몇 줄이면 되므로 채무로 넘기되, run 단계 M5(`T_min` 산출) 판정 전에 반드시 해소해야 한다.

## 감사자 재측정 (이 실행, 이 트리 `7ae1a3b86`)

| 항목 | 명령 | 관측 |
|---|---|---|
| HEAD·브랜치 | `git rev-parse HEAD` · `git branch --show-current` | `7ae1a3b8602c…` · `WT-always-loaded-headroom`, 작업 트리 깨끗함 |
| `develop` 이동 | `git rev-parse develop origin/develop` | 둘 다 `7fe658815eb0…` — 기준 커밋과 동일 |
| 2회차 이후 변경 | `git diff --stat b231487ab 7ae1a3b86` | SPEC 4파일 + 판정서 뼈대 1줄 + 2회차 보고서. `research.md` 는 바뀌지 않음 |
| REQ 정의 수 | `grep -cE '^- \*\*REQ-ALH-[0-9]{3}\*\*' spec.md` | `16` (001~016, 빈틈·중복 없음) |
| 옛 번호 잔존 | `grep -rn 'REQ-ALH-017'` | spec.md L25(0.2.0 HISTORY 행), progress.md L23(「0.2.0 당시 번호」 표) 두 곳뿐 — 둘 다 이력 표기 |
| clarification 마커 | `grep -rn 'NEEDS CLARIFICATION' <SPEC 디렉터리>` | 0행(`exit=1`) |
| `syscall` | `grep -c syscall` 5파일 | 전부 0 |
| D7 참조 | SPEC ID 추출 → `grep '^status:'` | DIET-001·DIET-002 `completed`, HEADROOM-001 자신 `draft` |
| D1 현 HEAD | AC-ALH-008 명령 원문(36경로) | `0` |
| D1 음성 대조 | `git log --no-merges 8fb81c948^1..8fb81c948 \| wc -l` / 같은 범위 `--first-parent` | `53` / `0` — 2회차 값 재현 |
| D1 양성 대조 | t1175 분기점 `b59a5d69c`(= `git merge-base 8fb81c948^1 8fb81c948^2`)에서 `git log --first-parent --no-merges b59a5d69c..8fb81c948 -- .claude/rules/moai internal/template/templates/.claude/rules/moai` | t1175 자신의 커밋 5건(`8e50ef148`·`f0893dc36`·`0acfa28e1`·`276391646`·`7bcdce760`). `--first-parent` 없이 같은 명령은 7건 — 흡수된 t1064 의 `43697af85`·`237d5e3e7` 이 더해진다 |
| `total_live` git archive 재현 | AC-ALH-002 L147 형식 그대로(`BH`=HEAD): `git archive HEAD CLAUDE.md AGENTS.md .moai/config/sections .claude/rules/moai \| tar -x -C $SCRATCH/live`, 이어 `xargs wc -m < count-set-live.txt \| tail -1` | 보관본 `199111 total`, 워킹 트리 `199111 total` — 일치. `.gitattributes` 가 `*.md`·`*.yaml` 을 `eol=lf` 로 고정해 보관본과 워킹 트리가 갈릴 여지도 없다 |
| `locale charmap` | `locale charmap` | `UTF-8` |
| `hash_live` 재현 | AC-ALH-004 L241-242 파이프라인을 보관본 트리에서 실행 | `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3`, 구속 줄 `170` — 선행 SPEC 기준선과 동일 |
| 분할 완결성·`¶` 제외 | AC-ALH-003 (2) awk 원문을 합성 트리(다중 바이트 `¶`·`é` 포함 본문, `¶1` 문단 행, yaml `(전체)` 행, `paths:` frontmatter 서문)에 BSD awk 로 실행 | `SPLIT-init-OK`. `C=104` = 세 파일 `wc -m` 합 104 |
| `net-negative` 변이 | 2회차 변이(`M1`·`bind=0`·`gov=N`·`c1~c4=Y`·`REJECT net-negative`)에 증거 `pointer_chars = 5`(< gross 20) 한 행, 증거에 `pointer_chars` 줄이 없는 한 행 | `ROWS BAD=0` 이지만 `NETNEG-init BAD=2` — 2회차 `BAD=0` 에서 바뀜. 게이트 FAIL |
| `count_set_init` 소비 | AC-ALH-006 `calc` 원문을 합성 TSV 에 제외 경로 없음 / `skill-routing.md` 제외로 실행 | `C=104`→`74`, 제외 파일 행이 `A`·`U`·`R`·`C` 모두에서 빠짐 |
| 17집합 재계산 | AC-ALH-006 L306, AC-ALH-007 L330-341 읽기 + 위 `calc` | `init_17` 접미사 줄 6개 + `verdict_init_17` 을 재계산 대상으로 소비 |
| `commands.log` 정규식 | L416 정규식을 10줄 표본에 실행 | `HOME=/x moai init`, `moai --debug init`, `moai  init`, `$SCRATCH/moai init`, `~/go/bin/moai init` 5건 적중. `moai initx`, `moai update`, `moai spec lint`, 정규식 자신을 담은 `grep` 줄은 비적중 |
| AC 계수 | `manager-docs.md` 의 `MOAI-AC-COUNTER` 본문을 그대로 실행 | stdout `10`, stderr `live=10 excluded=1 ambiguous=0`, `exit=0` |
| AC 코퍼스 게이트 | `go test ./internal/spec -run TestACCounterFullCorpusMatchesBaseline -count=1 -v` | `exit=0`, `--- PASS (6.90s)`. `absent-from-snapshot …HEADROOM-001/acceptance.md: COUNT 10` — 2회차 `COUNT 11` 에서 10 으로 |
| 하네스 헬퍼 | `grep -n 'func prepareSafeInitHome\|func runInitWithFlags\|func TestMain'` | `init_home_guard_test.go:145`, `init_deploy_exit_test.go:50`, `main_test.go:349`. `internal/cli/*headroom*` 는 아직 없음(run 단계 산출물) |
| slim 설치 | `init.go:446-452` `shouldDistributeAll` | `--all` 또는 환경변수 `MOAI_DISTRIBUTE_ALL=1/true` 이면 전체 배포. `catalog.yaml` 에 `rules` 0행 — 룰은 slim 선별 대상이 아님 |
| 린트 | `go build -o $SCRATCH/moai ./cmd/moai` → `$SCRATCH/moai spec lint --strict SPEC-ALWAYS-LOADED-HEADROOM-001; echo exit=$?` | `build exit=0` / `0 error(s), 0 warning(s)` / `lint exit=0`. INFO 1건 `OwnershipTransitionUnmeasured` — 대상은 최초 커밋 `10281a857`(2회차와 동일, 수용) |
| 수리 커밋 트레일러 | `git log -1 --format=%B 7ae1a3b86` | `Authored-By-Agent: manager-spec` 있음 |

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: `spec.md:L75-90` 에 REQ-ALH-001~016, 빈틈·중복 없음, 3자리 패딩.
- [PASS] MP-2 GEARS 준수(요구사항 층 `spec.md §C` 대상): 16개 모두 `shall`/`shall not` 을 가진 Ubiquitous·Event·Where·Unwanted 형태. 병합된 REQ-ALH-001(L75)은 긍정문과 금지문을 한 요구에 담은 Ubiquitous, REQ-ALH-013(L87)은 Unwanted, REQ-ALH-015(L89)는 조건절을 품은 Ubiquitous 로 2회차와 같은 수준에서 허용. AC 는 검증 층이라 여기서 채점하지 않았다.
- [PASS] MP-3 frontmatter: `spec.md:L2-14` — 12개 필드 모두 있음, `version: "0.3.0"` 따옴표, `tier: M`, 거부 별칭 없음.
- [N/A] MP-4 언어 중립성: 다언어 도구 SPEC 이 아니다. 추가될 Go 테스트는 저장소 내부 하네스이며 템플릿 경로가 아니다.
- [PASS] MP-5 D7: 참조 SPEC 둘 다 `completed`. BLOCKING 0.
- [PASS] MP-6 D8: `syscall` 0회.
- [PASS] MP-7 clarification gate: `plan.md`·`research.md` 에 마커 0행.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75–1.0 사이 | 2회차의 `count_set_init` 모순(N5)과 17집합 산출법(N4)은 해소(`acceptance.md:L34`, L294-306). 남은 것은 국소적이다: 절 분할의 제목 수준 미정(L54), `BH` 에서 하네스를 다시 돌리는 절차 미기재(L113), §D.3 첫 항목과 L156 의 줄 수 규칙 충돌, `research.md:L49` 의 옛 번호(N7) |
| Completeness | 0.95 | 0.75–1.0 사이 | 모든 절 존재, `### Out of Scope — …` 5개(`spec.md:L96-116`), REQ 16개로 Tier M 상한 안 |
| Testability | 0.70 | 0.50–0.75 사이 | 2회차 blocking 4건(흡수 형식, `S_init` 값, `net-negative`, 17집합)은 모두 기계 검사로 닫혔다(위 재측정). 그러나 `ADMIT` 행 `chars` 가 무상한이고 M2 `chars` 가 증거와 묶이지 않았다 — 변이 두 개가 모든 게이트를 통과(N1). 절 행과 문단 행의 이중 계상(N2), 계수 대상 파일을 목적지로 삼는 M1(N3)도 기계 검사가 없다 |
| Traceability | 1.00 | 1.00 | `acceptance.md:L11-22` 매트릭스와 `L461-478` §D.2 일치. REQ 16개 모두 AC 보유, AC 10개 모두 실재 REQ 를 가리킴. 병합된 폐기 수치 격리는 AC-ALH-001 L102-103 이 검사 |

총점은 네 차원의 조화평균 0.859 를 반올림했다.

## 2회차 결함 판정 (D1–D12)

| 결함 | 판정 | 근거 |
|---|---|---|
| D1 흡수 형식 (blocking) | **해소** | `acceptance.md:L359` 가 `--first-parent --no-merges`. 현 HEAD 0행. 음성 대조 53→0, 양성 대조에서 카드 자신의 커밋 5건은 잡고 흡수 커밋 2건은 뺀다(재측정 표). REQ-ALH-013(`spec.md:L87`)도 「카드 계보의 비병합 커밋」으로 바뀌었고, 흡수 방향이 뒤집히면 성립하지 않는다는 한계가 L387 에 적혔다. 잔여: 흡수 병합 커밋 **안에서** 계수 파일을 고치는 경우는 `--no-merges` 가 보지 못한다(N6, optional) |
| D2 `S_init` 값 미검증 (blocking) | **해소** | `total_init` 은 하네스를 `BH` 에서 재실행한 `$SCRATCH/init-surface` 로 재현(L113, L150), `total_live` 는 `git archive BH` 로 재현(L147-148, 감사자 실측 199111 일치), `hash_init`·`hash_live` 는 두 트리에서 파이프라인 재실행(L239-243, 감사자 실측 `d97b33d9…c6c3` 일치). 파일별 `Σ gross == wc -m`(L174-180) 을 BSD awk·다중 바이트 본문에서 실측 통과. 잔여: `BH ≠ HEAD` 일 때 `BH` 트리에서 하네스를 돌리는 방법이 적혀 있지 않다(N8, optional) |
| D3 `net-negative` 우회 (blocking) | **해소** | L81 조건표와 L203 awk 가 `M1`·`bind=0`·`gov=N`·`c1~c4=Y`·`dest≠-` 를 요구하고, L209-212 가 증거의 `pointer_chars ≥ gross` 를 검사한다. 2회차 변이 `BAD=0` → 이번 `NETNEG BAD=2`. 처방대로의 수리다. 잔여: `pointer_chars` 는 증거의 자기 신고 수치이며 포인터 원문과 기계적으로 묶이지 않는다(N5, optional) |
| D4 관측 설계 (blocking) | **해소** | 하네스가 `.git` 을 뺀 init 프로젝트 트리 전체를 내보낸다(L419, `plan.md:L33`). 관측은 사용자 범위 지시문이 없는 격리 `CLAUDE_CONFIG_DIR`·`HOME` 에서, `runtime_isolated = yes` 없이는 형태 1 불인정(L428, L435, L445). 사용자 지시문 혼입은 관측되지 않은 가설로 표기됐다(`spec.md:L51`). 잔여: 사본 디렉터리의 상위 경로에 있는 `CLAUDE.md` 도 런타임이 읽을 수 있는데 위치 제약이 없다(N9, optional) |
| D5 집합 모순 (blocking) | **해소** | `count_set_init` 한 줄과 `count-set-<s>.txt`(L34). TSV 는 계수 집합 파일만 담고(L48), 키 재생성이 `--paths count-set-$s.txt` 로 돌며(L170), 분할 대조(L177-179)가 TSV 파일 집합과 계수 집합을 파일 단위로 맞댄다. AC-ALH-006 이 그 행만으로 `C`·`A`·`U`·`R` 을 재계산하고 `total == current == C` 를 요구(L299-302). 합성 실측에서 제외 경로가 네 항 모두에서 빠진다 |
| D6 17집합 미검증 (blocking) | **해소** | `total/current/A_adm/R/U/T_min_init_17`·`verdict_init_17` 줄 요구(L43, L440-441), `calc init "$SR" init_17`(L306), `TOKEN-init_17` 재계산(L330-333), 두 토큰이 다르면 `UNDETERMINED` + `verdict_init_reason = count-set`(L335-341) |
| D7 REQ 예산 (blocking) | **해소** | REQ 16개. 옛 014 를 001 에 병합(`spec.md:L75`), 옛 015~017 → 014~016. 매트릭스(L11-22)·§D.2(L461-478)·`plan.md` 참조(L9, L73, L97-98) 모두 새 번호. 잔여: `research.md:L49` 가 조건 4 를 옛 번호 `REQ-ALH-015` 로 가리킨다(N7, optional) |
| D8 외부 AC 참조 (optional) | 해소 | L36·L238 `AC-ALD2-002 [REF]`. 카운터 `live=10 excluded=1`, 코퍼스 게이트 `COUNT 10` |
| D9 하네스 재실행·HEAD 귀속 (optional) | 해소 | `harness-run.txt` 첫 줄 `harness_head = <sha>`(L397), `build_head` 와 대조(L410). 재실행 지시(L113) |
| D10 정규식 (optional) | 해소 | L416 확장 정규식이 표본 5형태 적중·오탐 0. 자기 신고 한계를 L421 에 명시 |
| D11 REQ 구현 세부 (optional) | 해소 | REQ-ALH-016(`spec.md:L90`)에 함수명 없음. 이름은 `plan.md:L33`·AC-ALH-009 에만 |
| D12 린트 게이트 (optional) | 해소 | `plan.md:L91`, `plan.md:L99`, `acceptance.md:L456` 에 CI 판 `golangci-lint run ./internal/cli/...` |

## 신규 표면 검토 (v0.3.0 이 더한 것)

- **`gross` 평면 분할과 문단 행 제외.** 정의(L54)와 검사(L174-180)가 맞물린다. 문단 행은 분할 합에서도(L177), `C` 에서도(L299) 빠지므로 `total` 이중 계상은 없다. 다만 문단 행의 **허용 자수**는 `A` 에 그대로 들어가고, 「절 행과 문단 행 중 한 곳에만 둔다」(L483)는 규칙을 검사하는 줄이 없다 — N2. 「절」이 어느 제목 수준에서 끊기는지는 정해져 있지 않지만, 키 재생성과 분할 대조가 스크립트와의 자기 일관성을 요구하므로 검증 가능성에는 영향이 없다(분할 세밀도만 달라진다).
- **프로젝트 트리 전체 내보내기.** 계약(L419)이 `.git` 만 뺀 전체 복사를 요구하고, 형태 1 에서 18경로 밖 파일도 TSV 행을 갖게 한다(L445). `internal/cli` 의 `TestMain` 잔여물 가드(`main_test.go` — 패키지 작업 디렉터리에 `.moai` 가 생기면 FAIL)는 내보내기 대상이 `$SCRATCH` 이므로 충돌하지 않는다. 환경변수 `MOAI_DISTRIBUTE_ALL` 이 켜져 있으면 `--all` 없이도 전체 배포가 되는데, 하네스 계약이 이 변수를 비우라고 요구하지 않는다 — N10(optional).
- **`runtime_isolated`.** 형태 1 의 필수 줄이며 형식 검사가 있다(L435). 격리 여부 자체는 자기 신고이고 `runtime_source` 서술로 검토된다. 상위 디렉터리 `CLAUDE.md` 경로는 격리 조건에 빠져 있다 — N9.
- **`total_live` 의 git archive 재현.** L147-148 을 그대로 실행해 `199111` 을 얻었다. 워크트리 가드 아래에서도 이 형식(`git archive … | tar -x -C …` 와 `( cd … && xargs wc -m < "$OLDPWD/…" )`)은 거부되지 않았다. 다만 `count-set-live.txt` 의 **내용**은 줄 수(18)만 검사되고 18경로 목록과 대조되지 않는다 — N4.
- **`pointer_chars` 검사.** 변이로 발화를 확인했다. 수치는 증거 파일의 자기 신고다 — N5.
- **`harness_head`.** `head -1 | awk '{print $3}'` 이 `harness_head = <sha>` 의 세 번째 필드를 뽑아 `build_head` 와 대조한다. 하네스 파일이 측정 뒤 고쳐지면(예: M7 golangci-lint 수정) `harness_sha256` 이 달라져 FAIL 하므로, 수정 뒤에는 재측정이 강제된다 — 의도대로 엄격하다.

## Defects Found (structured defect-list)

N1. CHARS-UNBOUNDED — acceptance.md:L55·L188·L259, spec.md:L80 — `ADMIT` 행의 `chars` 는 「정수」만 검사되고(L188 `$4 !~ /^[0-9]+$/`) `gross` 이하라는 상한이 없다. M2 행의 `chars` 는 REQ-ALH-006 이 「실제 압축 시도로 잰다」고 요구하지만 AC-ALH-004 는 `post_hash` 만 읽고(L252, L259) 줄어든 자수는 증거와 대조하지 않는다. 합성 변이 실측: M1 `ADMIT` 행의 `chars` 를 `gross` 의 두 배로 두면 `ROWS BAD=0`·`NETNEG BAD=0`·`DEST BAD=0`·`SPLIT OK` 로 모두 통과하면서 `A` 가 20→40 이 되고, M2 `ADMIT` 문단 행(`gross` 16)의 `chars` 를 5000 으로 두면 역시 전부 통과하면서 `T_min` 이 음수가 된다. AC-ALH-006 은 판정서 줄이 이 부풀린 합과 일치하는지만 보므로 `CHECK=OK`, AC-ALH-007 은 `ACHIEVABLE` 을 요구해 `TOKEN=OK` 다. 결과는 상신 절차를 요구하지 않는 거짓 `ACHIEVABLE` 이다 — 상신 회피 방향이다. (M1 의 999999 같은 극단값은 목적지 수용량 검사 L219-222 에 걸리지만, 수용량 안의 부풀림과 M2 부풀림은 걸리지 않는다.) 2회차 결함 목록 밖의 항목이며 2회차 감사가 놓쳤다. — Severity: major — Class: blocking(채무로 이월, 아래 조건) — Required fix: AC-ALH-003 (3) awk 에 `if ($13=="ADMIT" && $4+0 > $3+0) bad++;` 와 `if ($13=="ADMIT" && $5 ~ /^M1/ && $4+0 != $3+0) bad++;` 를 넣는다(M1·M1′ 는 후보 전체를 옮기거나 지우므로 `chars == gross`). M2 는 증거 파일에 `pre_chars = N`·`post_chars = M` 줄을 요구하고 `N == gross` 이며 `chars == N − M` 인지 AC-ALH-004 의 증거 루프에서 검사한다. `M1′` 을 쓰는 경우 증거에 남는 중복 원본 경로를 적게 한다.

N2. SECTION-PARAGRAPH-DOUBLE-COUNT — acceptance.md:L53·L483·L296 — §D.3 은 「허용 자수는 절 행과 문단 행 중 한 곳에만 둔다」고 규정하지만 이를 검사하는 줄이 없다. 합성 변이 실측: 절 행 `## B`(M2 `ADMIT`, `chars = gross`)와 그 문단 행 `## B ¶1`(M2 `ADMIT`)을 함께 두면 모든 게이트가 통과하고 `A` 가 같은 문자를 두 번 센다(`A=69`, 그 파일의 비구속 자수를 넘음). — Severity: minor — Class: blocking(채무로 이월) — Required fix: AC-ALH-003 (3) 뒤에 다음 검사를 넣고 `OVERLAP BAD=0` 을 요구한다 — `awk -F'\t' 'NR>1 && $13=="ADMIT" {b=$2; isp=sub(/ ¶[0-9]+$/,"",b); k=$1 SUBSEP b; if (isp) p[k]=1; else s[k]=1} END {for (k in s) if (k in p) bad++; print "OVERLAP BAD="bad+0}' <TSV>`. 감사자가 BSD awk 로 실측했다: 이중 계상 변이 `OVERLAP BAD=1`, 정상 표 `OVERLAP BAD=0`.

N3. DEST-IN-COUNT-SET — acceptance.md:L63·L189·L219-223 — M1 목적지가 계수 집합 안의 파일이면 옮긴 문자가 합계에 그대로 남는데도 `A` 에 계상된다. 18경로 가운데 `workflow/skill-routing.md` 는 `paths:` 를 가진 룰이라 조건 3 을 형식상 통과할 수 있고, `count_set_init = 18`(또는 `live`)에서는 계수 대상이다. 합성 변이(`dest = workflow/skill-routing.md`)가 모든 게이트를 통과했다. — Severity: minor — Class: blocking(채무로 이월) — Required fix: AC-ALH-003 (3) 에 「`ADMIT` M1 행의 `.claude/rules/moai/<dest>` 가 `count-set-<s>.txt` 에 없다」 검사를 넣는다. 17집합 재계산에서는 이 행을 제외 파일과 같은 방식으로 다룰지 규칙을 한 줄 적는다.

N4. COUNT-SET-LIVE-UNPINNED — acceptance.md:L34·L145·L148 — `count-set-live.txt` 는 「18경로 고정」(L34, REQ-ALH-002)인데 검사는 줄 수 18 뿐이다(L145). 큰 파일 하나를 다른 경로로 바꿔 적어도 L148 재현값·AC-ALH-003·006 이 모두 그 목록 기준으로 자기 일관성을 유지한다. 같은 절에 18경로 `wc -m` 블록(L118-136)이 있어 사람이 그 블록을 돌리면 드러나지만, 판정 명령이 목록 파일을 쓰므로 기계 검사로는 막히지 않는다. `count_set_init = 18` 의 `count-set-init.txt` 도 같다. — Severity: minor — Class: optional — Required fix: L145 를 18경로 리터럴 목록과의 `diff` 로 바꾸고(`printf '%s\n' <18경로> | diff - .moai/reports/t1226/count-set-live.txt`), `count_set_init = 18` 에도 같은 대조를 두되 §D.3 의 누락 경로 예외만 허용한다.

N5. POINTER-CHARS-SELF-REPORT — acceptance.md:L81·L211·L314 — `net-negative` 기각은 증거 파일의 `pointer_chars = P` 수치만 본다. 포인터 원문과 그 `wc -m` 의 일치는 AC-ALH-006 의 검토 항목(L314)이며, 그것도 `ADMIT` 행의 `pointers-<s>.tsv` 에 한정돼 `net-negative` 행에는 걸리지 않는다. 2회차 처방(증거 파일 수치 검사)대로의 수리이므로 D3 은 해소로 판정했고, 이것은 그 처방의 잔여 한계다. — Severity: minor — Class: optional — Required fix: 증거 파일에 `pointer: <원문 한 줄>` 을 함께 요구하고 `P` 가 그 줄의 문자 수(개행 포함)와 같은지 검사한다.

N6. EVIL-MERGE-BLIND — acceptance.md:L359 — `--no-merges` 는 흡수 병합 커밋 **안에서** 계수 파일을 고친 경우(충돌 해소나 수동 편집)를 보지 못한다. 레인 규약상 흡수는 `git merge develop` 이므로 발생 가능하나 드물다. — Severity: minor — Class: optional — Required fix: 흡수 병합마다 `git diff --quiet <merge>^2 <merge> -- <36경로>` 가 0 인지(병합 결과가 흡수 대상 쪽과 같은지) 보는 한 줄을 덧붙이거나, 이 한계를 L387 의 한계 문단에 적는다.

N7. RESEARCH-STALE-REQ — research.md:L49 — 「조건 4 를 두 트리에서 판정하는 것(REQ-ALH-015)」. 0.3.0 에서 조건 4 는 REQ-ALH-014 이고 REQ-ALH-015 는 계수 집합 요구다. `research.md` 는 0.3.0 수리에서 바뀌지 않았다. — Severity: minor — Class: optional — Required fix: `REQ-ALH-014` 로 고친다.

N8. BH-RERUN-PROCEDURE — acceptance.md:L113·L149 — 「하네스를 `BH` 에서 다시 돌린다」가 `BH ≠ HEAD` 인 경우의 방법을 적지 않는다. 워크트리 세션은 브랜치를 바꿀 수 없고 bare `git worktree add` 도 금지다. — Severity: minor — Class: optional — Required fix: 「`git diff --quiet BH HEAD -- internal/ cmd/ pkg/ go.mod go.sum` 가 0 이면 HEAD 에서 재실행해도 된다, 아니면 `git archive BH` 로 모듈 전체를 `$SCRATCH` 에 풀어 그 안에서 실행한다」를 적는다.

N9. ANCESTOR-INSTRUCTIONS — acceptance.md:L428·L447 — 격리 조건이 `CLAUDE_CONFIG_DIR`·`HOME` 뿐이다. 런타임은 작업 디렉터리의 상위 경로에 있는 `CLAUDE.md` 도 읽을 수 있으므로(관측되지 않은 가설), 사본을 저장소 안에 두면 관측 합계가 오염될 수 있다. — Severity: minor — Class: optional — Required fix: 관측 방법에 「사본은 상위 경로에 `CLAUDE.md`·`.claude/` 가 없는 디렉터리(예: `$SCRATCH`)에 둔다」를 넣고 `runtime_source` 에 사본 경로를 적게 한다.

N10. DISTRIBUTE-ALL-ENV — acceptance.md:L419 — 고정 플래그는 `--all` 을 빼지만 `internal/cli/init.go:446-452` 는 환경변수 `MOAI_DISTRIBUTE_ALL=1/true` 로도 전체 배포를 켠다. 하네스 계약이 이 변수를 비우라고 요구하지 않는다. 룰 파일은 slim 선별 대상이 아니므로(`catalog.yaml` 에 `rules` 0행) 18경로에는 영향이 없을 가능성이 높지만, 형태 1 의 관측 집합(트리 전체)에는 영향을 줄 수 있다. — Severity: minor — Class: optional — Required fix: 계약에 `t.Setenv("MOAI_DISTRIBUTE_ALL", "")` 를 넣는다.

N11. MISSING-PATH-LINECOUNT — acceptance.md:L156 vs L482 — L156 은 `count_set_init = 18` 이면 `count-set-init.txt` 가 18줄이라고 하고, §D.3 첫 항목은 누락 경로를 빼되 `18` 표기를 유지하라고 한다. 누락이 생기면 둘을 함께 만족할 수 없다. 룰이 slim 선별 대상이 아니어서 실제로 걸릴 가능성은 낮다. — Severity: minor — Class: optional — Required fix: L156 에 「§D.3 누락 경로가 있으면 18 에서 누락 수를 뺀 값」을 덧붙인다.

## Regression Check (Iteration 3)

2회차 결함:
- D1: [RESOLVED] — `--first-parent` 형식, 음성·양성 대조 실측(재측정 표).
- D2: [RESOLVED] — 하네스·파이프라인 재실행 재현, `Σ gross == wc -m` 실측 통과.
- D3: [RESOLVED] — 변이 `BAD=0` → `NETNEG BAD=2`.
- D4: [RESOLVED] — 트리 전체 내보내기, 격리 관측, `runtime_isolated`.
- D5: [RESOLVED] — `count_set_init` 과 계수 집합 행 필터, `calc` 실측.
- D6: [RESOLVED] — 17집합 줄과 두 재계산.
- D7: [RESOLVED] — REQ 16개, 참조 갱신(잔여 N7 은 `research.md` 한 곳).
- D8–D12: [RESOLVED].

1회차 결함 재발 점검: D4·D6·D14(1회차 번호) 계열은 모두 이번 회차에 닫혔다. 세 회차 내내 변하지 않은 결함은 없다 — 정체 신호 없음.

## Recommendation

PASS-WITH-DEBT. 필수 통과 7개 모두 통과, 2회차 blocking 7건과 optional 5건 모두 해소(재측정 증거 위 표), 점수 0.86 으로 Tier M 문턱 0.80 이상이며 세 회차 연속 상승했다. 깨끗한 PASS 를 주지 않는 이유는 N1 하나다 — 판정 토큰의 입력인 `A_adm` 을 AC 를 모두 통과하면서 부풀릴 수 있다.

run 단계가 떠안는 채무(명시 목록):

1. **DEBT-1 (N1, major, 필수)** — `ADMIT` `chars` 상한(`chars ≤ gross`, M1·M1′ 는 `chars == gross`)과 M2 `chars == pre_chars − post_chars`(증거 줄 대조). **해소 시점: run 단계 M5 에서 `T_min` 을 적기 전.** 해소 방법은 manager-spec 의 `acceptance.md` 개정(AC-ALH-003 (3)·AC-ALH-004)이며, 개정 전에 run 을 끝내야 한다면 레인이 같은 검사를 `ac-verify.md` 에 추가 실행하고 sync-auditor 가 그 출력을 읽는다.
2. **DEBT-2 (N2, minor, 필수)** — 같은 절의 절 행과 문단 행이 동시에 `ADMIT` 인 경우를 FAIL 로 잡는 검사. 시점·방법은 DEBT-1 과 같다.
3. **DEBT-3 (N3, minor, 필수)** — M1 목적지가 계수 집합에 속하면 FAIL. 시점·방법은 DEBT-1 과 같다.
4. **DEBT-4 (N4–N11, optional)** — 계수 집합 목록 고정, `pointer` 원문 요구, 흡수 병합 내부 편집 한계, `research.md:L49` 번호, `BH` 재실행 절차, 상위 경로 지시문, `MOAI_DISTRIBUTE_ALL`, 누락 경로 줄 수. 오케스트레이터 재량. DEBT-1~3 개정 때 함께 처리하면 싸다.

acceptance.md 를 개정해도 AC 스냅숏 커밋 의무는 생기지 않는다 — 이 파일은 스냅숏에 기록돼 있지 않고(`absent-from-snapshot … COUNT 10`), 게이트는 부재 행을 보고만 한다. 새 AC 식별자를 만들지 않고 기존 AC 안에 검사 줄을 더하는 한 `COUNT` 도 10 으로 유지된다.

Implementation Kickoff Approval 은 이 판정과 무관하게 필수다. 이 PASS-WITH-DEBT 는 킥오프 승인을 대신하지 않는다.

## 검증 한계 (Gaps · Residual-risk)

- **Gaps**: 하네스 테스트 `TestHeadroomInitSurfaceExport` 는 아직 없어 `S_init` 재현 경로(L113, L150)와 AC-ALH-009 명령은 실행하지 못했다 — 실행 가능성은 기존 헬퍼·`TestMain` 구조에서 추론했을 뿐이다. 런타임 계수 관측(AC-ALH-010 형태 1)은 이 감사에서 시도하지 않았다. `candidates.py` 가 없으므로 AC-ALH-003 (1) 키 재생성은 형식만 읽었다. 흡수·변이 대조 중 「카드 계보 커밋이 계수 파일을 고친 경우」는 실제 이력(t1175 계보의 5건)으로 확인했고, 합성 저장소에서의 재현은 워크트리 가드가 격리 밖 git 호출을 거부해 하지 못했다.
- **Residual-risk**: `gov`·`c2`·`c3` 판정과 `runtime_isolated` 는 사람 판단·자기 신고에 기대며 검토 항목으로만 걸린다. 격리 `CLAUDE_CONFIG_DIR` 에서 런타임을 띄우려면 그 설정 디렉터리에서 인증이 필요할 수 있어 형태 1 관측 자체가 실패할 수 있다 — 그 경우 형태 2 로 닫히는 경로는 마련돼 있다.
