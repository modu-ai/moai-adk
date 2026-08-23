# M9 보고서 — 검증 스윕 기록 + run-phase 감사 신호

SPEC: `SPEC-FEEDBACK-AUTO-SUBMIT-001` (Tier L, cycle=tdd)
카드: t170 / lane-7 · 브랜치 `WT-auto-feedback` · 워크트리 `.claude/worktrees/t170`
범위: **기록 전용 절반.** 통합·PR 은 리드 보류 대상이며 이 마일스톤은 커밋에서 멈춘다.

---

## 1. Claim (주장)

1. `progress.md` §E.2 에 `### M9 — 검증 스윕` 을 추가해, 오케스트레이터가 `a6682a007` 에 대해 관측한 검증 스윕 결과를 **귀속을 명시한 채** 기록했다.
2. `progress.md` §E.3 의 placeholder `_<pending run-phase>_` 를 sync-phase 감사관이 단독으로 읽을 수 있는 run-phase 감사 신호로 채웠다 — YAML 신호 블록, 마일스톤 커밋 8건, AC 24/24 커버리지 표(소유 M + 판정 명령), 스윕 증거, 미검증 항목 7건, 외부 블로커 1건, 누적 잔여 위험 28건.
3. §E.4 는 손대지 않았다(`_<pending sync-phase>_` 그대로).
4. 이 보고서 파일을 5-섹션 형식으로 남겼다.

**주장하지 않는 것**: 이 에이전트는 어떤 검증 명령도 새로 실행하지 않았다. §E.3 의 모든 스윕 결과는 오케스트레이터 관측의 **인용**이며, 그 사실을 §E.3 본문과 M9 절 첫 문단에 명시했다.

---

## 2. Evidence (증거)

### 이 마일스톤이 직접 관측한 것

편집 전 placeholder 실재 확인:

```
$ grep -n '^## §E' .moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/progress.md
3:## §E.1 Plan-phase Audit-Ready Signal
18:## §E.2 Run-phase Evidence
1040:## §E.3 Run-phase Audit-Ready Signal
1044:## §E.4 Sync-phase Audit-Ready Signal
        (1042행 = `_<pending run-phase>_`)
```

AC 정의 개수 (acceptance.md):

```
$ grep -c '^### AC-F-' .moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/acceptance.md
24
```

§E.2 의 PASS 판정 전수 — `grep 'AC-F-0[0-9][0-9]' progress.md | grep -i PASS` 로 M1~M8 판정 줄을 열거해, 24개 id 가 모두 판정을 받았음을 확인했다. 두 어순(`AC-F-0NN 판정: **PASS**` / `판정: AC-F-0NN **PASS**`)이 함께 존재하며 문구를 통일하지 않았다(카드 지시).

마일스톤 커밋 8건:

```
$ git log --oneline -9
a6682a007 feat(...): M8 template mirror, key inventory, and 4-locale docs
23c5c18fa feat(...): M7 expose feedback section in the web console
38705eb85 feat(...): M6 skill gate clauses and wizard question
d2063308b feat(...): M5 moai feedback scrub and queue verbs
55dc0ec0a feat(...): M4 mask log and retry queue
3bcceffc7 feat(...): M3 vulnerability classifier reading the pre-mask body
e51475068 feat(...): M2 scrubber type contract and masking transforms
95fc239e3 feat(...): M1 feedback.auto_submit config key
3210da7d3 merge(release/v3.1.3): pick up peer lane landings before t170-2 run
```

편집 후 구조:

```
$ grep -n '^## §E' .moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/progress.md
3:   §E.1 / 18: §E.2 / 1076: §E.3 / 1217: §E.4

$ git diff --stat
 .../SPEC-FEEDBACK-AUTO-SUBMIT-001/progress.md | 175 +++++++++++++++++-
 1 file changed, 174 insertions(+), 1 deletion(-)
```

삭제 1줄은 placeholder `_<pending run-phase>_` 이며, §E.4 의 placeholder 는 그대로다.

### 인용한 것 (오케스트레이터가 `a6682a007`, clean tree 에서 관측)

- 영향 9개 패키지 `-count=1` 회귀 → 전부 `ok`, 예외 1건(`internal/cli/agentlint` `TestConstitutionCrossReference`)
- `go test -race -count=1 ./internal/feedback/...` → `ok … 1.856s`
- 7개 패키지 트리 `go vet` / `GOOS=windows go vet` → 양쪽 exit 0, 무출력
- `golangci-lint run --timeout=5m` → `0 issues.`
- `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...` → `ok … 24.045s`, agentemit `ok`
- `make build` 성공 + 직후 `git status --porcelain` 무출력

외부 블로커 귀속(오케스트레이터 측정): `origin/main` 의 `moai-constitution.md` `grep -c agent-authoring` = 1, `origin/release/v3.1.3` = 0, 제거 커밋 `243eb07ef`(t82 M4 예산 다이어트). 복구는 카드 **t189**(lane-8) 소관.

---

## 3. Baseline-attribution (baseline 귀속)

- 트리: `.claude/worktrees/t170`, 브랜치 `WT-auto-feedback`, HEAD `a6682a007`, base `3210da7d3`.
- 편집 시작 시점 `git status --porcelain` 무출력(clean) — 이 커밋의 diff 는 전부 이 마일스톤 귀속이다.
- 인용된 스윕은 **같은 커밋 `a6682a007`, 같은 clean tree** 에 대한 오케스트레이터 관측이다. 다른 트리·다른 시점의 수치를 끌어오지 않았다.
- AC 판정 근거 명령은 §E.2 에 남은 M1~M8 verbatim 출력에서 옮긴 것이며, 이 마일스톤이 재실행하지 않았다.

---

## 4. Gaps (미검증)

1. **어떤 검증 명령도 이 마일스톤에서 재실행하지 않았다.** 카드가 "다시 돌리지 말고 기록하라"고 지시했다. §E.3 의 스윕 수치는 전부 인용이다.
2. **M9 기록 커밋 자체는 스윕 대상이 아니다.** 이 커밋은 `progress.md` + 보고서 파일만 건드리므로 Go 빌드·테스트에 영향이 없지만, `a6682a007` 이후 트리에 대한 검증은 존재하지 않는다. §E.3 의 `run_commit_sha` 를 `a6682a007` 로 적고 그 사실을 주석으로 남긴 이유다.
3. **`TestConstitutionCrossReference` 를 직접 돌리지 않았다.** base 붉음과 귀속은 오케스트레이터의 `grep -c` 측정 인용이다.
4. **누적 잔여 위험 28건은 M2~M8 보고서 §5 의 취합이지 재검증이 아니다.** 각 항목이 지금도 유효한지 다시 확인하지 않았다 — 마일스톤 착지 시점의 기록이다.
5. **`.moai/reports/t170/` 의 M1 보고서가 없다.** M2~M8 만 존재하므로 M1 잔여 위험은 취합에 포함되지 않았다(§E.2 M1 절에는 판정과 근거가 남아 있다).
6. **통합·PR 관련 검증 전무.** push·`gh pr` 조회·릴리스 브랜치 상태 확인 모두 하지 않았다(리드 보류).

---

## 5. Residual-risk (잔여 위험)

1. **`run_commit_sha: a6682a007` 이 M9 커밋을 가리키지 않는다.** sync-phase 에서 backfill 하거나, 감사관이 "run 검증은 M8 헤드 기준"이라고 읽어야 한다. 이 SPEC 은 후자를 택해 주석으로 명시했다.
2. **외부 블로커 t189 가 먼저 착지하지 않으면 통합 후 전체 스위트가 붉다.** 리드의 통합 큐가 lane-8(t183→t189)을 앞에 둔 이유이며, 순서가 뒤집히면 이 카드의 PR CI 가 남의 결함으로 붉어진다.
3. **형제 SPEC `SPEC-TODO-ENABLE-FLAG-001` 과 공유 파일 9종.** 통합 시점의 병합 충돌은 인접 줄 수준이지만 개수 고정 테스트 4건은 재조정이 필요할 수 있다(§E.3 잔여 위험 22·23).
4. **§E.3 이 길다(≈140줄).** 감사관이 읽을 것을 전제로 자족성을 택했으나, 이후 마일스톤이 추가되면 이 절을 갱신할 때 어디를 고쳐야 하는지 분산된다.
5. **누적 잔여 위험 중 17·20 은 이 SPEC 의 강제력 한계 자체다.** 스크러버는 규약 강제이지 샌드박스가 아니고, `feedback.auto_submit` 은 Go 코드가 읽지 않는다. sync-phase 감사관이 "마스킹이 강제된다"로 읽지 않도록 §E.3 에 굵게 남겼다.
