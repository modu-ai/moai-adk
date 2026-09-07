---
id: SPEC-STATE-ANCHOR-VALIDATE-001
title: "상태 앵커 후보 검증 — Resolve의 ProjectDir/OriginalCwd 절대경로·존재 검증 (t510 F1 후속 경화)"
version: "0.1.0"
status: draft
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/stateanchor"
lifecycle: spec-anchored
tags: "state-anchor, validation, payload-trust, t510-f1, t537, hardening"
tier: S
era: V3R6
related_specs: [SPEC-STATE-ANCHOR-001]
depends_on: [SPEC-STATE-ANCHOR-001]
---

# SPEC: 상태 앵커 후보 검증 — Resolve의 ProjectDir/OriginalCwd 절대경로·존재 검증

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-08 | manager-spec | 최초 작성 — 카드 t537 plan-phase. 기원: t510 sync-audit F1(선택 경화 제안 — `.moai/reports/t510/sync-audit.md:120`, develop 워크트리에서 전문 재확인). 부모 SPEC `SPEC-STATE-ANCHOR-001`의 체인 1·2단(무조건 반환)을 검증-통과-폴스루로 바꾸는 **요구사항 변경**을 본 SPEC이 명시적으로 소유한다 |

카드: **t537** (Tier S · 선택 경화). 측정 기준 트리: `.claude/worktrees/t537`, 브랜치 `WT-resolve-validate` @ `52f863f36`. 모든 코드 좌표와 RED 근거는 이 트리에서 본 레인이 1차로 재판독한 것이다.

## 1. 문제 — 측정된 형태

부모 SPEC(`SPEC-STATE-ANCHOR-001`)이 세운 단일 상태-앵커 시접의 체인 3단은 검증되지만, 1·2단은 검증되지 않는다.

```go
// internal/stateanchor/stateanchor.go:64-76 (트리 52f863f36, 직접 판독)
func Resolve(s Session) string {
	if s.ProjectDir != "" {
		return s.ProjectDir        // ← 절대경로·존재 무검증 반환
	}
	if s.OriginalCwd != "" {
		return s.OriginalCwd       // ← 동일
	}
	dir := s.CurrentDir
	if dir == "" {
		dir = s.CWD
	}
	return FromDirectory(dir)      // ← 3단만 git 해석으로 검증됨
}
```

`ProjectDir`(stdin `workspace.project_dir`)과 `OriginalCwd`(stdin `worktree.original_cwd`)는 **빈 문자열이 아니기만 하면 그대로 앵커가 된다.** 절대경로 확인 없음, 존재 확인 없음, git 멤버십 확인 없음. 상대경로 페이로드는 상태줄 프로세스의 cwd 기준 상대경로로 해석되고, 존재하지 않는 절대경로는 그대로 앵커가 된다.

### 1.1 피해 기전 — 검증 없는 앵커는 임의 경로를 부활시킨다

앵커는 읽기 전용 값이 아니다. 상태줄의 상태 쓰기들은 쓰기 전에 앵커 유도 디렉터를 `MkdirAll`로 **생성**한다(본 레인 직접 판독):

| 소비자 | 좌표 | 착지 위치 |
|---|---|---|
| B1 세션 텔레메트리 | `internal/statusline/context_usage.go:197` | `<앵커>/.moai/state/context-usage/<sid>.json` |
| B2 landed | `internal/statusline/landed.go:256` | `<앵커>/.moai/state/landed/…` |
| B2b github counts | `internal/statusline/github.go:365` | `<앵커>/.moai/state/github/…` |

(부록 — **기원 인용의 정정**: t537 배차문의 피해 근거 목록은 `model_cache.go:55`를 앵커-유도 MkdirAll로 함께 나열했으나, 본 레인의 직접 판독에서 `WriteModelCache`는 `homeDir` 인자를 받아 `~/.moai/state/last-model.txt`에 쓰는 **홈-앵커** 경로다(`model_cache.go:47-60`, 호출 `metrics.go:55`). 앵커-유도 MkdirAll은 위 3곳이며, 정정으로 피해 서술은 좁아지지 않는다 — 상대경로·부재절대경로 페이로드의 부활 기전은 3곳에서 그대로 성립한다.)

따라서 stale(삭제된 프로젝트) 또는 거짓(적대적/오염된 stdin) `project_dir` 하나가 **존재하지 않는 임의 경로를 렌더당 새 디렉터 트리로 부활**시킨다. GH #1694로 접수된 stray-`.moai` 실패 형태와 동일한 물성이, 페이로드 경로로 재도입돼 있는 것이다. 시접의 문서 주석(`stateanchor.go:1-23`)이 말하는 앵커의 의미 — "`.moai/state/`가 읽히고 쓰이는 프로젝트 루트" — 는 무검증 반환과 모순된다.

### 1.2 기원 판정문 (전문 재확인)

develop 워크트리 `.moai/reports/t510/sync-audit.md:120`에서 전문을 직접 확인했다:

> **F1** [Low] [optional] `internal/stateanchor/stateanchor.go:64` — `Resolve`가 `ProjectDir`/`OriginalCwd`를 절대경로·존재 여부 무검증으로 반환한다. 적대적 stdin 페이로드가 상태 쓰기/읽기 앵커를 프로세스가 쓸 수 있는 임의 경로로 보낼 수 있다. 다만 신뢰 모델은 수리 전과 동일하고(같은 필드가 구 리졸버에도 흘렀음), 권한 경계 통과·유출 능력 증가는 없다. Required fix: 없음(승인된 설계). 선택 경화 — 앵커 후보에 `filepath.IsAbs` 요구 또는 디렉터 존재 검증을 후속 SPEC 후보로 기록.

본 카드가 그 "후속 SPEC 후보"다. F1이 Low/optional로 분류된 이유 — 신뢰 모델 불변·권한 경계 없음 — 는 유지되며, 본 SPEC은 그 등급 그대로 **경화(hardening)** 다. Required fix가 아니므로 폴스루 설계의 우아한 저하(REQ-SA-003 계승)를 해치는 강한 실패 동작은 요구하지 않는다.

### 1.3 소비자 지도 — 한 지점의 검증이 전부를 덮는다

본 레인 직접 판독: `Resolve`의 호출자는 정확히 하나다 — `internal/statusline/state_anchor.go:22-35`의 `resolveStateAnchor` (`grep -rn "stateanchor.Resolve" internal/ --include="*.go"` 비테스트 1매치). 이 어댑터 하나가 B1(`builder.go:181`)·B2(`backlog.go:31` board root → landed/github)·B3(`builder.go:292` goal 읽기)에 모두 흐른다. 따라서 **`Resolve` 내부 한 지점의 검증이 전 소비자를 덮는다** — 호출자 파일 변경이 필요 없다. `FromDirectory`(`internal/cli/deps.go:178`, B4 CLI 사슬)는 git 해석으로 존재성이 구조적으로 검증되는 별도 진입이며, 본 SPEC이 건드리지 않는다(REQ-SAV-005).

## 2. 원인 — 체인 3단의 검증 기준을 1·2단이 나눠 갖지 않았다

부모 SPEC의 수리는 "cwd를 앵커 후보에서 제거하고 git으로 답한다"였다. 그 과정에서 1·2단(stdin 필드)은 "런타임이 보고한 값이므로 신뢰한다"로 두었다 — F1이 지적하듯 승인된 설계였다. 그러나 3단(`FromDirectory`)은 같은 질문에 대해 이미 **존재성과 git 멤버십을 모두 검증**하는 형태로 착지했다. 세 단이 같은 계약("앵커는 `.moai/state/`가 읽히고 쓰이는 실재하는 프로젝트 루트")에 대해 다른 검증 수준을 갖는 것이 결함의 형태다. 수리는 새 검증층의 발명이 아니라, 3단이 이미 갖고 있는 계약의 **최소 검증 부분(절대경로 + 존재)을 1·2단에도 적용**하는 것이다. git 멤버십까지 1·2단으로 끌어올리지 않는 이유는 §5 옵션 (c)에서 기각한다.

## 3. 대표 mutant — 이 SPEC의 AC가 어떻게 만족될 수 있는가

"유효한 후보는 반환한다"만 요구하면 다음 구현들이 전부 통과하면서 결함을 남긴다.

1. **IsAbs-only 뮤턴트** (F1의 선택지 1) — 상대경로만 막고 **삭제된 절대경로는 그대로 통과**한다. 부활 기전(§1.1)이 살아남는다. AC-SAV-002가 잡는다.
2. **exists-only 뮤턴트** — `os.Stat`을 상대경로에도 적용하면 **상태줄 프로세스의 cwd 기준**으로 해석되므로, 상대 페이로드가 프로세스 cwd 아래 존재하는 디렉터를 가리키면 통과한다. 절대경로 요구가 없으면 존재 검증은 잘못된 기준점을 검증한다. AC-SAV-001이 잡는다.
3. **git-membership + root-equality 뮤턴트** — 기능 AC는 전부 통과하지만 (a) 후보 경로마다 git spawn을 렌더에 추가한다(~300ms 주기 경로의 비용 회귀), (b) **워크트리 세션의 정상 `project_dir`을 거절한다** — 워크트리 루트는 git common-dir 부모와 다르기 때문. AC-SAV-003의 워크트리 하위 케이스와 AC-SAV-007이 잡는다.
4. **hard-"" 뮤턴트** — 첫 후보가 검증 실패하면 즉시 ""를 반환한다. `project_dir`이 stale된 워크트리 세션이 `original_cwd`(그리고 git 워크업)로 구원받을 기회를 잃고 **상태 쓰기 전체가 꺼진다** — REQ-SA-003의 우아한 저하가 오히려 파괴된다. AC-SAV-002의 "폴스루" 단언이 잡는다.

## 4. 요구사항 (GEARS)

> **요구사항 변경 선언** — 부모 SPEC plan §D1이 체인 순서·중간 삽입을 "요구사항 변경(REQ-SA-002)"으로 고정했다. 본 SPEC은 그 요구사항 변경을 **명시적으로 소유**한다: 체인에 새 단을 삽입하거나 순서를 바꾸지 않고, 체인 1·2단의 **반환 조건**을 "비어 있지 않음"에서 "비어 있지 않음 **그리고** 검증 통과"로 강화한다. 순서(`project_dir` → `original_cwd` → git 워크업)와 3단 의미론은 불변이다. 부모 SPEC의 REQ-SA-002는 본 강화를 수용하도록 해석된다(본 SPEC이 그 변경의 SPEC이다).

### 경화 — 검증 술어

- **REQ-SAV-001** (Ubiquitous) — The state-anchor resolver shall validate every session-derived anchor candidate before returning it: a candidate (`ProjectDir`, `OriginalCwd`) shall be returned only when it is an absolute path (`filepath.IsAbs`) **and** names an existing directory (`os.Stat` succeeding with directory mode).
- **REQ-SAV-002** (Event-driven) — **When** a session-derived anchor candidate fails validation, the resolver shall fall through to the next step of the fixed precedence chain (`ProjectDir` → `OriginalCwd` → git walk-up) — it shall not return the failed candidate, and it shall not return "" before the remaining chain steps have been tried.
- **REQ-SAV-003** (Ubiquitous) — An empty candidate value shall mean "absent": it skips validation trivially and falls through, unchanged from current behavior.

### 보존 — 경화가 파괴해서는 안 되는 것

- **REQ-SAV-004** (Unwanted) — The resolver shall not add git membership, root-equality, or any other git-spawn validation to session-derived candidates; the candidate path shall remain git-free (the render's git budget is already spent on chain step 3 — §5 option (c) rejection).
- **REQ-SAV-005** (Unwanted) — The hardening shall not change `FromDirectory` behavior (git-validated by construction) and shall not modify any caller file: production changes are confined to `internal/stateanchor/stateanchor.go` and `internal/stateanchor/stateanchor_test.go`; the statusline adapter (`state_anchor.go`), its B1/B2/B2b/B3 consumers, and the display derivation (`extractProjectDirectory`, parent REQ-SA-004) stay untouched.
- **REQ-SAV-006** (Unwanted) — The validation shall not rewrite or normalize the candidate value (no `EvalSymlinks` rewrite of the returned anchor); the anchor's value provenance and the `.moai/state/` path scheme stay unchanged (parent D13 — the anchor VALUE may change for the better; the schema and path scheme may not).
- **REQ-SAV-007** (Event-driven) — **When** every session-derived candidate fails validation and the git walk-up also fails, the resolver shall return "" and callers shall skip the state write silently — the parent REQ-SA-003 "no project, no state" semantics preserved.

## 5. 설계 결정 — 검증 술어 선택 (lattice, 권고, 기각)

### 술어 격자

| 옵션 | 내용 | 잡는 것 | 못 잡는 것 / 비용 | 판정 |
|---|---|---|---|---|
| (a) | `filepath.IsAbs`만 | 상대경로 페이로드 | **삭제된 절대경로 통과** — 부활 기전 생존. I/O 0 | 불충분 — 단독 채택 기각 |
| (b) | 디렉터 존재(`os.Stat`)만 | stale/삭제 경로 | 상대경로가 **프로세스 cwd 기준**으로 해석돼 통과 가능 | 불충분 — 단독 채택 기각 |
| (a)+(b) | **절대경로 AND 존재 디렉터** | 상대경로 + stale/삭제 절대경로 — §1.1 부활 기전 양쪽 차단 | "존재하지만 무관한 디렉터"는 통과. 비용: 후보 존재 시 `os.Stat` 1회 — 3단이 이미 수행하는 git spawn보다 저렴 | **채택 (권고)** |
| (c) | git 멤버십 + root-equality | "무관한 디렉터"까지 | 후보마다 git spawn — 렌더 비용 회귀. **워크트리 세션의 정상 `project_dir`(워크트리 루트 ≠ common-dir 부모)을 거절** — 정상 경로를 깨는 과잉 검증. 단순성 사다리 위반(기존 3단이 git 순도를 이미 담당) | 기각 (REQ-SAV-004로 고정) |

**권고: (a)+(b)**. 근거: (a) 단독은 F1이 제안한 최소선이지만 부활 기전의 절반만 닫고, (b) 단독은 기준점을 틀린 곳에 둔다. 둘의 합은 `os.Stat` 1회라는 비용으로 두 피해층을 모두 닫고, 3단과의 검증 수준 격차(§2)를 git 미포함 선까지 메운다. (c)까지는 3단의 역할 침범 + 정상 워크트리 페이로드 거절이라 대가가 기능을 깬다.

### 실패 동작 — 폴스루 (권고) vs 즉시 ""

검증 실패 시 **다음 체인 단계로 폴스루**한다(REQ-SAV-002). 즉시 ""는 구현이 더 단순하지만, `project_dir`이 stale된 워크트리 세션(실사용 환경에서 흔한 상태 — 워크트리 삭제 후 `original_cwd`는 살아있는 형태)이 앵커를 아예 잃고 B1/B2/B3 상태 쓰기·읽기가 전부 꺼진다. 폴스루는 부모 REQ-SA-003의 "끝까지 못 찾으면 skip" 우아한 저하와 정확히 같은 결에 서 있고, 거짓 페이로드도 결국 git 워크업이 git 순도를 담보한다. F1이 Low/optional임을 감안하면 가용성을 잃는 hard-""는 과잉 반응이다.

### 기호 결정 3건

1. **빈 후보는 검증을 건너뛴다** (REQ-SAV-003) — 빈 문자열은 "부재"이지 "실패"가 아니므로 폴스루만 하면 된다. 기존 동작 불변.
2. **symlink 정규화는 하지 않는다** (REQ-SAV-006) — `os.Stat`은 symlink를 따르므로 "실재하는 디렉터" 판정 자체는 dual spelling(`/var/folders` vs `/private/var/folders` — 기존 테스트가 정규화하던 지점, `stateanchor_test.go:63-69`)에 관계없이 올바르다. 반환값은 **후보를 그대로** 반환하고 `EvalSymlinks` 재작성을 하지 않는다 — 값의 정규화는 앵커 값 provenance를 바꾸는 별도 변경이고, 본 SPEC의 결함(무검증)과 무관하다. 기존 테스트의 dual-spelling 정규화는 git 워크업(3단) 비교용으로 그대로 남는다.
3. **TOCTOU 인정** — 존재 판정 시점과 쓰기 시점 사이에 디렉터가 사라질 수 있다. 본 SPEC은 페이로드 신뢰 창을 닫는 것이 목적이지 레이스 불변 달성이 아니므로, 이 잔여 창은 Gaps(§7)에 기록한다.

## 6. 범위 밖 (Non-goals)

### Out of Scope — `internal/kanban` 결함 (카드 t536 전용 소관)

- `internal/kanban/state_dir.go:129-132`(`BacklogPathForRoot`가 `root` 인자를 `resolveStateDir`에 재투입 — §7 관측)와 `internal/kanban/todo_root.go:112-125`(home 미해석 폴백)의 수리는 본 SPEC이 하지 않는다. t536이 전용 소관이며, 본 SPEC의 접촉은 읽기 전용 코드 판독뿐이다(REQ-SAV-005의 범위 선언).

### Out of Scope — `FromDirectory`와 호출자 표면

- `FromDirectory`의 git 해석 의미론과 `internal/statusline/**` 전체(어댑터 `state_anchor.go` 포함), 표시 유도 `extractProjectDirectory`는 불변이다. 생산 변경은 `internal/stateanchor/` 2파일로 한정된다.

### Out of Scope — 앵커 값의 정규화 (EvalSymlinks 재작성)

- 검증에 통과한 후보를 symlink-해석 표기로 재작성하는 것은 값 provenance 변경이며 §5 기호 결정 2대로 하지 않는다.

### Out of Scope — git 멤버십 검증의 1·2단 확장

- §5 옵션 (c) 기각에 따라 후보 경로에 git spawn을 추가하는 어떤 구현도 본 SPEC을 위반한다(REQ-SAV-004). "무관한 디렉터" 통과는 인정된 잔여 위험이다.

### Out of Scope — 신뢰 모델 변경

- F1이 인정하듯 stdin 페이로드의 신뢰 모델(동일 사용자 파일시스템, 권한 경계 없음)은 본 경화로 바뀌지 않는다. 본 SPEC은 임의 경로 **부활**(MkdirAll 디렉터 생성)을 막는 것이지 페이로드 신뢰 문제를 해결하지 않는다.

## 7. 관측 (Out of Scope — t536 소관, 코드 판독 근거)

부모 SPEC의 리졸버 패밀리에는 "root를 받은 함수가 그 root를 다른 의미로 재해석하는" 결함 형태가 반복 관측된다. 본 레인이 코드 판독으로 재확인한 2건 — 둘 다 **t536 전용 소관**이며, 트리거도 다르다(페이로드 신뢰가 아니라 root-vs-statedir 혼동):

1. `BacklogPathForRoot`(`internal/kanban/state_dir.go:129-132`)는 `root` 인자를 그대로 `resolveStateDir`에 넣고, `resolveStateDir` → `StateDirForRoot`(`state_dir.go:46-48`, `filepath.Join(root, ".moai", "state", <todo>)`)가 그 위에 `.moai/state/todo`를 다시 접는다. 이미 해석된 상태 디렉터를 넘기면 `<root>/.moai/state/todo/.moai/state/todo/backlog.json` 이중 중첩 경로가 나온다 — 기전은 위 세 좌표의 직접 판독으로 기계적으로 확정된다.
2. `homeTodoQueueRoot`의 home-미해석 폴백(`internal/kanban/todo_root.go:112-125`)도 같은 형태로 `resolveStateDir(base, false)`에 디렉터 의미를 맡긴다(배포 코드).

주제적 연결: 두 결함 모양 모두 **"넘겨받은 것의 디렉터-의미를 검증하라"** 는 같은 처방이 답한다 — 이 관측이 본 카드를 패밀리 수리로 존재시키는 근거다. 단, kanban 쪽 수리는 카드별로 나뉜다 — **관측 1(BacklogPathForRoot 이중 접기, 루트→상태-디렉터 재유도)은 t549**, **관측 2(homeTodoQueueRoot 홈-미해석 폴백)는 t536**의 몫이다(t536 트리 배치가 `WT-home-fallback`인 것으로 뒷받침; §8의 t549 상호 참조와 정렬).

## 8. 미검증 항목 (Gaps)

- **t536 판정서 착지 시 상호 인용 갱신** — plan 시점 기준 lane-2 t536 판정서를 표준 증거 경로에서 찾지 못했다(확인: `.claude/worktrees/t536/.moai/reports/`, develop 트리, primary 트리). 본 SPEC §7의 kanban 관측은 판정서 인용 없이 **본 레인의 코드 판독(file:line)** 만으로 기술했다 — 조작하지 않았다. t536 판정서가 착지하면 상호 인용으로 갱신한다.
- **기원 배차문 근거 목록의 `model_cache.go:55`는 앵커-유도가 아니었다** — 본 레인 직접 판독으로 홈-앵커(`~/.moai/state/last-model.txt`)임이 확인됐고 §1.1 부록에 정정을 기록했다. 피해 서술은 3곳에서 성립하므로 결론 불변이나 논거를 정정했다(정정이 간극을 좁힌 방향).
- **TOCTOU 잔여 창** — §5 기호 결정 3. 존재 판정과 상태 쓰기 사이의 디렉터 소멸 레이스는 닫지 않는다(목적이 페이로드 신뢰 창이므로). 필요해지면 별도 SPEC.
- **`os.Stat` 비용은 판독 추정이다** — 후보 존재 시 1회 `os.Stat`은 3단의 git spawn 대비 저렴하다고 판단했지만, plan-phase에서 벤치마크로 측정하지는 않았다. Tier S 범위에서 비용 AC를 두지 않는다; 렌더 경로 비용이 문제가 되면 M1 이후 관측으로 확대한다.
- **"존재하지만 무관한 디렉터" 통과** — §5 (a)+(b) 채택의 인정된 잔여 위험. git 멤버십까지의 확장은 옵션 (c) 기각(§5)으로 본 SPEC 범위 밖이다.
- **t549 상호 참조 (run 진입 시 추가)** — §7 관측의 결함 패밀리(「root를 받은 함수가 그것을 무엇으로 취급하는가」)의 트리거-다른 형제 카드 t549가 발행됐다(kanban 루트→상태-디렉터 재유도, `internal/kanban` 소관 — 본 SPEC과 트리거·파일이 다르다). 수리는 t549의 몫이며 본 SPEC의 접촉은 이 상호 참조뿐이다.
