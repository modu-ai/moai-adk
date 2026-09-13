# t522 — lane verdict (.gitignore .agents 무시 범위 — 정책 판정, 코드 변경 없음)

card: t522 (Class C, Tier S~M)
worktree: .claude/worktrees/t522
branch: WT-agents-ignore-policy
base: 7e1859c28 (판정 시점 로컬 develop; 이후 병합 16f3b8a81도 .gitignore 불변이라 판정 유효)
measured by: lane-8 orchestrator, directly

## Claim

1. **카드 전제는 스테일이다.** 카드 문구 "이 저장소 .gitignore:133은 '.agents/'를 통째로
   무시한다"는 발행일(2026-09-07) 기준으로는 참이지만, 판정 시점 트리에서는 거짓이다.
   현행 루트 .gitignore:133-174는 `.agents/*`로 시작해 16개 published SKILL.md만
   재포함하는 정밀 정책이다.
2. **정책 판정 — 배포 템플릿 정책이 옳고, 루트도 이미 그에 정렬돼 있다.** 두 표면 모두
   "published 16 = 트랙, 미러/미등록 = 무시"로 일치한다. 배포 템플릿 변경이 필요하지
   않으므로 운영자 게이트는 트리거되지 않는다. "무시를 풀어 파일이 나타나게 하지 마라"
   조건도 공허하다 — 이미 16파일이 트랙 중이다.
3. **배포 측 정책은 가드로 봉인돼 있다.** `internal/template/gitignore_agents_mirror_test.go`
   의 TestGitignore_IgnoresSkillMirrorOnly(AC-CSC-015)가 배포 .gitignore에 좁은 미러
   패턴(`.agents/skills/moai*`) 존재와 루트 통째 무시 패턴 부재를 단언한다.

## Evidence

**(1) 루트 정책 인용** (`.gitignore:133-135` 및 136-137, 171-173 주석):

```
133  .agents/*
134  !.agents/skills/
135  .agents/skills/*
136  # Publish only the generated Codex command-skill entrypoints. Keep every other
137  # local .agents artifact, unlisted skill, and per-skill sidecar ignored.
138-153  !.agents/skills/moai-{clean,codemaps,e2e,feedback,fix,gate,goal,harness,loop,mx,plan,project,review,run,sync,todo}/
154  .agents/skills/*/*
155-170  !.agents/skills/moai-<동일 16명>/SKILL.md
171  # Template-distributed Codex command-skill publications ARE committed
172  # (SPEC-CODEX-COMMAND-SKILLS-001); re-include the template subtree excluded
173  # by the bare .agents/ rule above.
174  !internal/template/templates/.agents/
```

**(2) 루트 .agents 트랙 실측:**

```
$ git ls-files .agents/ | wc -l
16
(16파일 = .agents/skills/moai-<command>/SKILL.md, command 16종)
```

**(3) 귀속** (git pickaxe·log 실측, 이번 런):

```
$ git log --format='%h %ad %s' --date=short -S'.agents/*' -- .gitignore
1cfc6f544 2026-09-10 fix(ci): resolve review and release gate regressions

$ git show --stat --format='%h %ad %s' --date=short e7d2a1658 -- .gitignore .agents/
e7d2a1658 2026-09-07 feat(t503): command-to-skill publication emitter core (SPEC-CODEX-COMMAND-SKILLS-001 M1)
 .gitignore | 4 ++++

$ git log --format='%h %ad %s' --date=short --diff-filter=A -- .agents/skills/moai-clean/SKILL.md
1cfc6f544 2026-09-10
```

- 정밀 정책(`.agents/*`) 도입 + 루트 16파일 최초 커밋: **1cfc6f544 (2026-09-10)**
- 선행: **e7d2a1658 (2026-09-07, t503 M1)** — .gitignore +4줄(템플릿 서브트리 재포함 계열)
- 배포 측 가드 계보: **9a8a99667 (2026-08-31, SPEC-GITIGNORE-ROOT-GUARD-001, t377)**

→ 카드 전제는 **09-10 커밋에서 무효화**됐다(재개일 09-13 기준 3일 전). 카드가 인용한
"gitignore:133"의 행 번호도 현재 트리에서는 다른 내용을 가리킨다.

**(4) 배포 템플릿 정책 인용** (`internal/template/templates/.gitignore:204, 211-226`):

```
204  .agents/skills/moai*
211  !.agents/skills/moai-clean/  … 16개 디렉터리 재포함
주석 202-203: "Only the generated entries are ignored. The .agents/ root itself is NOT:
entries you create there, and source files placed there later, stay tracked."
```

## Baseline-attribution

모든 명령과 출력은 이번 런, 워크트리 `.claude/worktrees/t522`(WT-agents-ignore-policy @
7e1859c28)에서 수집. 이후 로컬 develop 병합(16f3b8a81)의 diff는 6파일이며 .gitignore를
포함하지 않아 본 판정은 착지 트리에도 그대로 유효하다.

## Gaps

- 루트 .gitignore의 133-174 외 섹션과 배포 .gitignore의 나머지 섹션 전체의 상호 대조는
  카드의 정책 질문 범위 밖으로 판단해 수행하지 않았다.
- `gitignore_agents_mirror_test.go`는 도입부(1-60행, AC-CSC-015 단언)만 판독했다.

## Residual-risk

- 루트 .gitignore:171-174의 주석은 더 이상 존재하지 않는 "bare .agents/ 규칙"을 참조하고,
  `.agents/*`가 내부 슬래시를 가져 루트 앵커인 이상 `!internal/template/templates/.agents/`
  재포함은 no-op다. 무해하나 스테일 주석 — 별도 정리 카드 후보.
- 루트는 비스킬 .agents 콘텐츠를 `.agents/*`로 숨기고 배포판은 숨기지 않는 미세 차이가
  남는다. 개발 저장소 소음 통제로 판단되며, 카드의 질문(출판 정책 어느 쪽이 옳은가)에는
  영향이 없다.
- 카드 처분(철회 vs landed 기록)은 운영자 판정 사항 — 본 판정서는 근거만 제공한다.
