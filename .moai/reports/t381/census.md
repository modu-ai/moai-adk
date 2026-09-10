# t381 census — 무시되는 증거 경로 인용 전수

- 측정 트리: `3f03d9c36` (`.claude/worktrees/t381`, 브랜치 `WT-ignored-evidence-cite`, `origin/develop` 와 동일)
- 이 문서가 `discovery.md` §2-§3 의 계수를 **대체한다**. 그쪽 계수는 배제 프로브로 얻은 것이라 무효.

---

## 1. 프로브 — 그리고 이 프로브가 못 잡는 것

census 를 다루는 카드이므로 프로브 자체를 먼저 적는다. **프로브가 대상의 일부를 구조적으로 배제하면 그 배제는 결과에 나타나지 않는다** — 이 카드에서 이미 세 번 발생했다(§1.2).

```bash
git grep -n '\.moai/state/verify' -- . ':!*.md'      # → 25줄  [카드 범위 census]
```

폭을 달리한 대조군, 같은 트리:

| 프로브 | 결과 | 용도 |
|---|---|---|
| `git grep -n '\.moai/state/verify' -- '*.go'` | **13** | lane-8 측정치와 일치 확인 |
| `git grep -n '\.moai/state/verify' -- . ':!*.md'` | **25** | 이 카드의 census |
| `git grep -n '\.moai/state/verify'` (전 확장자) | 531 | `.md` 포함 시 규모 |
| `git grep -n '\.moai/state' -- . ':!*.md'` | 492 | `/verify` 밖 포함 시 규모 |
| `git grep -n '\.moai/state'` | 3206 | 최대 폭 |

### 1.1 이 프로브가 못 잡는 것 (배제 목록)

| # | 못 잡는 것 | 규모 |
|---|---|---|
| B1 | `.md` 파일 — 의도적 제외(t375 소관). t375 가 실제로 `.md` 한정인지는 **미확인**(lane-8 run 중, develop 부재) | 531 − 25 = 506줄 |
| B2 | `.moai/state/` 안에서 `/verify` 가 **아닌** 경로 인용 (`session-memo.md`, `handoff/`, `session-msg/`, `config-cache.json` …) | 467줄 |
| B3 | **미추적 파일** — `git grep` 은 추적 파일만 본다 | 미측정 |
| B4 | `.moai/state/` **밖의** ignore 경로 인용 — `bin/`, `.moai/logs/`, `.moai/cache/`, `.claude/settings.local.json` | 미측정 |
| B5 | **런타임에 조립되는 경로** — 리터럴이 아니라 `filepath.Join(root, dir, name)` 형태면 어떤 문자열 프로브도 못 본다 | 미측정 |
| B6 | **행 경계를 넘는 인용** — 경로와 그것을 근거로 삼는 문장이 다른 줄에 있으면 행 단위 grep 은 관계를 못 본다 | 미측정 |

B2 는 규모가 카드 범위의 18배라 **범위를 넓히면 카드가 다른 카드가 된다** — 넓히지 않고 명시만 한다.

### 1.2 이 카드에서 프로브가 세 번 배제했다

기록해 둘 값이 있다. 셋 다 형태가 같다 — 프로브가 대상의 일부를 조용히 뺐고, 뺀 사실이 결과에 나타나지 않았다.

1. 배차문의 "전수 8건" — 명령에 `| head -8` 이 붙어 있었다. 앞 8줄이 전수로 불렸다.
2. 배차문의 primary 체크아웃 측정(12) vs 판정 트리(13) — `evidence_writer_zeroexec_test.go` 가 `main` 에 아예 없다(t341 이 develop 에만 착지). **측정 트리가 판정 트리가 아니었다.**
3. 이 카드의 1차 보충 정규식 `\.moai/state/[a-z-]+/t[0-9]+` — `/t<숫자>` 를 요구해 `snapshots/abc.json` 형태 3줄을 구조적으로 못 잡았다.

## 2. 분류 — 25줄 전수

축은 **"그 경로가 무엇인가"**다.

| 분류 | 수 | 뜻 |
|---|---|---|
| **C1 기계 소비자** | 6 | 코드가 그 경로를 *정의하거나 소비*한다. 경로가 곧 API |
| **C2 픽스처·샘플** | 4 | 파싱·검사 대상 문자열. 실재를 주장하지 않음 |
| **C3 인용** | 5 | 이미 존재한다고 **주장하는** 특정 산출물 지목 ← **수리 대상** |
| **C4 이 패턴을 지시하는 문서** | 8 | 에이전트에게 그 경로에 쓰라고 **시키는** 지시문 |
| **C5 기제 서술** | 2 | 특정 산출물 없이 기제만 언급 |

### C1 — 기계 소비자 (6) · carve-out

| 위치 | 형태 |
|---|---|
| `internal/verify/store.go:15` | `const SnapshotDir = ".moai/state/verify/snapshots"` — 이 상수가 경로를 정의한다 |
| `internal/verify/schema.go:44` | 위 상수의 doc comment |
| `internal/verify/store_test.go:26,44,45` | 위 상수의 모양 단언 |
| `internal/web/events.go:29` | `"verify": {".moai/state/verify"}` — 감시 대상 지정 |

가드를 분류 없이 넓히면 이 여섯이 오탐이 된다. **런타임에 만들어지는 경로이고, 만들어지기 전에 없는 것이 정상**이다.

### C2 — 픽스처·샘플 (4) · carve-out

| 위치 | 형태 |
|---|---|
| `internal/goal/evaluate_snapshot_test.go:48,71,166` | `attr: "snapshot .moai/state/verify/snapshots/abc.json key … exit 0"` — `fakeSnapshotSource` 에 주입해 verdict 페이로드로 실려 나오는지 보는 **입력 데이터**. 경로가 `abc.json` 이라 실재 파일을 가리키지도 않는다 |
| `internal/session/ignored_content_test.go:26` | `porcelain: "!! .moai/state/verify/t209/\n…"` — git porcelain 출력 **샘플**, 파싱 대상 |

배차문이 "눈으로 본 것"이라 한 goal 3줄은 읽고 판정했다: 실재를 주장하지 않는 데이터다.

### C3 — 인용 (5) · **수리 대상**

| 위치 | 인용 대상 | 본문 자립? |
|---|---|---|
| `internal/cli/mcp_glm.go:110` | `.moai/state/verify/t225/ac-amp-006-glm-differential-attempt1.md` | **자립** — 측정치(3667 vs 3480, budgets 3072 vs 1024, ratio 1.02)가 주석 본문에 있음 |
| `internal/cli/audit_pin_live_test.go:32` | `<repo>/.moai/state/verify/t225/` | 해당 없음 — 산출 위치 진술 (§2.1) |
| `internal/hook/evidence_writer_zeroexec_test.go:10` | `.moai/state/verify/t341/` | **의존** — "raw captures" 만 있고 값이 본문에 없음 |
| `.moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt:3` | `.moai/state/verify/t299-sync/0{2,3,4,5}-*.txt` | **의존** — 원본 실행 로그 |
| `.moai/reports/template-skill-improvement-plan-20260710.html:684` | `.moai/state/verify/eb01063e/skill-audit/*.json` | **의존** — "원시 데이터(JSON)" |

파일 종류가 셋이다: `.go` 3, `.txt` 1, `.html` 1. **`.md` 한정 가드가 못 잡는 것은 `.go` 만이 아니다.**

#### 2.1 `audit_pin_live_test.go:32` — 스스로 반증하는 인용

> Evidence lands in `<repo>/.moai/state/verify/t225/` (the card-scoped verify directory) **so the cited paths still resolve at audit time.**

```
$ git check-ignore -v .moai/state/verify/t225
.gitignore:298:.moai/state/	.moai/state/verify/t225

$ ls .moai/state/verify/
ls: .moai/state/verify/: No such file or directory
```

이 워크트리는 `origin/develop` 에서 방금 팠고, 인용 대상이 **실재하지 않는다**. 주석이 주장하는 성질을 ignore 규칙이 부정한다 — 인용이 끊긴 데 더해 **끊기지 않는다는 진술까지** 함께 실려 있다.

> ignore 규칙 위치는 이 트리에서 **`.gitignore:298`**. 배차문의 `:284` 는 다른 트리 기준이다.

### C4 — 이 패턴을 지시하는 문서 (8)

수리 대상은 아니지만 **이 결함 계열의 생산자**다.

| 위치 | 내용 |
|---|---|
| `internal/template/templates/.codex/agents/moai/manager-lead.toml:54,78,136,139,140,143,150` | 에이전트에게 `.moai/state/verify/$MOAI_SESSION_ID/` 에 검증 출력을 redirect 하라고 지시. **`:143` 은 "The path MUST resolve at audit time" 이라고 명시한다** — ignore 되는 디렉터리를 가리키면서 |
| `.moai/config/sections/workflow.yaml:31` | 같은 기제 언급 |

**`:143` 이 §2.1 과 같은 모순의 원본이다.** `audit_pin_live_test.go:32` 는 이 지시를 따른 결과로 읽힌다 — 코퍼스를 정리해도 지시가 남으면 재생산된다.

**[HARD] 겹침 경보**: 이 `.toml` 은 `internal/template/templates/.claude/agents/moai/manager-lead.md` 에서 `make agents-emit` 으로 **기계 방출**되는 사본이다(CLAUDE.local.md §2.0 — 손편집 금지). 그 `.md` 원본은 **lane-8 의 t375 편집 대상**(지시 2건 중 하나)이다. C4 를 건드리려면 t375 와 같은 파일을 만진다.

### C5 — 기제 서술 (2) · carve-out

| 위치 | 형태 |
|---|---|
| `.moai/reports/moai-autonomy-workflow-redesign-20260803.html:388` | "검증 가능 증거는 `.moai/state/verify/` 에 영속" — 특정 산출물 없음 |
| `.moai/reports/model-tier-redesign-20260712.html:57` | "도입 후 `.moai/state/verify/` 실측으로 보정을 권장" — 특정 산출물 없음 |

## 3. 요약

```
25줄 = C1 기계소비자 6 + C2 픽스처 4 + C3 인용 5 + C4 지시문 8 + C5 기제서술 2
수리 대상: C3 5줄 (.go 3 · .txt 1 · .html 1)
경계 사안:  C4 8줄 — 생산자이며, 원본 .md 가 t375(lane-8) 편집 대상
```

## 4. Gaps / 잔여 위험

- **Gaps**: B1-B6 (§1.1) 전부 미측정. t375 가드 범위 미확인. C3 5건의 수리 방향 미정(plan-phase 소관).
- **잔여 위험**: C3 를 정리해도 **C4 지시문이 남으면 같은 인용이 다시 생산된다**. 코퍼스 정리만으로는 닫히지 않는 축이고, 그 축은 t375 와 같은 파일 위에 있다.
- **잔여 위험**: 본문 의존형 인용(C3 3건)은 경로를 지우면 출처가 사라진다. 지우기·다시 가리키기·추적 위치로 옮기기가 서로 다른 결과를 낳는다.
