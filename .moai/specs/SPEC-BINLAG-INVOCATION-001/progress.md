# SPEC-BINLAG-INVOCATION-001 — 진행 기록

카드 t366 · 브랜치 `WT-lint-binary-lag` · 증거 경로 `.moai/reports/t366/`

## §E.1 Plan-phase Audit-Ready Signal

- 상태: plan-phase 산출물 4종 작성 완료(`spec.md` / `plan.md` / `acceptance.md` / `progress.md`).
- Tier: **S**. 근거는 `spec.md` §6의 어느 안을 골라도 예상 변경이 5파일 미만·300 LOC 미만이라는 점.
- 요구 8 / 수락 8. Tier S 상한(각 8) 정확히 충족.
- 작성 트리: `d7010f86a`. plan-phase의 모든 RED-now 셀이 그 트리에서 관측됐다.
- 미결 3건(`plan.md` §F M0의 `[NEEDS CLARIFICATION]`)은 **해소됐다** — 아래 §F 참조.
- plan-phase가 실행하지 않은 것: 양방향 재현(설계만), 명령 단위 census, 호출당 비교 비용 측정.

## §F Phase 4 Mode Selection

- 입력: Tier S · 범위 5파일 미만 · 도메인 1(문서/규칙) · 언어 markdown 전용 · 병렬 이득 없음.
- 평가: `direct` 선택 / `serial` 미선택(위임할 코드 작업 없음) / `fanout` 미선택(단일 도메인) / `sweep` 미선택(기계적 대량 변환 아님).
- **Decision: direct** — 코드 변경 0, 변경 파일 6개(규칙 1 + 미러 1 + SPEC 3 + 보고 1), 판단의 실체는 이 세션이 직접 실행한 측정이다. 위임하면 그 측정을 재도출해야 하므로 비용만 늘고 이득이 없다.
- 편향 주의: 이 세션이 관측을 수행하고 그 수락 기준을 스스로 PASS로 적었다. 독립 판정은 sync-phase 감사와 리드의 최종 판정이 맡는다(§E.3 참조).

## §E.2 Run-phase Evidence

측정 트리 **`968c9caf8`**(워크트리 `.claude/worktrees/t366`). 증거 원본: `.moai/reports/t366/evidence/`,
서사: `.moai/reports/t366/run-observation.md`.

### 게이트 결정 접수 — Option B

Implementation Kickoff Approval(2026-08-31)이 **Option B(절차만, 코드 무변경)** 를 채택했다.
A(루트 seam)·C(좁은 발화)는 조건부 후속으로 이월(조건: B를 세운 뒤에도 같은 오염이 계속 관측되면).
그 결과 요구·수락의 성격이 바뀌었고, 재서술을 `spec.md` §2.0 / HISTORY 0.2.0 / `acceptance.md` §D에 남겼다.

### 양방향 관측 (실행됨)

| | 다른 쪽 — 설치본 `343399d2f` | 같은 쪽 — 트리 빌드 `968c9caf8` |
|---|---|---|
| `moai spec lint` stdout 마지막 줄 | `0 error(s), 64 warning(s)` | `0 error(s), 177 warning(s)` |
| stdout 줄 수 | 68 | 181 |
| 종료 상태 | 0 | 0 |
| stderr 바이트 | 0 | 0 |

- `cmp -s installed.stdout tree.stdout` → exit 1 (다르다)
- 실행되지 않은 규칙: `MovingRefUnpinned` (t342, 착지 `84b3b7949`)
- `git merge-base --is-ancestor 343399d2f 84b3b7949` → exit 0 (설치본 빌드 이후 착지)
- 판정 자체는 존재·정확: `doctor` 가 설치본에 `warn ... binary: 343399d2f, HEAD: 968c9caf8`, 트리 빌드에 `ok ... matches source HEAD (968c9caf8)`

**결론** — 두 호출은 stderr(둘 다 0바이트)와 종료 상태(둘 다 0)로 **구별되지 않는다**. 113개 warning
분량의 규칙이 조용히 실행되지 않았고 출력은 그 사실을 말하지 않았다. 이것이 이 카드의 결함이다.

### 변경 파일 (코드 0줄)

| 파일 | 변경 |
|---|---|
| `.claude/rules/moai/core/verification-claim-integrity.md` | §2.2 Tool-provenance attribution 신설, Version 1.2.0 → 1.3.0, 로컬 provenance 한 줄 |
| `internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md` | 같은 §2.2 미러(중립 — SPEC ID·SHA·경로 없음), Version 1.3.0 |
| `.moai/specs/SPEC-BINLAG-INVOCATION-001/spec.md` | frontmatter 0.2.0 / in-progress, HISTORY 0.2.0, §2.0 요구 재성격, §6 결정 배너, §8 R-3 갱신 + R-4/R-5 추가 |
| `.moai/specs/SPEC-BINLAG-INVOCATION-001/acceptance.md` | §D 매트릭스 재작성(B 아래 분류 + 합산 제외), AC-001/003 GREEN 셀, AC-002/004 재서술, AC-005~008 자명 충족 주석, §D.1/§D.2 갱신 |
| `.moai/reports/t366/run-observation.md` | 신규 — 양방향 관측 서사 |
| `.moai/reports/t366/evidence/` | 신규 — 원시 stdout/stderr/규칙목록 6파일 |

### 공용 설치본 무변경 (REQ-BLI-008)

`~/go/bin/moai` 는 절차 전후로 `v3.1.2 / 343399d2f / built 2026-08-27T14:07:38Z`, 크기 `68955858`,
mtime `Aug 27 23:09` 로 동일하다. 재현 바이너리는 세션 전용 임시 디렉터리에 빌드했다.

## §E.3 Run-phase Audit-Ready Signal

- 실질 판정 대상 4건(AC-BLI-001 · 002 · 003 · 004) 전부 PASS. 자명 충족 4건(005-008)은 합산 제외 —
  발화가 없어 자동으로 참이 된 것이며 성취가 아니다(`spec.md` §8 R-4).
- **코드 변경 0이므로 테스트·lint 판정 대상이 없다. 이는 통과가 아니라 미해당이다.**
- **자기 판정 고지**: 이 세션이 관측을 수행하고 같은 세션에서 수락 기준을 PASS로 적었다. 독립 판정은
  sync-phase 감사와 리드의 최종 판정에 맡긴다.
- 미검증으로 남기는 것: `acceptance.md` §D.2 — 특히 **B안 전제 미검증(R-3)**, 적용 불가 관용 실측
  부재, `spec lint` 외 명령의 갈림 폭 미측정(R-5).

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: `completed`
- sync_complete_at: 2026-08-31
- sync_commit_sha: 57a6df488
- changelog_entry_position: `[Unreleased]` → `### Added`, first entry (card t366, `SPEC-BINLAG-INVOCATION-001`)
- frontmatter_status_transitions.spec_md: `in-progress → completed` (merged into this sync commit; `updated:` refreshed to 2026-08-31)
- b12_self_test_a (pre-emission grep): `grep -c 'BINLAG' CHANGELOG.md` → `0` before emission (checked; single entry now present)
- b12_self_test_b (AC count match): `grep -oE 'AC-BLI-[0-9]+' acceptance.md \| sort -u \| wc -l` → **8**; CHANGELOG cites 4 judged-PASS + 4 vacuously-satisfied-excluded, matching `acceptance.md` §D exactly
- b12_self_test_c (file path verification): all paths cited in the CHANGELOG entry verified to exist via `ls` before commit (`spec.md`, both `verification-claim-integrity.md` copies, `root.go`, `acceptance.md`, `run-observation.md`, `evidence/`, `progress.md`, `internal/binlag`)
- canary_compliance_check: not applicable — this SPEC defines a doctrine clause (§2.2 tool-provenance attribution), not a forward-looking policy with its own canary test
- Scope discipline: this sync commit modifies ONLY `CHANGELOG.md`, this file's §E.4, and `spec.md` frontmatter (`status:` + `updated:`). No body content in `spec.md`/`plan.md`/`acceptance.md` was touched.
- Not pushed by this session — the integration window is the lead's, per `.claude/rules/local/gitflow-lane-protocol.md`.
