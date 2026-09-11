# t622 — develop `ee99507fb` 흡수 뒤 재측정

- 흡수: `git merge --no-ff develop` → 병합 커밋 `7ac8b8491` (부모 `b25d1ba6e` · `ee99507fb`), 충돌 없음, exit 0
- `git merge-base develop HEAD` = `ee99507fbe3b4a22c6a0a74815723d222dfdc04d`
- 이 카드는 아직 범위 파일을 고치지 않았다. 아래 값은 모두 흡수 트리(`7ac8b8491`)에서 잰 것이다.

## 1. 줄번호 인용 — 템플릿 사본은 그대로, 로컬 사본만 밀림

SPEC 49행은 "인용 줄번호는 `b412f8a33` 기준이며 따로 적지 않으면 로컬·템플릿에서 같은 줄"이라고 적는다. 흡수 뒤 이 전제는 템플릿 사본에서만 참이다.

| 인용 | 템플릿(불변) | 로컬(흡수 뒤) | 측정 명령 |
|---|---|---|---|
| `manager-git.md:32` 기본값 설명 | 32 | 34 | `grep -n -E 'merge_method\|--squash' <copy>` |
| `manager-git.md:114` 병합 예시 | 114 | 116 | 같음 |
| `manager-git.md:156` 사전 읽기 배치 | 156 | 158 | `sed -n '146,172p'` |
| `manager-git.md` 148·166 `--auto-merge` 조건 | 148·166 | 150·168 | 같음 |
| `delivery.md` 337-338 · 343 · 348-349 · 355 · 404 | 그대로 | 362-363 · 368 · 373-374 · 380 · 429 (+25) | `grep -n -E 'merge_method\|--squash\|--merge' <copy>` |
| `workflows/sync.md` 사용법 95/85 · 플래그 114/104 | 85 · 104 | 95 · 114 (불변) | `grep -n -E -- '--merge' <copy>` |
| 명령 원본 `sync.md:3` · `sync.md.tmpl:3` | 3 | 3 | 같음 |
| `agent-common-protocol.md` Pre-Spawn 절 | 290~ (옛 2줄 배치) | 290~ (t635 Lane A/B 형태) | `grep -n -E 'Lane A\|fetch_status\|rev-list' <copy>` |

## 2. 로컬·템플릿 사본 차이 — SPEC A.4(89행) 기준선이 낡음

`diff <local> <template>` 덩어리 머리(흡수 뒤):

| 파일 | SPEC 89행 기준선 | 흡수 뒤 |
|---|---|---|
| `manager-git.md` | 바이트 동일 | `5,7c5` (develop 프론트매터 설명 +2줄, 로컬만) |
| `agent-common-protocol.md` | 바이트 동일 | `291,294d290` `296,297c292` `300c295` `302,306c297,298` `309,310c301,302` `314,318d305` (t635, 로컬만) |
| `delivery.md` | `275c275` `278c278` `479,480c479` | `9,11c9,11` `49,54c49,50` `154,163c150,152` `169,188c158,163` `300c275` `303c278` `447,448c422` `450,458c424,427` `510,511c479` |
| `doc-execution.md` | `138,143d137` | `9,11c9,11` `79,91d78` `118c105` `128,134c115,116` `144c126` `156,161d137` `173,174c149` `180c155` `182,189d156` |
| `quality-gates-context.md` | 바이트 동일 | `9,11c9,11` `48,49c48,49` `51c51` `138c138` `142,149c142` `151c144` `165c158` |
| `moai/SKILL.md` | 20개 덩어리 | 같은 20개 (변화 없음) |
| `references/reference.md` | `229d228` | 같음 |
| `workflows/sync.md` | `65,74d64` `81c71` | `29,31c29,31` 추가 |
| 명령 원본 | `2c2` | 같음 |

범위 파일 여섯 개에 develop 이 로컬 사본에만 넣은 차이가 생겼다. 이 카드 M6 의 사본 일치 판정(AC-GDP-013)이 이 차이를 "이 카드가 만든 차이"로 읽지 않으려면 기준선을 흡수 트리로 옮겨야 한다.

## 3. 사본 일치 테스트 기준선 — 이 카드 편집 전에 이미 빨강

명령: `go test ./internal/template/ -count=1 -run 'TestSanitizedPairParity|TestRuleTemplateMirrorDrift|TestHookWrapperCopiesStayIdentical' -v` → exit 1. 원본 출력: `.moai/reports/t622/absorb-mirror-baseline.txt`.

```
--- FAIL: TestRuleTemplateMirrorDrift/spec-workflow.md (0.00s)
--- FAIL: TestSanitizedPairParity/agent-common-protocol.md (0.01s)
--- FAIL: TestSanitizedPairParity/plan-auditor.md (0.01s)
sanitized_pair_parity_test.go:194: .claude/rules/moai/core/agent-common-protocol.md: normalized diff — 16 local-only, 5 template-only, ~5 reword pairs, net one-sided=11 (tolerance 4)
```

- `agent-common-protocol.md` 빨강의 원인은 t635 의 로컬 전용 Pre-Spawn 재작성이다(§2 표의 로컬 전용 덩어리). 이 카드 M5 가 고칠 바로 그 절이다.
- `spec-workflow.md`·`plan-auditor.md` 는 이 카드 범위 밖이다(plan.md §D 비접촉 목록에 `spec-workflow.md` 가 있다).
- `TestHookWrapperCopiesStayIdentical` 은 이 패키지에서 0건 선택됐다(`=== RUN` 줄 없음). 다른 패키지 소속이며 이 측정은 그 테스트를 보지 않았다.

## Gaps

- `TestHookWrapperCopiesStayIdentical` 은 재지 않았다(소속 패키지 미확인).
- develop 의 로컬 전용 편집이 의도된 분기인지, 템플릿 미러 누락인지는 판정하지 않았다 — 해당 카드(t635·t614 등) 소관이다.
