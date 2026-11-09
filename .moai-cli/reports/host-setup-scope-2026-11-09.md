# Claude Code · Codex 프로젝트 설정 지원 범위 보고 (2026-11-09, 리더)

기준 트리: origin/v2026.11.0 = 6ee819cae (읽기 전용 조사, 주 체크아웃 b37cdf2dd는 491 커밋 뒤라 사용하지 않음).
출처 표기는 `git show origin/v2026.11.0:<path>` 기준 줄 번호.

## 1. 한 줄 요약
moai-cli가 지금 실제로 쓰는 것은 공용 파일(AGENTS.md managed 구간, 선택적 AGENTS.local.md, .gitignore), 호스트별 role·skill 파일(.claude/, .agents/·.codex/), 그리고 각 호스트 홈 skill 폴더에 복사되는 `moai` skill뿐이다. **훅·MCP 설정·permissions·AGENTS.override.md·CLAUDE.md는 쓰지 않는다.** Claude settings.json 쓰기는 라이브러리에만 있고 CLI 명령이 없다. 실제 Claude/Codex 세션에서 로드되는지는 전부 not_run.

## 2. 호스트별 산출물

### Claude Code
| 산출물 | 원본 → 대상 | 구현 | 근거 |
|---|---|---|---|
| AGENTS.md managed 구간 | 프로필 → `<repo>/AGENTS.md` (`moai:managed` 마커 사이) | 예 (agents-apply / generation-apply) | src/agents.rs:16,161 · v2.1.277+이고 CLAUDE.md 변형이 없을 때만 로드, claude-doctor가 예측 (src/claude.rs:136-264) |
| CLAUDE.md / `@AGENTS.md` import 줄 | — | 설계만 (H2, 결정 Q-A 대기) | doctor가 제안만 (src/claude.rs:247-250); 기존 파일은 수정하지 않음 (src/agents.rs:365-373) |
| `moai` skill 홈 복사 | skills/moai/ → ~/.claude/skills/moai/ (단순 복사) | 예 (update-plan/apply) | src/update.rs:3-5,51-61 · 프로젝트 단위 복사는 HOLD(P04) |
| Claude 전용 skill frontmatter | claude::projection | 라이브러리만, CLI 없음 | src/claude.rs:576-589,632 · tools 부여·hooks 추가 필드는 거부 |
| 외부 skill | 로컬 체크아웃 → .claude/skills/<name> | 예 (skill-source-apply) | src/sources.rs:506-515 |
| 생성 skill 투영 | .claude/skills/<id>/SKILL.md | 예 (generation-apply) | src/generation.rs:575-578 |
| role 파일 | .claude/agents/<role>.md (name/description/tools) | 예 (opt-in) | src/generation.rs:342-354,516-518 |
| .claude/settings.json (outputStyle·statusLine만) | 3-way merge + 영수증 .moai-cli/manifests/claude-settings.json | 부분: 라이브러리만, CLI 없음 | src/claude.rs:266-269,429-570 · docs/claude-adapter-contract.md "Not done here" |
| output-styles·statusline 스크립트 | adapters/claude/… → .claude/output-styles/, .moai-cli/statusline.py | 임시 Python만, Rust 제품은 호출 안 함 | adapters/claude/install.py |
| hooks · permissions · MCP config | — | **미구현** (런처가 세션 시작 시 `--dangerously-skip-permissions`만 전달) | src/claude.rs:267-268, src/session.rs |

### Codex
| 산출물 | 원본 → 대상 | 구현 | 근거 |
|---|---|---|---|
| AGENTS.md managed 구간 | 공용 파일 동일, 32 KiB 초과 경고 | 예 | src/agents.rs:405 |
| AGENTS.override.md | — | 설계만 (H1) · 감지만 하고 생성 안 함 | src/hosts/codex.rs:274-276,346 · personal-coding-core §3.1–3.2 |
| `moai` skill 홈 복사 | skills/moai/ → ~/.agents/skills/moai/ | 예 (update-apply) | src/update.rs:54 |
| skill용 agents/openai.yaml | codex-projection-plan | 계획만, 쓰기 없음 | src/hosts/codex.rs:56-67,128-201 · 계약 갭 C5 |
| 외부·생성 skill | .agents/skills/<id>/ | 예 | src/sources.rs:508, src/generation.rs:577 |
| role 파일 | .codex/agents/<role>.toml (name/description/developer_instructions) | 예 (opt-in) | src/generation.rs:356-363,520-522 |
| .codex/config.toml (trust·status line·bundled skills·fallback·size) | — | **미구현** (TOML 파서 없음, 요청은 HOLD) | src/hosts/codex.rs:16-18,31-33,175-183 |
| hooks · MCP config | — | **미구현** (doctor가 .codex/hooks.json·config.toml 존재만 확인) | src/hosts/codex.rs:191,350-356 |
| 실행 플래그 | `codex -s workspace-write -a on-request` | 예 (런처만) | src/session.rs codex_argv |

실세션 검증: docs/support-matrix.md:5,26-29 — Claude·Codex 호출 시험 전부 not_run. 통과한 호스트 행은 `claude plugin validate`(정적) 하나.

## 3. 훅(hooks) 현황
- Claude Code: 0개. RC1 훅 집합은 리더 결정(P04=A)으로 비어 있음. settings 키 `hooks`(src/claude.rs:266-269)·skill frontmatter `hooks`(:577-578) 모두 거부. docs/host-support-matrix.md "MoAI RC1 hook은 0개".
- Codex: 0개. `"hooks":{"generated":[]}`(src/hosts/codex.rs:191); .codex/hooks.json·inline [hooks] 생성 없음(docs/codex-adapter-contract.md:10). trust·inline hooks는 not_evaluated. 갭 C4 "필요 시 별도 Issue", C7(훅 핸들러·exit 규칙 원문 대조) 미검증.
- 미지원: 어떤 훅 쓰기도, PermissionRequest 처리, Codex trust 검사, cloud hooks.
- 설계는 더 기대함: 설계 문서(textonly 293-318, 438-442)는 양 호스트 훅으로 검사·영수증 수집을 두고, 742행은 Go 기준선에 Codex 훅 이벤트 어댑터 12개가 있었다고 적음 — 이식된 것 없음.

## 4. 공통(통합) vs 개별(호스트별)
- 공통 단일 원본: AGENTS.md managed 구간(src/agents.rs:161-208)·AGENTS.local.md(:210); skills/moai/(SKILL.md frontmatter는 name+description만, workflows, routes.json)와 digest 고정 references/source-manifest.json — 호스트마다 독립 복사, 심링크 금지(설계 §4, 166행); .moai-cli/workflows/<id>/, .moai-cli/manifests/{generation,sources}.json.
- 호스트별: skill 폴더(.claude/skills vs .agents/skills), role 파일(.md vs .toml), Codex agents/openai.yaml(계획만), Claude settings.json 키(라이브러리만), 실행 플래그. 계획만 있고 미구현: Codex AGENTS.override.md, Claude용 CLAUDE.md import 또는 .claude/ 복사.
- 설계 매핑: 설계 textonly §4 표(154-160행), 배치 트리(555-575행).

## 5. 이슈 계획 커버리지
rc.2 카드(#2153–#2209) 중 호스트 설정을 건드리는 것:
- A1 #2160 `--mode` 인자(skills/moai), Codex는 leader 거부
- A4 #2163 `/goal` 붙여넣기 안내(양 호스트)
- A5 #2164 `-l/-a` 런처 플래그 폐기
- B3 #2170 SETUP 커밋·Issue 발행 승인 분리(H1 경계)
- B9 #2176 MCP parity 조사만(구현 blocked)

EP-H 제안(미발행, docs/decisions-2026-11-08.md:115-118 · personal-coding-core §8): H1 AGENTS.override.md 생성(Codex), H2 Claude 어댑터(CLAUDE.md import vs .claude/ 복사, Q-A 대기), H5 프로젝트 전용 agent·skill 조사.

**카드가 전혀 없는 갭**
1. 양 호스트 훅(Codex C4, 설계의 훅 역할)
2. Claude settings.json CLI 명령(outputStyle/statusLine) + Python 어댑터 제거
3. TOML 파서 + Codex config.toml(trust·status line·fallback·size) — C1–C3
4. Codex skill 복사 실제 설치 + openai.yaml, 홈 vs 프로젝트 위치 HOLD(C5)
5. 프로젝트 단위 skill 복사(P04 HOLD)
6. MCP 설정 생성(B9는 조사만)
7. 런처 skip 플래그를 대체할 `permissions` 설정 키(docs/rust-product-path.md:446 "later Issue")
8. 양 호스트 실세션 시험(AC-08C-3/4/6, Codex C6 — 운영자 몫)

## 6. 결론
이슈 계획은 **수립되지 않음**(부분). 호스트 설정 범위에서 rc.2는 런처·mode·goal 안내·MCP 조사까지만 다루고, 훅·settings/config 쓰기·override/CLAUDE.md·skill 설치 완성·permissions·실세션 검증은 카드가 없다. EP-H(H1·H2·H5)는 사용자 결정 Q-A/Q-B/Q-C 대기로 미발행.
