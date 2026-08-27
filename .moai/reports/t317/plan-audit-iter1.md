# SPEC 감사 보고서: SPEC-AGENT-EMIT-LINEAGE-001

- Iteration: 1/1 (Tier S → `plan_audit_tier_ceilings` S=1)
- **Verdict: FAIL**
- **Overall Score: 0.74** (조화평균) — Tier S PASS 임계 **0.75** 미달
- 감사 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t317` @ `48eb945df`, branch `WT-agent-emit-lineage`
- 저작자 추론 맥락은 M1 Context Isolation 에 따라 무시했다. 판정 근거는 `spec.md` / `plan.md` / `progress.md` 와 내가 이 트리에서 이 실행으로 잰 값뿐이다. `measurement.md` 는 증거 산출물로 읽되 그 주장들을 그대로 받지 않고 재실행해 검증했다.

FAIL 의 직접 사유는 점수가 아니라 **MP-7(clarification gate) 실패**다. 점수도 임계 아래이나, 두 사유는 독립적으로 각각 FAIL 을 강제한다.

---

## Must-Pass 결과

| 기준 | 결과 | 근거 |
|---|---|---|
| MP-1 REQ 번호 일관성 | **PASS** | `REQ-AEL-001`~`008` 연속, 결번·중복 정의 0, 3자리 zero-padding 일관 (`spec.md:65-79`). 중복 카운트로 보인 002/008 은 §4·§5 본문 참조 |
| MP-2 GEARS 형식 준수 | **PASS** | **요구 층(`REQ-XXX`)에 대해 판정.** 001·005 = Event-driven(`When …, the … shall`), 002 = Unwanted(`shall not`), 003·004·006·007 = Ubiquitous, 008 = Where. `spec.md:65-79`. §3 의 Given-When-Then 항목은 검증 층(`AC-XXX`)이므로 이 기준으로 감점하지 않았다 |
| MP-3 YAML frontmatter 유효성 | **PASS** | 정본 12필드 전부 존재(`spec.md:2-13`), 거부 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0건, `status: draft` 는 유효 enum, `version` 은 인용된 semver |
| MP-4 §22 언어 중립성 | **N/A (auto-pass)** | 이 저장소 자체의 Go/Makefile 배선에 한정된 단일 언어 SPEC. 배포 템플릿 트리에는 문서를 넣지 않기로 명시(`plan.md:55`)해 §25 중립성도 침범하지 않는다 |
| MP-5 D7 교차 SPEC 정합 | **PASS** | 본문 SPEC-ID 참조는 자기 자신 1건뿐. `status: draft` — retired/superseded/archived 아님. BLOCKING 없음 |
| MP-6 D8 크로스 플랫폼 규율 | **PASS (auto)** | `grep -c 'syscall' spec.md` → `0`. 해당 사안 없음 |
| MP-7 clarification gate | **FAIL** | 아래 D1 |

```console
$ grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/plan.md
plan.md:40:**[NEEDS CLARIFICATION: `make embed-check` 를 자동 호출 지점에 걸 것인가]** — 후보: (i) 수동 verb 로만 둔다, (ii) CI 의 빌드 잡 뒤에 건다, (iii) `moai doctor` 항목으로 편입한다.
```

`plan.md` 가 존재하므로 N/A 처리 불가. 미해결 마커 1건 = must-pass 실패.

---

## 차원별 점수

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 경로·동사·판정 명령이 모두 이름으로 박혀 있다. 다만 REQ-AEL-002 의 주어 범위가 모호하고(D4), M1 호출 지점이 미결이다(D1) |
| Completeness | 0.75 | 0.75 | 필수 절 전부 존재, frontmatter 12/12, `### Out of Scope — ` 5건 각각 불릿 보유, Tier S 산출물 집합 충족. 임베드 검사의 공허성 봉투가 미명세(D3) |
| Testability | 0.70 | 0.50↔0.75 사이 | 8건 중 6건은 실행 가능한 판정 명령을 갖춘 강한 AC. 그러나 AC-AEL-008 은 **실행으로 공허함을 확증**했고(D2), AC-AEL-006 은 대조군이 빠졌다(D6). 공허성 근절을 주제로 삼은 SPEC 에서 공허한 AC 가 남은 것은 0.75 밴드 서술보다 무겁다 |
| Traceability | 0.75 | 0.75 | REQ 8건 전부 AC 로 덮이고 고아 AC 0건. 다만 AC→REQ 방향 사상이 8건 **전부** `plan.md` 마일스톤 경유의 간접 사상(D5) |

조화평균 = 4 / (1/0.75 + 1/0.75 + 1/0.70 + 1/0.75) = **0.7368 → 0.74**. Tier S 임계 0.75 미달.

---

## 지정 감사 항목별 판정

### 1. 실측 8 의 동어반복 — **주장 성립, 내가 직접 재현했다**

같은 뮤턴트를 심어 두 테스트를 각각 걸었다.

```console
$ printf '\n# mutant-audit-t317\n' >> internal/template/templates/.codex/agents/moai/manager-git.toml
$ go test ./internal/template/agentemit/... -run TestEmbedFSPresenceAndByteEquality -count=1
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.405s      ← 뮤턴트 생존
$ go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission -count=1
--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.00s)
    golden_test.go:109: .codex/agents/moai/manager-git.toml: committed artifact differs from emission (sha256 mismatch)
FAIL                                                                        ← 대조군 사망
$ git checkout -- internal/template/templates/.codex/agents/moai/manager-git.toml
$ go test ./internal/template/agentemit/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.424s
```

코드로도 확인했다. `golden_test.go:271`(`fs.ReadFile(embedded, rel)`)과 `:277`(`os.ReadFile(committed)`)이 같은 온디스크 원본에서 파생되고, `go test` 가 매 실행 테스트 바이너리를 새로 컴파일하므로 양변이 함께 움직인다. **동어반복 판정은 참이다.**

**REQ-AEL-004/005 는 이 동어반복을 벗어난다.** 판정 대상을 "이미 존재하는 빌드 산출물"로 옮기면 임베드 바이트는 그 바이너리의 빌드 시점에 동결되고 커밋본은 이후 독립적으로 움직인다. 양변이 분리된다.

SPEC 이 "갓 빌드한 바이너리는 정의상 커밋본과 일치하므로 `build` 선행을 두면 안 된다"고 한 주장도 **참이다.** `Makefile:23-25` 의 `build` 는 `templ-generate` → `gen-catalog-hashes.go` → `go build` 순이고, `go build` 는 그 시점 온디스크 `templates/` 를 그대로 임베드한다. `.codex/*.toml` 을 건드리는 단계가 없으므로 빌드 직후 바이너리의 `.codex` 바이트 = 온디스크 커밋본이다. 따라서 `build` 를 선행으로 두면 검사가 다시 X와 X를 비교하게 된다. **REQ-AEL-004 의 "runs without rebuilding" 은 이 SPEC 에서 협상 불가한 조항이며, 이를 건드리는 AC 는 없다 — 이 갈래는 무너지지 않는다.**

### 2. AC-AEL-003 의 기계적 판정 가능성 — **성립**

뮤턴트가 명시되고(`printf` 한 줄), RED 가 특정되며(exit ≠ 0 + `manager-git.toml` 지목), 원복 절차가 있고(`git checkout --` → `make build` → 재실행 exit=0), **대조군까지 요구한다**(기존 `TestEmbedFSPresenceAndByteEquality` 가 같은 뮤턴트에 여전히 통과함을 기록). 실측 8 이 죽이지 못한 뮤턴트를 정확히 겨눈다. 다만 판정 범위에 구멍이 남는다 — D3.

### 3. M1 실현 가능성(저작자가 UNPROVEN 으로 표시) — **내가 실증했다. 성립한다**

저작자가 실행하지 않은 전제를 이 감사에서 직접 실행했다. 설치본 `moai`(v3.1.3-rc.5, commit `22df80e90`)로 스크래치 디렉터리에 배포:

```console
$ moai init <scratch>/d1 --non-interactive        # 기본 모드 (slim, --agent 기본값 claude)
---exit=0
$ find <scratch>/d1 -path '*/.codex/*' -name '*.toml' | wc -l
11
$ diff -rq <scratch>/d1/.codex/agents/moai internal/template/templates/.codex/agents/moai
Files .../manager-develop.toml and internal/template/.../manager-develop.toml differ
```

판정 세 가지:

1. **`.codex` 는 기본 배포 대상이다.** `--agent codex` 도 `--all` 도 필요 없었다. `plan.md:34` 의 "포함될 것으로 보이나 미실증" 과 `plan.md:67` 위험 행 "`.codex` 가 배포 대상이 아닐 가능성" 은 **닫혔다.**
2. **배포 경로는 바이트를 보존한다.** 11개 중 10개가 커밋본과 완전 일치. 유일한 차이인 `manager-develop.toml` 은 배포 경로의 변형이 아니라 **진짜 내용 스큐**다 — diff 내용이 테스트 범위 독트린 변경("run the full suite" → "run the tests the change can affect")으로, 설치본 빌드 커밋과 이 트리 base 사이의 실제 문서 변화다.
3. **그 한 건이 곧 REQ-AEL-004 가 겨눈 결함의 살아있는 실례다.** 설치된 바이너리가 커밋본과 다른 `.codex` 바이트를 싣고 있음을 소스↔소스 비교 없이 관측했다. 즉 이 감사의 측정 자체가 임베드 축 검사의 동작 프로토타입이고, **그 축은 공허하지 않다**(양성 검출이 실제로 났다).

**따라서 `plan.md:35` 의 2순위(바이너리에 임베드 자산 덤프 경로 추가)는 불필요하다.** 배포되는 바이너리에 유지자 전용 검사만을 위한 CLI 표면을 더하는 일이므로, 1순위가 성립한 지금 이것을 짓는 것은 표면 확대에 불과하다. run-phase 는 2순위를 재설계하지 말고 폐기하는 편이 낫다. 다만 1순위도 **동사가 미특정**이다(`init` 인지 `update` 인지 `plan.md:34` 가 말하지 않는다) — 위 실행이 `moai init <dir> --non-interactive` 로 성립함을 근거로 run-phase 가 이를 확정하면 된다.

**실현 가능성 결론: 구현 가능하다.** UNPROVEN 표시는 이 감사로 해소됐고, 2순위 폴백은 실행 불가능해서가 아니라 불필요해서 제거 대상이다.

### 4. Tier S 분류 — **성립. 상한 주장도 정확하다**

SSOT(`spec-workflow.md:138-152`) 대조:

- Tier S = `< 300 LOC` AND `< 5 files` → `plan.md:13-14` 의 3-5 파일 / <300 LOC 추정과 부합. M1 이 배포 추출 방식으로 확정되면 Makefile + 검사 스크립트 1 + 문서 1 로 실제 3파일권.
- **REQ 상한 8 / AC 상한 8, 두 축에 독립 적용** — SPEC 의 "요구 8 / 수락 8 … 둘 다 상한"(`spec.md:206`)은 SSOT 를 정확히 인용한 참인 주장이다. 실측 카운트도 REQ 8 / AC 8 로 정확히 일치.
- 산출물 집합: Tier S = spec.md + plan.md(AC 는 spec.md §3 인라인). 충족. `progress.md` 는 워크플로 공통 산출물이라 위반이 아니다.
- PASS 임계 = **0.75**.

"한 항목이라도 늘면 Tier 재판정" 이라는 자기 구속은 참이지만, 아래 D2·D3·D4 의 처방은 **기존 REQ 본문 확장 / AC 교체**로 흡수되므로 항목 수를 늘리지 않는다. Tier 재판정 없이 수리 가능하다. (D2 를 REQ-AEL-008 폐기로 처리하면 오히려 예산이 하나 빈다.)

### 5. 범위 경계 — **누수 없음**

- **C1↔C2 패리티**: 어느 REQ·AC 도 `.claude/agents/moai/` 를 대상으로 삼지 않는다. REQ-AEL-006 이 지목하는 것은 `internal/template/templates/.claude/agents/moai/*.md`(C2)로 올바르다.
- **`graph-freshness`**: 본문 전체에 요구로 등장하지 않는다.
- **`catalog.yaml` 의 `.codex` 미수록**: `spec.md:191` 이 관측으로만 기록하고 요구화하지 않는다. AC-AEL-004 는 `catalog.yaml` 쓰기를 명시적으로 판정 대상에서 뺀다(`spec.md:132`).

### 6. 전제 정합성 — **정정된 전제 위에 쓰였다. 역전된 전제의 잔재 없음**

정정 내용을 직접 재측정했다.

```console
$ sed -n '183p;238p' .github/workflows/ci.yml
        run: go test -coverprofile=coverage.out -covermode=atomic ./...
        run: go test -race -count=1 ./...
$ grep -rn "AGENTEMIT_UPDATE" .github/ | wc -l
0
```

CI 는 update 분기 없이 골든을 돌린다 → 소스 층 드리프트는 머지 경로에서 잡힌다. SPEC 은 이를 `spec.md:49`(실측 2)와 `spec.md:188`("CI 검출은 이미 존재한다. 새로 만들지 않는다")로 정면 수용하고, 남는 창을 §1.3 에서 **로컬 지연**과 **바이너리 노후** 둘로 한정한다. **"CI 가 못 본다"를 근거로 삼은 요구는 하나도 없다.** REQ-AEL-001 의 정당화는 CI 부재가 아니라 로컬 즉시성이다.

부수 검증 — AC-AEL-007 의 baseline 주장(0 히트)도 이 트리에서 재측정해 `0` 으로 일치했다. 제대로 귀속된 baseline 이다.

---

## 발견된 결함

**D1 — 미해결 `[NEEDS CLARIFICATION]` (clarification gate)**
`plan.md:40` — `make embed-check` 의 자동 호출 지점이 (i) 수동 verb / (ii) CI 빌드 잡 후행 / (iii) `moai doctor` 편입 사이에서 미결. — **Severity: critical — Class: blocking**
이것이 단순 절차 흠결이 아닌 이유: (i) 를 고르면 M1 의 산출물은 **아무도 부르지 않는 수동 동사**가 되고, §1.3 이 지목한 창 2(바이너리 노후)에는 기계적 방아쇠가 하나도 남지 않는다. 그러면 `spec.md:42` 가 스스로 내건 기준 — "우연이 잡아낸 결함이므로 처방은 절차가 아니라 **기계적 가드**여야 한다" — 를 M1 자신이 충족하지 못한다. 즉 이 마커는 M1 의 값어치를 좌우하는 결정이지 배선 세부가 아니다.
**필요 조치**: 오케스트레이터가 Implementation Kickoff Approval **이전에** `AskUserQuestion` 으로 (i)/(ii)/(iii) 를 확정하고, 결정과 그 근거를 `plan.md` M1 에 적어 마커를 제거한다. (ii)/(iii) 선택 시 사용자 가시 표면 확대를 함께 기록한다.

**D2 — AC-AEL-008 이 공허하다 (실행으로 확증)**
`spec.md:165-174` — 이 AC 의 시나리오는 **현재의 미수정 코드에서 이미 통과한다.**
```console
$ chmod 444 internal/template/templates/.codex/agents/moai/manager-git.toml
$ make agents-emit
    golden_test.go:98: update write .codex/agents/moai/manager-git.toml: open ../templates/.codex/agents/moai/manager-git.toml: permission denied
FAIL
make: *** [agents-emit] Error 1
$ chmod 644 internal/template/templates/.codex/agents/moai/manager-git.toml
```
`golden_test.go:97-99` 가 `os.WriteFile` 오류를 이미 `t.Fatalf` 로 세우므로, 권한 실패는 지금도 exit ≠ 0 이고 경로까지 이름으로 지목한다. REQ-AEL-008 을 하나도 구현하지 않아도 이 AC 는 GREEN 이다 — 구현 전후를 구분하지 못하는 AC 다. 게다가 REQ-AEL-008 의 실질(쓰인 바이트를 **되읽어 방출 결과와 대조**)은 권한 실패가 아니라 **조용한 내용 불일치**를 겨누는데, AC 는 이미 시끄러운 쪽을 시험한다. 공허성 근절을 주제로 세운 SPEC 에 공허한 AC 가 남은 것은 내부 모순이다. — **Severity: major — Class: blocking**
**필요 조치**: 둘 중 하나. (a) AC 를 내용 불일치 면으로 교체 — 쓰기 직후 파일이 방출 결과와 다르도록 만든 뒤(쓰기 후 외부 변조 주입 등) 검사가 exit ≠ 0 을 내는지 보이고, 같은 조건에서 **현행 코드는 통과함**을 대조군으로 기록한다. (b) REQ-AEL-008 + AC-AEL-008 을 함께 폐기한다(실측 3 이 이미 "작은 사안"으로 규정했고, 권한 면은 이미 닫혀 있다).

**D3 — 임베드 검사의 공허성 봉투가 미명세**
`spec.md:71-73`(REQ-AEL-004/005) 및 `spec.md:113-126`(AC-AEL-003) — 검사가 **몇 건을 비교했는지** 주장할 의무가 없고, **판정 대상 바이너리가 없을 때**의 거동도 규정되지 않았다. — **Severity: major — Class: blocking**
근거: 이 검사가 대체·보완하는 기존 테스트는 기수(cardinality)를 단언한다 — `golden_test.go:285`: `if count != 11 { t.Errorf("committed .codex/agents/moai carries %d TOMLs, want 11", count) }`. 새 검사에는 대응 조항이 없다. 추출이 부분적으로만 성공해 일부 경로만 비교해도(그 안에 뮤턴트 파일이 우연히 포함되면) AC-AEL-003 은 통과하고 AC-AEL-004 도 통과한다. `bin/moai` 부재 시에도 "비교 0건 → exit 0" 이 규정상 허용된다. 둘 다 이 SPEC 이 근절하겠다고 선언한 바로 그 실패 유형(공허한 가드)이다.
**필요 조치**: 신규 항목을 만들지 말고 기존 본문을 넓힌다 — REQ-AEL-004 에 "추출된 집합이 커밋된 방출물 전량을 덮는지 확인하고, 덮지 못하면 실패한다" + "판정 대상 바이너리가 없으면 성공이 아니라 실패로 종료한다"를 추가하고, AC-AEL-003 의 Then 에 기수 확인과 바이너리 부재 케이스를 넣는다. Tier S 예산 불변.

**D4 — REQ-AEL-002 의 무쓰기 의무가 임베드 검사를 형식상 덮지 않는다**
`spec.md:67` 의 주어는 "The drift check" 로, 문맥상 M2 의 소스 층 검사다. 임베드 검사(M1)의 무쓰기 의무는 `plan.md:36` 에만 "REQ-AEL-002 와 같은 규율" 이라는 표현으로 존재하며 — 이 표현 자체가 REQ-AEL-002 가 그것을 구속하지 않음을 인정한다 — 이를 검증하는 AC 도 없다. — **Severity: minor — Class: blocking**
이 구멍이 실제로 물리는 이유: 내가 실증한 1순위 추출 경로는 스크래치에 **프로젝트 전체를 배포**하는 무거운 쓰기 동작이다(git init 과 훅 설치까지 동반하는 것을 관측했다). 대상 디렉터리를 잘못 잡았을 때 저장소 트리를 오염시키지 않을 보장이 요구 층에 없다.
**필요 조치**: REQ-AEL-002 의 주어를 "이 SPEC 이 도입하는 모든 검사" 로 넓힌다(한 줄, 신규 REQ 불요).

**D5 — AC→REQ 사상이 8건 전부 간접이다**
`spec.md:87-174` — 어떤 AC 헤딩·본문도 `REQ-AEL-XXX` 를 인용하지 않는다. 사상은 `plan.md:24,42,53` 의 마일스톤 제목에만 있다. 덮개 자체는 완전하고(REQ 8건 전부 피복, 고아 AC 0건) 모호하지도 않으나, 두 문서를 겹쳐 읽어야만 성립한다. — **Severity: minor — Class: optional**
**필요 조치**: 각 AC 헤딩에 대응 REQ id 를 병기한다(AC 당 한 토큰).

**D6 — AC-AEL-006 에 대조군이 없다**
`spec.md:144-152` — `**[뮤테이션 확립]**` 로 표시했으나, `.PHONY` 추가 **이전** 상태에서 make 가 실제로 건너뛴다는 RED 관측을 요구하지 않는다. AC-AEL-003 이 대조군을 명시적으로 요구하는 것과 대비되는 비대칭이며, SPEC 이 스스로 세운 "실패시킬 수 있음을 보이지 못한 가드는 가드가 아니다"(`spec.md:98`) 기준에 미달한다. 부수적으로 Then 의 예상 문구("nothing to be done")도 GNU make 의 실제 출력(`make: 'agents-emit' is up to date.`)과 어긋난다. — **Severity: minor — Class: blocking**
**필요 조치**: 수정 전 `touch agents-emit && make agents-emit` 에서 레시피가 실행되지 않음을 기록하는 대조 단계를 한 줄 추가하고, 기대 문구를 실제 make 출력으로 맞춘다.

---

## 확인된 강점 (run-phase 가 그대로 써도 되는 부분)

기록해 둘 값이 있는 검증 결과다.

1. **동어반복 판정(실측 8)은 재현됐다** — 이 SPEC 의 하중을 받는 발견은 견고하다. M1 의 존재 이유가 성립한다.
2. **M1 1순위 추출 경로는 실증됐다** — `moai init <dir> --non-interactive` 로 `.codex` TOML 11건이 기본 배포되고 바이트가 보존된다. `plan.md:67` 의 위험 행은 닫혔고, 2순위 폴백은 불필요하다.
3. **임베드 축이 공허하지 않음을 양성 검출로 확인했다** — 설치본과 이 트리 사이 `manager-develop.toml` 1건 불일치. 검사가 잡아야 할 상태가 실제로 존재한다.
4. **정정된 전제가 전 구간에 반영됐다** — 역전된 "CI 가 못 본다" 전제 위에 세워진 요구는 없다.
5. **범위 차단이 실효적이다** — 배제 3영역 어디로도 REQ·AC 가 넘어가지 않는다.

---

## 권고

1. **D1 을 먼저 닫는다.** Implementation Kickoff Approval 이전에 `AskUserQuestion` 으로 M1 호출 지점을 확정하고 마커를 제거한다. 이것 없이는 재감사 자체가 성립하지 않는다(must-pass).
2. **D2 를 처리한다** — AC-AEL-008 을 내용 불일치 면으로 교체하거나, REQ-AEL-008 과 함께 폐기한다. 폐기가 더 정직한 선택으로 보인다: 권한 면은 이미 닫혀 있고 실측 3 도 이를 작은 사안으로 규정했다.
3. **D3·D4 를 기존 본문 확장으로 흡수한다** — REQ-AEL-004 에 기수·바이너리 부재 조항, REQ-AEL-002 에 주어 확대. 항목 수가 늘지 않으므로 Tier S 재판정은 불필요하다.
4. **D6 을 한 줄로 닫고**, D5 는 오케스트레이터 재량(선택 사항)으로 남긴다.
5. **`plan.md` M1 을 이 감사 결과로 갱신한다** — "미실증" 표기를 실증 결과로 대체하고, 2순위 폴백을 제거하며, 추출 동사를 `moai init <scratch> --non-interactive` 로 확정한다. run-phase 가 이미 밝혀진 것을 다시 설계하지 않게 된다.

수리 후 재감사는 위 D1-D6 델타에 한정한다(Tier S 상한 1회 초과분은 오케스트레이터 판단 사항).

---

## 감사 중 트리 변경 (전량 원복 확인)

이 감사는 뮤테이션 2건과 권한 변경 1건을 수행했고 모두 원복했다.

```console
$ git status --short
?? .moai/reports/t317/
?? .moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/
$ go test ./internal/template/agentemit/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.253s
```

추적 파일 수정 0건. 스크래치 배포물은 저장소 트리 밖(`/private/tmp/.../scratchpad/d1`)에만 생성했다.

## Gaps — 관측하지 않은 것

- `make build` 를 실행하지 않았다(트리를 변형하는 `catalog.yaml` in-place 쓰기와 `*_templ.go` 생성을 감사 중 유발하지 않기 위함). 따라서 AC-AEL-001/004 의 실제 build 거동은 판정하지 않았고, REQ-AEL-001 의 비용 주장(0.419s)도 build 문맥에서 재측정하지 않았다.
- 배포 추출 실증은 **설치본**(v3.1.3-rc.5 / `22df80e90`)으로 했다. 이 트리에서 빌드한 `bin/moai` 로는 확인하지 않았다 — 배포 경로의 구조적 성질을 묻는 측정이므로 결론은 유지되나, 버전 동일성까지 보인 것은 아니다.
- 11개 TOML 전부에 대해 임베드 테스트 공허성을 확인하지는 않았다(1건 표본).
- `moai init` 의 slim/`--all` 및 `--agent` 조합 전수를 시험하지 않았다. 기본 모드에서 `.codex` 가 배포된다는 것만 관측했다.

## Residual risk

- D3 의 기수 조항 없이 착지하면, 추출이 부분 성공하는 환경에서 새 검사가 조용히 약해질 수 있다 — 이 SPEC 이 겨눈 실패 유형의 재발이며, 재발 시 다시 우연에 의존해 발견해야 한다.
- M1 호출 지점이 (i) 수동으로 결정되면, 이 카드는 창 1(로컬 지연)만 닫고 창 2(바이너리 노후)는 열어 둔 채 종료된다. 그 결과가 수용 가능한지는 감사관이 아니라 운영자가 판단할 사안이다.
