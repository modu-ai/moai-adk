# SPEC-EVIDENCE-CITATION-CANON-001 — 진행 기록

카드: t375 · 워크트리 `.claude/worktrees/t375` · 브랜치 `WT-state-evidence-canon` · 기준 HEAD `b64043481`

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (조건부 — plan.md §B. M2가 §C.3을 결정한 직후 재확인 의무 있음)
- 산출물: spec.md · plan.md · acceptance.md · progress.md
- REQ 13건 (상한 16) / AC 15건 (상한 16)
- SPEC ID 정규식 검사: Bash 실행, 출력 `PASS`
- ID 중복 검사: `find .moai/specs -maxdepth 1 -type d -name 'SPEC-EVIDENCE*'` → `SPEC-EVIDENCE-CLAIM-INVARIANT-001` 1건만, 충돌 없음

### iter1 감사 수리 (2026-08-31)

`.moai/reports/t375/plan-audit-iter1.md` — FAIL 0.70, blocking 11건. 전부 수리했다.

| 결함 | 처리 |
|---|---|
| D1 하한 7이 모집단보다 두 자릿수 작음 | 트리별 하한 300으로 교체, 도출 명령을 AC 본문에 기재 (acceptance AC-ECC-010, plan §D.2) |
| D2 REQ-ECC-008 판정 AC 부재 | **AC-ECC-015 신설** — 방문 트리 목록 + 트리별 하한 + 미러 정합 + 방문 뮤테이션 |
| D3 131 미귀속 | 스크립트 도출값 124로 교체, 경위를 spec §1.1.1에 기록 |
| D4 §1.1 명령 2행이 자기 값을 못 만듦 | 출처를 커밋된 스크립트 2개로 교체, "532 정정 노트" 삭제, 추적 집합 사용 이유 명시 |
| D5 carve-out 누락 | `internal/web/events.go:29` 추가, §1.4 "하나"→"둘", AC-ECC-005·006 확장 |
| D6 AC-ECC-002가 오늘 통과 | 고정 문구 3종 grep으로 교체 (오늘 전부 0회 확인) |
| D7 허용목록 단위 미지정 | 파일 + 정확한 리터럴로 규정, "파일 전체 면제 금지" 단언 추가 (REQ-ECC-009, AC-ECC-013) |
| D8 `.gitkeep` 귀결 미명시 | **판정 변경** — `.gitkeep`·`!` 예외 줄 모두 두지 않음 (spec §4.2) |
| D9 판별식 미적용 | **적용했고 감사 제안과 다른 결론** — navigator 아래 전부 무시하지 않음 (spec §4.3) |
| D10 AC 3건 요구 층 공백 | REQ-ECC-002 주어를 doctrine 표면 문서로 확대 + REQ-ECC-013 신설 |
| D11 인용 넓이 상한 부재 | REQ-ECC-004에 "파일 하나를 이름 붙인다" 상한 통합 |
| D12 Tier 최대 케이스 미계산 | 11–18 조건부 + M2 재확인 의무 (plan §B) |
| D13 세션 디렉터리 124 | "최상위 124 = 세션 123 + snapshots 1"로 정정 |
| D15 부채 수치 오염 | 124 / 231 / 189로 갱신, `mcp_glm.go:110` 후속 후보 추가 |

**감사 소견 중 따르지 않은 것 1건 — D9.** 감사는 `fix-drafts/`의 잔여가 "요청됐으나 완료되지 않은 위임"을 뜻하므로 t373의 `chain/`과 같은 모양이라고 제안했다. §4.1 판별식의 숨은 전제(성공 경로에서 처분하는 코드가 있어야 잔존이 실패를 뜻한다)를 검사한 결과 그 전제가 성립하지 않는다:

```
grep -rn 'RemoveAll' internal/navigator/fix/ internal/cli/navigator_fix.go   (테스트 제외) → 0행
```

처분 코드가 없으므로 잔존은 성공한 실행에서도 남고 아무 신호가 아니다. 다만 결론은 감사가 원한 방향과 같다 — 무시하지 않는다 — 근거가 다르다(내용 기반 신호 `applied.json` + 존재-부재 논거 대칭). 상세는 spec.md §4.3.

### iter2 감사 종결 (2026-08-31) — PASS-WITH-DEBT 0.85

`.moai/reports/t375/plan-audit-iter2.md` — Tier M 임계 0.80 상회, iter1 대비 +0.15 단조. 재감사 상한 도달로 종결 판정. iter1 13건 중 11 종결 / 2 부분 종결 / 미변경 0. **감사가 iter1 D9를 철회했다** — §4.3의 반박이 옳다고 판정.

부채 4건을 이 판에서 닫았다.

| 항목 | 처리 |
|---|---|
| N1 (blocking) | 리드가 스크립트에 `SPEC_OWN_DIR` 배제를 추가·검증. SPEC 쪽 몫은 §1.1.2 신설 — 배제 이유(추적 집합 논거의 유효 기간)와 불변성 실측표(188 in → 184/515/346/124/231 무변). "커밋된"을 "이 카드가 plan-close로 함께 추적한다"로 정정 |
| D1 잔여 | §D.2.1 + AC-ECC-015 #2 — 하위트리 **집합 상등** 단언. 하한 300은 agents(21)·output-styles(3) 소실을 통과시키고, 그 둘이 `manager-lead.md`와 배너 3지점을 담는다. 독립성 시연(하한 통과 + 상등 실패) 포함 |
| D7 잔여 | AC-ECC-013 #1에 판정 명령 부여 — grep이 아니라 **빈 리터럴 항목 거부 뮤테이션 서브테스트**. grep은 오늘 목록에 대해서만 참이라 내일 추가될 항목을 막지 못한다 |
| N2 | §1.1에 반영 — 두 스크립트가 판별식을 각자 리터럴로 정의하므로 "갈라질 수 없다"가 아니라 "함께 고치고 둘 다 돌린다" |
| N3 | AC-ECC-010·plan §D.2의 미러 하한 도출을 4개 하위트리 열거로 교정(340 → **338**). 범위 밖 2개 명시 |
| N4 | §1.1.1에 단위 다리 신설 — 532는 맨 문자열, 515는 뒤에 `/`가 오는 형태, 차이 17 |
| N5 | REQ 재번호 — 013(경계 사례 기록)을 011로, 011·012(ignore 2건)를 012·013으로. 정의 순서가 001…013 단조가 되고 절 구성(§2.2 가드 / §2.3 ignore)도 유지된다. acceptance 매트릭스·plan 참조 동반 갱신 |

시연 방향이 4 → **6**으로 늘었다(하위트리 방문 · 허용목록 단위 추가).

## §F Phase 4 Mode Selection

Decision: **serial** (단일 sub-agent 순차 spawn)

**입력 파라미터** (t375 워크트리, base `b64043481`, 판정 시점 `origin/develop` = `3f03d9c36`)

| 항목 | 값 |
|---|---|
| tier | M |
| scope (파일 수) | 11~18 — §C.3 경계 3건이 carve-out이냐 교체냐로 갈림 |
| domain count | 3 (doctrine markdown / Go 가드 테스트 / `.gitignore` + 템플릿 미러) |
| file language mix | markdown 다수 + Go 1파일 + 방출 `.toml` 1 |
| concurrency benefit | LOW — 편집 대상들이 서로를 참조한다(규칙 본문 문구가 AC의 고정 문자열 grep을 결정하고, 미러는 원본을 따른다) |

**모드 평가**

| 모드 | 선택 | 사유 |
|---|---|---|
| `direct` | 미선택 | 오타·1줄 수정이 아니다. Go 가드 테스트 신설 + 문서 다수 편집 |
| `serial` | **선택** | 아래 |
| `fanout` | 미선택 | 다중 도메인이되 research-heavy가 아니라 coding/authoring-heavy다. 게다가 편집이 상호 의존한다 — M1이 정하는 고정 문구가 AC-ECC-002/003의 grep 대상이고, 미러(M6)는 M1·M3 결과를 복사한다. 병렬 쓰기 금지 규율에도 걸린다 |
| `sweep` | 미선택 | ~30파일 문턱 미달(11~18)이고, 단일 균일 변환 규칙이 아니다. 규칙 본문 개정·Go 테스트 신설·ignore 편집이 각각 다른 변환이다 |

**Decision: serial**

**정당화.** 이 카드의 편집은 한 방향으로 흐른다 — M1이 규칙 본문의 고정 문구를 정하면 그 문자열이 AC의 판정 명령이 되고, M3이 지시·출력스타일을 같은 문구로 맞추며, M6이 그것을 템플릿에 미러하고 `make agents-emit`으로 `.toml`을 방출한다. 앞 단계의 산출이 뒤 단계의 입력인 구조에서 병렬 fan-out은 이득이 없고, 같은 파일 계열에 두 writer를 붙이는 위험만 남는다. Anthropic의 coding-task parallelism caveat과도 같은 방향이다.

**경계 사례.** scope 11~18은 Tier M 구간(5~15)의 상단을 넘길 수 있다. plan.md §B가 조건부로 적었고, **M2 직후 Tier 재확인**이 [HARD]로 걸려 있다 — §C.3 경계 3건이 전부 교체로 결정되면 18이 되어 Tier L 재분류를 검토한다. 그 결정 전에 Tier를 확정하지 않는다.

## §E.2 Run-phase Evidence

측정 트리: 워크트리 `.claude/worktrees/t375`, 브랜치 `WT-state-evidence-canon`, 진입 HEAD `27df6bafb`.

### M2 — §C.3 경계 3건 판별식 적용 (REQ-ECC-011 / AC-ECC-009)

판별식은 "이 경로를 **최종적으로 읽는 것**이 사람인가 기계인가"다. 세 건 모두 산문이 아니라 **소비자 코드**를 읽어 판정했다.

| 파일 | 판정 | 이유 (관측) |
|---|---|---|
| `gate.md:122` | **carve-out** | `moai verify record`가 HEAD SHA로 키가 걸린 스냅샷 저장소에 쓰고, `moai verify check --key-current`가 같은 키로 읽는다. 사람이 이 경로를 판정 근거로 열지 않는다. 사전 판단과 일치 |
| `loop.md:115` | **carve-out** | 같은 저장소를 loop 완료 평가기가 기계적으로 읽는다. 문장 자신이 "mechanical read surface"라 적는다. 사전 판단과 일치 |
| `run.md:199` | **교체** | 사전 판단(경계 사례)이 옳았고, 코드가 방향을 정했다. `internal/harness/routing/outcome.go`의 `FinalizeOutcome`은 evidence **레코드의 `Kind`**만 보고 `KindVerifyPath` → `OutcomeSuccess`로 확정한다 — `--ref` 경로를 **여는 코드가 없다**(`grep -rn 'verify_path' --include='*.go' internal/` → 정의·enum 검증·테스트뿐). 즉 그 경로는 기계가 키로 조회하는 값이 아니라 **나중에 사람이 원장을 감사할 때 읽으라고 적어 둔 인용**이다. 판별식이 곧바로 "사람"을 가리킨다 |

### M2 — Tier 재확인 [HARD]

교체가 **1건**이므로 plan.md §B의 최대 케이스(3건 전부 교체 → 18파일)는 발생하지 않았다.

실제로 손댄 파일을 셌다(SPEC 산출물 제외):

```
git diff --name-only HEAD | grep -v '^\.moai/specs/' | grep -v '^\.moai/reports/' | wc -l
```

| 부류 | 파일 |
|---|---|
| 규칙 본문 (C.1) | `agent-common-protocol.md`, `agent-common-protocol-reference.md` |
| doctrine 표면 (C.2) | `manager-lead.md`, `output-styles/moai/moai.md` |
| 경계 교체 (C.3) | `skills/moai/workflows/run.md` |
| `.gitignore` (C.4) | 루트 + 템플릿 미러 |
| 가드 (C.5) | `internal/template/evidence_citation_guard_test.go` (신규) |
| 미러·방출 (C.6) | 위 5개의 `internal/template/templates/` 미러 + `.codex/…/manager-lead.toml` + `catalog.yaml` |

**판정: Tier M 유지.** Tier L 상향도 SPEC 분할도 하지 않는다. 근거 둘 — (a) 파일 수가 Tier M 구간(5–15)의 상단 안에 있고, 최대 케이스 18은 실현되지 않았다. (b) 질적 기준이 변하지 않았다: 새 실행 기제를 만들지 않았고(가드는 테스트 하나), 헌법 문서를 건드리지 않았다. `catalog.yaml`은 `make build`가 다시 만드는 파생물이라 판단 대상이 아니다.

### AC 매트릭스 (AC-ECC-001 ~ 015)

| AC | 판정 | 결정 명령 | 관측 출력 |
|---|---|---|---|
| 001 | **PASS** | `grep -c 'SHALL be persisted under'` (reference) / `grep -c 'evidence is persisted under'` (protocol) | `0` / `0` (사전 `1` / `1`) |
| 002 | **PASS** | `grep -c` × 3 (protocol) | `machine-local scratch`=1, `export before citing`=1, `.moai/reports/<card-id>`=1 (사전 전부 0) |
| 003 | **PASS** | `grep -c 'names one file'` / `grep -c -i 'never the directory\|wholesale'` (reference) | `1` / `1` (사전 0 / 0) |
| 004 | **PASS** | `grep -c -i 'residual-risk'` (reference) | `1` (사전 0) |
| 005 | **PASS** | `grep -c 'state/verify/snapshots'` / `grep -c 'internal/web/events.go'` (reference) | `1` / `1` (사전 0 / 0) |
| 006 | **PASS** | `go test ./internal/verify/... ./internal/web/...` | `ok` 양쪽, rc=0 |
| 007 | **PASS** | `grep -c 'canonical persistence location'` / `grep -c 'machine-local scratch'` (manager-lead) | `0` / `2` (사전 1 / 0) |
| 008 | **PASS** | `grep -c 'state/verify'` (output-style) | `0` (사전 3) |
| 009 | **PASS** | 이 절의 M2 표 3행 + `grep -c 'gate.md\|loop.md\|run.md' progress.md` | 3건 모두 판정·이유 기록 |
| 010 | **PASS** | `go test -v -run EvidenceCitation ./internal/template/` | `repo-root: scanned 363`, `template-mirror: scanned 338`, 양쪽 위반 0 (하한 300) |
| 011 | **PASS** | `TestEvidenceCitation_SyntheticMutant` | 위반 정확히 1건, 2행(옛 형태)만 |
| 012 | **PASS** | `TestEvidenceCitation_RealSentenceMutant` | 수리 이전 `agent-common-protocol.md:268` 문장 리터럴이 잡힌다 |
| 013 | **PASS** | `TestEvidenceCitation_Allowlist/{Unit,Size}` | 빈 리터럴·빈 파일 항목 거부, 항목 수 9 = 상수 |
| 014 | **PASS** | `grep -n 'moai/observability'` 양쪽 / `grep -c 'navigator'` 양쪽 / `find … -type d -name observability` | 각 1행(`*.jsonl`), `!` 줄 0, navigator 0 / 0, 디렉터리 0행 |
| 015 | **PASS** | `TestEvidenceCitation_Visitation` + 두 뮤테이션 | 트리 2개 방문, 하위트리 집합 4개와 상등(양쪽), 트리별 하한 개별 통과, 미러 정합 0=0 |

GAP 없음. 15/15 PASS.

### 여섯 방향 시연 (REQ-ECC-010 / plan §D.5)

반출: `.moai/reports/t375/guard-six-directions.txt`(통과 방향 + 뮤테이션 4·5·2·3·6, 판정 줄만), `.moai/reports/t375/guard-red-empty-allowlist.txt`(허용목록이 빈 상태의 RED).

| # | 방향 | 결과 |
|---|---|---|
| 1 | 통과 | 루트 363 / 미러 338 스캔, 위반 0 |
| 2 | 합성 뮤테이션 | 새 형태 1줄 + 옛 형태 1줄 → 옛 형태만 1건 |
| 3 | 실물 뮤테이션 | 수리 이전 실문장 리터럴 → 1건 |
| 4 | 트리 방문 뮤테이션 | 미러를 빼면 방문 트리 목록 단언이 실패, 하한은 통과 |
| 5 | 하위트리 방문 뮤테이션 | `agents`를 빼면 집합 상등이 실패, **같은 실행에서 하한은 통과**(342 ≥ 300 / 327 ≥ 300) — 두 단언의 독립성 |
| 6 | 허용목록 단위 뮤테이션 | 리터럴이 빈 항목·파일이 빈 항목 둘 다 검증자가 거부 |

**RED 증거의 성격.** 허용목록이 빈 상태에서 가드를 돌려 루트 9건 / 미러 15건을 실제 코퍼스에서 잡아냈고, 그 목록이 곧 허용목록 9개 항목의 입력이 됐다 — 항목은 "무엇이 이 경로를 인용하겠는가"를 헤아려 쓴 것이 아니라 스캐너가 보고한 줄에서 나왔다(§1.1.1이 두 번 미끄러진 바로 그 자리라 방식을 바꿨다). 미러 15건에는 수리 이전 문장(`SHALL be persisted under`, `canonical persistence location`, 배너 3지점)이 그대로 들어 있어, 가드가 **이 SPEC이 없애려는 결함을 실제로 잡는다**는 것이 그 출력으로 확인된다.

### 알려진 잔여 (이 카드가 닫지 않는다)

- `.go` 표면 13건은 가드 밖이다(spec.md §5, 카드 t381). **가드의 초록은 `.md` 표면에 대해서만 참이다.**
- 기존 문서 124개 / 인용 231건의 소급 정정은 범위 밖(spec.md §5).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-31
run_commit_sha: <backfill — this section is written in the same commit it describes>
run_status: complete
ac_pass_count: 15
ac_fail_count: 0
ac_gap_count: 0
preserve_list_post_run_count: 4   # measure_citations.py, measure_resolution.py, plan-audit-iter1.md, plan-audit-iter2.md — 전부 미변경
l44_pre_commit_fetch: not-run     # 레인은 push/merge하지 않는다 (리드가 통합을 소유)
l44_post_push_fetch: not-run      # 동상
new_warnings_or_lints_introduced: 0   # golangci-lint run --timeout=2m → "0 issues."
cross_platform_build:
  darwin_arm64: pass              # go build ./... rc=0
  windows_amd64: pass             # GOOS=windows GOARCH=amd64 go build ./... rc=0
coverage:
  internal/template: 86.3%        # 목표 85% 상회
  internal/template/agentemit: 91.3%
total_run_phase_files: 14         # SPEC 산출물·증거 반출 제외
m1_to_mN_commit_strategy: single-commit   # M1~M6이 서로 의존해 중간 커밋이 초록이 아니다 (mirror parity 단언)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
