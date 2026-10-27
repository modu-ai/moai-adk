# plan.md — SPEC-CLAUDE-INSTRUCTION-FILES-001

> stateless 산출물 (status 축 무상태). 카드 t1985, lane-16. HEAD `cb30e05b3e` 기준.

## §A Context

spec.md §A의 검증된 현행 지도가 구현의 출발점이다. 변경 표면은 4개 그룹으로, 스코프 맵 `instruction_migration.change_groups`와 일치한다:

1. **정규 이름·생성** — `internal/defs/files.go`, `internal/core/project/initializer.go`, `validator.go`, `internal/template/templates/AGENTS.md.tmpl`→CLAUDE.md 투영, `internal/template/harness_fs.go`
2. **마이그레이션·소유권 보존** — `internal/cli/migrate_local_instructions.go`, `internal/cli/update/plan/plan.go`, `internal/cli/init_resume.go`
3. **런타임 리더/작성기** — `internal/cli/harness/install.go`, `doctor_harness.go`, `internal/hook/instructions_loaded.go`, `pre_tool.go`(frozen 유지), `internal/harness/curator/dispatch.go`, `layer5.go`, `internal/config/token_budget_guard.go`
4. **자체 저장소 전환 + 룰/문서** — 이 저장소의 AGENTS.md/AGENTS.local.md, `.claude/rules/**`, `.claude/rules/local/**`, distributed template mirror, docs

## §B Known Issues

- **t1984 겹침**: `AGENTS.md.tmpl`, docs, 공유 감사 launcher/RPC 3파일(`internal/cli/codex_audit_launch.go`/`codex_audit_mcp.go`/`mcp_codex_audit.go`)이 t1984(lane-15)와 겹친다. 두 레인이 동시에 쓰면 안 된다.
- **이 저장소 자체가 첫 이관 대상**: 세션 하네스가 매 턴 AGENTS.md/AGENTS.local.md를 주입 중이다. 자체 전환은 run 단계에서 마이그레이션 verb 실행과 같은 커밋 경계에 둔다.
- **테스트 242 match-line**: 테스트 참조 대부분은 생성 방향 반전으로 자연 따라온다. historical/data-fixture 116줄은 건드리지 않는다(REQ-CIF-015).
- **`internal/cli/harness/install.go`의 CLAUDE.md 마커 블록**: 현재도 CLAUDE.md를 명명한다(주석 직독 확인). 방향 전환 후 이 파일의 표적이 정규 이름이 되도록 검증만 하면 된다 — 신규 메커니즘 불요.

## §C Pre-flight (게이트)

**[HARD] 순서 전제 — t1984.** 본 카드의 RUN 단계는 (a) card t1984가 착지(병합)되고 (b) 리더가 t1985를 해제한 뒤에만 시작한다. RUN 첫 커밋 전에:

1. t1984 브랜치를 본 워크트리에 흡수(`git merge`/rebase — 겹치는 `AGENTS.md.tmpl`·docs·감사 launcher/RPC의 t1984 측 변경을 먼저 받는다).
2. 흡수 후 `internal/cli/codex_audit_launch.go`/`codex_audit_mcp.go`/`mcp_codex_audit.go`의 존재·변경 여부를 재독; 본 카드는 이 3파일을 편집하지 않는다(spec.md §D Out of Scope).
3. RED 기준선 재측정: acceptance.md §D.1의 RED-now 명령을 흡수 후 트리에서 재실행해 SHA를 다시 고정한다(이관 후 재측정·재고정).

**검증 도구**: `go test` 대상 패키지 스코프, `golangci-lint`, `moai spec lint`(본 SPEC). 정적 확인(grep/ls)과 실행 검증을 모든 AC에서 구분해 기록한다 — 미실행은 결코 PASS가 아니다.

## §D Constraints

1. 기존 exclusive-write/backup/refuse/rollback 프리미티브 재사용 — 병렬 메커니즘·신규 ledger·신규 하위시스템 금지 (REQ-CIF-005).
2. 무손실·바이트 정확(백업), 비파괴 충돌 거부, 멱등, 읽기-쓰기 사이 원본 변경 중단 (REQ-CIF-006~011).
3. symlink·비정규·읽기불가·경로이탈 입력 거부 (REQ-CIF-009).
4. 공유/비공개/비밀 분리 — 모호 콘텐츠는 미해결·로컬 유지, 본문 로깅 금지 (REQ-CIF-012).
5. 역사 백업·감사 증거 불변 (REQ-CIF-015).
6. frozen 4-이름 계약 유지 (REQ-CIF-014).
7. 시간 추정 금지 — 우선순위 라벨과 단계 순서로만 계획.

## §E Self-Verification

- 요구사항 수 16 (REQ-CIF-001~016), AC 수 16 (AC-CIF-001~016) — Tier M 예산(각 ≤16) 준수. 대응은 acceptance.md §D.2/§D.3.
- 모든 AC는 정적 확인과 실행 검증을 구분 — 미실행은 결코 PASS가 아니다.
- 릴리스-블로킹 AC는 2-셀 채택(RED-now + green path)을 충족한다(acceptance.md §D.1).
- `moai spec lint` 본 SPEC 무결통과 — run/sync에서 관측해 기록.

## §F Milestones

역가능성(reversibility) 내림차순 — 방향 결정·데이터 이동이 앞, 기계적 따라오기가 뒤.

### M0 — 사전 흡수 게이트 (Priority: High, 결정 블로커)
- t1984 착지 확인·브랜치 흡수, 겹침 파일 재독, RED 기준선 재측정·SHA 재고정 (§C).
- 산출: 흡수 커밋 + 재측정 기록(progress.md §E.2).

### M1 — 정규 이름·생성 경로 전환 (Priority: High)
- `internal/defs/files.go`: ClaudeMD/AgentsMD 주석·문서 반전(CLAUDE 정규, AGENTS 레거시).
- `AGENTS.md.tmpl` → `CLAUDE.md.tmpl` 재명명(내용은 t1984 흡수본 기준) + deployer 투영 전환.
- `initializer.go` / `validator.go` / `init_resume.go` 프로브 방향 전환.
- `harness_fs.go` `normalizeHarnessReferences` 방향 반전(REQ-CIF-004) — 생성물이 AGENTS 참조를 만들지 않음을 테스트.
- AC: AC-CIF-001, 002, 004, 016(프로브 부분).

### M2 — 마이그레이션 동사 방향 전환 + 무손실 원칙 (Priority: High, 데이터 이동)
- `migrate_local_instructions.go` 방향 반전(AGENTS.local.md → CLAUDE.local.md; shared AGENTS.md → CLAUDE.md 경로 추가). 기존 프리미티브 재사용(REQ-CIF-005).
- 신규: 읽기-쓰기 사이 원본 변경 감지 중단(REQ-CIF-010), 바이트 정확 백업·권한 보존 검증, 경로이탈 거부.
- 테이블 주도 테스트: 정상 이동/공재 거부/양쪽 상이 거부/멱등 noop/symlink·비정규·읽기불가/부분 실패 rollback/동시 변경 중단.
- AC: AC-CIF-005~010.

### M3 — 런타임 리더/작성기 재타겟 (Priority: Medium)
- `doctor_harness.go`, `instructions_loaded.go`(`resolveInstructionFile` CLAUDE-우선 + 레거시 AGENTS 폴백 라벨), `curator/dispatch.go` TierSurfaceMap, `layer5.go` @import 문자열, `token_budget_guard.go` 측정 표면, `harness/install.go` 검증, `update/plan/plan.go` 리스크 분류 유지(4이름).
- frozen 계약 불변 테스트(`derived_test.go`/`frozen_compat_test.go`) 통과 확인(REQ-CIF-014).
- AC: AC-CIF-003, 012, 013, 016(나머지).

### M4 — 자체 저장소 전환 + 룰/문서 스윕 (Priority: Medium)
- 이 저장소 AGENTS.md/AGENTS.local.md를 마이그레이션 verb로 이동(무손실 백업, 커밋 경계에서). private 내용 CLAUDE.md 승격 없음 확인(REQ-CIF-012, REQ-CIF-016).
- `.claude/rules/**`, `.claude/rules/local/**`, template mirror의 활성 AGENTS 참조 방향 전환(단순 문자 치환이 아니라 참조 대상 전환; historical/data-fixture 인용은 보존).
- docs 동기화(manager-docs sync 단계와 연계). 역사 증거 불변 확인.
- AC: AC-CIF-011, 014, 015.

### M5 — 검증 총괄 (Priority: High)
- acceptance.md 전 AC 실행(정적/실행 구분 기록), `moai spec lint`, 패키지 스코프 테스트·커버리지 85%+, lint 통과 관측.
- 미실행/거부된 도구는 Gaps로 기록(verification-claim-integrity §3.1).

## §G Anti-Patterns

- **눈먼 전역 치환**: `AGENTS.md`→`CLAUDE.md` 문자열 일괄 바꿈으로 사용자 파일·역사 증거 손상 금지. 마이그레이션은 verb 통해서만.
- **병렬 메커니즘 발명**: 새 백업/락/원장 구조 금지 — 기존 프리미티브 재사용.
- **감사 launcher 동시 편집**: t1984 소유 3파일을 본 카드 run에서 건드리는 것.
- **모호 콘텐츠 자동 게시**: 분류 불확실 시 추측 금지 — 미해결 보존·보고.
- **빈 스윕 PASS**: grep 0-hit를 성공으로 읽지 않는다 — 스윕된 수를 먼저 확정(verification-completeness §1.1).
- **vague AC**: 미실행 테스트·관측 안 된 출력의 PASS 주장 금지.

## §H Cross-references

- spec.md §A 현행 지도, §C 요구사항, §D Out of Scope(t1984 경계)
- acceptance.md — AC-CIF-001~016, GUIDE-01..12 조정표
- progress.md — §E 진행 기록(§E.1만 plan-phase 작성)
- 스코프 맵: `/tmp/moai-factory-analysis-20261027/adk-native-codex-withdrawal-claude-guidance-scope.json`
