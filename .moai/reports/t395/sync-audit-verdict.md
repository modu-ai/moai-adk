# Sync-Audit Verdict: SPEC-BACKLOG-JSON-DISCLOSURE-001

카드: t395 · 트리 `.claude/worktrees/t395` · 브랜치 `WT-stale-backlog-json` @ `e8a30fa10`
origin/develop(`65196a5a7`) 대비 divergence `0 11` — 지시받은 핀과 일치, 감사 중 HEAD 이동 없음
측정 시점: 2026-09-02 · 플랫폼: darwin/arm64 · 측정자: sync-auditor (독립 재측정)

**Overall Verdict: FAIL**
**Overall Score: 89.9 / 100**

FAIL 사유는 단 하나 — D1, CHANGELOG에 실린 커버리지 문장이 **측정으로 거짓**이고 이 SPEC 자신의
§E.2 표가 그것을 반박한다. 코드는 고칠 것이 없고, 16개 AC 전부와 must-pass 두 차원(Functionality /
Security) 모두 통과한다. 남은 것은 한 문장 정정이며, 그 뒤 재감사는 D1 델타로만 좁혀도 된다.

이 판정은 lane의 §E.2를 추인한 것이 아니라, 아래 모든 값을 이 트리에서 직접 다시 재서 얻은 것이다.

---

## Dimension Scores

| Dimension | Score | Verdict | Evidence (이 실행에서 관측한 값) |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | `go test ./internal/kanban/ -run TestForemanQueueWatch -v` → 6/6 PASS, `ok internal/kanban 120.755s`. 실제 바이너리로 read verb 6종 재현: stdout 6/6 byte-identical, stderr 공개 1행/verb (json 있을 때), 0행 (없을 때) |
| Security (25%) | 95/100 | PASS | 공개 경로 read-only 확인: `0o400` DB에서 inspector 정상 반환 + lock 파일 미생성(`TestInspectArchiveVouch_ReadOnlyDatabase` PASS). backlog.json sha256 불변(`2e12bdc7…efc5` 전후 동일). 공개 문자열에 사용자 입력·경로·비밀값 보간 없음(상수 `vouch.Store`만) |
| Craft (20%) | 85/100 | PASS | `golangci-lint run --timeout=5m ./internal/kanban/... ./internal/cli/ ./internal/template/` → `0 issues.` · `go vet` exit 0 · 커버리지 kanban **86.5%** / cli **80.3%** / template **86.3%** · `todo_disclosure.go` 두 함수 **100.0%**. 감점: D2(항진 검사), D1이 드러낸 함수 커버리지 바닥 72.7% |
| Consistency (15%) | 82/100 | PASS | 미러 파리티 확인(`workflows/todo.md`·foreman `SKILL.md` local↔template **byte-identical**), catalog.yaml 해시 2건 재생성 반영, `@MX:ANCHOR fan_in=3` 존재. 감점: D1(공개 산출물의 거짓 수치), D4(한계 문단 누락) |

Must-pass firewall (Functionality + Security): **둘 다 통과** — firewall은 발동하지 않았다.
FAIL은 오로지 blocking 결함 D1이 만든 것이다.

---

## AC 매트릭스 — 감사자가 직접 잰 값

| AC | Sev | 판정 | 이 실행의 근거 |
|---|---|---|---|
| AC-BJD-001 | SHOULD | PASS | `NonAuthoritativeJSON` 이 `BacklogArchiveVouch` **필드**로 존재(`backlog_archive_vouch.go:41-53` 직접 확인). 4개 분기 테스트(StateD / SteadyState / LegacyJSON / NoStore) 통과 |
| AC-BJD-002 | **BLOCKING** | PASS | 실바이너리 재현. `list --json` / `list` / bare / `pr --json` / `why t1` / `history` 각 **stderr 공개 1행**. 문구: `todo: answered by the SQLite backlog store; the backlog.json beside it is NOT the queue — an export or a legacy leftover, whose contents can be arbitrarily stale` |
| AC-BJD-003 | **BLOCKING** | PASS | **직접 측정, 주장 수용 아님.** clean 대비 json-present stdout `cmp -s`: listjson IDENTICAL(285B) · list IDENTICAL(66B) · bare IDENTICAL(66B) · prjson IDENTICAL(76B) · why IDENTICAL(16B) · hist IDENTICAL(17B). stdout 내 공개 문자열 0건. `list --json` stdout은 여전히 유효 JSON이며 `last_seq:2`(DB) — 스테일 json의 `last_seq:9`가 아님 |
| AC-BJD-004 | SHOULD | PASS | json 부재 시 6종 모두 out=0 err=0 |
| AC-BJD-005 | **BLOCKING** | PASS | read surface 전체 실행 후 backlog.json sha256 전후 동일 |
| AC-BJD-006 | SHOULD | PASS | `0o400` DB에서 정상 반환, lock 파일 미생성 |
| AC-BJD-007 | SHOULD | PASS (범위 한정 — 아래 §"universal negative" 참조) | 합산 count **1**, 유일 non-zero 파일 `backlog_archive_vouch.go`. 상수 3개 모두 `:30,32,34`. 필드가 struct 내부에 있음 |
| AC-BJD-008 | **BLOCKING** | PASS | `FiresOnMutation/local` `/template` 둘 다 PASS. **falsifiability 실증**: `ShippedJSONTargetIsSilent` 가 같은 fixture·같은 창·같은 mutation 에서 침묵 → 두 타깃 모두에 통과하는 공허한 검사가 아님 |
| AC-BJD-009 | SHOULD | PASS | `FiresWithStaleJSONPresent` PASS — r1-repro.md 가 "추론했으나 관측 못함"으로 남긴 Case B 가 닫힘 |
| AC-BJD-010 | SHOULD | PASS | Given이 **실제로 구축·증거화**됨: `backlog.db-wal = 12392 bytes`, `cksum(backlog.db)` 창 전후 모두 `1190093710 40960` 로 불변. 미구축 시 `t.Fatalf("GAP …")` 로 실패하며 조용히 통과하지 않음 |
| AC-BJD-011 | SHOULD | PASS | 두 블록 byte-identical, `backlog.json` 미포함, `backlog.db` 포함 |
| AC-BJD-012 | **BLOCKING** | PASS (falsifiability 확인) | 두 grep 모두 zero. **미수리 트리에서 실패함을 확인**: `git grep … origin/develop` → 패턴1이 3건(SKILL.md:170, todo.md:17, foreman:95), 패턴2가 1건(todo.md:21). 즉 자기 범위 안의 항목을 실제로 본다 |
| AC-BJD-013 | SHOULD | PASS | 세 문장 모두 `.moai/state/todo/backlog.db` / 홈폴백 `…/backlog.db` 를 명명 |
| AC-BJD-014 | SHOULD | PASS | `git diff --exit-code origin/develop...HEAD -- .moai/docs/todo-queue-storage.md` → CLEAN |
| AC-BJD-015 | **BLOCKING** | PASS (falsifiability 확인) | 패턴1 정확히 1건(`todo-queue-storage.md:55` export 통제), 패턴2 0건. **미수리 트리 대비 확인**: develop 에서 패턴1 4건·패턴2 1건 → 두 패턴이 함께 있어야 4개 사이트를 덮는다는 주장이 측정으로 성립 |
| AC-BJD-016 | SHOULD | PASS (사실), **검증 수단은 결함** — D2·D3 | 실체는 참: 바이너리에 수리된 블록 존재(`grep -a -o 'cksum "$d"/backlog.db…'` 히트), 구 블록 `f=.moai/state/todo/backlog.json` **0건**. 그러나 §E.2가 인용한 두 검증 명령 **모두** 판별력이 없다 |

**16/16 PASS · GAP 0 · FAIL 0** — darwin/arm64 기준. Windows 에서 AC-BJD-008/009/010 은
`requirePOSIXWatchTools` 의 `t.Skipf` 로 **미측정이며 통과가 아니다**(CHANGELOG 서술 정확).
AC-BJD-011 은 `requirePOSIXWatchTools` 를 부르지 않아 Windows 에서도 실행되며, CHANGELOG 가
008/009/010 만 지목한 것도 정확하다.

### 전체 패키지 스위트 (감사자 실행)

| 패키지 | 결과 |
|---|---|
| `internal/kanban` | `ok 139.084s coverage: 86.5%` |
| `internal/cli` | `ok 492.127s coverage: 80.3%` |
| `internal/template` | `ok 30.360s coverage: 86.3%` |

`go test ./...` 는 돌리지 않았다(CLAUDE.local.md §4/§6). develop 의 `internal/web` i18n 2건은
측정하지 않았고 어느 쪽으로도 귀속하지 않는다.

---

## 리드가 지목한 6개 압박점 — 응답

**1. AC-BJD-008 의 falsifiability는 진짜인가.**
진짜다. `ShippedJSONTargetIsSilent` 는 구 블록을 Go 상수로 **핀**해 두고 같은 fixture·창·mutation 에서
침묵함을 보인다(PASS). fixture 에는 `backlog.json` 이 없으므로 구 스크립트는 매 반복 `cur=missing` 이고
`last` 가 초기화 이후 움직이지 않는다 — 구조적으로 발화 불가. 누군가 SKILL.md 의 타깃을 되돌리면
`FiresOnMutation` 이 붉어진다(같은 fixture에서 이미 관측된 침묵이 그 증거다). 이름만 믿지 않았다.

**2. AC-BJD-010 의 Given은 가정인가 구축인가.**
구축이며 증거화된다. 로그가 값을 남긴다: `backlog.db-wal = 12392 bytes`, `cksum(backlog.db)` 가
창 **전과 후 모두** `1190093710 40960` 로 불변. 미구축이면 `t.Fatalf("GAP — the Given was not built…")`
로 **실패**하지, 조용히 통과하지 않는다. plan-audit 이 제안했던 side-connection `wal_autocheckpoint=0`
기법은 실제로 틀렸고(해당 pragma는 per-connection), 대체된 기법은 **쓰는 연결 자신이** `wal_autocheckpoint(0)`
을 걸고 열린 채로 남는 방식이며 — 위 두 증거값이 그것이 작동했음을 보인다.
추가로 `DBOnlyTargetMissesWALDeferral` 통제가 PASS 라서, `backlog.db-wal` 을 함께 보는 것은
예방책이 아니라 **측정된 요구**다(naive `cksum backlog.db` 수리는 다른 침묵으로 갈아탈 뻔했다).

**3. AC-BJD-007 의 universal negative — 무엇을 세우고 무엇을 못 세우나.**
세우는 것: `internal/kanban/*.go` 안에서 `func Inspect.*Backlog.*(Store|Vouch)` **이름 모양**에
맞는 함수가 정확히 하나이고, 세 상수가 그 파일에만 있으며, 새 사실이 **별도 반환 타입이 아니라
struct 필드**라는 것.
세우지 **못하는** 것: "두 번째 store-identity 프로브가 어디에도 없다"는 명제. 다른 이름
(`probeQueueLayout` 등)이거나 `internal/cli` 등 다른 디렉터리에 있으면 이 grep 은 보지 못한다.
다만 이 SPEC 범위 안에서는 코드를 직접 읽어 보강했다 — `todo_disclosure.go` 는 자체 `os.Stat` 이
없고 `kanban.InspectBacklogArchiveVouch` 만 호출하며, `todo_history.go` 는 이미 손에 든 vouch 를
재사용한다(새 프로브 0). 즉 실질 주장은 참이지만, **grep 이 그것을 증명한 것은 아니다**.

**4. AC-BJD-012 / AC-BJD-015 는 미수리 트리에서 실패하는가.**
그렇다. develop 트리에 같은 두 쌍의 grep 을 걸어 확인했다(위 AC 표). 특히 패턴1은 홈폴백 사이트
`todo.md:21` 을 못 보고, 패턴2가 그것을 잡는다 — "한 패턴이면 4개 중 3개만 검사한다"는 SPEC 의
주장은 이 트리에서 **재현되는 측정**이다. 자기 범위 안의 항목을 보지 못하는 완결성 검사가 아니다.

**5. stdout 오염 — 주장 수용하지 않고 직접 쟀다.**
git 메타데이터를 갖춘 격리 fixture에서 실제 바이너리로 clean/json-present 두 벌을 받아
`cmp -s` 로 비교했다. `moai todo list --json` 포함 6종 전부 byte-identical(285/66/66/76/16/17 bytes).
`list --json` 의 stdout 은 유효 JSON 이며 DB 내용(`last_seq:2`, t1/t2)을 담고 스테일 json
(`last_seq:9`, t9)을 담지 않는다 — foreman 이 읽는 기계 표면은 오염되지 않았고, 답한 store 도 DB 다.
(lane 의 테스트는 `list` 만 돌리고 `list --json` 을 돌리지 않는다. 결과는 같지만, 리드가 지목한
정확히 그 표면은 lane 의 자동 검사 범위 밖이었다 — 지금 이 감사가 그 공백을 메운다.)

**6. CHANGELOG 의 커버리지·린트 수치는 이 트리에서 성립하는가.**
린트는 성립한다(`0 issues.`). 패키지 커버리지도 실질 성립한다(kanban 86.5% 일치, template 86.3% 일치,
cli 80.2% 기록 대 **80.3%** 실측 — 0.1pt 는 실행 간 변동 범위, 결함 아님).
**함수 단위 문장은 성립하지 않는다** → D1.

---

## Findings

### D1 — [MAJOR] [blocking] CHANGELOG 의 함수 커버리지 문장이 거짓이며, 이 SPEC 자신의 표가 반박한다

`CHANGELOG.md` (Unreleased → Added, t395 항목 마지막 문단):

> package coverage kanban 86.5% / cli 80.2% / template 86.3%, **with every changed function at or
> above 86.7%** and the new `todo_disclosure.go` at 100%.

`runTodoHistory` 는 이 SPEC 이 **바꾼 함수**다(`internal/cli/todo_history.go` 에 5행 추가 —
disclosure 호출). 이 트리에서 실측:

```
$ go tool cover -func=<cli profile>
internal/cli/todo_disclosure.go:35  discloseNonAuthoritativeBacklogJSON  100.0%
internal/cli/todo_disclosure.go:55  discloseQueueLayout                  100.0%
internal/cli/todo.go:310            runTodoList                           93.2%
internal/cli/todo_pr.go:118         runTodoPR                             88.1%
internal/cli/todo_why.go:20         newTodoWhyCmd                         86.7%
internal/cli/todo_history.go:77     runTodoHistory                        72.7%   ← 86.7% 미만
```

`progress.md` §E.2 의 "coverage (changed code)" 행이 **이미 72.7% 를 적고 있다**. 즉 산출물이
자기 자신과 모순한다. 같은 거짓 문장이 §E.2 "Gaps" 절에도 반복된다("Every function this SPEC
changed is at or above 86.7%") — 발원지는 §E.2, 전파처가 CHANGELOG 다.

왜 blocking 인가: 공개되는 릴리스 노트가 **측정 가능하게 거짓인 정량 주장**을 싣고 있고,
`verification-claim-integrity.md` §1 이 금하는 미관측 주장에 해당한다. 코드 결함은 아니다.

**Required fix**: CHANGELOG 의 해당 절을 실측대로 고친다 — 예: "changed functions 72.7–100%
(`runTodoHistory` 72.7%, 새 `todo_disclosure.go` 100%)" — 또는 해당 종속절을 삭제한다.
`progress.md` §E.2 Gaps 의 같은 문장도 함께 정정한다. 한 문장 편집이며 코드 변경은 필요 없다.

### D2 — [MINOR] [optional] AC-BJD-016 의 Go 검사는 항진명제다 (뮤테이션으로 증명)

`TestBacklogJSONDisclosure_EmbeddedTemplatesMatchSource` 는 `EmbeddedTemplates()`(=`//go:embed
all:templates` 를 `go test` 시점에 컴파일한 스냅샷)와 **같은 디스크 디렉터리**의 런타임 read 를
비교한다. 생성 파일이 없으므로 둘은 항상 같다.

뮤테이션으로 실증했다 — 템플릿 원본에 한 줄을 덧붙이고(스테일 바이너리라면 불일치해야 함) 검사를
돌렸더니 그대로 통과했다(`ok internal/template 0.418s`). 이후 원상 복구, 트리 clean 확인.
따라서 이 테스트의 실패 메시지 `"the binary predates the edit (run make build)"` 는 **도달 불가**다.

공정하게 적자면 §E.2 는 이 한계를 이미 Gap 으로 기록했다("proves embed-directive ↔ source parity
at test-build time, not that the on-disk `bin/moai` was rebuilt"). 그럼에도 AC 매트릭스는 이 명령을
AC-BJD-016 의 주 검증 수단으로 계속 인용한다.

**Required fix**: AC 매트릭스에서 이 테스트를 "소스↔embed 지시자 파리티" 로 재라벨하고, 실패
메시지에서 `make build` 문구를 뺀다. 바이너리 최신성은 별도 수단으로만 주장한다.

### D3 — [MINOR] [optional] `strings bin/moai | grep -c 'cksum "$d"/backlog.db'` 는 판별력이 없다

§E.2 AC-BJD-016 행은 이 프로브가 `→ 1` 이었다고 기록한다. 이 환경에서 재현되지 않는다:

```
$ strings bin/moai | grep -c 'cksum "$d"/backlog.db'      → 0
$ strings -a bin/moai | grep -c 'cksum "$d"/backlog.db'   → 0
$ grep -a -c 'cksum "$d"/backlog.db' bin/moai             → 0
# 그리고 이 트리에서 갓 빌드한 바이너리(문자열을 확실히 포함)도 → 0
```

원인: 이 환경의 `grep` 은 **ugrep 7.8.4** 이고, 패턴 중간의 비이스케이프 `$` 를 앵커로 취급한다.
확인:

```
$ printf 'cksum "$d"/backlog.db\n' > probe.txt
$ grep -c 'cksum "$d"/backlog.db'  probe.txt   → 0
$ grep -c 'cksum "\$d"/backlog.db' probe.txt   → 1
```

**실체 자체는 참이며 내가 따로 확인했다**: `grep -a -o 'cksum "\$d"/backlog.db…'` 가 `bin/moai` 와
신규 빌드 양쪽에서 히트하고, 구 블록 `f=.moai/state/todo/backlog.json` 은 0건이다. 문제는 사실이
아니라 **증거**다 — 존재할 때도 부재할 때도 0 을 내는 프로브는 아무것도 판별하지 못한다.

**Required fix**: `$` 를 이스케이프하거나(`'cksum "\$d"/backlog.db'`) `grep -a -F` 를 쓴다.
D2 와 합치면, 현재 AC-BJD-016 은 이 트리에서 **작동하는 검증 수단이 하나도 없다** — 실체는 참이지만
그것을 세운 것은 lane 의 두 명령이 아니라 이 감사의 재측정이다.

### D4 — [MINOR] [optional] "한계" 문단이 tool-path 독자 커버리지를 과장한다

CHANGELOG:

> The disclosure reaches only readers who go through the tool — someone running `cat` on
> `backlog.json` still gets a silent wrong answer

그런데 `moai todo --help` 는 **툴을 경유하는 독자**이고, `internal/cli/todo.go:96` 의 `Long` 은
지금도 "Operate the kanban backlog queue at `.moai/state/todo/backlog.json`" 이라고 단언한다.
이 문자열은 바이너리에 실제로 실려 나간다(바이너리 내 `state/todo/backlog.json` 잔존 2건 중 1건이
이것, 나머지 1건이 `todo-queue-storage.md` 의 export 통제행).

§E.2 는 이것을 "Adjacent defect found, deliberately NOT fixed" 로 **정직하게 기록했고 카드로
에스컬레이션했다** — 그 처리는 옳다. 누락은 CHANGELOG 한계 문단이 그 사실을 옮기지 않은 것이다.

**Required fix**: 한계 문단에 한 절 추가 — "`moai todo --help` 자체는 아직 `backlog.json` 을
단언한다(별도 카드)".

**범위 판단에 대한 의견 (리드 질의)**: `internal/cli/todo.go:96` 을 범위 밖으로 둔 결정은 **옳다**.
REQ-BJD-011/012 는 두 skill 파일에 묶여 있고 AC-BJD-012 의 grep 은 `.claude/skills/` 로 한정된다.
여기서 고쳤다면 AC 없는 코드 변경, 즉 범위 확장이었다. 다만 D4 대로 **한계 서술에는 실려야** 한다.

### D5 — [MINOR] [optional] 공개 프로브가 adopting 리졸버를 거친다

`discloseQueueLayout` → `newTodoStore()` → `resolveTodoQueueRoot()` →
`kanban.ResolveTodoQueueRootAdopting(...)` — **adopting** 변종이며 홈폴백 분기에서
`adoptLocalTodoQueue` 라는 파일시스템 부작용을 갖는다(`todo_root.go:76`).

순 신규 부작용은 없다 — 같은 verb 가 한 줄 뒤에서 `newTodoStore().Load()` 로 어차피 같은 호출을
하고 adoption 은 멱등이다. 다만 결과적으로 **공개가 verb 의 읽기 경로 자신이 홈 루트에 옮겨 놓은
backlog.json 을 보고 발화할 수 있다**. 첫 fixture(git 메타데이터 없음 → 홈폴백)에서 우연히 관측했다:
프로젝트 로컬에 backlog.json 을 두었더니 홈 루트가 큐인데도 공개가 발화했다.

AC-BJD-006 위반은 아니다 — 그 AC 는 **inspector**(`InspectBacklogArchiveVouch`)를 대상으로 하며
그것은 실제로 read-only 다(내가 확인). 기록해 두는 이유는 문서화된 의도("It touches no file and
takes no lock")가 함수 자신에 대해서는 참이지만 `discloseQueueLayout` 전체 경로에 대해서는
정확히 참은 아니기 때문이다.

**Required fix**(선택): `discloseQueueLayout` 이 순수 리졸버를 쓰도록 하거나, 주석에 adopting
경유 사실을 한 줄 남긴다.

### D6 — [INFO] [optional] 미러 divergence 는 19가 아니라 20

`.claude/skills/moai/SKILL.md` ↔ 템플릿: diff 라인 39 = CLAUDE_SKILL_DIR 쌍 19 + **로컬에만 있는
`Last Updated: 2026-07-07` 1행**. "19 pre-existing CLAUDE_SKILL_DIR lines" 는 거짓이 아니지만
divergence 전량은 아니다. 반면 나머지 두 미러 파일은 **byte-identical** 임을 확인했으므로,
"scoped, not whole-file" 이라는 서술 자체는 보수적으로 정확하다.

### D7 — [INFO] [optional] 같은 결함 클래스의 산문 2건 (범위 밖, 기능 결함 아님)

- `internal/web/screens.templ:286` / `screens_templ.go:1242` — 주석이 큐 파일을 `backlog.json` 이라
  단언한다. **기능 결함은 아니다**: `watchMap["kanban"]` 은 `.moai/state/todo` **디렉터리**를
  감시한다(`internal/web/events.go:32`). 확인함.
- `internal/hook/session_start_kanban.go:194` — 같은 성격의 주석. 동작은
  `kanban.QueuedBacklogCountForRoot` 에 위임하므로 무해.

`todo.go:96` 카드에 함께 묶어 두면 좋겠다.

---

## Gaps — 관측하지 **않은** 것

- **Windows**: AC-BJD-008/009/010 은 이 감사에서도 미측정이다(darwin 에서만 실행). skip 은 명시적
  `t.Skipf` 이므로 조용한 통과는 아니다.
- **`go test ./...` 미실행**: CLAUDE.local.md §4/§6 규율. 전량 판정은 CI 몫.
- **`internal/web` i18n 2건**: develop 이 이미 붉다고 전달받았으나 **재측정하지 않았고** 어느 쪽으로도
  귀속하지 않는다.
- **`cli` 커버리지의 pre-change 베이스라인 없음**: 80.3% 가 이 변경으로 움직였는지는 여전히 미귀속
  (lane 이 §E.2 에 Gap 으로 기록한 것과 동일하며, 나도 재기 않았다).
- **watch 의 잔여 사각**: cksum 충돌, 5초 폴 안의 add-then-revert, `backlog.db-shm`, 비-WAL 저널
  폴백, 단일 플랫폼/SQLite 빌드 — 모두 미측정이며 CHANGELOG 서술과 일치한다. 이 감사도 세 번째
  사각의 부재를 세우지 않았다.
- **스테일 `backlog.json` 의 작성자**: 여전히 미식별. 그래서 정리가 아닌 공개가 방어라는 §A.3 논리는
  이 감사에서도 유지된다.

## Residual risk

- D1 이 정정되지 않고 릴리스되면 거짓 정량 주장이 공개 릴리스 노트에 남는다.
- D2+D3 이 남으면 AC-BJD-016 은 **작동하는 자동 검증 없이** 계속 PASS 로 보고된다 — 다음에 누군가
  `make build` 를 빠뜨려도 초록이다. 실체는 지금 참이지만, 그것을 지키는 장치는 없다.
- `moai todo --help` 가 계속 `backlog.json` 을 단언하는 동안, 이 SPEC 이 막으려던 바로 그 오독을
  툴 자신의 도움말이 만들어낼 수 있다(D4, 별도 카드).

---

## 감사 위생

- 트리에 커밋하지 않았다. D2 뮤테이션은 `cp` 백업 후 즉시 복구했고 `git status --short` 무출력·
  `HEAD` 불변(`e8a30fa1…`) 확인.
- fixture 는 전부 `/tmp` 격리, 종료 시 삭제. primary 체크아웃의 살아있는 큐
  (`/Users/goos/MoAI/moai-adk-go/.moai/state/todo/`)는 읽지도 건드리지도 않았다.
  홈폴백 잔재(`~/.moai/todo/t395-*`)도 제거 확인.
- 이 보고서 외에 트리에 남긴 파일 없음.
