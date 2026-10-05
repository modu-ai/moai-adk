# Plan — SPEC-UPDATE-ADD-CODEX-001

## §A 맥락

- 카드: t589 — `moai update --add-codex` 신설 + CLAUDE.md→AGENTS.md 구조 전환 + init --force 재초기화 codex-add 경로 철폐 (운영자 지시 2026-09-09, Class C)
- 형제 카드: t585 (init 사이드 — harness 질문 + 3-way 배포). 이 카드는 **update 사이드만** 소유한다. 단 M3 는 init.go 를 건드리는 유일한 축인데 범위는 안내 출력 + 테스트 갱신뿐이다 — t585 가 소유하는 harness 질문·3-way 배포 표면과는 겹치지 않으며, t585 가 먼저 착지해 호출부 좌표가 움직이면 develop 병합 순서에서 흡수한다.
- 측정 트리: `.claude/worktrees/t589` @ `5caddeb2d`, branch `WT-add-codex-verb`

### 카드 등급 판정

Class C — 신설 동사(동작 결정) + 구조 전환(계약 문서 이동) + 경로 철폐(시맨틱 변경)로 하위 시스템을 넘는다. 3 working columns 전부 유효.

### Tier 판정 — M

- 영향 파일: `internal/cli/update.go`(플래그 + 동사), `internal/cli/update_codex_wiring.go`(ungated 래퍼), `internal/cli/update 버전 충돌 검증기`, `internal/cli/init.go`(안내), `internal/template/templates/AGENTS.md`, `internal/template/templates/CLAUDE.md`, 테스트 4~6종 → **5~15 파일, 신규 LOC < 1,000** → Tier M (artifact 3종 + progress).
- REQ 14건 / AC 14건 — Tier M 상한(16/16) 이내.
- design.md 미설치: 설계 결정 D1-D6 이 spec.md §1.2 + 본 문서 §D 에 담겨 있고, 측정(연구)은 plan-phase 에서 실행 완료 — Tier L 확장 요건이 없다.

## §B 알려진 결함·스테일 인용 (이미 측정됨)

- **B1 — 카드 좌표 `validator.go:258-284` 미해결**: `internal/cli/validator.go` 부재(spec.md §1.3). `.moai/` 백업의 os.Rename 호출부는 run-phase 에서 재탐색한다. M3 는 백업 메커니즘을 변경하지 않으므로(안내+테스트가 전부) 구속 아님.
- **B2 — 카드 좌표 `phase_test`/`validator_test` (internal/cli) 부재**: 실제 고정 테스트는 `init_test.go` + `init_agent_flag_test.go` + wizard/precedence/auth-ladder/tokens 6종(spec.md §1.1 M5). REQ-UAC-014 는 이 목록을 따른다.
- **B3 — update 에 이미 `--force` 가 있다**: 버전 매치 skip 해제 등 update 고유 의미. 새 플래그 `--add-codex` 와 이름 충돌 없음 — 다만 문서·도움말에서 두 --force 를 혼동하지 않게 주의.
- **B4 — `wiringFilesExist` 게이트의 판별 파일은 `.codex/hooks.json` 또는 `.codex/config.toml`** (SPEC-CODEX-WIRING-001 REQ-CW-009 구현 주석). M1 의 ungated 경로는 이 판별을 통과하지 않는다.

## §C 사전 점검 (run-phase 진입 시 재측정 — 값이 다르면 멈추고 보고)

측정 트리 `5caddeb2d` 기준 기대값. 다르면 다른 카드가 트리를 움직인 것이다 — 진행 전 blocker 보고.

```bash
# 1. RED 유지 확인 — 동사가 아직 없어야 한다
rg -n "add-codex" internal/cli/update.go          # 기대: 0 matches (empty)
./bin/moai update --add-codex                     # 기대: exit 1, "Unknown flag: --add-codex."

# 2. 바이트 상한 기준선
wc -c internal/template/templates/AGENTS.md       # 기대: 15415 (≤ 24576)
go test ./internal/config -run TestCodexContractByteCeiling   # 기대: ok

# 3. 스킬 발행 세트
ls internal/template/templates/.agents/skills/    # 기대: 16개 moai-*

# 4. 고정 테스트 재인벤토리 (B2 — 파일 집합이 늘었다면 REQ-UAC-014 범위 확장)
rg -l '"both"' internal/cli/*_test.go
rg -n "force" internal/cli/init_test.go

# 5. B1 클린 재측정 (단순 패턴, 치환 플래그 없이)
rg -n "os\.Rename" internal/cli/

# 6. 영향 패키지 기준선
go build ./... && go test ./internal/cli/... ./internal/codexwiring/... ./internal/config/...
```

추가 확인(미확정 — 첫 측정): 루트 계약 쌍 ↔ 템플릿 쌍 바이트 동일성 강제 가드 존재 여부(`internal/`에서 `templates/AGENTS.md` 참조 스캔). 측정 결과에 따라 M2 의 루트 쌓임 여부를 정한다(§H 위험 1).

## §D 구속 조건 (재논의 금지 — spec.md §1.2 결정 D1-D6 의 구현 쪽 정밀화)

- **D1** `--add-codex` 구현은 `codexwiring.Wire(".", out, errOut)` 직접 호출이다. `refreshCodexWiringBestEffort*`(게이트 버전)를 재사용해 게이트를 우회하는 "플래그 파라미터 추가" 형태로 우회 구현하지 않는다 — 게이트는 `RefreshWiring` 안에 있으므로 래퍼에 boolean을 넘기는 것도 우회 구현이다. 호출 위치는 `update.go:507` `refreshCodexWiringBestEffort` 바로 옆(같은 조건부 자리)이다. **분기 구속: `ErrValidationRefused` 는 하드 오류로 전파한다 — `--add-codex` 경로에서 exit ≠ 0. best-effort(경고 후 계속)가 허용되는 것은 IO 오류뿐이다.** 형제 래퍼 `refreshCodexWiringBestEffortAt`(`update_codex_wiring.go:16-20`)은 모든 오류를 경고로 삼키는 모델이며, 이 동사는 그 모델을 따르지 않는다.
- **D2** `.mcp.json` 무접촉. 어떤 병합·프로비저닝 코드도 `--add-codex` 경로에서 호출하지 않는다.
- **D3** AGENTS.md 에 `MOAI:LEARNED-WORKFLOW` 제목을 만들지 않는다. curator(`internal/harness/curator`)와 `internal/merge/strategies.go` 는 이 카드에서 **불변**이다.
- **D4** `init --force` 는 유지하되, `--agent codex|both` + already-initialized 조합에서 `moai update --add-codex` 안내를 출력한다. 재초기화 차단·플래그 제거는 하지 않는다.
- **D5** 템플릿 CLAUDE.md `^## [0-9]` 헤딩 수 18 불변. universal 이동(§6 품질 게이트 포인터, §7 일부)은 **본문 이동 + 제목·한 줄 포인터 잔류** 형태다.
- **D6 — 5 섹션 제목 + 바이트 예산 고정** (AC-UAC-009 grep 앵커):

| 섹션 제목 (템플릿 AGENTS.md, 영어) | 내용 (a-e) | 예산 |
|---|---|---|
| `## 8. Codex Web Console` | (a) 웹 콘솔 안내 | ≤ 1.0 KB |
| `## 9. Hook Event Coverage` | (b) codex 가 claude 의 11개 중 일부 훅 이벤트만 발화한다 — 어떤 이벤트가 빠지는지 델타 설명 | ≤ 1.5 KB |
| `## 10. Configuration Map` | (c) `.moai/config/sections/` 포인터 | ≤ 1.0 KB |
| `## 11. moai CLI Verbs` | (d) CLI 동사 표 | ≤ 1.5 KB |
| `## 12. Status Line Tokens` | (e) statusline 토큰 (구현 포인터 우선 — 토큰 목록을 복제해 낡게 하지 않는다) | ≤ 1.0 KB |

universal 이동분(§6 포인터) ≤ 0.5 KB. **총 증분 목표 ≤ 6.5 KB → M2 후 목표 ≤ 22,000 B** (하드 상한 24,576 B, 이 트리 시작점 15,415 B). 초과 시 본문 확장 금지, 포인터화로 해소.

## §E 자가 검증

- E1 — AC Binary PASS/FAIL 매트릭스 (acceptance.md §D, 각 행에 명령+실측 출력+HEAD SHA)
- E2 — `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- E3 — `go test ./internal/cli/... ./internal/codexwiring/... ./internal/config/...` (변경 영향 패키지 — 전체 스위트 로컬 금지, CI 몫)
- E4 — 경계 grep: `rg -n "add-codex" internal/cli/update.go` ≥ 1 (M1 후), `rg -n "MOAI:LEARNED-WORKFLOW" internal/template/templates/AGENTS.md` = 0 (M2 후)
- E5 — `golangci-lint run --timeout=2m` (신규 이슈 0 — 기준선 대비)
- E8 — RED 실패 출력: M1 착수 전 `./bin/moai update --add-codex` exit 1 재현본, M2 착수 전 5 섹션 앵커 0건, M3 착수 전 안내 문구 0건 (acceptance.md §D.5 장부 참조)

## §F 마일스톤

### 순서 구속 — 권고가 아니라 의존성

- **M3 는 M1 에 의존한다**: REQ-UAC-013 의 안내 문구가 `moai update --add-codex` 를 **이름으로** 참조하므로, 동사 없이 M3 를 착수하면 존재하지 않는 경로를 안내하게 된다.
- M2 는 M1 과 기술적으로 독립(템플릿 내용)이나 같은 브랜치에서 직렬 수행한다. 고정 순서 M1 → M2 → M3.

### M1 — update --add-codex 동사 + 배선 (REQ-UAC-001~008)

1. `updateCmd.Flags().Bool("add-codex", false, ...)` 등록 + help 문안(REQ-UAC-008)
2. ungated 래퍼(`refreshCodexWiringBestEffortAt` 형제, 예: `addCodexWiring`) — `codexwiring.Wire` 호출 (D1)
3. runUpdate 배선 — `update.go:507` 옆, `--dry-run` 프리뷰 경로와 `--check` 상호배타(기존 `validateUpdateVersionConflicts` 확장) 처리 (REQ-UAC-006/007)
4. 테스트: 스크랫치 프로젝트 3-산출물 생성 / 플래그 부재 시 무생성(회귀) / .mcp.json 바이트 동일 / 멱등 재실행 / 사용자 config.toml 보존 / dry-run 무기록 / check 조합 거절

**종료 조건**: AC-UAC-001~008 green.

### M2 — AGENTS.md 5 섹션 + universal 이동 (REQ-UAC-009~012)

1. §D6 표의 5 섹션을 `internal/template/templates/AGENTS.md` 에 저작 (예산 준수)
2. 템플릿 CLAUDE.md §6 품질 게이트 포인터의 universal 본문을 AGENTS.md 로 이동, 제목 18개 불변 (D5)
3. `make build` 재임베드 (템플릿 우선 규율 — `make agents-emit` 불필요, 에이전트 정의 미변경)
4. 바이트 가드 확인: `go test ./internal/config -run TestCodexContractByteCeiling` + `wc -c` ≤ 24,576

**종료 조건**: AC-UAC-009~012 green.

### M3 — init --force codex-add 철폐 + 고정 테스트 (REQ-UAC-013~014)

1. init already-initialized + `--agent codex|both` 경로에 안내 출력 추가 (`wireCodexUnlessClaude` 앞 or already-initialized 가드 인접 — run-phase 에서 호출부 확정)
2. §C #4 재인벤토리 결과의 고정 테스트 갱신 (같은 변경 내)
3. 영향 패키지 테스트 green

**종료 조건**: AC-UAC-013~014 green.

## §G AC별 mutant 노트

| AC | 잡아야 할 mutant | 판별 수단 |
|---|---|---|
| AC-UAC-001 | no-op 플래그(M-1) | AC-002 산출물 검사가 보조 |
| AC-UAC-002 | RefreshWiring 호출(M-2) — 게이트가 통과시키지 못함 | claude-only 스크랫치에서 wiring 파일 부재 확인 |
| AC-UAC-003 | 무조건 Wire 호출로 뒤집은 mutant | 플래그 부재 경로 회귀 |
| AC-UAC-004 | .mcp.json 재프로비저닝(M-3) | sha256 전후 비교 |
| AC-UAC-005 | 항상 재기록하는 mutant | sidecar sha 불변 |
| AC-UAC-006 | config.toml 덮어쓰기 mutant | 사용자 항목 grep |
| AC-UAC-007/008 | dry-run 이 쓰는 mutant / check 조합 통과 mutant | 파일 부재 + exit 코드 |
| AC-UAC-009 | 제목만 있고 내용 없는 껍데기 | 각 섹션 본문 앵커 2차 grep (§D.1) |
| AC-UAC-002 + §D.4-1 | 오류를 삼키는 래퍼 mutant(감사 F1) — `ErrValidationRefused` 를 경고로 바꾸고 exit 0 으로 끝나는 구현 | 신규 CLI 전파 테스트 `TestUpdateAddCodex_ValidationRefusalFailsLoud` 가 판정자 (§D.4 #1) |
| AC-UAC-010 | 상한 초과 본문 | fail-closed 가드 테스트가 판정자 |
| AC-UAC-011 | pure wrapper(M-5) / 마커 미러(M-4) | `^## [0-9]` 카운트 + 마커 grep |
| AC-UAC-013 | 안내 없이 통과하는 mutant | 안내 문구 grep |
| AC-UAC-014 | 테스트를 삭제로 "갱신"하는 mutant | 테스트 함수 수 감소 금지(같은 변경 내) |

## §H 잔여 위험

1. **루트 계약 쌍 드리프트**: M2 가 템플릿만 고치면 로컬 루트 `AGENTS.md`/`CLAUDE.md` 는 구판으로 남는다(오늘 두 쌍은 바이트 크기까지 동일). 동일성 강제 가드 존재 여부가 §C 확인 사항 — 가드가 있으면 M2 가 루트 쌍도 같은 변경에서 맞춘다. 없으면 다음 `moai update` 흡수를 기다리며, 그 기간의 루트·템플릿 내용 차이는 문서화된 상태다.
2. **statusline 토큰 설명의 낡음**: §D6 가 복제 금지·포인터 우선을 구속으로 박았다 — 내부 구현 변경 시 템플릿이 낡는 결함의 재발 차단.
3. **안내 문구의 시끄러움**: REQ-UAC-013 안내는 --agent codex|both 재초기화를 명시적으로 요청한 운영자에게도 출력된다. 의도된 행위(redirect-not-block, D4)이나 출력이 과하면 sync-phase 피드백으로 문구 조정한다.
4. **바이트 예비 소진 이후**: 5 섹션 저작이 목표(≤6.5KB)를 넘으면 섹션을 늘리지 말고 §D6 의 포인터화 규율을 적용한다. 상한 도달 시 다음 카드는 AGENTS.md 다이어트다.

## §I 교차 참조

- SPEC-CODEX-WIRING-001 — 배선 패키지(`codexwiring.Wire`/`RefreshWiring`), REQ-CW-001/003/008/009 계승·보존
- SPEC-UPDATE-MIRROR-HEAL-001 — `update.go:507` 인접 호출 자리의 선례(게이트와 무관하게 update 흐름에 붙는 보조 정비의 위치 관례)
- SPEC-UPDATE-VERSION-FLAG-001 — `validateUpdateVersionConflicts` 상호배타 매트릭스 (M1 이 확장)
- SPEC-INIT-HARNESS-PROMPT-001 — `resolveAgentWiringWithWizard` (M3 가 참조하는 --agent 해석점)
- 카드 t589 / 형제 t585 · 감사 `.moai/reports/init-tui-audit-20260909.md` §7 + 6차 탐사(CLAUDE.md 191행 census, 2026-09-09)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier M 판정
