---
id: SPEC-DOCS-TABCOUNT-DRIFT-001
title: 인수 기준 — 설정 탭 수·이름 드리프트 차단
version: "0.1.0"
status: draft
created: 2026-09-12
updated: 2026-09-12
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "README.{md,ko,ja,zh}, docs-site/content/**, internal/web"
lifecycle: spec-anchored
tags: "docs, guard, acceptance, t530"
---

# 인수 기준

모든 항목은 **명령 + 기대 출력**이다. 산문 단정은 인수 근거가 아니다.
모든 명령은 워크트리 루트(`.claude/worktrees/t530`)에서 실행한다.

## §D AC 매트릭스

| AC | 검증 대상 REQ | 형태 |
|---|---|---|
| AC-TCD-001 | REQ-TCD-001 | 열거 산출물 존재·완전성 |
| AC-TCD-002 | REQ-TCD-002 | A군 4자리 제거 |
| AC-TCD-003 | REQ-TCD-003, 008 | 대상 12파일 전체에서 수 0건 |
| AC-TCD-004 | REQ-TCD-004, 006 | 가드 존재 + GREEN |
| AC-TCD-005 | REQ-TCD-005 | 오탐 3자리 침묵 |
| AC-TCD-006 | REQ-TCD-004 | 변이 시험 — 깨뜨리면 FAIL |
| AC-TCD-007 | REQ-TCD-007 | 4로케일 패리티 |
| AC-TCD-008 | REQ-TCD-006 | 이름 목록 == 렌더 라벨 |
| AC-TCD-009 | REQ-TCD-009 | 스크린샷 미변경 + 결정 기록 |
| AC-TCD-010 | — | docs-site 빌드 경고 0 |

---

### AC-TCD-001 — 열거 산출물

**Given** 카드 t530 의 plan-phase 가 끝났을 때,
**When** `test -f .moai/reports/t530/tab-count-sites.md && grep -c '^| [ABCD][0-9]' .moai/reports/t530/tab-count-sites.md` 를 실행하면,
**Then** exit 0 이고 출력이 `28` 이상이다 (A군 4 + B군 8 + C군 8 + D군 8 = 28행).

### AC-TCD-002 — A군: 틀린 수 4자리 제거

**Given** M2 가 끝난 트리에서,
**When**
```bash
grep -nE '(9|nine|九|아홉)[^.。]{0,4}(tabs?|개 탭|タブ|个标签页)' \
  docs-site/content/ko/cli-reference/web.md docs-site/content/en/cli-reference/web.md \
  docs-site/content/ja/cli-reference/web.md docs-site/content/zh/cli-reference/web.md; echo "rc=$?"
```
를 실행하면,
**Then** 출력이 없고 `rc=1` 이다 (grep 무매치).

### AC-TCD-003 — 대상 12파일에서 손으로 적힌 탭 수 0건

**Given** M3 가 끝난 트리에서,
**When**
```bash
grep -rnE '([0-9]+|nine|ten|eleven|twelve|thirteen|fourteen|fifteen|九|十四|열네|아홉)[^.。]{0,4}(tabs?|개 탭|-탭|タブ|个标签页|标签页)' \
  README.md README.ko.md README.ja.md README.zh.md \
  docs-site/content/{ko,en,ja,zh}/cli-reference/web.md \
  docs-site/content/{ko,en,ja,zh}/advanced/moai-web-console.md; echo "rc=$?"
```
를 실행하면,
**Then** 출력이 없고 `rc=1` 이다.

> 기대값 `0건` 은 `plan.md §B` 의 자리별 처분이 20자리 전부 `removable` 로 확정된 데서 나온다
> (`must-stay` 자리 없음 ⇒ 허용 목록 없음). 이 트리에서 같은 명령을 지금 돌리면 **20행**이 나온다 —
> 즉 이 AC 는 현재 붉고, M3 가 그것을 0 으로 만든다.

### AC-TCD-004 — 가드 존재 + GREEN

**Given** M5 가 끝난 트리에서,
**When** `go test ./internal/web/ -run 'TestDocsTabContract' -v 2>&1 | tail -20` 을 실행하면,
**Then** `--- PASS: TestDocsTabContract` 가 출력에 있고 `ok  github.com/modu-ai/moai-adk/internal/web` 로 끝난다.
**And** 선택자가 조용히 0개를 고르지 않았음을 같은 출력의 `--- PASS` 행 존재로 확인한다
(`-run` 은 없는 이름에 대해 아무것도 고르지 않고도 `ok` 를 낸다).

### AC-TCD-005 — 오탐 침묵

**Given** 가드가 GREEN 인 트리에서,
**When** 아래 세 자리가 그대로 있음을 확인하고 가드를 다시 돌리면,
```bash
grep -c 'nine' README.md                                             # >= 1 (SVG 형태 수)
grep -c 'Eleven ref skills' README.md                                # 1
grep -c '4 タブ' docs-site/content/ja/claude-code/extensibility/plugins.md  # 1
go test ./internal/web/ -run 'TestDocsTabContract'                   # ok
```
**Then** 세 grep 이 모두 1 이상을 출력하고 가드는 여전히 `ok` 다 — 즉 이 자리들을 잡지 않는다.

### AC-TCD-006 — 변이 시험 (공허한 초록 차단)

**Given** 가드가 GREEN 인 트리에서,
**When** README.md 의 탭 이름 하나를 일부러 틀리게 바꾼 뒤(`Audit` → `Audits`) 가드를 돌리고, 되돌린 뒤 다시 돌리면,
**Then** 첫 실행은 `FAIL` 이고 실패 메시지가 **어느 파일의 어느 이름**이 어긋났는지 적시하며, 되돌린 뒤 실행은 `ok` 다.
**And** 두 출력이 `progress.md` §E.2 에 축어로 남는다.

### AC-TCD-007 — 4로케일 패리티

**Given** M3/M4 가 끝난 트리에서,
**When**
```bash
git diff --name-only origin/develop... -- docs-site/content | sed -E 's#docs-site/content/[a-z]+/##' | sort -u
git diff --name-only origin/develop... -- docs-site/content | wc -l
```
를 실행하면,
**Then** 첫 명령이 낸 상대 경로마다 ko/en/ja/zh 4본이 모두 변경 목록에 있다
(상대 경로 종류 수 × 4 == 둘째 명령의 수).

### AC-TCD-008 — 이름 목록이 렌더 라벨과 일치

**Given** M4 가 끝난 트리에서,
**When** `go test ./internal/web/ -run 'TestDocsTabContract/names' -v 2>&1 | grep -E '^(---|=== RUN)'` 를 실행하면,
**Then** `--- PASS` 가 나오고 FAIL 이 없다.
**And** 참조 정본이 `consoleTabs()` 순서임을 같은 테스트가 순서까지 단정한다 (집합 비교가 아니다).

### AC-TCD-009 — 스크린샷 미변경 + 결정 기록

**Given** 카드가 마감될 때,
**When**
```bash
git diff --name-only origin/develop... -- assets/images/; echo "rc=$?"
grep -c '## 5. 스크린샷 범위 판단' .moai/specs/SPEC-DOCS-TABCOUNT-DRIFT-001/spec.md
```
를 실행하면,
**Then** 첫 명령의 출력이 비어 있고(이미지 무변경), 둘째가 `1` 이다(결정과 근거가 기록됨).

### AC-TCD-010 — docs-site 빌드 경고 0

**Given** 문서 수정이 끝난 트리에서,
**When** `cd docs-site && hugo --gc --minify 2>&1 | grep -iE 'warn|error'; echo "rc=$?"` 를 실행하면,
**Then** 출력이 없고 `rc=1` 이다.

---

## §D.1 엣지 케이스

- **가드가 문서를 못 읽는 경우**: 파일 부재·경로 오류는 `t.Skip` 이 아니라 `t.Fatal` 이어야 한다. skip 은 판정 전 이탈이고, 초록으로 보인다.
- **이름 추출 개수 불일치**: 파서가 14개가 아닌 수를 뽑으면 그 자체로 FAIL — 0개를 뽑고 "비교할 것이 없다" 며 통과하는 경로를 막는다.
- **로케일별 수사 표기**: `十四` / `열네` / `fourteen` 이 각각 정규식에 들어 있는지 테스트 안의 표로 고정한다.

## §D.2 Definition of Done

1. AC-TCD-001 ~ 010 전부 PASS (명령 출력이 `progress.md` §E.2 에 축어로 남는다).
2. `go test ./internal/web/...` 전체 통과 — 새 가드가 기존 테스트를 깨지 않았다.
3. `.moai/reports/t530/tab-count-sites.md` 가 최종 트리 상태와 일치한다(수정된 행 번호 반영).
4. D군 12자리가 렌더 라벨(`GLM Settings` 계열)로 정합되었고, N2 가드가 그 일치를 단정한다 (§C 결정, 2026-09-12).
5. 후속 카드 2건(스크린샷 절차·재촬영)이 리드에게 카드 요청으로 전달됐다.
6. frontmatter `status` 가 소유 에이전트에 의해 전이됐다(본 에이전트는 `draft` 까지만 쓴다).
