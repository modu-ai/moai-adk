<!-- moai-template:moai-task@0.1.0 -->
## Redaction
- [x] 비밀값·개인정보·절대 홈 경로·원문 로그·비공개 내용 제거 확인

## Goal and user value
SETUP의 첫 vertical slice. `moai-cli`가 사용자 소유 `AGENTS.md`에 MoAI 관리 구간만 baseline/current/new 3-way로 병합하고, 선택 시 `AGENTS.local.md`를 1회 생성하며, 생성 provenance를 `.moai/manifests/generation.json`에 남긴다. 공통 skill package 뼈대(`skills/moai`)를 저장소에 두어 두 host projection의 canonical 원본으로 삼는다. 같은 입력의 재요청은 diff 0이어야 한다.

## Non-goals
- 실제 Claude Code / Codex host의 skill discovery·명시 호출·로딩 검증 (not_run)
- 인터뷰 orchestration, 외부 skill 설치, native agent 생성
- `<state_root>` per-file apply receipt, UPDATE 3-way(host 설치본), commit/push/release

## SPEC reference
- design: 홈 설치와 프로젝트 생성 기능 계약 §2·§4 (page_83f162ef14d08191917913738184f167) · 생성 파일 원문 템플릿 §1·§3 `project-agents`/`project-local`/`generation-manifest`/`router-core`/`router-routes`/`workflow-*` (page_396f36aa5d588191b92f8572bccc3bab) · 메인 계획 D11·D12·D16 (page_43faf222414c81919de71f7d7830e26c)
- status: 설계 결정은 확정, 구현 acceptance는 아래 AC로 판정. 승인 snapshot digest: 없음(unapproved draft 아님, 설계 Page 스냅샷 2026-11-08T04:49Z 기준)

## Acceptance criteria and verification
- AC-1 사용자 본문 보존: 관리 구간 밖 바이트 불변, CRLF·개행 없는 EOF 포함 · verify: `cargo test --offline --locked agents_generation_is_idempotent_and_preserves_user_body`
- AC-2 재요청 diff 0 (GENERATION-IDEMPOTENT) · verify: 같은 시험 + 이 저장소 dogfooding 2회차 sha256 비교
- AC-3 사용자 편집 보존·충돌 미적용 (GENERATION-USER-EDIT) · verify: `agents_user_edits_are_preserved_or_reported_as_conflict`
- AC-4 AGENTS.local.md opt-in(`none|ignored|tracked`), 재작성·재생성 없음 · verify: `agents_local_file_is_opt_in_and_never_rewritten`
- AC-5 다른 프로젝트 replay·plan 변조·대상 변경·manifest 변조 거부, CLAUDE.md/AGENTS.override.md 보존·보고 · verify: `agents_plan_rejects_replay_tamper_missing_reference_and_shadowing_is_reported`
- AC-6 symlink 부모 거부, 부분 적용 후 재개 (중단 후 재개) · verify: `agents_apply_refuses_symlinked_parents_and_recovers_from_partial_apply`
- AC-7 skill 뼈대: SKILL.md + routes.json(8 route) + workflows 8개, host 전용 메타데이터 없음 · verify: `moai-cli package-inspect skills/moai` digest 안정, grep allowed-tools/$ARGUMENTS 0건
- AC-8 독립 AUDIT 통과 · verify: 구현과 분리된 세션(agent-2)의 2차 감사 verdict

## Write scope and exclusions
- may change: `src/`, `tests/product_cli.rs`, `skills/moai/`, `docs/rust-product-path.md`, `.gitignore`, `.github/`
- must not touch: `.claude/`, `.moai/`(생성 대상 `project.yaml`·`manifests/` 제외), `references/`, `experiments/`, Go 저장소, 원격 `main`

## Dependencies
none (milestone 2026.11.0-rc.1의 첫 항목)

## Host scope
both (Claude Code and Codex) — 이번 slice는 host-neutral 파일 생성까지, host 로딩은 not_run

## Evidence at creation
로컬 작업 트리(브랜치 `2026.11.0-rc.1`, unborn HEAD, 미커밋)에서 실행:
- `cargo test --offline --locked` → exit 0, 통합 12 passed(기존 7 + 신규 5), 단위 3 passed
- `cargo clippy --offline --locked --all-targets -- -D warnings` → exit 0
- `cargo fmt --check` → exit 0
- `python3 -m unittest discover -s tests -p '*_test.py'` → 19 passed
- `git diff --check`, `git diff --cached --check` → exit 0
- dogfooding: `project-plan/apply`(`.moai/project.yaml` 생성) → `agents-plan/apply`(AGENTS.md append, AGENTS.local.md create, generation.json create) → 2회차 `state: unchanged`, 4개 파일 sha256 동일
- `skills/moai` digest: `sha256:454f1e5550027944832250985021f57c0a81acc25231a6de7d5cc31bbbed5253` (moai-package-tree-v1)
- not_run: host discovery·호출, Codex 32 KiB 실측, GitHub Issue form 렌더링(기본 브랜치 아님)
- blocked: 없음


> 위 원문 보존 구역의 측정값(통합 12 passed, Python 19, digest 454f1e55…)은 2026-11-08 발행 시점 값으로 **stale**이다. 최신 값은 아래 상세 설계 절과 소유 댓글 revision 5(통합 15/단위 3, Python 22, digest ba6d3a20…)를 따른다.

---

<!-- moai-template:moai-feature@0.3.0; 상세 설계 2026-11-08; 상관관계용 메타데이터일 뿐 권한이 아니다 -->
```yaml
logical_key: precursor-1973
milestone: 2026.11.0-rc.1
spec_ref:
  id: SPEC-RC1-precursor-1973
  revision: proposed
  digest: null
source_pages:
  - https://chatgpt.com/space/page_66acadc28bdc8191875556986da72e6b
claim:
  state: unclaimed
  session_id: null
  epoch: null
permission_scope: proposed
evidence_status: not_run
design_revision: 2026-11-08-r1
```
# [task] M1 SETUP slice: AGENTS.md 관리 구간 생성기(agents-plan/apply) + 공통 skill 뼈대(skills/moai) (정본 03-A/03-C/02-A의 선행 vertical slice)

## 목표와 사용자 가치
SETUP의 첫 vertical slice. `moai-cli`가 (1) `project-plan/apply`로 `.moai-cli/project.yaml` portable profile을 만들고(03-A 일부), (2) `agents-plan/apply`로 사용자 소유 `AGENTS.md`의 MoAI 관리 구간만 baseline/current/new 3-way로 병합하며 `AGENTS.local.md`를 opt-in 1회 생성하고 `.moai-cli/manifests/generation.json`에 provenance를 남기며(03-C 일부), (3) host 중립 공통 skill 뼈대 `skills/moai/`(SKILL.md·routes.json 8 route·workflows 8개)를 둔다(02-A 일부). 이미 완료: AC-1~AC-8 passed(댓글 revision 5: 4차 독립 감사 PASS, 5차 경로 분리 확인 PASS), commit f6808390 push, 보고된 시험값 cargo test 통합 15/단위 3·clippy·fmt·Python 22·diff --check ×2 exit 0, 경로 `.moai-cli/`·`~/.moai-cli/` 분리(Go `.moai/` 보존). 이 세션 관측: AGENTS.md 관리 구간 99-111행 존재(파일 15,150 bytes < Codex 32 KiB), `.moai-cli/project.yaml`·`manifests/generation.json` tracked, Python 22 tests OK. 가치: 03-A·03-C·02-A의 남은 작업이 이 slice의 실제 코드·시험·Gap 목록 위에서 시작되게 하고, 중복 구현을 막는다.

## 비목표와 범위
비목표(기존 유지 + 이관): host discovery·명시 호출·로딩 검증(→02-B/02-C), 인터뷰 orchestration(→03-A), 외부 skill 설치·provenance(→03-B), 맞춤 native agent/team·템플릿 필드 렌더링(→03-C), `<state_root>` per-file apply receipt·rollback(→01-B), 상태 어휘·exit code 3분류·protocol schema(→00-D), host 설치본 UPDATE 3-way(→04-B/04-C), Windows/Linux·CRLF·hard link 미지원 FS E2E(→08-C), Python adapter 구형 `.moai/` receipt 이행(→04-A). 이 Issue에서 새 구현은 하지 않는다. 범위: 완료된 것·이관할 Gap·증거 링크를 본문에 정리하고 종료 방식을 결정하는 것.

## 근거와 확정 결정
| 주장 | 출처 | 확인일 · 버전 | 상태 |
|---|---|---|---|
| AC-1~AC-8 PASS는 댓글 revision 4(4차 독립 감사 PASS: cargo test 15/3, clippy, fmt, Python 22, diff --check ×2)와 revision 5(5차 PASS: Go `.moai/`·`~/.moai/` 읽기·쓰기 코드 0, 구형 fallback 0)에 기록. revision 3은 revision 2의 '15 passed'가 미관측 값이었음을 정정(실측 14/1 → 수정 후 15/0). | #1973 댓글 id 6717787001(rev 3), 6717836188(rev 4), 6717874525(rev 5) | 2026-11-08 · commit f6808390 | 검증(댓글 존재) / stale(PASS 값은 이 세션 미측정) |
| 이 세션은 cargo test/clippy/fmt를 재실행하지 않았다(읽기 전용; target/ 쓰기 회피). Python 전체 22 tests OK만 관측. | local: PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p '*_test.py' | 2026-11-08  | 검증 |
| 구현 지도: `src/main.rs` 6 subcommand(package-inspect, project-plan, project-apply, package-resolve, agents-plan, agents-apply), `src/agents.rs`(33 KB), `src/lib.rs`(PROJECT_STATE_DIR/HOME_STATE_DIR), `tests/product_cli.rs` 통합 시험 15개(agents 8개 포함), stdin 64 KiB 상한, exit 0/1. | /Users/goos/MoAI/moai-adk/src/main.rs · tests/product_cli.rs · docs/rust-product-path.md | 2026-11-08 · Cargo.toml 2026.11.0-rc.1, clap =4.6.1, serde =1.0.228, serde_json =1.0.150, sha2 =0.10.9 | 검증 |
| skills/moai 뼈대: SKILL.md(name: moai, 디렉터리명 일치, host 전용 frontmatter 없음), references/routes.json(8 route, `classification_is_authorization: false`, SETUP/UPDATE만 `requires_binary: true`), workflows 8개. digest `sha256:ba6d3a20942f67ff80d576b56cb8215fb26f36fc153526e979674839d8fff063`(revision 2 보고, 이 세션 재계산 안 함). | /Users/goos/MoAI/moai-adk/skills/moai/SKILL.md · references/routes.json · #1973 댓글 rev 2 | 2026-11-08  | 검증 |
| Agent Skills 명세: `name` 1-64자 소문자·숫자·하이픈, 디렉터리명과 일치; `description` 1-1024자; `allowed-tools`는 experimental; SKILL.md 500줄 이하 권장; references/는 필요 시 로드. 현재 SKILL.md는 이 제약을 충족(name moai, 약 20줄). | https://agentskills.io/specification | 2026-11-08  | 검증 |
| Claude Code는 v2.1.277+에서 작업 디렉터리 또는 상위에 CLAUDE.md/.claude/CLAUDE.md/CLAUDE.local.md가 없을 때만 AGENTS.md를 읽고, `AGENTS.local.md`·`AGENTS.override.md`·`.agents/`는 "Not read". `/config` Project instructions로 `claude-md-and-agents-md` 선택 가능. 이 세션 자체도 AGENTS.md를 로드했음(system context) — 단 자동 발견 증거는 별도 세션에서 측정해야 함. | https://code.claude.com/docs/en/memory#agents-md | 2026-11-08 · Claude Code v2.1.277+ (v2.1.281 이전 일부 세션 제한) | 검증 |
| Codex는 `~/.codex`(CODEX_HOME) 전역 AGENTS.override.md→AGENTS.md, 프로젝트 루트→cwd 순으로 디렉터리당 1파일, 합산 `project_doc_max_bytes` 기본 32 KiB에서 중단. 현재 AGENTS.md 15,150 bytes. | https://learn.chatgpt.com/docs/agent-configuration/agents-md · local ls -la AGENTS.md | 2026-11-08  | 검증 |
| `cargo test --locked`는 Cargo.lock 변경 시 실패, `--offline`은 네트워크 차단, `--frozen`은 둘 다. 통합 시험 실행 시 바이너리가 자동 빌드되고 `CARGO_BIN_EXE_<name>`로 경로 전달. | https://doc.rust-lang.org/cargo/commands/cargo-test.html | 2026-11-08  | 검증 |
| 보존 주의: dogfooding이 Go 세대 `.moai/`에 `project.yaml`·`manifests/generation.json` 2파일을 새로 만들었고(덮어쓰기 0, Go 참조 0) 삭제·되돌리기 없이 보존 중. Go 소스는 루트 `AGENTS.md`·`AGENTS.local.md`를 frozen instruction file로 읽음(`internal/contract/frozen.go`). | #1973 댓글 rev 3·4, docs/github-plan-proposal.md §7 | 2026-11-08  | 미확인 |
| #1973은 blocked_by·blocking 모두 없음(정본 §5 leaf가 아니라 선행 slice). EP-04(#1979)·EP-03(#1978)의 sub-issue가 아니며 milestone #1에만 속함. | local: gh api …/issues/1973/dependencies/*; manifest links | 2026-11-08  | 검증 |
| 이 작업은 사용자 승인으로 PR 없이 `origin/2026.11.0-rc.1`(현 `v2026.11.0`)에 직접 push됐다(orphan commit, 61 files). 새 지시(이슈별 PR→v2026.11.0, 로컬 build/test 증거)는 이후 작업부터 적용된다. | #1973 댓글 rev 5; 사용자 지시 2026-11-08 | 2026-11-08  | 검증 |

## SPEC 요구(GEARS)
- [REQ-1973-01] 시스템은 항상 `AGENTS.md`의 `<!-- moai:managed:begin id=moai-workflow -->`…`end` 구간만 소유하고 구간 밖 바이트(CRLF·개행 없는 EOF 포함)를 보존해야 한다.
- [REQ-1973-02] WHEN 같은 profile·workflow_reference로 `agents-plan`을 다시 실행하면, 시스템은 모든 파일을 `unchanged`로 보고하고 `agents-apply`는 바이트 변화 0이어야 한다(GENERATION-IDEMPOTENT).
- [REQ-1973-03] IF 사용자가 관리 구간 안을 편집한 뒤 템플릿이 바뀌면, 시스템은 적용하지 않고 `conflict`로 보고해야 한다(GENERATION-USER-EDIT).
- [REQ-1973-04] WHERE `local_file`이 `none|ignored|tracked` 중 하나이면, 시스템은 `AGENTS.local.md`를 최대 1회 생성하고 이후 재작성·재생성하지 않으며 `created_once`에 기록해야 한다.
- [REQ-1973-05] IF plan digest 불일치·다른 프로젝트 replay·대상 변경·manifest 변조·workflow_reference 부재·symlink 부모가 감지되면, 시스템은 쓰기 없이 거부해야 한다.
- [REQ-1973-06] WHILE apply가 대상 파일 쓰기와 manifest 쓰기 사이에서 중단되면, 다음 apply는 렌더링과 같은 블록을 baseline으로 채택해 거짓 conflict 없이 재개해야 한다.
- [REQ-1973-07] 시스템은 항상 `<project>/.moai-cli/`와 `~/.moai-cli/`만 쓰고 Go 세대 `.moai/`·`~/.moai/`를 읽기·쓰기·이동하지 않아야 한다(GH-RC-PATH-ISOLATION).
- [REQ-1973-08] 시스템은 항상 `skills/moai/SKILL.md`가 Agent Skills 명세(name 소문자·디렉터리 일치, description ≤1024자)를 따르고 host 전용 메타데이터(`allowed-tools`, `$ARGUMENTS`)를 갖지 않아야 한다.
- [REQ-1973-09] 시스템은 항상 호출 결과에 `host_loaded: false`·`execution_verified: false`를 유지하고, host 로딩은 02-B/02-C에서 실측되기 전까지 not_run으로 보고해야 한다.
- [REQ-1973-10] WHEN 이 Issue를 종료하면, 시스템은 Gap 각각을 담당 leaf Issue(#1987 00-D, #1989 01-B, #1992/#1993 02-B/C, #1995 03-A, #1997 03-C, #1998 04-A, #2014 08-C)로 이관한 링크를 남기고 `closed_as completed`로 닫아야 한다(Gap이 미이관이면 열어 둔다).

## 인수 기준과 검증
- AC-1973-01: 사용자 본문 보존(CRLF·EOF) 시험 통과 · 검증: `CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-product-target cargo test --offline --locked agents_generation_is_idempotent_and_preserves_user_body` · 결과: stale (이전 세션 보고값, 이 세션 미측정 → 재측정 필요)
- AC-1973-02: 재요청 diff 0 · 검증: `같은 시험 + 임시 프로젝트 dogfooding 2회차 sha256 비교 (이 저장소 ˋ.moai/ˋ 안 dogfooding 금지)` · 결과: stale (이전 세션 보고값, 이 세션 미측정 → 재측정 필요)
- AC-1973-03: 사용자 편집 보존·충돌 미적용 · 검증: `cargo test --offline --locked agents_user_edits_are_preserved_or_reported_as_conflict` · 결과: stale (이전 세션 보고값, 이 세션 미측정 → 재측정 필요)
- AC-1973-04: AGENTS.local.md opt-in·재작성 없음 · 검증: `cargo test --offline --locked agents_local_file_is_opt_in_and_never_rewritten` · 결과: stale (이전 세션 보고값, 이 세션 미측정 → 재측정 필요)
- AC-1973-05: replay·변조·참조 부재 거부, shadowing 보고 · 검증: `cargo test --offline --locked agents_plan_rejects_replay_tamper_missing_reference_and_shadowing_is_reported` · 결과: stale (이전 세션 보고값, 이 세션 미측정 → 재측정 필요)
- AC-1973-06: symlink 부모 거부·부분 적용 재개 · 검증: `cargo test --offline --locked agents_apply_refuses_symlinked_parents_and_recovers_from_partial_apply interrupted_update_refreshes_baseline_instead_of_false_conflict` · 결과: stale (이전 세션 보고값, 이 세션 미측정 → 재측정 필요)
- AC-1973-07: skill 뼈대 digest 안정·host 전용 메타데이터 0 · 검증: `CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-product-target cargo run --offline --locked -q -- package-inspect skills/moai; grep -rc -E 'allowed-tools|\$ARGUMENTS' skills/moai | grep -v ':0'` · 결과: stale (이전 세션 보고값, 이 세션 미측정 → 재측정 필요)
- AC-1973-08: 독립 AUDIT PASS · 검증: `gh api repos/modu-ai/moai-adk/issues/1973/comments --jq '.[]|select(.body|test("감사 PASS|확인 PASS"))|.id'` · 결과: stale (이전 세션 보고값, 이 세션 미측정 → 재측정 필요)
- AC-1973-09: 경로 분리: Rust·Python에 Go `.moai/`·`~/.moai/` 참조 0 · 검증: `grep -rn -E '"\.moai/|\.moai"|~/\.moai[^-]' src tests adapters | grep -v moai-cli` · 결과: stale (이전 세션 보고값, 이 세션 미측정 → 재측정 필요)
- AC-1973-10: AGENTS.md 크기가 Codex 기본 budget 32 KiB 미만 · 검증: `wc -c AGENTS.md  # < 32768` · 결과: stale (이전 세션 보고값, 이 세션 미측정 → 재측정 필요)
- AC-1973-11: Gap 이관 링크가 담당 leaf Issue에 존재하고 본문에 기록됨 · 검증: `gh issue view 1973 -R modu-ai/moai-adk --json body --jq .body | grep -oE '#(1987|1989|1992|1993|1995|1997|1998|2014)' | sort -u | wc -l  # 8` · 결과: not_run (현재 본문에 이관 링크 없음)
- AC-1973-12: 전체 로컬 build/test 재측정이 종료 직전 기록됨 · 검증: `CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-product-target cargo test --offline --locked && CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-product-target cargo clippy --offline --locked --all-targets -- -D warnings && cargo fmt --check && PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p '*_test.py' && git diff --check` · 결과: not_run (이 세션 Python 22 OK만 관측)

## 구현 계획
1. 1. (H2 리더) `gh issue view 1973 --json body` sha256 기록 → 본문에 '완료된 것' 절(commit f6808390, AC-1~8 PASS 근거 댓글 rev 3·4·5 링크, 보고 시험값과 관측 주체 구분) 추가
2. 2. 본문에 'Gap 이관표' 추가: receipt→#1989, 상태 어휘·exit code·protocol→#1987, host discovery→#1992/#1993, 인터뷰→#1995, 템플릿 필드·native 역할→#1997, Python adapter receipt 이행→#1998, OS/CRLF/hard-link→#2014; 각 담당 leaf 그룹과 교차 확인(중복 구현 방지)
3. 3. 종료 직전 AC-1973-12 전체 로컬 재측정 1회 실행(cargo test/clippy/fmt·Python·diff --check) + `package-inspect skills/moai` digest 재계산 → 출력을 보고 댓글 revision 6에 첨부(사용자 지시: 로컬 build/test 증거)
4. 4. HOLD 2건(종료 방식, Go `.moai/` dogfooding 산물 처리) 사용자 결정 수신
5. 5. 결정에 따라 `closed_as completed` 또는 open 유지; 본문에 'PR 없이 push된 선행 작업' 명기
6. 6. 독립 검토(agent-2)가 이관표 누락 Gap 여부 확인

## 구현·검증 체크리스트
- [ ] 본문에 '완료된 것' 절 추가: commit f6808390, AC-1~8 PASS 근거(댓글 rev 3·4·5 링크), 보고 시험값과 관측 주체 구분
- [ ] 본문에 'Gap 이관표' 추가: receipt→#1989, 상태 어휘·exit code·protocol→#1987, host discovery→#1992/#1993, 인터뷰→#1995, 템플릿 필드·native 역할→#1997, Python adapter receipt 이행→#1998, OS/CRLF/hard-link→#2014
- [ ] 각 담당 leaf Issue 본문의 Dependencies/참조에 #1973 링크가 있는지 해당 그룹과 대조(중복 구현 방지)
- [ ] skills/moai digest 재계산해 본문 digest(ba6d3a20…)와 일치 확인
- [ ] AC-1973-12 전체 로컬 재측정 1회 실행·출력 첨부(사용자 지시: 로컬 build/test 증거)
- [ ] `.moai/project.yaml`·`.moai/manifests/generation.json`(Go 폴더 안 dogfooding 산물) 처리 방침 사용자 확인 → HOLD
- [ ] 종료 방식 결정(HOLD) 후 보고 댓글 revision 6 → closed_as completed 또는 open 유지
- [ ] 이 Issue는 PR 없이 push된 선행 작업임을 본문에 명기(소급 PR 없음)

## 의존성(실제 선행)
- none (선행 slice). 이 Issue의 산출물을 선행으로 삼는 leaf: #1995 03-A, #1997 03-C, #1991 02-A, #1989 01-B, #1987 00-D, #1998 04-A, #1992/#1993 02-B/C, #2014 08-C

## 파일·공유 interface 범위
- 변경 가능:
  - GitHub Issue #1973 본문(해시 확인 후 정확 diff)·보고 댓글 revision 6
  - docs/rust-product-path.md (Gap 이관 표기만, 코드 변경 없음)
- 변경 금지:
  - src/ · tests/product_cli.rs · skills/moai/ (새 구현은 담당 leaf에서)
  - .moai/ · ~/.moai/ · .claude/ · AGENTS.md 관리 구간 밖 · AGENTS.local.md
  - references/ · experiments/ · Go 저장소 · 원격 main
- 공유 interface:
  - `agents-plan/apply` typed JSON 계약(protocol_version, language, workflow_reference, local_file)
  - `.moai-cli/manifests/generation.json` 스키마(owned_files, created_once, profile_sha256)
  - `skills/moai/references/routes.json` 8 route·requires_binary 플래그
  - 관리 구간 마커 `<!-- moai:managed:begin id=moai-workflow -->`

## 테스트와 로컬 빌드
- CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-product-target cargo test --offline --locked
- CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-product-target cargo clippy --offline --locked --all-targets -- -D warnings
- cargo fmt --check
- PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p '*_test.py'
- git diff --check; git diff --cached --check
- CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-product-target cargo run --offline --locked -q -- package-inspect skills/moai
- 추가 fixture 없음(새 구현 없음); 이관 leaf가 각자 fixture 추가

## 완료 조건과 PR 연결
코드 변경 없음 → 새 PR 없음. 완료 조건: 본문에 완료 근거(commit f6808390bd83c53e44964ed1d4645727945ea1cc, 댓글 rev 3~5)와 Gap 이관표(8개 leaf 링크)가 있고, AC-1973-12 전체 로컬 재측정 출력이 댓글에 첨부되며, 종료 방식(HOLD)이 사용자 결정으로 확정된 뒤 `closed_as completed`. 독립 검토(agent-2)는 이관표의 누락 Gap 여부를 확인. 증거: 재측정 명령·exit code, digest 재계산 값, 이관 링크 목록.
통합 대상 브랜치: `v2026.11.0`. 이슈별 독립 worktree, 로컬 build/test 증거, 독립 review 뒤 PR.

## 미실행·위험
- not_run(이 세션): cargo test/clippy/fmt, package-inspect digest, Go 경로 참조 grep, Go 23 파일 sha256
- not_run(설계상): Claude/Codex host discovery·명시 호출, Codex 32 KiB 실측, hard link 미지원 FS, Windows/Linux CRLF, 동시 writer
- 위험: 보고 시험값(15/3)은 댓글 기반이며 rev 2에서 미관측 값이 PASS로 적힌 전례가 있음 → 종료 전 재측정 필수
- 위험: Go 세대 `.moai/` 안 dogfooding 산물 2파일이 Go 도구와 충돌할 가능성(참조 0으로 보고되었으나 재확인 안 함)
- 위험: AGENTS.md는 이 저장소에서 untracked(사용자 결정)라 다른 clone에서 관리 구간이 재생성될 수 있음
- 위험: Claude가 AGENTS.md를 읽는 조건(CLAUDE.md 부재, v2.1.277+)을 02-B가 실측해야 함

## 결정이 필요한 선택지 (미결은 HOLD)
- **#1973 종료 방식**  · 상태: HOLD
  - A. 지금 `closed_as completed`(AC 전부 PASS), Gap은 이관표 링크로만 · 영향: milestone closed 1/43, 후속 leaf가 Gap 책임 · 근거: 댓글 rev 5 AC-8 PASS; 설계 06:83 close ≠ 검증 (2026-11-08)
  - B. 8개 leaf에 이관 링크가 실제로 들어갈 때까지 open 유지 · 영향: Gap 누락 방지, Issue 한 개가 오래 열림 · 근거: 승인 계획 §10 G '남은 gap·close reason' (2026-11-08)
- **Go 세대 `.moai/project.yaml`·`.moai/manifests/generation.json`(dogfooding 산물) 처리** (P13 관련(격리 확정, 이행 단계 미결)) · 상태: HOLD
  - A. 보존(현 상태), 문서에 '사용자 변경 없음·Go 참조 없음' 기록 · 영향: Go 폴더에 비-Go 파일 잔존 · 근거: 댓글 rev 3·4, 사용자 보존 지시 (2026-11-08)
  - B. 사용자 확인 후 삭제 · 영향: Go 상태 디렉터리 정리, 되돌리기 불가 · 근거: 동일 (2026-11-08)
- **P06 관련: 상태 어휘·exit code·receipt 정렬을 00-D/01-B에서 할지 이 slice 후속 revision에서 할지** (P06) · 상태: HOLD
  - A. 00-D(#1987)·01-B(#1989)로 이관(권장 안 아님, 선택지) · 영향: #1973 코드 동결, 계약 변경은 leaf에서 · 근거: docs/rust-product-path.md Known gaps (2026-11-08)
  - B. #1973 revision 6에서 직접 수정 · 영향: leaf와 범위 중복 위험 · 근거: 승인 계획 §4 '같은 기능을 새 ID로 중복 생성하지 않음' (2026-11-08)

## 관련 논문·기술자료
- A Formal Investigation of Diff3 (Sanjeev Khanna, Keshav Kunal, Benjamin C. Pierce 2007 FSTTCS 2007, LNCS 4855, pp. 485-496, DOI 10.1007/978-3-540-77050-3_40) — 3-way merge(diff3)의 형식 명세와 '서로 떨어진 편집은 충돌하지 않는다'는 직관이 일반적으로 성립하지 않음을 보임 → 관리 구간 baseline/current/new 병합에서 자동 병합 대신 conflict로 멈추는 fail-safe 선택(REQ-1973-03)의 이론적 근거. 저자 호스팅 PDF 존재 확인(489 KB, 텍스트 추출은 실패), 서지는 Springer DOI·dblp 검색 결과 일치. · https://www.cis.upenn.edu/~sanjeev/papers/fsttcs07_diff3.pdf · 검증

## 베스트 프랙티스 채택/비채택
- Agent Skills 명세 준수: name=디렉터리명, description ≤1024자, SKILL.md ≤500줄, references/ 지연 로드, `allowed-tools` 미사용 · 판정: 채택 · 이유: 현재 skills/moai가 충족; 02-A 재설계의 기준 · 출처: https://agentskills.io/specification (2026-11-08)
- Claude Code AGENTS.md 로딩 조건(CLAUDE.md 부재·v2.1.277+)과 `claude-md-and-agents-md` 설정을 지원표에 기록; `AGENTS.local.md`는 자동 로드되지 않으므로 skill이 명시적으로 읽음 · 판정: 채택 · 이유: SKILL.md의 'read AGENTS.local.md explicitly' 지침과 일치, 02-B 실측 항목 · 출처: https://code.claude.com/docs/en/memory#agents-md (2026-11-08)
- Codex `project_doc_max_bytes`(32 KiB) 합산 한도를 AGENTS.md 생성 시 경고(현재 32 KiB 경고 구현됨) · 판정: 채택 · 이유: 02-C 실측 전까지 정적 검사로 대체 · 출처: https://learn.chatgpt.com/docs/agent-configuration/agents-md (2026-11-08)
- 재현 가능한 시험: `cargo test --offline --locked` + 격리 CARGO_HOME/TARGET_DIR · 판정: 채택 · 이유: AGENTS.md 개발 지침과 동일 · 출처: https://doc.rust-lang.org/cargo/commands/cargo-test.html (2026-11-08)
- `.gitignore` 판정은 문자열 규칙 일치 대신 `git check-ignore -v`로 측정 · 판정: HOLD · 이유: 현재 구현은 문자열 규칙 일치(Gap). 00-D/01-B 이관 시 채택 여부 결정 · 출처: https://git-scm.com/docs/git-check-ignore (2026-11-08)

## 조사 기록
- 2026-11-08 gh issue view 1973 --json body,labels; gh api issues/1973/comments(rev 1~5, 5건); dependencies blocked_by/blocking
- 2026-11-08 로컬: src/main.rs, Cargo.toml, tests/product_cli.rs 시험 이름 15개, skills/moai/SKILL.md·routes.json, docs/rust-product-path.md, AGENTS.md 관리 구간(99-111)·크기 15,150 B, git ls-files(.moai-cli 2파일 tracked), python unittest 22 OK
- 2026-11-08 /tmp/moai-rc1-approved-plan-20261108.md §4·§5(02-A·03-A·03-C 행)·§11 GH-RC-PATH-ISOLATION
- 2026-11-08 design-snapshot 00:343-360(두 host acceptance), 00:1397-1406(RC 격리 결정), 01:§7 P06·P13, 22(파일별 템플릿 목차), 12:234(격리 시험 계약)
- 2026-11-08 web: code.claude.com/docs/en/memory(AGENTS.md 절), learn.chatgpt.com agents-md, agentskills.io/specification, doc.rust-lang.org cargo-test, git-scm.com git-check-ignore; 논문 서지 검색·PDF 존재 확인(Khanna/Kunal/Pierce 2007)

## claim
상태: unclaimed. 배정은 리더의 조정 record에서 단일 writer로 기록하고 이 Issue의 소유 댓글에 표시한다. assignee·label은 lock이 아니다.

상태 규칙: Issue label·assignee는 lock·상태가 아니다. 진행·차단·완료는 소유 보고 댓글(stable run marker) 하나로 보고한다. 미결 기술 선택은 HOLD로 두고 템플릿이나 작성자가 확정하지 않는다. 실행하지 않은 검증은 not_run이다.
언어 규칙: 제목·본문·SPEC·체크리스트·설명 문구는 한국어로 쓰고, 코드 식별자·명령·경로·파일명은 원형을 유지한다.

출처: https://chatgpt.com/space/page_66acadc28bdc8191875556986da72e6b (Page updated_at 2026-11-08T05:42:40Z) · 상세 설계 근거: `.moai-cli/manifests/issue-design-research.corrected.json`