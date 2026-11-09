# RC1 기능 범위 선보고 — 빌드 5db3069a9 대비 최종 설계 (2026-11-09, 리더, 제안·미확정)

이 문서는 RC1 포함 범위 확정이 아니라 사용자 피드백을 받기 위한 현황 보고다.

## 기준
- 코드: `5db3069a94effe4ae330137dc96ff3a01c5be222`(origin/v2026.11.0). 스테이징 바이너리 `/tmp/moai-rc1-local-target/release/moai-cli`, `moai-cli 2026.11.0-rc.1`, sha256 `5fbcbf8a84559ef9de69117d5f811b5f18e5e3e2d64bedc078e6b6775ab46ae4`(설치 안 함).
- CI: rc-ci는 tip SHA 자체엔 없고 #2248 merge ref(run 42564487804)에서 통과.
- 설계: (1) 최신 정본 스냅샷 seq44(`references/space-pages/moai-adk-final-design-page-5490c4-seq44-20261109-1411.md`, sha256 234e1d47…, §1-§25) — 수신·검토. (2) 기능표의 행 번호는 이전 Downloads 사본(textonly, 862줄, sha256 b4242abd…) 기준으로 조사됨 — seq44와 절 번호가 다름(seq44는 §23-§25 추가). seq44 기준 재대조는 진행 중.

## 기능 상태(코드 실행 경로로 대조)
| # | 기능 | 상태 @5db3069a9 | 연결 | 다른 프로젝트에서 지금 사용 |
|---|---|---|---|---|
| 1 | 프로젝트 setup(profile + AGENTS.md managed 구간) | 포함·시험(project-plan/apply → agents-plan/apply). 딥 인터뷰는 skill 문서 절차뿐 | #1995 #1997 #2170 | 가능(소스 빌드, 호스트 로드 not_run) |
| 2 | AGENTS.override.md(Codex)·Claude 대응 | 미구현(감지만, CLAUDE.md 안 씀, @AGENTS.md import 제안만) | EP-H 미발행(Q-A) | 불가 |
| 3 | `moai` skill 설치·update·recovery | 부분: 사용자 범위 `~/.claude/skills/moai`·`~/.agents/skills/moai` 복사 + receipt. 프로젝트 범위 HOLD(P04), openai.yaml plan만. **결함 #2268: 기존 비관리 폴더에 RC1 파일을 섞음** | #1999 #2000 #2242 #2259(PR) #2268 | 조건부: 기존 moai skill 없는 HOME만 안전 |
| 4 | hooks·settings.json·config.toml·permissions·MCP 설정 | settings.json(outputStyle·statusLine)만 포함. hooks #2260·config.toml #2257·permissions #2251은 PR 대기, MCP 설정 미구현 | #2248 #2260 #2257 #2251 #2243 | settings만 가능 |
| 5 | 자가 학습 루프 | 미구현(op 없음, 승인 수준 미확정) | EP-H H3/H4 미발행 | 불가 |
| 6 | 5단계 PLAN→RUN→VERIFY→AUDIT→SYNC | 대부분 포함·시험(plan-admit, run-verify, audit-reduce, scope-check). SYNC는 절차(gh/git)이며 전용 op 없음 | #2002-#2004 #2007 | 조건부(CLI 가능, skill 호스트 로드 not_run) |
| 7 | agent/leader 자율 반복 | 부분: queue-next·leader-tick·claim·effect·worktree op 포함. mode workflow가 이 op들을 호출하지 않음(문서만). §25 루프는 #2255 #2263-#2265 진행 | #2161 #2162 #2255 #2263-#2265 | 조건부 |
| 8 | 호스트 런처 cc/glm/codex | 포함·시험(-l/-a 경고, 역할 세션에서만 skip 플래그) | #2053 #2164 #2251(PR) | 가능 |
| 9 | MCP | 미구현(HOLD-C1 '다음 버전'; 재설계 지시) | #2243 #2266(문서 PR) | 불가 |
| 10 | Atlas/Code Map | 미구현(스캐너 #2228 PR, 근거표 #2249 문서 PR) | EP-C | 불가 |
| 11 | 바이너리 설치·self-update·release 산출물 | install-inspect(읽기)만. self-update 없음. cargo-dist #2258 FAIL(SBOM) | #2193 #2016 | 소스 빌드만 |
| 12 | Forge ops(claim·report·audit-status) | 포함·시험. links-probe #2226 PR(FAIL: merge trailer) | #2005 #2173 #2172 | 조건부 |
| 12b-f | archive-plan / context budget / cost-report / parity-check / P09 layout | archive #2236 PR, context 포함(토크나이저 없음), cost·parity 포함, P09 #2235 PR | 각 Issue | 포함된 것만 |
| 13 | Pi·Desktop | 미구현(조사 문서만, 사용자 지시로 후순위) | #2199 | 불가 |
| 14 | 실세션 인증 하네스 | PR #2250, 실측 not_run | #2245 #2014 | 불가 |

## 호스트별 로드 차이(공식 문서 기준, 2026-11-08/09 확인)
| 항목 | Claude Code | Codex | moai-cli 현재 |
|---|---|---|---|
| AGENTS.md | v2.1.277+ 로드, CLAUDE.md 변형 있으면 억제 | 루트→cwd 폴더별 1파일, 32 KiB | managed 구간 생성 |
| AGENTS.override.md | 네이티브 아님 | 같은 폴더에서 AGENTS.md 대신 선택 | 생성 안 함 |
| skills | `~/.claude/skills`·`.claude/skills` | `~/.agents/skills`·`.agents/skills` | 사용자 범위만 |
| hooks | settings·plugin·frontmatter | hooks.json·config.toml(신뢰 프로젝트만) | 0개(#2260 PR) |
| 설정 | settings.json | config.toml | settings.json 2키만 |

## 기본 agent·skill 목록
설계(seq44 §1 L18·L61, §3 L62·L171): 기본 제공은 공통 `moai` skill 하나, 전문 agent·skill 묶음 기본 설치 없음, 프로젝트 인터뷰 뒤 필요한 역할만 생성. **이름이 정해진 기본 agent 목록은 설계에 없음.** RC1: `skills/moai`(route READ·SETUP·UPDATE·PLAN·RUN·VERIFY·AUDIT·SYNC + MODE-LEADER·MODE-AGENT), 역할 파일은 generation-plan/apply 입력으로만 생성(기본 역할 없음). 목록을 정할지는 사용자 결정(제안 대상).

## 로컬 사용·보존·복구 계획(제안)
1. 계획: 위 표에서 '가능' 항목만 대상. 2. 스테이징: 위 바이너리(설치 안 함). 3. 격리 시험: 임시 HOME·임시 프로젝트(#2262 agent-9 진행 중). 4. 사용자 경로 적용: #2268 수정 전에는 실제 HOME update-apply 금지(기존 `~/.claude/skills/moai` 10-22, `~/.agents/skills/moai` 10-29 존재). 적용 시 dry-run plan → 확인 → apply → recovery-plan. Go `moai`(~/go/bin/moai)·`.moai/`·`~/.moai/`·개인 AGENTS 파일은 건드리지 않음. 5. 실호스트 실행·공개 배포: 별도 승인.

## 미검증
빌드 외 로컬 시험은 이 보고 작성자가 실행하지 않음(격리 smoke: package-inspect·install-inspect만). 모든 host_loaded = not_run.
