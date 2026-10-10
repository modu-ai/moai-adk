# SPEC-ROLE-INJECTION-BUDGET-001 — Acceptance Criteria

> 두 칸 채택(`verification-completeness.md` §2): 모든 AC 는 RED-now 칸(명령 + 축자 stdout + exit + 트리 SHA)과 green 경로 칸(플립 마일스톤 + 통과 형태)을 함께 가진다. RED 명령은 전부 단일 호출(파이프·`&&`·`;` 체이닝·서브셸 없음)이며 exit 는 별도 필드로 기록한다. 축자 원문은 §C 증거 원장이 운반하고 표 셀은 원장 id 를 인용한다(권장 운반체 — 표 셀의 셸 메타문자 훼손 방지).
> RED 측정 트리: `2aab5f797` (브랜치 WT-3-2-1, 카드 워크트리 `.moai/worktrees/t1617`) — 2026-10-10 관측.

## §A. AC 목록

| AC | 요구 | 분류 | RED-now | Green 경로 |
|---|---|---|---|---|
| AC-RIB-001 | 조립 예산 테스트(REQ-RIB-001/002)가 템플릿·배포 양 트리에서 레지스트리 역할마다 조립 합본 ≤9,000 을 단정하고 breakdown 을 출력한다 | release-blocking | E1+E2 — core 17,793/18,114 > 3,946, exit 1 (이유: 현재 core 가 파생 예산을 초과 — 이 SPEC 이 고치는 상태 그 자체) | M1 작성 시 같은 명령이 FAIL 로 관측되고, M2 재작성 뒤 `go test -run TestRoleInjectionAssemblyBudget ./internal/hook/` 가 `ok` — 각 역할 행의 breakdown 이 9,000 이하를 보인다 |
| AC-RIB-002 | 재배치 감사 원장(REQ-RIB-003)이 core 의 모든 앵커 영역을 한 행 이상 가진다 | release-blocking | E3 — 원장 부재, exit 2 (이유: 감사 원장이 아직 작성되지 않음) | M2 가 `relocation-ledger.md` 를 작성 — 36행 이상, 모든 행이 목적지(companion 절 확인 포함) 명명; plan-audit·sync-audit 가 원장을 읽고 의미 보존을 판정 |
| AC-RIB-003 | 스텁 게이트(REQ-RIB-005)가 `*-core.md` 를 기계 열거해 표지 0·≤10,000 을 단정한다 | regression-guard | 오늘 이미 참(표지 0, 8,583/8,072 ≤10,000) — RED-now 대신 §1.1 관측-실패 완료: 게이트 자체의 모터 서브테스트(표지 심은 픽스처·초과 픽스처에서 FAIL 관측)가 채택 증거 | M1 작성, 모터 관측을 progress.md §E.2 에 기록; M2 전체 스위트 Green |
| AC-RIB-004 | InjectionFailed 경로(REQ-RIB-006)가 불일치를 이름 대고 `moai update` 를 안내한다 (4개 로캘 포함) | release-blocking | E4 — `moai update` 0히트, exit 1 (이유: 안내 문구가 아직 어디에도 없음) | M3 플립: 술어·로캘 테스트 `go test -run 'TestRoleRulesVersionSkew' ./internal/hook/` `ok`, 4개 로캘 행 모두 안내 포함 |
| AC-RIB-005 | 불일치 술어(REQ-RIB-007)가 픽스처로 단위 시험된다 — 라이브-리포 자기 단정 없음 | release-blocking | E5 — `template_version` 0히트, exit 1 (이유: role_rules.go 에 불일치 판독이 아직 없음) | M3 플립: 같음/다름/판독불가 3케이스 + 4로캘 안내 테스트 `ok`; 라이브-리포 단정 테스트는 존재하지 않음(부재를 grep 으로 확인) |
| AC-RIB-006 | 크기 게이트 doc 사다리가 1회로 정리되고 지시문·경고 문구가 통일된다 (REQ-RIB-008) | release-blocking | E6 — 사다리 문구 2회, exit 1 (이유: :393–413 중복이 현재 존재) | M3 플립: 같은 awk 가 exit 0 (n=1); `10000` 무천단위 표기 0히트; NOTE·경고 중복 서술 제거 |
| AC-RIB-007 | hooks-system.md 가 두 한도(50K 총 stdout 저장 / 10,000자 문자열당 전달)를 구분 기재한다 (REQ-RIB-009) | release-blocking | E7 — `10,000` 0히트, exit 1 (이유: 문자열당 한도 서술이 아직 없음) | M3 플립: 같은 grep ≥1; 두 한도 각각 실측 출처(Q4·lane-15 공지) 인용 행 존재 — 템플릿 미러 원본 편집 선행 |
| AC-RIB-008 | 재작성 뒤에도 로컬 gitflow 규칙의 Isolation 지목이 살아 있는 절로 해석된다 (REQ-RIB-010) | regression-guard | 오늘 참(절 :198 존재 + [HARD] 이동 금지 영역 생존) — RED-now 불가(rewrite 생존 성질). Green 경로가 생존을 관측 | M2: 재작성 후 `grep -c '## Isolation' factory-dispatch.md` ≥1 + 이동 금지 [HARD] 문장 생존 + 로컬 규칙 지목 문장과 대조 기록 |
| AC-RIB-009 | 템플릿 미러 정합 + `stub-delta:` 라벨의 상시 파일 델타 기록 (REQ-RIB-011) | release-blocking | E8 — progress.md 에 `stub-delta` 0히트, exit 1 (이유: 델타 기록이 아직 작성되지 않음; 라벨은 REQ-RIB-011 이 고정) | M2 미러 정합 유지(기존 rule-template 미러 테스트) + M3 가 progress.md §E.2 에 `stub-delta:` 라벨 행(스텁 전후 UTF-16 코드 단위·바이트, 1,000바이트 초과 성장 시 비산 문) 작성 → 같은 grep ≥1 |
| AC-RIB-010 | 오버플로 경로가 보존된다 — 기존 사다리 테스트가 재작성 후에도 Green (§D 제약) | regression-guard | 오늘 Green(`TestSessionStartRoleRulesSizeGate`) — 보존 성질이라 RED-now 불가 | M3: 전체 영향계열 스위트에서 동일 테스트 Green 유지 — 플립 없음이 곧 통과 형태 |
| AC-RIB-011 | 상위 원장 36행 after-text 가 재작성된 배포 트리와 정합한다 (REQ-ALB-015 상시 계약) | regression-guard | 편집 전 원장-정합은 현재 트리에서 이미 Green — 실행 가능한 RED-now 가 구성 불가(§2 undecidable 처분, release-blocking 제외). 채택 증거는 M2 편집 후 재관측이다 | M2: 규칙 편집 뒤 `go test ./internal/template/` 원장 테스트 — 36행 after-text 갱신 전 RED 관측 → 갱신 뒤 `ok`. 편집-후 관측이 채택 증거 |
| AC-RIB-012 | 라이브 레인 재진입(경계 소스 봉투 — startup + clear(핸드오프 대기 없음))에서 오버플로 공지 부재 + 합본 ≤10,000 관측 (§A.5 완료 기준 실측 반쪽) | regression-guard | 역사적 RED — lane-15 공지(§A.1 인용, 2026-10-09/10 관측)는 재실행 불가 → §2 undecidable 처분, release-blocking 제외 | M3: startup 재진입 + clear(핸드오프 대기 없음) 재진입 관측 — "역할 규칙 주입 초과" 부재 + systemMessage 이상 무, progress.md §E.2 에 세션·관측 기록 |
| AC-RIB-013 | 봉투 밖 재주입 세션(핸드오프 본문 클레임 clear · armed-goal compact)에서 REQ-ALB-010 사다리가 단언 대로 작동한다 — 넘침 파일 폴백 유지 + 게이트 발화가 systemMessage 로 가시 (REQ-RIB-012) | regression-guard | 사다리의 과잉 분기 동작은 오늘 Green(`TestSessionStartRoleRulesSizeGate` 서브테스트 b/c — 보존 성질이라 RED-now 불가, §2 처분) | M3: 기존 사다리 서브테스트 Green 유지 + 봉투 밖 소스(클레임된 핸드오프 본문 · armed goal 재주입)를 합성한 과잉 픽스처에서 폴백 유지·systemMessage 가시를 단정하는 서브테스트 Green |

## §B. 분류 요지

- **release-blocking 7개** — AC-RIB-001·002·004·005·006·007·009 — 각각 §2.1 네 요소(명령·축자 stdout·exit·트리 SHA)를 §C 원장 행으로 충족한다. 원장 항목은 8개(E1–E8): E1·E2 가 함께 AC-RIB-001 의 양 트리 관측이다. AC-RIB-001 의 단정 범위는 §A.5 경계 소스 봉투(startup · clear-핸드오프 대기 없음 · compact-goal 없음)로 재범위됐다(결정 d-20261010T072910Z-f1c3). AC-RIB-009 의 RED 는 REQ-RIB-011 이 고정한 `stub-delta:` 라벨의 부재 관측이다(E8).
- **regression-guard 6개** — AC-RIB-003·008·010·011·012·013 — §2 undecidable/보존-성질 처분을 따른다: release-blocking 상등을 받지 않으며, green 경로가 생존·보존·편집-후 재관측을 운반한다. AC-RIB-011 은 F3 수리로, AC-RIB-013 은 결정 d-20261010T072910Z-f1c3 의 봉투 밖 단언으로 regression-guard 다.
- **변이 탐침**(§2): AC-RIB-001은 core 를 3,947자로 늘리는 변이로 실패함을 테스트 자체가 보인다(단정이 경계값 3,946에서 작동). AC-RIB-003은 모터 서브테스트가 변이 관측을 내장한다.

## §C. 증거 원장 (RED-now — 트리 2aab5f797, 2026-10-10 관측; E1/E2는 F1 수리 임계 3,946으로 재관측 — 값·exit 동일)

```text
[E1] AC-RIB-001 (template tree)
$ python3 -c "c=open('internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md',encoding='utf-8').read(); S='<!-- moai:role-core-start -->'; E='<!-- moai:role-core-end -->'; rs=[x.split(E)[0] for x in c.split(S)[1:]]; j='\n\n'.join(rs); n=sum(2 if ord(ch)>=0x10000 else 1 for ch in j); print(n); exit(0 if n<=3946 else 1)"
17793
exit code: 1

[E2] AC-RIB-001 (deployed tree)
$ python3 -c "c=open('.claude/rules/moai/workflow/factory-dispatch.md',encoding='utf-8').read(); S='<!-- moai:role-core-start -->'; E='<!-- moai:role-core-end -->'; rs=[x.split(E)[0] for x in c.split(S)[1:]]; j='\n\n'.join(rs); n=sum(2 if ord(ch)>=0x10000 else 1 for ch in j); print(n); exit(0 if n<=3946 else 1)"
18114
exit code: 1

[E3] AC-RIB-002
$ awk 'END{exit !(NR>=36)}' .moai/specs/SPEC-ROLE-INJECTION-BUDGET-001/relocation-ledger.md
awk: can't open file .moai/specs/SPEC-ROLE-INJECTION-BUDGET-001/relocation-ledger.md
 source line number 1
exit code: 2

[E4] AC-RIB-004
$ grep -c 'moai update' internal/hook/role_rules.go
0
exit code: 1

[E5] AC-RIB-005
$ grep -c 'template_version' internal/hook/role_rules.go
0
exit code: 1

[E6] AC-RIB-006
$ awk '/REQ-ALB-009 retreat/{n++} END{print n; exit !(n==1)}' internal/hook/role_rules.go
2
exit code: 1

[E7] AC-RIB-007
$ grep -c '10,000' .claude/rules/moai/core/hooks-system.md
0
exit code: 1

[E8] AC-RIB-009
$ grep -c 'stub-delta' .moai/specs/SPEC-ROLE-INJECTION-BUDGET-001/progress.md
0
exit code: 1
```

참고(측정 귀속 — F1 수리 산식): 조립 합본 = 생산자 + `"\n\n"` + 헤더(뒤 개행 포함) + core + `"\n\n"` + 포인터 (role_rules.go:102–123 `assembleInjectionComposite` · :373–380 `roleRuleInjectionFor` — 헤더의 뒤 `\n\n` 뒤에 core가 직접 붙어 결합자 없음). 오버헤드 257(leader 최악). 생산자 상한 4,797 = 게이트 실측 합본 23,166 − 본 트리 역할 블록 18,369(= 2 + 123 + 18,114 + 2 + 128; 18,114에 로컬 꼬리 +321 포함). core 예산 3,946 = 9,000 − 4,797 − 257. 원문 전체는 `.moai/reports/t1617/measurements.md`.
