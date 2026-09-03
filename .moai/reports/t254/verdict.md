# t254 판정서 — GFM 표 셀 파이프 이스케이프 하자 (가짜-0 재측정)

- 판정: **FIXED** — 카드 전제(결함 존재)는 확인됐으나 **좌표는 스테일**이었다. 카드가 지목한 `research.md:32 / spec.md:55`는 실재하지 않는다(`SPEC-CODEX-WIRING-001`에 research.md 없음 — 4파일 구성; spec.md:55는 §A.2 표의 파이프 무관 행). lane-5 전례(lane-13 지시 경고)대로 재측정한 결과, 실 결함 표면은 progress 기록행 2건이다.

## 실제 결함 표면 (이름 나열)

- `.moai/specs/SPEC-CODEX-SIDECAR-GUARD-001/progress.md` — AC-CSG-003 행: `grep -cE '"\.codex"\|"\.codex/"' …`
- `.moai/specs/SPEC-CODEX-WIRING-001/progress.md` — AC-CW-003 행: `grep -cE 'enabled_tools\|disabled_tools' …`

**메커니즘**: `grep -E`에서 `\|`는 이스케이프된 **리터럴 파이프** 매치(대안이 아님). 표 셀에 기록된 이 형태는 파일 내용과 무관하게 항상 0을 내므로, 기록된 0이 참인지 렌더(실행) 기준으로 반증 불가 — 카드가 지적한 결함 그 자체.

## 렌더 형태 재측정 (develop `2660bcd09` + 바이너리 v3.1.2-1308-g65196a5a7, 2026-09-02)

| 검사 | 소스형(기록된 `\|`) | 렌더형(베어 `\|`) | 기록값 | 판정 |
|---|---|---|---|---|
| AC-CSG-003 | 0 (exit 1) | **0** (exit 1) | 0 | 값 참 — 형태만 결함 |
| AC-CW-003 열거 부재 | 0 (exit 1) | **0** (exit 1) | 0 | 값 참 — 형태만 결함 |
| AC-CW-003 default_tools | — | **1** (exit 0) | 1 | 일치 |
| AC-CW-003 TOML 파싱 | — | `TOML_OK` (exit 0) | TOML_OK | 일치 |
| AC-CW-001 | — | `moai init --help \| grep -c -- '--agent'` → **1** | 1 | 일치 |

**양성 대조**: 같은 `internal/cli/init_agent_flag_test.go`에 무인용 `.codex` 참조 9건(exit 0) — 인용 형태(`".codex"`)만 진짜 부재임을 확인. 세 기록값 모두 **참값** — AC·판정의 정정은 불필요, **명령 형태만** 결함.

## 변경 (이름 나열)

- `SPEC-CODEX-SIDECAR-GUARD-001/progress.md` AC-CSG-003 행 — 다중 `-e` 형태로 교체(파이프 0개: 소스==렌더==실행 동일) + 재측정 각주.
- `SPEC-CODEX-WIRING-001/progress.md` AC-CW-003 행 — 동일 교체 + 각주.
- 미수정: `AC-CW-001` 행의 `moai init --help \| grep …` — 셸 파이프의 셀 이스케이프는 **정당**한 형태(렌더==의도 명령, 값 재확인 1). acceptance.md의 정본 형태(베어 파이프, 표 밖)도 그대로.

## 5-섹션

**Claim**: 결함 행 2건이 실행 가능·반증 가능한 형태(다중 `-e`)로 교체됐고, 세 기록값(0/1/0)은 렌더 재측정으로 참임이 확인됐다.

**Evidence**: 위 표의 명령+출력 전부 (verbatim); `grep -c '\.codex' internal/cli/init_agent_flag_test.go` → `9` exit=0.

**Baseline-attribution**: `WT-pipe-render-check` 워크트리, develop `2660bcd09`, 2026-09-02 본 런. init 재측정 바이너리 `~/go/bin/moai` = v3.1.2-1308-g65196a5a7 (2026-09-02 빌드 — 설치본, 트리 HEAD와 다른 커밋임을 명시).

**Gaps**: /tmp 재측정 프로젝트는 오늘 빌드 바이너리 기준 — 기록 당시(v3.1.3 시절) 바이너리와의 동일성은 미검증(다만 생성기 출력이 달라졌다면 default_tools=1·TOML_OK도 함께 어긋났을 것 — 3측정 전부 일치).

**Residual-risk**: 같은 `-cE 'a\|b'` 형태가 다른 문서에 있는지 전수 스윕은 `git grep -nE "^\|.*\\\\\|" refs/heads/develop -- .moai/specs/`로 CODEX 2개 SPEC에서 3행(실결함 2 + CW-001 정당형 1)만 확인했다 — 타 도메인 문서의 동형 결함은 본 카드 스윕 범위 밖.
