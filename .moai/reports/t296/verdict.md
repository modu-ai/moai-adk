# t296 판정서 — '16-language neutrality contract' 본문 부재

- 판정: **FIXED** — 카드 전제 확인. `coding-standards.md`의 Language Policy 섹션은 "지시 문서는 영어"라는 전혀 다른 주제만 담고 있었고(grep "16"·언어 목록 0건 — 리드 2026-08-27 실측과 일치), 인용 파일들이 본문 없는 계약을 가리키고 있었다.

## 판정: 본문을 인용 지점에 직접 배치 (옵션 1)

인용 문구 전부가 `coding-standards.md § Language Policy (16-language neutrality contract)` 형태를 취하므로(아래 인용 스윕), 본문을 그 섹션 안에 서브섹션으로 배치하면 **인용 파일 무편집**으로 전부 해소된다. 신규 파일 생성+인용 14곳 재지향(옵션 2)보다 최소 변경이며, 배포 트리 도달성도 같은 위치에서 즉시 성립.

## 실측된 인용 표면 (develop `2660bcd09`, `git grep -c "16-language neutrality contract"`)

라이브 7파일(+skill-authoring 2히트): `.claude/rules/moai/development/skill-authoring.md` · `.claude/skills/moai-workflow-loop/SKILL.md` · `.../references/examples.md` · `.../references/reference.md` · `.claude/skills/moai/workflows/loop.md` · `.claude/skills/moai/workflows/project/doc-generation.md` · `.moai/docs/generic-patterns-guide.md` — 각각의 템플릿 미러 7파일(`internal/template/templates/` 하위 동일 경로). 카드의 "10곳"은 2026-08-27 시점 측정으로 현재 트리와 어긋남 — 본 판정서는 본 런 측정(이름 나열)을 따름. 미수정(역사 기록): `.moai/reports/t332/*`, `.moai/specs/*` 4건, 구 리포트 1건.

## 변경 (이름 나열)

- `internal/template/templates/.claude/rules/moai/development/coding-standards.md` — Language Policy 섹션에 `### Programming-Language Neutrality Contract (16 languages)` 서브섹션 추가: 16 언어 목록(동등 나열, Dart 캐논명 flutter), PRIMARY 없음·planned 격하 없음, 프로젝트 마커 기반 감지, 두 축(16 프로그래밍 언어=중립성 / 4 로케일=번역) 구분 주의문.
- `.claude/rules/moai/development/coding-standards.md` — 동일 블록(바이트 동일).
- 인용 14곳: 무편집(본문이 인용 지점에 착지했으므로).

## 5-섹션

**Claim**: 배포 템플릿만 있는 트리에서 임의 인용 경로를 따라가면 계약 본문(16 언어 동등 규율 + 두 축 주의문)에 도달한다.

**Evidence**:
- 인용 경로 유효성: `grep -l "coding-standards.md"` 대상 7파일 → `7` (전부 파일 경로 명시).
- 배포 트리 도달: `grep -c "Programming-Language Neutrality Contract" internal/template/templates/.claude/rules/moai/development/coding-standards.md` → `1`.
- `make build` → 임베드 재컴파일 성공; `go test ./internal/template/` → `ok ... 29.682s` (mirror parity 테스트 포함 GREEN).
- 추가 블록 양쪽 바이트 동일: `diff <(sed -n ...) ...` → `BLOCK_IDENTICAL`.
- 중립성: 본문 블록에 SPEC-ID/카드id/날짜 매치 0.

**Baseline-attribution**: `WT-lang-policy-body` 워크트리, develop `2660bcd09`, 2026-09-02 본 런.

**Gaps**: "배포 트리만 있는 상태"의 완전 시뮬레이션(클린 체크아웃에서 moai init 후 에이전트가 인용을 따라가는 E2E)은 미수행 — 본문이 템플릿 파일 자체에 있으므로 grep으로 대체. `.moai/docs/generic-patterns-guide.md`는 로컬/템플릿 양쪽에 존재하는데 그 인용은 이제 양쪽 모두 본문에 도달.

**Residual-risk**:
- 템플릿↔로컬 `coding-standards.md`에 **사전 기존** 1행 diff 존재(로컬에만 `git commit --no-verify` 금지 불릿 — SPEC-PRETOOL-GATE-MOVE-001). 본 카드 범위 밖이라 건드리지 않았고, 추가 블록은 양쪽 동일. mirror parity 테스트는 GREEN — 해당 테스트가 이 파일의 완전 바이트 동일성을 요구하지 않는 것으로 판독됨(테스트 통과가 직접 증거).
- 16 언어 목록은 이 파일에 인라인 복제됨 — 목록이 바뀌면 이 파일도 갱신 필요(단일 목록 SSOT가 트리에 없어 인라인 선택; 중복 드리프트는 후속 카드 후보).
