# Sync-phase 감사 — SPEC-TODO-DESTRUCTIVE-GUARD-001 (카드 t330)

| 항목 | 값 |
|---|---|
| 감사 대상 트리 | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t330` |
| 브랜치 / HEAD | `WT-todo-destructive-guard` / `7382ce247` |
| 비교 기준 | `origin/develop` = `5e194bba2` (감사 시점 재fetch), merge-base `d566ecc75` |
| 감사자 | sync-auditor (독립 fresh-judgment) |
| 감사 일자 | 2026-08-28 |
| 최종 판정 | **PASS-WITH-DEBT — 조화평균 0.890** (blocking 0건, debt 10건) |

> 감사 중 이 저장소의 실사용 대기열(`.moai/state/todo/`, primary 체크아웃의 큐)은 **한 번도 건드리지 않았다**. 모든 실행 검증은 `mktemp -d` + `git init`으로 만든 격리 저장소에서만 돌렸고, 비교용 구버전 바이너리는 `git archive 812ee01fc`를 `/tmp`로 풀어 별도 빌드했다. 브랜치 상태 변경·커밋·푸시는 없다.

---

## §1 Claim — 이 감사가 주장하는 것

1. **AC 개수는 16개가 맞다.** `acceptance.md`에서 직접 재도출했고, AC-TDG-001~016이 빠짐없이 연속한다. REQ도 16개로 티어 M 상한과 일치한다.
2. **16/16 PASS 주장은 지지된다.** 위험도가 높은 기준 9개를 직접 실행해 재현했고, 새 테스트 25개(kanban 10 + cli 15)가 모두 존재하며 통과한다. `internal/cli` 패키지 전체가 이 트리에서 초록이다.
3. **§A.4의 핵심 측정 3개는 내가 다시 재서 그대로 나왔다.** `origin/main` 0건, `origin/develop` 13건, 그리고 13건 중 **가장 이른 커밋은 `3030df58b`(plan-phase)** — 즉 "ref를 고쳐도 predicate는 구조적으로 항상 참"이라는 spec.md의 결론은 측정으로 뒷받침된다.
4. **잔여 위험 (b)(c)(d)는 참이다.** 구버전 바이너리는 보관 행을 가진 DB를 열고 쓰고, 보관 행은 그 쓰기를 살아남는다(직접 측정). 복원 위치는 즉시 왕복에서 정확하고 이후에는 클램프된다(테스트 + 코드 확인). 동결 테스트 2개는 **완화가 아니라 확장**됐다 — 두 테스트 모두 정확 집합 비교를 그대로 유지한다.
5. **잔여 위험 (a)는 참이며, 내가 세 번째 열화 경로를 추가로 측정했다.** ref가 없는 저장소에서 `--require-landed`는 거부하지 않고 stderr 경고만 남긴 채 **rc=0으로 통과**한다. 명세대로지만, 기계가 읽는 stdout만으로는 "가드가 통과시켰다"와 "가드가 아예 돌지 않았다"를 구별할 수 없다.
6. **sync 산출물에 관측되지 않은 주장 3건이 남아 있다.** `progress.md` §E.1과 `plan.md`가 spec.md v0.2.2가 명시적으로 정정한 틀린 커밋을 아직 그대로 싣고 있고, CHANGELOG의 커버리지 수치 하나와 커버리지 서술 하나가 기록된 측정과 어긋난다. 셋 다 동작에는 영향이 없다.

---

## §2 Evidence — 실행한 명령과 그 출력 그대로

### 2.1 트리 동일성과 범위

```
$ git rev-parse HEAD
7382ce247d5f44e5861cde23a3a77ed1b0b5bbf3
$ git branch --show-current
WT-todo-destructive-guard
$ git fetch origin develop; git rev-parse origin/develop
5e194bba27c146d8c2157d92b4a3fb3995919ff0
$ git merge-base --is-ancestor origin/develop HEAD  → "develop NOT ancestor (develop moved ahead)"
$ git rev-list --count --left-right origin/develop...HEAD
8	13
```

> **범위 주의**: `origin/develop..HEAD`의 diff에는 이 브랜치가 만들지 않은 삭제(예: `.moai/reports/t322/verdict.md`)가 섞여 나온다. develop이 병합 이후 8커밋 전진했기 때문이며, 이 브랜치의 행위가 아니다. 아래 판정은 모두 카드 커밋 13개와 실제 파일 내용을 근거로 삼았다.

### 2.2 AC / REQ 개수 재도출

```
$ grep -oE 'AC-TDG-[0-9]+' acceptance.md | sort -u | wc -l
      16
$ grep -oE 'REQ-TDG-[0-9]+' spec.md | sort -u | wc -l
      16
$ grep -oE '^\*\*AC-TDG-[0-9]+' acceptance.md | sort -u
AC-TDG-001 … AC-TDG-016   (16개, 결번 없음)
```

### 2.3 테스트 — 두 패키지

```
$ go test ./internal/kanban/... -count=1 -cover
ok  	github.com/modu-ai/moai-adk/internal/kanban	12.792s	coverage: 87.0% of statements

$ go test ./internal/cli/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	243.269s
rc=0

$ go vet ./internal/kanban/... ./internal/cli/...
vet rc=0
```

새 테스트 25개 전수 존재 확인(`grep -c "^func Test"`): `internal/kanban/backlog_archive_test.go` 10개, `internal/cli/todo_undone_test.go` 15개. kanban 아카이브 테스트 10개 개별 실행 전부 PASS.

### 2.4 §A.4 측정 재도출 (감사자 독립 실행)

```
$ git log origin/main    --perl-regexp --grep='\bt306\b' --oneline | wc -l
       0
$ git log origin/develop --perl-regexp --grep='\bt306\b' --oneline | wc -l
      13
$ git log origin/develop --perl-regexp --grep='\bt306\b' --oneline --reverse | head -1
3030df58b docs(SPEC-TODO-SQLITE-001): plan-phase artifacts — SQLite queue store + state dir rename (t306)
```

세 값 모두 spec.md v0.2.2가 적은 그대로다. 특히 **가장 이른 커밋이 plan-phase 산출물 커밋**이라는 정정이 사실로 확인됐고, 따라서 "ref를 고쳐도 predicate는 판별력을 잃을 게 아니라 애초에 없다"는 §A.4 결론은 측정으로 지지된다.

### 2.5 격리 저장소 실행 검증 (감사자 작성 픽스처)

두 바이너리를 세워 비교했다.

```
$ go build -o /tmp/t330-new-moai ./cmd/moai              # 현재 트리
$ git archive 812ee01fc | tar -x -C /tmp/t330-old-src
$ (cd /tmp/t330-old-src && go build -o /tmp/t330-old-moai ./cmd/moai)   # 아카이브 이전 바이너리
```

**AC-TDG-008 — `--expect` 불일치 거부 + 바이트 동일성**

```
rc=1 out=Error: mutate backlog …/backlog.json: mutation refused:
  backlog item t2 is "beta work", not matching --expect "zzz"
BYTE-IDENTICAL yes
```

**AC-TDG-009 — inconclusive 경로 (ref 없는 저장소)**

```
rc=0 out=note: --require-landed could not answer for t2
  (prlink: git log origin/main: exit status 128) — proceeding,
  because an unanswerable query is not evidence of not-landed
done t2
```

**AC-TDG-007 — 살아 있는 리더에 보관 행이 안 보이는가**

```
[list]   t1	queued	alpha work
         t3	queued	gamma work
[next]   t1	alpha work
         t3	gamma work
[why t2] t2: no findings          ← 정확히 이 한 줄 (AC가 요구한 exact-output)
[why t3] t3: no findings
[analyze] analyzed 1 pairs, recorded 0 findings
re-add "beta work" -> t4 3        ← 보관 카드와의 중복으로 걸리지 않음
```

**AC-TDG-003 / 015 — export 스트림 분리와 아카이브 적재**

```
STDOUT: exported 2 cards, 0 findings to …/backlog.json
STDERR: note: this export carries 1 archived card(s). A release predating the archive
        reads only the fields it knows, so it will discard them on its first write to
        the queue. Restore anything you still need with `moai todo undone <id>` …
items: ['t1', 't3']
archived: ['t2']
archived position: [1]
archived findings: [('t2','t3',0), ('t1','t2',1)]
live findings count: 0
last_seq: 3
```

가운데 카드(position 1)와 그것을 지칭한 **두 개의** finding이 원래 인덱스와 함께 보관됐고, 살아 있는 findings는 0이 됐다.

**AC-TDG-002 / 016 — 왕복 바이트 동일성 (깨끗한 저장소, 가운데 카드)**

```
STRICT ROUND TRIP BYTE-IDENTICAL: YES
```

복원 후 `undone t2` 재실행:

```
second undone -> rc=1  mutation refused: no archived backlog item t2
items: ['t1', 't2', 't3', 't4']
archived: []
findings: [('t2','t3','contains'), ('t1','t2','conflicts')]
```

**잔여 위험 (b) — 아카이브 이전 바이너리 대 보관 행을 가진 DB**

```
archived rows BEFORE old write: t1|0
OLD list rc=0 -> t2	queued	beta          ← 살아 있는 카드만 보인다
OLD add  rc=0 -> t3 2                        ← 여전히 쓸 수 있다
archived rows AFTER old write: t1|0          ← 보관 행이 구버전 쓰기를 살아남았다
schema_version: 1
NEW undone t1 after old write -> rc=0 undone t1 alpha   ← 이후 복원까지 된다
```

**JSON 강등 경로 — §E.3이 추론으로만 적은 손실을 측정**

```
exported json archived: ['t1']
--- files after OLD write:  backlog.db  backlog.db.stash  backlog.json.migrated  backlog.lock
backlog.json.migrated  has archived key: True  | archived: [t1 …]   ← 원본은 사이드카로 남는다
OLD-created db archived_items: Error: in prepare, no such table: archived_items
--- NEW binary re-reads what OLD left:
 items: ['t2', 't3'] | archived: []          ← 살아 있는 큐에서는 사라졌다
```

즉 stderr 공지 문구("첫 쓰기에서 그 필드를 버린다")는 **참이며 오히려 보수적**이다 — 구버전은 json을 덮어쓰는 대신 `backlog.json.migrated`로 옮기므로 디스크상 원본은 남는다.

**AC-TDG-013 — 재발급 방어를 강제 공격으로 검증**

```
add after archiving t1 -> t2 1                       ← 보관된 id를 재발급하지 않는다
# json의 last_seq를 손으로 0으로 낮추고 db를 치운 뒤
add with last_seq=0 + archived t1 -> t3 2            ← 여전히 t1을 재발급하지 않는다
```

`normalizeBacklogRecord`(`backlog_store.go:764-772`)가 보관 행의 id까지 high-water mark에 반영하므로, 손으로 낮춘 `last_seq`로도 충돌을 만들 수 없었다.

**AC-TDG-012 — 프롬프트 없음 (stdin 닫음)**

```
todo done t99   -> rc=1
todo undone t99 -> rc=1
todo undone     -> rc=1  (Accepts 1 arg(s), received 0)
```

### 2.6 동결 테스트 2개 — 완화인가 확장인가

`git diff abd4fbbbd..HEAD` 로 두 파일의 실제 변경을 읽었다.

`TestBacklogEngineSchemaShape`:

```go
-	want := []string{"findings", "idx_items_state", "items", "meta"}
+	want := []string{"archived_findings", "archived_items", "findings", "idx_items_state", "items", "meta"}
 	if strings.Join(got, ",") != strings.Join(want, ",") { … }
```

비교 연산은 그대로 **정확 집합 일치**다. 원소가 둘 늘었을 뿐 느슨해진 곳이 없다.

`TestTodoVerbSurfaceZeroDelta`: 동결 테이블 `frozenTodoSurface`는 **분기점 상태 그대로 두고**, 추가분을 별도 선언으로 분리했다.

- 플래그 비교: `wantFlags = frozen ∪ permittedFlagAdditions[use]` → `sort.Strings` → `strings.Join` 정확 비교. `dumpTodoSurface`도 `gotFlags`를 정렬하므로 양쪽 모두 정렬된 정확 비교다.
- 개수 검사: `len(live) == len(frozenTodoSurface) + len(permittedVerbAdditions)` — 추가와 삭제가 상쇄되는 구멍이 그대로 막혀 있다.
- 존재 검사: 선언된 추가 verb마다 실제 등록 여부를 확인해 공허한 통과를 막는다.

**판정: 확장이 맞다.** 정확-표면 가드는 보존됐다. 다만 `isPermittedVerbAddition`이 `strings.HasPrefix`를 쓰는 점은 남는다(F9, 이 카드가 도입한 것은 아님 — `theOnlyPermittedAddition`이 이미 같은 형태였다).

### 2.7 템플릿 미러 · 문서

```
$ cmp .claude/skills/moai/workflows/todo.md internal/template/templates/.claude/skills/moai/workflows/todo.md
cmp rc=0
$ grep -c undone (양쪽)  → 4 / 4
$ grep -nE 'SPEC-[A-Z0-9-]+|REQ-[A-Z]|AC-[A-Z]|\bt3[0-9][0-9]\b|20[0-9][0-9]-[0-9][0-9]-[0-9][0-9]|[0-9a-f]{9,40}' (템플릿 사본)
40:… `moai todo next <n> [--spec <SPEC-ID>]` …
176:… `moai todo next <n> --spec <SPEC-ID>` …
```

두 히트 모두 `<SPEC-ID>` **플레이스홀더**다. 실제 SPEC ID·카드 id·내부 날짜·커밋 SHA·`CLAUDE.local` 참조는 0건 — 템플릿 중립성 통과.

docs-site 4개 로케일:

```
$ wc -l  ko/en/ja/zh utility-commands/moai-todo.md   → 227 / 227 / 227 / 227
$ grep -c '^## '                                     → 9 / 9 / 9 / 9
$ grep -c 'undone'                                   → 4 / 4 / 4 / 4
```

4개 파일 모두 같은 5곳(`state` 필드 설명 · 사용 예시 3블록 추가 · `done` 행 · 새 `undone` 행 · `export-json` 행)이 동일하게 바뀌었다. ko 원본은 자연스러운 한국어 문어체이고 en/ja/zh는 ko 파생으로, 계산어투(calque)가 보이지 않는다.

**코드 대조**: 문서가 적은 동작을 Go 코드로 확인했다 — `done`은 `rec.ArchiveCard(id)`를 부르고(`todo.go`), `undone`은 `rec.RestoreCard(id)`를 부르며 아카이브 항목을 비우고(`backlog_store.go:278`), `export-json`은 아카이브를 포함해 마샬한 뒤 `discloseArchiveDowngradeCost`가 **아카이브가 실제로 있을 때만** stderr에 경고한다(`todo_export.go`). 문서 서술과 코드가 일치한다.

### 2.8 run-phase 증거 파일 대조

```
$ cat .moai/reports/t330/evidence/cover-cli.txt
ok  	github.com/modu-ai/moai-adk/internal/cli	300.387s	coverage: 79.8% of statements
$ cat .moai/reports/t330/evidence/cover-kanban.txt
ok  	github.com/modu-ai/moai-adk/internal/kanban	13.463s	coverage: 87.0% of statements
$ cat .moai/reports/t330/evidence/lint.txt
0 issues.
```

`evidence/ac005.sh` / `ac005.txt`는 구버전 바이너리 강등 검증을 재실행 가능한 형태로 보존하고 있다. 내가 §2.5에서 같은 절차를 독립적으로 다시 세워 같은 결론을 얻었다.

---

## §3 Baseline-attribution — 무엇에 대고 쟀나

| 측정 | 기준 |
|---|---|
| 테스트·vet·빌드 | 이 워크트리, HEAD `7382ce247`, 2026-08-28 이 세션 |
| `origin/main` / `origin/develop` grep 3건 | 감사 시점 재fetch한 `origin/develop = 5e194bba2` |
| 구버전 바이너리 비교 | `git archive 812ee01fc` → `/tmp/t330-old-src` → 별도 빌드. `acceptance.md`가 지정한 base tree와 같은 커밋 |
| 격리 실행 | `mktemp -d` 6개, 각각 `git init` + `CLAUDE_PROJECT_DIR` 고정. 실사용 큐 접촉 0 |
| 커버리지 | 이 세션 kanban 87.0% 직접 실측. cli 커버리지는 **재측정하지 않았다**(§4 Gaps) |
| 동결 테스트 판정 | `git diff abd4fbbbd..HEAD` 로 읽은 실제 diff |

---

## §4 Gaps — 관측하지 않은 것

- **`internal/cli` 커버리지를 다시 재지 않았다.** 이 세션은 `-cover` 없이 전체 패키지를 돌렸다(243초, rc=0). 79.8% 대 79.9%의 불일치(F2)는 *기록된 증거끼리의* 불일치로 판정한 것이지, 내가 세 번째 값을 잰 것이 아니다.
- **전체 스위트를 돌리지 않았다.** `./internal/kanban/...`과 `./internal/cli/...`만 실행했다. 모듈 전체 판정은 CI 몫이며, 이 브랜치는 아직 푸시되지 않아 **CI 실행 자체가 존재하지 않는다**.
- **Windows / Linux에서 컴파일하거나 실행하지 않았다.** 크로스 빌드조차 이 세션에서 다시 돌리지 않았다(run-phase의 `GOOS=windows go build rc=0` 기록을 그대로 둔다).
- **`-race`를 돌리지 않았다.**
- **docs-site를 hugo로 빌드하지 않았다.** 4개 로케일 대조는 행수·헤딩수·문자열 grep과 diff 판독까지이며, 렌더 결과는 보지 않았다.
- **실제 릴리스 바이너리(v3.0.x / v3.1.x)로 강등을 시험하지 않았다.** `812ee01fc` 트리에서 빌드한 바이너리만 썼다.
- **`moai spec lint` / `moai spec audit`을 돌리지 않았다.**
- **cross-model(codex/glm) 2차 의견을 구하지 않았다.** 이 카드에 `audit_model: multi` 지시가 없었고, 위임 프롬프트도 요구하지 않았다.
- **아카이브가 커진 상태의 성능을 실측하지 않았다.** F4는 코드 경로 판독에서 나온 것이지 벤치마크가 아니다.

---

## §5 Residual-risk — 관측한 것에도 불구하고 여전히 틀릴 수 있는 것

1. **아카이브가 커진 뒤의 쓰기 비용을 아무도 재지 않았다.** F4의 O(archive) 쓰기 증폭은 코드상 확실하지만, 카드 수백 장 규모에서 체감 지연이 생기는지는 미지수다.
2. **JSON 강등의 손실 폭은 구버전 구현에 의존한다.** 내가 잰 것은 `812ee01fc` 바이너리 하나의 동작(`backlog.json.migrated` 사이드카 보존)이다. 더 오래된 릴리스가 json을 제자리에서 덮어쓴다면 디스크상 원본조차 남지 않는다.
3. **`--require-landed`의 실질적 무해함이 오히려 위험할 수 있다.** ref가 없는 환경에서는 조용히 통과하고, ref가 있는 이 프로젝트에서는 구조적으로 항상 참이다(§2.4). 즉 이 플래그는 **현재 어떤 환경에서도 실질적 판별을 하지 못한다**. 문서와 help가 이를 밝히고 opt-in이라 blocking은 아니지만, "가드를 켰다"는 심리적 안전감 자체가 t306 사고의 재발 벡터다. t331 없이 이 플래그를 신뢰하는 운영자는 정확히 같은 방식으로 속는다.
4. **`export-json`이 새 `archived` 키를 쓴다.** 알려진 제3자 소비자는 없지만 산출물 모양이 바뀐 것은 사실이다.
5. **아카이브는 무한히 자란다.** 범위 밖 선언(spec.md §D)이며 doctrine에는 적혀 있으나 docs-site에는 없다(F8).
6. **`plan-audit` iter3가 닫았다는 인용 정정(D7/S1) 계열의 실패가 또 살아남았다.** F1이 바로 그 형태다 — 보고서를 다 읽고 나서 산출물 하나만 고치고 나머지를 놓치는 패턴이 세 번째로 재현됐다.

---

## §6 4-Dimension Scores

| 차원 | 점수 | 판정 | 근거 (실측 출력) |
|---|---|---|---|
| **Functionality (40%)** | 0.95 | PASS | `go test ./internal/cli/ -count=1` → `ok … 243.269s`, rc=0 · `go test ./internal/kanban/...` → `ok … coverage: 87.0%` · AC 16개 재도출(`grep -oE 'AC-TDG-[0-9]+' … wc -l` → `16`) · 고위험 기준 9개 격리 실행 재현(§2.5), 특히 `STRICT ROUND TRIP BYTE-IDENTICAL: YES` 와 `archived rows AFTER old write: t1\|0` · §A.4 측정 3개 독립 재도출(`0` / `13` / `3030df58b`) |
| **Security (25%)** | 0.92 | PASS | `LandedGrepArgs`가 `validCardToken.MatchString(cardID)`로 검증 후 **argv 형태**로 실행(셸 경유 없음) → 주입 표면 없음 · 모든 거부 경로에서 `BYTE-IDENTICAL yes` · `writeExportAtomic`은 동일 디렉터리 temp + rename · 아카이브는 기존 `Mutate` 락 안에서만 갱신 · Critical/High 0건. 감점 사유는 가용성 계열 2건(F6 fail-open이 stdout으로 구별 불가, F7 락 보유 중 무타임아웃 서브프로세스) |
| **Craft (20%)** | 0.82 | — | kanban 87.0%(목표 상회), cli 79.8%(85% 미달, 이 변경 귀속 아님) · `go vet` rc=0, `golangci-lint 0 issues` · 동결 테스트 2개 **정확 비교 보존 확인**(§2.6) · 감점: F4 쓰기 증폭 미공개, F5 죽은 코드 + 낡은 주석, F9 `HasPrefix` 허용폭 |
| **Consistency (15%)** | 0.88 | — | `cmp` rc=0 + 템플릿 중립성 위반 0건 · docs-site 4로케일 `227/227/227/227`, `9/9/9/9`, `4/4/4/4` 및 ko-canonical 원어 문체 · 문서 서술을 Go 코드로 대조 확인 · 감점: F1 정정된 사실이 두 산출물에 잔존, F2·F3 CHANGELOG 수치·서술 불일치, F8 docs-site 보존정책 누락 |

**Must-pass 방화벽**: Functionality 0.95 ≥ 0.80 PASS · Security 0.92 ≥ 0.80 PASS. 필수 통과 차원 모두 독립적으로 임계를 넘었으므로 방화벽에 걸리지 않는다.

**조화평균**

```
4 / (1/0.95 + 1/0.92 + 1/0.82 + 1/0.88)
= 4 / (1.052632 + 1.086957 + 1.219512 + 1.136364)
= 4 / 4.495465
= 0.8898
```

### 최종 판정 — **PASS-WITH-DEBT (0.890)**

blocking 0건. 셋 이상의 독립 경로(격리 실행, 구버전 바이너리 대조, 코드 판독)에서 명세가 약속한 동작이 확인됐고, 동결 가드가 실제로 보존됐으며, 잔여 위험 4건이 모두 참으로 검증됐다. 아래 debt 10건은 병합을 막지 않는다.

---

## §7 Findings

### F1 [DEBT · 중] — spec.md가 정정한 틀린 커밋이 다른 두 산출물에 그대로 남아 있다

`spec.md` v0.2.2가 명시적으로 정정했다:

> 정정: the earliest of the 13 ref-corrected matches is the plan-phase artifacts commit `3030df58b`, **not** the run merge `3cb258d62` (tenth of 13) — the prior wording named the wrong commit as earliest.

그런데 정정 이전 문장이 두 곳에 살아 있다.

- `progress.md:25` — "`origin/develop` names t306 in 13 commits, **the earliest being the run commit `3cb258d62`**"
- `plan.md:52` — "true — 13 matching commits, **earliest the run commit `3cb258d62`**"

내가 직접 재서 `--reverse | head -1` 이 `3030df58b`를 돌려준다(§2.4). 즉 두 문장은 **측정으로 반증된다**. spec.md가 권위 문서라 결론은 바뀌지 않지만, `progress.md`는 sync-phase가 갱신한 문서이고 §E.4 블록을 담은 바로 그 파일이다 — 감사자가 §E.1만 읽으면 반증된 사실을 인용하게 된다.

**요구되는 수정**: `progress.md:25`와 `plan.md:52`의 `3cb258d62`를 `3030df58b`(plan-phase 산출물 커밋)로 바꾸거나, spec.md §A.4를 참조하도록 문장을 축약한다.

### F2 [DEBT · 하] — CHANGELOG의 `internal/cli` 커버리지 수치가 기록된 측정과 다르다

CHANGELOG: `internal/cli` **79.9%**.
`progress.md` §E.3: `coverage: 79.8%`.
`evidence/cover-cli.txt`: `coverage: 79.8% of statements`.

이 트리의 어떤 기록도 79.9%를 지지하지 않는다. 귀속 없는 수치다.

**요구되는 수정**: CHANGELOG를 79.8%로 정정한다(또는 79.9%를 낳은 실행 로그를 evidence에 추가한다).

### F3 [DEBT · 하] — CHANGELOG의 커버리지 서술이 §E.3 자신의 수치보다 낙관적이다

CHANGELOG: "the new code itself is fully covered **except two stream-write-failure arms**".
`progress.md` §E.3이 실제로 나열한 미달 함수는 넷이다 — `todoWriteLine` 66.7%(스트림 쓰기), `ArchiveCard` 93.8%, `readArchive` 76.9%(SQL 실패), `writeArchive` 66.7%(SQL 실패). 즉 미커버 갈래는 스트림 쓰기 계열만이 아니다.

**요구되는 수정**: "두 개의 스트림 쓰기 갈래" → "스트림 쓰기 갈래와 SQL 실패 갈래"로 서술을 §E.3과 맞춘다.

### F4 [DEBT · 중] — 모든 큐 변경이 아카이브 테이블 전체를 지우고 다시 쓴다

`backlog_migrate.go:292`의 `e.writeArchive(ctx, tx, rec)`는 일반 쓰기 경로(`writeRecord`) 안에 있고, `writeArchive`는 매번 다음을 한다:

```go
DELETE FROM archived_items
DELETE FROM archived_findings
for i, entry := range rec.Archived { INSERT … }   // 전량 재삽입
```

따라서 `todo add` 한 번의 비용이 **완료된 카드 수에 비례**해 커진다. spec.md §D는 "아카이브가 무한히 자란다"를 범위 밖으로 선언했지만, 그 성장이 *모든 쓰기*에 부과되는 비용이라는 점은 §D에도 §E.3 잔여 위험에도 적혀 있지 않다. 읽기 경로(`readArchive`)도 매 open마다 아카이브 전체를 메모리로 올린다.

**요구되는 수정**: 잔여 위험에 이 쓰기 증폭을 명시하고, 보존 정책(t331 또는 후속 카드)의 동기로 기록한다. 코드 변경은 이 카드의 범위 밖이 맞다.

### F5 [DEBT · 하] — `RemoveFindingsNaming`이 프로덕션에서 죽었는데 주석은 여전히 호출된다고 말한다

`grep -rn "RemoveFindingsNaming" internal/` 결과 프로덕션 호출자는 0건이고, `backlog_findings_test.go`만 부른다. 그런데 `backlog_store.go:352-355`의 주석은 아직:

> Called when a card leaves the queue: a finding that outlives its subject points at nothing…

이제 카드가 큐를 떠날 때 호출되는 것은 `ArchiveCard`다. 테스트 파일의 서두 주석(`backlog_findings_test.go:122`)도 같은 낡은 계약을 반복한다.

**요구되는 수정**: 주석을 "현재 프로덕션 호출자 없음 — `ArchiveCard`가 대체함, 레거시/외부 재사용을 위해 보존"으로 고치거나, 삭제 여부를 운영자 판단에 올린다. (범위 규율상 임의 삭제는 하지 않는 것이 맞다.)

### F6 [DEBT · 중] — `--require-landed`의 fail-open이 기계가 읽는 표면에서 구별되지 않는다

격리 저장소(§2.5)에서 측정:

```
rc=0
stderr: note: --require-landed could not answer for t2 (prlink: git log origin/main: exit status 128) — proceeding …
stdout: done t2
```

`todo.go:20-22`가 "verb당 stdout 한 줄"을 기계 판독 표면으로 계약하고 있는데, 그 한 줄은 가드가 통과시킨 경우와 **가드가 아예 돌지 않은 경우**가 완전히 동일하다. 구별 정보는 stderr에만 있다. REQ-TDG-009 그대로이고 doctrine에도 밝혀져 있으므로 결함은 아니지만, 무인 레인이 `--require-landed`를 켜고 stdout만 파싱하면 아무 보호도 못 받으면서 받았다고 믿는다.

**요구되는 수정**: 후속 카드(t331)에서 stdout 한 줄에 판정 토큰(`done t2 landed=yes|unknown`)을 싣는 안을 검토한다. 이번 카드에서 고칠 필요는 없다.

### F7 [DEBT · 하] — 큐 락을 쥔 채 타임아웃 없는 서브프로세스를 돌린다

`todoRequireLanded`는 `store.Mutate(...)` 콜백 **안에서** 호출되므로, `git log`가 도는 동안 backlog advisory lock이 잡혀 있다. 락은 여러 레인이 공유하는 직렬화 지점이고, `GitLandedQuerier.Landed`에는 컨텍스트 타임아웃이 없다. 로컬 `git log`라 실질 위험은 작고 플래그는 opt-in이지만, 거부를 mutation 안에 두어 바이트 동일성을 얻는 설계의 대가로 기록해 둘 값어치가 있다.

**요구되는 수정**: 잔여 위험에 한 줄 기록. 대안(락 밖 선행 질의)은 REQ-TDG-011의 바이트 동일성 보장을 약화시키므로 권하지 않는다.

### F8 [DEBT · 하] — docs-site에 아카이브 보존 정책 경고가 없다

`.claude/skills/moai/workflows/todo.md`는 "The archive is not pruned: it grows, and a retention policy is the operator's decision"를 적었지만, docs-site 4로케일의 `moai-todo.md`에는 대응 문장이 없다. docs-site만 읽는 사용자는 무한 성장을 알 수 없다.

**요구되는 수정**: 4로케일 `state`/`export-json` 인근에 한 문장 추가(ko 원본 → en/ja/zh 파생).

### F9 [DEBT · 하] — 동결 가드의 추가 허용이 접두사 일치라 넓다

`isPermittedVerbAddition`은 `strings.HasPrefix(use, added)`를 쓰므로, 장래 `undone-all` 같은 verb가 선언 없이 통과한다. 이 카드가 도입한 결함은 아니고(`theOnlyPermittedAddition`도 같은 형태였다) 확장 과정에서 그대로 옮겨졌을 뿐이다.

**요구되는 수정**: 첫 토큰 완전 일치(`strings.Fields(use)[0] == added`)로 좁힌다. 선택 사항.

### F10 [DEBT · 하] — §E.4의 self-test 귀속이 얇다

`b12_self_test_c`는 "`ls` on all 6 implementation files + 4 test files + 2 doctrine files … all 12 resolved"라고만 적고, 실행한 명령의 출력을 남기지 않았다. 내가 독립적으로 12개 경로의 존재를 확인했으므로 **주장 자체는 참**이지만, VCI §3.2가 요구하는 "명령 + 그 출력 그대로"에는 미달한다(`b12_self_test_a`/`b`는 명령 형태를 남겨 기준을 만족한다).

**요구되는 수정**: 없음 또는 §E.4에 해당 `ls` 출력 경로를 덧붙인다.

---

## §8 병합 권고

**병합 가능.** blocking 0건이며, 필수 통과 차원 두 개가 독립적으로 임계를 넘었다.

병합 전에 값싸게 닫을 것을 권하는 것은 **F1·F2·F3 셋뿐**이다 — 셋 다 마크다운 한두 줄 수정이고, 세 문서(progress.md / plan.md / CHANGELOG)가 측정과 어긋난 채로 저장소에 들어가는 것을 막는다. 나머지 일곱은 후속 카드 재료이며, 특히 F4(쓰기 증폭)와 F6(fail-open 구별 불가)은 t331이 착지 판정을 다시 설계할 때 함께 읽혀야 한다.

`--require-landed`에 대해서는 한 가지를 분명히 남긴다. 이 감사에서 잰 바로는, 이 플래그는 ref가 없는 환경에서 조용히 통과하고 ref가 있는 이 프로젝트에서는 구조적으로 항상 참이다. 즉 **현재 실질적 판별력이 0이다.** SPEC이 이를 정직하게 문서화하고 opt-in으로 배송한 것은 옳은 판단이지만, 운영 문서 어디에도 "이 플래그를 켜는 것이 t306류 사고를 막지 못한다"는 문장이 한 줄로 서 있지는 않다. t331이 착지하기 전까지는 그 한 줄이 있는 편이 안전하다.

---

*이 보고서의 모든 실행 검증은 격리 저장소에서 수행됐으며, 이 저장소의 실사용 백로그 큐는 감사 전 과정에서 읽지도 쓰지도 않았다. 감사자는 이 파일 외의 어떤 파일도 수정하지 않았고, 커밋·푸시·머지를 수행하지 않았다.*
