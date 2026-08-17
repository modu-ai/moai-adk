# t47 — 검증 리포트 (ko 정본 승격 + 4로케일 재파생)

카드: t47 — README 4로케일 구조 부채 정리 → ko 신골격 정본 승격
워크트리: `.claude/worktrees/t47` (branch `WT-t47`, base `ca8c0b593` = origin/release/v3.1.1)
커밋: `3fa31ef4d` (규칙 개정) → `3c887d508` (ko 정본) → `c07804841` (en/ja/zh 재파생)

## 1. Claim (주장)

1. 규칙 개정: README 체인 서술이 en 정본 → ko 정본으로 5개 파일에서 개정됐다 (스킬 2 + specialist 2 + Runner 1). docs-site 체인(ko)은 불변.
2. ko 정본 확정: README.ko.md가 12-H2 골격의 정본이다. 문체는 해라체 문어체로 통일됐고(의도적 예외 2곳), en 구골격 고유 콘텐츠가 정보 단위로 흡수됐다 (대조표: absorption-map.md).
3. 재파생: README.md·README.ja.md·README.zh.md가 ko 골격에서 재작성됐고 4파일 구조가 완전히 일치한다.
4. 검증: 패리티·링크·이모지·Mermaid 검사 전부 통과. docs-site 무변경.
5. 사실 정정: en 구골격의 "12 agents × 3 = 36 cells"는 실측 11 에이전트 = 33셀로 정정됐다.

## 2. Evidence (증거)

패리티 (커밋 c07804841 트리에서 측정):

```
$ grep -c '^## ' README.md README.ko.md README.ja.md README.zh.md
README.zh.md:12 / README.md:12 / README.ja.md:12 / README.ko.md:12

$ grep -c '^### ' ...   → 59 / 59 / 59 / 59
$ grep -c '^```' ...    → 42 / 42 / 42 / 42 (짝수)
$ wc -l ...             → 756 / 756 / 756 / 756
```

URL 블랙리스트:

```
$ grep -n 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' README*.md
(매치 0건, exit 1)
```

Mermaid 방향 (4파일 전량):

```
$ grep -n '^```mermaid' -A1 README*.md | grep 'flowchart'
→ 12행 전부 "flowchart TD" (파일당 3개 × 4파일)
```

로케일 링크 치환:

```
$ grep -c 'adk.mo.ai.kr/en/' README.md  → 20
$ grep -c 'adk.mo.ai.kr/ja/' README.ja.md → 20
$ grep -c 'adk.mo.ai.kr/zh/' README.zh.md → 20
$ grep -c 'adk.mo.ai.kr/ko/' README.ko.md → 20
$ grep -n 'adk.mo.ai.kr/ko/' README.md README.ja.md README.zh.md | grep -v 'adk.mo.ai.kr/ko/ |'
→ 본문 잔여 0건 (4-로케일 문서 표의 "한국어" 행만 존재 — 정당)
```

내부 링크 대상 실재: `ls -la CHANGELOG.md LICENSE CONTRIBUTING.md assets/images/{moai-adk-og,kanban-five-sessions,moai-web-overview,moai-web-settings,deepswe-benchmark-2}.png` → 8파일 전부 존재. 로케일별 인포그래픽 `ls assets/images/ | grep -c 'infographic-en\|infographic-ja\|infographic-zh'` → 18 (6종 × 3로케일 전량).

이모지 스캔: 이모지 발견 위치 = 12-에이전트 표의 비용 색 원(🔴🟠🔵🩵⚪, 데이터 표기 — 구 en README가 갖던 동일 허용 형태) + statusline/FAQ 예시 코드블록 내부 + z.ai 팁 💡(구 4파일 공유 형식). 본문 산문의 이모지 0건.

docs-site 무변경: `git status --short docs-site/ | wc -l` → 0. 사유: 카드 범위는 README 표면이며 docs-site는 본래 ko 정본이라 체인 정합성이 자동 충족 — hugo 빌드·사이트 검사 대상 파일이 0개.

사실 정정 측정:

```
$ ~/go/bin/moai model profile --json | grep -c '"agent":'  → 11
```

en 구골격 "12 agents × 3 = 36 cells"의 12는 Explore 포함 카탈로그 수이며 매트릭스 행이 아님. 룰 문서(model-policy.md "11 agents × 3 = 33 cells; Go SSOT")와 일치하는 측정치로 4파일 전부 정정 반영.

Runner 구문·동작 검증: `node --input-type=module -e "import('./.claude/workflows/hns-oss-docs-run.js').then(m => console.log(JSON.stringify(m.derivedLocalesFor('readme-only'))))"` → `[{"surface":"readme","locale":"en","canonical":"ko","file":"README.md"},{...ja...},{...zh...}]` (ko-canonical 타깃 반환 확인). 잔여 en-canonical 서술 grep → 0건.

## 3. Baseline 귀속

모든 측정은 워크트리 `.claude/worktrees/t47`의 HEAD `c07804841` (커밋 후 working tree clean 기준)에서 이 세션이 직접 실행한 명령의 출력이다. `moai model profile`은 설치된 `~/go/bin/moai` (v3.1.0) 기준.

## 4. Gaps (미검증)

- **hugo 빌드 미실행**: docs-site 파일이 0개 변경돼 빌드 대상이 없음. README 자체는 hugo와 무관한 GitHub 표면.
- **렌더링 육안 미검사**: readme-sync 스킬의 "preview the markdown (GitHub-flavored)" 단계를 브라우저에서 돌리지 않았다 — 구조·링크는 기계 검증으로 커버, 시각 렌더링(표 깨짐 등)은 허브 리뷰에서 확인 필요.
- **ja/zh 문체 네이티브 검수**: GLM 5.3이 직접 번역했다. 구조·사실 보존은 기계 검증됐으나 일본어·중국어 모국어 관점의 문장 품질 검수는 사람/허브 리뷰 몫.
- **TRUST 5 상세 표 미반영**: 흡수 대조표 Gaps 참조 — 항목별 검증방법 컬럼은 요약 수준만 흡수.
- **스위처 헤더 4파일 전수 대조**: en은 확인했으나 ja/zh는 구조 동일성에 의존 (H3/펜스/줄 수 완전 일치가 상호 검증됨). 리뷰 시 3줄 육안 확인 권장.

## 5. 잔여 위험

- en 골격이 대폭 교체됐다 (14→12 H2). 외부 문서/번역 사이트가 en README를 인용하고 있었다면 앵커·섹션명이 깨질 수 있다 — GitHub 이슈/외부 링크 점검은 릴리즈 노트에서 안내할 가치가 있다.
- 이미지 재사용: three-axes 등 6종 인포그래픽은 과거 en 골격용으로 제작된 그림을 로케일별 변형으로 재배치했다. 그림 내부 텍스트가 "축(axes)" 어휘를 쓰고 있을 수 있는데 본문은 "세 가지 핵심 (three axes)"로 안내한다 — 이미지 재제작은 별도 작업.
- 병렬 fan-out 실패의 재발 가능성: 서브에이전트 200K 캡는 이 카드에서 3회 재현됐다 (t65 사례와 동일 기저). CLAUDE_CODE_MAX_CONTEXT_TOKENS 대응은 이미 t65에서 진행된 것으로 알고 있으나 이 워크트리 세션 스폰에는 적용되지 않았다 — 후속 카드 후보.

## 산출물

- `README.ko.md` (정본, 756줄) · `README.md` · `README.ja.md` · `README.zh.md` (재파생)
- 규칙 5파일: `.claude/skills/hns-oss-docs-{i18n-rules,readme-sync}/SKILL.md`, `.claude/agents/harness/hns-oss-docs-{content-author,locale-translator}-specialist.md`, `.claude/workflows/hns-oss-docs-run.js`
- `.moai/reports/t47/absorption-map.md` (손실-0 대조표) · 본 파일
