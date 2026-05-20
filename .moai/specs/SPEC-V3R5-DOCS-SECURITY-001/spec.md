---
id: SPEC-V3R5-DOCS-SECURITY-001
title: "docs-site Wave A — Security Foundation (CWE-732/214/345 사용자 가시 변화 반영)"
version: "0.1.0"
status: draft
created: 2026-05-20
updated: 2026-05-20
author: manager-spec
priority: P0
phase: "v2.20.0-rc1"
module: "docs-site/content/{ko,en,ja,zh}"
lifecycle: spec-anchored
tags: "docs, security, i18n, cwe-732, cwe-214, cwe-345, wave-a"
tier: M
---

# SPEC-V3R5-DOCS-SECURITY-001 — docs-site Wave A: Security Foundation

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-20 | manager-spec | Tier M 초안 작성. SPEC-V3R5-SECURITY-CRIT-001 (P0 보안 결함 3건 정정) 의 사용자 가시 변화를 docs-site 4-locale (ko/en/ja/zh)에 반영. 신규 `advanced/security-notes.md` + 기존 `advanced/settings-json.md` / `getting-started/update.md` / `multi-llm/cg-mode.md` 갱신. Tier 판단: 4 페이지 × 4 locale = **16 markdown files** 영향 → research doc §8.2 원안 Tier S 권고를 Tier M으로 상향 (Tier S 임계 <5 files 초과). |

## 1. Background

### 1.1 출처

본 SPEC은 `.moai/research/docs-site-v2.14-to-HEAD-update-plan-2026-05-20.md` (v2.14.0 → HEAD `03a2552a2` 범위, 313 commits / 50+ merged SPECs, 5명 병렬 reviewer 분석) §4 + §6.2 + §8.2 Wave A 첫 SPEC.

v2.14.0 → main HEAD 범위에서 SECURITY-CRIT-001 (PR #1032 머지 `03a2552a2`) 이 사용자 가시 보안 동작을 3건 변경했으나, `adk.mo.ai.kr` docs-site 어디에도 반영되지 않았다. 본 SPEC은 그 격차를 해소한다.

### 1.2 다루는 사용자 가시 변화

| 출처 | CWE | 사용자 가시 변화 |
|------|-----|------------------|
| SECURITY-CRIT-001 M1 (`b48bd86cb`) | CWE-732 / CWE-552 | `.claude/settings.local.json` 파일 mode `0o644 → 0o600` (소유자 전용 read/write) |
| SECURITY-CRIT-001 M2 (`10776c4b8`) | CWE-214 | `moai cg` 의 tmux env 주입이 argv 대신 source-file 경유 (`ps auxe`, `/proc/<pid>/cmdline` 에서 GLM token 비가시화) |
| SECURITY-CRIT-001 M3 (`ee1335282`) | CWE-345 | `moai update` checksum 검증 mandatory (`checksums.txt` 다운로드 3회 retry 후 실패 시 update 거부, `ErrChecksumUnavailable` sentinel) |

### 1.3 본 SPEC의 가치 (Why)

- **Enterprise 사용자 신뢰** — v2.20.0-rc1 release-blocker 보안 결함 3건의 정정 사실 + 권장 점검 절차를 docs-site 4-locale 공식 안내로 명문화.
- **OWASP / CWE 매핑 공개** — 감사·컴플라이언스 사용자가 변화 영향을 평가 가능.
- **회귀 방지 가이드** — 사용자가 자신의 `settings.local.json` permission, tmux env 노출 여부, update flow 가 새 보호를 받는지 자체 점검할 수 있는 체크리스트 제공.
- **다국어 동시성** — `docs-site-i18n-rules.md` §17.3 (ko canonical + en/ja/zh 동일 PR) 준수로 글로벌 사용자 동시 도달.

## 2. Goals

- **G1** — 신규 페이지 `advanced/security-notes.md` 를 ko/en/ja/zh 4-locale 모두 작성한다. 페이지는 6 섹션 구조 (Why / CWE-732 / CWE-214 / CWE-345 / 점검 체크리스트 / CHANGELOG 링크) 를 따른다.
- **G2** — 기존 `advanced/settings-json.md` 에 `settings.local.json` permission `0o600` 권고 및 자체 점검 명령을 4-locale 모두 추가한다.
- **G3** — 기존 `getting-started/update.md` 에 mandatory checksum 검증 동작과 실패 시 복구 절차를 4-locale 모두 추가한다.
- **G4** — 기존 `multi-llm/cg-mode.md` 에 tmux env source-file 방식 설명 (`~/.moai/.env.*` `0o600` 의존) 을 4-locale 모두 추가한다.
- **G5** — `CHANGELOG.md` `[Unreleased]` 섹션에 SECURITY 항목 추가 (CWE 키워드 + commit hash 참조 포함). docs-site 의 보안 페이지에서 이 항목을 직접 링크한다.
- **G6** — 4-locale 정합성을 `scripts/docs-i18n-check.sh` 와 `moai spec lint --strict` 로 자동 검증한다.

## 3. Non-Goals

- 신규 보안 결함 발견 또는 정정 — SECURITY-CRIT-001 이후에 새로 발견된 항목은 본 SPEC scope 외, 별도 SPEC.
- 다른 Wave (B/C/D/E) SPEC 페이지 — `harness-engineering.md`, `agent-guide.md`, `constitution.md` 등은 별도 Wave SPEC 처리.
- hugo 빌드 설정 변경 (`hugo.yaml`, `vercel.json`) — 본 SPEC은 콘텐츠만 작성.
- Version 스냅샷 발행 (`content/{locale}/v2/` 디렉토리 생성) — 모든 Wave 완성 후 v2.20.0-rc1 release 시점에 일괄 처리.
- cosign / sigstore 서명 기반 무결성 가이드 — SECURITY-CRIT-001 Non-Goals 와 정합. 별도 SPEC 후속.
- SECURITY-CRIT-001 의 source-of-truth (`spec.md`, commit body) 변경 — docs-site 는 downstream 만, upstream SPEC 은 frozen.

## 4. Stakeholders

| 역할 | 관심사 |
|------|--------|
| 메인테이너 (GOOS) | v2.20.0-rc1 release tag 발행 전 4-locale docs 갱신 완료 |
| Enterprise 사용자 | 공급망 (update) + 토큰 저장 (settings.local.json) + tmux (cg-mode) 의 보안 변화 사실 + 점검 절차 명시 |
| 다국어 사용자 (en/ja/zh) | ko 와 동등한 시점에 동등한 정보 도달 |
| 보안 감사인 | CWE 매핑 + 변경 commit hash + 자체 점검 명령 |
| 후속 Wave SPEC 작성자 | `advanced/security-notes.md` 를 cross-reference 가능한 안정적 anchor 로 활용 |

## 5. Requirements (EARS)

### 5.1 신규 페이지 `advanced/security-notes.md` (REQ-DSEC-001 ~ REQ-DSEC-003)

- **REQ-DSEC-001 [Ubiquitous]** — `docs-site/content/{ko,en,ja,zh}/advanced/security-notes.md` 4개 파일이 모두 존재한다.
- **REQ-DSEC-002 [Ubiquitous]** — 각 `security-notes.md` 는 다음 6 섹션 헤딩을 가진다 (locale 별 번역 가능, 의미 일관):
  1. Why (보안 변화 요약 + v2.20.0-rc1 context)
  2. CWE-732 / CWE-552 — settings.local.json permission hardening
  3. CWE-214 — tmux IPC token 노출 방지 (source-file injection)
  4. CWE-345 — update flow mandatory checksum verification
  5. 점검 체크리스트 (Self-Audit Checklist)
  6. References (CHANGELOG 링크 + SPEC-V3R5-SECURITY-CRIT-001 + 관련 commits)
- **REQ-DSEC-003 [Ubiquitous]** — 각 페이지의 hugo frontmatter 는 동일 구조 (`title`, `description`, `weight`, `tags`) 를 사용하며, ko canonical 의 `weight` 값을 다른 3 locale 이 동일하게 사용한다.

### 5.2 `advanced/settings-json.md` 갱신 (REQ-DSEC-004)

- **REQ-DSEC-004 [Event-driven]** — When 사용자가 `advanced/settings-json.md` 의 `settings.local.json` 섹션을 읽을 때, the system **shall** `0o600` permission 권고 + 자체 점검 명령 (예: `stat -c '%a' .claude/settings.local.json` 또는 macOS `stat -f '%A'`) + `advanced/security-notes.md` cross-reference 를 제공한다. 4-locale 모두 동일 정보를 포함한다.

### 5.3 `getting-started/update.md` 갱신 (REQ-DSEC-005)

- **REQ-DSEC-005 [Event-driven]** — When 사용자가 `getting-started/update.md` 를 읽을 때, the system **shall** 다음 정보를 제공한다 (4-locale 모두):
  - `moai update` 가 release 의 `checksums.txt` 를 mandatory 로 검증함을 명시
  - 다운로드 실패 시 3회 retry (지수 백오프 2s/4s/8s) 후 `ErrChecksumUnavailable` 로 abort 함을 명시
  - 실패 시 사용자 복구 절차 (네트워크 재시도 / proxy 확인 / `--skip-checksum` 같은 우회 옵션 부재 명시) 안내
  - `advanced/security-notes.md` cross-reference

### 5.4 `multi-llm/cg-mode.md` 갱신 (REQ-DSEC-006)

- **REQ-DSEC-006 [Event-driven]** — When 사용자가 `multi-llm/cg-mode.md` 를 읽을 때, the system **shall** 다음 정보를 제공한다 (4-locale 모두):
  - tmux env 주입이 source-file (`mkstemp` mode `0o600`) 경유 + `tmux source-file` 명령 사용 + 주입 직후 unlink 됨을 명시
  - argv 채널에 token 이 노출되지 않음을 명시 (CWE-214 mitigation)
  - `~/.moai/.env.glm` source 파일은 사용자 책임 하 `0o600` 유지 의무 강조
  - `advanced/security-notes.md` cross-reference

### 5.5 4-locale 동시성 (REQ-DSEC-007 ~ REQ-DSEC-010)

- **REQ-DSEC-007 [Ubiquitous]** — ko 페이지가 작성되면, 동일 PR 내에 en/ja/zh 페이지도 동시에 작성된다.
- **REQ-DSEC-008 [Ubiquitous]** — 4-locale 간 heading slug (앵커) 가 cross-reference 가능하도록 일관성 유지된다 (예: `#cwe-732-settings-local-json-permission-hardening` 같은 영문 slug 또는 locale 별 일관 규칙).
- **REQ-DSEC-009 [State-driven]** — While locale 별 페이지가 작성될 때, the system **shall** `docs-site-i18n-rules.md` §17.2 (Mermaid TD only — LR/BR 금지) 와 §17.1 (금지 URL `docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr` 사용 금지) 를 준수한다.
- **REQ-DSEC-010 [Unwanted]** — If 본문에 이모지가 포함된다면, then the system **shall not** 페이지 commit 을 허용하지 않는다 (이모지는 metadata 외 본문 금지 — research doc §6 i18n 정책).

### 5.6 CHANGELOG 갱신 (REQ-DSEC-011)

- **REQ-DSEC-011 [Ubiquitous]** — `CHANGELOG.md` 의 `[Unreleased]` 섹션에 SECURITY 항목이 추가되어 다음을 포함한다:
  - CWE-732 / CWE-552 (settings.local.json 0o600)
  - CWE-214 (tmux source-file injection)
  - CWE-345 (mandatory checksum verification)
  - SECURITY-CRIT-001 commit hashes (`b48bd86cb`, `10776c4b8`, `ee1335282`, `b4e7115cb`)
  - `advanced/security-notes.md` 페이지 링크

### 5.7 spec-lint clean (REQ-DSEC-012)

- **REQ-DSEC-012 [Ubiquitous]** — 본 SPEC 디렉토리 (`.moai/specs/SPEC-V3R5-DOCS-SECURITY-001/`) 는 `moai spec lint --strict` 실행 시 0 NEW 위반을 가진다 (pre-existing baseline 외).

### 5.8 Out of Scope (Constitution)

### 5.8.1 Out of Scope

- 신규 보안 결함 발견 또는 정정 — SECURITY-CRIT-001 이후에 새로 발견된 항목은 본 SPEC scope 외, 별도 SPEC.
- 다른 Wave (B/C/D/E) SPEC 페이지 — `harness-engineering.md`, `agent-guide.md`, `constitution.md` 등은 별도 Wave SPEC 처리.
- hugo 빌드 설정 변경 (`hugo.yaml`, `vercel.json`) — 본 SPEC은 콘텐츠만 작성.
- Version 스냅샷 발행 (`content/{locale}/v2/` 디렉토리 생성) — 모든 Wave 완성 후 v2.20.0-rc1 release 시점에 일괄 처리.
- cosign / sigstore 서명 기반 무결성 가이드 — SECURITY-CRIT-001 Non-Goals 와 정합. 별도 SPEC 후속.
- SECURITY-CRIT-001 의 source-of-truth (`spec.md`, commit body) 변경 — docs-site 는 downstream 만, upstream SPEC 은 frozen.
- `scripts/docs-i18n-check.sh` 자체 보강 — 신규 페이지 4-locale parity 자동 검출 누락 시 보강은 별도 follow-up SPEC.
- Native speaker 의 ja/zh 번역 품질 리뷰 — 본 SPEC 은 메인테이너 작성. native review 는 별도 운영 SPEC.

## 6. Risks

| ID | Risk | Severity | Mitigation |
|----|------|----------|------------|
| **R-DSEC-001** | ja/zh 번역 품질 격차 — 메인테이너의 ja/zh 숙련도 차이로 ko/en 대비 부정확하거나 부자연스러운 번역 발생 가능 | Medium | en 페이지를 brigde 로 활용 (영어→일본어/중국어 번역 시 의미 유실 최소화). 보안 용어는 영문 원문 (CWE, OWASP, source-file injection 등) 병기. 후속 native speaker 리뷰는 별도 SPEC 처리. |
| **R-DSEC-002** | security-notes.md 가 5K 단어 초과 시 §17.3 예외 적용 필요 (`translation_status: pending` 마킹) | Low | 본 SPEC 추정 ko 페이지 ~3K 단어, 4-locale 전체 ~12K 단어 — 단일 페이지는 5K 미만 예상. 초과 시 plan-phase 에서 페이지 분할 (예: cwe-732.md / cwe-214.md / cwe-345.md 분리) 결정. |
| **R-DSEC-003** | 4-locale heading slug 불일치 → cross-reference (anchor link) 깨짐 | Medium | hugo 자동 slug 대신 명시적 영문 slug 채택 권장 (예: `### CWE-732 — settings.local.json permission {#cwe-732}`). 사후 `scripts/docs-i18n-check.sh` 가 검증. |
| **R-DSEC-004** | docs-i18n-check.sh 가 신규 페이지의 4-locale parity 검증 누락 | Low | scripts/docs-i18n-check.sh 동작 사전 확인 (현재 `find {ko,en,ja,zh}/<rel>` 자동 비교 추정). 누락 시 plan-phase 에서 script 보강 또는 수동 AC 추가. |
| **R-DSEC-005** | SECURITY-CRIT-001 의 정확한 사용자 가시 변화가 향후 추가 정정으로 변경될 가능성 (post-merge follow-up) | Low | 본 SPEC 은 `03a2552a2` 머지 시점의 동작을 기준. 후속 변경 발생 시 별도 docs SPEC 으로 정정 (불변 패턴 아님). |

## 7. Acceptance Criteria 요약

본 절은 `acceptance.md` 와 1:1 대응한다. **7개 binary AC** = 4-locale parity (AC-DSEC-001) + i18n script PASS (AC-DSEC-002) + 금지 URL 0 (AC-DSEC-003) + Mermaid TD-only (AC-DSEC-004) + CHANGELOG SECURITY 항목 (AC-DSEC-005) + 본문 이모지 0 (AC-DSEC-006) + spec-lint clean (AC-DSEC-007).

상세 AC 는 `acceptance.md` 참조.

## 8. References

### 8.1 SPEC References

- SPEC-V3R5-SECURITY-CRIT-001 (`.moai/specs/SPEC-V3R5-SECURITY-CRIT-001/`) — upstream source of truth, status `implemented` v0.2.0, merged PR #1032 commit `03a2552a2`
- SPEC-V3R5-WORKFLOW-LEAN-001 — Tier M 3-artifact 규약 (merged `c0eb30da6`)
- SPEC-V3R5-LATE-BRANCH-001 — main 직접 commit, PR 시점 branch 분리 (merged `664cd6eae`)

### 8.2 Research / Rules

- `.moai/research/docs-site-v2.14-to-HEAD-update-plan-2026-05-20.md` §4 (SECURITY-CRIT-001 P0 보안 섹션 설계), §6.2 (Phase 1 — Security Foundation), §8.2 (Wave A 첫 SPEC)
- `.moai/docs/docs-site-i18n-rules.md` §17.1 / §17.2 / §17.3 / §17.4 (URL 표준 / Mermaid TD-only / 4-locale 동기 / 스냅샷 패치-NO)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (12-field canonical)
- `.claude/rules/moai/workflow/spec-workflow.md` (Tier S/M/L 차등)

### 8.3 Commits

- `b48bd86cb` — M1 settings.local.json 0o600 hardening (CWE-732/552)
- `10776c4b8` — M2 tmux sensitive env source-file injection (CWE-214)
- `ee1335282` — M3 mandatory checksum verification with retry (CWE-345)
- `b4e7115cb` — M4 cross-cutting verification + frontmatter
- `03a2552a2` — SECURITY-CRIT-001 merge to main (PR #1032)

### 8.4 CWE / OWASP

- CWE-732 — Incorrect Permission Assignment for Critical Resource
- CWE-552 — Files or Directories Accessible to External Parties
- CWE-214 — Invocation of Process Using Visible Sensitive Information
- CWE-345 — Insufficient Verification of Data Authenticity
