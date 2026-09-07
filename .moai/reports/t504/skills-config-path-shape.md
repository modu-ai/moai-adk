# t504 판정서 — codex `[[skills.config]]` path 값 모양 실측

- **SPEC**: SPEC-CODEX-SKILLCONFIG-SHAPE-001 (Tier S, measurement-only)
- **측정 일시**: 2026-09-07, 런타임 스탬프 `codex-cli 0.153.4` (`codex --version`, 실험 시점 직접 측정)
- **바이너리**: `/Users/goos/.local/bin/codex`
- **방법**: `CODEX_HOME` 격리 (t91 §9 선례) — 실사용자 `~/.codex` **기록 없음**
- **랩 아카이브**: `lab/` (본문 캡처 전체), 재현 스크립트 `harness.sh`(1차)·`harness2.sh`(2차)
- **원 실험 트리**: `/tmp/t504-hOuB8eKd` (세션 종료 후 OS가 지움 — 영구 사본은 `lab/`)

## 판정 요약 (한 문단)

**RQ1: DEAD — 엔트리는 로딩을 "만들지" 않는다. 그러나 스위치로는 산다(기발견 후 D-시리즈 정정).** codex-cli 0.153.4는 `[[skills.config]]` 배열 엔트리를 config.toml에서 **파싱한다** — 그러나 그 `path` 값은 **스킬 로딩을 생성하지 않는다**(V5: 실존 파일+`enabled=true`, 중립 경로 → 마커 0; 대조군 전부 정상: IV 격리 양성, N 디렉터 컨벤션 양성, V6 부재경로 음성). 부수 발견 2건: (1) **`enabled` 필드는 파서가 필수로 요구한다** — 이 필드가 없는 엔트리 하나면 codex 시작이 하드 에러로 죽는다(rc=1, 출력 0). (2) **죽은 경로에 대해 codex는 완전히 침묵한다**(rc=0, stderr 0).

**D-시리즈 정정(리드 지시 추간 셀)**: 살아있는 스킬(`$CODEX_HOME/skills/` 컨벤션으로 로드되는)에 대해 **파일 모양 엔트리+`enabled=false`는 실제로 로딩을 끈다**(D1f 마커 0 vs 통제 D2f 마커 1). 즉 이 키의 실측된 기능은 **"노출 생성"이 아니라 "기존 스킬의 게이트"**다 — 단 파일 모양 한정(디렉터 모양 엔트리의 enabled=false는 무효: D1d=1). 요약: **생성은 못 하고 차단은 한다(파일 모양).**

## 셀 매트릭스 (라운드 1)

| 셀 | 구성 | out(bytes) | err(bytes) | rc | T504MARKER/IV마커 |
|----|------|-----------|-----------|----|----|
| IV | 격리 유효성: 스크래치 홈 AGENTS.md 마커, 스킬 설정 없음 | 24957 | 0 | 0 | **IV마커=1 ✓** |
| N | 디텍터 기준: `$CODEX_HOME/skills/t504probe/`, skills.config 없음 | 24910 | 0 | 0 | **1 ✓** |
| V2 | 부재 경로, `enabled` 없음 | 0 | 151 | **1** | 0 (에러로 인한 것 — 무효) |
| V1 | 실존 파일, `enabled` 없음 | 0 | 151 | **1** | 0 (에러로 인한 것 — 무효) |
| V3 | 실존 디렉터, `enabled` 없음 | 0 | 151 | **1** | 0 (에러로 인한 것 — 무효) |
| V4 | 실존 파일 + `enabled = false` | 24752 | 0 | 0 | 0 |

1차 3셀(V1/V2/V3)의 rc=1 원인 — 세 셀 모두 동일 (verbatim, `lab/cells/V1.err`):

```
Error: /private/tmp/t504-hOuB8eKd/home-V1/config.toml:2:1: missing field `enabled`

Caused by:
    missing field `enabled`
    in `skills.config`
```

→ **발견 F1**: `enabled`는 옵션이 아니라 **필수 필드**다. 1차 설계의 V1/V2/V3는 의도한 변수(경로 유무)를 못 봤고, V1/V3/V4의 판정은 채택 불가(AC-CSS-001-03 게이트가 정확히 작동) — 2차 라운드로 재설계했다.

## 셀 매트릭스 (라운드 2 — `enabled = true` 정합 셀)

| 셀 | 구성 | out(bytes) | err(bytes) | rc | T504MARKER |
|----|------|-----------|-----------|----|----|
| V5 | **실존 파일 + enabled=true** (진짜 RQ1 셀) | 24753 | 0 | 0 | **0** |
| V6 | **부재 경로 + enabled=true** (깨끗한 무효 대조) | 24753 | 0 | 0 | **0** |
| V7 | **실존 디렉터 + enabled=true** | 24753 | 0 | 0 | **0** |

세 셀의 출력 크기가 동일(24753B)하고 stderr 전부 0 — 스키마는 통과하되 스킬 목록에 어떤 차이도 없다는 뜻이다. `t504probe`라는 이름이 출력 어디에도 나오지 않는다(마커 grep 0).

## D-시리즈 (리드 지시 추간 셀 — 살아있는 스킬을 끄는 실험)

질문: `$CODEX_HOME/skills/t504probe/SKILL.md`로 **실제 로드되는** 스킬에 대해, 그 스킬을 가리키는 `[[skills.config]]` 엔트리의 `enabled = false`가 로딩을 끄는가. 4셀 — 스킬 물리 위치는 4셀 모두 동일(컨벤션 경로), 엔트리의 path 모양과 enabled만 다름.

| 셀 | 엔트리 path 모양 | enabled | 마커 | out(bytes) | rc |
|----|----------------|---------|------|-----------|----|
| D2f | 파일 (`.../t504probe/SKILL.md`) | true | **1** ✓ | 24914 | 0 |
| D1f | 파일 (`.../t504probe/SKILL.md`) | false | **0** — 꺼짐 | 24754 | 0 |
| D2d | 디렉터 (`.../t504probe`) | true | **1** ✓ | 24914 | 0 |
| D1d | 디렉터 (`.../t504probe`) | false | **1** — 안 꺼짐 | 24914 | 0 |

stderr 전부 0, 노터치 해시 재확인 동일(`c45741c1…8638e01a`).

**통역 (행렬이 뒷받침하는 것까지만):**
- **D1 답 (파일 모양): 비활성화가 실제로 작동한다.** D2f(양성 통제)가 로드되는 같은 환경에서 D1f만 스킬 목록에서 사라졌다 — out 크기도 정확히 t504probe 엔트리 1개분(160B)만큼 작다.
- **D2 답:** 두 true 통제 모두 로드 — 엔트리의 존재 자체는 로딩을 깨지 않는다. D1f의 침묵은 `enabled=false`에 귀속된다.
- **D3 답 (모양 비대칭): 파일 모양만 스위치가 묶인다.** 디렉터 모양 엔트리는 true/false 모두 관성(D1d=D2d=1) — V7(디렉터 모양 생성 무력)과 정합. 매칭은 SKILL.md **파일 경로** 기준으로 관측됐다(D2f 마커 문맥: `t504probe: … (file: r0/…)` — 로드 경로는 r0 컨벤션 루트, 게이트는 파일 엔트리).
- 종합: 이 키의 실측된 기능 = **기존 스킬에 대한 파일별 게이트 레코드**. 노출 생성 무력(V5/V7), 파일 모양 게이트 유효(D1f/D2f), 디렉터 모양 완전 관성(V7/D1d/D2d), `enabled` 스키마 필수(F1).

## 채택성 게이트 (AC-CSS-001-01/02/03)

- IV 양성(격리 유효) ✓ — `CODEX_HOME` 격리는 0.153.4에서 유효. 폴백(사용자 홈 백업/복원) 불필요.
- N 양성(디텍터 유효) ✓ — 같은 바이너리에서 `$CODEX_HOME/skills/` 컨벤션은 노출됨.
- V6 음성 ✓ — 부재 경로+유효 스키마가 노출 없음을 깨끗하게 재현.
- 두 통제군이 같은 방향으로 실패한 경우 없음 → VOID 아님. **행렬 채택 가능.**

## 연구질문 판정

### RQ1 — `[[skills.config]]`는 로드되는가 → **DEAD (로딩 생성 축 — 게이트 기능은 D-시리즈 참조)**

근거: V5 (실존 파일+`enabled=true`, 중립 경로) 마커 0. 경로가 `$CODEX_HOME/skills/`도, cwd `.agents/skills/`도 아니므로 이 키 외에 노출을 설명할 통로가 없는데 노출이 없었다. 반면 같은 바이너리에서 N 셀(디렉터 컨벤션)은 노출됐다 — 디텍터가 스킬 노출을 관측할 수 있는 상태에서만 V5가 음성이었다는 것이 통제군이 보증한다.

단, **"완전히 무시"는 아니다**: 파싱은 일어나고 스키마 검증이 강제된다(F1: `enabled` 필수; 부재 시 codex 전체 rc=1). 즉 이 표면은 **"스키마는 검사하지만 로딩에는 안 쓰는"** 상태다.

### RQ2 — path 값 모양 → **파일·디렉터 모두 로드(생성) 안 됨 / `enabled`는 필수, 게이트 기능에서는 값이 유의미**

(본 절은 "노출 생성" 축의 측정이다. `enabled` 값의 실제 효과는 아래 D-시리즈가 재정의: 살아있는 스킬을 게이트할 때만 값이 작동한다.)

- 파일(`.../SKILL.md`): V5 마커 0. moai 엔트리가 쓰는 바로 그 모양이다.
- 디렉터: V7 마커 0.
- `enabled = false` (V4) vs `true` (V5): 둘 다 파싱 통과, 둘 다 노출 0 — 관측된 표면에서 값의 차이 없음.
- 부재 경로 + 유효 스키마: **완전 침묵** (V6 rc=0, stderr 0바이트). t451 docstring의 "neither prunes it nor complains" 중 complains 축 **확인**.
- 단, docstring이 모르는 반대 축의 결함: **`enabled` 누락 시 codex가 침묵하지 않고 죽는다**(F1). t451의 tri-state 파서(`skills.go`)는 `enabled`를 옵션으로 읽고 "Codex의 기본값은 관측되지 않았다"고만 말하는데, 실제는 "Codex가 필수로 요구하며 없으면 시작 실패"다 — doctor가 잡아야 할 **새 발견 등급**이다.

### RQ3 — 실환경 (읽기 전용, 격리 없음) → **49건 전부 enabled 보유·완전 침묵·노출 0**

- 실환경 49건 중 `enabled` 보유: **49/49** (`grep -A3 '^\[\[skills.config\]\]' | grep -c 'enabled\s*='` → 49). 그래서 이 머신의 codex가 F1 하드에러 없이 돌아간다.
- R 셀: rc=0, stderr 0바이트, 죽은 경로 49건에 대한 어떤 불평도 없음.
- R 출력의 유일한 `moai-` 일치: `moai-cowork` **플러그인 마켓플레이스** 루트 r34–r51 (`~/.codex/plugins/cache/moai-cowork/moai-*`) — moai 브랜드 cowork 플러그인들의 스킬이 **플러그인 루트 메커니즘**으로 로드되는 것이지 `skills.config` 엔트리가 아니었다. RQ3의 기대 부재(49건에서 오는 moai 스킬 콘텐츠 0)는 유지.
- 참고: roots 테이블 r0=`~/.codex/skills`, r1=`~/.agents/skills` — 실제 로딩 통로는 디렉터 컨벤션+플러그인 루트다.

## 노터치 증명 (AC-CSS-001-06)

```
before: c45741c10c3e844ae2fe768787338cfdc659dc41b2e2b75db83f5ebf8638e01a  /Users/goos/.codex/config.toml
after:  c45741c10c3e844ae2fe768787338cfdc659dc41b2e2b75db83f5ebf8638e01a  /Users/goos/.codex/config.toml
```

동일. R 셀 전 사본(backup-config.toml)은 만들어졌으나 복원이 일어날 일이 없었다(무효 확인 no-op) — 사본은 실값 API 키를 포함해 아카이브에서 제외됐다(아래 Gaps 보안 노트; 세션 한정 /tmp에만 존재). 실험 전체에서 실사용자 `~/.codex`에 기록한 셀은 없다(격리 셀 6+3건은 전부 스크래치 홈, R만 읽기).

## downstream 함의

- **t502 (skills.config 작업)**: **진행 확정(D-시리즈 판정)** — 비활성화 수단이 실재한다. 제약 2건이 설계에 들어가야 한다: (1) **`enabled`를 반드시 명시 발행**(F1 — 누락 시 그 사용자 codex 전면 장애), (2) **엔트리 path는 파일 모양(`.../SKILL.md`)으로** — 디렉터 모양은 게이트가 안 묶인다(D1d). 단, **노출 생성(프로비저닝)에는 이 키를 쓰지 않는다**(V5/V7) — 생성 통로는 `$CODEX_HOME/skills/` 디렉터 컨벤션(N 재확인)·`.agents/skills/`(t91)·플러그인 루트. M1 정본+심링크 전략과 정합.
- **t506 (유령 49건 정리)**: 삭제가 스킬 로딩에 미치는 영향 = **0** — 49건 전부 경로 부재라 로드될 것도 없고, D-시리즈 기준으로도 죽은 경로 엔트리는 게이트로서 아무것도 묶지 못한다(게이트는 실존 파일과의 경로 일치에 묶임). 삭제는 순수한 서류 정리다. 단 삭제 시 `enabled` 필드 보유 엔트리만 골라 안전하게 지우면 되고, 반대로 **`enabled`가 없는 엔트리를 새로 쓰는 사고가 생기면 그 사용자의 codex 전체가 깨진다**(F1) — 정리 도구는 이를 유지해야 한다.
- **t451 후속 (doctor)**: 두 가지 시사. (1) `enabled` 누락 엔트리 = codex 전면 장애 — doctor의 새 ERROR 등급 발견 후보(현 파서는 tri-state로 옵션 취급). (2) 죽은 경로 발견의 심각도 서술은 "remove the stale entries" 대신 "로딩에 영향 없는 서류 정리" 쪽이 실측과 더 정확하다. **본 카드에서 코드 수정 없음**(NFC-1) — 후속 카드 소관.

---

## Claim

codex-cli 0.153.4에서 `[[skills.config]]` 엔트리는 파싱·스키마 검증되며(`enabled` 필수 — 누락 시 codex 전면 하드 에러), `path` 값으로 스킬 로딩을 **생성하지는 않는다**(파일·디렉터 모두). 그러나 **파일 모양 엔트리는 그 SKILL.md에 정확히 맞물리는 게이트**로 작동해, `enabled = false`는 컨벤션으로 로드되던 살아있는 스킬의 노출을 실제로 끈다(디렉터 모양 엔트리는 게이트 무효). 실환경의 죽은 경로 49건은 codex가 침묵하며 노출에 영향이 없다.

## Evidence

위 표 전체 + verbatim 에러 본문 + 마커 grep 카운트(스캔 바이트 수 동반) + 노터치 해시 쌍(D-시리즈 후 재확인 포함). 모든 원본: `lab/cells/*.{out,err,rc}` (D-시리즈 포함), `lab/summary.txt`, `lab/summary2.txt`, `lab/summary3.txt`, `lab/version.txt`. 재현: `bash harness.sh` → `bash harness2.sh <EXP_ROOT>` → `bash harness3.sh <EXP_ROOT>`.

## Baseline-attribution

- 측정 대상 바이너리: `codex-cli 0.153.4` — `lab/version.txt` (실험 시작 시 `codex --version` 직접 측정, 세션 반입 값 아님)
- 측정 트리: worktree `.claude/worktrees/t504` @ `WT-codex-skillpath-shape`, base `ace1c5440` (origin/develop)
- 판정 명령: 각 셀 `codex debug prompt-input "hi"` (모델 호출 0회), 카운트는 디스크 파일 대상 `grep -c` (파이프 판정 없음)

## Gaps

- **데스크톱 앱 미관측**: 본 측정은 CLI(`debug prompt-input`) 경로다. Codex 데스크톱 앱이 `skills.config`를 자체 UI(스킬 레지스트리) 용도로 쓸 가능성은 배제하지 못했다 — 키의 "작성자"는 앱 측일 가능성이 높고(실환경 49건이 모두 `enabled`를 갖는 규칙적 모양), 앱에서의 의미는 미측정.
- **다른 codex 버전 미측정**: 0.153.4 단일 버전. t91(0.147대)·0.152.1 프로브와 비교해도 이 키의 로드 여부를 본 적은 이번이 처음이다.
- **`enabled` 이외 필드 미탐색**: `name` 등 다른 키가 로드를 활성화할 가능성은 시도하지 않았다(스키마 에러가 추가 필수 필드를 알려주지 않는한 없었음 — path+enabled로 rc=0 통과).
- **t504 큐 원문 미독**: `backlog.db` 직독 시도에서 테이블 스키마 불일치로 원문을 읽지 못했다(`.moai/backlog.db`, `~/.moai/backlog.db` 모두 예상 테이블 부재). 리드 배차 본문이 설계의 근거 문서였다.
- R 셀 cwd는 실험 proj 디렉터(중립) — 실제 프로젝트 컨텍스트에서의 노출 차이는 이번 범위 밖.
- **아카이브 제외 2건 (보안·위생)**: (1) 실사용자 config 사본 `backup-config.toml`은 실값 `Z_AI_API_KEY`(1건)를 포함해 **아카이브에서 제외**했다(비밀값 금지 규율 — 세션 한정 /tmp 사본으로만 존재, 실험 후 OS 폐기). 노터치 대조는 `lab/hash-before.txt`·`lab/hash-after.txt`가 보증한다. (2) codex가 스크래치 홈에 자동 생성한 시스템 스킬 번들(`skills/.system/` — skill-creator 등 codex 배포물)과 세션 잔여물은 증거가 아니라서 제외했다 — 피처 config·SKILL.md·셀 캡처·요약만 보존된다.

## Residual-risk

- 키가 "로드 무시"라는 판정을 codex가 향후 버전에서 뒤집을 수 있다 — 본 판정은 0.153.4 한정이며 t502가 이 키에 무언가를 쓰기로 결정하면 그 시점 버전에서 재측정해야 한다.
- `enabled` 필수 발견(F1)은 에러 메시지 기반의 관측이다 — codex가 조용히 기본값을 채우는 다른 실행 경로(예: 앱)가 있다면 F1은 CLI 한정 관측이다.
- 플러그인 루트(r34–r51 moai-cowork)의 스킬 내용은 본 측정의 대상이 아니었고, 그 표면의 동작은 별도 관측 없이는 알 수 없다.
