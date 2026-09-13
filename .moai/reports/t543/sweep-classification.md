# t543 스윕 분류 — 리터럴 base SHA 범위 판정식 22줄

기준: 리드 판정(2026-09-10) — **흡수 뒤에 평가되는 줄만 고친다.** 흡수를 이미 문구에서 다루거나, 다시 평가될 일이 없는 줄은 "해당 없음"으로 둔다.

인벤토리 명령: develop 의 `.moai/specs/*/acceptance.md`·`plan.md` 에서 `git diff --name-only <7~40자리 16진수>..HEAD` 형태를 찾는 grep. 결과는 `repro/sweep-inventory.txt`(22줄, SPEC 10개)다. SPEC 상태는 develop 의 각 `spec.md` frontmatter `status:` 를 이번 실행에서 읽었다.

| SPEC | 줄 | 상태 | 분류 | 이유 |
|---|---|---|---|---|
| SPEC-BINLAG-KEYGUARD-001 | acceptance.md:52, :331 | completed | 해당 없음(종결) | 종결된 SPEC 이라 흡수 뒤에 다시 평가될 일이 없다 |
| SPEC-CODEX-COVER-RESIDUAL-001 | acceptance.md:116 | completed | 해당 없음(종결) | 같음 |
| SPEC-CODEX-SKILL-DISABLE-001 | acceptance.md:120 | completed | 해당 없음(종결) | 같음. 줄 자체가 base 를 "develop 과의 merge-base" 로 설명한다 |
| SPEC-CODEX-STALE-SPLIT-FOURTH-001 | plan.md:76 | completed | 해당 없음(종결) | 같음 |
| SPEC-DOCS-LOCALE-PARITY-REPAIR-001 | acceptance.md:25, :26, :28, :39, :120, :122 | completed | 해당 없음(종결) | 같음 |
| SPEC-FACTORY-MODE-001 | acceptance.md:409, :436, plan.md:168 | completed | 해당 없음(종결) | 같음. plan.md:168 도 범위 판정식(`git diff --name-only 7171880a9..HEAD -- internal/`)을 제외 근거로 쓰지만, 종결된 SPEC 이다 |
| SPEC-FMT-GATE-001 | acceptance.md:51 | completed | 해당 없음(흡수를 이미 다룸) | 51행이 develop 흡수 뒤 10경로가 잡힌다는 사실을 적고, 판정식을 카드 커밋 귀속 형태로 이미 개정했다 |
| SPEC-SPEC-LINT-ID-ARG-001 | acceptance.md:49 | completed | 해당 없음(종결) | 같음. 줄 자체가 이 명령을 "옛 명령"으로 적고, 형제 카드의 착지로 거짓 FAIL 이 난다는 결함을 이미 기록한다 |
| SPEC-UPDATE-DATA-SURVIVAL-001 | acceptance.md:74, :927, plan.md:27 | completed | 해당 없음(종결) | 같음 |
| **SPEC-UPDATE-DOC-DRIFT-001** | **acceptance.md:659, :660, plan.md:54** | **draft** | **수정 대상** | 아직 run 을 거치지 않았으므로 레인 절차대로 진행하면 흡수 뒤 병합 트리에서 평가된다. `7f61332ef..HEAD` 는 흡수 순간 다른 카드의 변경을 끌어들인다 |

- 수정 대상: 1개 SPEC, 3줄. SPEC 본문은 manager-spec 소관이므로 그 경로로 고친다.
- 인벤토리의 한계: grep 은 `git diff --name-only <SHA>..HEAD` 형태만 잡는다. `git log <SHA>..HEAD`, `rev-list`, 변수에 담은 SHA 같은 다른 형태의 범위 판정식은 이 인벤토리에 없다.

## 넓힌 인벤토리 (같은 날, SPEC 수리 워커가 형태 밖의 줄을 찾은 뒤)

SPEC 수리 중 `SPEC-UPDATE-DOC-DRIFT-001` 의 AC-UDD-021(`acceptance.md:610-611`)이 같은 결함을 `git diff --stat` · `git log` 형태로 갖고 있는데 위 인벤토리가 놓쳤다는 사실이 드러났다. 그래서 형태를 넓혀 다시 쟀다.

- 명령: develop 의 `.moai/specs/*/acceptance.md`·`plan.md` 에서 `git diff` · `git log` · `git rev-list`(옵션 포함) 뒤에 `<7~40자리 16진수>..HEAD` 가 오는 줄을 찾는 grep. 결과 `repro/sweep-inventory-wide.txt`.
- 결과: **80줄, SPEC 27개.** 형태별로 `diff --name-only` 22 · `diff --stat` 17 · 기타 `diff` 30 · `log` 10 · `rev-list` 1.
- 새로 잡힌 SPEC 17개의 `status:` 를 develop 에서 읽었다. **17개 모두 completed** 다. 따라서 흡수 뒤 평가 대상이라는 분류는 바뀌지 않는다 — 여전히 draft 인 `SPEC-UPDATE-DOC-DRIFT-001` 한 곳뿐이다.
- 그 SPEC 안에서 넓힌 인벤토리가 잡은 줄은 5개다: AC-UDD-021 의 `acceptance.md:610`(`diff --stat`)과 `:611`(`log --oneline`), 이미 고친 AC-UDD-023 의 `:659`·`:660`, plan.md `:54`. **수정 대상에 AC-UDD-021 두 줄을 더한다.**
- 넓힌 grep 에도 한계가 남는다: 범위의 오른쪽 끝이 `HEAD` 가 아닌 경우, SHA 를 변수에 담아 쓰는 경우, `...` 세 점 범위는 잡지 않는다.
