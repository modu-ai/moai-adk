# SPEC-ALWAYS-LOADED-HEADROOM-001 — 진행 기록

카드 t1226 · Tier M · 워크트리 `.claude/worktrees/t1226` · 브랜치 `WT-always-loaded-headroom` · 기준 커밋 `7fe658815`

---

## §E.1 Plan-phase Audit-Ready Signal

- plan_complete_at: 2026-09-27
- plan_status: audit-ready
- spec 버전: 0.4.0 (plan-audit 3회차 PASS-WITH-DEBT 0.86 의 필수 채무 흡수본). 요구사항 16개(Tier M 상한), AC 10개(ID 불변)
- 산출물: `spec.md` · `plan.md` · `acceptance.md` · `research.md` · `progress.md`, 판정서 뼈대 `.moai/reports/t1226/verdict.md`, `.moai/reports/t1226/sec.py`(sha256 `d0e61541367abb06a170bd36b6376e51d51882380ca9f899f0d2934016a78547`, 원본과 일치)
- SPEC ID 정규식 검사(Bash 실행): `SPEC-ALWAYS-LOADED-HEADROOM-001` → `PASS`. 중복 확인 `ls .moai/specs | grep -c HEADROOM` → `0`(작성 전)
- 기준선(오케스트레이터 실측, 이 트리): 18파일 `wc -m` → `199111 total`, 잔여 49,111
- 동결 다중집합 재실행(manager-spec, 이 트리): `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3`
- 결정 지점: D1 판정 표면 = `S_init` 1차 + `S_live` 병기. D2 `S_init` 산출 = 커밋된 격리 하네스 테스트. D3 표면별 해시 실측. D4 런타임 계수 관측 요청, 불가 시 18·17 이중 판정. 해소가 필요한 clarification 마커 0건.

### plan-audit 1회차 결함 대응표 (번호는 0.2.0 당시 REQ 번호)

| 결함 | 반영 |
|---|---|
| D1 계수 집합 | REQ-ALH-016 · AC-ALH-010 신설. `skill-routing.md` 의 `paths:` 존중 여부는 가설로 표기(spec §A), 관측 불가 시 18·17 두 판정이 갈리면 `UNDETERMINED` |
| D2 `S_init` 실행 | REQ-ALH-017 · AC-ALH-009 신설. `prepareSafeInitHome` 기반 커밋 하네스 `TestHeadroomInitSurfaceExport`, 고정 플래그, `commands.log` 에 `moai init` 호출 0 건. 근거 코드 경로는 research §5 |
| D3 바이너리 출처 | `build_head`·`harness_sha256` 기계 줄(AC-ALH-002·009). 바이너리 대신 커밋된 하네스 파일 해시로 출처 고정 |
| D4 공허한 머리 검사 | `total_init`·`total_live`(재실행값과 일치)·`charmap`·절 내 `_미측정` 0건 검사. 뼈대에는 기계 줄이 없어 모든 MUST-PASS 가 FAIL 함을 acceptance 머리에 명시 |
| D5 백엔드 값 고정 | 형식 `레인 백엔드: <관측값> (출처: <관측 방법>)`, AC 는 형식만 검사. 뼈대 값은 `_미관측_` 으로 되돌림 |
| D6 표 완결성 | TSV 를 `candidates.py` 가 생성, AC 가 (file, section, gross, bind) 키를 재생성해 diff. REJECT 사유 열거와 열 대응 조건 |
| D7 조건 1 단서 | 용어 표에 단서 (a)/(b) 승계, `gov` 열, ADMIT 은 `gov=N` 요구 |
| D8 목적지 누적 | `dest` 열, `dest-sizes.tsv`, 두 트리 누적 `< 40000` awk, `pointers-<s>.tsv` 와 쌍 일대일 검사 |
| D9 UNDETERMINED 우회 | `gross` 열, `U = Σ gross(UNTRIED)`, `untried:<사유>` 필수, 토큰 재계산 awk, UNDETERMINED 도 상신 절차 필수 |
| D10 재조정 선택지 | `(a)|(b)|(공표값)|(기타)` + `J_includes_kanban_scope` 줄 |
| D11 F 정의 | `F(172ef22eb) = N` 서술값, 외삽 포함 표시, 판정 입력 아님 명시. `^F = ` 줄 0 건 검사 |
| D12 `/tmp` 금지 충돌 | 금지를 `evidence` 열과 `evidence:` 줄로 한정, 명령 기록은 `$SCRATCH` 표기 |
| D13 sec.py 출처 | plan 단계에서 커밋, sha256 기록(research §3), AC-ALH-005 해시 검사 |
| D14 흡수 내성 | AC-ALH-008 을 `git log --no-merges 7fe658815..HEAD -- <18경로 + 미러>` 로 변경. 흡수가 18경로를 바꾸면 재측정 규정(AC-ALH-002) |
| D15 AGENTS.md M1p | `$1 ~ /AGENTS\.md/ && $5!="M2" && ADMIT` 로 변경 |
| D16 표면별 해시 | `hash_init`·`hash_live` 실측 줄, AC-ALH-004 가 표면 값 사용 |
| D17 증거 존재 | 모든 행의 증거 파일 존재 검사 |
| D18 최소 해제 집합 | 「탐욕 해제 집합」으로 개명, 겹침 처리 규칙(plan M6) |
| D19 분할 범위·로케일 | 서문 후보 행, yaml 은 `config-data` 기각 행, `locale charmap` = UTF-8 기록 |
| D20 마커 리터럴 | plan·progress 의 설명 문장에서 리터럴 마커 표기 제거 |
| D21 트레일러 | 이번 커밋에 `Authored-By-Agent: manager-spec`. 최초 커밋 `10281a857` 의 INFO 는 이력 재작성 없이는 남는다 |

### plan-audit 3회차 — PASS-WITH-DEBT 0.86 과 채무 처분 (0.4.0)

판정: `.moai/reports/t1226/plan-audit-iter3.md`(커밋 `01ce1851e`). 필수 채무 셋은 run 착수 전에 기존 AC 안의 검사 줄로 흡수했다 — AC 추가·ID 변경 없음.

| 채무 | 처분 |
|---|---|
| DEBT-1 (N1, 필수) | **해소.** AC-ALH-003 (3) 에 `ADMIT` `chars ≤ gross`, M1·M1p `chars == gross`. AC-ALH-004 증거 루프에 M2 `pre_chars == gross`·`chars == pre_chars − post_chars`. M1p 증거에 `dup_source` 줄(`M1P-<s> BAD=0`) |
| DEBT-2 (N2, 필수) | **해소.** AC-ALH-003 (3) 에 감사자 awk 그대로 `OVERLAP-<s> BAD=0` |
| DEBT-3 (N3, 필수) | **해소.** AC-ALH-003 (3) 에 `DESTIN-<s> BAD=0`(ADMIT M1 목적지가 `count-set-<s>.txt` 안이면 FAIL). 17집합은 18집합에서 이미 FAIL 이므로 별도 규칙 불요 — 한 줄로 명시 |
| DEBT-4 N4 | **해소.** 18경로 리터럴 목록과 `count-set-live.txt` 내용 diff(`PIN-live-OK`), `count-set-init.txt` 도 같은 대조 규칙 |
| DEBT-4 N7 | **해소.** `research.md` 옛 번호를 REQ-ALH-014 로 |
| DEBT-4 N8 | **해소.** `BH ≠ HEAD` 이면 `BH` 를 새 격리 워크트리에서 열어 하네스 재실행 |
| DEBT-4 N10 | **해소.** 하네스 계약에 `t.Setenv("MOAI_DISTRIBUTE_ALL", "")` |
| DEBT-4 N11 | **해소.** `missing_init` 줄 도입, 줄 수 규칙을 「18(또는 17) − 누락 수」로 바꿔 §D.3 과 정합 |
| DEBT-4 N5·N6·N9 | **이월(선택).** `pointer_chars` 와 포인터 원문의 기계 대조, 흡수 병합 내부 편집 한계, 상위 경로 지시문 — run 단계 검토 항목으로 남는다 |

양성 대조(이번 실행, `$SCRATCH` 합성 TSV, BSD awk): 절 행과 `¶1` 문단 행을 동시에 `ADMIT` 으로 둔 표에서 `OVERLAP BAD=1`, 목적지 `workflow/skill-routing.md` 인 M1 `ADMIT` 에서 `DESTIN BAD=1`, 정상 `chars` 행에서 새 ROWS 조건 `BAD=0`.

### Implementation Kickoff — 자율 승인 기록

- 근거: `CLAUDE.local.md §31`(운영자 정책, 2026-09-26) — 「모든 킥오프 승인은 자율로 진행한다」, 진행 모드 자율. 카드 본문은 킥오프에 대한 운영자 게이트를 명시하지 않으므로 §31 예외에 해당하지 않는다.
- 승인 대상: SPEC-ALWAYS-LOADED-HEADROOM-001 v0.4.0, plan-audit 3회차 PASS-WITH-DEBT 0.86. 필수 채무 DEBT-1~3 은 위 표대로 이 개정에서 해소했고, 선택 채무 N5·N6·N9 는 run 단계 검토 항목으로 이월한다.
- 선택 기록: 판정 표면 `S_init` 1차(plan D1), 격리 하네스 경로(plan D2) — 모두 plan 의 권장안.
- 운영자 게이트로 남는 것: 카드 본문이 명시한 「동결 해시 해제 여부의 운영자 결정」(REQ-ALH-011·012). 이는 §31 이 인정하는 카드 본문 명시 예외이며, 리드가 리드 창에서 운영자에게 올린다. 레인은 결정하지 않는다.
- 킥오프 자율 승인은 plan-audit 판정과 별개의 게이트이며, PASS-WITH-DEBT 가 그것을 대신한 것이 아니라 §31 정책이 적용된 것이다.

### plan-audit 2회차 결함 대응표 (0.3.0)

| 결함 | 반영 |
|---|---|
| D1 흡수 형식 | AC-ALH-008 을 `git log --first-parent --no-merges 7fe658815..HEAD -- <36경로>` 로 교체, 한계(흡수 방향) 명시. REQ-ALH-013 을 「카드 계보의 비병합 커밋」으로. 1회차 처방이 틀렸음을 spec HISTORY 0.3.0 행에 기록 |
| D2 `S_init` 값 미검증 | AC-ALH-002 가 하네스를 `build_head` 에서 재실행해 `total_init` 을, `git archive` 로 `total_live` 를 재현. AC-ALH-004 가 두 트리에서 파이프라인을 재실행해 `hash_init`·`hash_live` 재현. AC-ALH-003 (2) 파일별 `Σ gross == wc -m` 분할 완결성 |
| D3 `net-negative` 우회 | 사유 조건을 `mech=M1`·`bind=0`·`gov=N`·`c1~c4=Y`·`dest≠-` 로 좁히고, 증거 `pointer_chars = P` 가 `P ≥ gross` 인지 검사 |
| D4 관측 설계 | 하네스가 init 프로젝트 트리 전체를 내보냄. 관측은 격리 `CLAUDE_CONFIG_DIR`·`HOME` 에서, `runtime_isolated = yes` 줄 필수. 관측자 사용자 범위 지시문 혼입은 가설로 표기(spec §A) |
| D5 집합 모순 | `count_set_init`(18/17/observed) 한 줄과 `count-set-<s>.txt` 도입. `total`·`current`·`A_adm`·`R`·`U` 를 그 집합 기준으로 정의(REQ-ALH-002·009·015), AC-ALH-006 이 집합 행만 합산 |
| D6 17집합 미검증 | `total/current/A_adm/R/U/T_min_init_17`·`verdict_init_17` 줄 요구, AC-ALH-006 `calc init <skill-routing 제외> init_17`, AC-ALH-007 토큰 재계산과 `count-set` 사유 규칙 |
| D7 REQ 예산 | 옛 REQ-ALH-014 를 REQ-ALH-001 에 병합, 옛 015~017 → 014~016. 16개. 매트릭스·추적성 표 갱신 |
| D8 외부 AC 참조 | 선행 SPEC 참조를 `AC-ALD2-002 [REF]` 로 표기 |
| D9 하네스 재실행 | `harness-run.txt` 첫 줄 `harness_head`, `build_head` 와 일치 검사. AC-ALH-002 에 재실행 절차 |
| D10 정규식 | `moai([[:space:]]+-[^[:space:]]+)*[[:space:]]+init` 로 확장, 자기 신고 한계를 Residual-risk 로 |
| D11 REQ 구현 세부 | REQ-ALH-016 에서 헬퍼 함수명 제거, 이름은 plan D2·AC-ALH-009 에만 |
| D12 린트 게이트 | M7·§D 제약·§D.1 게이트에 CI 판 `golangci-lint run ./internal/cli/...` |

### 린트

이 트리에서 빌드한 바이너리로 실행(설치본은 2026-09-25 빌드라 이 트리의 린트 규칙을 담는다는 보장이 없다):

```
$ go build -o $SCRATCH/moai ./cmd/moai
built
$ $SCRATCH/moai spec lint --strict SPEC-ALWAYS-LOADED-HEADROOM-001; echo "exit=$?"   # 0.4.0, 이 트리 재빌드
INFO      OwnershipTransitionUnmeasured  …/SPEC-ALWAYS-LOADED-HEADROOM-001/spec.md  1  SPEC SPEC-ALWAYS-LOADED-HEADROOM-001 transition "(none)" → "draft" expected owner "manager-spec" but commit 10281a8570f4d9b870cfa1d69e1a9cfac125dd36 (…) has no Authored-By-Agent trailer — ownership transition unmeasured

0 error(s), 0 warning(s)
exit=0
```

- `go test` 검증: AC-ALH-009 의 하네스 실행 한 줄이 `-run '^TestHeadroomInitSurfaceExport$'` 앵커와 공백 구분 PASS 판정(`--- PASS: TestHeadroomInitSurfaceExport `)을 쓴다. 린트 0 warning 으로 앵커 규칙 위반 없음.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
