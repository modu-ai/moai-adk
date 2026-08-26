# t274 격차 표 — CHANGELOG [3.1.3] 26항목 문서 존재 매핑 (사본)

> 출처: `.moai/specs/SPEC-DOCS-V313-CATCHUP-001/spec.md` §1 (version 0.3.0). 원본이 정본이다 — 본 사본은 카드 증거 경로 요건(`evidence: .moai/reports/t274/`)용 발췌.
> 판정 분류: **D** 3 (이미 문서화됨) / **U** 11 (갱신 착지) / **N** 4 (신규 페이지 착지 — 운영자 승인) / **NA** 8 (문서화 대상 아님, 근거 보존) + **V1–V8** (version SSOT 갭 — 격차 표 첫 항목 요건이었던 version-sync 정렬 포함)

## 1. 문제 — 측정된 형태 (검증된 격차 표)

CHANGELOG.md `## [3.1.3] - 2026-08-24` (177–306행)의 항목을 전수 추출했다: **Added 13 + Changed 4 + Fixed 9 = 26 항목** (`grep -n '^## \[' CHANGELOG.md` → `[3.1.3]` 177행, 다음 버전 `[3.1.2]` 307행; `[Unreleased]` 8행 항목은 범위 밖). 각 항목의 문서화 존재 여부를 docs-site 4로케일(`docs-site/content/{ko,en,ja,zh}/`, 파일 목록 체크섬 `98d2b226e6569dd7b07a8ce9ee4d3e5c` ×4 — 4로케일 파일 파리티 100%)과 README 4파일(`README.ko.md` ko-canonical + `README.md`/`README.ja.md`/`README.zh.md`, H2 12개 ×4)에 대해 grep으로 관측했다. 모든 관측은 이 워크트리(baseline `e07a6d0f4`)에서 이번 실행으로 얻은 것이다 (VCI §2).

판정 분류: **D** = 이미 문서화(작업 불필요) · **U** = 기존 페이지 갱신 · **N** = 신규 페이지 필요(구조 소관 — 별도 승인) · **NA** = 문서 표면 없음(결함 수정·내부 개선).

### 1.1 Added 13항목

| ID | 항목 | 판정 | 대상 · 근거(관측된 명령 결과) |
|----|------|------|-------------------------------|
| A1 | 루트 `AGENTS.md` standing contract (harness 공통 계약, 빌드 가드, 11 문서 stub화) | **N** | `grep -rln 'AGENTS\.md' docs-site/content` → 0파일; README 4파일 `grep -rn 'AGENTS\.md'` → 0행. 문서화 경로가 없음 — 신규 페이지 후보 |
| A2 | 11 에이전트 듀얼 게시 (`.codex/agents/moai/*.toml`, `agentemit` 결정적 생성) | **N** | `grep -rln '\.codex/agents' docs-site/content README…` → 0파일. A1과 동일 주제(codex dual-harness) |
| A3 | `.agents/skills` 스킬 미러 (codex-cli용, 사용자 저장소 밖) | **N** | `grep -rln '\.agents/skills' …` → 0파일. A1과 동일 주제 |
| A4 | `internal/codexadapter` — Codex 훅 어댑터 라이브러리 (11-이벤트 표, 아직 호출부 없음) | **N** | `grep -rln 'codexadapter\|hook adapter' …` → 0파일. A1과 동일 주제 — "아직 호출부 없음"이므로 문서 깊이는 승인 시 결정 |
| A5 | `/moai feedback` 스크러빙 계약 (`moai feedback scrub`/`queue` 동사, 취약점 분류기, 재시도 큐, `feedback.auto_submit`) | **U** | `moai-feedback.md` 95행에 auto_submit 게이트는 문서화됨(4로케일). 그러나 `grep -n 'scrub\|queue' ko/utility-commands/moai-feedback.md` → scrub/queue 동사·분류기·재시도 큐 0행 → 기존 페이지 갱신 |
| A6 | `workflow.todo.enabled` — 백로그 큐 끄기 스위치 (부재 시 on) | **U** | `grep -rln 'todo\.enabled' docs-site/content README…` → 0파일 → `advanced/config-sections.md`(+ `moai-todo.md` 크로스링크) 갱신 |
| A7 | `moai todo` 큐 자기 분석 (add 시 측정, `moai todo analyze`) | **U** (README만) | docs-site는 완전 문서화 — `grep -n 'analyze\|측정\|중복' docs-site/content/ko/utility-commands/moai-todo.md` → 95·97·99·166·195·196행 ("기록만 남김", "정확 중복 거절", "Jaccard 0.80"). 그러나 `sed -n '352p' README.ko.md` → 동사 나열 `add·list·next·done·unpick·drop·undrop·edit·move`에 `analyze` 부재 (4파일 동일) → README:352 한 줄 갱신 [D7] |
| A8 | MCP 5도구 선택적 `project_root` (워크트리 감사·재시작 주의 포함) | **D** | `grep -n 'project_root' ko/guides/mcp-server.md` → 96–114행 — 6도구 나열, 워크트리 사례 표, "재시작 전까지 예전 동작" 주의까지 완전 문서화(4로케일 존재) |
| A9 | `moai-domain-svg-infographic` 커넥터 지오메트리 검사 (`SVG070`–`SVG074`) | **NA** | `grep -rln 'SVG06\|SVG07\|aria-labelledby' …` → 0파일. SVG 규칙 세부는 원래 docs-site 문서 표면 밖(skill 내부 린트 규칙) — v3.1.3 미반영이 아니라 문서화된 적이 없음 |
| A10 | SVG 산출물 접근성 이름 (`SVG060`–`SVG064`) | **NA** | A9와 동일 근거 |
| A11 | `moai-domain-design-dna` 다이어그램 프로파일 (`.design-dna/` 지속, mermaid/drawio 임포터) | **U** | `grep -n 'design-dna\|diagram' ko/advanced/skill-guide.md` → 159행 스킬 소개 행만 존재 — 프로파일·지속·임포터 미반영 → 기존 행 갱신. README 4파일 745행 design-dna 절도 동일 갱신 대상 |
| A12 | `moai update`/`moai init` 스킬 미러 symlink→copy 폴백 통지 | **U** | `grep -n 'symlink' docs-site/content/ko/cli-reference/update.md docs-site/content/ko/getting-started/init-wizard.md` → 0행 — CHANGELOG가 `moai update`와 `moai init` 양쪽 통지이므로 두 페이지 모두 갱신 [D7] |
| A13 | `/moai gate` typecheck 축 (#1592) | **D** | `grep -n '타입' docs-site/content/ko/utility-commands/moai-gate.md` → 7·10·23·63·87·107행 — "린트·포맷·타입 검사·테스트" 4축을 이미 서술. #1592로 코드가 문서를 따라잡은 형태 — 문서 갱신 불필요 |

### 1.2 Changed 4항목

| ID | 항목 | 판정 | 대상 · 근거 |
|----|------|------|-------------|
| C1 | 에이전트 model/effort 매트릭스 → judgment-weighted 정책 (심사·조율 행 high, `manager-spec`/`manager-develop` 전 열 medium, `manager-docs` sonnet/low, 어떤 행도 max 아님) | **U** | `grep -n 'max\|medium\|sonnet/low' ko/advanced/profile-matrix.md` → 27–35행 매트릭스가 구 정책(`manager-spec opus/high`, `manager-develop opus/max`, `super-advisor opus/max`, `manager-docs opus/medium`), 47행 "max 두 행 배정", 51행 max 별칭 서술 → `profile-matrix.md` 갱신. **`multi-llm/model-policy.md` 112–123행에 같은 매트릭스 중복 존재 — 두 페이지 모두 갱신** |
| C2 | `manager-lead` 매트릭스 합류 (기존 inherit sentinel) | **U** | `profile-matrix.md` 27–35행 표에 `manager-lead` 행 없음 → C1 갱신에 포함 |
| C3 | GLM reasoning-effort 상한 `max`로 상향 (low 제외 전 effort → reasoning max, 무설정 기본 max) | **U** | `grep -rn 'reasoning' docs-site/content/ko/multi-llm/` → 0행. `model-policy.md` 201행 "GLM effort 오버레이" 개념 언급만 있고 Claude effort→GLM reasoning 매핑·상한은 미서술 → `model-policy.md` 갱신 |
| C4 | `ANTHROPIC_*` 환경변수 나열 완전판 | **NA** | `grep -n 'ANTHROPIC_' ko/cli-reference/launchers.md` → 59행 (GLM 자격증명 주입 설명). 인식 키 목록 전체를 나열하는 사용자 문서 표면이 없음 — 내부 개선, 문서화 의무 없음 |

### 1.3 Fixed 9항목

| ID | 항목 | 판정 | 대상 · 근거 |
|----|------|------|-------------|
| F1 | 홈 디렉터리 하위 상태 조회가 `~/.moai/state`로 귀결되던 결함 (조회가 프로젝트 루트에서 멈춤) | **NA** | `moai-clean.md`는 `--home` allowlist 중심(208–249행) — 결함 수정이지 기능 변화 아님 |
| F2 | 관리 루트의 symlink가 `moai update`를 벽돌로 만들던 결함 | **NA** | 결함 수정 이력 — 문서 의무 아님 (A12 갱신 시 함께 언급할 수 있으나 의무 아님) |
| F3 | codex/GLM 감사 백엔드가 안 읽은 코드에 판정 내리던 결함 (diff 수집 실패 → `inconclusive`, 백엔드 미호출) | **U** | `grep -c 'inconclusive' ko/advanced/multi-model-audit.md` → `0`. 수렴 절차(57행)에 diff-수집-실패 케이스 없음 → `multi-model-audit.md` 수렴 절차 갱신 |
| F4 | 명시된 `inconclusive` 판정을 pass로 합성하던 결함 | **U** | F3과 동일 grep — `inconclusive` 0행 → F3 갱신에 포함 |
| F5 | `audit_multi` 수렴 판정이 감사된 트리 아래 기록되지 않던 결함 | **U** | F3/F4와 동일 페이지 — `mcp-server.md` 96–114행(project_root)은 이미 문서화돼 있으므로 `multi-model-audit.md` 수렴 절차에 한 절 추가 |
| F6 | `project_root` 심볼릭 링크 경계 우회 결함 (canonicalize) | **NA** | 내부 경계 검사 디테일 — A8의 `mcp-server.md` 워크트리 사례 표가 사용자 관점을 이미 다룸 |
| F7 | `moai init` 수집·폐기되던 답변의 실제 적용 (autonomy tier, 4 워크플로 토글, audit/codex 선택) | **D** | `init-wizard.md` 7행 "정한 값은 전부 YAML로 저장" + 44행 Page 3 서술 — fix 후 문서와 이미 일치 |
| F8 | 웹 콘솔 서버 기동 중 SIGTERM 즉사 결함 | **NA** | `grep -n 'SIGTERM\|signal' ko/advanced/moai-web-console.md` → 0행. 내부 결함 수정 — 사용자 가시 문서 표면 아님 |
| F9 | constitution의 `agent-authoring` 교차참조 복구 | **NA** | 내부 rules 참조 복구 — 사용자 문서와 무관 |

### 1.4 항목 외 — version SSOT 갭 (카드 지시 조사, t272 잔여 재확인; 전수 재조사 [D1])

26항목과 별개로, i18n 규칙 §7 release-sync 의무(“모든 버전 표시는 릴리즈 PR에서 함께 갱신”)가 v3.1.3에서 지켜지지 않았음을 관측했다. 전수 재조사 명령: `grep -rn 'v3\.1\.[0-9]' docs-site/content/{ko,en,ja,zh} README.ko.md README.md README.ja.md README.zh.md` — 매칭에서 역사 인용(`added_in`/`new-badge` frontmatter, "v3.1.1에서 개명"류 서술)과 `--version` 플래그 문법 예시를 제외한 제품 버전 **표시(display)** 전부가 아래 8건이다:

| ID | 표면 | 관측 (RED-now) | 목표값 (green) |
|----|------|------|------|
| V1 | `docs-site/hugo.toml` 55–56행 | `version = "v3.1.2"`, `releaseDate = "2026-08-21"` | `v3.1.3` / `2026-08-24` |
| V2 | README 4파일 491행 statusline 예시 | `🗿 v3.1.2` ×4 | `🗿 v3.1.3` |
| V3 | README 4파일 766행 update-prompt 예시 | `🗿 v3.1.1 -> 🗿 v3.1.2` ×4 | `🗿 v3.1.2 -> 🗿 v3.1.3` |
| V4 | `docs-site/content/*/advanced/statusline.md` 22행 ×4로케일 | `🗿 v3.1.2` | `🗿 v3.1.3` |
| V5 | `docs-site/content/*/getting-started/faq.md` ×4로케일 | 37행 `🗿 v3.1.1 -> 🗿 v3.1.2` (ko 기준 라인; en 33행), 40–41행 설명 버전 표기, 46행 단독 `🗿 v3.1.2` | 예시 `🗿 v3.1.2 -> 🗿 v3.1.3`, 설명 동조, 단독 `🗿 v3.1.3` |
| V6 | `docs-site/content/*/guides/claude-cloud.md` ×4로케일 | `go install github.com/modu-ai/moai-adk/cmd/moai@v3.1.2` (ko 68행, en 66행) | `@v3.1.3` |
| V7 | README 4파일 493행 statusline 예시 브랜치 세그먼트 | `[WT] release/v3.1.2 +3` ×4 | `release/v3.1.3` |
| V8 | `docs-site/content/*/utility-commands/moai-feedback.md` 이슈 템플릿 표 예시값 — 로케일별 관측 상이 [R1 정정] | ko:62 `v3.1.1` (실제 스테일); en:62·ja:66·zh:66 플레이스홀더 `v10.8.0` (v3.1.x 문자열 아님 — `grep -n 'v3\.1\.\|v10\.8\.0'` 4로케일 재측정) | 4로케일 모두 `v3.1.3` (version-column example — 규칙 §7; 플레이스홀더도 version-sync 축이 읽는 표시이므로 정렬) |

**제외 기록 (관측했으나 갱신하지 않는 것)**: `cli-reference/update.md` 171행(`ko`)·175행(`en`)의 `moai update --version v3.1.0-rc1` — `--version` 플래그 문법 시연이지 제품 버전 표시가 아님. `added_in: "v3.1.1"` frontmatter와 `{{< new-badge v3.1.1 >}}`, "v3.1.1에서 개명/들어옴"류 서술(manager-lead.md:19, kanban-mode.md:217·281, agent-teams.md:102, home-hygiene.md, doctor.md 등) — 역사 인용은 규칙 §7이 명시적으로 보존 대상으로 지정. README 배지(24행)는 4파일 모두 `Release-v3.1.3`으로 이미 올바르다.

**이 갭의 원인(릴리즈 프로세스가 hugo.toml·예시 표시를 동기화하지 않는 구조)은 별도 카드 권장 사항이며 이 SPEC이 흡수하지 않는다** — 이 SPEC은 증상(스테일 값)만 바로잡는다 (§6 Out of Scope 참조).

### 1.5 집계

- 26항목 = **D 3** (A8·A13·F7) + **U 11** (A5·A6·A7-README·A11·A12·C1·C2·C3·F3·F4·F5) + **N 4** (A1–A4, codex dual-harness 주제로 통합 가능) + **NA 8** (A9·A10·C4·F1·F2·F6·F8·F9)
- 항목 외: version SSOT 갭 8건 (V1–V8)
- 작업 대상: U 11항목 + V1–V8 (기존 페이지/파일 갱신) + N 4항목(신규 페이지 — 승인 관문)
- 명령 표기 규약: 격차 표·AC의 `README…`·`README 4파일`은 `README.ko.md README.md README.ja.md README.zh.md` 전체 나열의 축약이며, `docs-site/content`는 네 로케일 전체를 가리킨다. 판정에 쓰는 grep 패턴 자체는 축약 없이 실행 가능한 형태로 기재한다 [D9].

