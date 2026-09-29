---
id: SPEC-JEV-SKILL-SUGGESTION-001
title: "plan — Jev skill-suggestion guidance skill"
version: "0.1.0"
created: 2026-09-30
author: manager-spec
---

# plan.md — SPEC-JEV-SKILL-SUGGESTION-001

## §A. Context

- Worktree: `.moai/worktrees/t1340` | Branch: `WT-jev-skill-suggest` | Base: develop @ `7a713a9a8`
- SPEC artifacts: `.moai/specs/SPEC-JEV-SKILL-SUGGESTION-001/` (spec/plan/acceptance/research/progress, Tier M + research.md)
- Scope: 스킬 2 사본 + 가드 테스트 1 파일 + catalog 엔트리 1 행. **비테스트 Go 변경 0.**
- PRESERVE (zero-edit 목록):
  - `internal/jevmeasure/gate_demo_test.go` (가드 — REQ-JEVG-003 이 byte-identity 를 요구)
  - `internal/cli/jev_question_design_skill_test.go` (형제 가드)
  - `internal/cli/mcp_jev.go`, `internal/cli/doctor_jev.go`, `internal/jev/**`, `internal/jevcred/**`, `internal/jevmeasure/**` (비테스트 Go 전부)
  - `.claude/skills/moai-ref-jev-question-design/**` 양 사본
  - `.moai/config/**`, `.claude/settings*.json` — 설정 변경 없음

## §B. Known Issues (축별 필터)

- **B2 cross-SPEC 정책 충돌 (본 카드의 최대 위험)** — 이 트리에는 소비자 출하를 금지하는
  살아 있는 가드(`TestNoConsumerCallPathShips`)와 복원 계약(REQ-JEVG-006)이 있다. 스킬
  본문이 "제안 기능을 쓸 수 있다"로 읽히면 REQ-JEVN-016(iii) 계열(사용 가능 기능으로
 presenting 금지) 위반이다. 본문은 게이트-오프 우선 + NO SIGNAL 원칙으로 시작하고, 호출
  경로 토큰 4종을 금지하는 새 가드 테스트가 회귀를 막는다. 가드 테스트·마커 목록을
  건드리는 것은 억제(suppression)로 금지.
- **B4 frontmatter 정본 스키마** — spec.md 12 필드, `created:`/`updated:`/`tags:` (snake_case 별칭 금지).
- **B6 spec-lint 헤딩 규약** — 배타 절은 `### Out of Scope — <topic>` H3 + `-` 불릿 (H2 단독은 MissingExclusions ERROR).
- **B8 작업 트리 위생** — `git add` 는 명시 pathspec 만. 런타임 관리 파일
  (`.moai/state/`, `.moai/logs/`) 편집 금지.
- **B10 PRESERVE 목록** — §A 열거 외 무접촉. 특히 타 카드 SPEC 디렉터리 무접촉.
- **B11 블로커 보고** — 질문이 필요하면 AskUserQuestion 대신 구조화 블로커 보고.

## §C. Pre-flight

```bash
git branch --show-current && git rev-parse --short HEAD   # WT-jev-skill-suggest 기대
go build ./...                                             # baseline green
GOOS=windows GOARCH=amd64 go build ./...                   # cross-platform baseline
ls .claude/skills/ | grep moai-jev-skill-suggestion        # 0 — RED-now 기준선 (plan 단계 관측됨)
grep -c "moai-jev-skill-suggestion" internal/template/catalog.yaml  # 0
golangci-lint run --timeout=2m ./internal/cli/... 2>&1 | tail -3    # baseline 대비 NEW 판별용
```

## §D. Constraints

- 커밋: Conventional Commits, `feat(SPEC-JEV-SKILL-SUGGESTION-001): M<N> ...`, 본문에 카드
  id(t1340) + `Authored-By-Agent:` 트레일러 + `🗿 MoAI` 종결. 레인은 push 하지 않는다 —
  병합은 리드 일괄(git-flow 레인 프로토콜).
- 금지: `--no-verify`, `--amend`, `go test ./...` 로컬 실행, 기존 가드 테스트 편집,
  스킬 본문에 SPEC ID/REQ 토큰/카드 id/날짜/SHA 삽입, 키 재질 인라인.
- 필수: 템플릿 원본 편집 후 로컬 사본 동기화(바이트 동일) + `make build`(카탈로그 해시
  재생성이 build 레시피에 내장됨 — 별도 대상 없음).
- 스킬 본문 언어: 영어 (자연어 정본형 중립성 C9). 본문 길이 목표 90-130 행.

## §E. Self-Verification (run-phase 인도물)

§E 항목은 VCI 5-절 형식(주장/증거/귀속/미검증/잔여위험)으로, (a) 명령 (b) verbatim 출력
(c) 트리 HEAD SHA (d) 각 명령의 exit code 를 함께 보고한다 — `grep -c` 0-적중은 exit 1
이므로 코드와 카운트를 쌍으로 기록한다 (plan-audit D4).

- E1: AC-JSK-001..011 PASS/FAIL 매트릭스 (acceptance.md §D)
- E2: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` 양쪽 exit 0
- E3: `go test ./internal/cli/ -run '^TestJevSkillSuggestionSkill(CarriesNoCallPath|CopiesStayIdentical)$'`
  (신설 가드 2테스트 — 앵커 선택자, 정확히 2건 sweep) — 양성 대조 포함 green;
  `go test ./internal/jevmeasure/ -run '^TestNoConsumerCallPathShips$'` green (기존 가드 무손상)
- E4: 양 사본 `cmp` 바이트 동일 + 중립성 grep 0 (`AC-JSK-003` 패턴)
- E5: `golangci-lint run ./internal/cli/...` — 신규 이슈 0 (baseline 대비)
- E6: 커밋 SHA 목록 + `git status --short` 청결 (§A PRESERVE 무접촉 포함)
- E7: 블로커 유무
- E8: RED 증거 — M1 에서 신설 가드 테스트가 스킬 파일 부재로 실패하는 verbatim 출력
  (존재 RED), M2 적용 후 동일 명령 green (RED→GREEN 쌍)

## §F. Milestones (TDD 순서)

- **M1 — 가드 테스트 선착 (RED)**: `internal/cli/jev_skill_suggestion_skill_test.go` 신설.
  형제 선례 복제: (1) `TestJevSkillSuggestionSkillCarriesNoCallPath` — 금지 토큰 4종
  (`internal/jev`, `mcp__moai__jev`, `jev_ask`, `moai jev`) 을 양 사본에서 부재 단언 +
  양성 대조 2건(`internal/cli/mcp_jev.go` → `internal/jev`,
  `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` → `jev_ask`); (2)
  `TestJevSkillSuggestionSkillCopiesStayIdentical` — 양 사본 바이트 동일. 이 시점 스킬
  파일이 없어 read-fail 로 RED — verbatim 출력 보존(E8).
- **M2 — 스킬 본문 + 미러 (GREEN)**: 템플릿 원본
  `internal/template/templates/.claude/skills/moai-jev-skill-suggestion/SKILL.md` 작성
  (spec.md §D 설계 + research.md §5 목차대로), 로컬 사본에 동일 내용 복제. M1 두 테스트
  green 확인. sentinel grep 표(acceptance.md §D.4-§D.7)가 전부 화면에 잡히는지 자가 확인.
- **M3 — 등록 + 빌드 + 검증 매트릭스**: `internal/template/catalog.yaml` 에 엔트리 추가
  (name/tier: core/path/hash/version — hash 는 `make build` 레시피의
  `gen-catalog-hashes.go --all` 이 재생성). `make build` → `go build ./...` +
  `GOOS=windows` → E1-E6 매트릭스 완성 → `go vet ./internal/cli/...`.

## §G. Anti-Patterns (금지)

- 가드를 통과하려고 스킬 본문의 위험 표현을 지우는 것(계약 약화) — 본문 내용은 spec.md
  §B REQ-JSK-004..007 이 소유하고, 가드는 토큰만 본다. 본문을 깎아 테스트를 통과하는
  것은 또 다른 억제다.
- "호출 예시" 추가 (셸 예제, MCP 도구명, import 경로) — REQ-JSK-002 위반이자 질문-디자인
  스킬이 세운 분할 파괴.
- 카탈로그 해시를 손으로 계산해 넣기 — `make build` 재생성을 우회하는 수동 값은 다음
  build 에 사라지며 drift 를 만든다.
- 스킬 본문에 "복원 예정" 류의 가용성 암시 — REQ-JEVN-016(iii) 정신 위반. 복원 산문은
  "기각/조건부 재검토" 어조로만.
- PRE-M1 baseline 을 M2 뒤에 재측정해 RED 를 흉내내기 — RED 는 M1 시점 파일 부재 상태에서
  관측된 것만 증거다.

## §H. Cross-references

- spec.md §F 추적 사슬 / acceptance.md §D AC 매트릭스 / research.md §3 전제 정정 전문
- `.claude/rules/moai/development/skill-authoring.md` — 스킬 프런트매터 스키마
- `.moai/docs/template-internal-isolation-doctrine.md` §25 — 중립성 클래스
