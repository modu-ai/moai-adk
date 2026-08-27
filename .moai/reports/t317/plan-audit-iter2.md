# SPEC 감사 보고서 (iter-2): SPEC-AGENT-EMIT-LINEAGE-001

- Iteration: 2/1 — **Tier S 상한(1) 초과분**. `harness.yaml:76` (`plan_audit_tier_ceilings.S: 1`). 상한 초과 재감사는 오케스트레이터 판단으로 개시됐으며, 판정 권한은 이 에이전트에 남는다
- **Verdict: PASS-WITH-DEBT**
- **Overall Score: 0.82** (조화평균) — Tier S PASS 임계 **0.75** 초과
- **점수 이동: 0.74 → 0.82, 단조 상승.** 회귀 없음 → STOP 에스컬레이션 조건 미충족
- 감사 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t317` @ `48eb945df`, branch `WT-agent-emit-lineage`, spec `version: 0.3.0`
- 저작자 추론 맥락은 M1 Context Isolation 에 따라 무시했다. `measurement.md` 의 실측 9·10 은 **증거로 읽되 권위로 받지 않았고**, 결론에 하중이 걸리는 부분은 이 트리 이 실행에서 다시 쟀다
- 범위: iter-1 의 D1-D6 델타 + 회귀 검사. iter-1 에서 통과시킨 항목(동어반복 전제, AC-AEL-003 판정가능성, Tier S, 범위 경계, 전제 정합성)은 **깨지지 않았는지만** 확인했다

---

## Must-Pass 결과

| 기준 | 결과 | 근거 |
|---|---|---|
| MP-1 REQ 번호 일관성 | **PASS** | `grep -n '^\*\*REQ-AEL-'` → 001,002,003,004,005,006,007 정의 7건. 결번·중복 0, 3자리 padding 일관. 008 정의는 0건(폐기 서술로만 등장) |
| MP-2 GEARS 형식 준수 | **PASS** | **요구 층(`REQ-XXX`)에 대해 판정.** 001 Event-driven(`When … shall`), 002 Unwanted(`shall not`), 003·004·006·007 Ubiquitous(`shall`), 005 Event-driven. `spec.md:67-85`. §3 의 Given-When-Then 은 검증 층(`AC-XXX`)이므로 이 기준으로 감점하지 않았다 |
| MP-3 YAML frontmatter 유효성 | **PASS** | 정본 12필드 전부 존재(`spec.md:2-13`) + `tier`/`era`. 거부 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0건. `version: "0.3.0"` 인용 semver, `status: draft` 유효 enum |
| MP-4 §22 언어 중립성 | **N/A (auto-pass)** | 이 저장소 Go/Makefile 배선 한정 단일 언어 SPEC. 문서 게재를 `CLAUDE.local.md` 로 못박아 배포 템플릿 중립성(§25)도 침범 안 함(`plan.md:64`) |
| MP-5 D7 교차 SPEC 정합 | **PASS** | 본문 `SPEC-` 참조는 자기 자신 1건. `status: draft`. BLOCKING 0 |
| MP-6 D8 크로스 플랫폼 규율 | **PASS (auto)** | `grep -c 'syscall' spec.md` → `0` |
| MP-7 clarification gate | **PASS ← iter-1 FAIL 에서 전환** | 아래 |

```console
$ grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/
$ echo "grep_rc=$?"
grep_rc=1
```

디렉터리 **전체**를 훑었다(spec.md / plan.md / progress.md). 마커 0건. iter-1 의 유일한 must-pass 실패가 닫혔다.

---

## 차원별 점수

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | D4·D1 이 닫혀 iter-1 의 두 감점 사유가 사라졌다. 그러나 v0.3.0 이 새 모호성을 하나 들여왔다 — REQ-AEL-004 의 doctor 조항이 배포 프로젝트에서 무엇을 대조하고 어떤 종료 코드를 내는지 미규정(D7). 합리적 엔지니어가 서로 다르게 구현할 수 있는 조항이므로 밴드 상향은 부당하다 |
| Completeness | 0.80 | 0.75↔1.0 | 필수 절 전부, frontmatter 12/12, `### Out of Scope —` 5건 각각 불릿 보유, Tier S 산출물 집합 충족. D3 의 공허성 봉투가 요구·수락 양쪽에 들어갔다. 감점은 v0.3.0 이 새 구현 표면(`internal/cli`)을 도입하고도 `module:` 필드와 `plan.md §B` 파일 수 추정을 갱신하지 않은 것(D8) |
| Testability | 0.90 | 0.75↔1.0 | 공허했던 AC-AEL-008 이 폐기되고, AC-AEL-006 이 문구 대조에서 **레시피 실행 여부**로 옮겨졌으며(내가 make 버전 문제를 재현해 확인), AC-AEL-003 에 기수·바이너리 부재 두 게이트가 붙었다. 7건 전부 실행 가능한 판정 명령을 갖는다. 감점은 REQ-AEL-004 의 v0.3.0 신규 조항(doctor 도달 + CI 금지)에 대응하는 판정 단계가 **한 줄도 없다**는 것 |
| Traceability | 0.85 | 0.75↔1.0 | 7 AC 헤딩 전부가 대응 REQ id 를 병기한다. REQ 7건 전부 피복, 고아 AC 0건, 두 문서를 겹쳐 읽을 필요가 사라졌다(D5 종결). 감점 사유는 REQ 단위가 아니라 **조항 단위** — REQ-AEL-004 의 doctor·CI 조항이 어떤 AC 에도 닿지 않는다 |

조화평균 = 4 / (1/0.75 + 1/0.80 + 1/0.90 + 1/0.85) = 4 / 4.87091 = **0.8212 → 0.82**.

---

## D1-D6 델타 판정

### D1 — 종결. 단, 결정의 두 근거는 **각각 참**이다 (내가 대조했다)

마커 제거는 위 MP-7 로 확증했다. 결정과 근거는 `plan.md:44-49` 에 기록됐고 요구 층 반영은 `spec.md:75`(REQ-AEL-004 말미) + `spec.md:77` 주석이다. 근거 두 개를 각각 판정했다.

**근거 1 — "(ii) CI 빌드 잡은 기계적으로 탈락한다": 논증 성립.**
CI 는 검사 대상 커밋을 체크아웃해 그 커밋에서 빌드하므로, 빌드된 바이너리의 `//go:embed` 바이트 = 그 커밋의 온디스크 `templates/` = 그 커밋의 커밋 산출물이다. 양변이 같은 원본에서 파생되므로 비교는 정의상 참이고 **실패할 수 없다** — iter-1 에서 내가 재현한 실측 8 의 동어반복과 **구조적으로 동형**이다(`go test` 가 매번 새 바이너리를 컴파일하는 것과, CI 가 매번 새 바이너리를 빌드하는 것이 같은 역할을 한다). 이 SPEC 이 겨눈 창은 **노후한 로컬/설치 바이너리**이고 CI 에는 그런 바이너리가 존재하지 않는다. 거짓 전제 위에 선 결정이 아니다.

한 가지 구분은 정확하다: CI 가 **소스 층** 드리프트를 잡는 것(실측 2)은 여전히 참이며, 근거 1 은 그것을 부정하지 않고 **임베드 축**만을 두고 말한다. SPEC 도 그렇게 쓰여 있다.

**근거 2 — "(iii) 는 이 코드베이스의 기존 처방이다": 인용이 정확하다.**

```console
$ sed -n '5,10p' internal/cli/doctor_mcp_version.go
// The host spawns the MCP server once per session and never respawns it on
// reinstall, so `make install` leaves the host talking to the previous build.
// Nothing surfaces that: tools/list simply lacks the new tools and the old
$ grep -n '"check"' internal/cli/doctor.go
58:	doctorCmd.Flags().String("check", "", "Run a specific check only (e.g., git, go, config)")
$ grep -n 'accepted by\|mcpServerVersionCheckName = ' internal/cli/doctor_mcp_version.go
25:// accepted by `moai doctor --check`).
26:const mcpServerVersionCheckName = "MCP Server Version"
$ grep -n 'func commitsMatch' internal/cli/doctor_mcp_version.go
99:func commitsMatch(a, b string) bool {
```

인용 4건(`doctor.go:58`, `doctor_mcp_version.go:24-26`, `checkMCPServerVersionAgainst`, `commitsMatch`) 전부 실재하며 행 번호까지 맞다. 인용문도 축자 일치한다. "같은 병(장수하는 산출물이 이전 빌드를 이고 있음), 같은 처방 형태"라는 유비도 성립한다.

**D1 판정: CLOSED. 결정은 거짓 전제 위에 있지 않다.** 다만 이 결정이 **새 결함 하나를 데려왔다** — D7 참조.

### D2 — 폐기 판정 성립. 세 논거 모두 참

저작자의 논거는 iter-1 의 내 처방(a)보다 강하고, 세 갈래를 각각 검증했다.

```console
$ sed -n '218,221p' "$(go env GOROOT)/src/os/file.go"
	if n != len(b) {
		err = io.ErrShortWrite
	}
$ sed -n '97,99p' internal/template/agentemit/golden_test.go
			if err := os.WriteFile(committedTOMLPath(p), data, 0o644); err != nil {
				t.Fatalf("update write %s: %v", p, err)
			}
```

1. **권한 실패** — iter-1 에서 내가 `chmod 444` 로 실행 확증했다. 미수정 코드에서 이미 exit≠0 + 경로 지목.
2. **부분 쓰기** — `os.WriteFile` 은 `f.Write` 를 거치고, `File.Write` 가 `n != len(b)` 에서 `io.ErrShortWrite` 를 돌려준다(위 인용). 그 오류를 `golden_test.go:98` 의 `t.Fatalf` 가 그대로 세운다. **인용이 정확하고 논증이 성립한다.**
3. **잔여 클래스의 도달 불가** — "쓰기가 오류 없이 끝났는데 저장된 바이트가 다르다"는 같은 프로세스 내 되읽기로 판정할 수 없다. `os.ReadFile` 은 페이지 캐시를 읽으므로 방금 쓴 바이트를 돌려주며, 디스크에 실제로 무엇이 있는지 묻지 못한다. 즉 REQ-AEL-008 이 지시하던 가드는 **내구성을 검증하는 것처럼 보이되 검증할 수 없는** 가드다. 이는 이 SPEC 이 근절 대상으로 선언한 공허성의 정의 그 자체이므로, (a) 로 갔다면 공허한 AC 를 공허한 REQ 로 바꾸는 셈이었다는 저작자의 판단이 옳다.

**D2 판정: CLOSED. 폐기는 실제 요구를 제거한 것이 아니다** — 제거된 것은 도달 불가능한 잔여 클래스를 겨눈 가드다. 카운트 7/7 로의 하락은 정당하다. (내 iter-1 처방을 저작자가 뒤집었고, 뒤집은 쪽이 옳다.)

### D3 — 종결

`spec.md:75` — REQ-AEL-004 가 이제 세 가지를 명문화한다: (1) 비교한 경로 수 보고 의무, (2) 그 수가 커밋 산출물 수보다 **적으면** exit≠0, (3) 판정 대상 바이너리 부재 시 **failure, never success**. `spec.md:130-138` — AC-AEL-003 이 두 게이트를 각각 실행 명령으로 갖는다(기수: `ls … | wc -l`, 부재: `BIN=/nonexistent/moai make embed-check`). 기수 기준값도 이 트리에서 재확인했다.

```console
$ ls internal/template/templates/.codex/agents/moai/*.toml | wc -l
      11
$ sed -n '284,286p' internal/template/agentemit/golden_test.go
	if count != 11 {
		t.Errorf("committed .codex/agents/moai carries %d TOMLs, want 11", count)
```

인용(`golden_test.go:285`)도 정확하다. **CLOSED.**

### D4 — 종결

`spec.md:69` — REQ-AEL-002 의 주어가 "Every check this SPEC introduces — the source-layer drift check of M2 and the embed-axis check of M1 alike" 로 확대돼 **양쪽을 문면으로 구속**한다. `plan.md:40` 은 종전의 양보적 표현("같은 규율")을 버리고 "REQ-AEL-002 가 이 검사를 **직접** 구속한다"로 바뀌었다.

```console
$ grep -n '같은 규율' .moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/plan.md
(0 히트)
```

**CLOSED.**

### D5 — 종결

7개 AC 헤딩 전부가 대응 REQ id 를 병기한다(`spec.md:95,108,121,145,151,161,180`). 사상 전량 검사:

| REQ | 피복 AC |
|---|---|
| 001 | AC-001, AC-004 |
| 002 | AC-002, AC-004 |
| 003 | AC-002, AC-005 |
| 004 | AC-003 |
| 005 | AC-003 |
| 006 | AC-007 |
| 007 | AC-006 |

미피복 REQ 0건, 고아 AC 0건, 존재하지 않는 REQ 를 가리키는 AC 0건. **CLOSED.**

### D6 — 저작자의 정정이 옳다. 내 iter-1 처방이 틀렸다

저작자가 내가 제시한 대체 문구(`make: 'agents-emit' is up to date.`)를 측정으로 반박했다. 이 머신에서 재현했다.

```console
$ make --version | head -1
GNU Make 3.81
$ touch agents-emit; make agents-emit; echo "exit=$?"; rm -f agents-emit
make: `agents-emit' is up to date.
exit=0
$ touch agents-emit; make agents-emit 2>&1 | head -1 | od -c | head -2; rm -f agents-emit
0000000    m   a   k   e   :       `   a   g   e   n   t   s   -   e   m
0000020    i   t   '       i   s       u   p       t   o       d   a   t
```

`od -c` 가 **비대칭 쌍**을 확정한다 — 여는 쪽 backtick(0x60), 닫는 쪽 apostrophe(0x27). GNU Make 4.x 는 대칭 인용을 내므로, 어느 문자열을 박아도 다른 make 에서 깨진다. **문구를 판정 기준에서 뺀 것은 정확한 교정이다.**

판정 가능성은 유지되는가 — 그렇다. 새 Then 은 "레시피가 실제로 실행되는지"이고, 판정 신호를 `AGENTEMIT_UPDATE=1 go test …` 출력의 출현으로 특정한다. `Makefile:27-28` 의 레시피 행이 `@` 접두 없이 쓰여 있어 make 가 명령행을 그대로 에코하므로, 이 신호는 관측 가능하다.

```console
$ sed -n '27,28p' Makefile
agents-emit: ## Regenerate the .codex/agents/moai TOMLs from the neutral .md layer
	AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission
```

Control(RED) 도 `spec.md:163-168` 에 실측 콘솔로 박혔다. **CLOSED.**

---

## 회귀 검사

| 항목 | 결과 | 근거 |
|---|---|---|
| REQ 001-007 결번·중복 없음, 008 부재 | PASS | 위 MP-1 |
| AC 7건, 전 REQ 피복, 고아 0 | PASS | 위 D5 표 |
| `progress.md` §E.1 정합 | PASS | `progress.md:8` "REQ **7** / AC **7**", `:10` "미해결 결정: **없음**" + doctor 편입 명시 — 현재 카운트·결정 상태와 일치 |
| Tier S 7/7 및 상한 인용 | PASS | `spec.md:224` / `plan.md:18` 모두 "상한 각 8, 두 축 독립". SSOT(`spec-workflow.md:140`) 대조 일치 |
| 산출물 집합 = Tier S | PASS | spec.md + plan.md(+ 워크플로 공통 progress.md). acceptance.md 없음, AC 는 §3 인라인 |
| 범위 밖 영역으로의 누수 | PASS | C1↔C2 를 대상 삼는 REQ·AC 0건, `graph-freshness` 요구화 0건, `catalog.yaml` 쓰기는 AC-AEL-004 가 명시적으로 판정 대상에서 제외(`spec.md:149`) |
| iter-1 통과 항목 파손 | 없음 | 동어반복 전제·"build 선행 금지" 조항(`plan.md:32`)·정정된 전제 서술 전부 온존 |

---

## 발견된 결함

### D7 — `moai doctor` 편입이 배포 사용자에게 무엇을 하는지 요구 층이 침묵한다 (신규)

`spec.md:75`(REQ-AEL-004) — **Severity: major — Class: blocking**

REQ-AEL-004 는 한 문장 안에서 두 가지를 동시에 못박는다: (α) 판정 지점은 `moai doctor` 항목으로 **도달 가능해야 한다**, (β) 판정 대상 바이너리가 없으면 **실패로 종료한다**. 두 조항 각각은 정당하다. 그러나 겹쳐 놓으면 배포 사용자 표면에서 충돌하고, SPEC 은 그 지점을 다루지 않는다.

측정 — 배포된 사용자 프로젝트를 만들어 상태를 쟀다(스크래치, 저장소 트리 밖):

```console
$ moai init <scratch>/proj --non-interactive; echo "init_exit=$?"
init_exit=0
$ ls -d <scratch>/proj/bin/moai
ls: .../proj/bin/moai: No such file or directory
$ ls -d <scratch>/proj/internal/template/templates/.codex
ls: .../proj/internal/template/templates/.codex: No such file or directory
$ cd <scratch>/proj && moai doctor >/dev/null 2>&1; echo "doctor_exit=$?"
doctor_exit=0
```

세 가지가 동시에 참이다. (1) 배포 프로젝트에 `bin/moai` 는 **없다** — `plan.md:33` 이 판정 대상 기본값으로 못박은 바로 그 경로다. (2) 커밋 산출물의 정의 경로(`spec.md:38` 의 C3 = `internal/template/templates/.codex/agents/moai/*.toml`)도 **없다** — 그것은 이 저장소의 소스 트리다. (3) `moai doctor` 는 지금 **exit 0** 이다.

그리고 doctor 는 Fail 을 종료 코드로 승격한다:

```console
$ sed -n '48,49p' internal/cli/doctor.go
	Long: "Run comprehensive system health checks ...\n\n" +
		"Exit codes: 0=no failing checks (warnings are advisory and do not fail the run), 1=one or more checks reported Fail.",
```

따라서 REQ-AEL-004 를 문면대로 구현하면 — doctor 항목으로 등록 + 대상 부재 시 실패 — **모든 배포 프로젝트에서 `moai doctor` 가 exit 1 로 뒤집힌다.** 이는 `plan.md:49` 가 운영자에게 제시한 수용 비용("`moai doctor` 출력에 **행 1개**가 추가된다. 그 이상은 없다")과 어긋난다. 실패 행 하나는 출력 한 줄이 아니라 종료 코드 변경이다. 운영자는 이 형태의 비용을 승인한 적이 없다.

합리적 엔지니어가 셋 중 무엇으로도 구현할 수 있다는 점이 이 결함의 실체다: (가) 문면대로 → 사용자 doctor 붕괴, (나) doctor 항목만 부재를 OK 로 처리 → verb 와 doctor 가 같은 로직을 공유한다는 `plan.md:34` 위반, (다) doctor 항목은 설치 바이너리 ↔ **프로젝트에 배포된** `.codex/`(측정: 11건 존재)를 대조 → 사용자에게 유의미한 검사가 되지만 SPEC 이 정의한 "커밋 산출물"과 다른 대상이다. 셋의 사용자 가시 결과가 서로 다르다.

`measurement.md` 측정 10 이 같은 비대칭의 절반(선례 항목은 부재를 OK 로 낸다)을 이미 관측하고 "run-phase 가 의도된 것으로 명시하는 편이 좋다"고 적었다. 나는 그 처방이 부족하다고 판정한다 — 이것은 표기 문제가 아니라 **요구 층의 미결 결정**이고, iter-1 의 D1 이 그러했듯 배선 세부가 아니라 산출물의 값어치를 좌우한다.

부수적으로 **판정 공백도 있다**: REQ-AEL-004 의 v0.3.0 신규 조항 셋(doctor 도달 / verb 도달 / CI 빌드 잡 금지) 중 어느 것도 AC 단계로 내려오지 않았다. AC-AEL-003 은 `make embed-check` 만 부른다. 실패시킬 수 없는 조항이 요구 층에 앉아 있다.

**필요 조치**: REQ-AEL-004 의 doctor 조항에 **대조 대상과 부재 거동을 배포 문맥까지 포함해** 한 문장으로 못박는다(예: doctor 항목은 저장소 소스 층이 존재할 때만 판정하고 그 밖에서는 항목을 내지 않는다 / 또는 배포된 `.codex/` 를 대상으로 삼는다 — 어느 쪽이든 명시). 동시에 AC-AEL-003 에 doctor 도달을 재는 한 줄을 더한다(`moai doctor --check "<항목명>"` 이 exit 0/≠0 을 의도대로 낸다 — 측정 10 이 이 호출 형태가 실제로 단일 항목만 남긴다는 것을 이미 세웠다). Tier S 항목 수는 늘지 않는다(기존 REQ·AC 본문 확장).

### D8 — v0.3.0 이 새 구현 표면을 들이고 범위 필드를 갱신하지 않았다 (신규)

`spec.md:11` (`module:`) + `plan.md:13` (§B 파일 수) — **Severity: minor — Class: blocking**

v0.3.0 은 `internal/cli` 에 doctor 항목을 등록하도록 요구했는데, `module:` 은 여전히 `"Makefile, internal/template/agentemit, internal/template/templates/.codex/agents/moai, CLAUDE.local.md"` 로 `internal/cli` 를 담지 않는다. `plan.md:13` 의 영향 파일 추정 "3-5" 도 doctor 등록 전 값 그대로다.

이것이 형식 흠결이 아닌 이유: Tier S 기준은 **`< 5 files`** 이고(`spec-workflow.md:140`), 추정 상단 "5" 는 이미 그 경계 밖이다. 여기에 doctor 등록(신규 검사 파일 + `doctor.go` 등록 편집, 통상 테스트 1건 동반)이 더해졌으므로 추정은 다시 계산돼야 한다. SPEC 스스로 "범위가 늘면 **Tier 를 다시 판정한다**"(`plan.md:18`)고 적어 놓고, 범위를 늘린 개정에서 그 재판정을 하지 않았다.

**필요 조치**: `module:` 에 `internal/cli` 추가, `plan.md §B` 파일 수 행을 doctor 등록 포함으로 재산출하고 그 값이 `< 5` 를 지키는지 명시. `≥ 5` 면 Tier M 승격(= acceptance.md 분리 + 임계 0.80)을 판정한다.

### 이월 없음

iter-1 의 D1-D6 은 **6건 전부 RESOLVED**. 3회 연속 잔존하는 정체(stagnation) 결함 없음.

---

## 회귀 표 (iter-1 결함 처리)

| iter-1 결함 | 상태 | 근거 |
|---|---|---|
| D1 미해결 clarification 마커 | **RESOLVED** | `grep -rn 'NEEDS CLARIFICATION' <spec dir>` → rc=1. 결정 + 두 근거 기록, 근거 인용 4건 코드 대조 일치 |
| D2 AC-AEL-008 공허 | **RESOLVED (폐기 경로)** | REQ/AC 008 정의 0건. 폐기 논거 3갈래 각각 검증 통과 |
| D3 공허성 봉투 미명세 | **RESOLVED** | REQ-AEL-004 기수·부재 조항 + AC-AEL-003 두 게이트 |
| D4 REQ-AEL-002 주어 | **RESOLVED** | 주어 확대 + `plan.md` 양보 표현 제거(grep 0) |
| D5 간접 사상 | **RESOLVED** | 7/7 헤딩에 REQ id 병기, 사상 전량 검사 통과 |
| D6 대조군 부재 + 문구 오류 | **RESOLVED (문구는 재정정)** | Control 실측 삽입 + 판정을 레시피 실행 여부로 이동. make 3.81 비대칭 인용 `od -c` 로 재현 |

---

## 감사 중 트리 변경 (전량 원복 확인)

이 감사는 뮤테이션을 심지 않았다. 수행한 쓰기는 (1) 저장소 루트에 `agents-emit` 빈 파일 2회 생성 후 즉시 삭제, (2) 세션 scratchpad 에 `moai init` 배포물 1건 생성 후 삭제뿐이다.

```console
$ git status --short
?? .moai/reports/t317/
?? .moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/
$ go test ./internal/template/agentemit/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.472s
```

추적 파일 수정 0건. `agents-emit` 잔재 없음(`git status` 에 미등장). 스크래치 배포물은 저장소 트리 밖에만 만들었고 제거를 확인했다.

---

## Gaps — 관측하지 않은 것

- `make build` 를 실행하지 않았다(트리를 변형하는 `catalog.yaml` in-place 쓰기와 `*_templ.go` 생성을 유발하지 않기 위함). AC-AEL-001/004 의 실제 build 거동은 이번에도 판정하지 않았다.
- D7 의 결과를 **가정적으로** 계산했다 — 아직 존재하지 않는 검사이므로, doctor 가 exit 1 로 뒤집히는 것을 실행으로 관측할 수는 없다. 내가 실행으로 세운 것은 그 결론의 세 전제(배포 프로젝트에 `bin/moai` 부재, 커밋 산출물 경로 부재, doctor 가 Fail 을 exit 1 로 승격)뿐이며, 셋 다 인용된 출력으로 붙였다.
- `moai doctor` 전체 실행의 항목 목록을 열거하지 않았다. exit code 만 쟀다.
- 11개 TOML 전부에 대한 재검증은 하지 않았다(기수 11 만 재확인).
- `measurement.md` 측정 9 의 배포 추출은 이 세션에서 재현하지 않았다 — iter-1 에서 내가 이미 독립 실행했고, 측정 9 가 그 결과와 일치함을 문서 대조로만 확인했다.

## Residual risk

- D7 을 요구 층에서 닫지 않고 run-phase 재량으로 넘기면, 세 구현안 중 (가)가 선택될 때 배포 사용자의 `moai doctor` 가 전면 실패한다. 그 회귀는 이 저장소 CI 에서 보이지 않는다 — CI 는 이 저장소 안에서 돌고, 거기에는 `bin/moai` 도 소스 층도 존재하기 때문이다. 즉 **이 SPEC 이 겨눈 것과 같은 형태의 눈먼 창**을 하나 새로 열 수 있다.
- D8 을 방치하면 Tier 판정 자체가 스테일한 추정 위에 서고, 파일 수가 5 를 넘길 때 임계 0.75 로 감사받은 SPEC 이 실제로는 0.80 대상이 된다.
- 이번 감사는 Tier S 상한(1회)을 초과한 2회차다. 다음 반복은 3회 상한에 걸리므로, D7·D8 을 닫은 뒤에는 델타 재감사 1회가 마지막이다.

---

## 권고

1. **D7 을 요구 층에서 닫는다.** REQ-AEL-004 의 doctor 조항에 배포 문맥의 대조 대상과 부재 거동을 명시하고, `plan.md:49` 의 수용 비용 서술을 그 결과에 맞춰 정정한다. 운영자가 승인한 것은 "행 1개"이므로, 종료 코드가 바뀌는 안을 고를 경우 재승인이 필요하다.
2. **AC-AEL-003 에 doctor 도달을 재는 한 줄을 더한다** — `moai doctor --check "<항목명>"`. 측정 10 이 이 호출 형태의 유효성을 이미 세웠으므로 판정 명령은 이미 손에 있다.
3. **D8 을 닫는다** — `module:` 에 `internal/cli` 추가, `plan.md §B` 파일 수 재산출 + Tier 재확인.
4. 1-3 은 전부 기존 본문 확장이므로 REQ/AC 개수를 늘리지 않는다. Tier 재판정이 필요한 축은 **파일 수**이지 항목 수가 아니다.
5. 이 SPEC 은 **현 상태로도 run-phase 진입 가능하다**(점수·must-pass 모두 통과). 다만 D7 은 착지 전에 닫혀야 하는 부채로 명시적으로 이월한다 — doctor 항목 등록 커밋이 D7 결정 없이 나가면 안 된다.
