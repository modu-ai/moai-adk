# SPEC-CODEX-HOME-BACKSLASH-001 — 진행 기록

카드: t571 · 워크트리: `.claude/worktrees/t571` · 브랜치: `WT-codex-home-backslash`

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier S + 카드 지시로 `acceptance.md` 분리)
- plan-audit 1회차 판정 `FAIL 0.66`(`.moai/reports/t571/plan-audit.md`, Tier S 임계 0.75; testability 0.50 · traceability 0.50 이 동인, must-pass 실패 0건). 그 판정의 D1~D12 에 대한 수리를 반영했다 — 새 AC 는 AC-CHB-009 하나이며 기존 번호는 재부여하지 않았다
- 기준선 HEAD: `ee194493f` (`origin/develop` `3ac58b5a1` + 로컬 develop 흡수)
- 기준선 증거: `.moai/reports/t571/repro-asymmetry.log` (`EXIT=0`) — 인용만, 재유도 없음
- 선행: `SPEC-CODEX-SKILL-PATH-SLASH-001` (로컬 develop 착지, 원격 미착지 — 서술이며 이 카드에서 재검증하지 않음)
- 선행의 `status` 필드: `implemented`(실측 — `grep -n '^status:' .moai/specs/SPEC-CODEX-SKILL-PATH-SLASH-001/spec.md` → `5:status: implemented`). 사전 점검 술어는 엄격히 `completed` 이므로 **미충족**이며, 이 카드는 기록 남기는 override 를 택한다 — 근거와 로그 경로는 §spec.md HISTORY 참조. 브랜치 착지 축과 status 필드 축은 서로 다른 축이다
- 구속력 있는 판별식: AC-CHB-001 두 팔 대칭 + 뮤턴트 2종(AC-CHB-002 옛 순서, AC-CHB-009 `IsAbs` 위 올리기)
- 미해결 clarification 마커 0건 (계수 근거: `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-HOME-BACKSLASH-001/` 의 매치는 이 줄 하나뿐이고, 이 줄은 계수 진술이지 마커가 아니다. 이 줄은 게이트의 grep 이 찾는 대괄호 리터럴을 **의도적으로 담지 않는다** — 담으면 마커 0건인 SPEC 에서 게이트가 자기 자신에게 걸린다)
- SPEC lint (경계 있는 호출 — 경로는 **위치 인자**다):

  ```
  $ moai spec lint .moai/specs/SPEC-CODEX-HOME-BACKSLASH-001/spec.md
  ✓ No findings — all SPEC documents are valid
  EXIT=0
  ```

  **이 초록의 범위 한계를 명시한다.** 린터는 넘긴 `spec.md` **하나만** 파싱한다. 이 카드의 AC 는 분리된 `acceptance.md` 에 있으므로 린터의 `AC→REQ coverage` 규칙은 그 파일을 **본 적이 없다** — 즉 이 `EXIT=0` 은 frontmatter·GEARS·Out-of-Scope 규약을 말할 뿐 REQ↔AC 커버리지에 대해서는 **아무것도 말하지 않는다.** `acceptance.md` 나 `plan.md` 를 린터에 넘기는 것은 범주 오류이며(둘 다 artifact-statelessness 규약대로 frontmatter 가 없다) `ParseFailure` 를 낸다. 커버리지는 별도 손 판정으로 세운다(§acceptance.md DoD).
- 사전 측정 SKIP 기준선: `.moai/reports/t571/skip-baseline-pre-m2.log` — M2 **적용 전** 트리에서 `go test ./internal/cli/... -count=1 -timeout 600s -v` 를 돌려 SKIP 계수와 그 계수를 잰 HEAD SHA 를 함께 고정했다. AC-CHB-006 은 이 수와의 **정확한 일치**를 요구한다(종전 문안 「새로 SKIP 으로 바뀐 0건」은 비교 대상이 없어 판정 불가였다)

## §E.2 Run-phase Evidence

기준선 귀속: 모든 측정은 **이번 실행**, 이 트리(`.claude/worktrees/t571`, 브랜치 `WT-codex-home-backslash`),
HEAD `768306d2779aae7bce52ac6a78ec0839a498a203` 에서 나왔다. plan-phase 가 인용한 `ee194493f` 는
흡수 전 HEAD 이며, 흡수(fast-forward)로 `768306d27` 이 됐다 — 리드가 알린 `a4855f0b2` 보다 앞선 값으로,
develop 이 재측정과 병합 사이에 한 번 더 나아갔기 때문이다. **의존성 코드는 병합 트리에서 다시 확인했다**:
`classifyCodexSkillPath` 는 흡수로 바뀌지 않았고 옛 순서 그대로였다.

### 변경 표면

| 파일 | 성격 | 규모 |
|---|---|---|
| `internal/cli/doctor_codex.go` | 수정 — `classifyCodexSkillPath` 분기 순서 재배치 + doc 주석 | 분기 4줄 + 주석 |
| `internal/cli/codex_skills_path_shape_test.go` | 신규 — 이 카드의 가드 전부 | 5 테스트 함수 |

새 shape 상수 0개, 새 production 시접 0개. `shapeName` 은 테스트 파일 안의 표이며 production 시접이 아니다.

### RED → GREEN (`.moai/reports/t571/red-before-fix.log`, `green-after-fix.log`)

수리 **전** 실행. `RED_EXIT=1`:

```
--- FAIL: TestCodexSkillPathBackslashSymmetry
    --- FAIL: .../home_relative_arm_matches_relative_arm
    --- FAIL: .../both_arms_are_oddly_formed
--- FAIL: TestJudgeCodexSkillEntryRefusesHomeRelativeBackslash
    --- FAIL: .../refuses_before_stat
    --- PASS: .../positive_control_clean_path_does_stat_once
--- FAIL: TestCodexBackslashRefusalIsSymmetricAcrossReadAndWrite
--- PASS: TestCodexSkillPathPreservedShapes            (4/4 — 보존 칸, 예정대로 초록)
--- PASS: TestJudgeCodexSkillEntryBackslashHomeStillExpands  (기각 설계 고정, 예정대로 초록)
```

양성 대조군이 RED 판에서도 PASS 였다는 것이 중요하다 — 계수기가 실제로 배선돼 있음을 세우므로,
`refuses_before_stat` 의 실패가 「계수기가 죽어 있어서」가 아니라 「stat 이 실제로 걸려서」임이 갈린다.

수리 **후** 같은 명령: `GREEN_EXIT=0`, 하위 테스트 포함 13 PASS / 0 FAIL.

### 뮤턴트 2종 — 하나씩 주입·복구, 주입 직전 트리 재확인

각 주입 전 `git status --porcelain` 과 `git rev-parse --short HEAD` 를 다시 읽어 이 카드의 변경만 있고
HEAD 가 `768306d27` 임을 확인했다. 두 뮤턴트를 동시에 주입하지 않았고, 각각 복구 후 재실행해 초록을 봤다.
이 창 안에서 빌드한 바이너리는 없다.

| 뮤턴트 | 주입 | 관측 (`EXIT=1`) | 로그 |
|---|---|---|---|
| AC-CHB-002 | 옛 순서 복원(`~/` 껍질이 백슬래시 검사보다 먼저) | 대칭 가드 RED 2건 + 실패 출력에 상수 이름 `codexPathHomeRelative` 2회. **AC-CHB-004 의 「stat 계수 0」도 같은 창에서 함께 RED** — 물려받은 RED-now 칸이 실재함이 관측됐다. 보존 칸 4개는 초록 유지 | `mutant-old-ordering.log` |
| AC-CHB-009 | 백슬래시 검사를 `IsAbs` **위로** | `--- FAIL: TestCodexSkillPathPreservedShapes/posix_absolute_with_backslash` — `"/tmp/a\\b" classified codexPathOddlyFormed, want codexPathAbsolute`. **그 칸 하나만** 붉어졌고 나머지 세 칸과 대칭 가드는 초록 | `mutant-isabs-hoist.log` |

AC-CHB-009 가 존재하는 이유가 여기서 갚아졌다: `/tmp/a\b` 칸은 이 카드 이전에도 이후에도, 그리고
AC-CHB-002 의 뮤턴트 아래에서도 초록이었다 — **한 번도 실패를 본 적이 없는 검사**였다. 이 뮤턴트가
그 칸을 한 번 붉게 만들어 채택 근거를 세운다.

복구 후 재실행: `green-after-mutants.log`, `EXIT=0`, 13 PASS / 0 FAIL.

### AC-CHB-008 규율 — 양성 대조군 포함

```
$ test -f internal/cli/codex_skills_path_shape_test.go   → EXIT=0
$ grep -c '[^A-Za-z]t\.Parallel()' internal/cli/codex_skills_path_shape_test.go   → 0
$ grep -c '[^A-Za-z]t\.Cleanup(' internal/cli/codex_skills_path_shape_test.go     → 5
```

양성 대조군: 시접 함수 하나에 병렬 opt-in 을 임시 주입하니 계수가 `0 → 1` 로 움직였고, 복구 후 다시 `0`.
대조군 없는 0 은 부재 증거가 아니므로 이 절차를 생략하지 않았다.

**여기서 자기적중을 한 번 겪었고 기록해 둔다.** 첫 측정이 `1` 을 냈는데 호출은 하나도 없었다 — 파일 머리의
주석이 「이 파일엔 그 토큰을 쓰지 않는다」고 **그 토큰을 적으면서** 설명하고 있었다. 이 카드에서 같은 형태가
세 번 나왔다(① plan-audit D11: SPEC 문구가 게이트의 마커를 담음 ② 수리 라운드: 완료조건 줄이 자기가 세는
토큰을 담음 ③ 여기). 셋 다 「검사가 자기 자신을 센다」는 한 가지 형태이고, 사람이 세 번 반복할 만큼 눈에
띄지 않는다. 수리는 검사식을 복잡하게 만드는 쪽이 아니라 산문에서 리터럴을 빼는 쪽으로 했다 — 명령이 그대로
판정이 되게 두는 것이 이 AC 의 요지였기 때문이다.

### 정적 검사 + 패키지 회귀 (AC-CHB-006)

```
$ go vet ./internal/cli/                                  → EXIT=0 (출력 없음)
$ go test ./internal/cli/... -count=1 -timeout 600s -v    → EXIT=0
```

| 항목 | 기준선(`768306d27`, 흡수 후·M2 전) | M2 적용 후 |
|---|---|---|
| SKIP 계수 (`grep -c '^ *--- SKIP: '`) | 30 | **30** |
| `FAIL` 줄 | 0 | **0** |
| 통과 패키지 (`^ok`) | — | 17 |
| SKIP 이름 집합 (경과시간 제거 후 `comm -3`) | — | **무출력 = 동일** |

증거: `skip-baseline-post-absorb.log`(기준선) · `post-m2-skiplines.txt`(이번 판정).

**대조식의 경과시간 제거가 여기서 값을 했다.** 접미사를 남긴 채 비교하면 같은 테스트가
`(0.01s)` 대 `(0.02s)` 로 어긋나 매 실행 1건짜리 가짜 델타가 난다(흡수 전후 대조에서 실제로 관측했고
`skip-baseline-post-absorb.log` 에 적어 뒀다). 항상 나는 가짜 경고는 읽는 사람에게 무시하는 습관을
가르치고, 진짜 회귀는 그 습관을 타고 지나간다.

**전체 스위트(`go test ./...`)는 로컬에서 돌리지 않았다** — CI 판정 몫이며, 병렬 레인이 동시에 돌려
머신을 마비시킨 전례가 있다.

### Gaps — 관측하지 않은 것

- **AC-CHB-007 은 한 번도 붉어진 적이 없다.** 그것이 고정하는 설계(검사를 확장 **뒤**로 이동)는 §3.2 가
  기각한 것이라 어떤 절차도 그 뮤턴트를 주입하지 않는다. 이 AC 가 세우는 것은 「홈 스텁이 실제로 먹었다」
  하나뿐이며(관측된 stat 인자 `C:\Users\x/ok/SKILL.md`), 그 이상을 주장하지 않는다. 관측으로 승격하려면
  세 번째 뮤턴트가 필요하고 이 카드는 두지 않았다.
- **원시 `-v` 출력은 보존하지 않았다.** 두 실행이 각 18,847줄 / 1.1MB 였고 그 중 ~99%가 이 파일이 주장하지
  않는 `=== RUN`·`--- PASS` 잡음이라, 판정에 쓰인 줄만 추린 47줄 추출본(`post-absorb-skiplines.txt`,
  `post-m2-skiplines.txt`)을 남겼다. 손실은 통과 테스트 잡음이며, 그 사실을 기준선 파일 안에 적어
  독자가 나중에 발견하게 두지 않았다. 전체 출력은 기록된 COMMAND 로 재생성 가능하다.
- **`golangci-lint` 는 돌리지 않았다.** `go vet` 만 실행했다.
- **크로스컴파일(`GOOS=windows`)은 돌리지 않았다.** 이 변경은 플랫폼 분기 없는 순수 분류 로직이지만,
  안 돌린 것은 안 돌린 것이므로 gap 으로 적는다. CI 매트릭스 몫이다.
- **흡수 후 SPEC 편집분은 감사받지 않았다.** plan-audit 2회차(`PASS 0.88`)는 `ee194493f` 트리를 채점했고,
  그 뒤 N1~N10 수리와 기준선 재고정이 들어갔다.

### Residual risk

- SKIP 30 은 호스트·환경에 묶인 수다(live 백엔드·TTY·크로스컴파일 술어로 건너뛰는 항목이 여럿). 다른
  환경에서 30 이 아니어도 회귀가 아닐 수 있으며, **판정은 이름 대조**이지 계수가 아니다.
- 이 카드는 `..` 이탈(`~/../../etc/x`)을 닫지 않는다. 그 형태는 백슬래시를 품지 않아 M2 를 그대로 통과하며,
  같은 삭제 경로에 도달한다. §spec.md §4 가 범위 밖으로 선언했고 후속 카드 소관이다.
- 선행 SPEC 의 `status` 는 여전히 `implemented` 다. run 진입은 기록된 override 로 했고
  (`.moai/logs/depends-on-override.log`), 이 카드가 그 SPEC 을 닫지 않는다.

## §E.3 Run-phase Audit-Ready Signal

- `run_status: audit-ready`
- AC 판정: **9/9 PASS** (AC-CHB-001 ~ 009). 각 명령·종료 코드·출력은 §E.2 와 `.moai/reports/t571/` 의
  로그에 있다. PASS-WITH-DEBT 0건, 이월 0건.
- 뮤턴트 2종 주입·관측·복구 완료. 소스는 원상 복구됐고 복구 후 초록을 재확인했다.
- 측정 HEAD: `768306d2779aae7bce52ac6a78ec0839a498a203`

## §E.4 Sync-phase Audit-Ready Signal

- `sync_status: audit-ready`
- 산출물: `CHANGELOG.md` `[Unreleased] → ### Fixed` 항목 1건(사전 중복 검사
  `grep -c 'SPEC-CODEX-HOME-BACKSLASH-001' CHANGELOG.md` → `0`), `progress.md` §E.2~§E.4,
  `.moai/reports/t571/` 증거 12파일.
- `sync_commit_sha: pending-backfill-sync`
- 병합: 리드에 창을 요청한다. 이 브랜치는 미푸시이며 워크트리가 작업의 유일본이므로, 원격 착지가 확인되기
  전에는 폐기하지 않는다.
