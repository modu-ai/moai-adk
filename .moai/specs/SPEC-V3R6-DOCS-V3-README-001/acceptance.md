# acceptance.md — SPEC-V3R6-DOCS-V3-README-001

> Testable acceptance criteria for README v3 rewrite. 본 문서의 모든 AC는 기계적(grep / diff / wc)으로 검증 가능해야 한다.

---

## §A. AC Summary

| AC ID | REQ | Severity | Axis | Verification method |
|-------|-----|----------|------|---------------------|
| AC-1 | REQ-README-001 | MUST | §1 agent catalog (en) — categories + tier-mapping table | grep "27 agents" 부재 + "8 retained" 존재 + archived name 부재 (categories 전역) + L335-360 tier-mapping line-range 내 archived active 행 부재 |
| AC-2 | REQ-README-001 | MUST | §1 agent catalog (ko) — categories + tier-mapping table + 5 stale count surfaces | grep "24개.*에이전트\|26개.*에이전트" 부재 (L40/L110/L308/L372 전수) + L380-410 tier-mapping 내 archived active 행 부재 + "8 retained" 동등 존재 |
| AC-3 | REQ-README-002 | MUST | §4.2 command set (en + ko 양쪽) | grep "47 Skills\|47개 스킬" 부재 (en+ko) + 17 `/moai` command 명시 |
| AC-4 | REQ-README-003 | MUST | §5 GLM tier (en) | grep "glm-5.2\[1m\]" 존재 + "GLM-5.1" 부재 (GLM 표 내) |
| AC-5 | REQ-README-003 | MUST | §5 GLM tier (ko) | ko 동일 grep |
| AC-6 | REQ-README-004 | MUST | en/ko sync | 양쪽 사실값 (agent count, GLM model, command count) 정확히 일치 |
| AC-7 | REQ-README-005 | MUST | statusline 보존 | "preset" retire 문구 + statusline v3 multi-line 보존 |
| AC-8 | REQ-README-006 | MUST | scope boundary | `git diff --stat`에 `.go` / `docs-site/` / `CLAUDE.md` 부재 |

---

## §B. AC Sub-ID Convention

본 SPEC의 AC는 단일 논리 AC 단위로 `AC-README-001a` / `AC-README-001b` 형태의 sub-ID를 사용하지 않는다 — 각 AC는 단일 grep/diff 검증으로 완결. 단, AC-6 (en/ko sync)은 en/ko 양쪽 검증이 짝을 이루므로, 검증 시나리오에서 `AC-6a (en)` / `AC-6b (ko)` 형태로 짝 검증을 명시할 수 있다 (AC body 내 표기이며, AC ID 자체는 `AC-6` 단일).

---

## §C. Detailed Acceptance Criteria (Given-When-Then)

### AC-1: Agent catalog 정합 (en, REQ-README-001, MUST)

**Given** `README.md`가 본 SPEC의 M1 마일스톤 적용을 받은 상태에서

**When** 다음 grep을 순차 실행하면:
```bash
# (a) stale "27 agents" 부재
grep -c "27 agents" README.md
# (b) stale archived agent name 부재 (categories 전역 — archived-agent-rejection.md 참조 컨텍스트 제외)
grep -iE "manager-strategy|manager-quality|manager-project|manager-brain|claude-code-guide|researcher|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-refactoring|expert-testing|expert-debug" README.md
# (b2) stale "Design System" category 행 부재 (Dn-1 — L296 콘텐츠 앵커 "(+ evaluator)"로 L318 skill-category 행과 구분)
grep -cF 'Design System** | 4 (+ evaluator)' README.md
# (c) "8 retained" 또는 동등 정확 표현 존재
grep -iE "8 retained|8 agents|retained agents" README.md
# (d) L335-360 "Agent Model Assignment by Tier" tier-mapping line-range 내 archived active 행 부재 (D2)
sed -n '335,360p' README.md | grep -iE "manager-strategy|manager-quality|manager-project|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-refactoring|expert-debug"
```

**Then**:
- (a) exit code 1 (count 0, 매치 없음)
- (b) exit code 1 (archived name 매치 없음 — `archived-agent-rejection.md` 참조 컨텍스트는 허용되므로, 해당 매치가 참조 링크 컨텍스트인지 수동 확인)
- (b2) count 0 (stale "Design System | 4 (+ evaluator)" 카테고리 행 제거됨 — Dn-1)
- (c) exit code 0 (1개 이상 매치)
- (d) exit code 1 (L335-360 line-range 내 archived active 행 매치 없음 — 10 archived 이름 전부 제거됨)

### AC-2: Agent catalog 정합 (ko, REQ-README-001, MUST)

**Given** `README.ko.md`가 본 SPEC의 M4 마일스톤 적용을 받은 상태에서

**When** 다음 grep을 실행하면:
```bash
# (a) stale "24개" / "26개" agent-count 부재 — L40/L110/L308/L372 전수 (D1 broadening: AI 누락 variant + 다양한 suffix 커버)
grep -E "24개.*에이전트|26개.*에이전트" README.ko.md
# (b) stale "52개 스킬" / "47개 스킬" 부재
grep -E "52개 스킬|47개 스킬" README.ko.md
# (b2) stale "Agency | 6" 카테고리 행 부재 (Dn-1 ko side — L342 콘텐츠 앵커; design system의 ko-locale 명칭이나 retained 8종에 해당하지 않는 stale 행)
grep -cF '**Agency** | 6' README.ko.md
# (c) 8 retained 동등 ko 표현 존재
grep -E "8개.*retained|8 retained|retained 에이전트" README.ko.md
# (d) L380-410 "티어별 에이전트 모델 할당" tier-mapping line-range 내 archived active 행 부재 (D2)
sed -n '380,410p' README.ko.md | grep -iE "manager-strategy|manager-quality|manager-project|manager-brain|researcher|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-refactoring|expert-testing|expert-debug"
```

**Then**:
- (a) exit code 1 (매치 없음 — L40/L110/L308/L372 4 stale surface 전부 정정됨)
- (b) exit code 1 (매치 없음)
- (b2) count 0 (stale "Agency | 6" 카테고리 행 제거됨 — Dn-1 ko side)
- (c) exit code 0 (1개 이상 매치)
- (d) exit code 1 (L380-410 line-range 내 archived active 행 매치 없음 — 11 archived 이름 전부 제거됨, ko는 `expert-testing` 추가 포함)

### AC-3: `/moai` command set 17 정정 (en + ko 양쪽, REQ-README-002, MUST)

**Given** `README.md` (M3) 및 `README.ko.md` (M4)가 본 SPEC의 적용을 받은 상태에서

**When** 다음 grep을 양쪽 파일에 실행하면:
```bash
# (a) en: stale "47 Skills" 헤더 부재
grep -c "47 Skills" README.md
# (b) ko: stale "47개 스킬" 헤더 부재 (D3 — ko L348 헤더 실존 인정)
grep -c "47개 스킬" README.ko.md
# (c) en: 17 `/moai` command 명시 존재 (둘 중 하나 이상)
grep -iE "17 (commands|slash|/moai)|/moai.*17" README.md
# (d) ko: 동일 (ko 번역 허용: "17개 `/moai` 명령" 등)
grep -E "17.*(명령|commands|/moai)|/moai.*17" README.ko.md
# (e) en: 17-command 리스트 중 plan/run/sync 코어 3개 존재
grep -cE "/moai plan|/moai run|/moai sync" README.md
```

**Then**:
- (a) exit code 1 (count 0)
- (b) exit code 1 (count 0)
- (c) exit code 0 (1개 이상 매치)
- (d) exit code 0 (1개 이상 매치)
- (e) count ≥ 3

### AC-4: GLM tier-model 정합 (en, REQ-README-003, MUST)

**Given** `README.md`가 본 SPEC의 M2 마일스톤 적용을 받은 상태에서

**When** 다음 grep을 실행하면:
```bash
# (a) "glm-5.2[1m]" 존재 (Opus/High tier)
grep -c 'glm-5\.2\[1m\]' README.md
# (b) GLM 표 컨텍스트 내 "GLM-5.1" 부재 (legacy pricing page 참조 컨텍스트는 제외)
grep -nE 'GLM-5\.1' README.md
# (c) Sonnet-tier "glm-4.7" 존재
grep -c 'glm-4\.7' README.md
# (d) Haiku-tier "glm-4.5-air" 존재
grep -ci 'glm-4\.5-air' README.md
```

**Then**:
- (a) count ≥ 1
- (b) exit code 1 (GLM tier 표 컨텍스트 내 매치 없음 — 수동으로 해당 라인이 pricing page URL 컨텍스트인지 확인)
- (c) count ≥ 1
- (d) count ≥ 1

### AC-5: GLM tier-model 정합 (ko, REQ-README-003, MUST)

**Given** `README.ko.md`가 본 SPEC의 M4 마일스톤 적용을 받은 상태에서

**When** AC-4와 동일 grep을 `README.ko.md`에 대해 실행하면

**Then**: AC-4와 동일 (a)/(b)/(c)/(d) 기대치.

### AC-6: en/ko 사실값 동기화 (REQ-README-004, MUST)

**Given** 양쪽 파일이 본 SPEC의 전 마일스톤(M1-M5) 적용을 받은 상태에서

**When** 다음 값을 양쪽에서 추출하여 비교하면:

| 사실값 | en 추출 grep | ko 추출 grep | 기대치 |
|--------|--------------|--------------|--------|
| Retained agent count | `grep -oE "[0-9]+ retained" README.md` | `grep -oE "[0-9]+ retained" README.ko.md` | 양쪽 모두 "8 retained" |
| GLM Opus model | `grep -oE 'glm-5\.2\[1m\]' README.md \| head -1` | `grep -oE 'glm-5\.2\[1m\]' README.ko.md \| head -1` | 양쪽 모두 "glm-5.2[1m]" |
| `/moai` command count | `grep -oE "17 (commands\|slash\|/moai)" README.md` | `grep -oE "17 (commands\|slash\|/moai)" README.ko.md` | 양쪽 모두 17 |

**Then**: 각 행의 en 값과 ko 값이 정확히 일치 (case-insensitive).

### AC-7: statusline 보존 (REQ-README-005, MUST)

**Given** 본 SPEC 작업이 완료된 상태에서

**When** 다음 grep을 양쪽 파일에 실행하면:
```bash
# (a) en: preset retire 문구 보존
grep -i "preset.*retired\|preset.*폐기\|preset shorthand" README.md
# (b) ko: 동일
grep -i "preset.*retired\|preset.*폐기\|preset 단축" README.ko.md
# (c) en: statusline v3 multi-line 보존
grep -i "multi-line\|statusline v3\|status line.*v3" README.md
# (d) ko: 동일
grep -i "멀티라인\|statusline v3" README.ko.md
```

**Then**: (a)(b)(c)(d) 모두 exit code 0 (1개 이상 매치) — 기존에 반영된 상태가 회귀하지 않았음을 확인.

### AC-8: scope boundary 준수 (REQ-README-006, MUST)

**Given** 본 SPEC의 모든 마일스톤 커밋이 완료된 상태에서

**When** 다음을 실행하면:
```bash
# (a) `.go` 파일 변경 부재
git diff --stat 4a6f4b4d3..HEAD -- '*.go' | wc -l
# (b) docs-site 변경 부재
git diff --stat 4a6f4b4d3..HEAD -- 'docs-site/' | wc -l
# (c) CLAUDE.md 변경 부재
git diff --stat 4a6f4b4d3..HEAD -- 'CLAUDE.md' | wc -l
# (d) template 변경 부재
git diff --stat 4a6f4b4d3..HEAD -- 'internal/template/templates/' | wc -l
# (e) README 2개 파일만 변경
git diff --stat 4a6f4b4d3..HEAD -- 'README.md' 'README.ko.md' | wc -l
```

**Then**:
- (a) count 0 (빈 출력)
- (b) count 0
- (c) count 0
- (d) count 0
- (e) count ≥ 2 (README.md + README.ko.md 2개 파일)

---

## §D. Edge Cases

### Edge-1: archived agent 이름이 `archived-agent-rejection.md` 참조 컨텍스트에서 등장

AC-1(b) / AC-2(b) 의 grep이 archived name을 매치할 때, 해당 매치가 "이름들은 archived이니 `archived-agent-rejection.md`를 참조하라"는 안내 컨텍스트인 경우는 PASS로 판정. 이 경우 매치 라인을 수동으로 읽어 참조 컨텍스트인지 확인. 단, AC-1(d) / AC-2(d)의 line-range grep (L335-360 / L380-410 tier-mapping table)은 참조 컨텍스트를 허용하지 않는다 — tier table 내 행은 항상 active model-policy 행으로 해석되므로, archived name이 해당 line-range에 등장하면 무조건 FAIL.

### Edge-1b: tier-mapping table line-range drift (AC-1(d) / AC-2(d) 라인 번호 이동)

본 SPEC의 M1/M4 작업이 README 다른 섹션을 축소/확장하면 L335-360 / L380-410 라인 번호가 이동할 수 있다. AC-1(d) / AC-2(d)의 `sed -n` line-range는 작성 시점(4a6f4b4d3) 기준이므로, run-phase 시점에 라인 번호가 이동했다면 먼저 `grep -n "Agent Model Assignment by Tier\|티어별 에이전트 모델 (할당|배정)"` 으로 실제 tier-mapping table 시작 라인을 찾고, 해당 테이블 끝 라인까지 동적으로 범위를 재설정하여 archived active 행 부재를 검증해야 한다. (Dn-2 — ko branch `(할당|배정)` broadening: README.ko.md L382 실제 헤더는 "티어별 에이전트 모델 배정"이며, narrow `할당` 패턴은 exit 1 no-match로 fallback 목적 상실.) 이것은 manager-develop의 M1/M4 run-phase 작업 시 기계적 검증 절차에 포함된다.

### Edge-2: "GLM-5.1"이 z.ai pricing page URL 설명 컨텍스트에서 등장

AC-4(b)의 grep이 "GLM-5.1"을 매치할 때, 해당 매치가 "z.ai는 GLM-5.1 등 다양한 모델을 제공한다"는 pricing page 참조 컨텍스트인 경우, README의 tier-model 표에는 `glm-5.2[1m]`가 기재되어 있으므로 PASS. 단, 본 SPEC의 권장사항은 GLM 표 내에서는 `glm-5.2[1m]`만 사용하고, legacy `GLM-5.1` 언급은 제거하는 것.

### Edge-3: en/ko 동기화 시 ko 번역 품질

AC-6은 사실값(count, model name)의 일치만 검증. ko 번역의 자연스러움(en의 의미를 ko가 정확히 반영하는가)은 기계적 검증이 불가능하므로, manager-develop의 M4 작업 시 번역 품질 review를 별도 수반. 본 AC의 기계적 검증은 사실값 일치에 한정.

### Edge-4: Progressive Disclosure 시스템 설명 보존

M3에서 "47 Skills" 헤더를 제거할 때, Progressive Disclosure 3-level 시스템(skill loading 메커니즘) 설명은 여전히 유효하므로 보존해야 함. AC-3은 "47 Skills" 헤더 부재만 검증하며, Progressive Disclosure 설명 자체의 존재는 별도 AC로 두지 않음 (본 SPEC 범위 밖 — skill 시스템 구조 변경이 아님).

---

## §E. Indirect Verification

본 SPEC은 documentation-only이므로 다음 간접 검증이 적용:

- **E-lint (spec-lint)**: `moai spec lint SPEC-V3R6-DOCS-V3-README-001` → 0 findings (FrontmatterInvalid / OwnershipTransitionInvalid 부재)
- **E-neutrality**: `internal/template/internal_content_leak_test.go` PASS — 본 SPEC은 template 변경 없음
- **E-build**: N/A (Go 코드 변경 없음)
- **E-test**: N/A (Go 코드 변경 없음)

---

## §F. Quality Gate Criteria (Definition of Done)

본 SPEC이 "완료"로 판정되기 위한 조건:

- [ ] AC-1 ~ AC-8 전수 PASS (위 Given-When-Then 시나리오 대로)
- [ ] `moai spec lint SPEC-V3R6-DOCS-V3-README-001` → 0 findings
- [ ] `git diff --stat 4a6f4b4d3..HEAD` → README.md + README.ko.md 2개 파일만
- [ ] M1 ~ M6 마일스톤 전수 커밋 존재
- [ ] en/ko 사실값 정확히 일치 (AC-6)
- [ ] statusline 섹션 회귀 없음 (AC-7)
- [ ] Go 코드 / docs-site / CLAUDE.md / template 변경 없음 (AC-8)

---

## §G. Forward-Looking Checks (후속 SPEC 권장)

본 SPEC 완료 후 다음이 별도 SPEC으로 권장됨 (본 SPEC scope 밖):

1. **skill-audit SPEC**: skill catalog의 정확한 총수 산출 (현재 "47" 또는 "52"로 불분명) — 본 SPEC은 "47 Skills" 헤더 제거만 수행
2. **DOCSITE-001**: docs-site 4-locale content (`docs-site/content/{en,ko,ja,zh}/`)의 동일 사실 정합
3. **GLM pricing SPEC**: README GLM 표의 input/output 단가를 z.ai 현재 pricing과 정합 (본 SPEC은 모델명만)

이 항목들은 본 SPEC의 AC가 아님 — forward-looking 참고용.

---

## HISTORY

- 2026-06-17: acceptance criteria authored. 8 AC (AC-1 ~ AC-8). 모든 AC는 grep/diff/wc 기반 기계적 검증 가능. documentation-only SPEC이므로 build/coverage AC는 N/A.
- 2026-06-17 (iter-2, v0.2.0): plan-auditor PASS-WITH-DEBT 0.83 → D1/D2/D3 정정. AC-1 (d) 추가 (en L335-360 tier-mapping archived active 행 부재), AC-2 (a) regex broadening (`24개.*에이전트|26개.*에이전트` — L308/L372 variant 커버) + (d) 추가 (ko L380-410 tier-mapping archived active 행 부재, `expert-testing` 포함 11종), AC-3 양 locale 커버 확장 (D3 — ko L348 "47개 스킬" 헤더 + L350-368 count 표 실존 인정). Edge-1b 추가 (line-range drift 대응 절차).
- 2026-06-17 (iter-3, v0.2.1): plan-auditor iter-2 추가 2 MINOR defect surgical patch. Dn-1 — AC-1 (b2) 추가 (en L296 "Design System** | 4 (+ evaluator)" 콘텐츠 앵커 grep), AC-2 (b2) 추가 (ko L342 "**Agency** | 6" 콘텐츠 앵커 grep). Dn-2 — Edge-1b ko branch `할당` → `(할당|배정)` broadening (README.ko.md L382 "배정" 반영). AC 총수 8 유지 (새 AC 추가 없이 기존 AC에 sub-letter (b2) 확장만).
