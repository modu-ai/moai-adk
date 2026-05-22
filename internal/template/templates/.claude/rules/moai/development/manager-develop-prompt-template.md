---
description: "manager-develop 위임 Prompt Template — Tier M/L SPEC run-phase 5-section 표준. SPEC 위임 작성 시에만 로드."
paths: ".moai/specs/**,.claude/agents/moai/manager-develop.md,.claude/skills/moai/workflows/run.md"
---

# manager-develop 위임 Prompt Template

> [ZONE:Evolvable] [HARD] 모든 `manager-develop` subagent 위임 prompt는 본 템플릿의 5개 섹션 (Context / Known Issues / Pre-flight / Constraints / Self-Verification Deliverables)을 포함해야 한다. 누락 시 재위임 반복 위험 증가.

본 rule은 W3 HARNESS-AUTONOMY-001 메타-분석 결과 (2026-05-20)에서 도출된 위임 품질 개선 사항을 표준화한다. 1-pass 위임으로 결함 사전 차단 목표.

## 1. 표준 위임 Prompt 5-Section 구조

### Section A — Context (위치 + 분기 + SPEC 산출물 경로)

명시 의무:
- 작업 위치 (project root absolute path)
- 현재 branch + HEAD SHA (manager가 추가 commit을 어디에 쌓을지 명확화)
- SPEC 산출물 경로 (`.moai/specs/SPEC-XXX/{spec,plan,acceptance,progress}.md`) + 라인 카운트
- plan-auditor verdict (PASS score + 재실행 권장 여부)
- 기존 인프라 (PRESERVE 대상 + EXTEND 대상)

### Section B — Known Issues 자동 주입 (가장 중요)

[ZONE:Evolvable] [HARD] 다음 8 카테고리의 known issues는 위임 prompt에 자동 포함되어야 한다. 누락 = 재위임 위험.

**B1. Cross-platform Build Tags**
- syscall 패키지 사용 시 build tag 강제 (lessons #21 W0 fix 패턴)
- 권장: `//go:build !windows` + `//go:build windows` 파일 분리
- 검증: `GOOS=windows GOARCH=amd64 go build ./...` 통과 의무

**B2. Cross-SPEC 정책 충돌 사전 스캔**
- 영향 받는 패키지의 retired/superseded SPEC 확인 (예: W3 ↔ V3R4 HARNESS retirement)
- `grep -r "Retired\|TestHarnessRetirement\|deprecation-marker" internal/<pkg>` 실행
- 충돌 발견 시: SPEC 본문에서 reversal 명시 또는 새 SPEC scope 정의

**B3. C-HRA-008 / Subagent Boundary Discipline**
- `internal/harness/`, `internal/hook/` 등 subagent 도메인 코드에 AskUserQuestion 호출 금지
- 검증: `grep -rn 'AskUserQuestion\|mcp__askuser' <pkg> | grep -v "_test.go" | grep -v "// "` 가 0 매치
- CI guard test 의무: `<pkg>/subagent_boundary_test.go`

**B4. Frontmatter Canonical Schema**
- `created:`/`updated:`/`tags:` 사용 (snake_case alias 금지)
- 참조: `.claude/rules/moai/development/spec-frontmatter-schema.md`

**B5. CI 3-tier 인지**
- spec-lint, golangci-lint, Test (per OS) 각각 별도 fail 가능
- W1/W2 chicken-and-egg 패턴 vs NEW 결함 구분
- 참조: lessons #19

**B6. spec-lint Heading 규약**
- `## Out of Scope` (h2) 만으로는 `MissingExclusions` ERROR
- `### <X.Y> Out of Scope` (h3) sub-section 필요 (W3 발견)

**B7. observer.go / capture path resolution**
- `input.CWD` empty → `os.Getwd()` fallback 시 working dir 누수 (`internal/hook/.moai/` anomaly 원인)
- 권장: `$CLAUDE_PROJECT_DIR` 우선 사용

**B8. Working Tree Hygiene**
- runtime-managed files (`.moai/harness/usage-log.jsonl`, `.moai/state/`) 변경 금지
- session_end의 `cleanupBogusRootDir` 의존 (`{}/`  literal directory는 cleanup 대상)
- 무관 untracked files은 commit 포함 금지 (`git add` specific path만)

### Section C — Pre-flight Check List (착수 전 의무 검증)

위임 받은 manager-develop가 코드 변경 전 실행:

```bash
# 1. 현재 branch + baseline 확인
git branch --show-current
git rev-parse HEAD

# 2. Cross-platform build 가능성 사전 확인
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. 기존 lint baseline 측정 (NEW vs pre-existing 구분 위해)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. PRESERVE 대상 파일 list 출력
ls <PRESERVE_GLOB>

# 5. 영향 패키지 retired/superseded SPEC 확인
grep -r "Retired\|TestHarnessRetirement\|superseded" internal/<target_pkg> || echo "no conflicts"
```

### Section D — Constraints (DO NOT VIOLATE)

각 위임 prompt에 explicit list:
- PRESERVE 대상 파일 enumeration (Brownfield strategy 적용 시)
- 무관 untracked/modified 파일 list (변경 금지)
- 금지 명령 (`--no-verify`, `--amend`, force-push to main, …)
- 사용 의무 명령 (Conventional Commits, `🗿 MoAI` trailer, …)
- C-HRA-008 같은 binary constraint (grep 0 매치)

### Section E — Self-Verification Deliverables

manager-develop이 완료 보고 시 자체 검증한 다음 항목을 강제 포함:

**E1. AC Binary PASS/FAIL Matrix**
| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-XXX-001 | PASS | `go test -run TestX ./pkg` | `PASS — ok  pkg 0.5s` |

**E2. Cross-Platform Build 결과**
```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

**E3. Coverage 측정 (≥85% threshold per package)**
```
$ go test -cover ./internal/<pkg>/...
```

**E4. Subagent Boundary Grep (C-HRA-008 류)**
```
$ grep -rn 'AskUserQuestion' <pkg> | grep -v "_test.go" | grep -v "// "
(no output expected)
```

**E5. Lint Status (NEW vs baseline 구분)**
```
$ golangci-lint run --timeout=2m
# NEW issues 발견 시 explicit report; pre-existing baseline은 별도 mark
```

**E6. Branch HEAD + Push 상태**
- 새 commits SHA 리스트
- `git push origin <branch>` 결과

**E7. Blocker Report (있을 시)**
- 위임 prompt에서 명시 안 된 사용자 결정 필요 항목 발견 시 structured 보고 (AskUserQuestion 절대 호출 금지)

## 2. 위임 Prompt 작성 Workflow (오케스트레이터 입장)

```
1. Section A 구성 (SPEC 산출물 경로 + 현재 git 상태)
2. Section B 자동 주입 (lessons memory에서 keyword 매칭으로 8 카테고리 select)
3. Section C 표준 pre-flight checks (위 그대로 복사)
4. Section D constraints (SPEC + PRESERVE list + working tree 상태에서 추출)
5. Section E deliverables (위 그대로 복사)
```

## 3. Anti-Patterns

본 템플릿 미준수 케이스 — 재위임 위험 증가:

- Section B 누락 → cross-platform build / cross-SPEC 충돌 사후 발견 (W3 케이스)
- Section E 누락 → orchestrator가 직렬 검증 (W3에서 ~10분 추가 손실)
- "implement the SPEC" 같은 1-liner 위임 → manager-develop이 prompt 부족으로 가정 누락
- PRESERVE 대상 enumeration 누락 → 의도치 않은 파일 수정

## 4. 관련 Layer (메타-분석 §3)

본 rule은 Layer A (위임 Prompt 품질 향상)의 표준화. 후속 Layer:

- Layer B (병렬 위임 — Agent Teams): 별도 rule 필요
- Layer C (Background CI watch): `gh pr checks --watch` 패턴 표준화
- Layer D (검증 병렬화): orchestrator self-discipline
- Layer F (lessons 자동 capture): SubagentStop hook 확장

## 5. 검증 (본 rule 자체 적용 결과)

다음 SPEC (W4 PROJECT-MEGA-001) 위임에서:
- 1-pass 성공률 측정 (목표: ≥80%, W3에서 33%)
- 재위임 횟수 (목표: ≤1, W3에서 3)
- 총 wall-time (목표: ≤30분, W3에서 91분)

---

Version: 1.0.0
Origin: W3 HARNESS-AUTONOMY-001 메타-분석 (2026-05-20)
Status: Active — applies to all manager-develop delegations
