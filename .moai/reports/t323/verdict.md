# t323 verdict — 카탈로그 무결성 해시가 스킬 디렉터리의 sub-files 를 관측하지 않는다

카드: 스킬 디렉터리의 SKILL.md 아닌 파일들이 카탈로그 무결성 해시에 안 보인다 — 배포되는데 검증 밖.
트리: `WT-catalog-tree-hash` @ develop `2660bcd09`, darwin.

## 판정 (카드의 세 질문에 대한 답)

- **(b) SKILL.md 만이 의도였는가** — 그렇다. `gen-catalog-hashes.go` 헤더 주석이 "hash only the
  root SKILL.md or skill.md (not sub-files)" 로, `catalog_hash_norm.go` 주석이 "NOT included in
  **v1** hash" 로 명시적 v1 축소였다. 그러나 그 축소의 대가를 담보하는 층은 존재하지 않는다:
  manifest 는 배포 시점 기록(원본↔배포본)이지 소스 무결성 검증이 아니고, mirror parity 테스트는
  rules/agents 등 특정 쌍만 본다. sub-files 의 소스-카탈로그 일치를 담보하는 곳은 **아무데도 없었다**.
- **(a) 확장한다** — 주장(카탈로그 무결성)과 실체(배포 경계 = 디렉터리 전체)를 일치시킨다.
  `//go:embed all:templates` 는 디렉터리 전체를 배포하므로 무결성 주장도 디렉터리 전체여야 한다.
  t317 계열이 지적한 "생성물·배포물과 검증 층의 동기화 부재"의 이 인스턴스를 닫는다.
- **(c) 비용** — 0 이었다. `make build` 가 이미 `gen-catalog-hashes --all` 을 자동 실행하므로
  개발자 워크플로 변화 없음. 해시 계산 비용은 274 파일 수준으로 미미 (전 패키지 테스트 25s 내 흡수).
  효과: sub-file 변경이 이제 catalog.yaml diff 로 관측된다 (t316 이 원했던 관측점의 확장).

## 나열 (카드 [HARD] 지시 — 세는 문장이 아니라 나열)

해시 관측 밖이던 파일 전체: `.moai/reports/t323/unhashed-before-t323.txt` — **274 개**,
카탈로그 45 항목 중 34 개 디렉터리 항목에서 수집. 성격 분류: `modules/ 83 · references/ 65 ·
scripts/ 48 · workflows/ 41 · templates/ 9 · schemas/ 2` (확장자: md 207 · svg 42 · json 8 ·
mustache 6 · gitkeep 5 · sh 3 · mjs 3). 전부 `//go:embed all:templates` 로 배포되는데 v1 해시는
루트 SKILL.md 하나만 봤다.

## Claim

1. 수리 전 트리에서 34 개 디렉터리 항목 전부가 전체-트리 집계 해시와 불일치했다 (RED — 결함 상태).
2. 수리 후 카탈로그가 전체 트리 집계와 일치하고, 검증자가 같은 단일 구현으로 재계산한다 (GREEN).
3. 생성자(스크립트, os.DirFS)와 검증자(테스트, 임베드 FS)가 **같은 함수**를 쓴다 — 쌍 드리프트의
   구조적 제거. 부수로 "byte-identical 로컬 복사" 규율(normalizeForHash)도 소멸시켰다.

## Evidence

| # | 명령 | 관측 출력 |
|---|------|----------|
| E1 (RED) | `go test ./internal/template/ -run TestCatalogHashCoversSkillSubfiles` (재생성 전) | 34 개 `CATALOG_HASH_SKINNY` — 예: `moai-ref-seo hash=39d3ad2b… whole-tree 64b86fae…`; `audited 34 directory entries`; FAIL |
| E2 | `go vet ./internal/template/scripts/ && go vet ./internal/template/` | 둘 다 OK — internal-package import 유효 (gopls 의 `not allowed` 진단은 워크트리 워크스페이스 거짓) |
| E3 | `go run ./internal/template/scripts/gen-catalog-hashes.go --all --dry-run` | 45 항목 — 디렉터리 34 개 `(whole tree)`, 파일 항목 단일; `moai-ref-seo 64b86fae…` 로 스크립트 값과 테스트 임베드-FS 값 교차 일치 |
| E4 | `go run … --all` 후 `git diff --stat catalog.yaml` | `34 insertions(+), 34 deletions(-)` — 34 개 디렉터리 항목의 hash 필드만 |
| E5 (GREEN) | `go test ./internal/template/ -run "TestCatalogHash" -count=1` → `ok 0.431s` |
| E6 (회귀) | `go test ./internal/template/ -count=1` | `ok github.com/modu-ai/moai-adk/internal/template 25.023s` |

## 구현 요약

- `internal/template/catalog_tree_hash.go` (신규): `ComputeDirTreeHash(fsys fs.FS, dir string)` —
  fs.WalkDir 로 정규 파일 전수 → NormalizeForHash → sha256 → `(relpath:digest)` 정렬 결합 → 최종
  sha256. 결정적(경로 정렬), 임베드 FS 와 os.DirFS 양쪽에서 동일.
- `internal/template/scripts/gen-catalog-hashes.go`: 디렉터리 항목이 ComputeDirTreeHash(os.DirFS)
  사용; 로컬 normalizeForHash 복사 제거 → template.NormalizeForHash 직접 사용 ("avoid import
  cycles" 주석의 근거가 거짓임을 E2 로 확인 후).
- `internal/template/catalog_tier_audit_test.go`: `TestManifestHashFormat` 의 디렉터리 분기가
  ComputeDirTreeHash 로 재계산 (생성자와 동일 함수); 신규 `TestCatalogHashCoversSkillSubfiles`
  (빈 스윕 가드 포함).
- `catalog.yaml`: 34 개 hash 필드 재생성.

## Baseline-attribution

모든 측정은 본 커밋 직전 이 워크트리(브랜치 `WT-catalog-tree-hash`, HEAD = `2660bcd09` = 로컬
develop)에서 이 실행으로 수행. E1 의 RED 는 카탈로그 재생성 이전 상태에서 관측.

## Gaps

- rename/특수문자 경로의 파일명은 -z 파싱 같은 특수 처리 없이 fs.WalkDir 가 직접 다루므로
  해당 없음 — 다만 카탈로그에 디렉터리 외 특수 형태(심링크 등)가 생기면 `!d.Type().IsRegular()`
  로 건너뛴다 (배포 임베드에 심링크 부재 실측은 안 함).
- catalog.yaml 의 `generated_at` 은 스크립트가 갱신하지 않는 기존 동작 그대로 (변경 안 함).
- `make build` 전체 사이클(agents-emit-check 등)은 실행하지 않았다 — 템플릿 원본 파일은 1개도
  변경하지 않았고 catalog.yaml 만 바뀌어, 이 카드에서 agents-emit 축은 자극되지 않는다.

## Residual-risk

- 무결성 주장의 실질 소비자는 여전히 개발기 테스트뿐이다 — 사용자 설치본의 런타임 검증 경로는
  존재하지 않는다 (이 카드가 만든 것이 아니라 처음부터 없던 축; t319·t316 계열과 별도).
- 카탈로그 변동 빈도가 sub-file 변경 때마다 올라간다 — release PR 의 diff 노이즈가 커지면
  배치 크기 관리로 흡수 필요.
