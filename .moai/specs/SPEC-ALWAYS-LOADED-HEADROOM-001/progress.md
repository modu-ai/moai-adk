# SPEC-ALWAYS-LOADED-HEADROOM-001 — 진행 기록

카드 t1226 · Tier M · 워크트리 `.claude/worktrees/t1226` · 브랜치 `WT-always-loaded-headroom` · 기준 커밋 `7fe658815`

---

## §E.1 Plan-phase Audit-Ready Signal

- plan_complete_at: 2026-09-27
- plan_status: audit-ready
- spec 버전: 0.2.0 (plan-audit 1회차 FAIL 0.62 수리본)
- 산출물: `spec.md` · `plan.md` · `acceptance.md` · `research.md` · `progress.md`, 판정서 뼈대 `.moai/reports/t1226/verdict.md`, `.moai/reports/t1226/sec.py`(sha256 `d0e61541367abb06a170bd36b6376e51d51882380ca9f899f0d2934016a78547`, 원본과 일치)
- SPEC ID 정규식 검사(Bash 실행): `SPEC-ALWAYS-LOADED-HEADROOM-001` → `PASS`. 중복 확인 `ls .moai/specs | grep -c HEADROOM` → `0`(작성 전)
- 기준선(오케스트레이터 실측, 이 트리): 18파일 `wc -m` → `199111 total`, 잔여 49,111
- 동결 다중집합 재실행(manager-spec, 이 트리): `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3`
- 결정 지점: D1 판정 표면 = `S_init` 1차 + `S_live` 병기. D2 `S_init` 산출 = 커밋된 격리 하네스 테스트. D3 표면별 해시 실측. D4 런타임 계수 관측 요청, 불가 시 18·17 이중 판정. 해소가 필요한 clarification 마커 0건.

### plan-audit 1회차 결함 대응표

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

### 린트

이 트리에서 빌드한 바이너리로 실행(설치본은 2026-09-25 빌드라 이 트리의 린트 규칙을 담는다는 보장이 없다):

```
$ go build -o $SCRATCH/moai ./cmd/moai
built
$ $SCRATCH/moai spec lint --strict SPEC-ALWAYS-LOADED-HEADROOM-001; echo "exit=$?"
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
