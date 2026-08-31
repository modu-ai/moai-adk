---
id: SPEC-CODEX-LAUNCH-VERB-001
title: "맨몸 moai codex 를 기동으로 — 기본 동사 역전, CODEX_HOME 명시 전달, -w 경로, 세 론처 규약 대조"
version: "0.1.0"
status: draft
created: 2026-08-31
updated: 2026-08-31
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "codex, launcher, launch-verb, codex-home, worktree, cross-launcher"
depends_on: [SPEC-CODEX-LAUNCHER-001]
related_specs: [SPEC-CODEX-INIT-001, SPEC-CODEX-WIRING-001, SPEC-WORKTREE-ENTRY-STRATEGY-001]
---

# SPEC-CODEX-LAUNCH-VERB-001 — 맨몸 `moai codex` 를 기동으로

## HISTORY

| 버전 | 날짜 | 변경 |
|---|---|---|
| 0.1.0 | 2026-08-31 | 최초 작성 (plan-phase, 카드 t391). 운영자 판정 2026-08-31 — SPEC-CODEX-LAUNCHER-001 `plan.md` §B 의 (b) 판정을 뒤집는다. 범위 4항(기본 동사 역전 / CODEX_HOME 명시 전달 / `-w` 경로 / 세 론처 대조 시험)으로 한정하고, settings 정리·프로필 적용 대칭성은 REQ-CL-013 과 충돌하므로 요구에서 제외. `-k`/`-f` 개방은 범위 밖 열린 질문 |

## §A. 승계 관계 — 무엇을 대체하는가

이 SPEC 은 **SPEC-CODEX-LAUNCHER-001 의 REQ-CL-002 를 대체한다** (supersede). 대체되는 원문:

> **REQ-CL-002** — The bare `moai codex` command shall print the readiness readout and exec nothing; launching shall require an explicit verb — `cli` … `status` shall be accepted as an explicit alias of the bare readout form.

SPEC-CODEX-LAUNCHER-001 은 `status: completed` 이므로 그 본문은 이 카드에서 고치지 않는다. 완결된 SPEC 에 HISTORY 항목이나 승계 포인터(`superseded_by` / `partially_superseded_by`)를 다는 일은 **sync-phase 소관**이며 run-phase 편집 대상이 아니다.

대체되지 **않는** 것을 명시한다 — 같은 SPEC 의 나머지 13개 REQ 는 전부 유효하다. 특히:

- **REQ-CL-013**(비변경 규율: 런처는 읽고 exec 할 뿐 쓰지 않는다)는 **그대로 유효**하며, 이 SPEC 의 REQ-CLV-006 이 그것을 재확인한다.
- **REQ-CL-011**(`app` 은 `codex app` 에 위임), **REQ-CL-003**(`--spawn` 계약), **REQ-CL-012**(바이너리 부재 시 단일 진단, launch 동사에만 적용) 는 무변경이다.

## §B. 측정 전제 (Verified baseline)

> 근거: `.moai/reports/t391/verdict.md` — 트리 `e79272713`(= 진입 시점 `origin/develop`)에서 수행된 5-절 판독. 아래 표의 "출처" 열이 판독본인 항목은 **이 SPEC 이 다시 잰 것이 아니라** 그 문서가 잰 것이다.

| # | 관측 | 출처 |
|---|---|---|
| B1 | 맨몸 비기동은 결함이 아니라 기록된 판정이다 — `plan.md` §B 가 (a)기동/(b)리드아웃 중 (b)를 택했다 (리드 판정 2026-08-24) | verdict C1 |
| B2 | (b)의 기각 근거("실수로 세션이 codex 에 넘어감")는 현행 구현에서 무게가 다르다 — cc/glm 은 `syscall.Exec` 프로세스 교체지만 codex 는 0.8.0 개정에서 `os/exec` 자식 + 종료코드 전파로 갔다 | verdict C2 |
| B3 | `codexInitOfferGate` 는 두 launch 동사가 지나는 **단일 통과 지점**이다 (정의 1건 `codex_init.go:152`, 호출 1건 `codex_launcher.go:256`) | verdict C4 |
| B4 | `codexDirectLaunch` 는 자식에게 `Env` 를 지정하지 않는다 — `grep 'c.Env\|Env ='` 출력 0행 | verdict C5 |
| B5 | 세 론처 교차 시험 파일은 **있다**(`codex_launcher_test.go`). 대조 항목이 그룹 소속·GroupID·spawn 진단 바이트뿐이고, "인자 없는 호출이 기동으로 이어지는가" 셀만 없다 | verdict C6 |
| B6 | codex 에는 `-w` 자체가 없다 (`codex_launcher.go` 의 `-w`/worktree 매치 0건). 대응 표면은 `cc.go:220-228` 의 `resolveWorktreeL2Path` + `normalizeWorktreeFlag` | verdict §6 (3) |
| B7 | **codex 바이너리에 `cli` 서브커맨드도 별칭도 없다.** `codex --help` 의 Commands 목록에 `cli` 가 없고, 같은 출력을 `grep -i cli` 하면 제목행·산문행 2행만 잡힌다. help 는 "If no subcommand is specified, options will be forwarded to the interactive CLI" 와 `codex [OPTIONS] [PROMPT]` 를 함께 명시한다 — 즉 `codex cli` 는 **프롬프트 `cli`** 로 해석된다 | 이 SPEC (2026-08-31 실측, `/Users/goos/.local/bin/codex`) |

| B8 | codex 최상위 help 에 **워크트리 플래그가 없다** — 같은 출력을 `worktree` 와 `-w,` 로 grep 하면 0행이다. 서브커맨드별 플래그와 숨은 플래그는 훑지 않았으므로 이 부재 주장은 최상위 표면에 한정된다 | 이 SPEC (2026-08-31 실측) |

B7·B8 은 이 SPEC 이 직접 잰 두 항목이며 판독본에는 없다 — 판독본 G5 가 "codex 바이너리를 실행하지 않았다"로 남긴 자리다. 이것이 REQ-CLV-004 의 근거다.

현행 구현은 `req.Args = append([]string{verb}, tail...)` 로 moai 쪽 동사 토큰을 자식 argv 첫 자리에 그대로 실어 보낸다(`codex_launcher.go:246`; 시험 `codex_launcher_test.go:465` 가 `want := append([]string{"cli"}, tail...)` 로 이를 고정하고 있다). `app` 은 실제 codex 서브커맨드라 성립하지만 `cli` 는 성립하지 않는다. 기본값을 기동으로 옮기는 일은 이 토큰 결함을 맨몸 경로로 승격시키므로, 함께 닫지 않으면 카드가 목표한 "맨몸이 기동한다"가 "맨몸이 프롬프트 `cli` 로 기동한다"가 된다.

## §C. 범위 밖 (out of scope)

이 카드가 **짓지 않는 것**을 먼저 못 박는다. 아래 항목은 전부 out of scope 이며 요구·수용 어느 쪽에도 나타나지 않는다.

### Out of Scope — settings 정리·프로필 적용 대칭성

- `moai cc` / `moai glm` 이 기동 전에 수행하는 settings 정리와 프로필 적용을 codex 에 옮기는 일. **REQ-CL-013 과 정면으로 충돌한다** — 런처는 읽고 exec 할 뿐 쓰지 않는다. cc/glm 이 그 단계를 갖는 것은 둘이 **같은 `claude` 바이너리**를 띄우기 때문이며 codex 에는 대응물이 없다 (운영자 판정 2026-08-31).
- `.claude/settings.local.json` · Claude 프로필 상태 · `CODEX_HOME` 하위 파일에 대한 모든 쓰기.

### Out of Scope — Kanban / Factory 진입 토큰

- `-k` / `-f` 를 codex 에 여는 문제. 판독본이 읽은 어떤 것도 이 질문에 답하지 않는다. **열린 질문이며 이 카드 범위 밖**이다 (운영자 판정 2026-08-31).

### Out of Scope — 완결 SPEC 본문 편집

- SPEC-CODEX-LAUNCHER-001 의 REQ 본문·HISTORY·frontmatter 편집. 승계 포인터 부착은 sync-phase 소관이다.
- REQ-CL-002 를 제외한 나머지 13 REQ 의 재기술.

### Out of Scope — auth 분류·배선 판정 재설계

- auth 2단 사다리(REQ-CL-008/009)와 배선 완전성 판정(REQ-CL-006)은 무변경이다. 이 카드는 **어느 동사가 무엇을 하는가**만 바꾼다.
- 리드아웃 6행의 내용과 형식.

### Out of Scope — 대화형 tty 왕복 단언

- tty 왕복은 CI 에서 관측 불가하므로 SPEC-CODEX-LAUNCHER-001 0.8.0 이 명시적 Gap 으로 남겼다. 이 SPEC 은 **같은 Gap 을 물려받으며** 그것을 닫으려 하지 않는다.

## §D. 요구사항 (GEARS)

### D.1 동사 표면 — 기본값 역전

- **REQ-CLV-001** — The bare `moai codex` command shall launch the Codex CLI, in the same manner as the explicit `cli` verb. This requirement supersedes REQ-CL-002 of SPEC-CODEX-LAUNCHER-001, whose bare-readout-and-launch-nothing clause it replaces in full.
- **REQ-CLV-002** — The `status` token shall remain an explicit alias that renders the readiness readout and launches nothing, so the readout stays reachable without launching; the `cli` verb shall remain accepted as an explicit synonym of the bare launching form.
- **REQ-CLV-003** — The verb-routing table shall remain the single closed set that decides routing: an unrouted token shall be rejected with the usage diagnostic, and no default branch shall route an unknown token to a launch.
- **REQ-CLV-004** — The argv handed to the codex child shall carry only tokens the codex binary itself accepts; the moai-side launching verb shall not be forwarded to the child as a positional argument, because the codex binary exposes no `cli` subcommand and reads such a token as a prompt. The `app` verb, which names a real codex subcommand, shall continue to be forwarded.

### D.2 환경 전달

- **REQ-CLV-005** — The launcher shall pass the resolved `CODEX_HOME` to the child process explicitly as an environment entry, rather than relying on ambient inheritance, and shall leave the remainder of the parent environment intact.
- **REQ-CLV-006** — The launcher shall not mutate `.claude/settings.local.json`, Claude profile state, or any file under `CODEX_HOME` on any verb, the bare form included — REQ-CL-013 of SPEC-CODEX-LAUNCHER-001 remains in force, and the environment entry REQ-CLV-005 requires shall be process-local, never a written file.

### D.3 워크트리 경로

- **REQ-CLV-007** — Where the operator passes `-w` (or `--worktree`) to a launching form, the launch shall occur with the named worktree root as its working directory, in place of the project root the launch would otherwise resolve.
- **REQ-CLV-008** — The `-w` value shall be resolved by the same rules `moai cc` applies — absolute paths validated against the accepted worktree prefixes, short names normalized — and, when an absolute path falls outside those prefixes, the system shall fail with that resolution's diagnostic and launch nothing.

### D.4 게이트·플랫폼 상속

- **REQ-CLV-009** — The SPEC-CODEX-INIT-001 offer gate shall remain the single choke point every launching form passes through, the bare form included; while the project wiring is incomplete and the session is non-interactive, the system shall issue no prompt and launch nothing.
- **REQ-CLV-010** — The codex launch path shall remain an `os/exec` child with verbatim exit-code propagation, and shall carry zero OS build tags and zero `syscall` imports, preserving the cross-platform property established by SPEC-CODEX-LAUNCHER-001 HISTORY 0.8.0.

### D.5 대조 시험과 배포 표면

- **REQ-CLV-011** — A single test location shall assert the bare-invocation convention of all three launchers (`cc`, `glm`, `codex`) together — whether an argument-free invocation leads to a launch — extending the existing cross-launcher comparison rather than forking a second one.
- **REQ-CLV-012** — Help text, examples, and any template-side documentation shall reflect the reversed default, stay language-neutral, and stay free of internal identifiers, satisfying the template neutrality guard.

## §E. 비기능 요구

- 리드아웃 경로는 fail-open 을 유지한다 — `status` 는 어느 프로브가 실패해도 나머지 행을 계속 보고한다.
- 기본값 역전은 **진단 가능성을 줄이지 않는다**: 바이너리가 없어도 `status` 는 여전히 성공해야 하며, 그 성질은 REQ-CL-012 가 이미 launch 동사에만 진단을 걸어 둔 결과로 성립한다.
- 경로 해석은 `os.UserHomeDir` 기반을 유지하며 macOS 편향 경로를 하드코딩하지 않는다.

## §F. 성공 판정

수용 기준은 `acceptance.md` 가 가진다. 이 SPEC 은 그 AC 표가 REQ-CLV-001..012 를 전수 덮고, 각 AC 가 이분 판정 가능할 때 완결로 본다.
